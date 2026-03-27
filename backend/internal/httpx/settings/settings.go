package settings

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
)

func Get(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg, err := settingsTx.Get(r.Context(), a.DB)
		if err != nil {
			slog.ErrorContext(r.Context(), "get settings failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}

		res.JSON(w, http.StatusOK, cfg)
	}
}

func Put(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in models.Settings
		var raw map[string]json.RawMessage
		current, err := settingsTx.Get(r.Context(), a.DB)
		if err != nil {
			slog.ErrorContext(r.Context(), "load current settings failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			slog.ErrorContext(r.Context(), "read settings body failed", "err", err)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Invalid request body")
			return
		}

		if err := json.Unmarshal(body, &in); err != nil {
			slog.ErrorContext(r.Context(), "decode settings failed", "err", err)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Invalid request body")
			return
		}
		if err := json.Unmarshal(body, &raw); err != nil {
			slog.ErrorContext(r.Context(), "decode settings raw failed", "err", err)
			res.Error(w, http.StatusBadRequest, "BAD_DATA", "Invalid request body")
			return
		}

		// tiny defaults / sanitise
		if in.InvoicePrefix == "" {
			in.InvoicePrefix = "INV-"
		}
		if in.Currency == "" {
			in.Currency = "GBP"
		}
		if in.DateFormat == "" {
			in.DateFormat = "dd/mm/yyyy"
		}
		if _, ok := raw["showItemTypeHeaders"]; !ok {
			in.ShowItemTypeHeaders = true
		}
		if _, ok := raw["startingInvoiceNumber"]; !ok {
			in.StartingInvoiceNumber = current.StartingInvoiceNumber
		}
		if in.StartingInvoiceNumber < 1 {
			res.Validation(w, res.Invalid("startingInvoiceNumber", "must be greater than 0"))
			return
		}
		in.CanEditStartingInvoiceNumber = current.CanEditStartingInvoiceNumber

		if err := settingsTx.Upsert(r.Context(), a.DB, in); err != nil {
			switch {
			case errors.Is(err, settingsTx.ErrStartingInvoiceNumberInvalid):
				res.Validation(w, res.Invalid("startingInvoiceNumber", "must be greater than 0"))
				return
			case errors.Is(err, settingsTx.ErrStartingInvoiceNumberLocked):
				res.Error(w, http.StatusConflict, "INVOICE_NUMBER_LOCKED", "Starting invoice number can only be changed when there are no invoices.")
				return
			}
			slog.ErrorContext(r.Context(), "save settings failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to save settings")
			return
		}

		cfg, err := settingsTx.Get(r.Context(), a.DB)
		if err != nil {
			slog.ErrorContext(r.Context(), "reload settings failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}

		res.JSON(w, http.StatusOK, cfg)
	}
}
