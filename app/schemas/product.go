package schemas

type ProductResponse struct {
	ID          string    `json:"ID"`
	UserId      string    `json:"UserId"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       int       `json:"price"`
	Image       *string   `json:"Image,omitempty"`
	Categories  *[]string `json:"Categories,omitempty"`
	Reviews     *[]string `json:"Reviews,omitempty"`
}
