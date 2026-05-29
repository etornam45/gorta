package core

import (
	"net/http"
	"time"
)

func SetSessionCookie(w http.ResponseWriter, token string, name string, domain string, duration time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    token,
		Path:     "/",
		Domain:   domain,
		Expires:  time.Now().Add(duration),
		MaxAge:   int(duration.Seconds()),
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func ClearSessionCookie(w http.ResponseWriter, name string, domain string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		Secure:   secure,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}
