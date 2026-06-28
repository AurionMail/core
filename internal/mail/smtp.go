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
		"password_len", len(password),
	)

	// 1. Connexion au serveur (Port 587) avec timeout
	var d net.Dialer
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	b.logger.Info("Dialing TCP connection to mail server")
	conn, err := d.DialContext(dialCtx, "tcp", b.ServerAddr)
	if err != nil {
		b.logger.Error("TCP dial failed", "error", err)
		return false, err
	}
	defer conn.Close()

	b.logger.Info("Initializing SMTP client over TCP")
	client, err := smtp.NewClient(conn, b.ServerAddr)
	if err != nil {
		b.logger.Error("Failed to create SMTP client", "error", err)
		return false, err
	}
	defer client.Close()

	// 2. Passage en TLS (STARTTLS)
	host, _, err := net.SplitHostPort(b.ServerAddr)
	if err != nil {
		b.logger.Error("Failed to split host and port", "server_addr", b.ServerAddr, "error", err)
		host = b.ServerAddr
	}

	b.logger.Info("Sending STARTTLS command", "server_name_used", host)
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false, // Let's Encrypt valide
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		b.logger.Error("STARTTLS handshake failed", "error", err)
		return false, err
	}
	b.logger.Info("STARTTLS handshake successful, connection is now encrypted")

	// 3. Authentification standard
	// Note : Go utilise en interne le 'host' passé ici pour valider la sécurité
	b.logger.Info("Attempting SASL PLAIN authentication",
		"auth_host_param", host,
		"auth_email", email,
	)

	auth := smtp.PlainAuth("", email, password, host)
	if err := client.Auth(auth); err != nil {
		// C'est ici que l'erreur brute de Stalwart ou de Go va être enregistrée
		b.logger.Warn("SMTP authentication rejected by server",
			"email", email,
			"error", err,
			"error_type", string(err.Error()),
		)
		return false, nil
	}

	b.logger.Info("SMTP authentication successful", "email", email)
	return true, nil
}
