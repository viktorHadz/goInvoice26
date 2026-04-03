package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/viktorHadz/goInvoice26/internal/app"
	"github.com/viktorHadz/goInvoice26/internal/httpx/res"
	authsvc "github.com/viktorHadz/goInvoice26/internal/service/auth"
)

func Me(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status, clearCookie, err := a.Auth.Status(r.Context(), readSessionToken(r, a))
		if err != nil {
			slog.ErrorContext(r.Context(), "auth status failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to load auth status")
			return
		}
		if clearCookie {
			http.SetCookie(w, a.Auth.ClearSessionCookie())
		}

		res.JSON(w, http.StatusOK, status)
	}
}

func Logout(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Auth.Logout(r.Context(), readSessionToken(r, a)); err != nil {
			slog.ErrorContext(r.Context(), "logout failed", "err", err)
			res.Error(w, http.StatusInternalServerError, "INTERNAL", "Failed to log out")
			return
		}

		http.SetCookie(w, a.Auth.ClearSessionCookie())
		res.NoContent(w)
	}
}

func GoogleStart(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode")
		redirectPath := r.URL.Query().Get("redirect")

		authURL, state, err := a.Auth.StartGoogleAuth(mode, redirectPath)
		if err != nil {
			slog.WarnContext(r.Context(), "google auth start rejected", "err", err, "mode", mode)
			http.Redirect(w, r, redirectErrorURL(a, mode, redirectPath, startErrorCode(err)), http.StatusFound)
			return
		}

		cookie, err := a.Auth.OAuthStateCookie(state)
		if err != nil {
			slog.ErrorContext(r.Context(), "create oauth state cookie failed", "err", err)
			http.Redirect(w, r, redirectErrorURL(a, state.Mode, state.Redirect, "oauth_state_failed"), http.StatusFound)
			return
		}

		http.SetCookie(w, cookie)
		http.Redirect(w, r, authURL, http.StatusFound)
	}
}

func GoogleCallback(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stateCookie, err := r.Cookie("invoicer_google_oauth")
		if err != nil {
			http.Redirect(w, r, redirectErrorURL(a, authsvc.ModeLogin, "/app", "invalid_oauth_state"), http.StatusFound)
			return
		}

		oauthState, err := a.Auth.ParseOAuthStateCookie(stateCookie.Value)
		if err != nil || oauthState.State != r.URL.Query().Get("state") {
			http.SetCookie(w, a.Auth.ClearOAuthStateCookie())
			http.Redirect(w, r, redirectErrorURL(a, oauthState.Mode, oauthState.Redirect, "invalid_oauth_state"), http.StatusFound)
			return
		}
		http.SetCookie(w, a.Auth.ClearOAuthStateCookie())

		if googleErr := r.URL.Query().Get("error"); googleErr != "" {
			http.Redirect(w, r, redirectErrorURL(a, oauthState.Mode, oauthState.Redirect, "google_"+googleErr), http.StatusFound)
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			http.Redirect(w, r, redirectErrorURL(a, oauthState.Mode, oauthState.Redirect, "missing_oauth_code"), http.StatusFound)
			return
		}

		principal, sessionToken, err := a.Auth.AuthenticateWithGoogle(r.Context(), oauthState.Mode, code)
		if err != nil {
			slog.WarnContext(r.Context(), "google auth callback failed", "err", err, "mode", oauthState.Mode)
			http.Redirect(w, r, redirectErrorURL(a, oauthState.Mode, oauthState.Redirect, callbackErrorCode(err)), http.StatusFound)
			return
		}

		http.SetCookie(w, a.Auth.SessionCookie(sessionToken, principal.ExpiresAt))
		http.Redirect(w, r, a.Auth.AppURL(oauthState.Redirect), http.StatusFound)
	}
}

func readSessionToken(r *http.Request, a *app.App) string {
	cookie, err := r.Cookie(a.Auth.SessionCookieName())
	if err != nil {
		return ""
	}

	return cookie.Value
}

func redirectErrorURL(a *app.App, mode, redirectPath, code string) string {
	path := "/login"
	if mode == authsvc.ModeSignup {
		path = "/signup"
	}

	target := a.Auth.AppURL(path)
	parsed, err := url.Parse(target)
	if err != nil {
		return target
	}

	query := parsed.Query()
	if code != "" {
		query.Set("error", code)
	}
	if redirectPath != "" {
		query.Set("redirect", redirectPath)
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func startErrorCode(err error) string {
	switch {
	case errors.Is(err, authsvc.ErrGoogleNotConfigured):
		return "google_not_configured"
	case errors.Is(err, authsvc.ErrInvalidMode):
		return "invalid_auth_mode"
	default:
		return "auth_start_failed"
	}
}

func callbackErrorCode(err error) string {
	switch {
	case errors.Is(err, authsvc.ErrGoogleNotConfigured):
		return "google_not_configured"
	case errors.Is(err, authsvc.ErrInvalidMode):
		return "invalid_auth_mode"
	case errors.Is(err, authsvc.ErrEmailNotVerified):
		return "google_email_not_verified"
	case errors.Is(err, authsvc.ErrAccountNotLinked):
		return "account_not_linked"
	case errors.Is(err, authsvc.ErrGoogleSubConflict):
		return "account_conflict"
	default:
		return "google_auth_failed"
	}
}
