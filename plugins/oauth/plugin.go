package oauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/etornam45/gorta/core"
	"golang.org/x/oauth2"
)

type Config struct {
	Providers       []Provider
	SuccessRedirect string
}

type Plugin struct {
	core      core.Core
	users     core.Storage
	accounts  Storage
	state     AuthState
	config    Config
	providers map[string]Provider
}

// FIXME: Storage is not a pointer, so we need to pass a pointer to the storage
func New(users core.Storage, accounts Storage, state AuthState, config Config) (*Plugin, error) {
	if users == nil {
		return nil, fmt.Errorf("oauth: user storage is required")
	}
	if accounts == nil {
		return nil, fmt.Errorf("oauth: account storage is required")
	}
	if state == nil {
		state = NewMemoryStateStore()
	}
	if len(config.Providers) == 0 {
		return nil, fmt.Errorf("oauth: at least one provider is required")
	}

	providers := make(map[string]Provider, len(config.Providers))
	for _, p := range config.Providers {
		name := p.Name()
		if name == "" {
			return nil, fmt.Errorf("oauth: provider name is required")
		}
		if _, exists := providers[name]; exists {
			return nil, fmt.Errorf("oauth: duplicate provider %q", name)
		}
		providers[name] = p
	}

	return &Plugin{
		users:     users,
		accounts:  accounts,
		state:     state,
		config:    config,
		providers: providers,
	}, nil
}

func (p *Plugin) Name() string { return "oauth" }

func (p *Plugin) Init(c core.Core) {
	p.core = c
}

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{Method: core.GET, Path: "/oauth/{provider}", Handler: http.HandlerFunc(p.handleAuthorize)},
		{Method: core.GET, Path: "/oauth/{provider}/callback", Handler: http.HandlerFunc(p.handleCallback)},
	}
}

func (p *Plugin) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	provider, err := p.providerFromRequest(r)
	if err != nil {
		writeProviderError(w, err)
		return
	}

	stateToken, err := core.GenerateToken()
	if err != nil {
		core.WriteError(w, err)
		return
	}

	if err := p.state.Save(r.Context(), OAuthState{
		Token:     stateToken,
		Provider:  ProviderName(provider.Name()),
		ExpiresAt: time.Now().Add(defaultStateTTL),
	}); err != nil {
		core.WriteError(w, err)
		return
	}

	cfg := p.providerConfig(provider)
	opts := []oauth2.AuthCodeOption{}
	if provider.Name() == string(Google) {
		// Google requires the access_type=offline parameter to get a refresh token
		//TODO:  Move this logic to the provider level so each provider can handle it differently
		opts = append(opts, oauth2.SetAuthURLParam("access_type", "offline"))
	}
	url := cfg.AuthCodeURL(stateToken, opts...)
	http.Redirect(w, r, url, http.StatusFound)
}

func (p *Plugin) handleCallback(w http.ResponseWriter, r *http.Request) {
	provider, err := p.providerFromRequest(r)
	if err != nil {
		writeProviderError(w, err)
		return
	}

	if errMsg := r.URL.Query().Get("error"); errMsg != "" {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "OAUTH_DENIED",
			"message": errMsg,
		})
		return
	}

	stateToken := r.URL.Query().Get("state")
	if stateToken == "" {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_STATE",
			"message": "missing state parameter",
		})
		return
	}

	savedState, err := p.state.Find(r.Context(), stateToken)
	if err != nil {
		if errors.Is(err, ErrOAuthStateNotFound) {
			core.WriteJSON(w, http.StatusBadRequest, map[string]string{
				"error":   "INVALID_STATE",
				"message": "state token is invalid or expired",
			})
			return
		}
		core.WriteError(w, err)
		return
	}
	defer p.state.Delete(r.Context(), stateToken) //nolint:errcheck

	if savedState.Provider != ProviderName(provider.Name()) {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_STATE",
			"message": "state provider mismatch",
		})
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_CODE",
			"message": "missing authorization code",
		})
		return
	}

	cfg := p.providerConfig(provider)
	token, err := cfg.Exchange(r.Context(), code)
	if err != nil {
		core.WriteJSON(w, http.StatusBadGateway, map[string]string{
			"error":   "TOKEN_EXCHANGE_FAILED",
			"message": "failed to exchange authorization code",
		})
		return
	}

	userInfo, err := provider.GetUser(r.Context(), token)
	if err != nil {
		core.WriteError(w, err)
		return
	}

	user, err := p.findOrCreateUser(r.Context(), provider, userInfo, token)
	if err != nil {
		core.WriteError(w, err)
		return
	}

	sessionToken, err := p.createSession(r.Context(), user.ID, r)
	if err != nil {
		core.WriteError(w, err)
		return
	}

	p.core.SetSessionCookie(w, sessionToken)
	http.Redirect(w, r, p.successRedirect(), http.StatusFound)
}

