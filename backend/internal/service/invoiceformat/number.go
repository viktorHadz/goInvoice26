package invoiceformat

import (
	"fmt"
	"regexp"
	"strings"
)

const defaultInvoicePrefix = "INV"

var trailingDashRegex = regexp.MustCompile(`-\s*$`)

func formatBaseLabel(prefix string, baseNumber int64) string {
	cleanPrefix := trailingDashRegex.ReplaceAllString(prefix, "")
	cleanPrefix = strings.TrimSpace(cleanPrefix)
	if cleanPrefix == "" {
		return fmt.Sprintf("%d", baseNumber)
	}
	return fmt.Sprintf("%s-%d", cleanPrefix, baseNumber)
}

func FormatInvoiceNumber(prefix string, baseNumber int64, revisionNo int64) string {
	cleanPrefix := prefix
	if cleanPrefix == "" {
		cleanPrefix = defaultInvoicePrefix
	}

	baseLabel := formatBaseLabel(cleanPrefix, baseNumber)
	if revisionNo <= 1 {
		return baseLabel
	}

	return fmt.Sprintf("%s.%d", baseLabel, revisionNo)
}

func FormatPaymentReceiptNumber(prefix string, baseNumber int64, revisionNo int64, receiptNo int64) string {
	cleanPrefix := prefix
	if cleanPrefix == "" {
		cleanPrefix = defaultInvoicePrefix
	}

	referenceLabel := FormatInvoiceNumber(cleanPrefix, baseNumber, revisionNo)
	if receiptNo <= 1 {
		return fmt.Sprintf("%s-PR-1", referenceLabel)
	}

	return fmt.Sprintf("%s-PR-%d", referenceLabel, receiptNo)
}
