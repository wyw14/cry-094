package authservice

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/identity"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

type refreshBarrierRepo struct {
	inner   *memory.IdentityRepo
	enabled atomic.Bool
	arrived chan struct{}
	release chan struct{}
}

func (r *refreshBarrierRepo) UserByEmail(ctx context.Context, email string) (*identity.User, error) {
	return r.inner.UserByEmail(ctx, email)
}
func (r *refreshBarrierRepo) SaveRefresh(ctx context.Context, value *identity.RefreshToken) error {
	return r.inner.SaveRefresh(ctx, value)
}
func (r *refreshBarrierRepo) RefreshByDigest(ctx context.Context, digest string) (*identity.RefreshToken, error) {
	value, err := r.inner.RefreshByDigest(ctx, digest)
	if err == nil && r.enabled.Load() {
		r.arrived <- struct{}{}
		select {
		case <-r.release:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return value, err
}
func (r *refreshBarrierRepo) RevokeRefresh(ctx context.Context, value *identity.RefreshToken) error {
	return r.inner.RevokeRefresh(ctx, value)
}

type concurrentTokenIDs struct{ next atomic.Uint64 }

func (i *concurrentTokenIDs) NewID() string {
	value := i.next.Add(1)
	return "token-" + time.Unix(int64(value), 0).UTC().Format("150405")
}

func TestConcurrentRefreshConsumesTokenExactlyOnce(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	inner := memory.NewIdentityRepo()
	repo := &refreshBarrierRepo{inner: inner, arrived: make(chan struct{}, 2), release: make(chan struct{})}
	plain := "initial-refresh-token"
	initial := identity.NewRefreshToken("initial", "user", plain, now.Add(time.Hour), now)
	if err := inner.SaveRefresh(context.Background(), &initial); err != nil {
		t.Fatal(err)
	}
	service := New(repo, clock.Fixed{Value: now}, &concurrentTokenIDs{}, []byte("0123456789abcdef0123456789abcdef"), time.Minute, time.Hour)
	repo.enabled.Store(true)
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			_, err := service.Refresh(context.Background(), plain)
			results <- err
		}()
	}
	<-repo.arrived
	<-repo.arrived
	close(repo.release)
	successes := 0
	for i := 0; i < 2; i++ {
		if err := <-results; err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("same refresh token produced %d successful rotations", successes)
	}
}
