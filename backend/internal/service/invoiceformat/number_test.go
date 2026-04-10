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
			name:       "first user-visible revision maps from db revision 2",
			prefix:     "INV-",
			baseNo:     7,
			revisionNo: 2,
			want:       "INV-7-Rev-1",
		},
		{
			name:       "higher revision remains shifted",
			prefix:     "INV-",
			baseNo:     7,
			revisionNo: 4,
			want:       "INV-7-Rev-3",
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
		name      string
		prefix    string
		baseNo    int64
		receiptNo int64
		want      string
	}{
		{name: "first receipt", prefix: "INV-", baseNo: 7, receiptNo: 1, want: "INV-7-PR-1"},
		{name: "later receipt", prefix: "INV-", baseNo: 7, receiptNo: 4, want: "INV-7-PR-4"},
		{name: "default prefix", prefix: "", baseNo: 7, receiptNo: 2, want: "INV-7-PR-2"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := FormatPaymentReceiptNumber(tc.prefix, tc.baseNo, tc.receiptNo)
			if got != tc.want {
				t.Fatalf("FormatPaymentReceiptNumber() = %q, want %q", got, tc.want)
			}
		})
	}
}
