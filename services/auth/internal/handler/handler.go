package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/auth/internal/service"
	"ecommerce/services/auth/internal/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     string `json:"role" binding:"required"`
}

type AuthHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) AuthHandler {
	return AuthHandler{service: service}
}

func (h *AuthHandler) RegisterNormal(c *gin.Context) {

	var requestData RegisterRequest
	err := c.ShouldBindJSON(&requestData)

	if err != nil {
		logger.Error("handler: failed to bind request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		logger.Error("handler: failed to register user", zap.Error(err))

		if strings.Contains(err.Error(), "service: email already exists") {
			c.JSON(http.StatusContinue, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	jwt, err := utils.GetJWT(user.ID, user.Email, user.Role)

	if err != nil {
		logger.Error("handler: failed to generate JWT", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"jwt":     jwt,
		"message": "User created successfully",
	})
	logger.Info("handler: successfully registered user", zap.String("id", user.ID))
}

func (h *AuthHandler) RegisterOAUTH(c *gin.Context) {}
