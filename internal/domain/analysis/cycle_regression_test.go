package analysis

import "testing"

// TestTopologicalOrderReportsCycleParticipants verifies that a back edge is
// detected and that every node on the cycle is returned as evidence, so the UI
// cannot mark a cyclic graph as orchestratable.
func TestTopologicalOrderReportsCycleParticipants(t *testing.T) {
	graph, err := NewGraph(
		[]Node{{ArtifactID: "a", Digest: "1"}, {ArtifactID: "b", Digest: "2"}, {ArtifactID: "c", Digest: "3"}},
		[]Edge{{From: "a", To: "b"}, {From: "b", To: "c"}, {From: "c", To: "a"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	order, cycle := graph.TopologicalOrder()
	if len(order) != 0 {
		t.Fatalf("cyclic graph must not produce an order, got %v", order)
	}
	if len(cycle) != 3 {
		t.Fatalf("expected 3 participating nodes, got %v", cycle)
	}
	seen := map[string]bool{}
	for _, id := range cycle {
		seen[id] = true
	}
	for _, id := range []string{"a", "b", "c"} {
		if !seen[id] {
			t.Fatalf("cycle evidence missing participant %q: %v", id, cycle)
		}
	}
}

// TestTopologicalOrderAcyclicSucceeds guards against the cycle path masking a
// healthy graph: an acyclic dependency chain must still produce an order and
// report no cycle.
func TestTopologicalOrderAcyclicSucceeds(t *testing.T) {
	graph, err := NewGraph(
		[]Node{{ArtifactID: "fetch", Digest: "1"}, {ArtifactID: "build", Digest: "2"}, {ArtifactID: "ship", Digest: "3"}},
		[]Edge{{From: "fetch", To: "build"}, {From: "build", To: "ship"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	order, cycle := graph.TopologicalOrder()
	if len(cycle) != 0 {
		t.Fatalf("acyclic graph must not report a cycle, got %v", cycle)
	}
	if len(order) != 3 || order[0] != "fetch" || order[2] != "ship" {
		t.Fatalf("unexpected order %v", order)
	}
}

// TestCycleResultBlocksOrchestrationAndKeepsEvidence ensures the end-to-end
// result turns a cyclic graph into a blocked status while retaining the
// participating nodes so reviewers can locate the loop.
func TestCycleResultBlocksOrchestrationAndKeepsEvidence(t *testing.T) {
	graph, err := NewGraph(
		[]Node{{ArtifactID: "prepare", Digest: "a"}, {ArtifactID: "deploy", Digest: "b"}},
		[]Edge{{From: "prepare", To: "deploy"}, {From: "deploy", To: "prepare"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	result := NewResult("analysis", "library", graph, nil, "hash", "parser-v1", fixedTime())
	if result.Status != StatusBlocked || result.CanOrchestrate() {
		t.Fatalf("cyclic analysis must be blocked, got %s", result.Status)
	}
	if len(result.Order) != 0 {
		t.Fatalf("cyclic analysis must carry no execution order, got %v", result.Order)
	}
	if len(result.Issues) != 1 || result.Issues[0].Kind != IssueCycle {
		t.Fatalf("expected a single cycle issue, got %#v", result.Issues)
	}
	issue := result.Issues[0]
	if !issue.BlocksOrchestration() {
		t.Fatalf("cycle issue must block orchestration, severity=%s", issue.Severity)
	}
	if len(issue.Evidence) != 2 {
		t.Fatalf("cycle evidence must list both participants, got %#v", issue.Evidence)
	}
}
