package outbox

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestSuccessfulEventIsDeliveredOnlyOnce(t *testing.T) {
	var calls atomic.Int32
	delivered := make(chan struct{}, 2)
	worker := New(2, func(context.Context, Event) error {
		calls.Add(1)
		delivered <- struct{}{}
		return nil
	})
	worker.Enqueue(Event{Topic: "analysis.completed", Payload: []byte("analysis-1")})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	select {
	case <-delivered:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("event was not delivered")
	}
	select {
	case <-delivered:
		cancel()
		<-done
		t.Fatalf("successful event was delivered %d times", calls.Load())
	case <-time.After(30 * time.Millisecond):
	}
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("worker stopped with %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("successful event was delivered %d times", calls.Load())
	}
}
