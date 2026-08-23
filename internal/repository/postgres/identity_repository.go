package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-094/internal/domain/common"
	"github.com/wyw14/cry-094/internal/domain/identity"
)

type IdentityRepository struct{ db *Database }

func NewIdentityRepository(db *Database) *IdentityRepository { return &IdentityRepository{db: db} }
func (r *IdentityRepository) UserByEmail(ctx context.Context, email string) (*identity.User, error) {
	var value identity.User
	err := r.db.Pool.QueryRow(ctx, `SELECT id,email,password_hash,active,created_at,version FROM users WHERE lower(email)=lower($1)`, email).Scan(&value.ID, &value.Email, &value.PasswordHash, &value.Active, &value.CreatedAt, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read user: %w", err)
	}
	return &value, nil
}
func (r *IdentityRepository) SaveRefresh(ctx context.Context, value *identity.RefreshToken) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO refresh_tokens(id,user_id,digest,expires_at,created_at) VALUES($1,$2,$3,$4,$5)`, value.ID, value.UserID, value.Digest, value.ExpiresAt, value.CreatedAt)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}
func (r *IdentityRepository) RefreshByDigest(ctx context.Context, digest string) (*identity.RefreshToken, error) {
	var value identity.RefreshToken
	err := r.db.Pool.QueryRow(ctx, `SELECT id,user_id,digest,expires_at,revoked_at,created_at FROM refresh_tokens WHERE digest=$1`, digest).Scan(&value.ID, &value.UserID, &value.Digest, &value.ExpiresAt, &value.RevokedAt, &value.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read refresh token: %w", err)
	}
	return &value, nil
}
func (r *IdentityRepository) RevokeRefresh(ctx context.Context, value *identity.RefreshToken) error {
	tag, err := r.db.Pool.Exec(ctx, `UPDATE refresh_tokens SET revoked_at=$2 WHERE id=$1 AND revoked_at IS NULL`, value.ID, value.RevokedAt)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return common.ErrConflict
	}
	return nil
}
