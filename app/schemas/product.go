package schemas

import "fusion/app/database/models"

type ProductResponse struct {
	ID          string            `json:"id"`
	UserID      string            `json:"user_id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Stock       int               `json:"stock"`
	Image       *string           `json:"image,omitempty"`
	Categories  []models.Category `json:"categories,omitempty"`
	Reviews     []models.Review   `json:"reviews,omitempty"`
}

type ProductUpdateRequest struct {
	Name        *string   `json:"name,omitempty"`
	Description *string   `json:"description,omitempty"`
	Price       *float64  `json:"price,omitempty"`
	Stock       *int      `json:"stock,omitempty"`
	Image       *string   `json:"image,omitempty"`
	Categories  *[]string `json:"categories,omitempty"`
}
