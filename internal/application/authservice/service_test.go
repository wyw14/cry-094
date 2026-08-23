package authservice

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/identity"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

type testClock struct{ value time.Time }

func (c testClock) Now() time.Time { return c.value }

type seqIDs struct{ value int64 }

func (i *seqIDs) NewID() string {
	n := atomic.AddInt64(&i.value, 1)
	return fmt.Sprintf("00000000-0000-4000-8000-%012d", n)
}

func newService(t *testing.T) (*Service, *memory.IdentityRepo, string) {
	t.Helper()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	user, err := identity.Register("owner", "owner@example.test", "secret-pass", now)
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	repo := memory.NewIdentityRepo(user)
	clk := testClock{value: now}
	ids := &seqIDs{}
	svc := New(repo, clk, ids, []byte("0123456789abcdef0123456789abcdef"), 15*time.Minute, time.Hour)
	tokens, err := svc.Login(context.Background(), "owner@example.test", "secret-pass")
	if err != nil {
		t.Fatalf("seed login: %v", err)
	}
	return svc, repo, tokens.Refresh
}

// TestRefreshConcurrentMintsSingleSuccessor drives the legacy race: two
// refreshes for the same short-lived token fire at once. Before the fix both
// won, minting two usable successors. After the fix only one wins; the loser is
// rejected (either by the conditional claim returning ErrConflict, or by
// observing the already-revoked record as not usable), so exactly one usable
// successor exists and the source is consumed once.
func TestRefreshConcurrentMintsSingleSuccessor(t *testing.T) {
	svc, repo, refresh := newService(t)

	const concurrency = 2
	var (
		wg         sync.WaitGroup
		barrier    sync.WaitGroup
		successCnt int64
		resultsMu  sync.Mutex
		results    []Tokens
	)
	barrier.Add(1)
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			barrier.Wait()
			tokens, err := svc.Refresh(context.Background(), refresh)
			if err == nil {
				atomic.AddInt64(&successCnt, 1)
				resultsMu.Lock()
				results = append(results, tokens)
				resultsMu.Unlock()
			}
		}()
	}
	barrier.Done()
	wg.Wait()

	if success := atomic.LoadInt64(&successCnt); success != 1 {
		t.Fatalf("expected exactly one successful refresh (no double-spend), got %d", success)
	}

	// The winner's successor must itself be single-use: a follow-up refresh of
	// the new token works once; a second immediate reuse is rejected because the
	// successor has already been revoked by the first use.
	winner := results[0]
	if _, err := svc.Refresh(context.Background(), winner.Refresh); err != nil {
		t.Fatalf("successor refresh should succeed once: %v", err)
	}
	if _, err := svc.Refresh(context.Background(), winner.Refresh); err == nil {
		t.Fatal("reusing the successor must fail, it was already consumed")
	}

	// Source token must be revoked exactly once, not double-consumed.
	stored, err := repo.RefreshByDigest(context.Background(), identity.TokenDigest(refresh))
	if err != nil {
		t.Fatalf("lookup source token: %v", err)
	}
	if stored.RevokedAt == nil {
		t.Fatal("source refresh token was not revoked")
	}
}
