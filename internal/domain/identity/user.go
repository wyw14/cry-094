package identity

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash []byte    `json:"-"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	Version      int64     `json:"version"`
}

func Register(id, email, password string, now time.Time) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return &User{ID: id, Email: email, PasswordHash: hash, Active: true, CreatedAt: now.UTC(), Version: 1}, nil
}
func (u User) VerifyPassword(password string) bool {
	return bcrypt.CompareHashAndPassword(u.PasswordHash, []byte(password)) == nil
}

type RefreshToken struct {
	ID        string
	UserID    string
	Digest    string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func NewRefreshToken(id, userID, plain string, expires, now time.Time) RefreshToken {
	return RefreshToken{ID: id, UserID: userID, Digest: TokenDigest(plain), ExpiresAt: expires.UTC(), CreatedAt: now.UTC()}
}
func (t RefreshToken) Usable(plain string, now time.Time) bool {
	return t.RevokedAt == nil && now.Before(t.ExpiresAt) && t.Digest == TokenDigest(plain)
}
func (t *RefreshToken) Revoke(now time.Time) { value := now.UTC(); t.RevokedAt = &value }
func TokenDigest(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
