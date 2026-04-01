package utils

import (
	"bytes"
	"ecommerce/services/email/internal/domain"
	"fmt"
	"html/template"
	"os"
	"sync"
)

var (
	verificationEmailTemplate *template.Template
	tmplOnce                  sync.Once
	tmplErr                   error
)

func GenerateHTMLBody(otp string) (string, error) {
	tmplOnce.Do(func() {
		verificationEmailTemplate, tmplErr = template.ParseFiles("./internal/templates/email_verification.html")
	})

	if tmplErr != nil {
		str, _ := os.Getwd()
		fmt.Println("PWD: " + str)
		return "", fmt.Errorf("utils: failed to parse verification email template: %w", tmplErr)
	}

	var body bytes.Buffer
	data := domain.EmailData{OTP: otp}
	if err := verificationEmailTemplate.Execute(&body, data); err != nil {
		return "", fmt.Errorf("utils: could not generate HTML body: %w", err)
	}

	return body.String(), nil
}
