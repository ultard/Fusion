package handlers

import (
	"fusion/app/database/models"
	"fusion/app/middleware"
	"fusion/app/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type OrderHandler struct {
	db *gorm.DB
}

// NewOrderHandler создает новый обработчик для заказов
func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{db: db}
}

// RegisterOrderRoutes регистрирует маршруты для заказов
func RegisterOrderRoutes(app *fiber.App, db *gorm.DB) {
	handler := NewOrderHandler(db)

	orderGroup := app.Group("/orders")

	orderGroup.Use(middleware.AuthMiddleware())
	orderGroup.Get("/", handler.GetOrders)
	orderGroup.Post("/", handler.CreateOrder)
	orderGroup.Put("/:id", handler.UpdateOrderStatus)
	orderGroup.Delete("/:id", handler.DeleteOrder)
}

// GetOrders возвращает список заказов для текущего пользователя
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)

	var orders []models.Order
	if err := h.db.Preload("Products").Where("user_id = ?", user.ID).Find(&orders).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve orders")
	}

	return c.JSON(orders)
}

// CreateOrder создает новый заказ или добавляет продукты в существующий заказ в статусе STAGING
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)

	type CreateOrderInput struct {
		ProductIDs []uuid.UUID `json:"product_ids"`
	}

	var input CreateOrderInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	// Проверяем, есть ли существующий заказ в статусе STAGING
	var order models.Order
	if err := h.db.Where("user_id = ? AND status = ?", user.ID, models.STAGING).First(&order).Error; err != nil {
		// Создаем новый заказ, если текущий в статусе STAGING не найден
		order = models.Order{
			ID:     uuid.New(),
			UserID: user.ID,
			Status: models.STAGING,
		}
		if err := h.db.Create(&order).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not create order")
		}
	}

	// Получаем продукты по переданным ID и добавляем их в заказ
	var products []models.Product
	if err := h.db.Where("id IN ?", input.ProductIDs).Find(&products).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not find products")
	}

	// Ассоциируем продукты с заказом
	if err := h.db.Model(&order).Association("Products").Append(&products); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add products to order")
	}

	return c.Status(fiber.StatusCreated).JSON(order)
}

// UpdateOrderStatus обновляет статус заказа по ID
func (h *OrderHandler) UpdateOrderStatus(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	user := c.Locals("current_user").(models.User)
	var order models.Order

	if err := h.db.Where("id = ? AND user_id = ?", parsedId, user.ID).First(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "order not found or access denied")
	}

	type UpdateStatusInput struct {
		Status models.OrderStatus `json:"status"`
	}

	var input UpdateStatusInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	order.Status = input.Status
	if err := h.db.Save(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update order status")
	}

	return c.JSON(order)
}

// DeleteOrder удаляет заказ по ID (например, отмена заказа)
func (h *OrderHandler) DeleteOrder(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	user := c.Locals("current_user").(models.User)
	if err := h.db.Where("id = ? AND user_id = ?", parsedId, user.ID).Delete(&models.Order{}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete order")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
