package handlers

import (
    "github.com/gin-gonic/gin"
    "aurion/core/internal/mail"
)

type MailHandler struct {
    service *mail.MailService
}

func NewMailHandler(service *mail.MailService) *MailHandler {
    return &MailHandler{service}
}

type SendMailRequest struct {
    To                   string        `json:"to"`
    Subject              string        `json:"subject"`
    CiphertextForSender  []byte        `json:"ciphertext_for_sender"`
    CiphertextForReceiver []byte       `json:"ciphertext_for_receiver"`
    Attachments          []mail.Attachment `json:"attachments"`
}

func (h *MailHandler) SendMail(c *gin.Context) {
    var req SendMailRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    userID := c.GetString("user_id")

    err := h.service.SendEncrypted(
        c,
        userID,
        req.To,
        req.Subject,
        req.CiphertextForSender,
        req.CiphertextForReceiver,
        req.Attachments,
    )

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "sent"})
}


func (h *MailHandler) ListMessages(c *gin.Context) {
    userID := c.GetString("user_id")
    folder := c.Query("folder")
    limit := parseIntOrDefault(c.Query("limit"), 50)
    offset := parseIntOrDefault(c.Query("offset"), 0)

    msgs, err := h.service.ListMessages(c, userID, folder, limit, offset)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, msgs)
}

func (h *MailHandler) GetMessage(c *gin.Context) {
    userID := c.GetString("user_id")
    id := c.Param("id")

    msg, err := h.service.GetMessage(c, userID, id)
    if err != nil {
        c.JSON(404, gin.H{"error": "not found"})
        return
    }

    c.JSON(200, msg)
}

func (h *MailHandler) DeleteMessage(c *gin.Context) {
    userID := c.GetString("user_id")
    id := c.Param("id")

    if err := h.service.DeleteMessage(c, userID, id); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "deleted"})
}

func (h *MailHandler) SetSeen(c *gin.Context) {
    userID := c.GetString("user_id")
    id := c.Param("id")

    var req struct {
        Seen bool `json:"seen"`
    }
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if err := h.service.SetSeen(c, userID, id, req.Seen); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "updated"})
}

func (h *MailHandler) UpdateTags(c *gin.Context) {
    userID := c.GetString("user_id")
    id := c.Param("id")

    var req struct {
        Tags []string `json:"tags"`
    }
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if err := h.service.UpdateTags(c, userID, id, req.Tags); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "updated"})
}

func (h *MailHandler) Search(c *gin.Context) {
    userID := c.GetString("user_id")
    q := c.Query("q")

    msgs, err := h.service.Search(c, userID, q)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, msgs)
}

func (h *MailHandler) ListMailboxes(c *gin.Context) {
    userID := c.GetString("user_id")

    boxes, err := h.service.ListMailboxes(c, userID)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, boxes)
}

func (h *MailHandler) CreateMailbox(c *gin.Context) {
    userID := c.GetString("user_id")

    var req struct {
        Name string `json:"name"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if err := h.service.CreateMailbox(c, userID, req.Name); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "created"})
}

func (h *MailHandler) RenameMailbox(c *gin.Context) {
    userID := c.GetString("user_id")

    var req struct {
        ID   string `json:"id"`
        Name string `json:"name"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if err := h.service.RenameMailbox(c, userID, req.ID, req.Name); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "renamed"})
}

func (h *MailHandler) DeleteMailbox(c *gin.Context) {
    userID := c.GetString("user_id")

    var req struct {
        ID string `json:"id"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    if err := h.service.DeleteMailbox(c, userID, req.ID); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "deleted"})
}

func (h *MailHandler) CreateDraft(c *gin.Context) {
    userID := c.GetString("user_id")

    var req struct {
        To      []string `json:"to"`
        Subject string   `json:"subject"`
        Payload []byte   `json:"payload"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    id, err := h.service.CreateDraft(c, userID, mail.OutgoingMessage{
        From:    userID,
        To:      req.To,
        Subject: req.Subject,
        Payload: req.Payload,
    })

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"id": id})
}

func (h *MailHandler) UpdateDraft(c *gin.Context) {
    userID := c.GetString("user_id")
    id := c.Param("id")

    var req struct {
        To      []string `json:"to"`
        Subject string   `json:"subject"`
        Payload []byte   `json:"payload"`
    }

    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    err := h.service.UpdateDraft(c, userID, id, mail.OutgoingMessage{
        From:    userID,
        To:      req.To,
        Subject: req.Subject,
        Payload: req.Payload,
    })

    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "updated"})
}

func (h *MailHandler) DeleteDraft(c *gin.Context) {
    userID := c.GetString("user_id")
    id := c.Param("id")

    if err := h.service.DeleteDraft(c, userID, id); err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, gin.H{"status": "deleted"})
}
