package models

import "encoding/json"

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

// To go into DB
type ProductCreate struct {
	ProductType     string
	PricingMode     string
	ProductName     string
	FlatPriceMinor  *int64
	HourlyRateMinor *int64
	MinutesWorked   *int64
	ClientID        int64
}

// Input received from frontend (clientID received from path params)
type ProductCreateIn struct {
	ProductType   *string      `json:"productType"`
	PricingMode   *string      `json:"pricingMode"`
	ProductName   *string      `json:"productName"`
	FlatPrice     *json.Number `json:"flatPrice,omitempty"`     // 25.65
	HourlyRate    *json.Number `json:"hourlyRate,omitempty"`    // 40.00
	MinutesWorked *json.Number `json:"minutesWorked,omitempty"` // 120
}

type ProductUpdate struct {
	ProductType     *string `json:"productType"`
	PricingMode     *string `json:"pricingMode"`
	ProductName     *string `json:"productName"`
	FlatPriceMinor  *int64  `json:"flatPriceMinor,omitempty"`
	HourlyRateMinor *int64  `json:"hourlyRateMinor,omitempty"`
	MinutesWorked   *int64  `json:"minutesWorked,omitempty"`
	ClientID        int64   `json:"clientId"`
}
