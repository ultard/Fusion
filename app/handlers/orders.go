package handlers

import (
	"fusion/app/database/models"
	"fusion/app/middleware"
	"fusion/app/schemas"
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

// CreateOrder создает новый заказ из выбранных товаров корзины
func (h *OrderHandler) CreateOrder(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)

	// Получаем данные из запроса
	var input schemas.CreateOrderRequest
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	// Проверяем наличие корзины для текущего пользователя
	var cart models.Cart
	if err := h.db.Preload("Products").Where("user_id = ?", user.ID).First(&cart).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "cart not found")
	}

	// Создаем новый заказ
	order := models.Order{
		UserID: user.ID,
		Status: models.CREATED, // Или другой статус по умолчанию
	}
	if err := h.db.Create(&order).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create order")
	}

	var orderProducts []models.CartProduct
	var productsToRemove []uuid.UUID
	for _, selectedProduct := range input.CartProductResponse {
		for _, cartProduct := range cart.Products {
			if cartProduct.ProductID.String() == selectedProduct.ID {
				orderProduct := models.CartProduct{
					CartID:    order.ID,
					ProductID: cartProduct.ProductID,
					Quantity:  cartProduct.Quantity,
				}

				orderProducts = append(orderProducts, orderProduct)
				productsToRemove = append(productsToRemove, cartProduct.ID)
				break
			}
		}
	}

	if err := h.db.Model(&order).Association("Products").Append(orderProducts); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add products to order")
	}

	if err := h.db.Where("id IN ?", productsToRemove).Delete(&models.CartProduct{}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove products from cart")
	}

	response := schemas.CreateOrderResponse{
		ID:           order.ID.String(),
		CartProducts: orderProducts,
		UserID:       order.UserID.String(),
		Status:       int(order.Status),
	}

	return c.Status(fiber.StatusCreated).JSON(response)
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
