package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"suporter-backend/internal/domain"
)

var ErrDonationNotFound = errors.New("donation not found")

type DonationRepository interface {
	Create(ctx context.Context, d *domain.Donation) error
	FindPendingByTotalAmount(ctx context.Context, streamerID uint64, totalAmount int64) (*domain.Donation, error)
	IsUniqueCodePending(ctx context.Context, streamerID uint64, code int) (bool, error)
	Update(ctx context.Context, d *domain.Donation) error
}

type gormDonationRepository struct {
	db *gorm.DB
}

func NewDonationRepository(db *gorm.DB) DonationRepository {
	return &gormDonationRepository{db: db}
}

func (r *gormDonationRepository) Create(ctx context.Context, d *domain.Donation) error {
	err := r.db.WithContext(ctx).Create(d).Error
	if err != nil {
		return fmt.Errorf("error creating donation record: %w", err)
	}
	return nil
}

func (r *gormDonationRepository) FindPendingByTotalAmount(ctx context.Context, streamerID uint64, totalAmount int64) (*domain.Donation, error) {
	var d domain.Donation
	// Find oldest pending matching donation
	err := r.db.WithContext(ctx).
		Where("streamer_id = ? AND total_amount = ? AND status = ?", streamerID, totalAmount, "pending").
		Order("created_at asc").
		First(&d).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDonationNotFound
		}
		return nil, fmt.Errorf("error finding pending donation: %w", err)
	}
	return &d, nil
}

func (r *gormDonationRepository) IsUniqueCodePending(ctx context.Context, streamerID uint64, code int) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&domain.Donation{}).
		Where("streamer_id = ? AND unique_code = ? AND status = ?", streamerID, code, "pending").
		Count(&count).Error
	if err != nil {
		return false, fmt.Errorf("error checking pending unique code: %w", err)
	}
	return count > 0, nil
}

func (r *gormDonationRepository) Update(ctx context.Context, d *domain.Donation) error {
	err := r.db.WithContext(ctx).Save(d).Error
	if err != nil {
		return fmt.Errorf("error updating donation record: %w", err)
	}
	return nil
}
