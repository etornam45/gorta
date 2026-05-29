package emailpassword

import (
	"strings"
	"unicode"

	"github.com/etornam45/gorta/core"
)

func ValidateEmail(email string) error {
	if len(email) == 0 {
		return core.ErrInvalidEmail
	}

	at := strings.LastIndex(email, "@")
	if at <= 0 {
		return core.ErrInvalidEmail
	}

	local := email[:at]
	domain := email[at+1:]

	if len(local) == 0 || len(domain) == 0 {
		return core.ErrInvalidEmail
	}

	if !strings.Contains(domain, ".") {
		return core.ErrInvalidEmail
	}

	if strings.ContainsAny(email, " \t\n\r") {
		return core.ErrInvalidEmail
	}

	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return core.ErrWeakPassword
	}
	return nil
}

func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func SanitizeName(name string) string {
	name = strings.TrimSpace(name)
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, name)
}
