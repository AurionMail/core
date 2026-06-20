package mail

import "context"

type MailBackend interface {
    VerifyCredentials(ctx context.Context, email, password string) (bool, error)
}