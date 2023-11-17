package smtp

import (
	"bytes"
	"time"

	"github.com/ry-animal/go-chi-api/assets"
	"github.com/ry-animal/go-chi-api/internal/funcs"

	"github.com/go-mail/mail/v2"

	htmlTemplate "html/template"
	textTemplate "text/template"
)

type Mailer struct {
	dialer *mail.Dialer
	from   string
}

func NewMailer(host string, port int, username, password, from string) *Mailer {
	dialer := mail.NewDialer(host, port, username, password)
	dialer.Timeout = 5 * time.Second

	return &Mailer{
		dialer: dialer,
		from:   from,
	}
}

func (m *Mailer) Send(recipient string, data any, patterns ...string) error {
	for i := range patterns {
		patterns[i] = "emails/" + patterns[i]
	}

	msg := mail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.from)

	ts, err := textTemplate.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, patterns...)
	if err != nil {
		return err
	}

	subject := new(bytes.Buffer)
	err = ts.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	msg.SetHeader("Subject", subject.String())

	plainBody := new(bytes.Buffer)
	err = ts.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	msg.SetBody("text/plain", plainBody.String())

	if ts.Lookup("htmlBody") != nil {
		ts, err := htmlTemplate.New("").Funcs(funcs.TemplateFuncs).ParseFS(assets.EmbeddedFiles, patterns...)
		if err != nil {
			return err
		}

		htmlBody := new(bytes.Buffer)
		err = ts.ExecuteTemplate(htmlBody, "htmlBody", data)
		if err != nil {
			return err
		}

		msg.AddAlternative("text/html", htmlBody.String())
	}

	for i := 1; i <= 3; i++ {
		err = m.dialer.DialAndSend(msg)

		if nil == err {
			return nil
		}

		time.Sleep(2 * time.Second)
	}

	return err
}
