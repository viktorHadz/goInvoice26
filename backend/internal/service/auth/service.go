package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/viktorHadz/goInvoice26/internal/accountscope"
	"github.com/viktorHadz/goInvoice26/internal/models"
	"github.com/viktorHadz/goInvoice26/internal/transaction/authTx"
)

const (
	ModeSignup = "signup"
	ModeLogin  = "login"
)

var (
	ErrGoogleNotConfigured = errors.New("google oauth not configured")
	ErrInvalidMode         = errors.New("invalid auth mode")
	ErrInvalidState        = errors.New("invalid oauth state")
	ErrEmailNotVerified    = errors.New("google email not verified")
	ErrAccountNotLinked    = errors.New("google account is not linked")
	ErrGoogleSubConflict   = errors.New("google account is linked to another user")
)

type Config struct {
	AppBaseURL         string
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	SessionCookieName  string
	SecureCookies      bool
	SessionTTL         time.Duration
	HTTPClient         *http.Client
}

type Service struct {
	db                 *sql.DB
	appBaseURL         string
	googleClientID     string
	googleClientSecret string
	googleRedirectURL  string
	sessionCookieName  string
	secureCookies      bool
	sessionTTL         time.Duration
	httpClient         *http.Client
}

type SessionPrincipal struct {
	UserID      int64
	AccountID   int64
	AccountName string
	Role        string
	ExpiresAt   time.Time
	User        models.AuthUser
	Account     models.AuthAccount
}

type OAuthState struct {
	State    string `json:"state"`
	Mode     string `json:"mode"`
	Redirect string `json:"redirect"`
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type GoogleProfile struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func NewService(db *sql.DB, cfg Config) *Service {
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}

	sessionTTL := cfg.SessionTTL
	if sessionTTL <= 0 {
		sessionTTL = 30 * 24 * time.Hour
	}

	cookieName := strings.TrimSpace(cfg.SessionCookieName)
	if cookieName == "" {
		cookieName = "invoicer_session"
	}

	return &Service{
		db:                 db,
		appBaseURL:         strings.TrimRight(strings.TrimSpace(cfg.AppBaseURL), "/"),
		googleClientID:     strings.TrimSpace(cfg.GoogleClientID),
		googleClientSecret: strings.TrimSpace(cfg.GoogleClientSecret),
		googleRedirectURL:  strings.TrimSpace(cfg.GoogleRedirectURL),
		sessionCookieName:  cookieName,
		secureCookies:      cfg.SecureCookies,
		sessionTTL:         sessionTTL,
		httpClient:         httpClient,
	}
}

func (s *Service) GoogleEnabled() bool {
	return s.googleClientID != "" && s.googleClientSecret != "" && s.googleRedirectURL != ""
}

func (s *Service) SessionCookieName() string {
	return s.sessionCookieName
}

func (s *Service) SetupRequired(ctx context.Context) (bool, error) {
	return authTx.SetupRequired(ctx, s.db)
}

func (s *Service) Status(ctx context.Context, sessionToken string) (models.AuthStatus, bool, error) {
	setupRequired, err := s.SetupRequired(ctx)
	if err != nil {
		return models.AuthStatus{}, false, err
	}

	status := models.AuthStatus{
		Authenticated: false,
		NeedsSetup:    setupRequired,
		GoogleEnabled: s.GoogleEnabled(),
	}

	if strings.TrimSpace(sessionToken) == "" {
		return status, false, nil
	}

	session, ok, err := s.ResolveSession(ctx, sessionToken)
	if err != nil {
		return models.AuthStatus{}, false, err
	}
	if !ok {
		return status, true, nil
	}

	status.Authenticated = true
	status.User = &session.User
	status.Account = &session.Account

	return status, false, nil
}

