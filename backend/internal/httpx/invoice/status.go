package invoice

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/params"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
)

type invoiceStatusBody struct {
	Status string `json:"status"`
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

		var current string
		err := a.DB.QueryRowContext(r.Context(), `
			SELECT i.status
			FROM invoices i
			WHERE i.client_id = ? AND i.base_number = ?
		`, clientID, baseNumber).Scan(&current)
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
		if !allowedStatusTransition(current, next) {
			res.Validation(w, res.Invalid("status", "transition not allowed from current status"))
			return
		}

		resExec, err := a.DB.ExecContext(r.Context(), `
			UPDATE invoices
			SET status = ?
			WHERE client_id = ? AND base_number = ?
		`, next, clientID, baseNumber)
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

func allowedStatusTransition(from, to string) bool {
	switch from {
	case "draft":
		return to == "issued" || to == "void" || to == "paid"
	case "issued":
		return to == "void" || to == "paid"
	case "paid":
		return to == "issued"
	case "void":
		return to == "issued"
	default:
		return false
	}
}
