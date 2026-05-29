package emailpassword

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/etornam45/gorta/core"
)

type Config struct {
	VerifyEmail       bool
	VerifyEmailDomain string
}

type Plugin struct {
	core    core.Core
	storage core.Storage
	creds   Storage
	mailer  Mailer // nil if no email features needed
	config  Config
}

// New creates the emailpassword plugin.
//
//	// Without email features
//	emailpassword.New(store, nil, emailpassword.Config{})
//
//	// With email verification
//	emailpassword.New(store, mailer, emailpassword.Config{
//	    VerifyEmail: true,
//	})
func New(storage core.Storage, creds Storage, mailer Mailer, config Config) (*Plugin, error) {

	if config.VerifyEmail && mailer == nil {
		return nil, fmt.Errorf("mailer is required for verify email")
	}

	return &Plugin{
		storage: storage,
		creds:   creds,
		mailer:  mailer,
		config:  config,
	}, nil
}

func (p *Plugin) Name() string { return "emailpassword" }

func (p *Plugin) Init(c core.Core) {
	p.core = c
}

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{Method: core.POST, Path: "/sign-up", Handler: http.HandlerFunc(p.handleSignUp)	},
		{Method: core.POST, Path: "/sign-in", Handler: http.HandlerFunc(p.handleSignIn)},
		{Method: core.GET, Path: "/verify-email", Handler: http.HandlerFunc(p.HandleVerifyEmail)},
	}
}

type SignUpInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type SignInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *Plugin) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var input SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_BODY",
			"message": "request body must be valid JSON",
		})
		return
	}

	result, err := p.SignUp(r.Context(), input, core.SessionMetaFromRequest(r))
	if err != nil {
		core.WriteError(w, err)
		return
	}

	if result.Token != "" {
		p.core.SetSessionCookie(w, result.Token)
	}

	core.WriteJSON(w, http.StatusCreated, map[string]any{
		"user":    result.User,
		"session": result.Session,
	})
}

func (p *Plugin) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var input SignInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_BODY",
			"message": "request body must be valid JSON",
		})
		return
	}

	result, err := p.SignIn(r.Context(), input, core.SessionMetaFromRequest(r))
	if err != nil {
		core.WriteError(w, err)
		return
	}

	p.core.SetSessionCookie(w, result.Token)
	core.WriteJSON(w, http.StatusOK, map[string]any{
		"user":    result.User,
		"session": result.Session,
	})
}

func (p *Plugin) handleSignOut(w http.ResponseWriter, r *http.Request) {
	token := core.ExtractToken(r, p.core.Config().CookieName)
	if token != "" {
		if err := p.core.RevokeSession(r.Context(), token); err != nil {
			fmt.Printf("[gorta] error signing out session: %v\n", err)
		}
	}

	p.core.ClearSessionCookie(w)
	core.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "signed out",
	})
}

func (p *Plugin) handleGetSession(w http.ResponseWriter, r *http.Request) {
	session := p.GetSession(r.Context())
	if session == nil {
		core.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHENTICATED",
			"message": "no active session",
		})
		return
	}

	core.WriteJSON(w, http.StatusOK, map[string]any{
		"user":    session.User,
		"session": session,
	})
}

func (p *Plugin) buildVerifyURL(token string) string {
	return p.core.Config().BaseURL + "/auth/verify-email?token=" + token
}

func sessionMetaFromRequest(r *http.Request) core.SessionMeta {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return core.SessionMeta{
		IPAddress: ip,
		UserAgent: r.Header.Get("User-Agent"),
	}
}

func (p *Plugin) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	core.WriteJSON(w, http.StatusOK, p.GetSession(r.Context()))
}

func (p *Plugin) HandleVerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Token is required",
		})
		return
	}

	user, err := p.VerifyEmail(r.Context(), token)
	if err != nil {
		core.WriteError(w, err)
		return
	}

	core.WriteJSON(w, http.StatusOK, map[string]any{
		"user": user,
	})
}
