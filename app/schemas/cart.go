package schemas

type CartResponse struct {
	ID       string                `json:"id"`
	Products []CartProductResponse `json:"products"`
	UserID   string                `json:"user_id"`
}

type CartProductResponse struct {
	ID       string `json:"id"`
	Quantity int    `json:"quantity"`
}
