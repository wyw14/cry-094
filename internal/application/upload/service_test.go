package upload

import (
	"context"
	"encoding/binary"
	"fmt"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/platform/files"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

type sequenceID struct{ value int }

func (s *sequenceID) NewID() string { s.value++; return fmt.Sprintf("id-%d", s.value) }
func TestUploadNormalizesEncodingAndStoresDigestBoundBytes(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	teams := memory.NewTeamRepo()
	owner, _ := team.New("team", "Operations", "actor", now)
	if err := teams.Create(ctx, owner); err != nil {
		t.Fatal(err)
	}
	artifacts := memory.NewArtifactRepo()
	store := files.New()
	ids := &sequenceID{}
	service := New(teams, artifacts, store, clock.Fixed{Value: now}, ids)
	library, err := service.CreateLibrary(ctx, "team", "actor", "Deploy", map[string]string{"tier": "prod"})
	if err != nil {
		t.Fatal(err)
	}
	content := utf16LE("jq . $CONFIG")
	artifact, err := service.Upload(ctx, Request{TeamID: "team", LibraryID: library.ID, ActorID: "actor", Filename: "check.ps1", MIME: "text/plain", Content: content, Version: 1})
	if err != nil {
		t.Fatal(err)
	}
	if artifact.Encoding != script.EncodingUTF16LE {
		t.Fatalf("got %s", artifact.Encoding)
	}
	stored, err := store.Get(ctx, artifact.StorageKey)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != "jq . $CONFIG" || files.Digest(stored) != artifact.Digest {
		t.Fatal("normalized immutable bytes must match digest")
	}
}
func utf16LE(value string) []byte {
	result := []byte{0xff, 0xfe}
	for _, r := range value {
		pair := make([]byte, 2)
		binary.LittleEndian.PutUint16(pair, uint16(r))
		result = append(result, pair...)
	}
	return result
}
