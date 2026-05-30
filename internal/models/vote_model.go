package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteModel struct {
	gorm.Model
	ID       uuid.UUID   `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID   uuid.UUID   `gorm:"column:user_id;index;" json:"user_id"`
	User     UserModel   `gorm:"foreignKey:UserID" json:"user"`
	OptionID uuid.UUID   `gorm:"column:option_id;uniqueIndex:idx_user_option" json:"option_id"`
	Option   OptionModel `gorm:"foreignKey:OptionID" json:"option"`
	PollID   uuid.UUID   `gorm:"column:poll_id;index;" json:"poll_id"`
	Poll     PollModel   `gorm:"foreignKey:PollID" json:"poll"`
}
