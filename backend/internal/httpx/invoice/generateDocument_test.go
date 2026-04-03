package invoice

import (
	"testing"

	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

func hasFieldError(errs []res.FieldError, field string) bool {
	for _, err := range errs {
		if err.Field == field {
			return true
		}
	}
	return false
}

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

func TestValidateQuickDocumentInvoice_AcceptsValidPayload(t *testing.T) {
	in := validInvoiceInput()

	got, errs := validateQuickDocumentInvoice(in, in.Overview.ClientID, in.Overview.BaseNumber)
	if len(errs) > 0 {
		t.Fatalf("validateQuickDocumentInvoice() returned errs: %#v", errs)
	}

	if got.Totals.TotalMinor != 12000 {
		t.Fatalf("totalMinor = %d, want 12000", got.Totals.TotalMinor)
	}
}

func TestValidateQuickDocumentInvoice_RejectsInvalidLineStructure(t *testing.T) {
	in := validInvoiceInput()
	in.Lines = nil

	_, errs := validateQuickDocumentInvoice(in, in.Overview.ClientID, in.Overview.BaseNumber)
	if !hasFieldError(errs, "lines") {
		t.Fatalf("expected lines validation error, got %#v", errs)
	}
}

func TestValidateQuickDocumentInvoice_RejectsSortOrderZero(t *testing.T) {
	in := validInvoiceInput()
	in.Lines[0].SortOrder = 0

	_, errs := validateQuickDocumentInvoice(in, in.Overview.ClientID, in.Overview.BaseNumber)
	if !hasFieldError(errs, "lines[0].sortOrder") {
		t.Fatalf("expected sortOrder validation error, got %#v", errs)
	}
}
