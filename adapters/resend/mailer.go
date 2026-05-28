package resend

import (
	"context"
	"fmt"

	"github.com/resend/resend-go/v2"

	"github.com/etornam45/gorta/interfaces"
)

type Mailer struct {
	client *resend.Client
	config Config
}

type Config struct {
	FromEmail string
	FromName  string
	APIKey    string
}

func NewMailer(config Config) interfaces.Mailer {	
	return &Mailer{
		client: resend.NewClient(config.APIKey),
		config: config,
	}
}



func (m *Mailer) SendVerificationEmail(ctx context.Context, to, verifyURL string) error {
	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    m.config.FromEmail,
		To:      []string{to},
		Subject: "Verify your email",
		Html:    fmt.Sprintf("<p>Click <a href='%s'>here</a> to verify your email</p>", verifyURL),
		Text:    fmt.Sprintf("Click %s to verify your email", verifyURL),
	})
	return err
}

func (m *Mailer) SendPasswordReset(ctx context.Context, to, resetURL string) error {
	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    m.config.FromEmail,
		To:      []string{to},
		Subject: "Reset your password",
		Html:    fmt.Sprintf("<p>Click <a href='%s'>here</a> to reset your password</p>", resetURL),
		Text:    fmt.Sprintf("Click %s to reset your password", resetURL),
	})
	return err
}


func (m *Mailer) SendMagicLink(ctx context.Context, to, magicLink string) error {
	_, err := m.client.Emails.SendWithContext(ctx, &resend.SendEmailRequest{
		From:    m.config.FromEmail,
		To:      []string{to},
		Subject: "Magic Link",
		Html:    fmt.Sprintf("<p>Click <a href='%s'>here</a> to login</p>", magicLink),
		Text:    fmt.Sprintf("Click %s to login", magicLink),
	})
	return err
}