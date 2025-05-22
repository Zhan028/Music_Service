package email

import (
	"gopkg.in/gomail.v2"
)

type GomailSender struct {
	From     string
	Host     string
	Port     int
	Username string
	Password string
}

func NewGomailSender(from, host, username, password string, port int) *GomailSender {
	return &GomailSender{
		From:     from,
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
	}
}

func (s *GomailSender) SendEmail(to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.From)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(s.Host, s.Port, s.Username, s.Password)
	return d.DialAndSend(m)
}
