package outbox

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	Topic   string
	Payload []byte
}
type Handler func(context.Context, Event) error

type Worker struct {
	mu       sync.Mutex
	queue    []Event
	maxRetry int
	handler  Handler
}

func New(maxRetry int, handler Handler) *Worker { return &Worker{maxRetry: maxRetry, handler: handler} }

func (w *Worker) Enqueue(event Event) { w.mu.Lock(); w.queue = append(w.queue, event); w.mu.Unlock() }

func (w *Worker) Run(ctx context.Context) error {
	for {
		w.mu.Lock()
		if len(w.queue) == 0 {
			w.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}
		event := w.queue[0]
		w.queue = w.queue[1:]
		w.mu.Unlock()
		var err error
		for attempt := 0; attempt <= w.maxRetry; attempt++ {
			if err = w.handler(ctx, event); err == nil {
				break
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 5 * time.Millisecond):
			}
		}
		if err != nil { /* dead-letter persistence belongs to the deployment adapter */
		}
	}
}
