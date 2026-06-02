package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"ecommerce/pkg/logger"
	"ecommerce/services/search/internal/domain"
	"ecommerce/services/search/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// CatalogConsumer listens to catalog events from RabbitMQ and indexes/deletes
// products in Elasticsearch. Follows the same pattern as PaymentConsumer in the
// order service.
type CatalogConsumer struct {
	rabbitChannel *amqp.Channel
	searchService service.SearchService
}

// CatalogEventPayload represents the message published by the catalog service.
type CatalogEventPayload struct {
	Event   string                 `json:"event"`
	Product map[string]interface{} `json:"product,omitempty"`
	// For delete events, public_id is at the top level
	PublicID string `json:"public_id,omitempty"`
}

// NewCatalogConsumer creates a new CatalogConsumer.
func NewCatalogConsumer(ch *amqp.Channel, svc service.SearchService) *CatalogConsumer {
	return &CatalogConsumer{
		rabbitChannel: ch,
		searchService: svc,
	}
}

// StartListening declares the queue, binds it to the catalog_events exchange,
// and starts consuming messages. It blocks until the context is cancelled.
func (c *CatalogConsumer) StartListening(ctx context.Context) error {
	queue, err := c.rabbitChannel.QueueDeclare(
		"search_service_catalog_queue", // name
		true,                           // durable
		false,                          // delete when unused
		false,                          // exclusive
		false,                          // no-wait
		nil,                            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	err = c.rabbitChannel.ExchangeDeclare(
		"catalog_events", // name
		"topic",          // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Bind with wildcard to catch product.created, product.updated, product.deleted
	err = c.rabbitChannel.QueueBind(
		queue.Name,       // queue name
		"product.#",      // routing key pattern
		"catalog_events", // exchange name
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	messages, err := c.rabbitChannel.Consume(
		queue.Name,
		"search_service_consumer", // consumer tag
		false,                     // auto-ack (set to false for manual ack)
		false,                     // exclusive
		false,                     // no-local
		false,                     // no-wait
		nil,                       // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	logger.Info("Search Service is now listening for catalog events...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down catalog consumer gracefully...")
			return nil
		case msg, ok := <-messages:
			if !ok {
				logger.Error("RabbitMQ channel closed")
				return nil
			}
			c.processMessage(ctx, msg)
		}
	}
}

// processMessage handles a single catalog event message.
func (c *CatalogConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var payload CatalogEventPayload

	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logger.Error("Failed to unmarshal catalog event, dropping message",
			zap.Error(err), zap.String("body", string(msg.Body)))
		msg.Nack(false, false)
		return
	}

	logger.Info("Received catalog event", zap.String("event", payload.Event))

	switch payload.Event {
	case "product.created", "product.updated":
		doc, err := c.parseProductDocument(payload.Product)
		if err != nil {
			logger.Error("Failed to parse product document from event",
				zap.Error(err), zap.String("event", payload.Event))
			msg.Nack(false, false)
			return
		}

		if err = c.searchService.IndexProduct(ctx, doc); err != nil {
			logger.Error("Failed to index product in Elasticsearch",
				zap.Error(err), zap.String("public_id", doc.PublicID))
			msg.Nack(false, true) // requeue for transient failures
			return
		}

		logger.Info("Successfully indexed product",
			zap.String("event", payload.Event),
			zap.String("public_id", doc.PublicID))

	case "product.deleted":
		publicID := payload.PublicID
		if publicID == "" {
			logger.Error("product.deleted event missing public_id, dropping message")
			msg.Nack(false, false)
			return
		}

		if err := c.searchService.DeleteProduct(ctx, publicID); err != nil {
			logger.Error("Failed to delete product from Elasticsearch",
				zap.Error(err), zap.String("public_id", publicID))
			msg.Nack(false, true)
			return
		}

		logger.Info("Successfully deleted product from index",
			zap.String("public_id", publicID))

	default:
		logger.Info("Unknown catalog event type, ignoring", zap.String("event", payload.Event))
	}

	msg.Ack(false)
}

// parseProductDocument converts the product map from the event payload to a ProductDocument.
func (c *CatalogConsumer) parseProductDocument(product map[string]interface{}) (domain.ProductDocument, error) {
	if product == nil {
		return domain.ProductDocument{}, fmt.Errorf("product data is nil")
	}

	// Re-marshal and unmarshal to leverage JSON tags
	data, err := json.Marshal(product)
	if err != nil {
		return domain.ProductDocument{}, fmt.Errorf("failed to marshal product map: %w", err)
	}

	var doc domain.ProductDocument
	if err = json.Unmarshal(data, &doc); err != nil {
		return domain.ProductDocument{}, fmt.Errorf("failed to unmarshal product document: %w", err)
	}

	if doc.PublicID == "" {
		return domain.ProductDocument{}, fmt.Errorf("product document missing public_id")
	}

	return doc, nil
}
