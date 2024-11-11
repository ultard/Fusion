package models

import (
	"github.com/google/uuid"
	"time"
)

type Verification struct {
	ID    uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Type  string    `gorm:"type:varchar(50);not null"`
	Token string    `gorm:"uniqueIndex;not null"`

	UserID uuid.UUID `gorm:"type:uuid;not null"`
	User   User

	ExpiresAt time.Time
	CreatedAt time.Time
}
