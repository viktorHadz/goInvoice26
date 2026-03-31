package billing

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	billingsvc "github.com/viktorHadz/goInvoice26/internal/service/billing"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

type checkoutSyncRequest struct {
	SessionID string `json:"sessionId"`
}

func CreateCheckoutSession(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || principal.Role != "owner" {
			slog.WarnContext(r.Context(), "billing checkout rejected for non-admin user")
			res.Error(w, http.StatusForbidden, "BILLING_OWNER_ONLY", "Only the workspace admin can manage billing")
			return
		}

		link, err := a.Billing.CreateCheckoutSession(
			r.Context(),
			principal.AccountID,
			principal.AccountName,
			principal.Email,
		)
		if err != nil {
			handleBillingError(w, r, "create checkout session", err)
			return
		}

		res.JSON(w, http.StatusOK, link)
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
	default:
		slog.ErrorContext(r.Context(), action+" failed", "err", err, "accountID", accountscope.AccountID(r.Context()))
		res.Error(w, http.StatusBadGateway, "BILLING_PROVIDER_ERROR", "Stripe could not complete that request")
	}
}
