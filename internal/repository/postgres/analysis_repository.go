package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-094/internal/domain/analysis"
	"github.com/wyw14/cry-094/internal/domain/common"
)

type AnalysisRepository struct{ db *Database }

func NewAnalysisRepository(db *Database) *AnalysisRepository { return &AnalysisRepository{db: db} }
func (r *AnalysisRepository) Save(ctx context.Context, value *analysis.Result) error {
	graph, err := json.Marshal(value.Graph)
	if err != nil {
		return err
	}
	issues, err := json.Marshal(value.Issues)
	if err != nil {
		return err
	}
	checks, err := json.Marshal(value.CapabilityChecks)
	if err != nil {
		return err
	}
	_, err = r.db.Pool.Exec(ctx, `INSERT INTO analyses(id,library_id,graph,execution_order,issues,capability_checks,status,artifact_hash,parser_build,completed_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`, value.ID, value.LibraryID, graph, value.Order, issues, checks, value.Status, value.ArtifactHash, value.ParserBuild, value.CompletedAt)
	if err != nil {
		return fmt.Errorf("insert analysis: %w", err)
	}
	return nil
}
func (r *AnalysisRepository) Get(ctx context.Context, id string) (*analysis.Result, error) {
	var value analysis.Result
	var graph, issues, checks []byte
	err := r.db.Pool.QueryRow(ctx, `SELECT id,library_id,graph,execution_order,issues,capability_checks,status,artifact_hash,parser_build,completed_at FROM analyses WHERE id=$1`, id).Scan(&value.ID, &value.LibraryID, &graph, &value.Order, &issues, &checks, &value.Status, &value.ArtifactHash, &value.ParserBuild, &value.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read analysis: %w", err)
	}
	if err := json.Unmarshal(graph, &value.Graph); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(issues, &value.Issues); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(checks, &value.CapabilityChecks); err != nil {
		return nil, err
	}
	return &value, nil
}
