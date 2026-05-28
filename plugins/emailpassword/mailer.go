package emailpassword

import "context"

type Mailer interface {
	SendVerificationEmail(ctx context.Context, to, verifyURL string) error
	SendPasswordReset(ctx context.Context, to, resetURL string) error
}