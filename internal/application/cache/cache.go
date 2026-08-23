package cache

import (
	"context"
	"sync"

	"github.com/wyw14/cry-094/internal/domain/analysis"
)

type Key struct{ LibraryID, ArtifactHash, ParserBuild, InventoryHash string }
type Store struct {
	mu        sync.RWMutex
	values    map[Key]analysis.Result
	byLibrary map[string]map[Key]bool
}

func New() *Store {
	return &Store{values: make(map[Key]analysis.Result), byLibrary: make(map[string]map[Key]bool)}
}
func (s *Store) Get(_ context.Context, key Key) (analysis.Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[key]
	return value, ok
}
func (s *Store) Put(_ context.Context, key Key, value analysis.Result) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.values[key] = value
	if s.byLibrary[key.LibraryID] == nil {
		s.byLibrary[key.LibraryID] = map[Key]bool{}
	}
	s.byLibrary[key.LibraryID][key] = true
}
func (s *Store) InvalidateLibrary(_ context.Context, libraryID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := s.byLibrary[libraryID]
	for key := range keys {
		delete(s.values, key)
	}
	delete(s.byLibrary, libraryID)
	return len(keys)
}
