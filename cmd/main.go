package main

import (
	"database/sql"
	"net/http"

	"log"
	"os"
	"time"

	gorta "github.com/etornam45/gorta"
	_ "github.com/mattn/go-sqlite3"

	"github.com/etornam45/gorta/adapters/resend"
	sqladapter "github.com/etornam45/gorta/adapters/sql"
	emailpassword "github.com/etornam45/gorta/plugins/emailpassword"
	"github.com/etornam45/gorta/plugins/magiclink"
	"github.com/etornam45/gorta/plugins/oauth"
	"github.com/etornam45/gorta/plugins/oauth/providers"
)

func main() {
	db, err := sql.Open("sqlite3", "./gorta.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connected")
	runMigrations(db, "adapters/sql/schema.sql")
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatal(err)
	}
	gortaConfig := gorta.Config{
		Secret:          cfg.Secret,
		SessionDuration: cfg.SessionDuration,
		BaseURL:         cfg.BaseURL,
		SecureCookies:   cfg.SecureCookies,
		CookieDomain:    cfg.CookieDomain,
		CookieName:      cfg.CookieName,
	}

	mailer := resend.NewMailer(resend.Config{
		APIKey:    cfg.ResendAPIKey,
		FromEmail: cfg.FromEmail,
		FromName:  cfg.FromName,
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
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
	})
	github := providers.NewGitHubProvider(providers.GitHubConfig{
		ClientID:     cfg.GitHubClientID,
		ClientSecret: cfg.GitHubClientSecret,
	})
	oauthPlugin, err := oauth.New(storage, storage, nil, oauth.Config{
		Providers:       []oauth.Provider{google, github},
		SuccessRedirect: cfg.BaseURL + "/protected",
	})
	if err != nil {
		log.Fatal(err)
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
		gorta.WriteJSON(w, http.StatusOK, map[string]string{
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
	user := gorta.GetUser(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	gorta.WriteJSON(w, http.StatusOK, map[string]string{
		"message": "Hello, " + user.Name + "!",
		"user":    user.Email,
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
