package billingcatalog

import (
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/billingplan"
)

const (
	IntervalMonthly = "monthly"
	IntervalYearly  = "yearly"
)

type Config struct {
	SingleMonthlyPriceID string
	SingleYearlyPriceID  string
	TeamMonthlyPriceID   string
	TeamYearlyPriceID    string
}

type Selection struct {
	Plan     string
	Interval string
}

func NormalizeInterval(interval string) string {
	switch strings.TrimSpace(strings.ToLower(interval)) {
	case IntervalMonthly, "month":
		return IntervalMonthly
	case IntervalYearly, "year", "annual", "annually":
		return IntervalYearly
	default:
		return ""
	}
}

func NormalizeSelection(plan, interval string) Selection {
	return Selection{
		Plan:     billingplan.Normalize(plan),
		Interval: NormalizeInterval(interval),
	}
}

func DetermineFromPriceID(priceID string, cfg Config) Selection {
	priceID = strings.TrimSpace(priceID)
	switch {
	case priceID != "" && priceID == strings.TrimSpace(cfg.SingleMonthlyPriceID):
		return Selection{Plan: billingplan.PlanSingle, Interval: IntervalMonthly}
	case priceID != "" && priceID == strings.TrimSpace(cfg.SingleYearlyPriceID):
		return Selection{Plan: billingplan.PlanSingle, Interval: IntervalYearly}
	case priceID != "" && priceID == strings.TrimSpace(cfg.TeamMonthlyPriceID):
		return Selection{Plan: billingplan.PlanTeam, Interval: IntervalMonthly}
	case priceID != "" && priceID == strings.TrimSpace(cfg.TeamYearlyPriceID):
		return Selection{Plan: billingplan.PlanTeam, Interval: IntervalYearly}
	default:
		return Selection{}
	}
}

func PriceIDFor(plan, interval string, cfg Config) string {
	switch NormalizeSelection(plan, interval) {
	case Selection{Plan: billingplan.PlanSingle, Interval: IntervalMonthly}:
		return strings.TrimSpace(cfg.SingleMonthlyPriceID)
	case Selection{Plan: billingplan.PlanSingle, Interval: IntervalYearly}:
		return strings.TrimSpace(cfg.SingleYearlyPriceID)
	case Selection{Plan: billingplan.PlanTeam, Interval: IntervalMonthly}:
		return strings.TrimSpace(cfg.TeamMonthlyPriceID)
	case Selection{Plan: billingplan.PlanTeam, Interval: IntervalYearly}:
		return strings.TrimSpace(cfg.TeamYearlyPriceID)
	default:
		return ""
	}
}

func PlanAvailable(plan string, cfg Config) bool {
	return DefaultIntervalForPlan(plan, cfg) != ""
}

func IntervalAvailable(plan, interval string, cfg Config) bool {
	return PriceIDFor(plan, interval, cfg) != ""
}

func AnyConfigured(cfg Config) bool {
	return strings.TrimSpace(cfg.SingleMonthlyPriceID) != "" ||
		strings.TrimSpace(cfg.SingleYearlyPriceID) != "" ||
		strings.TrimSpace(cfg.TeamMonthlyPriceID) != "" ||
		strings.TrimSpace(cfg.TeamYearlyPriceID) != ""
}

func DefaultPlan(cfg Config) string {
	switch {
	case PlanAvailable(billingplan.PlanSingle, cfg):
		return billingplan.PlanSingle
	case PlanAvailable(billingplan.PlanTeam, cfg):
		return billingplan.PlanTeam
	default:
		return ""
	}
}

func DefaultIntervalForPlan(plan string, cfg Config) string {
	switch billingplan.Normalize(plan) {
	case billingplan.PlanSingle:
		switch {
		case strings.TrimSpace(cfg.SingleMonthlyPriceID) != "":
			return IntervalMonthly
		case strings.TrimSpace(cfg.SingleYearlyPriceID) != "":
			return IntervalYearly
		default:
			return ""
		}
	case billingplan.PlanTeam:
		switch {
		case strings.TrimSpace(cfg.TeamMonthlyPriceID) != "":
			return IntervalMonthly
		case strings.TrimSpace(cfg.TeamYearlyPriceID) != "":
			return IntervalYearly
		default:
			return ""
		}
	default:
		return ""
	}
}

func DefaultCheckoutSelection(plan, interval string, cfg Config) Selection {
	selection := NormalizeSelection(plan, interval)
	if selection.Plan == "" {
		selection.Plan = DefaultPlan(cfg)
	}
	if selection.Interval == "" {
		selection.Interval = DefaultIntervalForPlan(selection.Plan, cfg)
	}
	return selection
}
