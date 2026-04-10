package invoiceformat

import "testing"

func TestFormatInvoiceNumber(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		baseNo     int64
		revisionNo int64
		want       string
	}{
		{
			name:       "base revision omits suffix",
			prefix:     "INV-",
			baseNo:     7,
			revisionNo: 1,
			want:       "INV-7",
		},
		{
			name:       "first saved revision keeps dotted suffix",
			prefix:     "INV-",
			baseNo:     7,
			revisionNo: 2,
			want:       "INV-7.2",
		},
		{
			name:       "higher revision stays dotted",
			prefix:     "INV-",
			baseNo:     7,
			revisionNo: 4,
			want:       "INV-7.4",
		},
		{
			name:       "empty prefix uses default",
			prefix:     "",
			baseNo:     7,
			revisionNo: 1,
			want:       "INV-7",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatInvoiceNumber(tc.prefix, tc.baseNo, tc.revisionNo)
			if got != tc.want {
				t.Fatalf("FormatInvoiceNumber() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestFormatPaymentReceiptNumber(t *testing.T) {
	tests := []struct {
		name       string
		prefix     string
		baseNo     int64
		revisionNo int64
		receiptNo  int64
		want       string
	}{
		{name: "base receipt", prefix: "INV-", baseNo: 7, revisionNo: 1, receiptNo: 1, want: "INV-7-PR-1"},
		{name: "revision receipt", prefix: "INV-", baseNo: 7, revisionNo: 2, receiptNo: 4, want: "INV-7.2-PR-4"},
		{name: "default prefix", prefix: "", baseNo: 7, revisionNo: 3, receiptNo: 2, want: "INV-7.3-PR-2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatPaymentReceiptNumber(tc.prefix, tc.baseNo, tc.revisionNo, tc.receiptNo)
			if got != tc.want {
				t.Fatalf("FormatPaymentReceiptNumber() = %q, want %q", got, tc.want)
			}
		})
	}
}
