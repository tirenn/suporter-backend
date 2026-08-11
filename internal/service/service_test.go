package service_test

import (
	"context"
	"testing"
	"time"

	"suporter-backend/internal/config"
	"suporter-backend/internal/domain"
	"suporter-backend/internal/service"
)

type mockUserRepo struct {
	users  map[string]*domain.User
	nextID uint64
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:  make(map[string]*domain.User),
		nextID: 1,
	}
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	u.ID = m.nextID
	m.nextID++
	m.users[u.Email] = u
	return nil
}

func (m *mockUserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	if u, ok := m.users[email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func TestAuthService_RegisterAndLogin(t *testing.T) {
	cfg := &config.Config{
		JWTSecret:   "test-secret-key-12345",
		JWTExpiryHr: 24,
	}

	repo := newMockUserRepo()
	authSvc := service.NewAuthService(repo, cfg)

	ctx := context.Background()

	// 1. Test Register
	regReq := domain.RegisterRequest{
		Name:     "Test Streamer",
		Email:    "streamer@test.com",
		Password: "SecurePassword123!",
	}

	resp, err := authSvc.Register(ctx, regReq)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Errorf("Expected non-empty access token")
	}

	if resp.User.Email != "streamer@test.com" {
		t.Errorf("Expected email streamer@test.com, got %s", resp.User.Email)
	}

	// 2. Test Token Validation
	claims, err := authSvc.ValidateToken(resp.AccessToken)
	if err != nil {
		t.Fatalf("Token validation failed: %v", err)
	}

	if claims.UserID != resp.User.ID || claims.Email != "streamer@test.com" {
		t.Errorf("Claims mismatch: %v", claims)
	}

	// 3. Test Login
	loginReq := domain.LoginRequest{
		Email:    "streamer@test.com",
		Password: "SecurePassword123!",
	}

	loginResp, err := authSvc.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if loginResp.AccessToken == "" {
		t.Errorf("Expected non-empty login access token")
	}
}

func TestSSEBroker_ProjectIsolation(t *testing.T) {
	broker := service.NewSSEBroker()

	proj1Chan := broker.Subscribe("prj_111")
	proj2Chan := broker.Subscribe("prj_222")

	defer broker.Unsubscribe("prj_111", proj1Chan)
	defer broker.Unsubscribe("prj_222", proj2Chan)

	alert1 := domain.Alert{
		ID:      1,
		Name:    "Alice",
		Message: "Hello Project 1!",
	}

	broker.Broadcast("prj_111", alert1)

	select {
	case received := <-proj1Chan:
		if received.Name != "Alice" {
			t.Errorf("Expected alert for Alice, got %s", received.Name)
		}
	case <-time.After(100 * time.Millisecond):
		t.Errorf("Timed out waiting for alert on project 1")
	}

	select {
	case unexpected := <-proj2Chan:
		t.Errorf("Project 2 received unexpected alert meant for Project 1: %v", unexpected)
	case <-time.After(50 * time.Millisecond):
		// Success! Project 2 isolated from Project 1.
	}
}
