package invoice

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

type invoiceStatusBody struct {
	Status string `json:"status"`
}

type statusTransitionRules struct {
	CanReturnIssuedToDraft bool
	CanReopenPaidToIssued  bool
}

// PatchInvoiceStatus updates invoices.status with allowed transitions only.
func PatchInvoiceStatus(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientID, ok := params.ValidateParam(w, r, "clientID")
		if !ok {
			return
		}
		baseNumber, ok := params.ValidateParam(w, r, "baseNumber")
		if !ok {
			return
		}

		var body invoiceStatusBody
		if ok := res.DecodeJSON(w, r, &body); !ok {
			return
		}

		next := strings.TrimSpace(strings.ToLower(body.Status))
		if next == "" {
			res.Validation(w, res.Required("status"))
			return
		}

		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "patch invoice status missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		var (
			current       string
			rules         statusTransitionRules
			revisionCount int64
			totalMinor    int64
			depositMinor  int64
			paidMinor     int64
		)
		err = a.DB.QueryRowContext(r.Context(), `
			SELECT
				i.status,
				COUNT(DISTINCT rev.id) AS revision_count,
				cur.total_minor,
				cur.deposit_minor,
				COALESCE((SELECT SUM(p.amount_minor) FROM payments p WHERE p.invoice_id = i.id), 0) AS paid_minor
			FROM invoices i
			JOIN invoice_revisions cur
				ON cur.id = i.current_revision_id
			LEFT JOIN invoice_revisions rev
				ON rev.invoice_id = i.id
			WHERE i.account_id = ? AND i.client_id = ? AND i.base_number = ?
			GROUP BY i.id, i.status, cur.total_minor, cur.deposit_minor
		`, accountID, clientID, baseNumber).Scan(&current, &revisionCount, &totalMinor, &depositMinor, &paidMinor)
		if errors.Is(err, sql.ErrNoRows) {
			res.Error(w, http.StatusNotFound, "NOT_FOUND", "Invoice not found")
			return
		}
		if err != nil {
			slog.ErrorContext(r.Context(), "patch invoice status load failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}

		current = strings.TrimSpace(strings.ToLower(current))
		rules = statusTransitionRules{
			CanReturnIssuedToDraft: revisionCount <= 1,
			CanReopenPaidToIssued:  paidMinor != expectedPaidMinor(totalMinor, depositMinor),
		}
		if !allowedStatusTransition(current, next, rules) {
			res.Validation(w, res.Invalid("status", invalidStatusTransitionMessage(current, next, rules)))
			return
		}

		resExec, err := a.DB.ExecContext(r.Context(), `
			UPDATE invoices
			SET status = ?
			WHERE account_id = ? AND client_id = ? AND base_number = ?
		`, next, accountID, clientID, baseNumber)
		if err != nil {
			slog.ErrorContext(r.Context(), "patch invoice status update failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "DATABASE_ERROR", "Database error")
			return
		}
		n, _ := resExec.RowsAffected()
		if n == 0 {
			res.Error(w, http.StatusNotFound, "NOT_FOUND", "Invoice not found")
			return
		}

		res.JSON(w, http.StatusOK, map[string]any{"status": next})
	}
}

func expectedPaidMinor(totalMinor, depositMinor int64) int64 {
	expected := totalMinor - depositMinor
	if expected < 0 {
		return 0
	}
	return expected
}

func invalidStatusTransitionMessage(from, to string, rules statusTransitionRules) string {
	switch {
	case from == to:
		return "status is already set to " + to
	case from == "issued" && to == "draft" && !rules.CanReturnIssuedToDraft:
		return "issued invoices with saved revisions cannot return to draft"
	case from == "paid" && to == "issued" && !rules.CanReopenPaidToIssued:
		return "fully paid invoices cannot return to issued"
	default:
		return "transition not allowed from current status"
	}
}

func allowedStatusTransition(from, to string, rules statusTransitionRules) bool {
	if from == to {
		return false
	}

	switch from {
	case "draft":
		return to == "issued"
	case "issued":
		if to == "draft" {
			return rules.CanReturnIssuedToDraft
		}
		return to == "void" || to == "paid"
	case "paid":
		return to == "issued" && rules.CanReopenPaidToIssued
	case "void":
		return false
	default:
		return false
	}
}
