`[Gorta]` is a minimal, adapter-based authentication library for Go.

#### Manifesto
Gorta handles the hard parts of auth (password hashing, session management,
OAuth flows, and secure cookie handling) while leaving data persistence
entirely to you. You implement the [Adapter] interface against your own
database, ORM, and schema. Gorta never touches your database directly.

#### Quick Start

```go
a, err := gorta.New(myAdapter, gorta.Config{
	Secret: os.Getenv("AUTH_SECRET"), // must be 32+ characters
})

// Mount Gorta's HTTP routes
mux.Handle("/auth/", a.Handler())

// Protect routes with middleware
mux.Handle("/dashboard", a.RequireAuth()(dashboardHandler))

// Apply session loading to all routes
http.ListenAndServe(":8080", a.Middleware()(mux))
```


#### Sending Email
Sending verification emails is oprional

To create your oun `mailer` implement this interface

```go
type Mailer interface {
	SendVerificationEmail(ctx context.Context, to, verifyURL string) error
	SendPasswordReset(ctx context.Context, to, resetURL string) error
}
```

#### Resend Mailer Example

```go
// Usage
mailer := resend.NewMailer(resend.Config{
	APIKey:    "YOUR_API_KEY",
	FromEmail: "email",
	FromName:  "Name",
})

auth, err := gorta.New(sqladapter.New(db), config, mailer)
```


### Errors

Gorta defines sentinel errors in errors.go. Your Adapter must return
these (e.g. [ErrUserNotFound]) rather than raw database errors, so
Gorta can map them to the correct HTTP responses.