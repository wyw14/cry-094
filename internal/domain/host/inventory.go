package host

import (
	"sort"
	"time"

	"github.com/wyw14/cry-094/internal/domain/script"
)

type Tool struct {
	Name    string                 `json:"name"`
	Version script.SemanticVersion `json:"version"`
	Path    string                 `json:"path,omitempty"`
}

type Inventory struct {
	ID         string          `json:"id"`
	TeamID     string          `json:"team_id"`
	TargetName string          `json:"target_name"`
	OS         string          `json:"os"`
	Tools      map[string]Tool `json:"tools"`
	EnvNames   map[string]bool `json:"environment_names"`
	ProbedAt   time.Time       `json:"probed_at"`
	ProbeBuild string          `json:"probe_build"`
	Version    int64           `json:"version"`
}

func (i Inventory) ToolNames() []string {
	names := make([]string, 0, len(i.Tools))
	for name := range i.Tools {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (i Inventory) Satisfies(name string, constraint *script.Constraint) bool {
	tool, ok := i.Tools[name]
	if !ok {
		return false
	}
	return constraint == nil || constraint.SatisfiedBy(tool.Version)
}
