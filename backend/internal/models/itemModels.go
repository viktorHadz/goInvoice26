package models

type Product struct {
	ID              int64   `json:"id"`
	ProductType     string  `json:"productType"` // style/sample
	PricingMode     string  `json:"pricingMode"` // flat/hourly
	ProductName     string  `json:"productName"`
	FlatPriceMinor  *int64  `json:"flatPriceMinor,omitempty"` // 1000 --> £10
	HourlyRateMinor *int64  `json:"hourlyRateMinor,omitempty"`
	MinutesWorked   *int64  `json:"minutesWorked,omitempty"`
	ClientID        int64   `json:"clientId"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       *string `json:"updated_at,omitempty"`
}
