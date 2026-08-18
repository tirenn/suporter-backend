package service_test

import (
	"context"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/repository"
	"suporter-backend/internal/service"
)

func TestAuthService_RegisterAndLogin(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %s", err)
	}
	defer db.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm database: %s", err)
	}

	userRepo := repository.NewUserRepository(gormDB)
	cfg := &config.Config{
		JWTSecret: "test-secret-key-12345",
	}
	authServ := service.NewAuthService(userRepo, cfg)

	ctx := context.Background()
	regReq := domain.RegisterRequest{
		Name:     "Test User",
		Username: "testuser",
		Password: "Password123!",
		Role:     "streamer",
	}

	// 1. Mock Check Existing Username (Return ErrRecordNotFound)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("testuser", 1).
		WillReturnError(gorm.ErrRecordNotFound)

	// 2. Mock Insert User (Return ID = 1)
	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	regRes, err := authServ.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Register error: %v", err)
	}

	if regRes.AccessToken == "" {
		t.Errorf("expected access token to be generated")
	}

	if regRes.User.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", regRes.User.Username)
	}

	if regRes.User.Role != "streamer" {
		t.Errorf("expected role 'streamer', got %s", regRes.User.Role)
	}

	// Test Login
	loginReq := domain.LoginRequest{
		Username: "testuser",
		Password: "Password123!",
	}

	// Mock FindByUsername for Login
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE username = $1 ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("testuser", 1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "name", "role", "webhook_key", "is_active"}).
			AddRow(1, "testuser", regRes.User.PasswordHash, "Test User", "streamer", regRes.User.WebhookKey, true))

	loginRes, err := authServ.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login error: %v", err)
	}

	if loginRes.AccessToken == "" {
		t.Errorf("expected login access token to be generated")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}
