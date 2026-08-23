package analysisservice

import (
	"context"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/analysis"
	"github.com/wyw14/cry-094/internal/domain/host"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
	"github.com/wyw14/cry-094/internal/parser"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/platform/files"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

func TestRepeatedToolUseKeepsDeclaredVersionConstraint(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	teams := memory.NewTeamRepo()
	owner, _ := team.New("team", "Operations", "analyst", now)
	if err := teams.Create(ctx, owner); err != nil {
		t.Fatal(err)
	}
	artifacts := memory.NewArtifactRepo()
	library := &script.Library{ID: "library", TeamID: "team", Name: "Checks", Tags: map[string]string{}, Version: 1, Created: now}
	if err := artifacts.CreateLibrary(ctx, library); err != nil {
		t.Fatal(err)
	}
	content := []byte("# requires: jq >=2.1\njq . first.json\njq . second.json\n")
	objects := files.New()
	artifact := &script.Artifact{ID: "artifact", LibraryID: library.ID, Number: 1, Filename: "check.sh", Shell: script.ShellPOSIX, Encoding: script.EncodingUTF8, Digest: files.Digest(content), StorageKey: "team/library/check", Size: int64(len(content)), UploadedBy: "analyst", UploadedAt: now, Version: 1}
	if err := objects.Put(ctx, artifact.StorageKey, content); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.CreateArtifact(ctx, artifact); err != nil {
		t.Fatal(err)
	}
	service := New(teams, artifacts, memory.NewAnalysisRepo(), objects, parser.Factory{BuildID: "parser-v7"}, clock.Fixed{Value: now}, &acceptanceVersionIDs{})
	inventory := host.Inventory{Tools: map[string]host.Tool{"jq": {Name: "jq", Version: script.SemanticVersion{Major: 2}}}, EnvNames: map[string]bool{}}
	result, err := service.Analyze(ctx, "team", "analyst", library.ID, []string{artifact.ID}, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != analysis.StatusBlocked {
		t.Fatalf("incompatible repeated tool use produced status %s", result.Status)
	}
	conflict := false
	for _, issue := range result.Issues {
		if issue.Kind == analysis.IssueVersionConflict && issue.Severity == analysis.SeverityBlocking {
			conflict = true
		}
	}
	if !conflict {
		t.Fatalf("version conflict evidence missing: %#v", result.Issues)
	}
}

type acceptanceVersionIDs struct{}

func (*acceptanceVersionIDs) NewID() string { return "analysis-id" }
