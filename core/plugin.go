package core

import (
	"context"
	"net/http"
	"time"
)

type Plugin interface {
	Name() string
	Init(c Core) // FIXME: Core is not a pointer, so we need to pass a pointer to the core
	Routes() []Route
}

type Core interface {
	CreateSession(ctx context.Context, userID string, meta SessionMeta) (*Session, string, error)
	RevokeSession(ctx context.Context, token string) error
	ValidateSession(ctx context.Context, token string) (*Session, error)

	SetSessionCookie(w http.ResponseWriter, token string)
	ClearSessionCookie(w http.ResponseWriter)

	RequireAuth() func(http.Handler) http.Handler
	Config() CoreConfig
}

type CoreConfig struct {
	SessionDuration time.Duration
	BaseURL         string
	SecureCookies   bool
	CookieName      string
}

type RouteMethod string

const (
	GET     RouteMethod = "GET"
	POST    RouteMethod = "POST"
	PUT     RouteMethod = "PUT"
	DELETE  RouteMethod = "DELETE"
	PATCH   RouteMethod = "PATCH"
	OPTIONS RouteMethod = "OPTIONS"
	HEAD    RouteMethod = "HEAD"
)

type Route struct {
	Method  RouteMethod
	Path    string
	Handler http.Handler
}
