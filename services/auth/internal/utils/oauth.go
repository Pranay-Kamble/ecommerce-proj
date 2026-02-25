package utils

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var OAuth *oauth2.Config

func InitOAuth() error {
	clientId := os.Getenv("GOOGLE_CLIENT_ID")
	if clientId == "" {
		return fmt.Errorf("utils: env variable CLIENT_ID not found")
	}

	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if clientSecret == "" {
		return fmt.Errorf("utils: env variable CLIENT_SECRET not found")
	}

	redirectUrl := os.Getenv("REDIRECT_URL")
	if redirectUrl == "" {
		return fmt.Errorf("utils: env variable REDIRECT_URL not found")
	}

	scopes := []string{
		"openid",
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}

	endpoint := google.Endpoint

	OAuth = &oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  redirectUrl,
		Endpoint:     endpoint,
		Scopes:       scopes,
	}

	return nil
}
