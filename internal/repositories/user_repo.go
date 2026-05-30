package repository

import (
	"context"
	"fmt"

	"gorm.io/gorm"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
	"roly-poly/pkg/token"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.UserModel) error
	FindByApiKey(ctx context.Context, apiKey string) (*models.UserModel, error)
}

type GormUserRepository struct {
	db *gorm.DB
}

func (r *GormUserRepository) FindByApiKey(ctx context.Context, apiKey string) (*models.UserModel, error) {
	var data models.UserModel
	if err := r.db.WithContext(ctx).Where(constants.FindByApiKeyQuery, apiKey).First(&data).Error; err != nil {
		return nil, fmt.Errorf("find user by api key: %w", err)
	}
	return &data, nil
}

func (r *GormUserRepository) Create(ctx context.Context, user *models.UserModel) error {
	key, err := token.Generate()
	if err != nil {
		return fmt.Errorf("generate api key: %w", err)
	}
	user.ApiKey = key

	if err := r.db.WithContext(ctx).Create(&user).Error; err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func NewGormUserRepository(db *gorm.DB) UserRepository {
	return &GormUserRepository{db}
}
