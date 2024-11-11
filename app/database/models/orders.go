package models

import "github.com/google/uuid"

// OrderStatus определяет статус заказа
type OrderStatus int32

const (
	STAGING OrderStatus = iota
	CREATED
	BILLED
	DELIVERED
)

type Order struct {
	ID     uuid.UUID   `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID uuid.UUID   `gorm:"type:uuid;not null"`
	Status OrderStatus `gorm:"type:int;default:0"`

	User     User
	Products []Product `gorm:"many2many:order_products"`
}
