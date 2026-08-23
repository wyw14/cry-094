package integration

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/wyw14/cry-094/internal/repository/postgres"
)

func openDatabase(t *testing.T) *postgres.Database {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	db, err := postgres.Open(ctx, url, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	return db
}
func TestSerializableTransactionRollsBackAllWrites(t *testing.T) {
	db := openDatabase(t)
	ctx := context.Background()
	marker := fmt.Sprintf("rollback-%d", time.Now().UnixNano())
	err := db.WithinTx(ctx, func(ctx context.Context, tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO outbox_events(id,topic,payload,next_attempt_at,created_at) VALUES(gen_random_uuid(),$1,'{}',now(),now())`, marker); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("transaction should return error")
	}
	var count int
	if err := db.Pool.QueryRow(ctx, `SELECT count(*) FROM outbox_events WHERE topic=$1`, marker).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("rollback left %d rows", count)
	}
}
