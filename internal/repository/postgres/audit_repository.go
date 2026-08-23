package postgres

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-094/internal/domain/audit"
)

type AuditRepository struct{ db *Database }

func NewAuditRepository(db *Database) *AuditRepository { return &AuditRepository{db: db} }
func (r *AuditRepository) Append(ctx context.Context, value audit.Event) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO audit_events(id,team_id,actor_id,source,action,resource_type,resource_id,before_data,after_data,reason,occurred_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`, value.ID, value.TeamID, value.ActorID, value.Source, value.Action, value.ResourceType, value.ResourceID, value.Before, value.After, value.Reason, value.OccurredAt)
	if err != nil {
		return fmt.Errorf("append immutable audit event: %w", err)
	}
	return nil
}
