package ports

import (
	"context"
	"time"

	"github.com/wyw14/cry-094/internal/domain/analysis"
	"github.com/wyw14/cry-094/internal/domain/audit"
	"github.com/wyw14/cry-094/internal/domain/host"
	"github.com/wyw14/cry-094/internal/domain/precheck"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
)

type TeamRepository interface {
	Create(context.Context, *team.Team) error
	Get(context.Context, string) (*team.Team, error)
	Save(context.Context, *team.Team, int64) error
}

type ArtifactRepository interface {
	CreateLibrary(context.Context, *script.Library) error
	GetLibrary(context.Context, string) (*script.Library, error)
	CreateArtifact(context.Context, *script.Artifact) error
	GetArtifact(context.Context, string) (*script.Artifact, error)
	ListArtifacts(context.Context, string, string, int) ([]script.Artifact, string, error)
}

type AnalysisRepository interface {
	Save(context.Context, *analysis.Result) error
	Get(context.Context, string) (*analysis.Result, error)
}

type PlanRepository interface {
	Save(context.Context, *precheck.Plan) error
	Get(context.Context, string) (*precheck.Plan, error)
}

type AuditRepository interface {
	Append(context.Context, audit.Event) error
}

type Parser interface {
	Name() string
	Build() string
	Parse(context.Context, script.Artifact, []byte) (script.ParseConclusion, error)
}

type ObjectStore interface {
	Put(context.Context, string, []byte) error
	Get(context.Context, string) ([]byte, error)
}

type HostProbe interface {
	Probe(context.Context, host.Inventory) (host.Inventory, error)
}

type Signer interface {
	Sign(context.Context, string) (signature, keyID string, err error)
}

type Clock interface{ Now() time.Time }
