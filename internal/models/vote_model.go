package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VoteModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	UserID    uuid.UUID      `gorm:"column:user_id;uniqueIndex:idx_user_option;index;" json:"user_id"`
	User      UserModel      `gorm:"foreignKey:UserID" json:"user"`
	OptionID  uuid.UUID      `gorm:"column:option_id;uniqueIndex:idx_user_option" json:"option_id"`
	Option    OptionModel    `gorm:"foreignKey:OptionID" json:"option"`
	PollID    uuid.UUID      `gorm:"column:poll_id;index;" json:"poll_id"`
	Poll      PollModel      `gorm:"foreignKey:PollID" json:"poll"`
}
