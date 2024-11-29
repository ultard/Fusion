package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type Product struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID      uuid.UUID
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" gorm:"type:decimal(10,2)"`
	Stock       int     `json:"stock"`
	Image       *string
	Categories  []Category `gorm:"many2many:product_category;"`
	Reviews     []Review
	User        User

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Category struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string
	Description string
	Products    []Product `gorm:"many2many:product_category;"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Review struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid;"`
	Rating    float64
	Comment   string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

type Favourite struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	ProductID uuid.UUID `gorm:"type:uuid;index"`
	UserID    uuid.UUID `gorm:"type:uuid;index"`

	User    User
	Product Product
}
