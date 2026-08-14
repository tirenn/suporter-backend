package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error)
	Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error)
	MobileLogin(ctx context.Context, username, password string) (*domain.AuthResponse, error)
	ValidateToken(tokenString string) (*domain.JWTClaims, error)
	GetPublicProfile(ctx context.Context, username string) (*domain.StreamerPublicProfile, error)
	UpdateQRISUrl(ctx context.Context, userID uint64, qrisUrl string) error
	IsUserActive(ctx context.Context, userID uint64) (bool, error)
	GetProfile(ctx context.Context, userID uint64) (*domain.User, error)
	RegenerateWebhookKey(ctx context.Context, userID uint64) (*domain.User, error)
}

type authService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *authService) Register(ctx context.Context, req domain.RegisterRequest) (*domain.AuthResponse, error) {
	name := strings.TrimSpace(req.Name)
	if len(name) < 3 {
		return nil, fmt.Errorf("full name must be at least 3 characters")
	}

	username := strings.ToLower(strings.TrimSpace(req.Username))
	if len(username) < 3 {
		return nil, fmt.Errorf("username must be at least 3 characters")
	}

	role := strings.ToLower(strings.TrimSpace(req.Role))
	if role != "streamer" {
		role = "streamer"
	}

	// Check if username already exists
	existingUser, err := s.userRepo.FindByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("username is already taken")
	}

	if len(req.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}
	var hasUpper, hasNumber, hasSymbol bool
	for _, r := range req.Password {
		if r >= 'A' && r <= 'Z' {
			hasUpper = true
		} else if r >= '0' && r <= '9' {
			hasNumber = true
		} else if strings.ContainsRune("!@#$%^&*(),.?\":{}|<>_+-=[]\\/~`", r) {
			hasSymbol = true
		}
	}
	if !hasUpper || !hasNumber || !hasSymbol {
		return nil, fmt.Errorf("password must contain at least one uppercase letter, one number, and one symbol")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var webhookKey string
	bytesKey := make([]byte, 16)
	if _, err := rand.Read(bytesKey); err == nil {
		webhookKey = "wk_" + hex.EncodeToString(bytesKey)
	} else {
		webhookKey = "wk_" + uuid.New().String()
	}

	var webhookSecret string
	bytesSecret := make([]byte, 24)
	if _, err := rand.Read(bytesSecret); err == nil {
		webhookSecret = "whsec_" + hex.EncodeToString(bytesSecret)
	} else {
		webhookSecret = "whsec_" + uuid.New().String()
	}

	u := &domain.User{
		Name:          strings.TrimSpace(req.Name),
		Username:      username,
		PasswordHash:  string(hashedPassword),
		Role:          role,
		WebhookKey:    webhookKey,
		WebhookSecret: webhookSecret,
	}

	if err := s.userRepo.Create(ctx, u); err != nil {
		return nil, err
	}

	token, err := s.generateJWT(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken: token,
		User:        *u,
	}, nil
}

func (s *authService) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
	username := strings.ToLower(strings.TrimSpace(req.Username))
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	if !u.IsActive {
		return nil, errors.New("akun Anda belum aktif, silakan hubungi admin")
	}

	token, err := s.generateJWT(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken: token,
		User:        *u,
	}, nil
}

func (s *authService) ValidateToken(tokenString string) (*domain.JWTClaims, error) {
	cleanToken := tokenString
	if strings.HasPrefix(tokenString, "Bearer ") {
		cleanToken = strings.TrimPrefix(tokenString, "Bearer ")
	}

	token, err := jwt.ParseWithClaims(cleanToken, &domain.JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*domain.JWTClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// GetPublicProfile returns a safe subset of a streamer's data for unauthenticated donors.
func (s *authService) GetPublicProfile(ctx context.Context, username string) (*domain.StreamerPublicProfile, error) {
	u, err := s.userRepo.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
	if err != nil {
		return nil, fmt.Errorf("streamer not found")
	}
	if u.Role != "streamer" {
		return nil, fmt.Errorf("streamer not found")
	}
	if !u.IsActive {
		return nil, errors.New("streamer sedang tidak aktif")
	}
	return &domain.StreamerPublicProfile{
		Name:     u.Name,
		Username: u.Username,
		QRISUrl:  u.QRISUrl,
	}, nil
}

// UpdateQRISUrl updates the QRIS image URL for a streamer.
func (s *authService) UpdateQRISUrl(ctx context.Context, userID uint64, qrisUrl string) error {
	return s.userRepo.UpdateQRISUrl(ctx, userID, qrisUrl)
}

func (s *authService) generateJWT(user *domain.User) (string, error) {
	claims := &domain.JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *authService) IsUserActive(ctx context.Context, userID uint64) (bool, error) {
	u, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return false, err
	}
	return u.IsActive, nil
}

func (s *authService) MobileLogin(ctx context.Context, username, password string) (*domain.AuthResponse, error) {
	username = strings.ToLower(strings.TrimSpace(username))
	u, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("username atau password salah")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("username atau password salah")
	}

	if !u.IsActive {
		return nil, errors.New("akun Anda belum aktif, silakan hubungi admin")
	}

	token, err := s.generateJWT(u)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &domain.AuthResponse{
		AccessToken: token,
		User:        *u,
	}, nil
}

func (s *authService) RegenerateWebhookKey(ctx context.Context, userID uint64) (*domain.User, error) {
	var webhookKey string
	bytesKey := make([]byte, 16)
	if _, err := rand.Read(bytesKey); err == nil {
		webhookKey = "wk_" + hex.EncodeToString(bytesKey)
	} else {
		webhookKey = "wk_" + uuid.New().String()
	}

	var webhookSecret string
	bytesSecret := make([]byte, 24)
	if _, err := rand.Read(bytesSecret); err == nil {
		webhookSecret = "whsec_" + hex.EncodeToString(bytesSecret)
	} else {
		webhookSecret = "whsec_" + uuid.New().String()
	}

	if err := s.userRepo.UpdateWebhookCredentials(ctx, userID, webhookKey, webhookSecret); err != nil {
		return nil, err
	}

	// Fetch updated user profile
	return s.userRepo.FindByID(ctx, userID)
}

func (s *authService) GetProfile(ctx context.Context, userID uint64) (*domain.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}
