package handlers

import (
    "time"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"

    "aurion/core/internal/db/repository"
    "aurion/core/internal/security"

	"strings"
    "encoding/base64"
    "crypto/hmac"
    "crypto/sha256"
)

type AuthHandler struct {
    Users    *repository.UserRepository
    Sessions *repository.SessionRepository
    secret   []byte
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository, secret []byte) *AuthHandler {
    return &AuthHandler{users, sessions, secret}
}

type SignupRequest struct {
    Email    string `json:"email"`
    Password string `json:"password"`
    SaltClient string `json:"salt_client"`
    SaltServer string `json:"salt_server"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
    var req SignupRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    email := strings.ToLower(req.Email)
    hash := security.HashPassword(req.Password)

    user, err := h.Users.CreateUser(c, email, hash, req.SaltServer, req.SaltClient)
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

type SaltRequest struct {
    Email string `json:"email"`
}

func (h *AuthHandler) GetSalt(c *gin.Context) {
    var req SaltRequest
    if err := c.BindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": "invalid request"})
        return
    }

    email := strings.ToLower(strings.TrimSpace(req.Email))
    start := time.Now()

    user, err := h.Users.GetUserByEmail(c, email)

    // Fake deterministic salts
    fakeServer := fakeSalt(email+"server", h.secret)
    fakeClient := fakeSalt(email+"client", h.secret)

    saltServer := fakeServer
    saltClient := fakeClient
    var userID any = nil

    if err == nil {
        saltServer = user.SaltServer
        saltClient = user.SaltClient
        userID = user.ID
    }

    // Constant-time response (50ms)
    minDuration := 50 * time.Millisecond
    elapsed := time.Since(start)
    if elapsed < minDuration {
        time.Sleep(minDuration - elapsed)
    }

    c.JSON(200, gin.H{
        "id":          userID,
        "salt_server": saltServer,
        "salt_client": saltClient,
    })
}


func fakeSalt(email string, secret []byte) string {
    mac := hmac.New(sha256.New, secret)
    mac.Write([]byte(email))
    sum := mac.Sum(nil)
    return base64.RawStdEncoding.EncodeToString(sum[:16])
}


