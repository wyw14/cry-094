package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-094/internal/domain/common"
	"github.com/wyw14/cry-094/internal/domain/precheck"
)

type PlanRepository struct{ db *Database }

func NewPlanRepository(db *Database) *PlanRepository { return &PlanRepository{db: db} }
func (r *PlanRepository) Save(ctx context.Context, value *precheck.Plan) error {
	steps, err := json.Marshal(value.Steps)
	if err != nil {
		return err
	}
	tag, err := r.db.Pool.Exec(ctx, `INSERT INTO precheck_plans(id,analysis_id,team_id,state,steps,created_by,reviewed_by,review_reason,content_digest,signature,signer_key_id,created_at,reviewed_at,signed_at,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) ON CONFLICT(id) DO UPDATE SET state=EXCLUDED.state,steps=EXCLUDED.steps,reviewed_by=EXCLUDED.reviewed_by,review_reason=EXCLUDED.review_reason,signature=EXCLUDED.signature,signer_key_id=EXCLUDED.signer_key_id,reviewed_at=EXCLUDED.reviewed_at,signed_at=EXCLUDED.signed_at,version=EXCLUDED.version WHERE precheck_plans.version=EXCLUDED.version-1`, value.ID, value.AnalysisID, value.TeamID, value.State, steps, value.CreatedBy, nullable(value.ReviewedBy), nullable(value.ReviewReason), value.ContentDigest, nullable(value.Signature), nullable(value.SignerKeyID), value.CreatedAt, nullableTime(value.ReviewedAt), nullableTime(value.SignedAt), value.Version)
	if err != nil {
		return fmt.Errorf("save plan: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return common.ErrVersionConflict
	}
	return nil
}
func (r *PlanRepository) Get(ctx context.Context, id string) (*precheck.Plan, error) {
	var value precheck.Plan
	var steps []byte
	err := r.db.Pool.QueryRow(ctx, `SELECT id,analysis_id,team_id,state,steps,created_by,COALESCE(reviewed_by,''),COALESCE(review_reason,''),content_digest,COALESCE(signature,''),COALESCE(signer_key_id,''),created_at,COALESCE(reviewed_at,'0001-01-01'),COALESCE(signed_at,'0001-01-01'),version FROM precheck_plans WHERE id=$1`, id).Scan(&value.ID, &value.AnalysisID, &value.TeamID, &value.State, &steps, &value.CreatedBy, &value.ReviewedBy, &value.ReviewReason, &value.ContentDigest, &value.Signature, &value.SignerKeyID, &value.CreatedAt, &value.ReviewedAt, &value.SignedAt, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read plan: %w", err)
	}
	if err := json.Unmarshal(steps, &value.Steps); err != nil {
		return nil, err
	}
	return &value, nil
}
func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
func nullableTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}
