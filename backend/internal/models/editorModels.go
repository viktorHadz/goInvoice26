package models

type InvoiceEditorResponse struct {
	Totals InvoiceEditorTotals `json:"totals"`
	Lines  []InvoiceEditorLine `json:"lines"`
}

type InvoiceEditorTotals struct {
	BaseNumber        int64   `json:"baseNumber"`
	RevisionNo        int64   `json:"revisionNo"`
	IssueDate         string  `json:"issueDate"`
	DueByDate         *string `json:"dueByDate,omitempty"`
	ClientName        string  `json:"clientName"`
	ClientCompanyName string  `json:"clientCompanyName"`
	ClientAddress     string  `json:"clientAddress"`
	ClientEmail       string  `json:"clientEmail"`
	Note              *string `json:"note,omitempty"`

	VATRate       int64  `json:"vatRate"`
	VATAmountMin  int64  `json:"vatAmountMin"`
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
	Name         string `json:"name"`
	LineType     string `json:"lineType"`
	Quantity     int64  `json:"quantity"`
	UnitPriceMin int64  `json:"unitPriceMin"`
	LineTotalMin int64  `json:"lineTotalMin"`
	SortOrder    int64  `json:"sortOrder"`
}