func (s *Service) ResolveSession(ctx context.Context, sessionToken string) (SessionPrincipal, bool, error) {
	tokenHash := hashToken(sessionToken)
	if tokenHash == "" {
		return SessionPrincipal{}, false, nil
	}

	session, ok, err := authTx.GetSessionByTokenHash(ctx, s.db, tokenHash, time.Now())
	if err != nil {
		return SessionPrincipal{}, false, err
	}
	if !ok {
		return SessionPrincipal{}, false, nil
	}

	if err := authTx.TouchSession(ctx, s.db, session.ID, time.Now()); err != nil {
		return SessionPrincipal{}, false, err
	}

	principal := SessionPrincipal{
		UserID:      session.UserID,
		AccountID:   session.AccountID,
		AccountName: session.AccountName,
		Role:        session.User.Role,
		ExpiresAt:   session.ExpiresAt,
		User: models.AuthUser{
			ID:        session.User.ID,
			Name:      session.User.Name,
			Email:     session.User.Email,
			AvatarURL: session.User.AvatarURL,
			Role:      session.User.Role,
		},
		Account: models.AuthAccount{
			ID:   session.AccountID,
			Name: session.AccountName,
		},
	}

	return principal, true, nil
}

func (s *Service) StartGoogleAuth(mode, redirectPath string) (string, OAuthState, error) {
	if !s.GoogleEnabled() {
		return "", OAuthState{}, ErrGoogleNotConfigured
	}

	mode = normalizeMode(mode)
	if mode == "" {
		return "", OAuthState{}, ErrInvalidMode
	}

	state, err := randomToken(24)
	if err != nil {
		return "", OAuthState{}, fmt.Errorf("generate oauth state: %w", err)
	}

	redirectPath = sanitizeRedirectPath(redirectPath)
	params := url.Values{}
	params.Set("client_id", s.googleClientID)
	params.Set("redirect_uri", s.googleRedirectURL)
	params.Set("response_type", "code")
	params.Set("scope", "openid email profile")
	params.Set("state", state)
	params.Set("prompt", "select_account")

	authURL := "https://accounts.google.com/o/oauth2/v2/auth?" + params.Encode()

	return authURL, OAuthState{
		State:    state,
		Mode:     mode,
		Redirect: redirectPath,
	}, nil
}

