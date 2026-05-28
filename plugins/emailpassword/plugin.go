package emailpassword

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/etornam45/gorta/core"
	"github.com/etornam45/gorta/interfaces"
)

type Config struct {
	VerifyEmail bool
	BaseURL string
}

type Plugin struct {
	storage core.Storage
	creds   Storage
	mailer  interfaces.Mailer // nil if no email features needed
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
//	    BaseURL:     "https://myapp.com",
//	})
func New(storage core.Storage, creds Storage, mailer interfaces.Mailer, config Config) (*Plugin, error) {

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

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{Method: core.POST, Path: "/sign-up", Handler: http.HandlerFunc(p.handleSignUp)	},
		{Method: core.POST, Path: "/sign-in", Handler: http.HandlerFunc(p.handleSignIn)},
	}
}

type signUpInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (p *Plugin) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var input signUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("INVALID_BODY", "request body must be valid JSON"))
		return
	}

	result, err := p.signUp(r.Context(), input, core.SessionMetaFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}

	if result.token != "" {
		setSessionCookie(w, result.token, p.sessionDuration())
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":    result.user,
		"session": result.session,
	})
}

type signUpResult struct {
	user    *core.User
	session *core.Session
	token   string
}

func (p *Plugin) signUp(ctx context.Context, input signUpInput, meta core.SessionMeta) (*signUpResult, error) {
	input.Email = normalizeEmail(input.Email)
	input.Name = sanitizeName(input.Name)

	if err := validateEmail(input.Email); err != nil {
		return nil, err
	}
	if err := validatePassword(input.Password); err != nil {
		return nil, err
	}

	// Check for existing user before hashing — Argon2id is slow by design
	if _, err := p.storage.FindUserByEmail(ctx, input.Email); err == nil {
		return nil, core.ErrUserAlreadyExists
	} else if !errors.Is(err, core.ErrUserNotFound) {
		return nil, fmt.Errorf("emailpassword: checking email: %w", err)
	}

	hash, err := core.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user, err := p.storage.CreateUser(ctx, core.User{
		ID:        core.GenerateID(),
		Email:     input.Email,
		Name:      input.Name,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		return nil, fmt.Errorf("emailpassword: creating user: %w", err)
	}

	if err := p.creds.CreateCredential(ctx, user.ID, hash); err != nil {
		p.storage.DeleteUser(ctx, user.ID) //nolint:errcheck
		return nil, fmt.Errorf("emailpassword: storing credential: %w", err)
	}

	session, token, err := p.createSession(ctx, user.ID, meta)
	if err != nil {
		return &signUpResult{user: user}, nil
	}

	return &signUpResult{user: user, session: session, token: token}, nil
}

type signInInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *Plugin) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var input signInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody("INVALID_BODY", "request body must be valid JSON"))
		return
	}

	session, token, err := p.signIn(r.Context(), input, core.SessionMetaFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}

	setSessionCookie(w, token, p.sessionDuration())
	writeJSON(w, http.StatusOK, map[string]any{
		"user":    session.User,
		"session": session,
	})
}

func (p *Plugin) signIn(ctx context.Context, input signInInput, meta core.SessionMeta) (*core.Session, string, error) {
	input.Email = normalizeEmail(input.Email)

	user, err := p.storage.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, core.ErrUserNotFound) {
			core.VerifyPassword(input.Password, core.DUMMY_HASH) //nolint:errcheck
			return nil, "", core.ErrInvalidCredentials
		}
		return nil, "", fmt.Errorf("emailpassword: finding user: %w", err)
	}

	hash, err := p.creds.FindCredentialByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", fmt.Errorf("emailpassword: finding credential: %w", err)
	}

	ok, err := core.VerifyPassword(input.Password, hash)
	if err != nil || !ok {
		return nil, "", core.ErrInvalidCredentials
	}

	session, token, err := p.createSession(ctx, user.ID, meta)
	if err != nil {
		return nil, "", err
	}
	session.User = user

	return session, token, nil
}


func (p *Plugin) createSession(ctx context.Context, userID string, meta core.SessionMeta) (*core.Session, string, error) {
	token, err := core.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	session := core.Session{
		ID:        core.GenerateID(),
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(p.sessionDuration()),
		IPAddress: meta.IPAddress,
		UserAgent: meta.UserAgent,
		CreatedAt: time.Now(),
	}

	created, err := p.storage.CreateSession(ctx, session)
	if err != nil {
		return nil, "", fmt.Errorf("emailpassword: creating session: %w", err)
	}

	return created, token, nil
}

func (p *Plugin) sessionDuration() time.Duration {
	return 7 * 24 * time.Hour // plugins inherit or can override
}

func validateEmail(email string) error {
	if len(email) == 0 {
		return core.ErrInvalidEmail
	}
	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return core.ErrInvalidEmail
	}
	if !strings.Contains(email[at+1:], ".") {
		return core.ErrInvalidEmail
	}
	if strings.ContainsAny(email, " \t\n\r") {
		return core.ErrInvalidEmail
	}
	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 {
		return core.ErrWeakPassword
	}
	return nil
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func sanitizeName(name string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, strings.TrimSpace(name))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

func writeError(w http.ResponseWriter, err error) {
	status, code, message := errorToHTTP(err)
	writeJSON(w, status, errBody(code, message))
}

func errBody(code, message string) map[string]string {
	return map[string]string{"error": code, "message": message}
}

func errorToHTTP(err error) (int, string, string) {
	switch {
	case errors.Is(err, core.ErrInvalidCredentials):
		return http.StatusUnauthorized, "INVALID_CREDENTIALS", "invalid email or password"
	case errors.Is(err, core.ErrUserAlreadyExists):
		return http.StatusConflict, "USER_ALREADY_EXISTS", "a user with this email already exists"
	case errors.Is(err, core.ErrWeakPassword):
		return http.StatusBadRequest, "WEAK_PASSWORD", "password must be at least 8 characters"
	case errors.Is(err, core.ErrInvalidEmail):
		return http.StatusBadRequest, "INVALID_EMAIL", "invalid email address"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "an unexpected error occurred"
	}
}

func setSessionCookie(w http.ResponseWriter, token string, d time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gorta_session",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(d),
		MaxAge:   int(d.Seconds()),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}