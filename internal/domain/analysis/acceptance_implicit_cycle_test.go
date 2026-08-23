package analysis

import (
	"testing"
	"time"
)

func TestImplicitDependencyCycleBlocksOrchestration(t *testing.T) {
	graph, err := NewGraph(
		[]Node{{ArtifactID: "prepare", Digest: "a"}, {ArtifactID: "deploy", Digest: "b"}, {ArtifactID: "verify", Digest: "c"}},
		[]Edge{
			{From: "prepare", To: "deploy", Contract: "manifest", Evidence: "declared output"},
			{From: "deploy", To: "verify", Contract: "receipt", Evidence: "declared output"},
			{From: "verify", To: "prepare", Contract: "shared.env", Implicit: true, Evidence: "sourced library"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	result := NewResult("analysis", "library", graph, nil, "hash", "parser-v1", time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if result.Status != StatusBlocked {
		t.Fatalf("cycle with implicit edge produced status %s", result.Status)
	}
	if result.CanOrchestrate() {
		t.Fatal("cyclic dependency graph must not be orchestratable")
	}
	found := false
	for _, issue := range result.Issues {
		if issue.Kind == IssueCycle && issue.Severity == SeverityBlocking && len(issue.Evidence) == 3 {
			found = true
		}
	}
	if !found {
		t.Fatalf("blocking evidence chain missing: %#v", result.Issues)
	}
}
