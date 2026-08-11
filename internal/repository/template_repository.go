package repository

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"suporter-backend/internal/domain"
)

var ErrTemplateNotFound = errors.New("template not found")

type TemplateRepository interface {
	Create(ctx context.Context, template *domain.Template) error
	FindByProjectID(ctx context.Context, projectID uint64) ([]*domain.Template, error)
	FindByID(ctx context.Context, id uint64) (*domain.Template, error)
	Update(ctx context.Context, template *domain.Template) error
	Delete(ctx context.Context, id uint64) error
}

type gormTemplateRepository struct {
	db *gorm.DB
}

func NewTemplateRepository(db *gorm.DB) TemplateRepository {
	return &gormTemplateRepository{db: db}
}

func (r *gormTemplateRepository) Create(ctx context.Context, t *domain.Template) error {
	err := r.db.WithContext(ctx).Create(t).Error
	if err != nil {
		return fmt.Errorf("error creating template: %w", err)
	}
	return nil
}

func (r *gormTemplateRepository) FindByProjectID(ctx context.Context, projectID uint64) ([]*domain.Template, error) {
	var templates []*domain.Template
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Order("created_at desc").Find(&templates).Error
	if err != nil {
		return nil, fmt.Errorf("error finding templates: %w", err)
	}
	return templates, nil
}

func (r *gormTemplateRepository) FindByID(ctx context.Context, id uint64) (*domain.Template, error) {
	var t domain.Template
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&t).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTemplateNotFound
		}
		return nil, fmt.Errorf("error finding template by id: %w", err)
	}
	return &t, nil
}

func (r *gormTemplateRepository) Update(ctx context.Context, t *domain.Template) error {
	err := r.db.WithContext(ctx).Save(t).Error
	if err != nil {
		return fmt.Errorf("error updating template: %w", err)
	}
	return nil
}

func (r *gormTemplateRepository) Delete(ctx context.Context, id uint64) error {
	err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&domain.Template{}).Error
	if err != nil {
		return fmt.Errorf("error deleting template: %w", err)
	}
	return nil
}
