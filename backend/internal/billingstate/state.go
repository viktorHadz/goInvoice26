package billingstate

import "strings"

const (
	StatusInactive           = "inactive"
	StatusTrialing           = "trialing"
	StatusActive             = "active"
	StatusPastDue            = "past_due"
	StatusCanceled           = "canceled"
	StatusUnpaid             = "unpaid"
	StatusIncomplete         = "incomplete"
	StatusIncompleteExpired  = "incomplete_expired"
	StatusPaused             = "paused"
	StatusCheckoutIncomplete = "checkout_incomplete"
)

func Normalize(status string) string {
	switch strings.TrimSpace(strings.ToLower(status)) {
	case StatusTrialing:
		return StatusTrialing
	case StatusActive:
		return StatusActive
	case StatusPastDue:
		return StatusPastDue
	case StatusCanceled:
		return StatusCanceled
	case StatusUnpaid:
		return StatusUnpaid
	case StatusIncomplete:
		return StatusIncomplete
	case StatusIncompleteExpired:
		return StatusIncompleteExpired
	case StatusPaused:
		return StatusPaused
	case StatusCheckoutIncomplete:
		return StatusCheckoutIncomplete
	default:
		return StatusInactive
	}
}

func GrantsAccess(status string) bool {
	switch Normalize(status) {
	case StatusActive, StatusTrialing:
		return true
	default:
		return false
	}
}
