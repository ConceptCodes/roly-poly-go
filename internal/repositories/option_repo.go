package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/models"
)

type OptionRepository interface {
	FindAll(ctx context.Context) ([]*models.OptionModel, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.OptionModel, error)
	CreateMany(ctx context.Context, options []*models.OptionModel) error
	Update(ctx context.Context, poll *models.OptionModel) error
	Delete(ctx context.Context, poll *models.OptionModel) error
}

type GormOptionRepository struct {
	db *gorm.DB
}

func (r *GormOptionRepository) FindAll(ctx context.Context) ([]*models.OptionModel, error) {
	var data []*models.OptionModel
	if err := r.db.WithContext(ctx).Preload("Poll").Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find all options: %w", err)
	}
	return data, nil
}

func (r *GormOptionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.OptionModel, error) {
	var data models.OptionModel
	if err := r.db.WithContext(ctx).Preload("Poll").First(&data, id).Error; err != nil {
		return nil, fmt.Errorf("find option by id %s: %w", id, err)
	}
	return &data, nil
}

func (r *GormOptionRepository) CreateMany(ctx context.Context, options []*models.OptionModel) error {
	if err := r.db.WithContext(ctx).Create(&options).Error; err != nil {
		return fmt.Errorf("create options: %w", err)
	}
	return nil
}

func (r *GormOptionRepository) Update(ctx context.Context, option *models.OptionModel) error {
	if err := r.db.WithContext(ctx).Model(&models.OptionModel{}).Where("id = ?", option.ID).Updates(option).Error; err != nil {
		return fmt.Errorf("update option %s: %w", option.ID, err)
	}
	return nil
}

func (r *GormOptionRepository) Delete(ctx context.Context, option *models.OptionModel) error {
	if err := r.db.WithContext(ctx).Delete(&option).Error; err != nil {
		return fmt.Errorf("delete option %s: %w", option.ID, err)
	}
	return nil
}

func NewGormOptionRepository(db *gorm.DB) OptionRepository {
	return &GormOptionRepository{db}
}
