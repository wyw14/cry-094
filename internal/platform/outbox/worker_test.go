package outbox

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestWorkerRetriesThenStopsOnCancellation(t *testing.T) {
	var calls atomic.Int32
	worker := New(2, func(context.Context, Event) error { calls.Add(1); return errors.New("offline") })
	worker.Enqueue(Event{Topic: "analysis.completed"})
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()
	if err := worker.Run(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("got %v", err)
	}
	if calls.Load() != 3 {
		t.Fatalf("got %d attempts", calls.Load())
	}
}
