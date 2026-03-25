package settings

import (
	"encoding/json"
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

		if err := settingsTx.Upsert(r.Context(), a.DB, in); err != nil {
			slog.ErrorContext(r.Context(), "save settings failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to save settings")
			return
		}

		res.JSON(w, http.StatusOK, in)
	}
}
