// Central validation logic for the entire program
package validate

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

func RuneLen(s string) int { return utf8.RuneCountInString(s) }

func hasControlOrInvalidSeparators(s string) bool {
	for _, r := range s {
		// \u2028 and \u2029 are line separators that can mess up logs/UI
		if unicode.IsControl(r) || r == '\u2028' || r == '\u2029' {
			return true
		}
	}
	return false
}

func hasNewlineOrTab(s string) bool {
	for _, r := range s {
		if r == '\n' || r == '\r' || r == '\t' {
			return true
		}
	}
	return false
}

type TextRules struct {
	Field      string
	Required   bool
	Min, Max   int  // rune length bounds; 0 means "no bound"
	SingleLine bool // reject \n \r \t
	Trim       bool
}

// Text validates + sanitizes a string field.
// Returns sanitized value + per-field errors.
// IMPORTANT: If Required is false and the value is empty, it's valid and returns no errors.
func Text(value string, rules TextRules) (string, []res.FieldError) {
	if rules.Trim {
		value = strings.TrimSpace(value)
	}

	// Required check first
	if value == "" {
		if rules.Required {
			return value, []res.FieldError{res.Required(rules.Field)}
		}
		return value, nil
	}

	var errs []res.FieldError

	// Length checks
	n := RuneLen(value)
	if rules.Min > 0 && n < rules.Min {
		errs = append(errs, res.MinLen(rules.Field, rules.Min))
	}
	if rules.Max > 0 && n > rules.Max {
		errs = append(errs, res.MaxLen(rules.Field, rules.Max))
	}

	// Character checks
	if hasControlOrInvalidSeparators(value) {
		errs = append(errs, res.Invalid(rules.Field, "contains invalid characters"))
	}

	if rules.SingleLine && hasNewlineOrTab(value) {
		errs = append(errs, res.Invalid(rules.Field, "must be single-line"))
	}

	return value, errs
}

// Email validates + sanitizes an email.
// Email is optional by default: empty => ok.
// Enforce "required" outside by using TextRules.Required or explicit check.
func Email(field, value string, maxRunes int) (string, []res.FieldError) {
	value = strings.TrimSpace(value)

	if value == "" {
		return value, nil
	}

	var errs []res.FieldError

	if maxRunes > 0 && RuneLen(value) > maxRunes {
		errs = append(errs, res.MaxLen(field, maxRunes))
		return value, errs
	}

	if hasControlOrInvalidSeparators(value) {
		errs = append(errs, res.Invalid(field, "contains invalid characters"))
		return value, errs
	}

	// stdlib parser also rejects "Name <a@b.com>" by requiring exact match
	addr, err := mail.ParseAddress(value)
	if err != nil || addr.Address != value {
		errs = append(errs, res.Invalid(field, "email format is invalid"))
	}

	return value, errs
}
