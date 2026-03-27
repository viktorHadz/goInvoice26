package invoice

import "testing"

func TestAllowedStatusTransition(t *testing.T) {
	tests := []struct {
		from  string
		to    string
		rules statusTransitionRules
		want  bool
	}{
		{from: "draft", to: "draft", want: false},
		{from: "draft", to: "issued", want: true},
		{from: "draft", to: "paid", want: true},
		{from: "draft", to: "void", want: false},
		{from: "issued", to: "paid", want: true},
		{from: "issued", to: "void", want: true},
		{from: "issued", to: "draft", rules: statusTransitionRules{CanReturnIssuedToDraft: true}, want: true},
		{from: "issued", to: "draft", rules: statusTransitionRules{CanReturnIssuedToDraft: false}, want: false},
		{from: "paid", to: "issued", rules: statusTransitionRules{CanReopenPaidToIssued: true}, want: true},
		{from: "paid", to: "issued", rules: statusTransitionRules{CanReopenPaidToIssued: false}, want: false},
		{from: "paid", to: "void", want: false},
		{from: "void", to: "issued", want: false},
		{from: "void", to: "void", want: false},
	}

	for _, tt := range tests {
		if got := allowedStatusTransition(tt.from, tt.to, tt.rules); got != tt.want {
			t.Fatalf("allowedStatusTransition(%q, %q) = %v, want %v", tt.from, tt.to, got, tt.want)
		}
	}
}
