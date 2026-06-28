// mail/smtp.go
package mail

import (
	"context"
	"crypto/tls"
	"net"
	"net/smtp"
	"time"
)

type SMTPBackend struct {
	ServerAddr string // Exemple: "localhost:587" ou "mail.aurion.lan:587"
}

func NewSMTPBackend(addr string) *SMTPBackend {
	return &SMTPBackend{ServerAddr: addr}
}

func (b *SMTPBackend) VerifyCredentials(ctx context.Context, email, password string) (bool, error) {
	// 1. Connexion au serveur (Port 587) avec timeout
	var d net.Dialer
	dialCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	conn, err := d.DialContext(dialCtx, "tcp", b.ServerAddr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, b.ServerAddr)
	if err != nil {
		return false, err
	}
	defer client.Close()

	// 2. Passage en TLS (STARTTLS) - Requis pour sécuriser l'envoi du mot de passe
	host, _, _ := net.SplitHostPort(b.ServerAddr)
	tlsConfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}
	if err := client.StartTLS(tlsConfig); err != nil {
		return false, err
	}

	// 3. Authentification standard (Go gère tout car il sait que la connexion est TLS)
	auth := smtp.PlainAuth("", email, password, host)
	if err := client.Auth(auth); err != nil {
		return false, nil // Mauvais identifiants
	}

	return true, nil // Identifiants corrects
}
