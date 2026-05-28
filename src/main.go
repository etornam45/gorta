package main

import (
	"database/sql"
	"net/http"

	"log"
	"os"
	"time"

	"github.com/etornam45/gorta"
	sqladapter "github.com/etornam45/gorta/adapters/sql"
	_ "github.com/mattn/go-sqlite3"

	"github.com/etornam45/gorta/adapters/resend"
)

func main() {
	db, err := sql.Open("sqlite3", "./gorta.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("Database connected")
	runMigrations(db, "adapters/sql/schema.sql")

	mailer := resend.NewMailer(resend.Config{
		APIKey:    "Your API Key",
		FromEmail: "your@email.com",
		FromName:  "Your Name",
	})

	a, err := gorta.New(sqladapter.New(db), gorta.Config{
		Secret:            "asdfadflajsdkfjalksdjfkljaskdjfldsafsadfadsfsd",
		SessionDuration:   1 * time.Hour,
		SecureCookies:     false,
		CookieDomain:      "localhost",
		CookieName:        "gorta_session",
		VerifyEmail:       true,
		VerifyEmailDomain: "localhost:8080",
	}, mailer)
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	log.Println("Server is running http://localhost:8080")

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		gorta.WriteJSON(w, http.StatusOK, map[string]string{
			"message": "Hello, World!",
		})
	})
	mux.Handle("/auth/", http.StripPrefix("/auth", a.Handler()))
	mux.Handle("/protected", a.RequireAuth()(http.HandlerFunc(ProtectedHandler)))
	http.ListenAndServe(":8080", a.Middleware()(mux))
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
