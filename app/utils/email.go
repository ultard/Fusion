package utils

import (
	"bytes"
	"fmt"
	"gopkg.in/gomail.v2"
	"html/template"
	"os"
	"path/filepath"
)

type EmailService interface {
	SendEmail(to, subject, templateName string, data interface{}) error
}

type emailService struct {
	smtpHost string
	smtpPort int
	username string
	password string
	sender   string
}

func NewEmailService(smtpHost string, smtpPort int, username, password, sender string) EmailService {
	return &emailService{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
		username: username,
		password: password,
		sender:   sender,
	}
}

// SendEmail отправляет электронное письмо с html-шаблоном
func (s *emailService) SendEmail(to, subject, templateName string, data interface{}) error {
	baseDir := "templates"
	cleanTemplateName := filepath.Clean(templateName)

	templatePath := fmt.Sprintf("%s/%s.html", baseDir, cleanTemplateName)
	tmplContent, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New("email").Parse(string(tmplContent))
	if err != nil {
		return err
	}

	var tpl bytes.Buffer
	if err := tmpl.Execute(&tpl, data); err != nil {
		return err
	}

	m := gomail.NewMessage()
	m.SetHeader("From", s.sender)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", tpl.String())

	d := gomail.NewDialer(s.smtpHost, s.smtpPort, s.username, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
