package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-094/internal/domain/common"
	"github.com/wyw14/cry-094/internal/domain/script"
)

type ArtifactRepository struct{ db *Database }

func NewArtifactRepository(db *Database) *ArtifactRepository { return &ArtifactRepository{db: db} }
func (r *ArtifactRepository) CreateLibrary(ctx context.Context, value *script.Library) error {
	_, err := r.db.Pool.Exec(ctx, `INSERT INTO script_libraries(id,team_id,name,tags,version,created_at) VALUES($1,$2,$3,$4,$5,$6)`, value.ID, value.TeamID, value.Name, value.Tags, value.Version, value.Created)
	if err != nil {
		return fmt.Errorf("insert library: %w", err)
	}
	return nil
}
func (r *ArtifactRepository) GetLibrary(ctx context.Context, id string) (*script.Library, error) {
	var value script.Library
	err := r.db.Pool.QueryRow(ctx, `SELECT id,team_id,name,tags,version,created_at FROM script_libraries WHERE id=$1`, id).Scan(&value.ID, &value.TeamID, &value.Name, &value.Tags, &value.Version, &value.Created)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read library: %w", err)
	}
	return &value, nil
}
func (r *ArtifactRepository) CreateArtifact(ctx context.Context, value *script.Artifact) error {
	return r.db.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		var teamID string
		if err := tx.QueryRow(ctx, `SELECT team_id FROM script_libraries WHERE id=$1 FOR SHARE`, value.LibraryID).Scan(&teamID); err != nil {
			return fmt.Errorf("lock library: %w", err)
		}
		_, err := tx.Exec(ctx, `INSERT INTO script_artifacts(id,library_id,number,filename,shell,encoding,digest,storage_key,size_bytes,uploaded_by,uploaded_at,parser_build,version) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`, value.ID, value.LibraryID, value.Number, value.Filename, value.Shell, value.Encoding, value.Digest, value.StorageKey, value.Size, value.UploadedBy, value.UploadedAt, value.ParserBuild, value.Version)
		if err != nil {
			return fmt.Errorf("insert artifact: %w", err)
		}
		return nil
	})
}
func (r *ArtifactRepository) GetArtifact(ctx context.Context, id string) (*script.Artifact, error) {
	var value script.Artifact
	err := r.db.Pool.QueryRow(ctx, `SELECT id,library_id,number,filename,shell,encoding,digest,storage_key,size_bytes,uploaded_by,uploaded_at,parser_build,version FROM script_artifacts WHERE id=$1`, id).Scan(&value.ID, &value.LibraryID, &value.Number, &value.Filename, &value.Shell, &value.Encoding, &value.Digest, &value.StorageKey, &value.Size, &value.UploadedBy, &value.UploadedAt, &value.ParserBuild, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, common.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read artifact: %w", err)
	}
	return &value, nil
}
func (r *ArtifactRepository) ListArtifacts(ctx context.Context, libraryID, cursor string, limit int) ([]script.Artifact, string, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	cursorTime := time.Unix(0, 0).UTC()
	cursorID := ""
	if cursor != "" {
		if err := r.db.Pool.QueryRow(ctx, `SELECT uploaded_at,id FROM script_artifacts WHERE id=$1`, cursor).Scan(&cursorTime, &cursorID); err != nil {
			return nil, "", fmt.Errorf("resolve cursor: %w", err)
		}
	}
	rows, err := r.db.Pool.Query(ctx, `SELECT id,library_id,number,filename,shell,encoding,digest,storage_key,size_bytes,uploaded_by,uploaded_at,parser_build,version FROM script_artifacts WHERE library_id=$1 AND (uploaded_at,id)>($2,$3) ORDER BY uploaded_at,id LIMIT $4`, libraryID, cursorTime, cursorID, limit+1)
	if err != nil {
		return nil, "", fmt.Errorf("list artifacts: %w", err)
	}
	defer rows.Close()
	values := make([]script.Artifact, 0, limit+1)
	for rows.Next() {
		var value script.Artifact
		if err := rows.Scan(&value.ID, &value.LibraryID, &value.Number, &value.Filename, &value.Shell, &value.Encoding, &value.Digest, &value.StorageKey, &value.Size, &value.UploadedBy, &value.UploadedAt, &value.ParserBuild, &value.Version); err != nil {
			return nil, "", err
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	next := ""
	if len(values) > limit {
		next = values[limit-1].ID
		values = values[:limit]
	}
	return values, next, nil
}
