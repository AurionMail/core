package jmap
import(
	"time"
	"aurion/core/internal/mail"
)
func extractEmail(raw interface{}) string {
    arr, ok := raw.([]interface{})
    if !ok || len(arr) == 0 {
        return ""
    }
    m, ok := arr[0].(map[string]interface{})
    if !ok {
        return ""
    }
    email, _ := m["email"].(string)
    return email
}

func extractEmails(raw interface{}) []string {
    arr, ok := raw.([]interface{})
    if !ok {
        return []string{}
    }

    emails := make([]string, 0, len(arr))
    for _, item := range arr {
        m, ok := item.(map[string]interface{})
        if !ok {
            continue
        }
        if email, ok := m["email"].(string); ok {
            emails = append(emails, email)
        }
    }
    return emails
}



func extractBody(m map[string]interface{}) string {
    bodyValues, ok := m["bodyValues"].(map[string]interface{})
    if !ok {
        return ""
    }

    body, ok := bodyValues["body"].(map[string]interface{})
    if !ok {
        return ""
    }

    value, _ := body["value"].(string)
    return value
}

func extractSnippet(m map[string]interface{}) string {
    if preview, ok := m["preview"].(string); ok {
        return preview
    }
    return ""
}

func extractDate(m map[string]interface{}) time.Time {
    if s, ok := m["receivedAt"].(string); ok {
        t, err := time.Parse(time.RFC3339, s)
        if err == nil {
            return t
        }
    }
    return time.Time{}
}

func extractSeen(m map[string]interface{}) bool {
    kw, ok := m["keywords"].(map[string]interface{})
    if !ok {
        return false
    }
    if v, ok := kw["$seen"].(bool); ok {
        return v
    }
    return false
}

func extractTags(m map[string]interface{}) []string {
    kw, ok := m["keywords"].(map[string]interface{})
    if !ok {
        return []string{}
    }

    tags := []string{}
    for k, v := range kw {
        if k == "$seen" {
            continue
        }
        if b, ok := v.(bool); ok && b {
            tags = append(tags, k)
        }
    }
    return tags
}

func extractAttachments(m map[string]interface{}) []mail.Attachment {
    arr, ok := m["attachments"].([]interface{})
    if !ok {
        return nil
    }

    out := make([]mail.Attachment, 0, len(arr))

    for _, item := range arr {
        a, ok := item.(map[string]interface{})
        if !ok {
            continue
        }


        out = append(out, mail.Attachment{
            Filename: a["name"].(string),
            MimeType: a["type"].(string),
            Data:     nil, 
        })
    }

    return out
}

