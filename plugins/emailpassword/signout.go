package emailpassword

import (
	"context"
	"fmt"

	"github.com/etornam45/gorta/core"
)

func (p *Plugin) SignOut(ctx context.Context, token string) error {
	storage, ok := p.storage.(core.Storage)
	if !ok {
		return fmt.Errorf("gorta: error revoking session")
	}
	return storage.RevokeSession(ctx, token)
}

func (p *Plugin) SignOutEverywhere(ctx context.Context, userID string) error {
	storage, ok := p.storage.(core.Storage)
	if !ok {
		return fmt.Errorf("gorta: error revoking all sessions")
	}
	return storage.RevokeAllSessions(ctx, userID)
}
