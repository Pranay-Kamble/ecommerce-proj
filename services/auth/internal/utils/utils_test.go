package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	plainPasswords := []string{
		"password1", "password2",
		"CRICKET321", "notAGoodPassword",
		"wowowowowow", "aVeryBadPassword",
		"heheheh", "1234567890",
	}

	for _, plainPassword := range plainPasswords {

		hash, err := HashPassword(plainPassword)
		assert.NoError(t, err)
		assert.NotEmpty(t, hash)

		isValid := VerifyPassword(plainPassword, hash)
		assert.True(t, isValid, "Password should verify successfully: %s", plainPassword)
	}
}
