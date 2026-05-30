package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/etornam45/gorta/plugins/oauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
}

type GoogleProvider struct {
	config oauth2.Config
}

func NewGoogleProvider(cfg GoogleConfig) *GoogleProvider {
	return &GoogleProvider{
		config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{"openid", "email", "profile"},
			Endpoint:     google.Endpoint,
		},
	}
}

func (g *GoogleProvider) Name() string { return string(oauth.Google) }

func (g *GoogleProvider) Config() *oauth2.Config { return &g.config }

func (g *GoogleProvider) AuthorizeOptions(_ *oauth.OAuthState) ([]oauth2.AuthCodeOption, error) {
	return []oauth2.AuthCodeOption{
		oauth2.SetAuthURLParam("access_type", "offline"),
	}, nil
}

func (g *GoogleProvider) ExchangeOptions(_ *oauth.OAuthState) []oauth2.AuthCodeOption {
	return nil
}

func (g *GoogleProvider) GetUser(ctx context.Context, token *oauth2.Token) (*oauth.UserInfo, error) {
	client := g.config.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, fmt.Errorf("google: fetching userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("google: userinfo returned %d: %s", resp.StatusCode, body)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("google: decoding userinfo: %w", err)
	}

	id, _ := raw["id"].(string)
	email, _ := raw["email"].(string)
	name, _ := raw["name"].(string)
	picture, _ := raw["picture"].(string)
	verified, _ := raw["verified_email"].(bool)

	if id == "" {
		return nil, fmt.Errorf("google: userinfo missing id")
	}
	if email == "" {
		return nil, fmt.Errorf("google: userinfo missing email")
	}

	return &oauth.UserInfo{
		ID:            id,
		Email:         email,
		EmailVerified: verified,
		Name:          name,
		Picture:       picture,
		Raw:           raw,
	}, nil
}
