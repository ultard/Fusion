package middleware

import (
	"fusion/app/database/models"
	"fusion/app/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strings"
)

// AuthMiddleware проверяет токен доступа пользователя
func AuthMiddleware(permissions ...string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		services := c.Locals("services").(AppServices)

		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing Authorization header")
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 {
			return fiber.NewError(fiber.StatusBadRequest, "invalid Authorization header")
		}

		tokenString := parts[1]
		token, err := services.JWT.ValidateToken(tokenString)

		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token")
		}

		claims, ok := token.Claims.(*utils.JwtCustomClaim)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid token claims")
		}

		userID := claims.UserID
		userUUID, err := uuid.Parse(userID)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid user ID format")
		}

		var user models.User
		if err := services.DB.First(&user, userUUID).Error; err != nil {
			return fiber.NewError(fiber.StatusNotFound, "user not found")
		}

		if len(permissions) > 0 {
			roleMatch := user.HasPermissions(permissions...)

			if !roleMatch {
				return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
			}
		}

		c.Locals("current_user", user)
		return c.Next()
	}
}
