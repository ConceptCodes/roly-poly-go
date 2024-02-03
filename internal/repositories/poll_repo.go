package repository

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/constants"
	"roly-poly/internal/models"
)

type PollRepository interface {
	FindAll() ([]*models.PollModel, error)
	FindByID(id uuid.UUID) (*models.PollModel, error)
	Create(poll *models.PollModel) error
	Update(poll *models.PollModel) error
	Delete(poll *models.PollModel) error
	OwnsPoll(userId, pollId uuid.UUID) (bool, error)
}

type GormPollRepository struct {
	db *gorm.DB
}

func (r *GormPollRepository) FindAll() ([]*models.PollModel, error) {
	var data []*models.PollModel
	if err := r.db.Preload("User").Find(&data).Error; err != nil {
		return nil, err
	}
	return data, nil
}

func (r *GormPollRepository) FindByID(id uuid.UUID) (*models.PollModel, error) {
	var data models.PollModel
	if err := r.db.Preload("User").First(&data, id).Error; err != nil {
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

func (r *GormPollRepository) Update(poll *models.PollModel) error {
	if err := r.db.Save(&poll).Error; err != nil {
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

func (r *GormPollRepository) OwnsPoll(userId, pollId uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.Model(&models.PollModel{}).Where(constants.FindByUserIdAndIdQuery, pollId, userId).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func NewGormPollRepository(db *gorm.DB) PollRepository {
	return &GormPollRepository{db}
}
