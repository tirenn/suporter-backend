package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"suporter-backend/internal/domain"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	FindByUUID(ctx context.Context, uuid string) (*domain.Project, error)
	FindByUserID(ctx context.Context, userID uint64) ([]*domain.Project, error)
	Update(ctx context.Context, project *domain.Project) error
	Delete(ctx context.Context, uuid string) error
}

type gormProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) ProjectRepository {
	return &gormProjectRepository{db: db}
}

func (r *gormProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	err := r.db.WithContext(ctx).Create(p).Error
	if err != nil {
		return fmt.Errorf("error creating project: %w", err)
	}
	return nil
}

func (r *gormProjectRepository) FindByUUID(ctx context.Context, uuid string) (*domain.Project, error) {
	var p domain.Project
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProjectNotFound
		}
		return nil, fmt.Errorf("error finding project by uuid: %w", err)
	}
	return &p, nil
}

func (r *gormProjectRepository) FindByUserID(ctx context.Context, userID uint64) ([]*domain.Project, error) {
	var projects []*domain.Project
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at desc").Find(&projects).Error
	if err != nil {
		return nil, fmt.Errorf("error finding projects by user_id: %w", err)
	}
	return projects, nil
}

func (r *gormProjectRepository) Update(ctx context.Context, p *domain.Project) error {
	err := r.db.WithContext(ctx).Save(p).Error
	if err != nil {
		return fmt.Errorf("error updating project: %w", err)
	}
	return nil
}

func (r *gormProjectRepository) Delete(ctx context.Context, uuid string) error {
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).Delete(&domain.Project{}).Error
	if err != nil {
		return fmt.Errorf("error deleting project: %w", err)
	}
	return nil
}
