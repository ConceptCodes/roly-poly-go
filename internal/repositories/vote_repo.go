package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/models"
)

type VoteRepository interface {
	Create(vote *models.VoteModel) error
	FindByUserAndPoll(userID, pollID uuid.UUID) ([]*models.VoteModel, error)
}

type GormVoteRepository struct {
	db *gorm.DB
}

func (r *GormVoteRepository) Create(vote *models.VoteModel) error {
	if err := r.db.Create(&vote).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormVoteRepository) FindByUserAndPoll(userID, pollID uuid.UUID) ([]*models.VoteModel, error) {
	var data []*models.VoteModel
	if err := r.db.Where("user_id = ? AND poll_id = ?", userID, pollID).Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func NewGormVoteRepository(db *gorm.DB) VoteRepository {
	return &GormVoteRepository{db}
}
