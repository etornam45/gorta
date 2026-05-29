<p align="center">
  <img src="./gorta.png" width="100" />
</p>

**`[Gorta]`** is a minimal, adapter-based authentication library for Go.

#### Manifesto
Gorta handles the hard parts of auth (password hashing, session management,
OAuth flows, and secure cookie handling) while leaving data persistence
entirely to you. You implement the [Adapter] interface against your own
database, ORM, and schema. Gorta never touches your database directly.

#### Quick Start

```go	
import github.com/etornam45/gorta
```

**Storage**
```go
storage := sqladapter.New(db)
```
**Plugins**
```go
emailpasswordPlugin, err := emailpassword.New(storage, storage, mailer, emailpassword.Config{
	VerifyEmail: true,
})

magiclinkPlugin, err := magiclink.New(storage, mailer, magiclink.Config{
	Expiration: 1 * time.Hour,
})
```
**Usage**
```go
a, err := gorta.New(
	storage,
	config,
	gorta.WithPlugin(emailpasswordPlugin),
	gorta.WithPlugin(magiclinkPlugin),
)

// Mount Gorta's HTTP routes
mux.Handle("/auth/", a.Handler())

// Protect routes with middleware
mux.Handle("/dashboard", a.RequireAuth()(dashboardHandler))

// Apply session loading to all routes
http.ListenAndServe(":8080", a.Middleware()(mux))
```


#### Sending Email
Now every plugin *may have* it's own `storage` and `plugin` interface you must implement

```go
// SQL storage for magiclink
type Storage interface {
	CreateVerification(ctx context.Context, v Verification) (*Verification, error)
	FindVerificationByToken(ctx context.Context, token string) (*Verification, error)
	DeleteVerification(ctx context.Context, id string) error
}

// the DB layer
type SQLAdapter struct {
	db *sql.DB
}

func New(db *sql.DB) *SQLAdapter {
	return &SQLAdapter{db: db}
}

func (a *SQLAdapter) CreateVerification(ctx context.Context, v magiclink.Verification) (*magiclink.Verification, error) {
	_, err := a.db.ExecContext(ctx,
		`INSERT INTO verifications (id, identifier, token, expires_at, created_at)
		 VALUES ($1, $2, $3, $4, $5)`,
		v.ID, v.Identifier, v.Token, v.ExpiresAt, v.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("sqladapter: creating verification: %w", err)
	}
	return &v, nil
}
```


#### OAuth
You only have Google and Github. Hoping to add more providers

```go
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
```


#### Auth Methods

- [x] Email Passord
- [x] Social Login (Google, Github)
- [x] Magic Links
- [ ] Passkeys (WebAuthn)
- [ ] Email OTP
- [ ] Phone Number
- [ ] Username
- [ ] Anonymous
- [ ] One‑Tap Sign‑In


#### Security & Compliance

- [ ] Two‑Factor Authentication (2FA / TOTP)
- [ ] Captcha
- [ ] Rate Limiting
- [ ] CSRF Protection


#### Payments & Billing
- [ ] Stripe
- [ ] Polar
- [ ] Autumn
- [ ] Creem