package clients

import (
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/validate"
)

//---------------------------------
// Validation Wrappers  => Validate and append to errors
//---------------------------------

// validates a concrete string field, returns sanitized value and appends errors.
func validateString(value string, rules validate.TextRules, errs *[]res.FieldError) string {
	v, e := validate.Text(value, rules)
	*errs = append(*errs, e...)
	return v
}

// validates a *string only if present (PATCH semantics), writes back sanitized value.
func validateStringPtr(ptr **string, rules validate.TextRules, errs *[]res.FieldError) {
	if *ptr == nil {
		return
	}
	v, e := validate.Text(**ptr, rules)
	*errs = append(*errs, e...)
	**ptr = v
}

// validates email only if present (PATCH semantics), writes back sanitized value.
func validateEmailPtr(ptr **string, field string, maxRunes int, errs *[]res.FieldError) {
	if *ptr == nil {
		return
	}
	v, e := validate.Email(field, **ptr, maxRunes)
	*errs = append(*errs, e...)
	**ptr = v
}

// ---------------------------------
// Validation Wrappers  => CREATE and POST
// ---------------------------------
// Sanitizes input received from client when creating a new client
func ValidateCreate(in models.CreateClient) (models.CreateClient, error) {
	errs := []res.FieldError{}

	in.Name = validateString(in.Name, validate.TextRules{
		Field: "name", Required: true, Min: 2, Max: 50, SingleLine: true, Trim: true,
	}, &errs)

	in.CompanyName = validateString(in.CompanyName, validate.TextRules{
		Field: "companyName", Max: 70, SingleLine: true, Trim: true,
	}, &errs)

	in.Address = validateString(in.Address, validate.TextRules{
		Field: "address", Max: 70, SingleLine: true, Trim: true,
	}, &errs)

	in.Email, _ = validate.Email("email", in.Email, 50)
	_, emailErrs := validate.Email("email", in.Email, 50)
	errs = append(errs, emailErrs...)

	if len(errs) > 0 {
		return in, res.Validation(errs...)
	}
	return in, nil
}

// Sanitizes input received from client when updating a new client
func ValidateUpdate(in models.UpdateClient) (models.UpdateClient, error) {
	errs := []res.FieldError{}

	validateStringPtr(&in.Name, validate.TextRules{
		Field: "name", Required: true, Min: 2, Max: 50, SingleLine: true, Trim: true,
	}, &errs)

	validateStringPtr(&in.CompanyName, validate.TextRules{
		Field: "companyName", Max: 70, SingleLine: true, Trim: true,
	}, &errs)

	validateStringPtr(&in.Address, validate.TextRules{
		Field: "address", Max: 70, SingleLine: true, Trim: true,
	}, &errs)

	validateEmailPtr(&in.Email, "email", 50, &errs)

	// Reject empty PATCH!
	if in.Name == nil && in.CompanyName == nil && in.Address == nil && in.Email == nil {
		errs = append(errs, res.Invalid("request", "no fields to update"))
	}

	if len(errs) > 0 {
		return in, res.Validation(errs...)
	}
	return in, nil
}
