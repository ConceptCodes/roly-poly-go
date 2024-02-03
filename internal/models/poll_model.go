package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/constants"
)

type PollModel struct {
	gorm.Model
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Title       string    `gorm:"not null" json:"title"`
	Description string    `gorm:"type:varchar(100)" json:"description"`
	Closed      time.Time `gorm:"" json:"closed"`
	Public      bool      `gorm:"default:true" json:"public"`
	UserID      uuid.UUID `gorm:"column:user_id;index;" json:"user_id"`
	User        UserModel `gorm:"foreignKey:UserID" json:"user"`
}

func (PollModel) TableName() string {
	return fmt.Sprintf(constants.DBTablePrefix, "polls")
}

func (poll *PollModel) Simple() *PollModel {
	return &PollModel{}
}
