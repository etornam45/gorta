package gorta

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

type contextKey string

const (
	sessionContextKey contextKey = "gorta_session"
	userContextKey    contextKey = "gorta_user"
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
				if errors.Is(err, ErrSessionExpired) || errors.Is(err, ErrSessionNotFound) {
					a.clearSessionCookie(w)
				}
				next.ServeHTTP(w, r)
				return
			}

			ctx := context.WithValue(r.Context(), sessionContextKey, session)
			ctx = context.WithValue(ctx, userContextKey, session.User)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (a *Auth) RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if GetSession(r.Context()) == nil {
				writeJSON(w, http.StatusUnauthorized, map[string]string{
					"error":   "UNAUTHENTICATED",
					"message": "you must be signed in to access this resource",
				})
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (a *Auth) extractToken(r *http.Request) string {
	if cookie, err := r.Cookie(a.config.CookieName); err == nil {
		return cookie.Value
	}

	bearer := r.Header.Get("Authorization")
	if strings.HasPrefix(bearer, "Bearer ") {
		return strings.TrimPrefix(bearer, "Bearer ")
	}

	return ""
}

// GetSession returns the session from the context, or nil if not authenticated.
func GetSession(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionContextKey).(*Session)
	return s
}

// GetUser returns the authenticated user from the context, or nil if not authenticated.
func GetUser(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey).(*User)
	return u
}

// IsAuthenticated reports whether the request context has a valid session.
func IsAuthenticated(ctx context.Context) bool {
	return GetSession(ctx) != nil
}
