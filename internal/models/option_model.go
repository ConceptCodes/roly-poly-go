package models

import (
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"roly-poly/internal/constants"
)

type OptionModel struct {
	gorm.Model
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Label  string    `gorm:"not null;unique_index" json:"label"`
	PollID uuid.UUID `gorm:"column:poll_id;index;" json:"poll_id"`
	Poll   PollModel `gorm:"foreignKey:PollID" json:"poll"`
	Votes  uint      `gorm:"default:0" json:"votes"`
}

func (OptionModel) TableName() string {
	return fmt.Sprintf(constants.DBTablePrefix, "options")
}

func (option *OptionModel) Simple() *OptionModel {
	return &OptionModel{
		Label: option.Label,
		Votes: option.Votes,
	}
}
