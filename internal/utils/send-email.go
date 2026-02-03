package utils

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"e-commerce.com/internal/config"
)

func SendVerificationEmail(to, verificationLink string) error {
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"
	senderEmail := config.AppConfig.SMTPEmail
	appPassword := config.AppConfig.SMTPPassword

	// Email content
	subject := "Verify Your Email Address"
	body := fmt.Sprintf(`
	<html>
		<body>
			<h2>Welcome to Our Service!</h2>
			<p>Please click the link below to verify your email address:</p>
			<a href="%s">Verify Email</a>
			<p>If you didn't request this, please ignore this email.</p>
		</body>
	</html>`, verificationLink)

	// Email headers
	headers := map[string]string{
		"From":         senderEmail,
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=utf-8",
	}

	// Build the message
	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n" + body)

	// Setup authentication
	auth := smtp.PlainAuth("", senderEmail, appPassword, smtpHost)

	// Set up a dialer with timeout
	conn, err := net.DialTimeout("tcp", smtpHost+":"+smtpPort, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, smtpHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	if err = client.StartTLS(&tls.Config{
		ServerName: smtpHost,
	}); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	if err = client.Mail(senderEmail); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get writer: %w", err)
	}

	_, err = writer.Write([]byte(message.String()))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}
