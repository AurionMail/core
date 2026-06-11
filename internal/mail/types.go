package mail
import(
	"time"
)
type OutgoingMessage struct {
    From      string
    To        []string
    Subject   string
    Payload   []byte // encrypted or not
    Attachments []Attachment
}

type Message struct {
    ID        string
    From      string
    To        []string
    Subject   string
    Payload   []byte // chiffré
    Snippet   string
    Date      time.Time
    Seen      bool
    Tags      []string
    Attachments []Attachment
}

type Attachment struct {
    Filename string
    MimeType string
    Data     []byte
}

type Mailbox struct {
    ID    string   `json:"id"`
    Name  string   `json:"name"`
    Role  string   `json:"role"` // inbox, sent, drafts, trash, spam, null
    Total int      `json:"total"`
    Unread int     `json:"unread"`
}

