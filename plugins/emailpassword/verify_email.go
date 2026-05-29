package emailpassword

import (
	"context"
	"time"

	"github.com/etornam45/gorta/core"
	"fmt"
)

func (p *Plugin) VerifyEmail(ctx context.Context, token string) (*core.User, error) {
	if token == "" {
		return nil, core.ErrInvalidToken
	}
	verification, err := p.creds.EmailPasswordFindVerificationByToken(ctx, token)
	if err != nil {
		fmt.Printf("[gorta] error finding email password verification: %v\n", err)
		return nil, core.ErrVerificationNotFound
	}
	if verification.ExpiresAt.Before(time.Now()) {
		fmt.Printf("[gorta] error email password verification expired: %v\n", err)
		return nil, core.ErrVerificationExpired
	}
	user, err := p.storage.FindUserByEmail(ctx, verification.Identifier)
	if err != nil {
		fmt.Printf("[gorta] error finding user: %v\n", err)
		return nil, core.ErrUserNotFound
	}
	if user.EmailVerified {
		return nil, core.ErrEmailNotVerified
	}
	_, err = p.storage.UpdateUser(ctx, user.ID, core.UpdateUserInput{
		EmailVerified: &[]bool{true}[0],
	})
	if err != nil {
		fmt.Printf("[gorta] error updating user: %v\n", err)
		return nil, err
	}

	// clean up the verification
	if err := p.creds.DeleteVerification(ctx, verification.ID); err != nil {
		fmt.Printf("[gorta] error deleting verification: %v\n", err)
	}
	return user, nil
}
