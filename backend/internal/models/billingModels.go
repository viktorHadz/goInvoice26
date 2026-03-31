package models

type AuthBilling struct {
	Configured        bool   `json:"configured"`
	Status            string `json:"status"`
	AccessGranted     bool   `json:"accessGranted"`
	CanManage         bool   `json:"canManage"`
	PortalAvailable   bool   `json:"portalAvailable"`
	CurrentPeriodEnd  string `json:"currentPeriodEnd,omitempty"`
	CancelAtPeriodEnd bool   `json:"cancelAtPeriodEnd"`
}

type BillingSessionLink struct {
	URL string `json:"url"`
}
