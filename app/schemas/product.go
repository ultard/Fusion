package schemas

import "fusion/app/database/models"

type ProductResponse struct {
	ID          string             `json:"id"`
	UserID      string             `json:"user_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Price       float64            `json:"price"`
	Image       *string            `json:"image,omitempty"`
	Categories  *[]models.Category `json:"categories,omitempty"`
	Reviews     *[]models.Review   `json:"reviews,omitempty"`
}
