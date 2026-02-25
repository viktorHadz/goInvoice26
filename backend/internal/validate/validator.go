// Central validation logic for the entire program
package validate

import (
	"math"
	"net/mail"
	"strconv"
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

type IntRules struct {
	Field    string
	Required bool
	Min, Max *int64 // bounds inclusive; nil means unset
	Trim     bool
}

// MoneyRules validates currency-like decimals and converts to minor units (decimal=2).
type MoneyRules struct {
	Field              string
	Required           bool
	MinMinor, MaxMinor *int64 // bounds in minor units
	Trim               bool
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

func Int64(value string, rules IntRules) (int64, []res.FieldError) {
	if rules.Trim {
		value = strings.TrimSpace(value)
	}

	if value == "" {
		if rules.Required {
			return 0, []res.FieldError{res.Required(rules.Field)}
		}
		return 0, nil
	}

	if hasControlOrInvalidSeparators(value) {
		return 0, []res.FieldError{res.Invalid(rules.Field, "contains invalid characters")}
	}

	n, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, []res.FieldError{res.Invalid(rules.Field, "must be an integer")}
	}

	var errs []res.FieldError
	if rules.Min != nil && n < *rules.Min {
		errs = append(errs, res.Invalid(rules.Field, "value below minimum"))
	}
	if rules.Max != nil && n > *rules.Max {
		errs = append(errs, res.Invalid(rules.Field, "value above maximum"))
	}
	return n, errs
}

// MoneyMinor parses strings like "25", "25.6", "25.65" into minor units (2 after decimal point).
// Rejects commas, scientific notation, and >2 decimal places. No float rounding.
func MoneyMinor(value string, rules MoneyRules) (int64, []res.FieldError) {
	if rules.Trim {
		value = strings.TrimSpace(value)
	}

	if value == "" {
		if rules.Required {
			return 0, []res.FieldError{res.Required(rules.Field)}
		}
		return 0, nil
	}

	if hasControlOrInvalidSeparators(value) {
		return 0, []res.FieldError{res.Invalid(rules.Field, "contains invalid characters")}
	}

	neg := false
	switch value[0] {
	case '+':
		value = value[1:]
	case '-':
		neg = true
		value = value[1:]
	}

	if value == "" {
		return 0, []res.FieldError{res.Invalid(rules.Field, "must be a number")}
	}

	var intPart, fracPart string
	if i := strings.IndexByte(value, '.'); i >= 0 {
		intPart = value[:i]
		fracPart = value[i+1:]
		if strings.IndexByte(fracPart, '.') >= 0 {
			return 0, []res.FieldError{res.Invalid(rules.Field, "must be a valid decimal")}
		}
	} else {
		intPart = value
		fracPart = ""
	}

	if intPart == "" {
		intPart = "0"
	}

	for _, r := range intPart {
		if r < '0' || r > '9' {
			return 0, []res.FieldError{res.Invalid(rules.Field, "must be a number")}
		}
	}
	for _, r := range fracPart {
		if r < '0' || r > '9' {
			return 0, []res.FieldError{res.Invalid(rules.Field, "must be a number")}
		}
	}

	if len(fracPart) > 2 {
		return 0, []res.FieldError{res.Invalid(rules.Field, "must have at most 2 decimal places")}
	}

	ip, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return 0, []res.FieldError{res.Invalid(rules.Field, "number out of range")}
	}

	// normalize frac to 2 digits
	if len(fracPart) == 1 {
		fracPart += "0"
	} else if len(fracPart) == 0 {
		fracPart = "00"
	}

	fp, err := strconv.ParseInt(fracPart, 10, 64)
	if err != nil {
		return 0, []res.FieldError{res.Invalid(rules.Field, "number out of range")}
	}

	// overflow-safe combine
	if ip > (math.MaxInt64-fp)/100 {
		return 0, []res.FieldError{res.Invalid(rules.Field, "number out of range")}
	}

	minor := ip*100 + fp
	if neg {
		minor = -minor
	}

	var errs []res.FieldError
	if rules.MinMinor != nil && minor < *rules.MinMinor {
		errs = append(errs, res.Invalid(rules.Field, "value below minimum"))
	}
	if rules.MaxMinor != nil && minor > *rules.MaxMinor {
		errs = append(errs, res.Invalid(rules.Field, "value above maximum"))
	}
	return minor, errs
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
