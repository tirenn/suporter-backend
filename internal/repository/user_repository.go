package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"suporter-backend/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindByWebhookKey(ctx context.Context, key string) (*domain.User, error)
	FindByID(ctx context.Context, id uint64) (*domain.User, error)
	UpdateQRISUrl(ctx context.Context, userID uint64, qrisUrl string) error
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
		return fmt.Errorf("error creating user: %w", err)
	}
	return nil
}

func (r *gormUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user by username: %w", err)
	}
	return &u, nil
}

func (r *gormUserRepository) FindByWebhookKey(ctx context.Context, key string) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).Where("webhook_key = ?", key).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user by webhook key: %w", err)
	}
	return &u, nil
}

func (r *gormUserRepository) FindByID(ctx context.Context, id uint64) (*domain.User, error) {
	var u domain.User
	err := r.db.WithContext(ctx).First(&u, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("error finding user by id: %w", err)
	}
	return &u, nil
}

func (r *gormUserRepository) UpdateQRISUrl(ctx context.Context, userID uint64, qrisUrl string) error {
	result := r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("qris_url", qrisUrl)
	if result.Error != nil {
		return fmt.Errorf("error updating QRIS URL: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}
