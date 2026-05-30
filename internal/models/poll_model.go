package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PollModel struct {
	gorm.Model
	ID          uuid.UUID     `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Title       string        `gorm:"not null" json:"title"`
	Description string        `gorm:"type:varchar(100)" json:"description"`
	Closed      time.Time     `gorm:"" json:"closed"`
	Public      bool          `gorm:"default:true" json:"public"`
	UserID      uuid.UUID     `gorm:"column:user_id;index;" json:"user_id"`
	User        UserModel     `gorm:"foreignKey:UserID" json:"user"`
	Options     []OptionModel `gorm:"foreignKey:PollID" json:"options"`
}
