package emailpassword

import (
	"context"
)

func (p *Plugin) SignOut(ctx context.Context, token string) error {
	return p.core.RevokeSession(ctx, token)
}

func (p *Plugin) SignOutEverywhere(ctx context.Context, userID string) error {
	return p.storage.RevokeAllSessions(ctx, userID)
}
