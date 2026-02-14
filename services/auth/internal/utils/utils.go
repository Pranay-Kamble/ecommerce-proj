package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
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
