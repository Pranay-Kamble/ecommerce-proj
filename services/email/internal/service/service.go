package service

import (
	"context"
	"ecommerce/services/email/internal/utils"
	"fmt"
	"os"

	"github.com/wneessen/go-mail"
)

type EmailService interface {
	SendVerificationEmail(ctx context.Context, to string, OTP string) error
}

type emailService struct {
	smtp *mail.Client
}

func NewEmailService(smtp *mail.Client) EmailService {
	return &emailService{smtp: smtp}
}

func (e *emailService) SendVerificationEmail(ctx context.Context, to, OTP string) error {
	message := mail.NewMsg()
	err := message.To(to)
	if err != nil {
		return fmt.Errorf("service: could not attach the receiver's email address: %w", err)
	}

	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		return fmt.Errorf("service: could not get receiver's email address")
	}

	err = message.From(from)
	if err != nil {
		return fmt.Errorf("service: could not attach the receiver's email address: %w", err)
	}

	payload, err := utils.GenerateHTMLBody(OTP)
	if err != nil {
		return fmt.Errorf("service: could not generate HTML body: %w", err)
	}

	message.Subject("OTP for Email Verification")
	message.SetBodyString("text/html", payload)

	err = e.smtp.DialAndSendWithContext(ctx, message)
	if err != nil {
		return fmt.Errorf("service: could not send email verification email: %w", err)
	}

	return nil
}
