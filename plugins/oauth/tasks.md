

Callback arrives with a verified Google identity.

Case 1: Account exists (returning user)
  FindAccountByProvider("google", googleUserID) → found
  → load user from account.UserID
  → create session

Case 2: Account doesn't exist, email matches existing user (account linking)
  FindUserByEmail(googleEmail) → found
  → create Account linking this user to Google
  → create session

Case 3: Brand new user
  CreateUser(name, email, picture from Google)
  → create Account
  → create session



API 

```go			
google := oauth.NewGoogleProvider(oauth.GoogleConfig{
	ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
	ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
	// RedirectURL built from core.Config().BaseURL automatically
})

github := oauth.NewGitHubProvider(oauth.GitHubConfig{
	ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
	ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
})

auth, err := gorta.New(store, config,
	gorta.WithPlugin(emailpassword.New(store, store, mailer, emailpassword.Config{})),
	gorta.WithPlugin(
		oauth.New(store, store, oauth.Config{
			Providers: []oauth.Provider{google, github},
		}),
	),
)
```