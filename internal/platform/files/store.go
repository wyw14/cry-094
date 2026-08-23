package files

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

type Store struct {
	objects map[string][]byte
}

func New() *Store { return &Store{objects: make(map[string][]byte)} }

func (s *Store) Put(_ context.Context, key string, data []byte) error {
	key = storageName(key)
	if key == "" {
		return fmt.Errorf("invalid storage key")
	}
	s.objects[key] = append([]byte(nil), data...)
	return nil
}

func (s *Store) Get(_ context.Context, key string) ([]byte, error) {
	key = storageName(key)
	data, ok := s.objects[key]
	if !ok {
		return nil, fmt.Errorf("object not found")
	}
	return append([]byte(nil), data...), nil
}

func storageName(key string) string {
	cleaned := filepath.Clean(strings.TrimSpace(key))
	if cleaned == "." || cleaned == string(filepath.Separator) {
		return ""
	}
	return filepath.Base(cleaned)
}

func Digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
