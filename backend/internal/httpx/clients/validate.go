package clients

import (
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/validate"
)

func ValidateCreate(client models.CreateClient) (models.CreateClient, []res.FieldError) {
	var errs []res.FieldError

	client.Name, errs = text(client.Name, validate.TextRules{
		Field: "name", Required: true, Min: 2, Max: 50, SingleLine: true, Trim: true,
	}, errs)

	client.CompanyName, errs = text(client.CompanyName, validate.TextRules{
		Field: "companyName", Max: 70, SingleLine: true, Trim: true,
	}, errs)

	client.Address, errs = text(client.Address, validate.TextRules{
		Field: "address", Max: 70, SingleLine: true, Trim: true,
	}, errs)

	client.Email, errs = email(client.Email, "email", 50, errs)

	return client, errs
}

func ValidateUpdate(client models.UpdateClient) (models.UpdateClient, []res.FieldError) {
	var errs []res.FieldError

	client.Name, errs = textPtr(client.Name, validate.TextRules{
		Field: "name", Min: 2, Max: 50, SingleLine: true, Trim: true, Required: true,
	}, errs)

	client.CompanyName, errs = textPtr(client.CompanyName, validate.TextRules{
		Field: "companyName", Max: 70, SingleLine: true, Trim: true,
	}, errs)

	client.Address, errs = textPtr(client.Address, validate.TextRules{
		Field: "address", Max: 70, SingleLine: true, Trim: true,
	}, errs)

	client.Email, errs = emailPtr(client.Email, "email", 50, errs)

	// Reject empties:
	// check if name is not nill pointer first (crashes program) then check if its empty
	if client.Name != nil && *client.Name == "" {
		errs = append(errs, res.Required("name"))
	}
	if client.Name == nil && client.CompanyName == nil && client.Address == nil && client.Email == nil {
		errs = append(errs, res.Invalid("request", "no fields to update"))
	}

	return client, errs
}

// --------------------
// helpers
// --------------------

func text(value string, rules validate.TextRules, errs []res.FieldError) (string, []res.FieldError) {
	v, e := validate.Text(value, rules)
	return v, append(errs, e...)
}

func email(value string, field string, maxRunes int, errs []res.FieldError) (string, []res.FieldError) {
	v, e := validate.Email(field, value, maxRunes)
	return v, append(errs, e...)
}

func textPtr(ptr *string, rules validate.TextRules, errs []res.FieldError) (*string, []res.FieldError) {
	if ptr == nil {
		return nil, errs
	}
	v, e := validate.Text(*ptr, rules)
	errs = append(errs, e...)
	out := v
	return &out, errs
}

func emailPtr(ptr *string, field string, maxRunes int, errs []res.FieldError) (*string, []res.FieldError) {
	if ptr == nil {
		return nil, errs
	}
	v, e := validate.Email(field, *ptr, maxRunes)
	errs = append(errs, e...)
	out := v
	return &out, errs
}
