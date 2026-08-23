package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-094/internal/domain/common"
	"github.com/wyw14/cry-094/internal/domain/team"
)

type TeamRepository struct{ db *Database }

func NewTeamRepository(db *Database) *TeamRepository { return &TeamRepository{db: db} }
func (r *TeamRepository) Create(ctx context.Context, value *team.Team) error {
	members, err := json.Marshal(value.Members)
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, `INSERT INTO teams(id,name,members,version,created_at) VALUES($1,$2,$3,$4,$5)`, value.ID, value.Name, members, value.Version, value.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert team: %w", err)
	}
	return nil
}
func (r *TeamRepository) Get(ctx context.Context, id string) (*team.Team, error) {
	var value team.Team
	var members []byte
	err := r.db.Pool.QueryRow(ctx, `SELECT id,name,members,version,created_at FROM teams WHERE id=$1`, id).Scan(&value.ID, &value.Name, &members, &value.Version, &value.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read team: %w", err)
	}
	if err := json.Unmarshal(members, &value.Members); err != nil {
		return nil, fmt.Errorf("decode team members: %w", err)
	}
	return &value, nil
}
func (r *TeamRepository) Save(ctx context.Context, value *team.Team, expectedVersion int64) error {
	members, err := json.Marshal(value.Members)
	if err != nil {
		return err
	}
	tag, err := r.db.Pool.Exec(ctx, `UPDATE teams SET name=$2,members=$3,version=$4 WHERE id=$1 AND version=$5`, value.ID, value.Name, members, value.Version, expectedVersion)
	if err != nil {
		return fmt.Errorf("update team: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return common.ErrVersionConflict
	}
	return nil
}
