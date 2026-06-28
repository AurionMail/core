// mail/smtp.go
package mail

import (
	"context"
	"crypto/tls"
	"log/slog"
	"net"
	"net/smtp"
	"time"
)

type SMTPBackend struct {
	ServerAddr string
	logger     *slog.Logger
}

func NewSMTPBackend(addr string, logger *slog.Logger) *SMTPBackend {
	return &SMTPBackend{
		ServerAddr: addr,
		logger:     logger,
	}
}

func (b *SMTPBackend) VerifyCredentials(ctx context.Context, email, password string) (bool, error) {
	b.logger.Info("Starting credentials verification",
		"server_addr", b.ServerAddr,
		"email", email,
	)

	var d net.Dialer
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	// 1. Extraction propre du Hostname sans le port dès le départ
	host, _, err := net.SplitHostPort(b.ServerAddr)
	if err != nil {
		b.logger.Error("Failed to split host and port", "server_addr", b.ServerAddr, "error", err)
		host = b.ServerAddr
	}

	conn, err := d.DialContext(dialCtx, "tcp", b.ServerAddr)
	if err != nil {
		b.logger.Error("TCP dial failed", "error", err)
		return false, err
	}
	defer conn.Close()

	// CRITIQUE : Passer 'host' (sans le port) et non 'b.ServerAddr'
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		b.logger.Error("Failed to create SMTP client", "error", err)
		return false, err
	}
	defer client.Close()

	// 2. Passage en TLS
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		b.logger.Error("STARTTLS handshake failed", "error", err)
		return false, err
	}
	b.logger.Info("STARTTLS handshake successful, connection is now encrypted")

	// 3. Authentification standard
	b.logger.Info("Attempting SASL PLAIN authentication", "auth_host", host, "auth_email", email)

	// Ici, 'host' matchera exactement le 'host' de NewClient et le EHLO de Stalwart
	auth := smtp.PlainAuth("", email, password, host)
	if err := client.Auth(auth); err != nil {
		b.logger.Warn("SMTP authentication rejected by server",
			"email", email,
			"error", err,
		)
		return false, nil
	}

	b.logger.Info("SMTP authentication successful", "email", email)
	return true, nil
}
