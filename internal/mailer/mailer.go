package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	"github.com/go-mail/mail/v2"
)

/* Embeds the /templates files into the program. */
//go:embed templates/*
var templateFS embed.FS

/* Connection to the SMTP server */
type Mailer struct {
	dailer *mail.Dialer
	sender string
}

/* Sets up a new SMTP server */
func New(host string, port int, username string, password string, sender string) Mailer {
	dailer := mail.NewDialer(host, port, username, password)
	dailer.Timeout = 5 * time.Second

	return Mailer{
		dailer: dailer,
		sender: sender,
	}
}

/* Sends the activation email to the user */
func (m Mailer) Send(recipient string, tmplFile string, data any) error {
	// create email to send
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+tmplFile)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())

	// attempt to send message 3 times, else give up
	for i := 1; i <= 3; i++ {
		err = m.dailer.DialAndSend(msg)
		if err == nil {
			return nil
		}

		time.Sleep(1 * time.Second)
	}

	return err
}
