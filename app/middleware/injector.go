package middleware

import (
	"fusion/app/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type AppServices struct {
	Config utils.AppConfig
	DB     *gorm.DB
	JWT    utils.JWTService
	Email  utils.EmailService
}

// InjectorMiddleware проверяет токен доступа пользователя
func InjectorMiddleware(config utils.AppConfig, db *gorm.DB, jwt utils.JWTService, email utils.EmailService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals("services", AppServices{
			Config: config,
			DB:     db,
			JWT:    jwt,
			Email:  email,
		})

		return c.Next()
	}
}
