package main

import (
	"net/smtp"
	"strings"
)

type EmailSender interface {
	SendEmail(from string, to []string, subject string, body string) error
}

type StdMailer struct {
	auth          smtp.Auth
	smtpServerURL string
}

func NewStdMailer(username, password, host, smtpServerURL string) *StdMailer {
	return &StdMailer{
		auth:          smtp.PlainAuth("", username, password, host),
		smtpServerURL: smtpServerURL,
	}
}

func (m *StdMailer) SendEmail(from string, to []string, subject string, body string) error {
	emailTo := strings.Join(to, ",")
	msg := []byte("To: " + emailTo + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body)

	err := smtp.SendMail(m.smtpServerURL, m.auth, from, to, msg)
	if err != nil {
		return err
	}

	return nil
}
