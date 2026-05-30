package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

type PollRepository interface {
	FindAll(ctx context.Context, userID uuid.UUID, publicOnly bool) ([]*models.PollModel, error)
	FindByID(ctx context.Context, id uuid.UUID) (*models.PollModel, error)
	FindByIDForUser(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) (*models.PollModel, error)
	Create(ctx context.Context, poll *models.PollModel) error
	CreateWithOptions(ctx context.Context, poll *models.PollModel, options []*models.OptionModel) error
	Update(ctx context.Context, poll *models.PollModel) error
	Delete(ctx context.Context, poll *models.PollModel) error
	DeleteWithOptions(ctx context.Context, pollID uuid.UUID) error
	CastVotes(ctx context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error
	OwnsPoll(ctx context.Context, userId, pollId uuid.UUID) (bool, error)
	CountVotesByUserAndPoll(ctx context.Context, userID, pollID uuid.UUID) (int64, error)
}

type GormPollRepository struct {
	db *gorm.DB
}

func (r *GormPollRepository) FindAll(ctx context.Context, userID uuid.UUID, publicOnly bool) ([]*models.PollModel, error) {
	var data []*models.PollModel

	query := r.db.WithContext(ctx).Preload("User").Preload("Options")

	if publicOnly {
		query = query.Where("public = ?", true)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&data).Error; err != nil {
		return nil, fmt.Errorf("find all polls: %w", err)
	}
	return data, nil
}

func (r *GormPollRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.PollModel, error) {
	var data models.PollModel
	if err := r.db.WithContext(ctx).Preload("User").Preload("Options").First(&data, id).Error; err != nil {
		return nil, fmt.Errorf("find poll by id %s: %w", id, err)
	}
	return &data, nil
}

func (r *GormPollRepository) FindByIDForUser(ctx context.Context, id uuid.UUID, requesterID uuid.UUID) (*models.PollModel, error) {
	var data models.PollModel
	if err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Options").
		Where("id = ? AND (public = ? OR user_id = ?)", id, true, requesterID).
		First(&data).Error; err != nil {
		return nil, fmt.Errorf("find poll %s for user %s: %w", id, requesterID, err)
	}
	return &data, nil
}

func (r *GormPollRepository) Create(ctx context.Context, poll *models.PollModel) error {
	if err := r.db.WithContext(ctx).Create(&poll).Error; err != nil {
		return fmt.Errorf("create poll: %w", err)
	}
	return nil
}

func (r *GormPollRepository) CreateWithOptions(ctx context.Context, poll *models.PollModel, options []*models.OptionModel) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&poll).Error; err != nil {
			return fmt.Errorf("create poll in tx: %w", err)
		}
		for _, option := range options {
			option.PollID = poll.ID
		}
		if err := tx.Create(&options).Error; err != nil {
			return fmt.Errorf("create options in tx: %w", err)
		}
		return nil
	})
}

func (r *GormPollRepository) Update(ctx context.Context, poll *models.PollModel) error {
	if err := r.db.WithContext(ctx).Model(&models.PollModel{}).Where("id = ?", poll.ID).Updates(poll).Error; err != nil {
		return fmt.Errorf("update poll %s: %w", poll.ID, err)
	}
	return nil
}

func (r *GormPollRepository) Delete(ctx context.Context, poll *models.PollModel) error {
	if err := r.db.WithContext(ctx).Delete(&poll).Error; err != nil {
		return fmt.Errorf("delete poll %s: %w", poll.ID, err)
	}
	return nil
}

func (r *GormPollRepository) CastVotes(ctx context.Context, pollID, userID uuid.UUID, optionIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, optionID := range optionIDs {
			var opt models.OptionModel
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				First(&opt, "id = ?", optionID).Error; err != nil {
				return fmt.Errorf("lock option %s: %w", optionID, err)
			}

			vote := &models.VoteModel{
				UserID:   userID,
				OptionID: optionID,
				PollID:   pollID,
			}
			if err := tx.Create(&vote).Error; err != nil {
				return fmt.Errorf("create vote for option %s: %w", optionID, err)
			}

			if err := tx.Model(&models.OptionModel{}).Where("id = ?", optionID).
				UpdateColumn("votes", gorm.Expr("votes + 1")).Error; err != nil {
				return fmt.Errorf("increment vote count for option %s: %w", optionID, err)
			}
		}
		return nil
	})
}

func (r *GormPollRepository) DeleteWithOptions(ctx context.Context, pollID uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Where("poll_id = ?", pollID).Delete(&models.VoteModel{}).Error; err != nil {
			return fmt.Errorf("delete votes for poll %s: %w", pollID, err)
		}
		if err := tx.WithContext(ctx).Where("poll_id = ?", pollID).Delete(&models.OptionModel{}).Error; err != nil {
			return fmt.Errorf("delete options for poll %s: %w", pollID, err)
		}
		if err := tx.WithContext(ctx).Where("id = ?", pollID).Delete(&models.PollModel{}).Error; err != nil {
			return fmt.Errorf("delete poll %s: %w", pollID, err)
		}
		return nil
	})
}

func (r *GormPollRepository) OwnsPoll(ctx context.Context, userId, pollId uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.PollModel{}).
		Where(constants.FindByUserIdAndIdQuery, userId, pollId).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("owns poll check: %w", err)
	}
	return count > 0, nil
}

func (r *GormPollRepository) CountVotesByUserAndPoll(ctx context.Context, userID, pollID uuid.UUID) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.VoteModel{}).
		Where("user_id = ? AND poll_id = ?", userID, pollID).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count votes for user %s poll %s: %w", userID, pollID, err)
	}
	return count, nil
}

func NewGormPollRepository(db *gorm.DB) PollRepository {
	return &GormPollRepository{db}
}
