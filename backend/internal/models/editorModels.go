package models

type INVBookRevision struct {
	ID         int64   `json:"id"`
	RevisionNo int64   `json:"revisionNo"`
	IssueDate  string  `json:"issueDate"`
	DueByDate  *string `json:"dueByDate,omitempty"`
	UpdatedAt  *string `json:"updatedAt,omitempty"`
}

type INVBookInvoice struct {
	ID        int64             `json:"id"`
	BaseNo    int64             `json:"baseNo"`
	Status    string            `json:"status"`
	Revisions []INVBookRevision `json:"revisions"`
}

type INVBookOut struct {
	Items []INVBookInvoice `json:"items"`
}
