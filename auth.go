package gorta

import (
	"errors"
	"fmt"
	"log"
	"time"

)

type Config struct {
	CookieName      string // default: "gorta_session"
	Secret          string
	SessionDuration time.Duration
	SecureCookies   bool
	CookieDomain    string
	VerifyEmail     bool
	VerifyEmailDomain  string
}

type Auth struct {
	adapter Adapter
	config  Config
	logger  *log.Logger
	mailer  Mailer
}

func New(adapter Adapter, config Config, mailer Mailer) (*Auth, error) {
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
	if config.VerifyEmail && mailer == nil {
		return nil, errors.New("mailer is required for verify email")
	}
	return &Auth{adapter: adapter, config: config, mailer: mailer}, nil
}

func (a *Auth) logError(message string, err error) {
	if a.logger != nil {
		fmt.Println("[gorta] %s: %v", message, err)
		// a.logger.Printf("[gorta] %s: %v", message, err)
	}
}
