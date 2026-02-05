package clients

import (
	"time"
)

type Client struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	CompanyName string     `json:"company_name"`
	Address     string     `json:"address"`
	Email       string     `json:"email"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}
type CreateClientInput struct {
	Name        string `json:"name" binding:"required"`
	CompanyName string `json:"company_name"`
	Address     string `json:"address"`
	Email       string `json:"email"`
}
