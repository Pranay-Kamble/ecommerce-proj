package utils

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var (
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
)

func HashPassword(password string) (string, error) {
	plainTextPassword := []byte(password)

	hash, err := bcrypt.GenerateFromPassword(plainTextPassword, bcrypt.DefaultCost)

	if errors.Is(err, bcrypt.ErrPasswordTooLong) {
		return "", errors.New("utils: password too long")
	}

	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func VerifyPassword(password, hash string) bool {
	byteHash := []byte(hash)
	bytePassword := []byte(password)

	err := bcrypt.CompareHashAndPassword(byteHash, bytePassword)

	return err == nil
}

func InitKeys() error {
	publicKeyPEM, err := os.ReadFile("../secrets/public.pem")
	if err != nil {
		return fmt.Errorf("utils: failed to read public key: %w", err)
	}
	publicKey, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyPEM)
	if err != nil {
		return fmt.Errorf("utils: failed to parse public key: %w", err)
	}

	privateKeyPEM, err := os.ReadFile("../secrets/private.pem")
	if err != nil {
		return fmt.Errorf("utils: failed to read private key: %w", err)
	}
	privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyPEM)
	if err != nil {
		return fmt.Errorf("utils: failed to parse private key: %w", err)
	}
	return nil
}

func GetJWT(ID, email, role string) (string, error) {
	if privateKey == nil {
		return "", errors.New("utils: private key is empty")
	}

	claims := jwt.MapClaims{
		"id":    ID,
		"email": email,
		"role":  role,
		"iat":   time.Now().Unix(),
		"exp":   time.Now().Add(15 * time.Minute).Unix(),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	signedToken, err := jwtToken.SignedString(privateKey)

	if err != nil {
		return "", fmt.Errorf("utils: failed to sign token: %w", err)
	}

	return signedToken, nil
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
