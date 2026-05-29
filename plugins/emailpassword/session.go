package emailpassword

import (
	"context"

	"github.com/etornam45/gorta/core"
)

func (p *Plugin) GetSession(ctx context.Context) *core.Session {
	return core.SessionFromContext(ctx)
}
