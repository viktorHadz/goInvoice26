package validate

import (
	"testing"
)

func TestValidateText(t *testing.T) {
	var tests = []struct {
		TName      string
		Field      string
		Input      string
		Expected   string
		Required   bool
		Min, Max   int
		SingleLine bool
		Trim       bool
		Pass       bool
	}{
		{
			TName: "Testing name base case", Field: "name", Input: "Viktor Veselinov Hadzhiyski", Required: true, SingleLine: true, Pass: true,
		},
		{
			TName: "Input Trim with assert", Field: "name", Input: "        Vik      H            ", Expected: "Vik      H", Required: true, SingleLine: true, Trim: true, Pass: true,
		},
		{
			TName: "Testing required with missing input", Field: "name", Required: true, SingleLine: true, Pass: false,
		},
		{
			TName: "Testing address with near limit string", Input: "37A Glenton Road, SE13 5RS, Lewisham, London, United Kingdom",
			Field: "address", Max: 70, SingleLine: true, Pass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.TName, func(t *testing.T) {
			out, errs := Text(
				tt.Input,
				TextRules{Field: tt.Field, Required: tt.Required, Min: tt.Min, Max: tt.Max, SingleLine: tt.SingleLine, Trim: tt.Trim},
			)
			if tt.Trim {
				if out != tt.Expected {
					t.Errorf(
						"Output normalisation not the same as Expected. Output(\"%v\") != Expected(\"%v\")", out, tt.Expected,
					)
				}
				t.Logf("Input after Trim: (%v) Expected: (%v)", out, tt.Expected)
			}

			if len(errs) > 0 && tt.Pass {
				t.Errorf("Field %v with input %v, failed when expected to pass. Errors: %v", tt.Field, tt.Input, errs)
			}
			if len(errs) <= 0 && !tt.Pass {
				t.Errorf("Field %v with input %v, passed when expected to fail.", tt.Field, tt.Input)
			}

		})
	}

}
