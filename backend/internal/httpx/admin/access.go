package admin

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type directAccessGrantRequest struct {
	Email string `json:"email"`
	Plan  string `json:"plan"`
	Note  string `json:"note"`
}

type promoCodeRequest struct {
	Code         string `json:"code"`
	DurationDays int    `json:"durationDays"`
}

type promoCodeStatusRequest struct {
	Active bool `json:"active"`
}

func Overview(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requirePlatformAdmin(w, r, a); !ok {
			return
		}

		directGrants, err := accessTx.ListDirectAccessGrants(r.Context(), a.DB)
		if err != nil {
			slog.ErrorContext(r.Context(), "list direct access grants failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load direct access grants")
			return
		}

		promoCodes, err := accessTx.ListPromoCodes(r.Context(), a.DB)
		if err != nil {
			slog.ErrorContext(r.Context(), "list promo codes failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load promo codes")
			return
		}

		if directGrants == nil {
			directGrants = []models.DirectAccessGrant{}
		}
		if promoCodes == nil {
			promoCodes = []models.PromoCode{}
		}

		res.JSON(w, http.StatusOK, models.PlatformAccessOverview{
			DirectGrants: directGrants,
			PromoCodes:   promoCodes,
		})
	}
}

func CreateDirectAccessGrant(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePlatformAdmin(w, r, a)
		if !ok {
			return
		}

		var req directAccessGrantRequest
		if !res.DecodeJSON(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Email) == "" {
			res.Validation(w, res.Required("email"))
			return
		}

		grant, err := accessTx.CreateDirectAccessGrant(
			r.Context(),
			a.DB,
			req.Email,
			req.Plan,
			req.Note,
			principal.UserID,
		)
		if err != nil {
			switch {
			case errors.Is(err, accessTx.ErrInvalidEmail):
				res.Validation(w, res.Invalid("email", "must be a valid email address"))
				return
			case errors.Is(err, accessTx.ErrInvalidAccessPlan):
				res.Validation(w, res.Invalid("plan", "must be single or team"))
				return
			case errors.Is(err, accessTx.ErrDirectAccessGrantExists):
				res.Error(w, http.StatusConflict, "DIRECT_ACCESS_GRANT_EXISTS", "That email already has a direct access grant.")
				return
			default:
				slog.ErrorContext(r.Context(), "create direct access grant failed", "err", err)
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to create direct access grant")
				return
			}
		}

		res.JSON(w, http.StatusCreated, grant)
	}
}

func DeleteDirectAccessGrant(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requirePlatformAdmin(w, r, a); !ok {
			return
		}

		grantID, err := strconv.ParseInt(chi.URLParam(r, "grantID"), 10, 64)
		if err != nil || grantID <= 0 {
			res.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid direct access grant")
			return
		}

		deleted, err := accessTx.DeleteDirectAccessGrant(r.Context(), a.DB, grantID)
		if err != nil {
			slog.ErrorContext(r.Context(), "delete direct access grant failed", "err", err, "grantID", grantID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to delete direct access grant")
			return
		}
		if !deleted {
			res.NotFound(w, "Direct access grant not found")
			return
		}

		res.NoContent(w)
	}
}

func CreatePromoCode(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := requirePlatformAdmin(w, r, a)
		if !ok {
			return
		}

		var req promoCodeRequest
		if !res.DecodeJSON(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Code) == "" {
			res.Validation(w, res.Required("code"))
			return
		}
		if req.DurationDays <= 0 {
			res.Validation(w, res.Invalid("durationDays", "must be greater than zero"))
			return
		}

		promoCode, err := accessTx.CreatePromoCode(r.Context(), a.DB, req.Code, req.DurationDays, principal.UserID)
		if err != nil {
			switch {
			case errors.Is(err, accessTx.ErrInvalidPromoCode):
				res.Validation(w, res.Invalid("code", "use 3-64 letters, numbers, dashes, or underscores"))
				return
			case errors.Is(err, accessTx.ErrPromoCodeExists):
				res.Error(w, http.StatusConflict, "PROMO_CODE_EXISTS", "That promo code already exists.")
				return
			default:
				slog.ErrorContext(r.Context(), "create promo code failed", "err", err)
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to create promo code")
				return
			}
		}

		res.JSON(w, http.StatusCreated, promoCode)
	}
}

func UpdatePromoCodeStatus(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requirePlatformAdmin(w, r, a); !ok {
			return
		}

		promoCodeID, err := strconv.ParseInt(chi.URLParam(r, "promoCodeID"), 10, 64)
		if err != nil || promoCodeID <= 0 {
			res.Error(w, http.StatusBadRequest, "INVALID_ID", "Invalid promo code")
			return
		}

		var req promoCodeStatusRequest
		if !res.DecodeJSON(w, r, &req) {
			return
		}

		updated, err := accessTx.SetPromoCodeActive(r.Context(), a.DB, promoCodeID, req.Active)
		if err != nil {
			slog.ErrorContext(r.Context(), "update promo code status failed", "err", err, "promoCodeID", promoCodeID)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to update promo code")
			return
		}
		if !updated {
			res.NotFound(w, "Promo code not found")
			return
		}

		res.NoContent(w)
	}
}

func requirePlatformAdmin(w http.ResponseWriter, r *http.Request, a *app.App) (userscope.Principal, bool) {
	principal, ok := userscope.PrincipalFromContext(r.Context())
	if !ok || !a.Auth.IsPlatformAdminEmail(principal.Email) {
		res.Error(w, http.StatusForbidden, "PLATFORM_ADMIN_ONLY", "Only the platform admin can manage app access.")
		return userscope.Principal{}, false
	}

	return principal, true
}
