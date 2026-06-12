package mail

import (
    "context"
    "fmt"
    "aurion/core/internal/wkd"
    "aurion/core/internal/db/repository"
)

type MailService struct {
    backend     MailBackend
    identities  *repository.IdentityRepository
    publicKeys  *repository.IdentityPublicKeyRepository
    privateKeys *repository.IdentityPrivateKeyRepository
}

func NewMailService(
    backend MailBackend,
    identities *repository.IdentityRepository,
    pub *repository.IdentityPublicKeyRepository,
    priv *repository.IdentityPrivateKeyRepository,
) *MailService {
    return &MailService{
        backend:     backend,
        identities:  identities,
        publicKeys:  pub,
        privateKeys: priv,
    }
}

func (s *MailService) SendEncrypted(
    ctx context.Context,
    fromUserID string,
    toEmail string,
    subject string,
    ciphertextForSender []byte,
    ciphertextForReceiver []byte,
    attachments []Attachment,
) error {

    if len(ciphertextForSender) == 0 {
        return fmt.Errorf("missing ciphertext for sender copy")
    }

    //
    // 1. check for key with WKD
    //
    hasReceiverKey := false

    extKey, err := wkd.LookupWKD(toEmail)
    if err == nil && extKey != "" {
        hasReceiverKey = true
    }

    //
    // 2. if key
    //
    if hasReceiverKey && len(ciphertextForReceiver) == 0 {
        return fmt.Errorf("missing ciphertext for receiver while receiver has a public key")
    }

    //
    // 3. constrcut message
    //
    outgoing := OutgoingMessage{
        From:        fromUserID,
        To:          []string{toEmail},
        Subject:     subject,
        Payload:     ciphertextForReceiver, // chiffré ou plaintext fallback
        Attachments: attachments,
    }

    //
    // 4. send via backend
    //
    if err := s.backend.SendMessage(ctx, outgoing); err != nil {
        return err
    }

    //
    // 5. store encrypted copy
    //
    if err := s.backend.StoreSentCopy(ctx, OutgoingMessage{
        From:        fromUserID,
        To:          []string{toEmail},
        Subject:     subject,
        Payload:     ciphertextForSender,
        Attachments: attachments,
    }); err != nil {
        return fmt.Errorf("failed to store sent copy: %w", err)
    }

    return nil
}

func (s *MailService) GetMessage(ctx context.Context, userID string, id string) (Message, error) {
    return s.backend.GetMessage(ctx, userID, id)
}

func (s *MailService) ListMessages(ctx context.Context, userID string, folder string, limit int, offset int) ([]Message, error) {
    return s.backend.ListMessages(ctx, userID, folder, limit, offset)
}

func (s *MailService) SetSeen(ctx context.Context, userID, id string, seen bool) error {
    return s.backend.SetSeen(ctx, userID, id, seen)
}

func (s *MailService) UpdateTags(ctx context.Context, userID, id string, tags []string) error {
    return s.backend.UpdateTags(ctx, userID, id, tags)
}

func (s *MailService) DeleteMessage(ctx context.Context, userID, id string) error {
    return s.backend.DeleteMessage(ctx, userID, id)
}

func (s *MailService) Search(ctx context.Context, userID string, query string) ([]Message, error) {
    return s.backend.Search(ctx, userID, query)
}

func (s *MailService) ListMailboxes(ctx context.Context, userID string) ([]Mailbox, error) {
    return s.backend.ListMailboxes(ctx, userID)
}

func (s *MailService) CreateMailbox(ctx context.Context, userID, name string) error {
    return s.backend.CreateMailbox(ctx, userID, name)
}

func (s *MailService) RenameMailbox(ctx context.Context, userID, id, name string) error {
    return s.backend.RenameMailbox(ctx, userID, id, name)
}

func (s *MailService) DeleteMailbox(ctx context.Context, userID, id string) error {
    return s.backend.DeleteMailbox(ctx, userID, id)
}

func (s *MailService) CreateDraft(ctx context.Context, userID string, msg OutgoingMessage) (string, error) {
    return s.backend.CreateDraft(ctx, userID, msg)
}

func (s *MailService) UpdateDraft(ctx context.Context, userID, id string, msg OutgoingMessage) error {
    return s.backend.UpdateDraft(ctx, userID, id, msg)
}

func (s *MailService) DeleteDraft(ctx context.Context, userID, id string) error {
    return s.backend.DeleteDraft(ctx, userID, id)
}
