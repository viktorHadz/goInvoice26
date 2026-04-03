package models

type AuthBilling struct {
	Configured             bool   `json:"configured"`
	Status                 string `json:"status"`
	AccessGranted          bool   `json:"accessGranted"`
	AccessSource           string `json:"accessSource,omitempty"`
	AccessExpiresAt        string `json:"accessExpiresAt,omitempty"`
	PromoCode              string `json:"promoCode,omitempty"`
	PromoExpired           bool   `json:"promoExpired"`
	CanManage              bool   `json:"canManage"`
	PortalAvailable        bool   `json:"portalAvailable"`
	TrialDays              int    `json:"trialDays"`
	Plan                   string `json:"plan,omitempty"`
	Interval               string `json:"interval,omitempty"`
	SeatLimit              int    `json:"seatLimit"`
	SinglePlanAvailable    bool   `json:"singlePlanAvailable"`
	TeamPlanAvailable      bool   `json:"teamPlanAvailable"`
	SingleMonthlyAvailable bool   `json:"singleMonthlyAvailable"`
	SingleYearlyAvailable  bool   `json:"singleYearlyAvailable"`
	TeamMonthlyAvailable   bool   `json:"teamMonthlyAvailable"`
	TeamYearlyAvailable    bool   `json:"teamYearlyAvailable"`
	CurrentPeriodEnd       string `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd      bool   `json:"cancelAtPeriodEnd"`
}

type PublicBillingCatalog struct {
	Configured              bool   `json:"configured"`
	TrialDays               int    `json:"trialDays"`
	SingleMonthlyAvailable  bool   `json:"singleMonthlyAvailable"`
	SingleYearlyAvailable   bool   `json:"singleYearlyAvailable"`
	TeamMonthlyAvailable    bool   `json:"teamMonthlyAvailable"`
	TeamYearlyAvailable     bool   `json:"teamYearlyAvailable"`
	SingleMonthlyPriceLabel string `json:"singleMonthlyPriceLabel,omitempty"`
	SingleYearlyPriceLabel  string `json:"singleYearlyPriceLabel,omitempty"`
	TeamMonthlyPriceLabel   string `json:"teamMonthlyPriceLabel,omitempty"`
	TeamYearlyPriceLabel    string `json:"teamYearlyPriceLabel,omitempty"`
}

type BillingSessionLink struct {
	URL string `json:"url"`
}
