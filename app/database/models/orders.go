package models

import (
	"github.com/google/uuid"
	"time"
)

// OrderStatus определяет статус заказа
type OrderStatus int32

const (
	CREATED OrderStatus = iota
	STAGING
	BILLED
	SENT
	DELIVERED
	ACCEPTED
)

type Order struct {
	ID     uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID uuid.UUID   `gorm:"type:uuid;not null"`
	Status OrderStatus `gorm:"type:int;default:0"`

	User     User
	Products []CartProduct `gorm:"many2many:order_products;"`
}

type OrderProduct struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrderID   uuid.UUID `gorm:"type:uuid;primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;primaryKey"`
	Product   Product
	Quantity  int `gorm:"not null,default:1"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
