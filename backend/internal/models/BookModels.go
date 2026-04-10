package models

type INVBookRevision struct {
	ID         int64   `json:"id"`
	RevisionNo int     `json:"revisionNo"`
	IssueDate  string  `json:"issueDate"`
	DueByDate  *string `json:"dueByDate,omitempty"`
	UpdatedAt  *string `json:"updatedAt,omitempty"`
}

type INVBookHistoryItem struct {
	ID          int64   `json:"id"`
	Type        string  `json:"type"`
	CreatedAt   string  `json:"createdAt"`
	RevisionNo  *int    `json:"revisionNo,omitempty"`
	ReceiptNo   *int    `json:"receiptNo,omitempty"`
	IssueDate   *string `json:"issueDate,omitempty"`
	DueByDate   *string `json:"dueByDate,omitempty"`
	PaymentDate *string `json:"paymentDate,omitempty"`
	AmountMinor *int64  `json:"amountMinor,omitempty"`
	Label       *string `json:"label,omitempty"`
}

type INVBookInvoice struct {
	ID                int64             `json:"id"`
	ClientID          int64             `json:"clientId"`
	ClientName        string            `json:"clientName"`
	ClientCompanyName string            `json:"clientCompanyName"`
	BaseNo            int               `json:"baseNo"`
	Status            string            `json:"status"`
	LatestRevisionNo  int               `json:"latestRevisionNo"`
	IssueDate         string            `json:"issueDate"`
	DueByDate         *string           `json:"dueByDate,omitempty"`
	TotalMinor        int64             `json:"totalMinor"`
	DepositMinor      int64             `json:"depositMinor"`
	PaidMinor         int64             `json:"paidMinor"`
	BalanceDueMinor   int64             `json:"balanceDueMinor"`
	Revisions         []INVBookRevision `json:"revisions"`
	History           []INVBookHistoryItem `json:"history"`
}

type INVBookOut struct {
	Items   []INVBookInvoice `json:"items"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	Count   int              `json:"count"`
	Total   int              `json:"total"`
	HasMore bool             `json:"hasMore"`
}
