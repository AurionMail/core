package handlers

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "aurion/core/internal/db/repository"
    "aurion/core/internal/security"

	"strings"
)

type AuthHandler struct {
    Users    *repository.UserRepository
    Sessions *repository.SessionRepository
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository) *AuthHandler {
    return &AuthHandler{users, sessions}
}

type SignupRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
    var req SignupRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    email := strings.ToLower(req.Email)
    hash := security.HashPassword(req.Password)

    user, err := h.Users.CreateUser(c, email, hash)
    if err != nil {
        c.JSON(400, gin.H{"error": err})
        return
    }

    token := uuid.New().String()
    expires := time.Now().Add(30 * 24 * time.Hour)

    session, err := h.Sessions.CreateSession(c, user.ID.String(), token, expires)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to create session"})
        return
    }

    c.JSON(200, gin.H{
        "user_id": user.ID,
        "email":   user.Email,
        "token":   session.Token,
    })
}

type LoginRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req LoginRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    user, err := h.Users.GetUserByEmail(c, strings.ToLower(req.Email))
    if err != nil {
        c.JSON(401, gin.H{"error": "invalid credentials"})
        return
    }

    if !security.VerifyPassword(req.Password, user.PasswordHash) {
        c.JSON(401, gin.H{"error": "invalid credentials"})
        return
    }

    token := uuid.New().String()
    expires := time.Now().Add(30 * 24 * time.Hour)

    session, err := h.Sessions.CreateSession(c, user.ID.String(), token, expires)
    if err != nil {
        c.JSON(500, gin.H{"error": "failed to create session"})
        return
    }

    c.JSON(200, gin.H{
        "user_id": user.ID,
        "email":   user.Email,
        "token":   session.Token,
    })
}

func (h *AuthHandler) Session(c *gin.Context) {
    token := extractBearer(c)
    if token == "" {
        c.JSON(401, gin.H{"error": "missing token"})
        return
    }

    session, err := h.Sessions.GetSessionByToken(c, token)
    if err != nil || session.ExpiresAt.Before(time.Now()) {
        c.JSON(401, gin.H{"error": "invalid session"})
        return
    }

    c.JSON(200, gin.H{
        "user_id": session.UserID,
        "email":   "TODO: fetch email",
    })
}

