package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
	"suporter-backend/internal/service"
)

type AuthHandler struct {
	authService service.AuthService
}

func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register godoc
// @Summary Register a new user account
// @Description Register user with name, email, and password. Returns JWT token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.RegisterRequest true "User Registration Details"
// @Success 201 {object} domain.AuthResponse
// @Failure 400 {object} map[string]string "Bad Request"
// @Failure 409 {object} map[string]string "Email already exists"
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.authService.Register(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, repository.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "Email is already registered"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login user account
// @Description Authenticate email & password, returns JWT token.
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "User Credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 401 {object} map[string]string "Invalid email or password"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
