package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/etornam45/gorta/core"
	"github.com/etornam45/gorta/plugins/oauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yahoo"
)

type YahooConfig struct {
	ClientID     string
	ClientSecret string
}

type YahooProvider struct {
	config oauth2.Config
}

func NewYahooProvider(cfg YahooConfig) *YahooProvider {
	return &YahooProvider{
		config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     yahoo.Endpoint,
		},
	}
}

func (y *YahooProvider) Name() string { return string(oauth.Yahoo) }

func (y *YahooProvider) Config() *oauth2.Config { return &y.config }

func (y *YahooProvider) AuthorizeOptions(state *oauth.OAuthState) ([]oauth2.AuthCodeOption, error) {
	verifier := oauth2.GenerateVerifier()
	state.CodeVerifier = verifier

	nonce, err := core.GenerateToken()
	if err != nil {
		return nil, fmt.Errorf("yahoo: generating nonce: %w", err)
	}

	return []oauth2.AuthCodeOption{
		oauth2.S256ChallengeOption(verifier),
		oauth2.SetAuthURLParam("nonce", nonce),
	}, nil
}

func (y *YahooProvider) ExchangeOptions(state *oauth.OAuthState) []oauth2.AuthCodeOption {
	if state.CodeVerifier == "" {
		return nil
	}
	return []oauth2.AuthCodeOption{oauth2.VerifierOption(state.CodeVerifier)}
}

func (y *YahooProvider) GetUser(ctx context.Context, token *oauth2.Token) (*oauth.UserInfo, error) {
	client := y.config.Client(ctx, token)
	resp, err := client.Get("https://api.login.yahoo.com/openid/v1/userinfo")
	if err != nil {
		return nil, fmt.Errorf("yahoo: fetching user: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("yahoo: userinfo returned %d: %s", resp.StatusCode, body)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("yahoo: decoding userinfo: %w", err)
	}

	id, _ := raw["sub"].(string)
	email, _ := raw["email"].(string)
	name, _ := raw["name"].(string)
	picture, _ := raw["picture"].(string)
	// verified, _ := raw["verified"].(bool)
	verified := true

	return &oauth.UserInfo{
		ID:            id,
		Email:         email,
		EmailVerified: verified,
		Name:          name,
		Picture:       picture,
		Raw:           raw,
	}, nil
}
