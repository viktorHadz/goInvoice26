package billing

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/billingplan"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	billingsvc "github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/transaction/accessTx"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type checkoutSyncRequest struct {
	SessionID string `json:"sessionId"`
}

type billingPlanRequest struct {
	Plan     string `json:"plan"`
	Interval string `json:"interval"`
	Redirect string `json:"redirect,omitempty"`
}

type promoCodeRedeemRequest struct {
	Code string `json:"code"`
}

func PublicCatalog(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		catalog, err := a.Billing.PublicCatalog(r.Context())
		if err != nil {
			slog.ErrorContext(r.Context(), "load public billing catalog failed", "err", err)
			res.Error(w, http.StatusBadGateway, "BILLING_PROVIDER_ERROR", "Stripe could not load billing details")
			return
		}

		res.JSON(w, http.StatusOK, catalog)
	}
}

func CreateCheckoutSession(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != "owner" {
			slog.WarnContext(r.Context(), "billing checkout rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		var req billingPlanRequest
		if r.ContentLength > 0 && !res.DecodeJSON(w, r, &req) {
			return
		}

		link, err := a.Billing.CreateCheckoutSession(
			r.Context(),
			principal.AccountID,
			principal.AccountName,
			principal.Email,
			req.Plan,
			req.Interval,
			req.Redirect,
		)
		if err != nil {
			handleBillingError(w, r, "create checkout session", err)
			return
		}

		res.JSON(w, http.StatusOK, link)
	}
}

func ChangeSubscriptionPlan(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != authTx.UserRoleOwner {
			slog.WarnContext(r.Context(), "billing plan change rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		var req billingPlanRequest
		if !res.DecodeJSON(w, r, &req) {
			return
		}
		req.Plan = billingplan.Normalize(req.Plan)
		if req.Plan == "" {
			res.Validation(w, res.Invalid("plan", "must be single or team"))
			return
		}

		if req.Plan == billingplan.PlanSingle {
			seatUsage, err := authTx.GetTeamSeatUsage(r.Context(), a.DB, principal.AccountID)
			if err != nil {
				slog.ErrorContext(r.Context(), "load team seat usage before downgrade failed", "err", err, "accountID", principal.AccountID)
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to verify team access before changing the plan")
				return
			}
			if seatUsage.MemberCount+seatUsage.PendingInviteCount > billingplan.SingleSeatLimit {
				res.Error(
					w,
					http.StatusConflict,
					"BILLING_PLAN_DOWNGRADE_BLOCKED",
					"Remove extra teammates and pending invites before switching to the single-user plan.",
				)
				return
			}
		}

		if err := a.Billing.ChangeSubscriptionPlan(r.Context(), principal.AccountID, req.Plan, req.Interval); err != nil {
			handleBillingError(w, r, "change subscription plan", err)
			return
		}

		res.NoContent(w)
	}
}

func CreatePortalSession(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != "owner" {
			slog.WarnContext(r.Context(), "billing portal rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		link, err := a.Billing.CreatePortalSession(r.Context(), principal.AccountID)
		if err != nil {
			handleBillingError(w, r, "create billing portal session", err)
			return
		}

		res.JSON(w, http.StatusOK, link)
	}
}

func SyncCheckoutSession(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != "owner" {
			slog.WarnContext(r.Context(), "billing checkout sync rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		var req checkoutSyncRequest
		if !res.DecodeJSON(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.SessionID) == "" {
			res.Validation(w, res.Required("sessionId"))
			return
		}

		if err := a.Billing.SyncCheckoutSession(r.Context(), principal.AccountID, req.SessionID); err != nil {
			handleBillingError(w, r, "sync checkout session", err)
			return
		}

		res.NoContent(w)
	}
}

func RedeemPromoCode(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != authTx.UserRoleOwner {
			slog.WarnContext(r.Context(), "promo code redeem rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		var req promoCodeRedeemRequest
		if !res.DecodeJSON(w, r, &req) {
			return
		}
		if strings.TrimSpace(req.Code) == "" {
			res.Validation(w, res.Required("code"))
			return
		}

		result, err := accessTx.RedeemPromoCode(
			r.Context(),
			a.DB,
			principal.AccountID,
			principal.UserID,
			req.Code,
			time.Now(),
			a.AccessLedgerSecret,
			a.PromoRedemptionRetentionDays,
		)
		if err != nil {
			switch {
			case errors.Is(err, accessTx.ErrInvalidPromoCode):
				res.Validation(w, res.Invalid("code", "enter a valid promo code"))
				return
			case errors.Is(err, accessTx.ErrPromoCodeNotFound):
				res.Error(w, http.StatusNotFound, "PROMO_CODE_NOT_FOUND", "That promo code could not be found.")
				return
			case errors.Is(err, accessTx.ErrPromoCodeInactive):
				res.Error(w, http.StatusConflict, "PROMO_CODE_INACTIVE", "That promo code is no longer active.")
				return
			case errors.Is(err, accessTx.ErrPromoCodeAlreadyRedeemed):
				res.Error(w, http.StatusConflict, "PROMO_CODE_ALREADY_REDEEMED", "That promo code has already been used for this workspace.")
				return
			case errors.Is(err, accessTx.ErrAccessAlreadyGranted):
				res.Error(w, http.StatusConflict, "PROMO_ACCESS_ALREADY_ACTIVE", "This workspace already has active access.")
				return
			default:
				slog.ErrorContext(r.Context(), "redeem promo code failed", "err", err, "accountID", principal.AccountID)
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to redeem promo code")
				return
			}
		}

		res.JSON(w, http.StatusOK, result)
	}
}

func CancelSubscription(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != "owner" {
			slog.WarnContext(r.Context(), "billing cancellation rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		if err := a.Billing.CancelSubscriptionAtPeriodEnd(r.Context(), principal.AccountID); err != nil {
			switch err {
			case billingsvc.ErrSubscriptionNotFound:
				res.Error(w, http.StatusNotFound, "BILLING_SUBSCRIPTION_NOT_FOUND", "There is no active subscription to cancel")
				return
			default:
				handleBillingError(w, r, "cancel subscription", err)
				return
			}
		}

		res.NoContent(w)
	}
}

func StripeWebhook(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
		if err != nil {
			res.Error(w, http.StatusBadRequest, "BAD_JSON", "Invalid webhook payload")
			return
		}

		if err := a.Billing.HandleWebhook(r.Context(), payload, r.Header.Get("Stripe-Signature")); err != nil {
			switch {
			case err == billingsvc.ErrNotConfigured:
				res.Error(w, http.StatusServiceUnavailable, "BILLING_NOT_CONFIGURED", "Billing is not configured")
			case err == billingsvc.ErrWebhookSignature:
				res.Error(w, http.StatusBadRequest, "BILLING_WEBHOOK_INVALID", "Webhook signature could not be verified")
			default:
				slog.ErrorContext(r.Context(), "stripe webhook failed", "err", err)
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to process billing webhook")
			}
			return
		}

		res.NoContent(w)
	}
}

func handleBillingError(w http.ResponseWriter, r *http.Request, action string, err error) {
	switch err {
	case billingsvc.ErrNotConfigured:
		slog.WarnContext(r.Context(), action+" rejected because billing is not configured", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusServiceUnavailable, "BILLING_NOT_CONFIGURED", "Billing is not configured")
	case billingsvc.ErrCustomerNotFound:
		slog.WarnContext(r.Context(), action+" rejected because no stripe customer exists", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusNotFound, "BILLING_CUSTOMER_NOT_FOUND", "No billing customer exists for this account yet")
	case billingsvc.ErrCheckoutPending:
		slog.InfoContext(r.Context(), action+" is waiting for Stripe confirmation", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusConflict, "BILLING_CHECKOUT_PENDING", "Stripe is still confirming the subscription")
	case billingsvc.ErrInvalidCheckoutSync:
		slog.WarnContext(r.Context(), action+" rejected because checkout session does not belong to account", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusBadRequest, "BILLING_CHECKOUT_INVALID", "That checkout session does not belong to this account")
	case billingsvc.ErrInvalidPlan:
		slog.WarnContext(r.Context(), action+" rejected because billing plan is invalid", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusBadRequest, "BILLING_PLAN_INVALID", "That billing plan is not supported")
	case billingsvc.ErrInvalidInterval:
		slog.WarnContext(r.Context(), action+" rejected because billing interval is invalid", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusBadRequest, "BILLING_INTERVAL_INVALID", "That billing interval is not supported")
	case billingsvc.ErrPlanUnavailable:
		slog.WarnContext(r.Context(), action+" rejected because billing plan is not available", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusConflict, "BILLING_PLAN_UNAVAILABLE", "That billing selection is not available yet")
	case billingsvc.ErrPlanAlreadyActive:
		slog.InfoContext(r.Context(), action+" skipped because billing plan is already active", "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusConflict, "BILLING_PLAN_ALREADY_ACTIVE", "That billing selection is already active")
	default:
		slog.ErrorContext(r.Context(), action+" failed", "err", err, "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusBadGateway, "BILLING_PROVIDER_ERROR", "Stripe could not complete that request")
	}
}
