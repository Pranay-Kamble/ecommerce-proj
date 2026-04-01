package workers

import (
	"context"
	"sync"
	"time"

	"ecommerce/pkg/logger"
	"ecommerce/services/payment/internal/domain"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OutboxWorker struct {
	db       *gorm.DB
	rabbitMQ *amqp.Channel
}

func NewOutboxWorker(db *gorm.DB, rabbitMQ *amqp.Channel) *OutboxWorker {
	return &OutboxWorker{db: db, rabbitMQ: rabbitMQ}
}

func (ow *OutboxWorker) StartOutboxWorker(ctx context.Context) {
	logger.Info("worker: Outbox processor started successfully, polling every 5 seconds")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("worker: Outbox processor shutting down gracefully")
			return
		case <-ticker.C:
			ow.processOutboxEvents(ctx)
		}
	}
}

func (ow *OutboxWorker) processOutboxEvents(ctx context.Context) {
	var events []*domain.OutboxEvent

	err := ow.db.WithContext(ctx).Where("processed = ?", false).Limit(50).Find(&events).Error
	if err != nil {
		logger.Error("worker: failed to retrieve unprocessed events", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return
	}

	logger.Info("worker: picked up pending outbox events", zap.Int("event_count", len(events)))

	var waitgroup sync.WaitGroup

	for _, event := range events {
		waitgroup.Add(1)

		go func(e *domain.OutboxEvent) {
			defer waitgroup.Done()

			logger.Info("worker: attempting to publish event",
				zap.String("event_id", e.ID.String()),
				zap.String("event_type", e.EventType),
			)

			innerErr := ow.brokerPublishEvent(ctx, e)

			if innerErr != nil {
				logger.Error("worker: failed to publish event to RabbitMQ",
					zap.String("event_id", e.ID.String()),
					zap.Error(innerErr),
				)
				return
			}

			dbErr := ow.db.WithContext(ctx).Delete(e).Error
			if dbErr != nil {
				logger.Error("worker: CRITICAL - event published but failed to delete from DB",
					zap.String("event_id", e.ID.String()),
					zap.Error(dbErr),
				)
				return
			}

			logger.Info("worker: successfully published and deleted event",
				zap.String("event_id", e.ID.String()),
				zap.String("routing_key", "payment."+e.EventType),
			)

		}(event)
	}

	waitgroup.Wait()
}

func (ow *OutboxWorker) brokerPublishEvent(ctx context.Context, event *domain.OutboxEvent) error {
	routingKey := "payment." + event.EventType

	err := ow.rabbitMQ.PublishWithContext(ctx,
		"payment_events",
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    event.ID.String(),
			Body:         []byte(event.Payload),
		},
	)

	return err
}
