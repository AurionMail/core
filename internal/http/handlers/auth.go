package handlers

import (
	"database/sql"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"aurion/core/internal/db/repository"
	"aurion/core/internal/mail"
	"aurion/core/internal/security"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type AuthHandler struct {
	Users       *repository.UserRepository
	Sessions    *repository.SessionRepository
	MailBackend mail.MailBackend
	secret      []byte
}

func NewAuthHandler(users *repository.UserRepository, sessions *repository.SessionRepository, mailBackend mail.MailBackend, secret []byte) *AuthHandler {
	return &AuthHandler{users, sessions, mailBackend, secret}
}

type SignupRequest struct {
	Email                   string `json:"email"`
	Password                string `json:"password"`
	ServerPassword          string `json:"server_password"`
	EncryptedServerPassword string `json:"encrypted_server_password"`
	SaltClient              string `json:"salt_client"`
	SaltServer              string `json:"salt_server"`
}

func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	email := strings.ToLower(req.Email)

	if h.MailBackend != nil {
		valid, err := h.MailBackend.VerifyCredentials(c.Request.Context(), email, req.ServerPassword)
		if err != nil || !valid {
			c.JSON(401, gin.H{"error": "external mail server authentication failed"})
			return
		}
	}

	hash := security.HashPassword(req.Password)

	// Instead of creating a new user, fetch existing user by email and update
	user, err := h.Users.GetUserByEmail(c, email)
	if err != nil {
		c.JSON(400, gin.H{"error": "user not found"})
		return
	}

	// Update password hash and salts
	user.PasswordHash = hash
	user.SaltServer = req.SaltServer
	user.SaltClient = req.SaltClient
	if req.EncryptedServerPassword != "" {
		user.ServerPasswordEncrypted = sql.NullString{
			String: req.EncryptedServerPassword,
			Valid:  true,
		}
	}

	if _, err := h.Users.UpdateUserByEmail(c, user); err != nil {
		c.JSON(500, gin.H{"error": "failed to update user"})
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

type VerifyCredentialsRequest struct {
	Email          string `json:"email"`
	ServerPassword string `json:"server_password"`
}

// VerifyCredentials permet à l'UI de tester si les identifiants IMAP/JMAP sont bons en amont.
func (h *AuthHandler) VerifyCredentials(c *gin.Context) {
	var req VerifyCredentialsRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	if h.MailBackend == nil {
		c.JSON(501, gin.H{"error": "mail backend integration not configured"})
		return
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	valid, err := h.MailBackend.VerifyCredentials(c.Request.Context(), email, req.ServerPassword)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to contact external mail server"})
		return
	}

	if !valid {
		c.JSON(401, gin.H{"error": "external mail server authentication failed"})
		return
	}

	c.JSON(200, gin.H{"status": "valid"})
}

func fakeSalt(email string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(email))
	sum := mac.Sum(nil)
	return base64.RawStdEncoding.EncodeToString(sum[:16])
}

// GetServerLogin extrait le token, valide la session et retourne le mot de passe serveur chiffré de l'utilisateur.
func (h *AuthHandler) GetServerLogin(c *gin.Context) {
	// 1. Extraction et validation du token de session
	token := extractBearer(c)
	if token == "" {
		c.JSON(401, gin.H{"error": "missing token"})
		return
	}

	session, err := h.Sessions.GetSessionByToken(c, token)
	if err != nil || session.ExpiresAt.Before(time.Now()) {
		c.JSON(401, gin.H{"error": "invalid or expired session"})
		return
	}

	// 2. Récupération de l'utilisateur associé à la session
	// Note : Selon l'implémentation de ton Repository, GetUserById peut prendre un string ou un uuid.UUID.
	user, err := h.Users.GetUserById(c, session.UserID)
	if err != nil {
		c.JSON(404, gin.H{"error": "user not found"})
		return
	}

	// 3. Retour de la propriété attendue
	// Adapte 'ServerPasswordEncrypted' selon le nom exact du champ dans ton struct User (ex: user.ServerPasswordEncrypted)
	c.JSON(200, gin.H{
		"server_password_encrypted": user.ServerPasswordEncrypted.String,
	})
}
