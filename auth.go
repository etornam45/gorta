package gorta

import (
	"errors"
	"log"
	"time"
)

type Config struct {
	CookieName      string // default: "gorta_session"
	Secret          string
	SessionDuration time.Duration
	SecureCookies   bool
	CookieDomain    string
}

type Auth struct {
	adapter Adapter
	config  Config
	logger  *log.Logger
}

func New(adapter Adapter, config Config) (*Auth, error) {
	if adapter == nil {
		return nil, errors.New("adapter is required")
	}
	if len(config.Secret) < 32 {
		return nil, errors.New("secret must be at least 32 characters")
	}
	if config.SecureCookies && config.CookieDomain == "" {
		return nil, errors.New("cookie domain is required for secure cookies")
	}
	if config.SessionDuration == 0 {
		config.SessionDuration = 7 * 24 * time.Hour
	}
	return &Auth{adapter: adapter, config: config}, nil
}

func (a *Auth) logError(message string, err error) {
	if a.logger != nil {
		a.logger.Printf("[gorta] %s: %v", message, err)
	}
}
