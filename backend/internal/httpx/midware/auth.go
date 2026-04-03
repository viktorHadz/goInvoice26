package midware

import (
	"net/http"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
	"github.com/viktorHadz/goInvoice26/internal/userscope"
)

func RequireAuth(a *app.App) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(a.Auth.SessionCookieName())
			if err != nil || cookie.Value == "" {
				res.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Please sign in to continue")
				return
			}

			principal, ok, err := a.Auth.ResolveSession(r.Context(), cookie.Value)
			if err != nil {
				res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to validate your session")
				return
			}
			if !ok {
				http.SetCookie(w, a.Auth.ClearSessionCookie())
				res.Error(w, http.StatusUnauthorized, "UNAUTHENTICATED", "Please sign in to continue")
				return
			}

			ctx := accountscope.WithAccountID(r.Context(), principal.AccountID)
			ctx = userscope.WithPrincipal(ctx, userscope.Principal{
				UserID:               principal.UserID,
				AccountID:            principal.AccountID,
				AccountName:          principal.AccountName,
				Email:                principal.User.Email,
				Role:                 principal.Role,
				Name:                 principal.User.Name,
				BillingStatus:        principal.Billing.Status,
				BillingPlan:          principal.Billing.Plan,
				BillingAccessGranted: principal.Billing.AccessGranted,
			})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireBillingAccess(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, ok := userscope.PrincipalFromContext(r.Context())
		if !ok || !principal.BillingAccessGranted {
			res.Error(
				w,
				http.StatusPaymentRequired,
				"SUBSCRIPTION_REQUIRED",
				"Active billing or a valid access grant is required to access the workspace",
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func RequireOwner(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if userscope.Role(r.Context()) != authTx.UserRoleOwner {
			res.Error(w, http.StatusForbidden, "FORBIDDEN", "Only the workspace admin can manage teammates")
			return
		}

		next.ServeHTTP(w, r)
	})
}
