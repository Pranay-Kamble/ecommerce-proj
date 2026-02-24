package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/auth/internal/client"
	"ecommerce/services/auth/internal/service"
	"ecommerce/services/auth/internal/utils"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type VerifyRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6,numeric"`
}
type AuthHandler struct {
	service     service.AuthService
	emailClient client.EmailClient
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func NewAuthHandler(service service.AuthService, emailClient client.EmailClient) AuthHandler {
	return AuthHandler{
		service:     service,
		emailClient: emailClient,
	}
}

func (h *AuthHandler) RegisterNormal(c *gin.Context) {

	var requestData RegisterRequest
	err := c.ShouldBindJSON(&requestData)

	if err != nil {
		logger.Error("handler: failed to bind request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "incorrect request body"})
		return
	}

	if requestData.Role != "buyer" && requestData.Role != "seller" && requestData.Role != "logistic" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid role"})
		return
	}

	user, err := h.service.Register(c.Request.Context(),
		requestData.Name,
		requestData.Email,
		requestData.Password,
		requestData.Role,
		"email",
		"",
	)

	if err != nil {
		if strings.Contains(err.Error(), "service: email already exists") {
			logger.Error("handler: failed to register user (email already exists)")
			c.JSON(http.StatusConflict, gin.H{"error": "email already exists"})
			return
		}

		logger.Error("handler: failed to register user due to internal error", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	otp, err := h.service.CreateOTP(c.Request.Context(), user.Email, time.Minute*10)

	if err != nil {
		logger.Error("handler: failed to create OTP", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	go func(email, otp string) {
		err := h.emailClient.SendVerificationEmail(email, otp)
		if err != nil {
			logger.Error("handler: failed to send verification email", zap.Error(err))
		}
	}(user.Email, otp)

	c.JSON(http.StatusOK, gin.H{
		"message": "registered successfully, please verify email now",
	})
}

func (h *AuthHandler) RegisterOAUTH(c *gin.Context) {}

func (h *AuthHandler) GetPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "pong"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var request LoginRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		logger.Error("handler: failed to bind request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	email := strings.ToLower(request.Email)
	password := request.Password

	userInfo, err := h.service.Login(c.Request.Context(), email, password)
	if err != nil {

		errorString := err.Error()

		if strings.Contains(errorString, "service: email does not exist") || strings.Contains(errorString, "service: invalid password") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}

		if strings.Contains(errorString, "service: user is not verified") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "please verify email"})
			return
		}

		logger.Error("handler: failed to login", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.issueTokensAndRespond(c, userInfo.ID, userInfo.Email, userInfo.Role, "User logged in", http.StatusOK)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")

	if err != nil {
		logger.Error("handler: failed to get refresh token from cookie: ", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid refresh token"})
		return
	}

	newTokenString, tokenUser, err := h.service.RotateRefreshToken(c.Request.Context(), refreshToken)

	if err != nil {
		if strings.Contains(err.Error(), "service: refresh token not found") ||
			strings.Contains(err.Error(), "service: refresh token is expired") ||
			strings.Contains(err.Error(), "service: refresh token already used or revoked") ||
			strings.Contains(err.Error(), "service: user not found") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid refresh token"})
			return
		}

		logger.Error("handler: failed to rotate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	jwt, err := utils.GetJWT(tokenUser.ID, tokenUser.Email, tokenUser.Role)
	if err != nil {
		logger.Error("handler: failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.SetCookie("refreshToken", newTokenString, 60*60*24*7, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"msg": "refresh successful", "jwt": jwt})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	refreshToken, err := c.Cookie("refreshToken")

	if errors.Is(err, http.ErrNoCookie) {
		c.JSON(http.StatusOK, gin.H{"msg": "logout successful"})
		return
	} else if err != nil {
		logger.Error("handler: failed to get refresh token from cookie", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	err = h.service.Logout(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.SetCookie("refreshToken", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"msg": "logout successful"})
}

func (h *AuthHandler) Verify(c *gin.Context) {
	var request VerifyRequest
	err := c.ShouldBindJSON(&request)
	if err != nil {
		logger.Error("handler: failed to bind request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.service.VerifyEmail(c.Request.Context(), request.Email, request.OTP)
	if err != nil {
		if strings.Contains(err.Error(), "service: failed to delete OTP from OTP repository") {
			logger.Error("handler: failed to delete OTP from OTP repository", zap.Error(err))
		} else {
			logger.Error("handler: failed to verify OTP from OTP repository", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
	}

	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid OTP"})
		return
	}

	h.issueTokensAndRespond(c, user.ID, user.Email, user.Role, "User logged in", http.StatusOK)
	logger.Info("handler: successfully registered user", zap.String("id", user.ID))
}

func (h *AuthHandler) ResendOTP(c *gin.Context) {
	var requestData ResendOTPRequest

	if err := c.ShouldBindJSON(&requestData); err != nil {
		logger.Error("handler: failed to bind resend otp request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "A valid email is required"})
		return
	}

	otp, err := h.service.ResendOTP(c.Request.Context(), requestData.Email)
	if err != nil {
		logger.Error("handler: resend OTP blocked", zap.Error(err))
		if strings.Contains(err.Error(), "user is already verified") {
			c.JSON(http.StatusConflict, gin.H{"error": "This email is already verified. Please log in."})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "If email is registered and unverified, a new OTP has been sent.",
		})
		return
	}
	go func(targetEmail, generatedOTP string) {
		err := h.emailClient.SendVerificationEmail(targetEmail, generatedOTP)
		if err != nil {
			logger.Error("handler: failed to send resend verification email", zap.Error(err))
		}
	}(requestData.Email, otp)

	c.JSON(http.StatusOK, gin.H{
		"message": "If email is registered and unverified, a new OTP has been sent.",
	})
}

func (h *AuthHandler) issueTokensAndRespond(c *gin.Context, userID, email, role, successMsg string, statusCode int) {
	jwt, err := utils.GetJWT(userID, email, role)
	if err != nil {
		logger.Error("handler: failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	refreshToken, hashedRefreshToken, familyId, err := utils.GetRefreshTokenString()
	if err != nil {
		logger.Error("handler: failed to generate refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	_, err = h.service.SaveRefreshToken(c.Request.Context(), userID, hashedRefreshToken, familyId)
	if err != nil {
		logger.Error("handler: failed to save refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.SetCookie("refreshToken", refreshToken, 60*60*24*7, "/", "", false, true)

	c.JSON(statusCode, gin.H{
		"jwt": jwt,
		"msg": successMsg,
	})
}
