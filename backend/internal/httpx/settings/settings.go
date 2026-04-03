package settings

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/settingsTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

func Get(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "get settings missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}
		cfg, err := settingsTx.Get(r.Context(), a.DB, accountID)
		if err != nil {
			slog.ErrorContext(r.Context(), "get settings failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}
		cfg.ReadOnly = userscope.Role(r.Context()) != "owner"

		res.JSON(w, http.StatusOK, cfg)
	}
}

func Put(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if userscope.Role(r.Context()) != "owner" {
			res.Error(w, http.StatusForbidden, "SETTINGS_OWNER_ONLY", "Only the workspace admin can edit settings")
			return
		}

		accountID, err := accountscope.Require(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "put settings missing account scope", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}
		var in models.Settings
		var raw map[string]json.RawMessage
		current, err := settingsTx.Get(r.Context(), a.DB, accountID)
		if err != nil {
			slog.ErrorContext(r.Context(), "load current settings failed", "err", err, "account_id", accountID)
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
		if _, ok := raw["logoUrl"]; ok {
			res.Validation(w, res.Invalid("logoUrl", "is read-only and must be managed via the logo endpoint"))
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
		in.ReadOnly = false
		in.LogoURL = current.LogoURL
		in.LogoAssetID = current.LogoAssetID
		in.LogoStorageKey = current.LogoStorageKey

		if err := settingsTx.Upsert(r.Context(), a.DB, accountID, in); err != nil {
			switch {
			case errors.Is(err, settingsTx.ErrStartingInvoiceNumberInvalid):
				res.Validation(w, res.Invalid("startingInvoiceNumber", "must be greater than 0"))
				return
			case errors.Is(err, settingsTx.ErrStartingInvoiceNumberLocked):
				res.Error(w, http.StatusConflict, "INVOICE_NUMBER_LOCKED", "Starting invoice number can only be changed when there are no invoices.")
				return
			}
			slog.ErrorContext(r.Context(), "save settings failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to save settings")
			return
		}

		cfg, err := settingsTx.Get(r.Context(), a.DB, accountID)
		if err != nil {
			slog.ErrorContext(r.Context(), "reload settings failed", "err", err, "account_id", accountID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load settings")
			return
		}
		cfg.ReadOnly = false

		res.JSON(w, http.StatusOK, cfg)
	}
}
