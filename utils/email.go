package utils

import (
	"mywetransfergo/config"

	"gopkg.in/gomail.v2"
)

func SendMail(receiver, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", config.App.SMTPFrom)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	d := gomail.NewDialer(
		config.App.SMTPHost,
		config.App.SMTPPort,
		config.App.SMTPUsername,
		config.App.SMTPPassword,
	)

	return d.DialAndSend(m)
}
