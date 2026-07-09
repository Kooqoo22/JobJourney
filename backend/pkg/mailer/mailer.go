package mailer

type Mailer interface {
	SendVerificationEmail(to, name, link string) error
	SendPasswordResetEmail(to, name, link string) error
}
