package magiclink

import "context"

type Mailer interface {
	SendMagicLink(ctx context.Context, to, magicLink string) error
}