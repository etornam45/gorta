package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"log"
	"os"
	"strconv"
	"strings"
	"time"

	gorta "github.com/etornam45/gorta"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"

	"github.com/etornam45/gorta/adapters/resend"
	sqladapter "github.com/etornam45/gorta/adapters/sql"
	"github.com/etornam45/gorta/core"
	emailpassword "github.com/etornam45/gorta/plugins/emailpassword"
	"github.com/etornam45/gorta/plugins/magiclink"
	"github.com/etornam45/gorta/plugins/oauth"
	"github.com/etornam45/gorta/plugins/oauth/providers"
)

func main() {
	_ = godotenv.Load()
	db, err := sql.Open("sqlite3", "./gorta.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connected")
	runMigrations(db, "adapters/sql/schema.sql")

	sessionDuration, err := parseDuration(os.Getenv("SESSION_DURATION"))
	if err != nil {
		log.Fatal("SESSION_DURATION: ", err)
	}
	secureCookies, err := parseBool(os.Getenv("COOKIE_SECURE"))
	if err != nil {
		log.Fatal("COOKIE_SECURE: ", err)
	}

	
	mailer := resend.NewMailer(resend.Config{
		APIKey:    os.Getenv("RESEND_API_KEY"),
		FromEmail: os.Getenv("FROM_EMAIL"),
		FromName:  os.Getenv("FROM_NAME"),
	})
	
	storage := sqladapter.New(db)
	emailpasswordPlugin, err := emailpassword.New(storage, storage, mailer, emailpassword.Config{
		VerifyEmail: false,
	})
	if err != nil {
		log.Fatal(err)
	}
	
	magiclinkPlugin, err := magiclink.New(storage, mailer, magiclink.Config{
		Expiration: 1 * time.Hour,
	})
	if err != nil {
		log.Fatal(err)
	}
	
	google := providers.NewGoogleProvider(providers.GoogleConfig{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	})
	github := providers.NewGitHubProvider(providers.GitHubConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
	})
	oauthPlugin, err := oauth.New(storage, storage, nil, oauth.Config{
		Providers:       []oauth.Provider{google, github},
		SuccessRedirect: os.Getenv("BASE_URL") + "/protected",
	})
	if err != nil {
		log.Fatal(err)
	}
	
	gortaConfig := gorta.Config{
		Secret:          os.Getenv("SECRET"),
		SessionDuration: sessionDuration,
		BaseURL:         os.Getenv("BASE_URL"),
		SecureCookies:   secureCookies,
		CookieDomain:    os.Getenv("COOKIE_DOMAIN"),
		CookieName:      os.Getenv("COOKIE_NAME"),
	}
	a, err := gorta.New(
		storage,
		gortaConfig,
		gorta.WithPlugin(emailpasswordPlugin),
		gorta.WithPlugin(magiclinkPlugin),
		gorta.WithPlugin(oauthPlugin),
	)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		core.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})
	mux.Handle("/auth/", http.StripPrefix("/auth", a.Handler()))
	mux.Handle("/protected", a.RequireAuth()(http.HandlerFunc(ProtectedHandler)))

	log.Println("Server is running http://localhost:8080")
	if err := http.ListenAndServe(":8080", a.Middleware()(mux)); err != nil {
		log.Fatal(err)
	}
}

func ProtectedHandler(w http.ResponseWriter, r *http.Request) {
	session := gorta.GetSession(r.Context())
	if session == nil {
		core.WriteJSON(w, http.StatusUnauthorized, map[string]string{
			"error":   "UNAUTHENTICATED",
			"message": "no active session",
		})
		return
	}
	core.WriteJSON(w, http.StatusOK, map[string]any{
		"message": "Hello, " + session.User.Name + "!",
		"session": *session,
	})
}

func runMigrations(db *sql.DB, schemaPath string) {
	content, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Println("Error reading schema file: ", err.Error())
		return
	}
	_, err = db.Exec(string(content))
	if err != nil {
		log.Println("Error running migrations: ", err.Error())
		return
	}
	log.Println("Migrations completed successfully")
}

func parseBool(s string) (bool, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return false, nil
	}
	return strconv.ParseBool(s)
}

func parseDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, nil
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid duration %q: %w", s, err)
	}
	return d, nil
}
