package gorta

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/etornam45/gorta/core"
)

type Config struct {
	Secret string
	SessionDuration time.Duration
	CookieName string
	CookieDomain string
	SecureCookies bool
	Logger *slog.Logger
}

type Auth struct {
	storage core.Storage
	plugins []core.Plugin
	config  Config
	logger  *slog.Logger
}

// New creates a new Auth instance with the given core storage and plugins.
//
//	store := sqlstore.New(db)
//	auth, err := gorta.New(store,
//	    gorta.WithPlugin(emailpassword.New(store, mailer)),
//	    gorta.WithPlugin(magiclink.New(store, mailer)),
//	)
func New(storage core.Storage, config Config, opts ...Option) (*Auth, error) {
	if storage == nil {
		return nil, errors.New("gorta: storage is required")
	}
	if len(config.Secret) < 32 {
		return nil, fmt.Errorf("gorta: secret must be at least 32 characters, got %d", len(config.Secret))
	}
	if config.SessionDuration == 0 {
		config.SessionDuration = 7 * 24 * time.Hour
	}
	if config.CookieName == "" {
		config.CookieName = "gorta_session"
	}
	if config.Logger == nil {
		config.Logger = slog.Default()
	}

	a := &Auth{
		storage: storage,
		config:  config,
		logger:  config.Logger,
	}

	for _, opt := range opts {
		opt(a)
	}

	return a, nil
}

type Option func(*Auth)

func WithPlugin(p core.Plugin) Option {
	return func(a *Auth) {
		a.plugins = append(a.plugins, p)
	}
}

func (a *Auth) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /session", a.handleGetSession)
	mux.HandleFunc("POST /sign-out", a.handleSignOut)
	mux.HandleFunc("GET /me", a.RequireAuth()(http.HandlerFunc(a.handleGetMe)).ServeHTTP)

	for _, p := range a.plugins {
		for _, r := range p.Routes() {
			pattern := string(r.Method) + " " + r.Path
			mux.Handle(pattern, r.Handler)
		}
	}

	return mux
}

func (a *Auth) createSession(ctx context.Context, userID string, meta core.SessionMeta) (*core.Session, string, error) {
	token, err := core.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	session := core.Session{
		ID:        core.GenerateID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(a.config.SessionDuration),
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: time.Now(),
	}

	created, err := a.storage.CreateSession(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("gorta: creating session: %w", err)
	}

	return created, token, nil
}

func (a *Auth) ValidateSession(ctx context.Context, token string) (*core.Session, error) {
	session, err := a.storage.FindSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		go a.storage.DeleteSession(context.Background(), session.ID) //nolint:errcheck
		return nil, core.ErrSessionExpired
	}

	user, err := a.storage.FindUserByID(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("gorta: finding session user: %w", err)
	}

	session.User = user
	return session, nil
}

func (a *Auth) RevokeSession(ctx context.Context, token string) error {
	session, err := a.storage.FindSessionByToken(ctx, token)
	if err != nil {
		if errors.Is(err, core.ErrSessionNotFound) {
			return nil
		}
		return err
	}
	return a.storage.DeleteSession(ctx, session.ID)
}

type contextKey string

const (
	sessionKey contextKey = "gorta_session"
	userKey    contextKey = "gorta_user"
)

func (a *Auth) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := a.extractToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			session, err := a.ValidateSession(r.Context(), token)
			if err != nil {
				if errors.Is(err, core.ErrSessionExpired) || errors.Is(err, core.ErrSessionNotFound) {
					a.clearSessionCookie(w)
				}
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), sessionKey, session)
			ctx = context.WithValue(ctx, userKey, session.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Auth) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if GetSession(r.Context()) == nil {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "UNAUTHENTICATED",
					"message": "you must be signed in",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if GetSession(r.Context()) == nil {
				WriteJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "UNAUTHENTICATED",
					"message": "you must be signed in",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetSession(ctx context.Context) *core.Session {
	s, _ := ctx.Value(sessionKey).(*core.Session)
	return s
}

func GetUser(ctx context.Context) *core.User {
	u, _ := ctx.Value(userKey).(*core.User)
	return u
}

func (a *Auth) extractToken(r *http.Request) string {
	if cookie, err := r.Cookie(a.config.CookieName); err == nil {
		return cookie.Value
	}
	if bearer := r.Header.Get("Authorization"); strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}
	return ""
}

func (a *Auth) SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.config.CookieName,
		Value:    token,
		Path:     "/",
		Domain:   a.config.CookieDomain,
		Expires:  time.Now().Add(a.config.SessionDuration),
		MaxAge:   int(a.config.SessionDuration.Seconds()),
		Secure:   a.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *Auth) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     a.config.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   a.config.CookieDomain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   a.config.SecureCookies,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (a *Auth) handleGetSession(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r.Context())
	if session == nil {
		WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHENTICATED",
			"message": "no active session",
		})
		return
	}
	WriteJSON(w, http.StatusOK, map[string]any{
		"user":    session.User,
		"session": session,
	})
}

func (a *Auth) handleSignOut(w http.ResponseWriter, r *http.Request) {
	token := a.extractToken(r)
	if token != "" {
		if err := a.RevokeSession(r.Context(), token); err != nil {
			a.logger.Error("revoking session", "error", err)
		}
	}
	a.clearSessionCookie(w)
	WriteJSON(w, http.StatusOK, map[string]string{"message": "signed out"})
}

func (a *Auth) handleGetMe(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, GetUser(r.Context()))
}

func SessionMetaFromRequest(r *http.Request) core.SessionMeta {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return core.SessionMeta{IPAddress: ip, UserAgent: r.Header.Get("User-Agent")}
}

func (a *Auth) logError(msg string, err error) {
	a.logger.Error(msg, "error", err)
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func WriteError(w http.ResponseWriter, err error) {
	status, code, message := errorToHTTP(err)
	WriteJSON(w, status, map[string]string{"error": code, "message": message})
}

func errorToHTTP(err error) (status int, code, message string) {
	switch {
	case errors.Is(err, core.ErrInvalidCredentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password"
	case errors.Is(err, core.ErrUserNotFound):
		return http.StatusNotFound, "USER_NOT_FOUND", "user not found"
	case errors.Is(err, core.ErrUserAlreadyExists):
		return http.StatusConflict, "USER_ALREADY_EXISTS", "a user with this email already exists"
	case errors.Is(err, core.ErrSessionNotFound), errors.Is(err, core.ErrInvalidToken):
		return http.StatusUnauthorized, "INVALID_SESSION", "session is invalid"
	case errors.Is(err, core.ErrSessionExpired):
		return http.StatusUnauthorized, "SESSION_EXPIRED", "session has expired"
	case errors.Is(err, core.ErrEmailNotVerified):
		return http.StatusForbidden, "EMAIL_NOT_VERIFIED", "please verify your email address"
	case errors.Is(err, core.ErrWeakPassword):
		return http.StatusBadRequest, "WEAK_PASSWORD", "password must be at least 8 characters"
	case errors.Is(err, core.ErrInvalidEmail):
		return http.StatusBadRequest, "INVALID_EMAIL", "invalid email address"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred"
	}
}