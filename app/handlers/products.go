package handlers

import (
	"fusion/app/database/models"
	"fusion/app/middleware"
	"fusion/app/schemas"
	"fusion/app/utils"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type ProductHandler struct {
	db *gorm.DB
}

// RegisterProductRoutes регистрирует маршруты для продуктов
func RegisterProductRoutes(app *fiber.App, db *gorm.DB) {
	handler := &ProductHandler{
		db: db,
	}

	productGroup := app.Group("/products")
	productGroup.Get("/", handler.GetProducts)
	productGroup.Get("/:id", handler.GetProduct)

	productGroup.Use(middleware.AuthMiddleware())
	productGroup.Post("/", handler.CreateProduct)
	productGroup.Put("/:id", handler.UpdateProduct)
	productGroup.Delete("/:id", handler.DeleteProduct)

	productGroup.Post("/:id/reviews", handler.CreateReview)
	productGroup.Delete("/:id/reviews", handler.RemoveReview)

	productGroup.Post("/:id/favorites", handler.AddToFavorites)
	productGroup.Delete("/:id/favorites", handler.RemoveFromFavorites)
}

// GetProducts возвращает список всех продуктов
func (h *ProductHandler) GetProducts(c *fiber.Ctx) error {
	var products []models.Product
	if err := h.db.
		Preload("Review").
		Preload("Category").
		Find(&products).
		Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not retrieve products")
	}

	response := make([]schemas.ProductResponse, len(products))
	for i, product := range products {
		response[i] = schemas.ProductResponse{
			ID:          product.ID.String(),
			UserID:      product.UserID.String(),
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Image:       product.Image,
			Categories:  product.Categories,
			Reviews:     product.Reviews,
		}
	}

	return c.JSON(response)
}

// GetProduct возвращает продукт по ID
func (h *ProductHandler) GetProduct(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	var product models.Product
	if err := h.db.
		Preload("Review").
		Preload("Category").
		First(&product, "id = ?", parsedId).
		Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}

	response := schemas.ProductResponse{
		ID:          product.ID.String(),
		UserID:      product.UserID.String(),
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Image:       product.Image,
		Categories:  product.Categories,
		Reviews:     product.Reviews,
	}

	return c.JSON(response)
}

// CreateProduct создает новый продукт
func (h *ProductHandler) CreateProduct(c *fiber.Ctx) error {
	var product models.Product
	if err := c.BodyParser(&product); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if err := h.db.Create(&product).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create product")
	}

	return c.Status(fiber.StatusCreated).JSON(product)
}

// UpdateProduct обновляет информацию о продукте
func (h *ProductHandler) UpdateProduct(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	user := c.Locals("current_user").(models.User)

	var product models.Product
	if err := h.db.Where("id = ? AND user_id = ?", parsedId, user.ID).First(&product).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found or access denied")
	}

	var updateFields schemas.ProductUpdateRequest
	if err := c.BodyParser(&updateFields); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	if updateFields.Name != nil {
		product.Name = *updateFields.Name
	}
	if updateFields.Description != nil {
		product.Description = *updateFields.Description
	}
	if updateFields.Image != nil {
		product.Image = updateFields.Image
	}
	if updateFields.Price != nil {
		product.Price = *updateFields.Price
	}
	if updateFields.Stock != nil {
		product.Stock = *updateFields.Stock
	}
	if updateFields.Categories != nil {
		var categories []models.Category
		for _, categoryName := range *updateFields.Categories {
			categories = append(categories, models.Category{Name: categoryName})
		}
		product.Categories = &categories
	}

	product.Image = updateFields.Image

	if err := h.db.Save(&product).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not update product")
	}

	return c.SendStatus(fiber.StatusAccepted)
}

// DeleteProduct удаляет продукт по ID
func (h *ProductHandler) DeleteProduct(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	user := c.Locals("current_user").(models.User)
	if err := h.db.Where("id = ? AND user_id = ?", parsedId, user.ID).Delete(&models.Product{}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not delete product")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// CreateReview создает новый отзыв о продукте
func (h *ProductHandler) CreateReview(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	var product models.Product
	if err := h.db.First(&product, parsedId).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}

	user := c.Locals("current_user").(models.User)
	if err := h.db.Where("product_id = ? AND user_id = ?", product.ID, user.ID).First(&models.Review{}).Error; err == nil {
		return fiber.NewError(fiber.StatusBadRequest, "review already exists")
	}

	var review struct {
		Rating  float64 `json:"rating"`
		Comment string  `json:"comment"`
	}
	if err := c.BodyParser(&review); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid input")
	}

	newReview := models.Review{
		ProductID: product.ID,
		UserID:    user.ID,
		Rating:    review.Rating,
		Comment:   review.Comment,
	}

	if err := h.db.Create(&newReview).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not create review")
	}

	return c.Status(fiber.StatusCreated).JSON(newReview)
}

// RemoveReview удаляет отзыв о продукте
func (h *ProductHandler) RemoveReview(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	user := c.Locals("current_user").(models.User)
	if err := h.db.Where("product_id = ? AND user_id = ?", parsedId, user.ID).Delete(&models.Review{}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove review")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// AddToFavorites добавляет продукт в избранное
func (h *ProductHandler) AddToFavorites(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	var product models.Product
	if err := h.db.First(&product, parsedId).Error; err != nil {
		return fiber.NewError(fiber.StatusNotFound, "product not found")
	}

	user := c.Locals("current_user").(models.User)
	favorite := models.Favourite{ProductID: product.ID, UserID: user.ID}

	if err := h.db.Where("product_id = ? AND user_id = ?", product.ID, user.ID).FirstOrCreate(&favorite).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not add to favorites")
	}

	return c.SendStatus(fiber.StatusCreated)
}

// RemoveFromFavorites удаляет продукт из избранного
func (h *ProductHandler) RemoveFromFavorites(c *fiber.Ctx) error {
	parsedId, err := utils.ParseRouteID(c)
	if err != nil {
		return err
	}

	user := c.Locals("current_user").(models.User)
	if err := h.db.Where("product_id = ? AND user_id = ?", parsedId, user.ID).Delete(&models.Favourite{}).Error; err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "could not remove from favorites")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
