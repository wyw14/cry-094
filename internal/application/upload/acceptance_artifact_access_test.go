package upload

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/platform/files"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

func TestArtifactReadsStayInsideAuthorizedTeam(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	teams := memory.NewTeamRepo()
	teamA, _ := team.New("team-a", "Operations A", "viewer-a", now)
	teamB, _ := team.New("team-b", "Operations B", "owner-b", now)
	if err := teams.Create(ctx, teamA); err != nil {
		t.Fatal(err)
	}
	if err := teams.Create(ctx, teamB); err != nil {
		t.Fatal(err)
	}
	artifacts := memory.NewArtifactRepo()
	library := &script.Library{ID: "library-b", TeamID: "team-b", Name: "Private checks", Tags: map[string]string{}, Version: 1, Created: now}
	if err := artifacts.CreateLibrary(ctx, library); err != nil {
		t.Fatal(err)
	}
	content := []byte("jq . secret.json")
	objects := files.New()
	artifact := &script.Artifact{ID: "artifact-b", LibraryID: library.ID, Number: 1, Filename: "private.sh", Shell: script.ShellPOSIX, Encoding: script.EncodingUTF8, Digest: files.Digest(content), StorageKey: "team-b/library-b/private", Size: int64(len(content)), UploadedBy: "owner-b", UploadedAt: now, Version: 1}
	if err := objects.Put(ctx, artifact.StorageKey, content); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.CreateArtifact(ctx, artifact); err != nil {
		t.Fatal(err)
	}
	service := New(teams, artifacts, objects, clock.Fixed{Value: now}, &artifactAccessIDs{})
	items, _, listErr := service.List(ctx, "team-a", "viewer-a", library.ID, "", 20)
	downloaded, _, downloadErr := service.Download(ctx, "team-a", "viewer-a", artifact.ID)
	if listErr == nil || len(items) != 0 {
		t.Fatalf("cross-team listing exposed %#v", items)
	}
	if downloadErr == nil || len(downloaded) != 0 {
		t.Fatalf("cross-team download exposed %q", downloaded)
	}
}

type artifactAccessIDs struct{}

func (*artifactAccessIDs) NewID() string { return "unused" }
