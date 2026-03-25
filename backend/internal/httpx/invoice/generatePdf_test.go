package invoice

import "testing"

func TestBuildPDFFilename(t *testing.T) {
	tests := []struct {
		name       string
		baseNumber int64
		revisionNo int64
		want       string
	}{
		{
			name:       "fallback when base is invalid",
			baseNumber: 0,
			revisionNo: 1,
			want:       "Invoice.pdf",
		},
		{
			name:       "base invoice has no revision suffix",
			baseNumber: 1,
			revisionNo: 1,
			want:       "Invoice-1.pdf",
		},
		{
			name:       "first visible revision maps from db revision 2",
			baseNumber: 1,
			revisionNo: 2,
			want:       "Invoice-1-Rev-1.pdf",
		},
		{
			name:       "higher revision keeps shifted suffix",
			baseNumber: 9,
			revisionNo: 5,
			want:       "Invoice-9-Rev-4.pdf",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildPDFFilename(tc.baseNumber, tc.revisionNo)
			if got != tc.want {
				t.Fatalf("buildPDFFilename() = %q, want %q", got, tc.want)
			}
		})
	}
}
