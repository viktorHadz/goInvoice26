package invoiceformat

import (
	"fmt"
	"regexp"
	"strings"
)

const defaultInvoicePrefix = "INV-"
var trailingDashRegex = regexp.MustCompile(`-\s*$`)

func formatBaseLabel(prefix string, baseNumber int64) string {
	cleanPrefix := trailingDashRegex.ReplaceAllString(prefix, "")
	cleanPrefix = strings.TrimSpace(cleanPrefix)
	if cleanPrefix == "" {
		return fmt.Sprintf("%d", baseNumber)
	}
	return fmt.Sprintf("%s - %d", cleanPrefix, baseNumber)
}

// FormatInvoiceNumber centralizes user-facing invoice numbering for all renderers.
//
// Display contract:
//   - revision <= 1 => "{prefix}{baseNumber}"
//   - revision > 1  => "{prefix}{baseNumber}.{revision-1}"
//
// This keeps DB revision storage unchanged while exposing client-facing numbers
// as base, base.1, base.2, ... across PDF and future DOCX renderers.
func FormatInvoiceNumber(prefix string, baseNumber int64, revisionNo int64) string {
	cleanPrefix := prefix
	if cleanPrefix == "" {
		cleanPrefix = defaultInvoicePrefix
	}

	baseLabel := formatBaseLabel(cleanPrefix, baseNumber)
	if revisionNo <= 1 {
		return baseLabel
	}

	return fmt.Sprintf("%s.%d", baseLabel, revisionNo-1)
}
