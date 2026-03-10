package models

type FEInvoiceIn struct {
	Overview InvoiceCreateIn
	Lines    []LineCreateIn
	Totals   TotalsCreateIn
}

type InvoiceCreateIn struct {
	ClientID int64 `json:"clientId"`
	// CurrentRevisionID int64   `json:"currentRevisionId"` ---> Set by server
	BaseNumber int64 `json:"baseNumber"`
	// Status            string  `json:"status"` ---> Set by server
	// CreatedAt         string  `json:"createdAt"`
	IssueDate         string  `json:"issueDate"`
	DueByDate         *string `json:"dueByDate"`
	ClientName        string  `json:"clientName"`
	ClientCompanyName string  `json:"clientCompanyName"`
	ClientAddress     string  `json:"clientAddress"`
	ClientEmail       string  `json:"clientEmail"`
	Note              *string `json:"note"`
}

type LineCreateIn struct {
	// RevisionID     int64  `json:"revisionId"` ---> Set by server
	ProductID      *int64 `json:"productId"` // NULLABLE
	Name           string `json:"name"`
	LineType       string `json:"lineType"`       // custom(default), style, sample,
	PricingMode    string `json:"pricingMode"`    // flat(default) hourly
	Quantity       int64  `json:"quantity"`       // default 1
	MinutesWorked  *int64 `json:"minutesWorked"`  // >= 0 || NULL
	UnitPriceMinor int64  `json:"unitPriceMinor"` // >= 0
	LineTotalMinor int64  `json:"lineTotalMinor"` // qty * unite price minor
	SortOrder      int64  `json:"sortOrder"`
}

// Can use to doublecheck if they match
type TotalsCreateIn struct {
	VATRate        int64 `json:"vatRate"`  // VAT RATE, percent based units  2000 = 20%
	VatAmountMinor int64 `json:"vatMinor"` // VAT ammount in minor

	DepositType  string `json:"depositType"` // none(default), percent, fixed
	DepositRate  int64  `json:"depositRate"`
	DepositMinor int64  `json:"depositMinor"`

	DiscountType  string `json:"discountType"` // none(default), percent, fixed
	DiscountRate  int64  `json:"discountRate"`
	DiscountMinor int64  `json:"discountMinor"`

	PaidMinor int64 `json:"paidMinor"`

	SubtotalAfterDisc int64 `json:"subtotalAfterDiscountMinor"`
	SubtotalMinor     int64 `json:"subtotalMinor"`
	TotalMinor        int64 `json:"totalMinor"`
	BalanceDue        int64 `json:"balanceDueMinor"`
}
