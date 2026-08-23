package local

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Signer struct {
	keyID  string
	secret []byte
}

func NewSigner(keyID string, secret []byte) (*Signer, error) {
	if keyID == "" || len(secret) < 16 {
		return nil, fmt.Errorf("local signer requires key id and at least 16 secret bytes")
	}
	return &Signer{keyID: keyID, secret: append([]byte(nil), secret...)}, nil
}
func (s *Signer) Sign(ctx context.Context, digest string) (string, string, error) {
	if err := ctx.Err(); err != nil {
		return "", "", err
	}
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(digest))
	return hex.EncodeToString(mac.Sum(nil)), s.keyID, nil
}
