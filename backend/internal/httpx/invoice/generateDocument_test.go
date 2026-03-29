package invoice

import "testing"

func TestBuildDocumentFilename(t *testing.T) {
	tests := []struct {
		name       string
		baseNumber int64
		revisionNo int64
		ext        string
		want       string
	}{
		{
			name:       "fallback when base is invalid",
			baseNumber: 0,
			revisionNo: 1,
			ext:        "pdf",
			want:       "Invoice.pdf",
		},
		{
			name:       "base invoice uses base number",
			baseNumber: 12,
			revisionNo: 1,
			ext:        "docx",
			want:       "Invoice-12.docx",
		},
		{
			name:       "revision uses display suffix",
			baseNumber: 12,
			revisionNo: 4,
			ext:        ".pdf",
			want:       "Invoice-12-Rev-3.pdf",
		},
		{
			name:       "blank extension falls back",
			baseNumber: 5,
			revisionNo: 1,
			ext:        "",
			want:       "Invoice-5.bin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildDocumentFilename(tc.baseNumber, tc.revisionNo, tc.ext)
			if got != tc.want {
				t.Fatalf("buildDocumentFilename() = %q, want %q", got, tc.want)
			}
		})
	}
}
