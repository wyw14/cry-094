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
	CacheGeneration  uint64            `json:"cache_generation"`
	CachePublished   bool              `json:"cache_published"`
	InvalidatedAt    *time.Time        `json:"invalidated_at,omitempty"`
}

type CapabilityCheck struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
	Required   bool   `json:"required"`
}

func NewResult(id, libraryID string, graph Graph, issues []Issue, artifactHash, parserBuild string, now time.Time) Result {
	order, cycle := graph.TopologicalOrder()
	if len(cycle) > 0 {
		issues = append(issues, Issue{Kind: IssueCycle, Severity: SeverityBlocking,
			Summary:  "script dependency graph contains a cycle",
			Evidence: cycleEvidence(cycle), Resolution: "break the reported data dependency cycle"})
	}
	status := StatusReady
	for _, issue := range issues {
		if issue.BlocksOrchestration() {
			status = StatusBlocked
			break
		}
	}
	return Result{ID: id, LibraryID: libraryID, Graph: graph, Order: order, Issues: issues,
		Status: status, ArtifactHash: artifactHash, ParserBuild: parserBuild, CompletedAt: now.UTC()}
}

func cycleEvidence(ids []string) []EvidenceStep {
	steps := make([]EvidenceStep, 0, len(ids))
	for _, id := range ids {
		steps = append(steps, EvidenceStep{ArtifactID: id, Detail: "participates in dependency cycle"})
	}
	return steps
}

func (r Result) CanOrchestrate() bool { return r.Status == StatusReady }

func (r *Result) BindCacheGeneration(generation uint64) {
	r.CacheGeneration = generation
	r.CachePublished = false
	r.InvalidatedAt = nil
}

func (r Result) BelongsToCacheGeneration(generation uint64) bool {
	return r.CacheGeneration <= generation
}

func (r Result) CanPublishAt(generation uint64) bool {
	if r.InvalidatedAt != nil {
		return false
	}
	return r.BelongsToCacheGeneration(generation)
}

func (r *Result) AcceptCachePublication(generation uint64) {
	r.CacheGeneration = generation
	r.CachePublished = true
	r.InvalidatedAt = nil
}

func (r *Result) MarkCacheInvalidated(now time.Time) {
	value := now.UTC()
	r.InvalidatedAt = &value
	r.CachePublished = false
}
