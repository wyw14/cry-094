package idempotency

import (
	"context"
	"sync"
)

type Store struct {
	mu      sync.Mutex
	results map[string][]byte
}

func New() *Store { return &Store{results: make(map[string][]byte)} }

func (s *Store) Replay(_ context.Context, key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.results[key]
	return append([]byte(nil), value...), ok
}

func (s *Store) Save(_ context.Context, key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.results[key]; !exists {
		s.results[key] = append([]byte(nil), value...)
	}
}
