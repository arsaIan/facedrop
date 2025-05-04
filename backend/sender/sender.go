package sender

import (
	"bytes"
	"crypto/tls"
	"facedrop/config"
	"facedrop/logger"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
)

type EmailSender struct {
	cfg *config.Config
}

type EmailData struct {
	Subject      string
	Body         string
	DownloadLink string
}

func NewEmailSender(cfg *config.Config) *EmailSender {
	return &EmailSender{cfg: cfg}
}

func (e *EmailSender) SendZipFile(to []string, subject string, body string, downloadLink string) error {
	// Check if the necessary config parameters are provided
	if e.cfg.EmailConfig.Username == "" || e.cfg.EmailConfig.Password == "" || e.cfg.EmailConfig.SMTPHost == "" || e.cfg.EmailConfig.SMTPPort == 0 {
		return fmt.Errorf("username or password or smtp host or smtp port is empty")
	}

	// Create HTML content
	tmpl, err := template.New("email").Parse(emailTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var htmlContent bytes.Buffer
	err = tmpl.Execute(&htmlContent, EmailData{
		Subject:      subject,
		Body:         body,
		DownloadLink: downloadLink,
	})
	if err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	// Create the email message
	var message bytes.Buffer

	// Write headers
	message.WriteString(fmt.Sprintf("From: %s\r\n", e.cfg.EmailConfig.From))
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ",")))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	message.WriteString("\r\n")

	// Write HTML content
	message.Write(htmlContent.Bytes())

	// Create TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: false,
		ServerName:         e.cfg.EmailConfig.SMTPHost,
	}

	// Connect to the SMTP server
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%d", e.cfg.EmailConfig.SMTPHost, e.cfg.EmailConfig.SMTPPort), tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, e.cfg.EmailConfig.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Authenticate
	auth := smtp.PlainAuth("", e.cfg.EmailConfig.Username, e.cfg.EmailConfig.Password, e.cfg.EmailConfig.SMTPHost)
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	// Set sender and recipients
	if err := client.Mail(e.cfg.EmailConfig.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", addr, err)
		}
	}

	// Send the email
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	_, err = w.Write(message.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	logger.Info("Email sent successfully", logger.String("to", strings.Join(to, ",")))
	return nil
}
