package utils

import (
	"os"
	"strconv"

	"gopkg.in/gomail.v2"
)

func SendMailWithAttachment(sender, receiver, subject, body, attachmentPath string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", sender)
	m.SetHeader("To", receiver)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)
	m.Attach(attachmentPath)

	port, _ := strconv.Atoi(os.Getenv("SMTP_PORT"))
	d := gomail.NewDialer(
		os.Getenv("SMTP_HOST"),
		port,
		os.Getenv("SMTP_USERNAME"),
		os.Getenv("SMTP_PASSWORD"),
	)

	return d.DialAndSend(m)
}

// Ensure the file ends with a newline
