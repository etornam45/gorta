package magiclink

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/etornam45/gorta"
	"github.com/etornam45/gorta/core"
)

type Config struct {
	BaseURL    string
	Expiration time.Duration
}

type Plugin struct {
	storage Storage
	mailer  Mailer // nil if no email features needed
	config  Config
}

func New(storage Storage, mailer Mailer, config Config) (*Plugin, error) {
	return &Plugin{
		storage: storage,
		mailer:  mailer,
		config:  config,
	}, nil
}

func (p *Plugin) Name() string { return "magiclink" }

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{Method: core.POST, Path: "/magic-link/request", Handler: http.HandlerFunc(p.handleMagicLinkRequest)},
		{Method: core.GET, Path: "/magic-link/verify", Handler: http.HandlerFunc(p.handleMagicLinkVerify)},
	}
}

type magicLinkRequestInput struct {
	Email string `json:"email"`
}

func (p *Plugin) handleMagicLinkRequest(w http.ResponseWriter, r *http.Request) {
	var input magicLinkRequestInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	token, err := core.GenerateToken()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	verification, err := p.storage.CreateVerification(ctx, Verification{
		ID:         core.GenerateID(),
		Identifier: input.Email,
		Token:      token,
		ExpiresAt:  now.Add(p.config.Expiration),
		CreatedAt:  now,
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	magicLink := fmt.Sprintf("%s/auth/magic-link/verify?token=%s", p.config.BaseURL, verification.Token)
	if err := p.mailer.SendMagicLink(ctx, input.Email, magicLink); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	gorta.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Magic link sent to " + input.Email,
	})
}


// FIXME: Do I have to hash the verification token? B4 storing it in the database?

func (p *Plugin) handleMagicLinkVerify(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	verification, err := p.storage.FindVerificationByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if verification.ExpiresAt.Before(time.Now()) {
		http.Error(w, "verification expired", http.StatusGone)
		return
	}

	store, ok := p.storage.(core.Storage)
	if !ok {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	user, err := store.FindUserByEmail(r.Context(), verification.Identifier)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	sessionToken, err := core.GenerateToken()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	meta := core.SessionMetaFromRequest(r)
	session, err := store.CreateSession(r.Context(), core.Session{
		ID:        core.GenerateID(),
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(p.config.Expiration),
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: time.Now(),
	})
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	gorta.WriteJSON(w, http.StatusOK, map[string]any{
		"user":    user,
		"session": session,
	})
}
