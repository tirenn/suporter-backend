package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"suporter-backend/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")
var ErrEmailAlreadyExists = errors.New("email already registered")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
}

type gormUserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &gormUserRepository{db: db}
}

func (r *gormUserRepository) Create(ctx context.Context, u *domain.User) error {
	err := r.db.WithContext(ctx).Create(u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) || contains(err.Error(), "unique") || contains(err.Error(), "duplicate") {
			return ErrEmailAlreadyExists
		}
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (r *gormUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user by email: %w", err)
	}
	return &u, nil
}

func (r *gormUserRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user by id: %w", err)
	}
	return &u, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && s != "" && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
