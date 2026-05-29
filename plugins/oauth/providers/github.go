package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/etornam45/gorta/plugins/oauth"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GitHubConfig struct {
	ClientID     string
	ClientSecret string
}

type GitHubProvider struct {
	config oauth2.Config
}

func NewGitHubProvider(cfg GitHubConfig) *GitHubProvider {
	return &GitHubProvider{
		config: oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       []string{"read:user", "user:email"},
			Endpoint:     github.Endpoint,
		},
	}
}

func (g *GitHubProvider) Name() string { return string(oauth.GitHub) }

func (g *GitHubProvider) Config() *oauth2.Config { return &g.config }

func (g *GitHubProvider) GetUser(ctx context.Context, token *oauth2.Token) (*oauth.UserInfo, error) {
	client := g.config.Client(ctx, token)

	userRaw, err := g.fetchJSON(client, "https://api.github.com/user")
	if err != nil {
		return nil, fmt.Errorf("github: fetching user: %w", err)
	}

	id := stringifyGitHubID(userRaw["id"])
	if id == "" {
		return nil, fmt.Errorf("github: user missing id")
	}

	name, _ := userRaw["name"].(string)
	if name == "" {
		name, _ = userRaw["login"].(string)
	}
	picture, _ := userRaw["avatar_url"].(string)

	email, verified, err := g.primaryEmail(client, userRaw)
	if err != nil {
		return nil, err
	}
	if email == "" {
		return nil, fmt.Errorf("github: no primary email found")
	}

	return &oauth.UserInfo{
		ID:            id,
		Email:         email,
		EmailVerified: verified,
		Name:          name,
		Picture:       picture,
		Raw:           userRaw,
	}, nil
}

func (g *GitHubProvider) primaryEmail(client *http.Client, userRaw map[string]any) (string, bool, error) {
	if email, ok := userRaw["email"].(string); ok && email != "" {
		return email, true, nil
	}

	emailsRaw, err := g.fetchJSONSlice(client, "https://api.github.com/user/emails")
	if err != nil {
		return "", false, fmt.Errorf("github: fetching emails: %w", err)
	}

	for _, item := range emailsRaw {
		email, _ := item["email"].(string)
		primary, _ := item["primary"].(bool)
		verified, _ := item["verified"].(bool)
		if email != "" && primary {
			return email, verified, nil
		}
	}
	for _, item := range emailsRaw {
		email, _ := item["email"].(string)
		verified, _ := item["verified"].(bool)
		if email != "" {
			return email, verified, nil
		}
	}
	return "", false, nil
}

func (g *GitHubProvider) fetchJSON(client *http.Client, url string) (map[string]any, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var raw map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func (g *GitHubProvider) fetchJSONSlice(client *http.Client, url string) ([]map[string]any, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, body)
	}

	var raw []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func stringifyGitHubID(v any) string {
	switch id := v.(type) {
	case float64:
		return strconv.FormatInt(int64(id), 10)
	case string:
		return id
	default:
		return fmt.Sprint(v)
	}
}
