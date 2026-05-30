package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OptionModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Label     string         `gorm:"not null;uniqueIndex:idx_poll_option_label" json:"label"`
	PollID    uuid.UUID      `gorm:"column:poll_id;uniqueIndex:idx_poll_option_label;index;" json:"poll_id"`
	Poll      PollModel      `gorm:"foreignKey:PollID" json:"poll"`
	Votes     uint           `gorm:"default:0" json:"votes"`
}
