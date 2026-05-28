package auth

import "context"

func (a *Auth) SignOut(ctx context.Context, token string) error {
	return a.RevokeSession(ctx, token)
}

func (a *Auth) SignOutEverywhere(ctx context.Context, userID string) error {
	return a.RevokeAllSessions(ctx, userID)
}
