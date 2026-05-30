package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

type PollRepository interface {
	FindAll(userID uuid.UUID, publicOnly bool) ([]*models.PollModel, error)
	FindByID(id uuid.UUID) (*models.PollModel, error)
	Create(poll *models.PollModel) error
	CreateWithOptions(poll *models.PollModel, options []*models.OptionModel) error
	Update(poll *models.PollModel) error
	Delete(poll *models.PollModel) error
	OwnsPoll(userId, pollId uuid.UUID) (bool, error)
}

type GormPollRepository struct {
	db *gorm.DB
}

func (r *GormPollRepository) FindAll(userID uuid.UUID, publicOnly bool) ([]*models.PollModel, error) {
	var data []*models.PollModel

	query := r.db.Preload("User").Preload("Options")

	if publicOnly {
		query = query.Where("public = ?", true)
	} else {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (r *GormPollRepository) FindByID(id uuid.UUID) (*models.PollModel, error) {
	var data models.PollModel
	if err := r.db.Preload("User").Preload("Options").First(&data, id).Error; err != nil {
		return nil, err
	}
	return &data, nil
}

func (r *GormPollRepository) Create(poll *models.PollModel) error {
	if err := r.db.Create(&poll).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormPollRepository) CreateWithOptions(poll *models.PollModel, options []*models.OptionModel) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&poll).Error; err != nil {
			return err
		}
		for _, option := range options {
			option.PollID = poll.ID
		}
		if err := tx.Create(&options).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *GormPollRepository) Update(poll *models.PollModel) error {
	if err := r.db.Model(&models.PollModel{}).Where("id = ?", poll.ID).Updates(poll).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormPollRepository) Delete(poll *models.PollModel) error {
	if err := r.db.Delete(&poll).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormPollRepository) CastVotes(pollID, userID uuid.UUID, optionIDs []uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, optionID := range optionIDs {
			vote := &models.VoteModel{
				UserID:   userID,
				OptionID: optionID,
				PollID:   pollID,
			}
			if err := tx.Create(&vote).Error; err != nil {
				return err
			}
			if err := tx.Model(&models.OptionModel{}).Where("id = ?", optionID).
				UpdateColumn("votes", gorm.Expr("votes + 1")).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *GormPollRepository) DeleteWithOptions(pollID uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("poll_id = ?", pollID).Delete(&models.VoteModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("poll_id = ?", pollID).Delete(&models.OptionModel{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ?", pollID).Delete(&models.PollModel{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *GormPollRepository) OwnsPoll(userId, pollId uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.PollModel{}).Where(constants.FindByUserIdAndIdQuery, userId, pollId).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func NewGormPollRepository(db *gorm.DB) PollRepository {
	return &GormPollRepository{db}
}
