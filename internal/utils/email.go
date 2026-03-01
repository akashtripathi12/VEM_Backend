package utils

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"

	"github.com/akashtripathi12/TBO_Backend/internal/config"
)

// SendEmail sends an email using SMTP
// It supports both STARTTLS (Port 587) and Implicit SSL (Port 465)
func SendEmail(cfg *config.Config, to []string, subject string, body string) error {
	from := cfg.SMTPEmail
	password := cfg.SMTPPass
	smtpHost := cfg.SMTPHost
	smtpPort := cfg.SMTPPort

	// Validation
	if from == "" || password == "" {
		return fmt.Errorf("SMTP credentials missing: check SMTP_EMAIL and SMTP_PASS")
	}

	// Remove spaces from Gmail App Password if present
	password = strings.ReplaceAll(password, " ", "")

	// Message construction
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s\r\n", to[0], subject, body))

	// Authentication
	auth := smtp.PlainAuth("", from, password, smtpHost)

	// Handling Implicit SSL (Port 465)
	if smtpPort == "465" {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: false,
			ServerName:         smtpHost,
		}

		conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", smtpHost, smtpPort), tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect via TLS: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, smtpHost)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Quit()

		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		if err = client.Mail(from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		for _, addr := range to {
			if err = client.Rcpt(addr); err != nil {
				return fmt.Errorf("failed to set recipient %s: %w", addr, err)
			}
		}

		w, err := client.Data()
		if err != nil {
			return fmt.Errorf("failed to open data writer: %w", err)
		}

		_, err = w.Write(message)
		if err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}

		err = w.Close()
		if err != nil {
			return fmt.Errorf("failed to close data writer: %w", err)
		}

		log.Printf("📧 Email sent to %v via Port 465 (SSL)", to)
		return nil
	}

	// Standard SMTP (Port 587/STARTTLS)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, to, message)
	if err != nil {
		return fmt.Errorf("failed to send email via Port %s: %w", smtpPort, err)
	}

	log.Printf("📧 Email sent to %v via Port %s", to, smtpPort)
	return nil
}
