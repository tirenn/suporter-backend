package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/middleware"
	"suporter-backend/internal/service"
)

type AuthHandler struct {
	authService        service.AuthService
	recaptchaSecretKey string
}

func NewAuthHandler(authService service.AuthService, recaptchaSecretKey string) *AuthHandler {
	return &AuthHandler{
		authService:        authService,
		recaptchaSecretKey: recaptchaSecretKey,
	}
}

// Register godoc
// @Summary Register a new streamer account
// @Description Register streamer with name, username, and password. Requires Google reCAPTCHA v2 token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "User Registration Details"
// @Success 201 {object} domain.AuthResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Verify Google reCAPTCHA v2 token
	if err := middleware.VerifyRecaptcha(h.recaptchaSecretKey, req.RecaptchaToken, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verifikasi CAPTCHA gagal: " + err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login streamer account
// @Description Authenticate username & password. Requires Google reCAPTCHA v2 token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "User Credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 401 {object} map[string]string "Invalid username or password"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Verify Google reCAPTCHA v2 token
	if err := middleware.VerifyRecaptcha(h.recaptchaSecretKey, req.RecaptchaToken, c.ClientIP()); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Verifikasi CAPTCHA gagal: " + err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// MobileLogin godoc
// @Summary Mobile Streamer Login (No reCAPTCHA)
// @Description Direct login endpoint for native mobile clients (Android) using username and password.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.MobileLoginRequest true "Mobile User Credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/auth/mobile-login [post]
func (h *AuthHandler) MobileLogin(c *gin.Context) {
	var req domain.MobileLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.authService.MobileLogin(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetStreamerPublicProfile godoc
// @Summary Get a streamer's public profile (name + QRIS URL)
// @Description Returns streamer's display name and QRIS QR image URL for the donation page.
// @Tags Streamers
// @Produce json
// @Param username path string true "Streamer Username"
// @Success 200 {object} domain.StreamerPublicProfile
// @Failure 404 {object} map[string]string "Streamer not found"
// @Router /api/v1/streamers/{username} [get]
func (h *AuthHandler) GetStreamerPublicProfile(c *gin.Context) {
	username := c.Param("username")
	profile, err := h.authService.GetPublicProfile(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, profile)
}

// UpdateProfile godoc
// @Summary Update streamer's QRIS URL (requires auth)
// @Description Allows a logged-in streamer to update their QRIS QR code image URL.
// @Tags Authentication
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param request body domain.UpdateProfileRequest true "Profile Update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/profile [put]
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	if err := h.authService.UpdateQRISUrl(c.Request.Context(), userID, req.QRISUrl); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "QRIS URL updated successfully", "qris_url": req.QRISUrl})
}

// RegenerateWebhookKey godoc
// @Summary Regenerate user's static webhook key (requires auth)
// @Description Invalidates the old static webhook key and generates a new secure wk_... key.
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/profile/webhook-key [put]
func (h *AuthHandler) RegenerateWebhookKey(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.authService.RegenerateWebhookKey(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to regenerate webhook key: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetProfile godoc
// @Summary Get streamer's current profile (requires auth)
// @Description Returns the authenticated streamer's latest profile (including is_active, webhook_key, etc.).
// @Tags Authentication
// @Security BearerAuth
// @Produce json
// @Success 200 {object} domain.User
// @Failure 401 {object} map[string]string "Unauthorized"
// @Router /api/v1/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.authService.GetProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch profile: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}
