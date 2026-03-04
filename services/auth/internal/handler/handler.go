package handler

import (
	"ecommerce/pkg/logger"
	"ecommerce/services/auth/internal/client"
	"ecommerce/services/auth/internal/service"
	"ecommerce/services/auth/internal/utils"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sixafter/nanoid"
	"go.uber.org/zap"
)

type RegisterRequest struct {
	Name     string `json:"name" binding:"required" example:"Pranay Kamble"`
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
	Role     string `json:"role" binding:"required" example:"buyer"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=8" example:"SecurePass123!"`
}

type VerifyRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
	OTP   string `json:"otp" binding:"required,len=6,numeric" example:"123456"`
}

type AuthHandler struct {
	service     service.AuthService
	emailClient client.EmailClient
}

type ResendOTPRequest struct {
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}

type OAuthProfile struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func NewAuthHandler(service service.AuthService, emailClient client.EmailClient) AuthHandler {
	return AuthHandler{
		service:     service,
		emailClient: emailClient,
	}
}

// RegisterNormal godoc
// @Summary      Register a new user
// @Description  Creates a new user account (buyer, seller, or logistic) and triggers an async email verification OTP.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      RegisterRequest  true  "User Registration Details"
// @Success      200      {object}  map[string]interface{} "registered successfully, please verify email now"
// @Failure      400      {object}  map[string]interface{} "incorrect request body or invalid role"
// @Failure      409      {object}  map[string]interface{} "email already exists"
// @Failure      500      {object}  map[string]interface{} "internal server error"
// @Router       /register [post]
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

	c.JSON(http.StatusOK, gin.H{"message": "registered successfully, please verify email now"})
}

// GetPing godoc
// @Summary      Health check the server
// @Description  Use this to check if server is active and running.
// @Tags         System
// @Produce      json
// @Success      200      {object}  map[string]interface{} "pong"
// @Router       /ping [get]
func (h *AuthHandler) GetPing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"msg": "pong"})
}

// Login godoc
// @Summary      Login an existing user
// @Description  Validates credentials and returns a JWT in the JSON body and a Refresh Token in an HttpOnly cookie.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      LoginRequest  true  "User Login Credentials"
// @Success      200      {object}  map[string]interface{} "JWT and success message"
// @Failure      400      {object}  map[string]interface{} "Invalid request body"
// @Failure      401      {object}  map[string]interface{} "Invalid email/password or unverified email"
// @Failure      500      {object}  map[string]interface{} "Internal server error"
// @Router       /login [post]
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

// Refresh godoc
// @Summary      Refresh Access Token
// @Description  Rotates the refresh token securely. Requires the 'refreshToken' HttpOnly cookie to be present.
// @Tags         Session Management
// @Produce      json
// @Success      200      {object}  map[string]interface{} "New JWT and message"
// @Failure      400      {object}  map[string]interface{} "Invalid or missing refresh token cookie"
// @Failure      401      {object}  map[string]interface{} "Token expired, revoked, or theft detected"
// @Failure      500      {object}  map[string]interface{} "Internal server error"
// @Router       /refresh [post]
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

// Logout godoc
// @Summary      Logout a user
// @Description  Revokes the refresh token in the database and clears the HttpOnly cookie.
// @Tags         Session Management
// @Produce      json
// @Success      200      {object}  map[string]interface{} "logout successful"
// @Failure      500      {object}  map[string]interface{} "Internal server error"
// @Router       /logout [post]
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

// Verify godoc
// @Summary      Verify Email OTP
// @Description  Verifies the 6-digit OTP sent to the user's email and issues login tokens upon success.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      VerifyRequest  true  "Email and OTP"
// @Success      200      {object}  map[string]interface{} "JWT and success message"
// @Failure      400      {object}  map[string]interface{} "Invalid request body"
// @Failure      401      {object}  map[string]interface{} "Invalid OTP"
// @Failure      500      {object}  map[string]interface{} "Internal server error"
// @Router       /verify [post]
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

// ResendOTP godoc
// @Summary      Resend Verification OTP
// @Description  Generates and emails a new 6-digit OTP if the user exists and is unverified.
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request  body      ResendOTPRequest  true  "Email address"
// @Success      200      {object}  map[string]interface{} "OTP sent message"
// @Failure      400      {object}  map[string]interface{} "A valid email is required"
// @Failure      409      {object}  map[string]interface{} "Email already verified"
// @Router       /resend-otp [post]
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

// issueTokensAndRespond is an internal helper, so it does NOT get Swagger annotations.
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

// GoogleLogin godoc
// @Summary      Initiate Google OAuth
// @Description  Redirects the user to the Google Consent screen. Cannot be tested directly in Swagger UI due to CORS/Redirects.
// @Tags         OAuth 2.0
// @Success      307  "Redirects to accounts.google.com"
// @Router       /google/login [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	state := utils.HashUsingSHA256(nanoid.ID(time.Now().String()))
	c.SetCookie("oauthstate", state, 60*5, "/", "", false, true)

	oauthURL := utils.OAuth.AuthCodeURL(state)
	c.Redirect(http.StatusTemporaryRedirect, oauthURL)
}

// GoogleCallback godoc
// @Summary      Google OAuth Callback
// @Description  Handles the redirect from Google, exchanges the code for a profile, and issues login tokens.
// @Tags         OAuth 2.0
// @Param        state query string true "CSRF State Token"
// @Param        code  query string true "Authorization Code"
// @Success      201  {object}  map[string]interface{} "JWT and success message"
// @Failure      400  {object}  map[string]interface{} "Invalid state, cookie, or authorization code"
// @Failure      500  {object}  map[string]interface{} "Internal server error during exchange or parsing"
// @Router       /google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	oauthstate := c.Request.URL.Query().Get("state")

	browserOauthstate, err := c.Cookie("oauthstate")

	if errors.Is(err, http.ErrNoCookie) || oauthstate == "" || oauthstate != browserOauthstate {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid state or cookie missing"})
		return
	}

	authorizationCode := c.Request.URL.Query().Get("code")
	if authorizationCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "authorization code is required"})
		return
	}

	token, err := utils.OAuth.Exchange(c.Request.Context(), authorizationCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid authorization code"})
		return
	}

	oauthClient := utils.OAuth.Client(c.Request.Context(), token)
	response, err := oauthClient.Get("https://www.googleapis.com/oauth2/v2/userinfo")

	if err != nil {
		logger.Error("handler: failed to get user info from Google APIs: ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("handler: failed to close response body cleanly: ", zap.Error(err))
		}
	}(response.Body)

	bodyBytes, err := io.ReadAll(response.Body)

	var profile OAuthProfile
	if err := json.Unmarshal(bodyBytes, &profile); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse google response"})
		return
	}

	user, err := h.service.OAuthLogin(c.Request.Context(), profile.Email, profile.ID, profile.Name)
	if err != nil {
		logger.Error("handler: failed to complete OAuth Login/Register: ", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	h.issueTokensAndRespond(c, user.ID, user.Email, user.Role, "User logged in", http.StatusCreated)
}

// GetPublicKey godoc
// @Summary      Share Public Key
// @Description  Exposes a public endpoint to share the Public Key used for verification.
// @Tags         Authentication
// @Success      200  {object}  map[string]interface{} "Success Public Key"
// @Failure      500  {object}  map[string]interface{} "Internal server error during exchange or parsing"
// @Router       /public-key [get]
func (h *AuthHandler) GetPublicKey(c *gin.Context) {
	publicKey, err := utils.GetPublicKeyString()
	if err != nil {
		logger.Error("handler: failed to get public key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"publicKey": publicKey,
	})
}
