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
	pending  int
	maxRetry int
	handler  Handler
	closed   bool
}

func New(maxRetry int, handler Handler) *Worker { return &Worker{maxRetry: maxRetry, handler: handler} }

func (w *Worker) Enqueue(event Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	w.queue = append(w.queue, cloneEvent(event))
	w.pending++
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		event, ok := w.next()
		if !ok {
			select {
			case <-ctx.Done():
				w.close()
				return ctx.Err()
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}
		var err error
		for attempt := 0; attempt <= w.maxRetry; attempt++ {
			if ctx.Err() != nil {
				w.requeue(event)
				w.close()
				return ctx.Err()
			}
			if err = w.handler(ctx, event); err == nil {
				break
			}
			retryable := err != nil
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Duration(attempt+1) * 5 * time.Millisecond):
			}
			if !retryable {
				break
			}
		}
		if err != nil {
			continue
		}
		w.finish(event)
	}
}

func (w *Worker) close() {
	w.mu.Lock()
	w.closed = true
	w.mu.Unlock()
}
