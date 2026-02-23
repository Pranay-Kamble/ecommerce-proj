package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/email/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EmailHandler interface {
	VerificationEmail(ctx *gin.Context)
}

type VerificationEmailHandler struct {
	To  string `json:"to" binding:"email,required"`
	OTP string `json:"otp" binding:"required,len=6,numeric"`
}

type emailHandler struct {
	service service.EmailService
}

func NewEmailHandler(service service.EmailService) EmailHandler {
	return &emailHandler{service: service}
}

func (e *emailHandler) VerificationEmail(c *gin.Context) {
	var body VerificationEmailHandler
	err := c.ShouldBindJSON(&body)

	if err != nil {
		logger.Error("handler: could not bind request: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "OTP and Receiver's email address is required"})
		return
	}

	err = e.service.SendVerificationEmail(c, body.To, body.OTP)
	if err != nil {
		logger.Error("handler: could not send verification email", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "email sent successfully"})
}
