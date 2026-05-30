package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/models"
)

type VoteRepository interface {
	Create(ctx context.Context, vote *models.VoteModel) error
	FindByUserAndPoll(ctx context.Context, userID, pollID uuid.UUID) ([]*models.VoteModel, error)
}

type GormVoteRepository struct {
	db *gorm.DB
}

func (r *GormVoteRepository) Create(ctx context.Context, vote *models.VoteModel) error {
	if err := r.db.WithContext(ctx).Create(&vote).Error; err != nil {
		return fmt.Errorf("create vote: %w", err)
	}
	return nil
}

func (r *GormVoteRepository) FindByUserAndPoll(ctx context.Context, userID, pollID uuid.UUID) ([]*models.VoteModel, error) {
	var data []*models.VoteModel
	if err := r.db.WithContext(ctx).Where("user_id = ? AND poll_id = ?", userID, pollID).Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find votes for user %s poll %s: %w", userID, pollID, err)
	}
	return data, nil
}

func NewGormVoteRepository(db *gorm.DB) VoteRepository {
	return &GormVoteRepository{db}
}
