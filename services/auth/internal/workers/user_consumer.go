package workers

import (
	"context"
	"ecommerce/pkg/broker"
	"ecommerce/pkg/logger"
	"ecommerce/services/auth/internal/repository"
	"encoding/json"

	"go.uber.org/zap"
)

type UserOnboardedEvent struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

func StartUserEventsConsumer(r *broker.RabbitMQClient, userRepo repository.UserRepository, queueName string) {
	ctx := context.Background()
	channel, err := r.Consume(ctx, queueName)
	if err != nil {
		logger.Fatal("workers: failed to start user events consumer: ", zap.Error(err))
	}

	go func() {
		for event := range channel {
			var message UserOnboardedEvent

			err = json.Unmarshal(event.Body, &message)
			if err != nil {
				logger.Error("workers: failed to unmarshal event body", zap.Error(err))
				_ = event.Ack(false)
				continue
			}

			if event.RoutingKey == "seller.onboarded" && message.Status == "onboarded" {
				err = userRepo.UpdateOnboardingStatus(ctx, message.UserID, true)
				if err != nil {
					logger.Error("workers: failed to update onboarding status", zap.Error(err))
					_ = event.Ack(false)
					continue
				}
				logger.Info("workers: successfully onboarded seller", zap.String("user_id", message.UserID))
			}

			_ = event.Ack(false)
		}
	}()
}
