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

