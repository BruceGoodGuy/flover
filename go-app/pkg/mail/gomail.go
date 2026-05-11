package mail

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"text/template"

	gomail "github.com/mrz1836/go-mail"
)

type MailInterface interface {
	Send(ctx context.Context, to string, subject string, templateName string, data any) error
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

func (mail *Mail) Send(ctx context.Context, to string, subject string, templateName string, data any) error {
	// 1. Build the path to the template file
	// The templates are located in the "templates" folder at the project root
	templatePath := filepath.Join("templates", templateName+".html")

	// 2. Parse and execute the HTML template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("failed to parse template %s at %s: %w", templateName, templatePath, err)
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	// 3. Prepare the email using go-mail
	m := mail.Mail.NewEmail()
	m.HTMLContent = body.String() // The executed HTML string
	m.Recipients = []string{to}
	m.Subject = subject

	// 4. Send the email
	if err := mail.Mail.SendEmail(ctx, m, mail.Provider); err != nil {
		fmt.Printf("error in SendEmail: %s using provider: %v\n", err.Error(), mail.Provider)
		return err
	}

	return nil
}
