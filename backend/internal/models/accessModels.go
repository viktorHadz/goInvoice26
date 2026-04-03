package models

type DirectAccessGrant struct {
	ID        int64  `json:"id"`
	Email     string `json:"email"`
	Plan      string `json:"plan"`
	Note      string `json:"note,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type PromoCode struct {
	ID              int64  `json:"id"`
	Code            string `json:"code"`
	DurationDays    int    `json:"durationDays"`
	Active          bool   `json:"active"`
	RedemptionCount int    `json:"redemptionCount"`
	CreatedAt       string `json:"createdAt"`
}

type PromoRedemptionResult struct {
	Code         string `json:"code"`
	DurationDays int    `json:"durationDays"`
	ExpiresAt    string `json:"expiresAt"`
}

type PlatformAccessOverview struct {
	DirectGrants []DirectAccessGrant `json:"directGrants"`
	PromoCodes   []PromoCode         `json:"promoCodes"`
}

type AccountAccessState struct {
	AccessGranted bool   `json:"accessGranted"`
	Source        string `json:"source,omitempty"`
	Plan          string `json:"plan,omitempty"`
	ExpiresAt     string `json:"expiresAt,omitempty"`
	PromoCode     string `json:"promoCode,omitempty"`
	PromoExpired  bool   `json:"promoExpired"`
}
