package models

type FEInvoiceIn struct {
	Overview InvoiceCreateIn
	Lines    []LineCreateIn
	Totals   TotalsCreateIn
}

type InvoiceCreateIn struct {
	ClientID          int64   `json:"clientId"`
	BaseNumber        int64   `json:"baseNumber"`
	IssueDate         string  `json:"issueDate"`
	DueByDate         *string `json:"dueByDate"`
	ClientName        string  `json:"clientName"`
	ClientCompanyName string  `json:"clientCompanyName"`
	ClientAddress     string  `json:"clientAddress"`
	ClientEmail       string  `json:"clientEmail"`
	Note              *string `json:"note"`
}

type LineCreateIn struct {
	ProductID      *int64 `json:"productId"`
	Name           string `json:"name"`
	LineType       string `json:"lineType"`
	PricingMode    string `json:"pricingMode"`
	Quantity       int64  `json:"quantity"`
	MinutesWorked  *int64 `json:"minutesWorked"`
	UnitPriceMinor int64  `json:"unitPriceMinor"`
	LineTotalMinor int64  `json:"lineTotalMinor"`
	SortOrder      int64  `json:"sortOrder"`
}

type TotalsCreateIn struct {
	VATRate        int64 `json:"vatRate"`
	VatAmountMinor int64 `json:"vatMinor"`

	DepositType  string `json:"depositType"`
	DepositRate  int64  `json:"depositRate"`
	DepositMinor int64  `json:"depositMinor"`

	DiscountType  string `json:"discountType"`
	DiscountRate  int64  `json:"discountRate"`
	DiscountMinor int64  `json:"discountMinor"`

	PaidMinor int64 `json:"paidMinor"`

	SubtotalAfterDisc int64 `json:"subtotalAfterDiscountMinor"`
	SubtotalMinor     int64 `json:"subtotalMinor"`
	TotalMinor        int64 `json:"totalMinor"`
	BalanceDue        int64 `json:"balanceDueMinor"`
}

type InvoicePDFIssuer struct {
	CompanyName    string
	Email          string
	Phone          string
	CompanyAddress string
	LogoURL        string
}

type InvoicePDFItem struct {
	Name      string
	LineType  string
	Quantity  string
	ItemPrice string
	ItemTotal string
	SortOrder int64
}

type InvoicePDFData struct {
	Title              string
	InvoiceNumberLabel string
	Currency           string

	IssueAt string
	DueDate *string
	Note    *string

	Issuer InvoicePDFIssuer
	Client CreateClient

	Lines []InvoicePDFItem

	Totals TotalsCreateIn

	PaymentTerms   string
	PaymentDetails string
	NotesFooter    string
}