func (s *Service) AuthenticateWithGoogle(ctx context.Context, mode, code string) (SessionPrincipal, string, error) {
	mode = normalizeMode(mode)
	if mode == "" {
		return SessionPrincipal{}, "", ErrInvalidMode
	}
	if !s.GoogleEnabled() {
		return SessionPrincipal{}, "", ErrGoogleNotConfigured
	}

	profile, err := s.fetchGoogleProfile(ctx, code)
	if err != nil {
		return SessionPrincipal{}, "", err
	}
	if !profile.EmailVerified {
		return SessionPrincipal{}, "", ErrEmailNotVerified
	}

	user, ok, err := authTx.GetUserByGoogleSub(ctx, s.db, profile.Sub)
	if err != nil {
		return SessionPrincipal{}, "", err
	}
	if ok {
		if err := authTx.UpdateUserProfile(ctx, s.db, user.ID, profile.Name, profile.Picture); err != nil {
			return SessionPrincipal{}, "", err
		}
		return s.createSessionForUser(ctx, user.ID)
	}

	user, ok, err = authTx.GetUserByEmail(ctx, s.db, profile.Email)
	if err != nil {
		return SessionPrincipal{}, "", err
	}
	if ok {
		if user.GoogleSub != "" && user.GoogleSub != profile.Sub {
			return SessionPrincipal{}, "", ErrGoogleSubConflict
		}
		if err := authTx.UpdateGoogleIdentity(ctx, s.db, user.ID, profile.Sub, profile.Name, profile.Picture); err != nil {
			return SessionPrincipal{}, "", err
		}
		return s.createSessionForUser(ctx, user.ID)
	}

	if mode == ModeSignup {
		setupRequired, err := s.SetupRequired(ctx)
		if err != nil {
			return SessionPrincipal{}, "", err
		}
		if !setupRequired {
			return SessionPrincipal{}, "", authTx.ErrSetupAlreadyComplete
		}

		user, err := authTx.CreateInitialOwner(ctx, s.db, authTx.CreateGoogleUserParams{
			Name:      profile.Name,
			Email:     profile.Email,
			GoogleSub: profile.Sub,
			AvatarURL: profile.Picture,
			Role:      authTx.UserRoleOwner,
			AccountID: accountscope.DefaultAccountID,
		})
		if err != nil {
			return SessionPrincipal{}, "", err
		}

		return s.createSessionForUser(ctx, user.ID)
	}

	accountID, allowed, err := authTx.AllowedAccountIDForEmail(ctx, s.db, profile.Email)
	if err != nil {
		return SessionPrincipal{}, "", err
	}
	if allowed {
		user, err := authTx.CreateMemberFromGoogle(ctx, s.db, authTx.CreateGoogleUserParams{
			Name:      profile.Name,
			Email:     profile.Email,
			GoogleSub: profile.Sub,
			AvatarURL: profile.Picture,
			Role:      authTx.UserRoleMember,
			AccountID: accountID,
		})
		if err != nil {
			return SessionPrincipal{}, "", err
		}
		if err := authTx.DeleteInviteByEmail(ctx, s.db, accountID, profile.Email); err != nil {
			return SessionPrincipal{}, "", err
		}

		return s.createSessionForUser(ctx, user.ID)
	}

	setupRequired, err := s.SetupRequired(ctx)
	if err != nil {
		return SessionPrincipal{}, "", err
	}
	if setupRequired {
		return SessionPrincipal{}, "", authTx.ErrSetupRequired
	}

	return SessionPrincipal{}, "", ErrAccountNotLinked
}

func (s *Service) Logout(ctx context.Context, sessionToken string) error {
	if strings.TrimSpace(sessionToken) == "" {
		return nil
	}

	return authTx.DeleteSessionByTokenHash(ctx, s.db, hashToken(sessionToken))
}

func (s *Service) SessionCookie(token string, expiresAt time.Time) *http.Cookie {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	return &http.Cookie{
		Name:     s.sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
		MaxAge:   maxAge,
		Expires:  expiresAt.UTC(),
	}
}

func (s *Service) ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     s.sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	}
}

func (s *Service) OAuthStateCookie(state OAuthState) (*http.Cookie, error) {
	payload, err := json.Marshal(state)
	if err != nil {
		return nil, fmt.Errorf("marshal oauth state: %w", err)
	}

	return &http.Cookie{
		Name:     "invoicer_google_oauth",
		Value:    base64.RawURLEncoding.EncodeToString(payload),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
		MaxAge:   int((10 * time.Minute).Seconds()),
		Expires:  time.Now().Add(10 * time.Minute).UTC(),
	}, nil
}

func (s *Service) ClearOAuthStateCookie() *http.Cookie {
	return &http.Cookie{
		Name:     "invoicer_google_oauth",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   s.secureCookies,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	}
}

func (s *Service) ParseOAuthStateCookie(raw string) (OAuthState, error) {
	if strings.TrimSpace(raw) == "" {
		return OAuthState{}, ErrInvalidState
	}

	payload, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return OAuthState{}, fmt.Errorf("decode oauth state cookie: %w", err)
	}

	var state OAuthState
	if err := json.Unmarshal(payload, &state); err != nil {
		return OAuthState{}, fmt.Errorf("unmarshal oauth state cookie: %w", err)
	}
	if state.State == "" || normalizeMode(state.Mode) == "" {
		return OAuthState{}, ErrInvalidState
	}

	state.Mode = normalizeMode(state.Mode)
	state.Redirect = sanitizeRedirectPath(state.Redirect)

	return state, nil
}

