package jmap

import (
    "context"
    "aurion/core/internal/mail"
	"fmt"
)

type JMAPBackend struct {
    client *JMAPClient
}

func NewJMAPBackend(url, username, password string) *JMAPBackend {
    return &JMAPBackend{
        client: NewJMAPClient(url, username, password),
    }
}

func (b *JMAPBackend) SendMessage(ctx context.Context, msg mail.OutgoingMessage) error {

    emailCreate := map[string]interface{}{
        "mailboxIds": map[string]bool{
            "outbox": true,
        },
        "subject": msg.Subject,
        "bodyValues": map[string]interface{}{
            "body": map[string]string{
                "value": string(msg.Payload),
            },
        },
        "textBody": []map[string]string{
            {"partId": "body"},
        },
        "from": []map[string]string{
            {"email": msg.From},
        },
        "to": buildEmailers(msg.To),
    }

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            // Étape 1 : Email/set
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": msg.From,
                    "create": map[string]interface{}{
                        "email1": emailCreate,
                    },
                },
                "c1",
            },
            // Étape 2 : EmailSubmission/set
            []interface{}{
                "EmailSubmission/set",
                map[string]interface{}{
                    "accountId": msg.From,
                    "create": map[string]interface{}{
                        "sub1": map[string]interface{}{
                            "emailId": "#email1",
                        },
                    },
                },
                "c2",
            },
        },
    }

    // 3. call jmap server
    _, err := b.client.Call(ctx, req)
    if err != nil {
        return fmt.Errorf("jmap send failed: %w", err)
    }

    return nil
}


func (b *JMAPBackend) StoreSentCopy(ctx context.Context, msg mail.OutgoingMessage) error {

    emailCreate := map[string]interface{}{
        "mailboxIds": map[string]bool{
            "sent": true,
        },
        "subject": msg.Subject,
        "bodyValues": map[string]interface{}{
            "body": map[string]string{
                "value": string(msg.Payload), // ciphertext
            },
        },
        "textBody": []map[string]string{
            {"partId": "body"},
        },
        "from": []map[string]string{
            {"email": msg.From},
        },
        "to": buildEmailers(msg.To),
    }

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": msg.From,
                    "create": map[string]interface{}{
                        "email1": emailCreate,
                    },
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    if err != nil {
        return fmt.Errorf("jmap store sent copy failed: %w", err)
    }

    return nil
}


func (b *JMAPBackend) ListMessages(ctx context.Context, userID string, folder string, limit int, offset int) ([]mail.Message, error) {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/query",
                map[string]interface{}{
                    "accountId": userID,
                    "filter": map[string]interface{}{
                        "inMailbox": folder,
                    },
                    "limit":    limit,
                    "position": offset,
                },
                "c1",
            },
            []interface{}{
                "Email/get",
                map[string]interface{}{
                    "accountId": userID,
                    "ids":       "#c1",
                },
                "c2",
            },
        },
    }

    resp, err := b.client.Call(ctx, req)
    if err != nil {
        return nil, err
    }

    // extract messages
    methodResponses := resp["methodResponses"].([]interface{})
    getResp := methodResponses[1].([]interface{})[1].(map[string]interface{})
    list := getResp["list"].([]interface{})

    messages := make([]mail.Message, 0, len(list))

    for _, raw := range list {
        m := raw.(map[string]interface{})

        messages = append(messages, mail.Message{
            ID:      m["id"].(string),
            Subject: m["subject"].(string),
            From:    extractEmail(m["from"]),
            To:      extractEmails(m["to"]),
            Payload: []byte(extractBody(m)),
        })
    }

    return messages, nil
}


func (b *JMAPBackend) GetMessage(ctx context.Context, userID string, id string) (mail.Message, error) {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/get",
                map[string]interface{}{
                    "accountId": userID,
                    "ids":       []string{id},
                },
                "c1",
            },
        },
    }

    resp, err := b.client.Call(ctx, req)
    if err != nil {
        return mail.Message{}, err
    }

    methodResponses := resp["methodResponses"].([]interface{})
    getResp := methodResponses[0].([]interface{})[1].(map[string]interface{})
    list := getResp["list"].([]interface{})

    if len(list) == 0 {
        return mail.Message{}, fmt.Errorf("not found")
    }

    m := list[0].(map[string]interface{})

    return mail.Message{
        ID:      id,
        Subject: m["subject"].(string),
        From:    extractEmail(m["from"]),
        To:      extractEmails(m["to"]),
        Payload: []byte(extractBody(m)),
    }, nil
}


func (b *JMAPBackend) DeleteMessage(ctx context.Context, userID string, id string) error {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": userID,
                    "destroy":   []string{id},
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}


func (b *JMAPBackend) SetSeen(ctx context.Context, userID string, id string, seen bool) error {

    keywords := map[string]bool{}
    if seen {
        keywords["$seen"] = true
    }

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": userID,
                    "update": map[string]interface{}{
                        id: map[string]interface{}{
                            "keywords": keywords,
                        },
                    },
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}


