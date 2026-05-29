package core

import "context"

type contextKey string

const (
	sessionContextKey contextKey = "gorta_session"
	userContextKey    contextKey = "gorta_user"
)

func SessionFromContext(ctx context.Context) *Session {
	s, _ := ctx.Value(sessionContextKey).(*Session)
	return s
}

func UserFromContext(ctx context.Context) *User {
	u, _ := ctx.Value(userContextKey).(*User)
	return u
}

func ContextWithSession(ctx context.Context, session *Session) context.Context {
	ctx = context.WithValue(ctx, sessionContextKey, session)
	if session != nil && session.User != nil {
		ctx = context.WithValue(ctx, userContextKey, session.User)
	}
	return ctx
}
