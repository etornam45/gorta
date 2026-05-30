package gorta

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/etornam45/gorta/core"
)

type Config struct {
	Secret          string
	SessionDuration time.Duration
	BaseURL         string
	CookieName      string
	CookieDomain    string
	SecureCookies   bool
	Logger          *slog.Logger
}

type Auth struct {
	storage core.Storage
	plugins []core.Plugin
	config  Config
	routes  map[string]http.Handler
	logger  *slog.Logger
}

// New creates a new Auth instance with the given core storage and plugins.
//
//	store := sqlstore.New(db)
//	auth, err := gorta.New(store, config,
//	    gorta.WithPlugin(emailpassword.New(store, mailer)),
//	    gorta.WithPlugin(magiclink.New(store, mailer)),
//	)
//	mux := http.NewServeMux()
//	mux.Handle("/auth/", a.Handler())
//	http.ListenAndServe(":8080", a.Middleware()(mux))
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
		routes:  make(map[string]http.Handler),
	}

	for _, opt := range opts {
		opt(a)
	}

	for _, p := range a.plugins {
		p.Init(a)
	}

	return a, nil
}

type Option func(*Auth)

func WithPlugin(p core.Plugin) Option {
	return func(a *Auth) {
		a.plugins = append(a.plugins, p)

		for _, r := range p.Routes() {
			a.routes[string(r.Method)+" "+r.Path] = r.Handler
		}
	}
}

func (a *Auth) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /session", a.handleGetSession)
	mux.HandleFunc("POST /sign-out", a.handleSignOut)
	mux.HandleFunc("GET /me", a.RequireAuth()(http.HandlerFunc(a.handleGetMe)).ServeHTTP)

	// Print the routes
	a.logger.Info("\n\n[core]\n")
	a.logger.Info("GET /session")
	a.logger.Info("POST /sign-out")
	a.logger.Info("GET /me")


	// FIXME: Consider using a more efficient way to register routes
	a.logger.Info("\n\n[plugins]\n")
	for path, handler := range a.routes {
		mux.Handle(path, handler)
		a.logger.Info(path)
	}

	return mux
}

func (a *Auth) CreateSession(ctx context.Context, userID string, meta core.SessionMeta) (*core.Session, string, error) {
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
					a.ClearSessionCookie(w)
				}
				next.ServeHTTP(w, r)
				return
			}

			ctx := core.ContextWithSession(r.Context(), session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Auth) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if GetSession(r.Context()) == nil {
				core.WriteJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "UNAUTHENTICATED",
					"message": "you must be signed in",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
// TODO: REMOVE THIS
func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if GetSession(r.Context()) == nil {
				core.WriteJSON(w, http.StatusUnauthorized, map[string]string{
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
	return core.SessionFromContext(ctx)
}

func GetUser(ctx context.Context) *core.User {
	return core.UserFromContext(ctx)
}

func (a *Auth) Config() core.CoreConfig {
	return core.CoreConfig{
		SessionDuration: a.config.SessionDuration,
		BaseURL:         a.config.BaseURL,
		SecureCookies:   a.config.SecureCookies,
		CookieName:      a.config.CookieName,
	}
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

func (a *Auth) ClearSessionCookie(w http.ResponseWriter) {
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
		core.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHENTICATED",
			"message": "no active session",
		})
		return
	}
	core.WriteJSON(w, http.StatusOK, map[string]any{
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
	a.ClearSessionCookie(w)
	core.WriteJSON(w, http.StatusOK, map[string]string{"message": "signed out"})
}

func (a *Auth) handleGetMe(w http.ResponseWriter, r *http.Request) {
	core.WriteJSON(w, http.StatusOK, GetSession(r.Context()))
}
