package cache

import (
	"context"
	"testing"

	"github.com/wyw14/cry-094/internal/domain/analysis"
)

func TestInvalidationRejectsAnalysisStartedOnOlderGeneration(t *testing.T) {
	ctx := context.Background()
	store := New()
	generation := store.Generation(ctx, "library")
	key := Key{LibraryID: "library", ArtifactHash: "artifact-v1", ParserBuild: "parser-v1", InventoryHash: "host-v1", LibraryGeneration: generation}
	result := analysis.Result{ID: "analysis-v1", LibraryID: "library"}
	result.BindCacheGeneration(generation)
	if removed := store.InvalidateLibrary(ctx, "library"); removed != 0 {
		t.Fatalf("empty invalidation removed %d entries", removed)
	}
	if store.Put(ctx, key, result) {
		t.Fatal("stale analysis was published after library invalidation")
	}
	if value, ok := store.Get(ctx, key); ok {
		t.Fatalf("stale graph remained visible: %#v", value)
	}
}
