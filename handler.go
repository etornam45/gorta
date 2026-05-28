package gorta

import (
	"encoding/json"
	"net/http"
)

func (a *Auth) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("POST /sign-up", http.HandlerFunc(a.handleSignUp))
	mux.Handle("POST /sign-in", http.HandlerFunc(a.handleSignIn))
	mux.Handle("POST /sign-out", http.HandlerFunc(a.handleSignOut))
	mux.Handle("GET /session", http.HandlerFunc(a.handleGetSession))
	mux.Handle("GET /me", a.RequireAuth()(http.HandlerFunc(a.HandleGetMe)))
	return mux
}

func (a *Auth) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var input SignUpInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_BODY",
			"message": "request body must be valid JSON",
		})
		return
	}

	result, err := a.SignUp(r.Context(), input, sessionMetaFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}

	if result.Token != "" {
		a.setSessionCookie(w, result.Token)
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"user":    result.User,
		"session": result.Session,
	})
}

func (a *Auth) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var input SignInInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"error":   "INVALID_BODY",
			"message": "request body must be valid JSON",
		})
		return
	}

	result, err := a.SignIn(r.Context(), input, sessionMetaFromRequest(r))
	if err != nil {
		writeError(w, err)
		return
	}

	a.setSessionCookie(w, result.Token)
	writeJSON(w, http.StatusOK, map[string]any{
		"user":    result.User,
		"session": result.Session,
	})
}

func (a *Auth) handleSignOut(w http.ResponseWriter, r *http.Request) {
	token := a.extractToken(r)
	if token != "" {
		if err := a.SignOut(r.Context(), token); err != nil {
			a.logError("signing out session", err)
		}
	}

	a.clearSessionCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{
		"message": "signed out",
	})
}

func (a *Auth) handleGetSession(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r.Context())
	if session == nil {
		writeJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHENTICATED",
			"message": "no active session",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":    session.User,
		"session": session,
	})
}

func sessionMetaFromRequest(r *http.Request) SessionMeta {
	ip := r.RemoteAddr
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		ip = forwarded
	}
	return SessionMeta{
		IPAddress: ip,
		UserAgent: r.Header.Get("User-Agent"),
	}
}

func (a *Auth) HandleGetMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, GetSession(r.Context()))
}
