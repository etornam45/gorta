package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Secret          string
	SessionDuration time.Duration
	BaseURL         string
	Port            string
	SecureCookies   bool
	CookieName      string
	CookieDomain    string
	DatabasePath    string

	ResendAPIKey string
	FromEmail    string
	FromName     string

	GoogleClientID     string
	GoogleClientSecret string
	GitHubClientID     string
	GitHubClientSecret string
}

func LoadConfig() (AppConfig, error) {
	_ = godotenv.Load()
	return loadConfigFromEnv()
}

func loadConfigFromEnv() (AppConfig, error) {
	cfg := AppConfig{
		Port:         envOr("PORT", "8080"),
		BaseURL:      envOr("BASE_URL", "http://localhost:8080"),
		CookieName:   envOr("COOKIE_NAME", "gorta_session"),
		CookieDomain: envOr("COOKIE_DOMAIN", "localhost"),
		DatabasePath: envOr("DATABASE_PATH", "./gorta.db"),
		FromEmail:    envOr("FROM_EMAIL", "onboarding@resend.dev"),
		FromName:     envOr("FROM_NAME", "Gorta"),
	}

	cfg.Secret = os.Getenv("SECRET")
	if len(cfg.Secret) < 32 {
		return cfg, fmt.Errorf("SECRET must be at least 32 characters")
	}

	cfg.ResendAPIKey = os.Getenv("RESEND_API_KEY")

	duration, err := parseDuration(envOr("SESSION_DURATION", "24h"))
	if err != nil {
		return cfg, fmt.Errorf("SESSION_DURATION: %w", err)
	}
	cfg.SessionDuration = duration

	secure, err := parseBool(envOr("COOKIE_SECURE", "false"))
	if err != nil {
		return cfg, fmt.Errorf("COOKIE_SECURE: %w", err)
	}
	cfg.SecureCookies = secure

	cfg.GoogleClientID = os.Getenv("GOOGLE_CLIENT_ID")
	cfg.GoogleClientSecret = os.Getenv("GOOGLE_CLIENT_SECRET")
	cfg.GitHubClientID = os.Getenv("GITHUB_CLIENT_ID")
	cfg.GitHubClientSecret = os.Getenv("GITHUB_CLIENT_SECRET")

	return cfg, nil
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseBool(s string) (bool, error) {
	return strconv.ParseBool(strings.TrimSpace(s))
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	return time.ParseDuration(s)
}

func redact(s string) string {
	if s == "" {
		return "(not set)"
	}
	if len(s) <= 8 {
		return "****"
	}
	return s[:4] + "..." + s[len(s)-4:]
}
