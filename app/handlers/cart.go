package handlers

import (
	"fusion/app/database/models"
	"fusion/app/middleware"
	"fusion/app/schemas"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CartRoute struct {
	db *gorm.DB
}

func RegisterCartRoute(app *fiber.App, db *gorm.DB) {
	handler := &CartRoute{
		db: db,
	}

	cartGroup := app.Group("/cart")
	cartGroup.Use(middleware.AuthMiddleware())
	cartGroup.Get("/cart", handler.GetCart)
	cartGroup.Post("/cart", handler.AddToCart)
	cartGroup.Delete("/cart", handler.RemoveFromCart)
}

func (h *CartRoute) GetCart(c *fiber.Ctx) error {
	user := c.Locals("current_user").(models.User)

	var cart models.Cart
	if err := h.db.
		Preload("Products").
		Where("user_id = ?", user.ID).
		First(&cart).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve cart")
	}

	response := schemas.CartResponse{
		ID:       cart.ID.String(),
		UserID:   cart.UserID.String(),
		Products: make([]schemas.CartProductResponse, len(cart.Products)),
	}

	for i, p := range cart.Products {
		response.Products[i] = schemas.CartProductResponse{
			ID:       p.ID.String(),
			Quantity: p.Quantity,
		}
	}

	return c.JSON(response)
}

func (h *CartRoute) AddToCart(c *fiber.Ctx) error {
	type AddToCartInput struct {
		ProductID string `json:"product_id"`
		Quantity  int    `json:"quantity"`
	}

	user := c.Locals("current_user").(models.User)

	var input AddToCartInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	var product models.Product
	if err := h.db.First(&product, "id = ?", input.ProductID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}

	var cart models.Cart
	if err := h.db.
		Preload("Products").
		Where("user_id = ?", user.ID).
		First(&cart).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve cart")
	}

	if cart.ID == uuid.Nil {
		cart = models.Cart{
			UserID: user.ID,
		}

		if err := h.db.Create(&cart).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "could not create cart")
		}
	} else {
		for _, p := range cart.Products {
			if p.ID == product.ID {
				return fiber.NewError(fiber.StatusBadRequest, "product already in cart")
			}
		}
	}

	cartProduct := models.CartProduct{
		CartID:    cart.ID,
		ProductID: product.ID,
		Quantity:  input.Quantity,
	}

	if err := h.db.Create(&cartProduct).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add product to cart")
	}

	response := schemas.CartResponse{
		ID:       cart.ID.String(),
		UserID:   cart.UserID.String(),
		Products: make([]schemas.CartProductResponse, len(cart.Products)),
	}

	for i, p := range cart.Products {
		response.Products[i] = schemas.CartProductResponse{
			ID:       p.ID.String(),
			Quantity: p.Quantity,
		}
	}

	return c.JSON(response)
}

func (h *CartRoute) RemoveFromCart(c *fiber.Ctx) error {
	type RemoveFromCartInput struct {
		ProductID string `json:"product_id"`
	}

	user := c.Locals("current_user").(models.User)

	var input RemoveFromCartInput
	if err := c.BodyParser(&input); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	var product models.Product
	if err := h.db.First(&product, "id = ?", input.ProductID).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}

	var cart models.Cart
	if err := h.db.
		Preload("Products").
		Where("user_id = ?", user.ID).
		First(&cart).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve cart")
	}

	for _, p := range cart.Products {
		if p.ProductID == product.ID {
			if err := h.db.Delete(&p).Error; err != nil {
				return fiber.NewError(fiber.StatusInternalServerError, "could not delete product from cart")
			}
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}
