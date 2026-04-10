package models

type InvoiceEditorResponse struct {
	Status          string                     `json:"status"`
	Totals          InvoiceEditorTotals        `json:"totals"`
	Lines           []InvoiceEditorLine        `json:"lines"`
	History         []InvoiceEditorHistoryItem `json:"history"`
	SelectedReceipt *InvoiceEditorReceipt      `json:"selectedReceipt,omitempty"`
	Payments []InvoiceEditorPayment `json:"payments"`
}

type InvoiceEditorTotals struct {
	BaseNumber        int64   `json:"baseNumber"`
	RevisionNo        int64   `json:"revisionNo"`
	IssueDate         string  `json:"issueDate"`
	SupplyDate        *string `json:"supplyDate,omitempty"`
	DueByDate         *string `json:"dueByDate,omitempty"`
	ClientName        string  `json:"clientName"`
	ClientCompanyName string  `json:"clientCompanyName"`
	ClientAddress     string  `json:"clientAddress"`
	ClientEmail       string  `json:"clientEmail"`
	Note              *string `json:"note,omitempty"`

	VATRate       int64  `json:"vatRate"`
	VATAmountMin  int64  `json:"vatAmountMinor"`
	DiscountType  string `json:"discountType"`
	DiscountRate  int64  `json:"discountRate"`
	DiscountMinor int64  `json:"discountMinor"`
	DepositType   string `json:"depositType"`
	DepositRate   int64  `json:"depositRate"`
	DepositMinor  int64  `json:"depositMinor"`
	SubtotalMinor int64  `json:"subtotalMinor"`
	TotalMinor    int64  `json:"totalMinor"`
	PaidMinor     int64  `json:"paidMinor"`
}

type InvoiceEditorLine struct {
	ProductID     *int64  `json:"productId,omitempty"`
	PricingMode   *string `json:"pricingMode,omitempty"`
	MinutesWorked *int64  `json:"minutesWorked,omitempty"`
	Name          string  `json:"name"`
	LineType      string  `json:"lineType"`
	Quantity      int64   `json:"quantity"`
	UnitPriceMin  int64   `json:"unitPriceMinor"`
	LineTotalMin  int64   `json:"lineTotalMinor"`
	SortOrder     int64   `json:"sortOrder"`
}

type InvoiceEditorPayment struct {
	ID          int64   `json:"id"`
	AmountMinor int64   `json:"amountMinor"`
	PaymentDate string  `json:"paymentDate"`
	PaymentType string  `json:"paymentType"`
	Label       *string `json:"label,omitempty"`
}

type InvoiceEditorHistoryItem struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	CreatedAt   string  `json:"createdAt"`
	RevisionNo  *int64  `json:"revisionNo,omitempty"`
	ReceiptNo   *int64  `json:"receiptNo,omitempty"`
	IssueDate   *string `json:"issueDate,omitempty"`
	DueByDate   *string `json:"dueByDate,omitempty"`
	PaymentDate *string `json:"paymentDate,omitempty"`
	AmountMinor *int64  `json:"amountMinor,omitempty"`
	Label       *string `json:"label,omitempty"`
}

type InvoiceEditorReceipt struct {
	ID               int64   `json:"id"`
	ReceiptNo        int64   `json:"receiptNo"`
	PaymentDate      string  `json:"paymentDate"`
	AmountMinor      int64   `json:"amountMinor"`
	Label            *string `json:"label,omitempty"`
	AppliedRevisionNo int64  `json:"appliedRevisionNo"`
	CreatedAt        string  `json:"createdAt"`
}