func (b *JMAPBackend) UpdateTags(ctx context.Context, userID string, id string, tags []string) error {

    keywords := map[string]bool{}
    for _, t := range tags {
        keywords[t] = true
    }

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": userID,
                    "update": map[string]interface{}{
                        id: map[string]interface{}{
                            "keywords": keywords,
                        },
                    },
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}


func (b *JMAPBackend) Search(ctx context.Context, userID string, query string) ([]mail.Message, error) {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            // 1. Email/query
            []interface{}{
                "Email/query",
                map[string]interface{}{
                    "accountId": userID,
                    "filter": map[string]interface{}{
                        "text": query,
                    },
                    "limit": 50,
                },
                "c1",
            },
            // 2. Email/get
            []interface{}{
                "Email/get",
                map[string]interface{}{
                    "accountId": userID,
                    "ids":       "#c1",
                },
                "c2",
            },
        },
    }

    resp, err := b.client.Call(ctx, req)
    if err != nil {
        return nil, err
    }

    // extract messages
    methodResponses := resp["methodResponses"].([]interface{})
    getResp := methodResponses[1].([]interface{})[1].(map[string]interface{})
    list := getResp["list"].([]interface{})

    messages := make([]mail.Message, 0, len(list))

    for _, raw := range list {
        m := raw.(map[string]interface{})

        messages = append(messages, mail.Message{
            ID:      m["id"].(string),
            Subject: m["subject"].(string),
            From:    extractEmail(m["from"]),
            To:      extractEmails(m["to"]),
            Payload: []byte(extractBody(m)),
        })
    }

    return messages, nil
}

func (b *JMAPBackend) ListMailboxes(ctx context.Context, userID string) ([]mail.Mailbox, error) {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Mailbox/get",
                map[string]interface{}{
                    "accountId": userID,
                },
                "c1",
            },
        },
    }

    resp, err := b.client.Call(ctx, req)
    if err != nil {
        return nil, err
    }

    methodResponses := resp["methodResponses"].([]interface{})
    getResp := methodResponses[0].([]interface{})[1].(map[string]interface{})
    list := getResp["list"].([]interface{})

    out := make([]mail.Mailbox, 0, len(list))

    for _, raw := range list {
        m := raw.(map[string]interface{})

        out = append(out, mail.Mailbox{
            ID:     m["id"].(string),
            Name:   m["name"].(string),
            Role:   getString(m["role"]),
            Total:  int(m["totalEmails"].(float64)),
            Unread: int(m["unreadEmails"].(float64)),
        })
    }

    return out, nil
}

func (b *JMAPBackend) CreateMailbox(ctx context.Context, userID, name string) error {
    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Mailbox/set",
                map[string]interface{}{
                    "accountId": userID,
                    "create": map[string]interface{}{
                        "new": map[string]interface{}{
                            "name": name,
                        },
                    },
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}

func (b *JMAPBackend) RenameMailbox(ctx context.Context, userID, id, newName string) error {
    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Mailbox/set",
                map[string]interface{}{
                    "accountId": userID,
                    "update": map[string]interface{}{
                        id: map[string]interface{}{
                            "name": newName,
                        },
                    },
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}

func (b *JMAPBackend) DeleteMailbox(ctx context.Context, userID, id string) error {
    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Mailbox/set",
                map[string]interface{}{
                    "accountId": userID,
                    "destroy":   []string{id},
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}

func (b *JMAPBackend) CreateDraft(ctx context.Context, userID string, msg mail.OutgoingMessage) (string, error) {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": userID,
                    "create": map[string]interface{}{
                        "draft": map[string]interface{}{
                            "mailboxIds": map[string]bool{
                                "drafts": true,
                            },
                            "subject": msg.Subject,
                            "to":      buildEmailers(msg.To),
                            "bodyValues": map[string]interface{}{
                                "body": map[string]interface{}{
                                    "value": string(msg.Payload),
                                },
                            },
                        },
                    },
                },
                "c1",
            },
        },
    }

    resp, err := b.client.Call(ctx, req)
    if err != nil {
        return "", err
    }

    methodResponses := resp["methodResponses"].([]interface{})
    setResp := methodResponses[0].([]interface{})[1].(map[string]interface{})
    created := setResp["created"].(map[string]interface{})
    draft := created["draft"].(map[string]interface{})

    return draft["id"].(string), nil
}

func (b *JMAPBackend) UpdateDraft(ctx context.Context, userID, id string, msg mail.OutgoingMessage) error {

    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": userID,
                    "update": map[string]interface{}{
                        id: map[string]interface{}{
                            "subject": msg.Subject,
                            "to":      buildEmailers(msg.To),
                            "bodyValues": map[string]interface{}{
                                "body": map[string]interface{}{
                                    "value": string(msg.Payload),
                                },
                            },
                        },
                    },
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}

func (b *JMAPBackend) DeleteDraft(ctx context.Context, userID, id string) error {
    req := map[string]interface{}{
        "using": []string{
            "urn:ietf:params:jmap:core",
            "urn:ietf:params:jmap:mail",
        },
        "methodCalls": []interface{}{
            []interface{}{
                "Email/set",
                map[string]interface{}{
                    "accountId": userID,
                    "destroy":   []string{id},
                },
                "c1",
            },
        },
    }

    _, err := b.client.Call(ctx, req)
    return err
}
