package client

import (
	"bytes"
	"ecommerce/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

type EmailClient interface {
	SendVerificationEmail(email string, otp string) error
}

type emailClient struct {
	baseUrl string
	client  *http.Client
}

func NewEmailClient(baseUrl string) EmailClient {
	return &emailClient{
		baseUrl: baseUrl,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

func (e *emailClient) SendVerificationEmail(toEmail string, otp string) error {
	payload := map[string]string{
		"to":  toEmail,
		"otp": otp,
	}

	body, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("client: failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/verification-email", e.baseUrl)
	response, err := e.client.Post(url, "application/json", bytes.NewBuffer(body))

	if err != nil {
		return fmt.Errorf("client: failed to send verification email: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("client: failed to close response body: ", zap.Error(err))
		}
	}(response.Body)

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("client: failed to send verification email: %s", response.Status)
	}

	return nil
}