func (s *Service) AppURL(path string) string {
	path = sanitizeRedirectPath(path)
	if s.appBaseURL == "" {
		return path
	}

	base, err := url.Parse(s.appBaseURL)
	if err != nil {
		return path
	}

	ref, err := url.Parse(path)
	if err != nil {
		return s.appBaseURL
	}

	return base.ResolveReference(ref).String()
}

func (s *Service) createSessionForUser(ctx context.Context, userID int64) (SessionPrincipal, string, error) {
	user, err := authTx.GetUserByID(ctx, s.db, userID)
	if err != nil {
		return SessionPrincipal{}, "", err
	}

	token, err := randomToken(32)
	if err != nil {
		return SessionPrincipal{}, "", fmt.Errorf("generate session token: %w", err)
	}
	expiresAt := time.Now().Add(s.sessionTTL)

	if err := authTx.CreateSession(ctx, s.db, user.ID, user.AccountID, hashToken(token), expiresAt); err != nil {
		return SessionPrincipal{}, "", err
	}

	principal, ok, err := s.ResolveSession(ctx, token)
	if err != nil {
		return SessionPrincipal{}, "", err
	}
	if !ok {
		return SessionPrincipal{}, "", errors.New("session was created but could not be reloaded")
	}

	return principal, token, nil
}

func (s *Service) fetchGoogleProfile(ctx context.Context, code string) (GoogleProfile, error) {
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://oauth2.googleapis.com/token", strings.NewReader(url.Values{
		"client_id":     []string{s.googleClientID},
		"client_secret": []string{s.googleClientSecret},
		"code":          []string{code},
		"grant_type":    []string{"authorization_code"},
		"redirect_uri":  []string{s.googleRedirectURL},
	}.Encode()))
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("create google token request: %w", err)
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenRes, err := s.httpClient.Do(tokenReq)
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("exchange google code: %w", err)
	}
	defer tokenRes.Body.Close()

	if tokenRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(tokenRes.Body, 4<<10))
		return GoogleProfile{}, fmt.Errorf("google token exchange failed: %s", strings.TrimSpace(string(body)))
	}

	var tokenPayload googleTokenResponse
	if err := json.NewDecoder(tokenRes.Body).Decode(&tokenPayload); err != nil {
		return GoogleProfile{}, fmt.Errorf("decode google token response: %w", err)
	}
	if tokenPayload.AccessToken == "" {
		return GoogleProfile{}, errors.New("google token response missing access token")
	}

	infoReq, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("create google userinfo request: %w", err)
	}
	infoReq.Header.Set("Authorization", "Bearer "+tokenPayload.AccessToken)

	infoRes, err := s.httpClient.Do(infoReq)
	if err != nil {
		return GoogleProfile{}, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer infoRes.Body.Close()

	if infoRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(infoRes.Body, 4<<10))
		return GoogleProfile{}, fmt.Errorf("google userinfo request failed: %s", strings.TrimSpace(string(body)))
	}

	var profile GoogleProfile
	if err := json.NewDecoder(infoRes.Body).Decode(&profile); err != nil {
		return GoogleProfile{}, fmt.Errorf("decode google userinfo response: %w", err)
	}
	if strings.TrimSpace(profile.Sub) == "" || strings.TrimSpace(profile.Email) == "" {
		return GoogleProfile{}, errors.New("google userinfo response missing required fields")
	}

	profile.Email = strings.TrimSpace(strings.ToLower(profile.Email))
	profile.Name = strings.TrimSpace(profile.Name)
	profile.Picture = strings.TrimSpace(profile.Picture)

	return profile, nil
}

func normalizeMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case ModeSignup:
		return ModeSignup
	case ModeLogin:
		return ModeLogin
	default:
		return ""
	}
}

func sanitizeRedirectPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "/app"
	}

	return path
}

func hashToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}

	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomToken(length int) (string, error) {
	buf := make([]byte, length)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}
