package models

import "github.com/google/uuid"

type Permissions struct {
	ID     uint      `gorm:"primaryKey"`
	Name   string    `gorm:"uniqueIndex;not null"`
	UserID uuid.UUID `gorm:"type:uuid"`
	User   User
}
