package schemas

import "fusion/app/database/models"

type CreateOrderRequest struct {
	CartProductResponse []CartProductResponse `json:"products"`
}

type CreateOrderResponse struct {
	ID           string               `json:"id"`
	CartProducts []models.CartProduct `json:"cart_products"`
	UserID       string               `json:"user_id"`
	Status       int                  `json:"status"`
}
