package analysisservice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/host"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/platform/files"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

type cancellationParser struct {
	started chan struct{}
	release chan struct{}
}

func (p *cancellationParser) Name() string  { return "cancellation-parser" }
func (p *cancellationParser) Build() string { return "cancellation-build" }
func (p *cancellationParser) Parse(ctx context.Context, _ script.Artifact, _ []byte) (script.ParseConclusion, error) {
	close(p.started)
	select {
	case <-ctx.Done():
		return script.ParseConclusion{}, ctx.Err()
	case <-p.release:
		return script.ParseConclusion{}, context.Canceled
	}
}

type cancellationParserFactory struct{ parser ports.Parser }

func (f cancellationParserFactory) For(script.Shell) (ports.Parser, error) { return f.parser, nil }

func TestAnalyzeStopsBeforePublishingWhenCallerCancels(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	teams := memory.NewTeamRepo()
	owner, err := team.New("team", "Operations", "analyst", now)
	if err != nil {
		t.Fatal(err)
	}
	if err := teams.Create(context.Background(), owner); err != nil {
		t.Fatal(err)
	}
	artifacts := memory.NewArtifactRepo()
	library := &script.Library{ID: "library", TeamID: "team", Name: "Checks", Tags: map[string]string{}, Version: 1, Created: now}
	if err := artifacts.CreateLibrary(context.Background(), library); err != nil {
		t.Fatal(err)
	}
	content := []byte("echo ready\n")
	objects := files.New()
	artifact := &script.Artifact{ID: "artifact", LibraryID: "library", Number: 1, Filename: "check.sh", Shell: script.ShellPOSIX, Encoding: script.EncodingUTF8, Digest: files.Digest(content), StorageKey: "team/library/check", Size: int64(len(content)), UploadedBy: "analyst", UploadedAt: now, Version: 1}
	if err := objects.Put(context.Background(), artifact.StorageKey, content); err != nil {
		t.Fatal(err)
	}
	if err := artifacts.CreateArtifact(context.Background(), artifact); err != nil {
		t.Fatal(err)
	}
	parser := &cancellationParser{started: make(chan struct{}), release: make(chan struct{})}
	service := New(teams, artifacts, memory.NewAnalysisRepo(), objects, cancellationParserFactory{parser: parser}, clock.Fixed{Value: now}, &testIDs{})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := service.Analyze(ctx, "team", "analyst", "library", []string{"artifact"}, host.Inventory{Tools: map[string]host.Tool{}, EnvNames: map[string]bool{}})
		done <- err
	}()
	select {
	case <-parser.started:
	case <-time.After(time.Second):
		t.Fatal("parser did not start")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("expected caller cancellation, got %v", err)
		}
	case <-time.After(150 * time.Millisecond):
		close(parser.release)
		<-done
		t.Fatal("analysis continued after caller cancellation")
	}
}
