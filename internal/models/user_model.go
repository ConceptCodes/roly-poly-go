package models

import (
	"fmt"
	"roly-poly/internal/constants"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserModel struct {
	gorm.Model
	ID        uuid.UUID `gorm:"type:uuid;primary_key;unique_index;default:uuid_generate_v4()" json:"id"`
	ApiKey    uuid.UUID `gorm:"type:uuid;primary_key;unique_index;default:uuid_generate_v4()" json:"api_key"`
	Enabled   bool      `gorm:"default:true" json:"enabled"`
	FirstName string    `gorm:"not null" json:"first_name"`
	LastName  string    `gorm:"not null" json:"last_name"`
}

func (user *UserModel) Simple() *UserModel {
	return &UserModel{
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Enabled:   user.Enabled,
		ApiKey:    user.ApiKey,
	}
}

func (UserModel) TableName() string {
	return fmt.Sprintf(constants.DBTablePrefix, "users")
}
