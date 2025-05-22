package usecase

type EmailSender interface {
	SendEmail(to, subject, body string) error
}
