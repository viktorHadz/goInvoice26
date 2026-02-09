package models

type Client struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	CompanyName string  `json:"company_name"`
	Address     string  `json:"address"`
	Email       string  `json:"email"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   *string `json:"updated_at,omitempty"`
}
type CreateClient struct {
	Name        string `json:"name" binding:"required"`
	CompanyName string `json:"company_name"`
	Address     string `json:"address"`
	Email       string `json:"email"`
}
type UpdateClientInput struct {
	Name        *string `json:"name"`
	CompanyName *string `json:"company_name"`
	Address     *string `json:"address"`
	Email       *string `json:"email"`
}
