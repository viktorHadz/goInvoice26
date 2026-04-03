package billing

import (
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/billingcatalog"
	"github.com/viktorHadz/goInvoice26/internal/billingplan"
)

func TestDetermineSubscriptionSelection_PrefersMetadataWhenPriceIDChanges(t *testing.T) {
	selection := determineSubscriptionSelection(
		stripeSubscription{
			Metadata: map[string]string{
				"plan":     billingplan.PlanTeam,
				"interval": billingcatalog.IntervalYearly,
			},
			Items: struct {
				Data []struct {
					ID    string `json:"id"`
					Price struct {
						ID string `json:"id"`
					} `json:"price"`
				} `json:"data"`
			}{
				Data: []struct {
					ID    string `json:"id"`
					Price struct {
						ID string `json:"id"`
					} `json:"price"`
				}{
					{
						ID: "si_123",
						Price: struct {
							ID string `json:"id"`
						}{
							ID: "price_legacy_team_yearly",
						},
					},
				},
			},
		},
		billingcatalog.Config{
			SingleMonthlyPriceID: "price_single_monthly",
			SingleYearlyPriceID:  "price_single_yearly",
			TeamMonthlyPriceID:   "price_team_monthly",
			TeamYearlyPriceID:    "price_team_yearly",
		},
	)

	if selection.Plan != billingplan.PlanTeam {
		t.Fatalf("selection.Plan = %q, want %q", selection.Plan, billingplan.PlanTeam)
	}
	if selection.Interval != billingcatalog.IntervalYearly {
		t.Fatalf("selection.Interval = %q, want %q", selection.Interval, billingcatalog.IntervalYearly)
	}
}

func TestFormatStripePriceLabel(t *testing.T) {
	if got := formatStripePriceLabel("gbp", 500, billingcatalog.IntervalMonthly); got != "£5 / month" {
		t.Fatalf("formatStripePriceLabel() = %q, want %q", got, "£5 / month")
	}
}
