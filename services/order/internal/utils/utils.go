package utils

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

var publicKey *rsa.PublicKey = nil

func GetPublicKey() error {
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		return errors.New("no auth service url found")
	}

	authServiceURL += "/api/v1/auth/public-key"

	response, err := http.Get(authServiceURL)
	if err != nil {
		return fmt.Errorf("middleware: failed to get public key from auth service: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("middleware: failed to get public key from auth service: status_code: %d", response.StatusCode)
	}

	var body map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&body)

	if err != nil {
		return fmt.Errorf("middleware: failed to decode response from auth service: %w", err)
	}

	publicKeyString, ok := body["publicKey"].(string)
	if !ok {
		return errors.New("middleware: failed to get public key from auth service: publicKey not found in response")
	}

	publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyString))
	if err != nil {
		return fmt.Errorf("middleware: failed to parse public key from auth service: %w", err)
	}

	return nil
}

func VerifyJWT(tokenString string) (jwt.MapClaims, error) {
	if publicKey == nil {
		return nil, errors.New("utils: public key is empty")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("utils: unexpected signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("utils: failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("utils: invalid token claims format")
}
