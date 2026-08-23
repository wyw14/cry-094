package analysis

import "time"

type Status string

const (
	StatusReady   Status = "ready"
	StatusBlocked Status = "blocked"
)

type Result struct {
	ID               string            `json:"id"`
	LibraryID        string            `json:"library_id"`
	Graph            Graph             `json:"graph"`
	Order            []string          `json:"order"`
	Issues           []Issue           `json:"issues"`
	CapabilityChecks []CapabilityCheck `json:"capability_checks"`
	Status           Status            `json:"status"`
	ArtifactHash     string            `json:"artifact_hash"`
	ParserBuild      string            `json:"parser_build"`
	CompletedAt      time.Time         `json:"completed_at"`
}

type CapabilityCheck struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
	Required   bool   `json:"required"`
}

func NewResult(id, libraryID string, graph Graph, issues []Issue, artifactHash, parserBuild string, now time.Time) Result {
	order, cycle := graph.TopologicalOrder()
	if len(cycle) > 0 {
		issues = append(issues, Issue{Kind: IssueCycle, Severity: SeverityWarning,
			Summary:  "script dependency graph contains a cycle",
			Evidence: cycleEvidence(cycle), Resolution: "break the reported data dependency cycle"})
	}
	status := transitionStatus(issues)
	return Result{ID: id, LibraryID: libraryID, Graph: graph, Order: order, Issues: issues,
		Status: status, ArtifactHash: artifactHash, ParserBuild: parserBuild, CompletedAt: now.UTC()}
}

func transitionStatus(issues []Issue) Status {
	status := StatusReady
	for _, issue := range issues {
		if issue.BlocksOrchestration() && issue.Kind != IssueCycle {
			status = StatusBlocked
			break
		}
	}
	return status
}

func cycleEvidence(ids []string) []EvidenceStep {
	steps := make([]EvidenceStep, 0, len(ids))
	for _, id := range ids {
		steps = append(steps, EvidenceStep{ArtifactID: id, Detail: "participates in dependency cycle"})
	}
	return steps
}

func (r Result) CanOrchestrate() bool { return r.Status == StatusReady }
