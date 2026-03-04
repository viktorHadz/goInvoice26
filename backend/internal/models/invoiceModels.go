package models

// All data for an invoice - HTTP GET
type Invoice struct {
	InvoiceID int64
	// Invoices
	ClientID          string
	CurrentRevisionID *int64  // > 0 || Null - new invoice set to NULL e.g. DRAFT
	BaseNumber        int64   // invoice number e.g."SAM-1"
	Status            *string // draft(default), issued, paid, void
	CreatedAt         *string
	UpdatedAt         *string
	// Invoice Revisions
	RevisionNumber    int64   // creating new so revision number is 1
	IssuedDate        string  // cant be null
	DueByDate         *string // CAN be NULLABLE leave pointer for DB
	ClientName        string
	ClientCompanyName string
	ClientAddress     string
	ClientEmail       string
	Note              *string
	VATRate           int64  // VAT RATE, percent based units  2000 = 20%
	DiscountType      string // none(default), percent, fixed
	DiscountMinor     int64
	SubtotalMinor     int64
	VatAmountMinor    int64 // VAT ammount in minor
	TotalMinor        int64
	RevisionCreatedAt *string
	// Payments
	PaymentType string // payment(default), deposit
	PaidMinor   *int64
	// Line Items
	Line []LineItem
}

// Allows loading items separately from invoice. Insert last
type LineItem struct {
	RevisionID        int64
	ProductID         *int64 // NULLABLE
	Name              string
	LineType          string // custom(default), style, sample,
	PricingMode       string // flat(default) hourly
	Quantity          int64  // default 1
	UnitPriceMinor    int64  // >= 0
	LineSubtotalMinor int64
	MinutesWorked     *int64 // >= 0 || NULL
}

// Makes a new draft invoice clientID retrieved from path pa
type NewInvoice struct {
	CurrentRevisionID *int64  `json:"currentRevisionID,omitempty"`
	BaseNumber        int64   `json:"baseNumber"`
	Status            *string `json:"status,omitempty"`
	CreatedAt         *string `json:"createdAt,omitempty"`
	UpdatedAt         *string `json:"updatedAt,omitempty"`
}
