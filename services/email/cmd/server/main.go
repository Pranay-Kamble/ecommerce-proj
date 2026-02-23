package main

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/email/internal/handler"
	"ecommerce/services/email/internal/service"
	"log"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/wneessen/go-mail"
	"go.uber.org/zap"
)

func main() {
	r := gin.Default()
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("main: Error loading .env file")
		return
	}

	environment := os.Getenv("ENV_TYPE")
	if environment == "" {
		environment = "prod"
	}
	logger.Init(environment)

	smtpServerAddress := os.Getenv("SMTP_SERVER")
	if smtpServerAddress == "" {
		smtpServerAddress = "smtp.gmail.com"
	}

	username := os.Getenv("SMTP_USER")
	if username == "" {
		logger.Fatal("main: Error loading SMTP username")
		return
	}

	password := os.Getenv("SMTP_PASS")
	if password == "" {
		logger.Fatal("main: Error loading SMTP password")
		return
	}

	smtpPort := 587
	portStr := os.Getenv("SMTP_PORT")
	if portStr != "" {
		parsedPort, err := strconv.Atoi(portStr)
		if err != nil {
			logger.Fatal("main: SMTP_PORT must be a valid number")
			return
		}
		smtpPort = parsedPort
	}

	smtpServer, err := mail.NewClient(smtpServerAddress, mail.WithPort(smtpPort), mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithTLSPolicy(mail.TLSMandatory),
		mail.WithUsername(username), mail.WithPassword(password))

	if err != nil {
		logger.Fatal("main: Error initializing SMTP connection")
		return
	}

	defer func(emailClient *mail.Client) {
		err := emailClient.Close()
		if err != nil {
			logger.Error("main: Error closing SMTP connection cleanly")
		}
	}(smtpServer)

	emailService := service.NewEmailService(smtpServer)
	emailHandler := handler.NewEmailHandler(emailService)

	handler.RegisterRoutes(r, emailHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	if err := r.Run(":" + port); err != nil {
		logger.Fatal("main: failed to start server", zap.Error(err))
	}
}
