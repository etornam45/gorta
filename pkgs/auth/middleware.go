package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/etornam45/gorta/internals"
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
				if errors.Is(err, internals.ErrSessionExpired) || errors.Is(err, internals.ErrSessionNotFound) {
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
				WriteJSON(w, http.StatusUnauthorized, map[string]string{
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
func GetSession(ctx context.Context) *internals.Session {
	s, _ := ctx.Value(sessionContextKey).(*internals.Session)
	return s
}

// GetUser returns the authenticated user from the context, or nil if not authenticated.
func GetUser(ctx context.Context) *internals.User {
	u, ok := ctx.Value(userContextKey).(*internals.User)
	if !ok {
		fmt.Println("User not found in context, Did you forget to add the middleware? Got: ", ctx.Value(userContextKey))
		return nil
	}
	return u
}

// IsAuthenticated reports whether the request context has a valid session.
func IsAuthenticated(ctx context.Context) bool {
	return GetSession(ctx) != nil
}
