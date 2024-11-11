package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Session struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID   uuid.UUID `gorm:"type:uuid;index;not null"`
	Token    string
	Agent    string
	IP       string
	IsActive bool

	IssuedAt  time.Time `gorm:"default:now()"`
	ExpiresAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	User User `gorm:"foreignKey:UserID"`
}
