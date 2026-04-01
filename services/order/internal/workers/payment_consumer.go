package workers

import (
	"context"
	"encoding/json"
	"fmt"

	"ecommerce/pkg/logger"
	"ecommerce/services/order/internal/repository" // <-- Import your repository package
	"ecommerce/services/order/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

type PaymentConsumer struct {
	rabbitChannel *amqp.Channel
	orderService  service.OrderService
	cartRepo      repository.CartRepository
}

type PaymentEventPayload struct {
	OrderID string `json:"order_id"`
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
}

func NewPaymentConsumer(ch *amqp.Channel, svc service.OrderService, cartRepo repository.CartRepository) *PaymentConsumer {
	return &PaymentConsumer{
		rabbitChannel: ch,
		orderService:  svc,
		cartRepo:      cartRepo,
	}
}
func (c *PaymentConsumer) StartListening(ctx context.Context) error {
	queue, err := c.rabbitChannel.QueueDeclare(
		"order_service_payment_queue", // name
		true,                          // durable
		false,                         // delete when unused
		false,                         // exclusive
		false,                         // no-wait
		nil,                           // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare a queue: %w", err)
	}

	err = c.rabbitChannel.ExchangeDeclare(
		"payment_events", // name
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

	err = c.rabbitChannel.QueueBind(
		queue.Name,          // queue name
		"payment.OrderPaid", // routing key we are listening for
		"payment_events",    // exchange name
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	messages, err := c.rabbitChannel.Consume(
		queue.Name,
		"order_service_consumer", // consumer tag
		false,                    // auto-ack (set to false)
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	logger.Info("Order Service is now listening for payment events...")

	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down payment consumer gracefully...")
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

func (c *PaymentConsumer) processMessage(ctx context.Context, msg amqp.Delivery) {
	var payload PaymentEventPayload

	if err := json.Unmarshal(msg.Body, &payload); err != nil {
		logger.Error("Failed to unmarshal payment event, dropping message", zap.Error(err))
		msg.Nack(false, false)
		return
	}

	if payload.Status == "paid" || payload.Status == "success" {

		err := c.orderService.UpdateOrderStatus(ctx, payload.OrderID, "paid")
		if err != nil {
			logger.Error("Failed to update order status in DB", zap.Error(err), zap.String("order", payload.OrderID))
			msg.Nack(false, true)
			return
		}

		order, err := c.orderService.GetOrder(ctx, payload.OrderID, payload.UserID)
		if err != nil {
			logger.Error("Failed to fetch order to clear cart", zap.Error(err))
		} else {
			err = c.cartRepo.ClearCart(ctx, order.UserID)
			if err != nil {
				logger.Error("Failed to clear cart, but order was paid", zap.Error(err))
			} else {
				logger.Info("Cart successfully cleared for user", zap.String("user_id", order.UserID))
			}
		}
	}

	msg.Ack(false)
	logger.Info("Order successfully marked as paid from RabbitMQ event!", zap.String("order_id", payload.OrderID))
}
