package handlers

import (
	"fusion/app/database/models"
	"fusion/app/middleware"
	"fusion/app/schemas"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UsersRoute struct {
	db       *gorm.DB
	validate *validator.Validate
}

func RegisterUserRoutes(app *fiber.App, db *gorm.DB) {
	handler := UsersRoute{
		db:       db,
		validate: validator.New(),
	}

	userGroup := app.Group("/users")
	meGroup := userGroup.Group("/me")
	meGroup.Use(middleware.AuthMiddleware())
	meGroup.Get("/", handler.getCurrentUser)
	meGroup.Patch("/", handler.updateUser)
	meGroup.Delete("/", handler.deleteUser)

	userGroup.Get("/:id", handler.getUserById)
}

// getCurrentUser возвращает список всех пользователей
func (h UsersRoute) getCurrentUser(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)
	response := schemas.UserMeResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		Avatar:   user.Avatar,
	}

	return c.JSON(response)
}

// getUser возвращает пользователя по ID
func (h UsersRoute) getUserById(c *fiber.Ctx) error {
	id := c.Params("id")
	parsedId, err := uuid.Parse(id)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid product ID")
	}

	var user models.User
	if err := h.db.First(&user, parsedId).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "user not found")
	}

	response := schemas.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Avatar:   user.Avatar,
	}

	return c.JSON(response)
}

func (h UsersRoute) updateUser(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)
	var input schemas.UserUpdateRequest
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "could not parse request body")
	}

	if input.Username != nil {
		user.Username = *input.Username
	}
	if input.Email != nil {
		user.Email = *input.Email
	}
	if input.Avatar != nil {
		user.Avatar = input.Avatar
	}

	if err := h.db.Save(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update user")
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// deleteUser удаляет пользователя по ID
func (h UsersRoute) deleteUser(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)
	if err := h.db.Delete(&user).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete user")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
