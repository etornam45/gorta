package magiclink

import (
	"github.com/etornam45/gorta/core"
	"net/http"
)
import "github.com/etornam45/gorta/interfaces"

type Config struct {
	BaseURL string
}

type Plugin struct {
	storage core.Storage
	mailer  interfaces.MagicMailer // nil if no email features needed
	config  Config
}

func (p *Plugin) Name() string { return "magiclink" }

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{Method: core.POST, Path: "/magic-link/request", Handler: (http.HandlerFunc(p.handleMagicLinkRequest))},
		{Method: core.GET, Path: "/magic-link/verify", Handler: (http.HandlerFunc(p.handleMagicLinkVerify))},
	}
}

type magicLinkRequestInput struct {
	Email string `json:"email"`
}

func (p *Plugin) handleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	
}

func (p *Plugin) handleMagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	
}