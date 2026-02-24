package service

import (
	"context"
	"ecommerce/services/email/internal/utils"
	"fmt"
	"os"

	"github.com/resend/resend-go/v2"
)

type EmailService interface {
	SendVerificationEmail(ctx context.Context, to string, OTP string) error
}

type emailService struct {
	client *resend.Client
}

func NewEmailService(client *resend.Client) EmailService {
	return &emailService{client: client}
}

func (e *emailService) SendVerificationEmail(ctx context.Context, to, OTP string) error {
	from := os.Getenv("FROM_EMAIL")
	if from == "" {
		from = "onboarding@resend.dev"
	}

	payload, err := utils.GenerateHTMLBody(OTP)
	if err != nil {
		return fmt.Errorf("service: could not generate HTML body: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    from,
		To:      []string{to},
		Subject: "OTP for Email Verification",
		Html:    payload,
	}

	_, err = e.client.Emails.SendWithContext(ctx, params)
	if err != nil {
		return fmt.Errorf("service: could not send email verification email via resend: %w", err)
	}

	return nil
}
