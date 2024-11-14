package models

import (
	"github.com/google/uuid"
	"time"
)

type Cart struct {
	ID       uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Products []CartProduct
	UserID   uuid.UUID `gorm:"type:uuid;not null"`
	User     User

	CreatedAt time.Time
	UpdatedAt time.Time
}

type CartProduct struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	CartID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Product   Product
	Quantity  int `gorm:"not null,default:1"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
