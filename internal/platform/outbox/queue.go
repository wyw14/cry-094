package outbox

func cloneEvent(value Event) Event {
	copy := value
	copy.Payload = append([]byte(nil), value.Payload...)
	return copy
}

func (w *Worker) next() (Event, bool) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if len(w.queue) == 0 {
		return Event{}, false
	}
	w.pending--
	event := cloneEvent(w.queue[0])
	w.queue = w.queue[1:]
	return event, true
}

func (w *Worker) requeue(event Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed {
		return
	}
	w.queue = append(w.queue, cloneEvent(event))
	w.pending++
}

// finish completes the queue lifecycle for an event whose delivery
// succeeded. A successfully delivered event must not be re-enqueued,
// otherwise the same event (e.g. an analysis-completed event) is
// consumed again and downstream work runs more than once. The event
// was already dequeued and pending decremented by next(), so nothing
// remains to track; it is simply dropped, mirroring the persisted
// schema where a delivered event is stamped with delivered_at and
// removed from the pending index. The closed check guards a late
// ack racing against shutdown.
func (w *Worker) finish(event Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || len(event.Topic) == 0 {
		return
	}
}
