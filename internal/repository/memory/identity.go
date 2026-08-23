package memory

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/wyw14/cry-094/internal/domain/identity"
)

type IdentityRepo struct {
	mu      sync.RWMutex
	users   map[string]*identity.User
	refresh map[string]*identity.RefreshToken
}

func NewIdentityRepo(users ...*identity.User) *IdentityRepo {
	repo := &IdentityRepo{users: make(map[string]*identity.User), refresh: make(map[string]*identity.RefreshToken)}
	for _, user := range users {
		copy := *user
		copy.PasswordHash = append([]byte(nil), user.PasswordHash...)
		repo.users[strings.ToLower(user.Email)] = &copy
	}
	return repo
}
func (r *IdentityRepo) UserByEmail(_ context.Context, email string) (*identity.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.users[strings.ToLower(email)]
	if !ok {
		return nil, fmt.Errorf("user not found")
	}
	copy := *value
	copy.PasswordHash = append([]byte(nil), value.PasswordHash...)
	return &copy, nil
}
func (r *IdentityRepo) SaveRefresh(_ context.Context, value *identity.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.refresh[value.Digest]; ok {
		return fmt.Errorf("refresh token exists")
	}
	copy := *value
	r.refresh[value.Digest] = &copy
	return nil
}
func (r *IdentityRepo) RefreshByDigest(_ context.Context, digest string) (*identity.RefreshToken, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.refresh[digest]
	if !ok {
		return nil, fmt.Errorf("refresh token not found")
	}
	copy := *value
	return &copy, nil
}
func (r *IdentityRepo) RevokeRefresh(_ context.Context, value *identity.RefreshToken) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	stored, ok := r.refresh[value.Digest]
	if !ok {
		return fmt.Errorf("refresh token not found")
	}
	if stored.RevokedAt != nil {
		return nil
	}
	if value.RevokedAt == nil {
		return fmt.Errorf("refresh token revocation timestamp required")
	}
	copy := *stored
	revokedAt := *value.RevokedAt
	copy.RevokedAt = &revokedAt
	r.refresh[value.Digest] = &copy
	return nil
}
