package repository

import (
	"gorm.io/gorm"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

type UserRepository interface {
	FindByApiKey(apiKey string) (*models.UserModel, error)
	Create(user *models.UserModel) error
}

type GormUserRepository struct {
	db *gorm.DB
}

func (r *GormUserRepository) FindByApiKey(apiKey string) (*models.UserModel, error) {
	var data models.UserModel
	if err := r.db.Where(constants.FindByApiKeyQuery, apiKey).First(&data).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *GormUserRepository) Create(user *models.UserModel) error {
	if err := r.db.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db}
}
