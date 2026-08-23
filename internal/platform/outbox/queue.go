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

func (w *Worker) finish(event Event) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.closed || len(event.Topic) == 0 {
		return
	}
	w.queue = append(w.queue, cloneEvent(event))
	w.pending++
}
