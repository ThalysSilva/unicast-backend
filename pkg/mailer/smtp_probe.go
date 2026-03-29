package mailer

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"time"
)

func TestSMTPConnection(config SmtpAuthentication, timeout time.Duration) error {
	address := net.JoinHostPort(config.Host, strconv.Itoa(config.Port))

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("tcp dial failed: %w", err)
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))

	client, err := smtp.NewClient(conn, config.Host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("smtp client init failed: %w", err)
	}
	defer func() {
		_ = client.Close()
	}()

	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("smtp hello failed: %w", err)
	}

	if ok, _ := client.Extension("STARTTLS"); !ok {
		return fmt.Errorf("smtp server does not support STARTTLS")
	}

	if err := client.StartTLS(&tls.Config{ServerName: config.Host}); err != nil {
		return fmt.Errorf("smtp STARTTLS failed: %w", err)
	}

	if ok, _ := client.Extension("AUTH"); !ok {
		return fmt.Errorf("smtp server does not support AUTH")
	}

	if err := client.Auth(smtp.PlainAuth("", config.Username, config.Password, config.Host)); err != nil {
		return fmt.Errorf("smtp AUTH failed: %w", err)
	}

	if err := client.Quit(); err != nil {
		return fmt.Errorf("smtp QUIT failed: %w", err)
	}

	return nil
}
