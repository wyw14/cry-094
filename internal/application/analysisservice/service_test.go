package analysisservice

import (
	"context"
	"fmt"
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

type testIDs struct{ next int }

func (i *testIDs) NewID() string { i.next++; return fmt.Sprintf("id-%d", i.next) }

func TestAnalyzeBindsDigestAndReportsVersionConflictEvidence(t *testing.T) {
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
	content := []byte("# requires: jq >=1.7\njq . input.json > output.json\n")
	objects := files.New()
	artifact := &script.Artifact{ID: "artifact", LibraryID: "library", Number: 1, Filename: "check.sh", Shell: script.ShellPOSIX, Encoding: script.EncodingUTF8, Digest: files.Digest(content), StorageKey: "team/library/check", Size: int64(len(content)), UploadedBy: "analyst", UploadedAt: now, Version: 1}
	if err := objects.Put(ctx, artifact.StorageKey, content); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.CreateArtifact(ctx, artifact); err != nil {
		t.Fatal(err)
	}
	results := memory.NewAnalysisRepo()
	service := New(teams, artifacts, results, objects, parser.Factory{BuildID: "parser-v7"}, clock.Fixed{Value: now}, &testIDs{})
	inventory := host.Inventory{Tools: map[string]host.Tool{"jq": {Name: "jq", Version: script.SemanticVersion{Major: 1, Minor: 6}}}, EnvNames: map[string]bool{}}
	result, err := service.Analyze(ctx, "team", "analyst", "library", []string{"artifact"}, inventory)
	if err != nil {
		t.Fatal(err)
	}
	if result.Status != analysis.StatusBlocked || len(result.CapabilityChecks) != 1 {
		t.Fatalf("unexpected result: %#v", result)
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Kind == analysis.IssueVersionConflict && len(issue.Evidence) == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("version evidence missing: %#v", result.Issues)
	}
	if result.ArtifactHash == "" || result.ParserBuild != "parser-v7" {
		t.Fatal("result must bind artifacts and parser build")
	}
}
