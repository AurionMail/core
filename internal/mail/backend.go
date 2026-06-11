package mail
import (
    "context"
)

type MailBackend interface {
    SendMessage(ctx context.Context, msg OutgoingMessage) error
	StoreSentCopy(ctx context.Context, msg OutgoingMessage) error
    ListMessages(ctx context.Context, userID string, folder string, limit int, offset int) ([]Message, error)
    GetMessage(ctx context.Context, userID string, id string) (Message, error)
    DeleteMessage(ctx context.Context, userID string, id string) error
    SetSeen(ctx context.Context, userID string, id string, seen bool) error
    UpdateTags(ctx context.Context, userID string, id string, tags []string) error
    Search(ctx context.Context, userID string, query string) ([]Message, error)
}