func (p *Plugin) findOrCreateUser(ctx context.Context, provider Provider, info *UserInfo, token *oauth2.Token) (*core.User, error) {
	account, err := p.accounts.FindAccountByProvider(ctx, provider.Name(), info.ID)
	if err == nil {
		user, err := p.users.FindUserByID(ctx, account.UserID)
		if err != nil {
			return nil, fmt.Errorf("oauth: loading linked user: %w", err)
		}
		return user, nil
	}
	if !errors.Is(err, core.ErrAccountNotFound) {
		return nil, fmt.Errorf("oauth: finding account: %w", err)
	}

	var user *core.User
	existing, err := p.users.FindUserByEmail(ctx, info.Email)
	if err == nil {
		user = existing
	} else if errors.Is(err, core.ErrUserNotFound) {
		now := time.Now()
		created, err := p.users.CreateUser(ctx, core.User{
			ID:            core.GenerateID(),
			Email:         info.Email,
			EmailVerified: info.EmailVerified,
			Name:          info.Name,
			Image:         info.Picture,
			CreatedAt:     now,
			UpdatedAt:     now,
		})
		if err != nil {
			return nil, fmt.Errorf("oauth: creating user: %w", err)
		}
		user = created
	} else {
		return nil, fmt.Errorf("oauth: finding user by email: %w", err)
	}

	var expiresAt *time.Time
	if !token.Expiry.IsZero() {
		expiresAt = &token.Expiry
	}

	_, err = p.accounts.CreateAccount(ctx, Account{
		ID:                core.GenerateID(),
		UserID:            user.ID,
		Provider:          ProviderName(provider.Name()),
		ProviderAccountID: info.ID,
		AccessToken:       token.AccessToken,
		RefreshToken:      token.RefreshToken,
		ExpiresAt:         expiresAt,
		CreatedAt:         time.Now(),
	})
	if err != nil {
		return nil, fmt.Errorf("oauth: linking account: %w", err)
	}

	return user, nil
}

func (p *Plugin) createSession(ctx context.Context, userID string, r *http.Request) (string, error) {
	_, token, err := p.core.CreateSession(ctx, userID, core.SessionMetaFromRequest(r))
	return token, err
}

var (
	errInvalidProvider = errors.New("oauth: provider is required")
	errUnknownProvider = errors.New("oauth: unknown provider")
)

func (p *Plugin) providerFromRequest(r *http.Request) (Provider, error) {
	name := r.PathValue("provider")
	if name == "" {
		return nil, errInvalidProvider
	}
	provider, ok := p.providers[name]
	if !ok {
		return nil, errUnknownProvider
	}
	return provider, nil
}

func writeProviderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, errInvalidProvider):
		core.WriteJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_PROVIDER",
			"message": "provider is required",
		})
	case errors.Is(err, errUnknownProvider):
		core.WriteJSON(w, http.StatusNotFound, map[string]string{
			"error":   "UNKNOWN_PROVIDER",
			"message": "unknown oauth provider",
		})
	default:
		core.WriteError(w, err)
	}
}

func (p *Plugin) providerConfig(provider Provider) *oauth2.Config {
	cfg := *provider.Config()
	cfg.RedirectURL = p.callbackURL(provider.Name())
	return &cfg
}

func (p *Plugin) callbackURL(provider string) string {
	return p.core.Config().BaseURL + "/auth/oauth/" + provider + "/callback"
}

func (p *Plugin) successRedirect() string {
	if p.config.SuccessRedirect != "" {
		return p.config.SuccessRedirect
	}
	return p.core.Config().BaseURL
}
