package models

// All data for an invoice - HTTP GET
type Invoice struct {
	// invoices
	ClientID          string
	CurrentRevisionID int64
	BaseNumber        int64  // invoice number
	Status            string // draft, issued, paid, void
	CreatedAt         string
	UpdatedAt         string
	// invoice_revisions
	IssuedDate        string
	UpdatedDate       string
	ClientName        string
	ClientCompanyName string
	ClientAddress     string
	ClientEmail       string
	Note              string
	DiscountType      string
	DiscountValue     int64
	SubtotalMinor     int64
	VATRate           int64 // Percent
	VatAmountMinor    int64
	TotalMinor        int64
	RevisionCreatedAt string
	// Payments
	PaidMinor int64
}

// Allows loading items separately from invoice
type InvoiceItems struct {
	ProductID         int64
	RevisionID        int64
	Name              string
	LineType          string // style, sample, custom
	PricingMode       string // flat hourly
	Quantity          int64
	UnitPriceMinor    int64
	LineSubtotalMinor int64
	MinutesWorked     int64
}

// All data necessary for creating an invoice - HTTP CREATE
type createInvoice struct {
	// id
	// client_id
	// base_number
	// status
	// 'draft','issued','paid','void'

}
