package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/payment/internal/service"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"
	"go.uber.org/zap"
)

type WebhookHandler struct {
	paymentSvc    service.PaymentService
	webhookSecret string
}

func NewWebhookHandler(paymentSvc service.PaymentService, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{paymentSvc: paymentSvc, webhookSecret: webhookSecret}
}

func (wh *WebhookHandler) handleWebhook(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 65536)

	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		logger.Error("handler: failed to read request body", zap.Error(err))
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "failed to read request body"})
		return
	}

	signatureHeader := c.GetHeader("Stripe-Signature")
	if signatureHeader == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing Stripe-Signature header"})
		return
	}

	event, err := webhook.ConstructEvent(payload, signatureHeader, wh.webhookSecret)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "verification failed"})
		return
	}

	if event.Type == "checkout.session.completed" {
		var session stripe.CheckoutSession
		err = json.Unmarshal(event.Data.Raw, &session)

		if err != nil {
			logger.Error("handler: failed to unmarshal checkout session", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		err = wh.paymentSvc.MarkPaymentAsSuccess(c.Request.Context(), session.ID)

		if err != nil {
			logger.Error("handler: failed to mark payment as success ", zap.Error(err))
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to mark payment as success"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}
