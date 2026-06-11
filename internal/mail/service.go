package mail

import (
    "context"
    "aurion/core/internal/db/repository"
	
	"fmt"
)

type MailService struct {
    backend     MailBackend
    publicKeys  *repository.PublicKeyRepository
    privateKeys *repository.PrivateKeyRepository
}

func NewMailService(
    backend MailBackend,
    pub *repository.PublicKeyRepository,
    priv *repository.PrivateKeyRepository,
) *MailService {
    return &MailService{backend, pub, priv}
}

func (s *MailService) SendEncrypted(
    ctx context.Context,
    fromUserID string,
    toEmail string,
    subject string,
    ciphertextForSender []byte,    // always encrypted with sender key
    ciphertextForReceiver []byte,  // encrypted for receiver OR plaintext fallback
    attachments []Attachment,
) error {

    if len(ciphertextForSender) == 0 {
        return fmt.Errorf("missing ciphertext for sender copy")
    }

    //
    // check if reeceiver has pub key
    //
    _, err := s.publicKeys.GetPrimaryPublicKeyByEmail(ctx, toEmail)
    hasReceiverKey := (err == nil)


    if hasReceiverKey && len(ciphertextForReceiver) == 0 {
        return fmt.Errorf("missing ciphertext for receiver while receiver has a public key")
    }

    outgoing := OutgoingMessage{
        From:        fromUserID,
        To:          toEmail,
        Subject:     subject,
        Payload:     ciphertextForReceiver, // encrypted OR plaintext fallback
        Attachments: attachments,
    }

    //
    // 5. send with backend
    //
    if err := s.backend.SendMessage(ctx, outgoing); err != nil {
        return err
    }

    //
    // 6. store sender's copy
    //
    if err := s.backend.StoreSentCopy(ctx, OutgoingMessage{
        From:        fromUserID,
        To:          toEmail,
        Subject:     subject,
        Payload:     ciphertextForSender,
        Attachments: attachments,
    }); err != nil {
        return fmt.Errorf("failed to store sent copy: %w", err)
    }

    return nil
}


func (s *MailService) GetMessage(
    ctx context.Context,
    userID string,
    id string,
) (Message, error) {

    msg, err := s.backend.GetMessage(ctx, userID, id)
    if err != nil {
        return Message{}, err
    }

    
    return msg, nil
}

func (s *MailService) ListMessages(
    ctx context.Context,
    userID string,
    folder string,
    limit int,
    offset int,
) ([]Message, error) {
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
