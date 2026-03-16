package models

type INVBookRevision struct {
	ID         int64   `json:"id"`
	RevisionNo int     `json:"revisionNo"`
	IssueDate  string  `json:"issueDate"`
	DueByDate  *string `json:"dueByDate,omitempty"`
	UpdatedAt  *string `json:"updatedAt,omitempty"`
}

type INVBookInvoice struct {
	ID             int64             `json:"id"`
	BaseNo         int               `json:"baseNo"`
	Status         string            `json:"status"`
	LatestRevision int               `json:"latestRevision"`
	Revisions      []INVBookRevision `json:"revisions"`
}

type INVBookOut struct {
	Items   []INVBookInvoice `json:"items"`
	Limit   int              `json:"limit"`
	Offset  int              `json:"offset"`
	Count   int              `json:"count"`
	Total   int              `json:"total"`
	HasMore bool             `json:"hasMore"`
}
