package workers

import (
	"context"
	"ecommerce/pkg/logger"
	"ecommerce/services/payment/internal/domain"
	"sync"
	"time"

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
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("OutboxWorker is shutting down")
			return
		case <-ticker.C:
			ow.processOutboxEvents(ctx)
		}
	}
}

func (ow *OutboxWorker) processOutboxEvents(ctx context.Context) {
	events, err := gorm.G[*domain.OutboxEvent](ow.db).Where("processed = ?", false).Limit(50).Find(ctx)

	if err != nil {
		logger.Error("worker: failed to retrieve unprocessed events: ", zap.Error(err))
		return
	}

	if len(events) == 0 {
		return
	}

	waitgroup := &sync.WaitGroup{}
	for _, event := range events {
		waitgroup.Add(1)
		go func(e *domain.OutboxEvent) {
			defer waitgroup.Done()
			innerErr := ow.brokerPublishEvent(ctx, e)

			if innerErr != nil {
				logger.Error("worker: failed to publish event", zap.Error(innerErr))
			} else {
				ow.db.Delete(e)
				logger.Info("worker: event is published", zap.Any("event", e))
			}
		}(event)
	}

	waitgroup.Wait()
}

func (ow *OutboxWorker) brokerPublishEvent(ctx context.Context, event *domain.OutboxEvent) error {
	routingKey := "payment." + event.EventType

	err := ow.rabbitMQ.PublishWithContext(ctx,
		"payment_events", // exchange declared in main.go
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
