package pdf

import (
	"fmt"
	"strings"
)

const notApplicableValue = " "

type invoicePDFPricing struct {
	itemPrice  string
	timeWorked string
	hourlyRate string
}

func formatDurationMinutes(minutes int64) string {
	if minutes < 0 {
		minutes = 0
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if hours == 0 {
		return fmt.Sprintf("%dm", remainingMinutes)
	}
	if remainingMinutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh %dm", hours, remainingMinutes)
}

func buildInvoicePDFPricing(
	pricingMode string,
	unitPriceMinor int64,
	minutes *int64,
	currency string,
) invoicePDFPricing {
	if strings.TrimSpace(strings.ToLower(pricingMode)) != "hourly" {
		return invoicePDFPricing{
			itemPrice:  formatMoney(unitPriceMinor, currency),
			timeWorked: notApplicableValue,
			hourlyRate: notApplicableValue,
		}
	}

	timeWorked := notApplicableValue
	if minutes != nil {
		timeWorked = formatDurationMinutes(*minutes)
	}

	return invoicePDFPricing{
		itemPrice:  notApplicableValue,
		timeWorked: timeWorked,
		hourlyRate: formatMoney(unitPriceMinor, currency) + "/hr",
	}
}
