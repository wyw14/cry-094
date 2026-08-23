package cache

import (
	"context"
	"sync"
	"time"

	"github.com/wyw14/cry-094/internal/domain/analysis"
)

type Key struct {
	LibraryID, ArtifactHash, ParserBuild, InventoryHash string
	LibraryGeneration                                   uint64
}
type Store struct {
	mu            sync.RWMutex
	values        map[Key]analysis.Result
	byLibrary     map[string]map[Key]bool
	generation    map[string]uint64
	pending       map[string][]Key
	lastPublished map[string]time.Time
}

func New() *Store {
	return &Store{
		values: make(map[Key]analysis.Result), byLibrary: make(map[string]map[Key]bool),
		generation: make(map[string]uint64), pending: make(map[string][]Key),
		lastPublished: make(map[string]time.Time),
	}
}
func (s *Store) Get(_ context.Context, key Key) (analysis.Result, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.values[key]
	return value, ok
}
func (s *Store) Generation(_ context.Context, libraryID string) uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.generation[libraryID]
}
func (s *Store) Put(_ context.Context, key Key, value analysis.Result) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	current := s.generation[key.LibraryID]
	if !value.CanPublishAt(current) {
		return false
	}
	value.AcceptCachePublication(current)
	s.values[key] = value
	if s.byLibrary[key.LibraryID] == nil {
		s.byLibrary[key.LibraryID] = map[Key]bool{}
	}
	s.byLibrary[key.LibraryID][key] = true
	s.pending[key.LibraryID] = append(s.pending[key.LibraryID], key)
	s.lastPublished[key.LibraryID] = time.Now().UTC()
	return true
}
func (s *Store) InvalidateLibrary(_ context.Context, libraryID string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := s.byLibrary[libraryID]
	for key := range keys {
		delete(s.values, key)
	}
	delete(s.byLibrary, libraryID)
	s.generation[libraryID]++
	for _, key := range s.pending[libraryID] {
		if value, ok := s.values[key]; ok {
			value.MarkCacheInvalidated(time.Now())
			s.values[key] = value
		}
	}
	delete(s.lastPublished, libraryID)
	return len(keys)
}
