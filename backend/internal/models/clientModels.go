package models

type Client struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	CompanyName string  `json:"companyName"`
	Address     string  `json:"address"`
	Email       string  `json:"email"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   *string `json:"updated_at,omitempty"`
}
type CreateClient struct {
	Name        string `json:"name" binding:"required"`
	CompanyName string `json:"companyName"`
	Address     string `json:"address"`
	Email       string `json:"email"`
}

type UpdateClient struct {
	Name        *string `json:"name"`
	CompanyName *string `json:"companyName"`
	Address     *string `json:"address"`
	Email       *string `json:"email"`
}
