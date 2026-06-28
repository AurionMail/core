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
	ServerAddr string // Exemple: "mail.aurion.lan:465"
}

func NewSMTPBackend(addr string) *SMTPBackend {
	return &SMTPBackend{ServerAddr: addr}
}

func (b *SMTPBackend) VerifyCredentials(ctx context.Context, email, password string) (bool, error) {
	// Sécurité : Appliquer un timeout via le contexte pour ne pas bloquer les requêtes HTTP
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 1. Connexion TLS implicite (Port 465)
	d := tls.Dialer{
		Config: &tls.Config{
			// Passer à true uniquement en développement local si Stalwart utilise un certificat auto-signé
			InsecureSkipVerify: false,
		},
	}

	conn, err := d.DialContext(dialCtx, "tcp", b.ServerAddr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	// 2. Initialisation du client SMTP Go
	client, err := smtp.NewClient(conn, b.ServerAddr)
	if err != nil {
		return false, err
	}
	defer client.Quit()

	// 3. Extraction du Hostname pour SASL Plain Auth
	host, _, err := net.SplitHostPort(b.ServerAddr)
	if err != nil {
		host = b.ServerAddr
	}

	auth := smtp.PlainAuth("", email, password, host)

	// 4. Tentative d'authentification auprès de Stalwart
	if err := client.Auth(auth); err != nil {
		// Une erreur ici signifie généralement "Authentication failed" (535)
		return false, nil
	}

	return true, nil
}
