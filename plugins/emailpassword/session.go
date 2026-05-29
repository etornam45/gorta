package emailpassword

import (
	"context"
	"net/http"

	"github.com/etornam45/gorta/core"
)

func (p *Plugin) GetSession(ctx context.Context) *core.Session {
	return core.SessionFromContext(ctx)
}

func (p *Plugin) clearSessionCookie(w http.ResponseWriter) {
	core.ClearSessionCookie(w, core.DefaultSessionCookie, "", false)
}
