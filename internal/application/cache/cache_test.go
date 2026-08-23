package cache

import (
	"context"
	"sync"
	"testing"

	"github.com/wyw14/cry-094/internal/domain/analysis"
)

func TestConcurrentCacheInvalidationDoesNotReturnCrossVersionResult(t *testing.T) {
	store := New()
	ctx := context.Background()
	key := Key{LibraryID: "library", ArtifactHash: "old", ParserBuild: "v1", InventoryHash: "host"}
	store.Put(ctx, key, analysis.Result{ID: "old"})
	start := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); <-start; store.InvalidateLibrary(ctx, "library") }()
	go func() {
		defer wg.Done()
		<-start
		store.Put(ctx, Key{LibraryID: "library", ArtifactHash: "new", ParserBuild: "v2", InventoryHash: "host"}, analysis.Result{ID: "new"})
	}()
	close(start)
	wg.Wait()
	if value, ok := store.Get(ctx, key); ok && value.ID != "old" {
		t.Fatalf("cache key returned wrong result %#v", value)
	}
}
