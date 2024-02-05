package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/models"
)

type OptionRepository interface {
	FindAll() ([]*models.OptionModel, error)
	FindByID(id uuid.UUID) (*models.OptionModel, error)
	CreateMany(options []*models.OptionModel) error
	Update(poll *models.OptionModel) error
	Delete(poll *models.OptionModel) error
}

type GormOptionRepository struct {
	db *gorm.DB
}

func (r *GormOptionRepository) FindAll() ([]*models.OptionModel, error) {
	var data []*models.OptionModel
	if err := r.db.Preload("roly_poly_poll").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (r *GormOptionRepository) FindByID(id uuid.UUID) (*models.OptionModel, error) {
	var data models.OptionModel
	if err := r.db.Preload("roly_poly_poll").First(&data, id).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *GormOptionRepository) CreateMany(options []*models.OptionModel) error {
	if err := r.db.Create(&options).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormOptionRepository) Update(option *models.OptionModel) error {
	if err := r.db.Save(&option).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormOptionRepository) Delete(option *models.OptionModel) error {
	if err := r.db.Delete(&option).Error; err != nil {
		return err
	}
	return nil
}

func NewGormOptionRepository(db *gorm.DB) OptionRepository {
	return &GormOptionRepository{db}
}
