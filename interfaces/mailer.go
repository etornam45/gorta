package interfaces

import "context"

type MagicMailer interface {
	SendMagicLink(ctx context.Context, to, magicLink string) error
}


type VerificationMailer interface {
	SendVerificationEmail(ctx context.Context, to, verifyURL string) error
	SendPasswordReset(ctx context.Context, to, resetURL string) error
}

type Mailer interface {}