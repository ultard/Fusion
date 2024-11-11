package utils

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// ParseRouteID извлекает и проверяет параметр id как UUID.
func ParseRouteID(c *fiber.Ctx) (uuid.UUID, error) {
	id := c.Params("id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid product ID")
	}
	return parsedId, nil
}
