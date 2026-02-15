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

func TestInitKeys(t *testing.T) {

	err := InitKeys()

	assert.NoError(t, err, "InitKeys failed: check if the .pem files are in the right folder")
}

func TestJWTWorkflow(t *testing.T) {
	err := InitKeys()
	assert.NoError(t, err)

	testID := "nano-12345"
	testEmail := "test@example.com"
	testRole := "admin"

	tokenString, err := GetJWT(testID, testEmail, testRole)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString, "Token should not be empty")

	claims, err := VerifyJWT(tokenString)

	assert.NoError(t, err, "Valid token should not throw an error")
	assert.NotNil(t, claims, "Claims should be returned")

	assert.Equal(t, testID, claims["id"])
	assert.Equal(t, testEmail, claims["email"])
	assert.Equal(t, testRole, claims["role"])
}

func TestVerifyJWTHacker(t *testing.T) {
	err := InitKeys()
	assert.NoError(t, err)

	fakeToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.fake.signature"
	_, err = VerifyJWT(fakeToken)
	assert.Error(t, err, "A fake token MUST throw an error")

	validToken, _ := GetJWT("123", "hacker@evil.com", "user")
	tamperedToken := validToken[:len(validToken)-5] + "XXXXX"

	_, err = VerifyJWT(tamperedToken)
	assert.Error(t, err, "A tampered signature MUST be rejected")
}
