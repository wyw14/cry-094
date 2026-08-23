package analysis

import (
	"testing"
	"time"
)

func TestGraphReportsCycleEvidenceAndBlocks(t *testing.T) {
	graph, err := NewGraph([]Node{{ArtifactID: "prepare", Digest: "a"}, {ArtifactID: "deploy", Digest: "b"}}, []Edge{{From: "prepare", To: "deploy", Contract: "manifest"}, {From: "deploy", To: "prepare", Contract: "rollback"}})
	if err != nil {
		t.Fatal(err)
	}
	result := NewResult("analysis", "library", graph, nil, "hash", "parser-v1", fixedTime())
	if result.Status != StatusBlocked {
		t.Fatalf("got %s", result.Status)
	}
	if len(result.Issues) != 1 || result.Issues[0].Kind != IssueCycle || len(result.Issues[0].Evidence) != 2 {
		t.Fatalf("missing cycle evidence: %#v", result.Issues)
	}
}
func fixedTime() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }
