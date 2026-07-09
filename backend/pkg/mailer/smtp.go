package mailer

import (
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/Kooqoo22/JobJourney/backend/config"
)

var ErrNotConfigured = errors.New("smtp is not configured")

type smtpMailer struct {
	cfg config.SMTPConfig
}

func NewSMTP(cfg config.SMTPConfig) Mailer {
	return &smtpMailer{cfg: cfg}
}

func (m *smtpMailer) SendVerificationEmail(to, name, link string) error {
	subject := "Verify your JobJourney email"
	body := fmt.Sprintf(
		"<p>Hi %s,</p><p>Welcome to JobJourney. Confirm your email address to activate email reminders.</p><p><a href=\"%s\">Verify my email</a></p><p>This link expires in 24 hours. If you did not create an account, ignore this email.</p>",
		htmlEscape(name), htmlEscape(link),
	)
	return m.send(to, subject, body)
}

func (m *smtpMailer) SendPasswordResetEmail(to, name, link string) error {
	subject := "Reset your JobJourney password"
	body := fmt.Sprintf(
		"<p>Hi %s,</p><p>We received a request to reset your password. Choose a new one using the link below.</p><p><a href=\"%s\">Reset my password</a></p><p>This link expires in 1 hour and can be used once. If you did not request this, ignore this email.</p>",
		htmlEscape(name), htmlEscape(link),
	)
	return m.send(to, subject, body)
}

func (m *smtpMailer) send(to, subject, htmlBody string) error {
	if m.cfg.Host == "" || m.cfg.FromEmail == "" {
		return ErrNotConfigured
	}

	from := m.cfg.FromEmail
	fromHeader := from
	if m.cfg.FromName != "" {
		fromHeader = fmt.Sprintf("%s <%s>", m.cfg.FromName, from)
	}

	msg := strings.Join([]string{
		"From: " + fromHeader,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=\"UTF-8\"",
		"",
		htmlBody,
	}, "\r\n")

	addr := m.cfg.Host + ":" + m.cfg.Port
	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}
	return smtp.SendMail(addr, auth, from, []string{to}, []byte(msg))
}

func htmlEscape(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
	)
	return replacer.Replace(s)
}
