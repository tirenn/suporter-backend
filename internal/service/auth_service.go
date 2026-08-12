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
	ValidateToken(tokenString string) (*domain.JWTClaims, error)
	GetPublicProfile(ctx context.Context, username string) (*domain.StreamerPublicProfile, error)
	UpdateQRISUrl(ctx context.Context, userID uint64, qrisUrl string) error
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	var webhookKey string
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err == nil {
		webhookKey = "wk_" + hex.EncodeToString(bytes)
	} else {
		webhookKey = "wk_" + uuid.New().String()
	}

	u := &domain.User{
		Name:         strings.TrimSpace(req.Name),
		Username:     username,
		PasswordHash: string(hashedPassword),
		Role:         role,
		WebhookKey:   webhookKey,
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
