package mail

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"

	gomail "github.com/mrz1836/go-mail"
)

type MailInterface interface {
	Send(cx context.Context) error
}

type Mail struct {
	Mail     *gomail.MailService
	Provider gomail.ServiceProvider
}

func (m *Mail) NewMail() *Mail {
	mail := new(gomail.MailService)
	mail.FromName = "No Reply"
	mail.FromUsername = "no-reply"
	mail.FromDomain = os.Getenv("EMAIL_FROM_DOMAIN")

	// Provider
	mail.SMTPHost = os.Getenv("EMAIL_SMTP_HOST")
	mail.SMTPPort, _ = strconv.Atoi(os.Getenv("EMAIL_SMTP_PORT"))
	mail.SMTPUsername = os.Getenv("EMAIL_SMTP_USERNAME")
	mail.SMTPPassword = os.Getenv("EMAIL_SMTP_PASSWORD")

	provider := gomail.SMTP

	// Start the service
	err := mail.StartUp()
	if err != nil {
		log.Printf("error in StartUp: %s using provider: %x", err.Error(), provider)
		panic("Can't start mail service")
	}

	return &Mail{Mail: mail, Provider: provider}
}

func (mail *Mail) Send(ctx context.Context, to string, content string, subject string) error {
	m := mail.Mail.NewEmail()
	m.HTMLContent = fmt.Sprintf("<html><body>%s</body></html>", content)
	m.Recipients = []string{to}
	m.Subject = subject
	if err := mail.Mail.SendEmail(context.Background(), m, mail.Provider); err != nil {
		log.Printf("error in SendEmail: %s using provider: %x", err.Error(), mail.Provider)
		return err
	}

	return nil
}
