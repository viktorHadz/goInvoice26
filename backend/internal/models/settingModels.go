package models

type Settings struct {
	CompanyName    string `json:"companyName"`
	Email          string `json:"email"`
	Phone          string `json:"phone"`
	CompanyAddress string `json:"companyAddress"`
	InvoicePrefix  string `json:"invoicePrefix"`
	Currency       string `json:"currency"`
	DateFormat     string `json:"dateFormat"`
	PaymentTerms   string `json:"paymentTerms"`
	PaymentDetails string `json:"paymentDetails"`
	NotesFooter    string `json:"notesFooter"`
	LogoURL        string `json:"logoUrl"`
}
