package validate

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

func RuneLen(s string) int { return utf8.RuneCountInString(s) }

func HasControlChars(s string) bool {
	for _, r := range s {
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return true
		}
	}
	return false
}

type TextRules struct {
	Field      string
	Required   bool
	Min, Max   int  // in runes (characters)
	SingleLine bool // reject \n \r \t
	Trim       bool
}

func Text(s string, rules TextRules) (string, []res.FieldError) {
	errs := []res.FieldError{}

	if rules.Trim {
		s = strings.TrimSpace(s)
	}

	if rules.Required && s == "" {
		return s, append(errs, res.Required(rules.Field))
	}
	if s == "" {
		return s, errs
	}

	n := RuneLen(s)
	if (rules.Min > 0 && n < rules.Min) || (rules.Max > 0 && n > rules.Max) {
		errs = append(errs, res.OutOfRange(rules.Field, rules.Min, rules.Max))
	}

	if HasControlChars(s) {
		errs = append(errs, res.Invalid(rules.Field, "contains invalid characters"))
	} else if rules.SingleLine {
		for _, r := range s {
			if r == '\n' || r == '\r' || r == '\t' {
				errs = append(errs, res.Invalid(rules.Field, "must be single-line"))
				break
			}
		}
	}

	return s, errs
}

func Email(field, email string, maxRunes int) (string, []res.FieldError) {
	errs := []res.FieldError{}
	email = strings.TrimSpace(email)

	if email == "" {
		return email, errs // optional; enforce required outside if needed
	}

	if RuneLen(email) > maxRunes {
		return email, append(errs, res.MaxLen(field, maxRunes))
	}
	if HasControlChars(email) {
		return email, append(errs, res.Invalid(field, "contains invalid characters"))
	}

	// stdlib parser also rejects "Name <a@b.com>" by requiring exact match
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email {
		return email, append(errs, res.Invalid(field, "email format is invalid"))
	}

	return email, errs
}
