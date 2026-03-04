package utils

import (
	"crypto/rsa"
	"ecommerce/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	publicKey *rsa.PublicKey
	keyMutex  sync.RWMutex
)

func VerifyJWT(tokenString string) (jwt.MapClaims, error) {
	pubKey, err := getPublicKey()
	if err != nil {
		return nil, fmt.Errorf("utils: auth key error: %w", err)
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return pubKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("utils: failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("utils: invalid token signature")
}

func getPublicKey() (*rsa.PublicKey, error) {
	keyMutex.RLock()
	if publicKey != nil {
		keyMutex.RUnlock()
		return publicKey, nil
	}
	keyMutex.RUnlock()

	keyMutex.Lock()
	defer keyMutex.Unlock()

	if publicKey != nil {
		return publicKey, nil
	}

	key, err := fetchPublicKeyFromAuth()
	if err != nil {
		return nil, err
	}

	publicKey = key
	return publicKey, nil
}

func fetchPublicKeyFromAuth() (*rsa.PublicKey, error) {
	authService := os.Getenv("AUTH_SERVICE_BASE_URL")
	if authService == "" {
		return nil, errors.New("AUTH_SERVER_BASE_URL is missing")
	}

	url := strings.TrimRight(authService, "/") + "/api/v1/auth/public-key"

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to reach auth service: %w", err)
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("failed to close auth service: %w", zap.Error(err))
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth server returned status %d", resp.StatusCode)
	}

	var payload struct {
		PublicKey string `json:"publicKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("failed to decode json payload: %w", err)
	}

	return jwt.ParseRSAPublicKeyFromPEM([]byte(payload.PublicKey))
}
