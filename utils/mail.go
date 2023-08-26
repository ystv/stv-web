package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

type (
	// Mailer encapsulates the dependency
	Mailer struct {
		*mail.SMTPClient
		Defaults Defaults
	}

	// Defaults allows for default recipients to be set
	Defaults struct {
		DefaultTo   string
		DefaultCC   []string
		DefaultBCC  []string
		DefaultFrom string
	}

	// MailConfig represents a configuration to connect to an SMTP server
	MailConfig struct {
		Host     string
		Port     int
		Username string
		Password string
	}

	// Mail represents an email to be sent
	Mail struct {
		Subject     string
		To          string
		Cc          []string
		Bcc         []string
		From        string
		Error       error
		UseDefaults bool
		Tpl         *template.Template
		TplData     interface{}
	}
)

// NewMailer creates a new SMTP client
func NewMailer(config MailConfig) (*Mailer, error) {
	smtpServer := mail.SMTPServer{
		Host:           config.Host,
		Port:           config.Port,
		Username:       config.Username,
		Password:       config.Password,
		Encryption:     mail.EncryptionSTARTTLS,
		Authentication: mail.AuthLogin,
		ConnectTimeout: 10 * time.Second,
		SendTimeout:    10 * time.Second,
		KeepAlive:      true,
		TLSConfig:      &tls.Config{InsecureSkipVerify: true},
	}

	smtpServer.KeepAlive = true

	smtpClient, err := smtpServer.Connect()
	if err != nil {
		return nil, err
	}
	return &Mailer{smtpClient, Defaults{}}, nil
}

// AddDefaults adds the default recipients
func (m *Mailer) AddDefaults(defaults Defaults) {
	m.Defaults.DefaultTo = defaults.DefaultTo
	m.Defaults.DefaultCC = defaults.DefaultCC
	m.Defaults.DefaultBCC = defaults.DefaultBCC
	m.Defaults.DefaultFrom = defaults.DefaultFrom
}

// CheckSendable verifies that the email can be sent
func (m *Mailer) CheckSendable(item Mail) error {
	if item.UseDefaults {
		if m.Defaults.DefaultTo == "" {
			if item.To == "" {
				return fmt.Errorf("field UseDefaults was used and there were no defaults and to field set")
			}
		}
	} else if item.To == "" {
		return fmt.Errorf("no To field is set")
	}
	return nil
}

// SendMail sends a template email
func (m *Mailer) SendMail(item Mail) error {
	err := m.CheckSendable(item)
	if err != nil {
		return err
	}
	var (
		to, from string
		cc, bcc  []string
	)
	if item.UseDefaults {
		to = m.Defaults.DefaultTo
		cc = m.Defaults.DefaultCC
		bcc = m.Defaults.DefaultBCC
		from = m.Defaults.DefaultFrom
	} else {
		to = item.To
		cc = item.Cc
		bcc = item.Bcc
		from = item.From
	}
	body := bytes.Buffer{}
	err = item.Tpl.Execute(&body, item.TplData)
	if err != nil {
		return fmt.Errorf("failed to exec tpl: %w", err)
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject(item.Subject)
	if len(item.Cc) != 0 {
		email.AddCc(cc...)
	}
	if len(item.Bcc) != 0 {
		email.AddBcc(bcc...)
	}
	email.SetBody(mail.TextHTML, body.String())
	if email.Error != nil {
		return fmt.Errorf("failed to set mail data: %w", email.Error)
	}
	return email.Send(m.SMTPClient)
}

// SendErrorMail sends a standard template error email
func (m *Mailer) SendErrorMail(item Mail) error {
	err := m.CheckSendable(item)
	if err != nil {
		return err
	}
	var (
		to, from string
		cc, bcc  []string
	)
	if item.UseDefaults {
		to = m.Defaults.DefaultTo
		cc = m.Defaults.DefaultCC
		bcc = m.Defaults.DefaultBCC
		from = m.Defaults.DefaultFrom
	} else {
		to = item.To
		cc = item.Cc
		bcc = item.Bcc
		from = item.From
	}
	errorTemplate := template.New("Error Template")
	errorTemplate = template.Must(errorTemplate.Parse("An error occurred!<br><br>{{.}}<br><br>>We apologise for the inconvenience."))
	body := bytes.Buffer{}
	err = errorTemplate.Execute(&body, item.Error)
	if err != nil {
		return fmt.Errorf("failed to exec tpl: %w", err)
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject("Non-fatal error - YSTV STV")
	if len(cc) != 0 {
		email.AddCc(cc...)
	}
	if len(bcc) != 0 {
		email.AddBcc(bcc...)
	}
	email.SetBody(mail.TextHTML, body.String())
	if email.Error != nil {
		return fmt.Errorf("failed to set mail data: %w", email.Error)
	}
	return email.Send(m.SMTPClient)
}

// SendErrorFatalMail sends a standard template error fatal email
func (m *Mailer) SendErrorFatalMail(item Mail) error {
	err := m.CheckSendable(item)
	if err != nil {
		return err
	}
	var (
		to, from string
		cc, bcc  []string
	)
	if item.UseDefaults {
		to = m.Defaults.DefaultTo
		cc = m.Defaults.DefaultCC
		bcc = m.Defaults.DefaultBCC
		from = m.Defaults.DefaultFrom
	} else {
		to = item.To
		cc = item.Cc
		bcc = item.Bcc
		from = item.From
	}
	errorTemplate := template.New("Fatal Error Template")
	errorTemplate = template.Must(errorTemplate.Parse("<body><p style=\"color: red;\">A <b>FATAL ERROR</b> OCCURRED!<br><br><code>{{.}}</code></p><br><br>We apologise for the inconvenience.</body>"))
	body := bytes.Buffer{}
	err = errorTemplate.Execute(&body, item.Error)
	if err != nil {
		return fmt.Errorf("failed to exec tpl: %w", err)
	}
	email := mail.NewMSG()
	email.SetFrom(from).AddTo(to).SetSubject("FATAL ERROR - YSTV STV")
	if len(cc) != 0 {
		email.AddCc(cc...)
	}
	if len(bcc) != 0 {
		email.AddBcc(bcc...)
	}
	email.SetBody(mail.TextHTML, body.String())
	if email.Error != nil {
		return fmt.Errorf("failed to set mail data: %w", email.Error)
	}
	return email.Send(m.SMTPClient)
}
