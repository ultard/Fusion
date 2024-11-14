package models

import (
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

// User представляет пользователя системы
type User struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Avatar          *string   `gorm:"type:varchar(255);"`
	Email           string    `gorm:"uniqueIndex;not null"`
	Phone           *string
	Password        string `gorm:"not null"`
	Username        string `gorm:"uniqueIndex;not nul"`
	IsEmailVerified bool   `gorm:"default:false"`

	Verifications []Verification
	Permissions   []Permissions `gorm:"many2many:user_permissions"`
	Products      []Product
	Favourites    []Favourite
	Sessions      []Session
	Orders        []Order
	Cart          *Cart

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (u *User) HasPermissions(permissions ...string) bool {
	for _, requiredPermission := range permissions {
		for _, permission := range u.Permissions {
			if permission.Name == requiredPermission {
				return true
			}
		}
	}

	return false
}

func formatPhoneNumber(phone string) (string, error) {
	re := regexp.MustCompile(`[^0-9+]`)
	normalizedPhone := re.ReplaceAllString(phone, "")

	if !regexp.MustCompile(`^\+?[1-9]\d{1,14}$`).MatchString(normalizedPhone) {
		return "", fmt.Errorf("invalid phone number format")
	}

	if !strings.HasPrefix(normalizedPhone, "+") {
		normalizedPhone = "+" + normalizedPhone
	}

	return normalizedPhone, nil
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.Phone != nil {
		formattedPhone, err := formatPhoneNumber(*u.Phone)
		if err != nil {
			return err
		}
		u.Phone = &formattedPhone
	}
	return nil
}

func (u *User) BeforeUpdate(tx *gorm.DB) error {
	if u.Phone != nil {
		formattedPhone, err := formatPhoneNumber(*u.Phone)
		if err != nil {
			return err
		}
		u.Phone = &formattedPhone
	}
	return nil
}
