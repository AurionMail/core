package mail

import (
	"context"
	"crypto/tls"
	"net/textproto"
	"time"
)

type IMAPBackend struct {
	ServerAddr string
}

func NewIMAPBackend(addr string) *IMAPBackend {
	return &IMAPBackend{ServerAddr: addr}
}

func (b *IMAPBackend) VerifyCredentials(ctx context.Context, email, password string) (bool, error) {
	d := tls.Dialer{
		Config: &tls.Config{InsecureSkipVerify: false},
	}
	
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := d.DialContext(dialCtx, "tcp", b.ServerAddr)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	client := textproto.NewConn(conn)
	defer client.Close()

	// Lire la bannière d'accueil du serveur IMAP
	_, _, err = client.ReadResponse(0)
	if err != nil {
		return false, err
	}

	// Envoyer la commande de login IMAP standard A001 LOGIN email password
	id, err := client.Cmd("A001 LOGIN %s %s", email, password)
	if err != nil {
		return false, err
	}

	client.StartResponse(id)
	defer client.EndResponse(id)

	_, _, err = client.ReadResponse(0)
	if err != nil {
		// Une erreur de réponse signifie généralement "Invalid credentials"
		return false, nil
	}

	return true, nil
}