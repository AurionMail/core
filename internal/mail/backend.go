package mail

import (
    "context"
)

type MailBackend interface {
    // --- Send and store ---
    SendMessage(ctx context.Context, msg OutgoingMessage) error
    StoreSentCopy(ctx context.Context, msg OutgoingMessage) error

    // --- Messages ---
    ListMessages(ctx context.Context, userID string, folder string, limit int, offset int) ([]Message, error)
    GetMessage(ctx context.Context, userID string, id string) (Message, error)
    DeleteMessage(ctx context.Context, userID string, id string) error
    SetSeen(ctx context.Context, userID string, id string, seen bool) error
    UpdateTags(ctx context.Context, userID string, id string, tags []string) error
    Search(ctx context.Context, userID string, query string) ([]Message, error)

    // --- Mailboxes ---
    ListMailboxes(ctx context.Context, userID string) ([]Mailbox, error)
    CreateMailbox(ctx context.Context, userID string, name string) error
    RenameMailbox(ctx context.Context, userID string, id string, newName string) error
    DeleteMailbox(ctx context.Context, userID string, id string) error

    // --- Drafts ---
    CreateDraft(ctx context.Context, userID string, msg OutgoingMessage) (string, error)
    UpdateDraft(ctx context.Context, userID string, id string, msg OutgoingMessage) error
    DeleteDraft(ctx context.Context, userID string, id string) error
}
