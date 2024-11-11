package schemas

import "github.com/google/uuid"

type UserResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Avatar   *string   `json:"avatar,omitempty"`
}

type UserMeResponse struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   *string   `json:"avatar,omitempty"`
}
