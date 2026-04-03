package billingplan

import "strings"

const (
	PlanSingle = "single"
	PlanTeam   = "team"

	SingleSeatLimit = 1
	TeamSeatLimit   = 5
)

func Normalize(plan string) string {
	switch strings.TrimSpace(strings.ToLower(plan)) {
	case PlanSingle:
		return PlanSingle
	case PlanTeam:
		return PlanTeam
	default:
		return ""
	}
}

func DetermineFromPriceID(priceID, singlePriceID, teamPriceID string) string {
	priceID = strings.TrimSpace(priceID)
	switch {
	case priceID != "" && priceID == strings.TrimSpace(teamPriceID):
		return PlanTeam
	case priceID != "" && priceID == strings.TrimSpace(singlePriceID):
		return PlanSingle
	default:
		return ""
	}
}

func SeatLimit(plan string) int {
	switch Normalize(plan) {
	case PlanSingle:
		return SingleSeatLimit
	case PlanTeam:
		return TeamSeatLimit
	default:
		return 0
	}
}

func SupportsTeam(plan string) bool {
	return Normalize(plan) == PlanTeam
}
