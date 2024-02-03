package models

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/constants"
)

type VoteModel struct {
	gorm.Model
	ID       uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID   uuid.UUID   `gorm:"column:user_id;index;" json:"user_id"`
	User     UserModel   `gorm:"foreignKey:UserID" json:"user"`
	OptionID uuid.UUID   `gorm:"column:option_id;index;" json:"option_id"`
	Option   OptionModel `gorm:"foreignKey:OptionID" json:"option"`
	PollID   uuid.UUID   `gorm:"column:poll_id;index;" json:"poll_id"`
	Poll     PollModel   `gorm:"foreignKey:PollID" json:"poll"`
}

func (VoteModel) TableName() string {
	return fmt.Sprintf(constants.DBTablePrefix, "votes")
}

func (vote *VoteModel) Simple() *VoteModel {
	return &VoteModel{}
}
