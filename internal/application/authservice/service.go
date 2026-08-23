package authservice

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/identity"
)

type Repository interface {
	UserByEmail(context.Context, string) (*identity.User, error)
	SaveRefresh(context.Context, *identity.RefreshToken) error
	RefreshByDigest(context.Context, string) (*identity.RefreshToken, error)
	RevokeRefresh(context.Context, *identity.RefreshToken) error
}
type IDGenerator interface{ NewID() string }
type Tokens struct {
	Access    string `json:"access_token"`
	Refresh   string `json:"refresh_token"`
	ExpiresIn int64  `json:"expires_in"`
}
type Service struct {
	repo       Repository
	clock      ports.Clock
	ids        IDGenerator
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func New(repo Repository, clock ports.Clock, ids IDGenerator, secret []byte, accessTTL, refreshTTL time.Duration) *Service {
	return &Service{repo: repo, clock: clock, ids: ids, secret: append([]byte(nil), secret...), accessTTL: accessTTL, refreshTTL: refreshTTL}
}
func (s *Service) Login(ctx context.Context, email, password string) (Tokens, error) {
	user, err := s.repo.UserByEmail(ctx, email)
	if err != nil || !user.Active || !user.VerifyPassword(password) {
		return Tokens{}, fmt.Errorf("invalid credentials")
	}
	return s.issue(ctx, user.ID)
}
func (s *Service) Refresh(ctx context.Context, plain string) (Tokens, error) {
	digest := identity.TokenDigest(plain)
	record, err := s.repo.RefreshByDigest(ctx, digest)
	if err != nil {
		return Tokens{}, fmt.Errorf("refresh token lookup failed: %w", err)
	}
	now := s.clock.Now()
	if !record.Usable(plain, now) {
		return Tokens{}, fmt.Errorf("refresh token is not usable")
	}
	// Atomically claim the refresh token by revoking it before minting the
	// successor. RevokeRefresh is conditional on revoked_at being NULL, so of two
	// concurrent refreshes for the same token only one can win the claim; the
	// loser gets ErrConflict and must not mint a second set of credentials.
	// Revoking before issuing is intentionally fail-closed: if issuing fails the
	// token is consumed and the caller must re-authenticate rather than risk a
	// double-spend window where two usable successors exist.
	record.Revoke(now)
	if err := s.repo.RevokeRefresh(ctx, record); err != nil {
		return Tokens{}, err
	}
	rotated, err := s.issue(ctx, record.UserID)
	if err != nil {
		return Tokens{}, fmt.Errorf("issue rotated credentials: %w", err)
	}
	return rotated, nil
}
func (s *Service) issue(ctx context.Context, userID string) (Tokens, error) {
	now := s.clock.Now()
	claims := jwt.MapClaims{"sub": userID, "iat": now.Unix(), "exp": now.Add(s.accessTTL).Unix(), "jti": s.ids.NewID()}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return Tokens{}, err
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return Tokens{}, err
	}
	refresh := base64.RawURLEncoding.EncodeToString(raw)
	record := identity.NewRefreshToken(s.ids.NewID(), userID, refresh, now.Add(s.refreshTTL), now)
	if err := s.repo.SaveRefresh(ctx, &record); err != nil {
		return Tokens{}, err
	}
	return Tokens{Access: access, Refresh: refresh, ExpiresIn: int64(s.accessTTL.Seconds())}, nil
}
