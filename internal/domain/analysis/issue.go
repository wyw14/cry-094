package analysis

type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityBlocking Severity = "blocking"
)

type IssueKind string

const (
	IssueCycle           IssueKind = "cycle"
	IssueMissing         IssueKind = "missing_dependency"
	IssueVersionConflict IssueKind = "version_conflict"
	IssueUnsafeCommand   IssueKind = "unsafe_command"
)

type EvidenceStep struct {
	ArtifactID string `json:"artifact_id"`
	Dependency string `json:"dependency"`
	Detail     string `json:"detail"`
	Line       int    `json:"line,omitempty"`
}

type Issue struct {
	Kind       IssueKind      `json:"kind"`
	Severity   Severity       `json:"severity"`
	Summary    string         `json:"summary"`
	Evidence   []EvidenceStep `json:"evidence"`
	Resolution string         `json:"resolution"`
}

func (i Issue) BlocksOrchestration() bool { return i.Severity == SeverityBlocking }
