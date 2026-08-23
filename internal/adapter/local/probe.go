package local

import (
	"context"
	"github.com/wyw14/cry-094/internal/domain/host"
	"time"
)

type Probe struct {
	Tools       map[string]host.Tool
	Environment []string
	Build       string
	Now         func() time.Time
}

func (p Probe) Probe(ctx context.Context, value host.Inventory) (host.Inventory, error) {
	if err := ctx.Err(); err != nil {
		return host.Inventory{}, err
	}
	value.Tools = make(map[string]host.Tool, len(p.Tools))
	for name, tool := range p.Tools {
		value.Tools[name] = tool
	}
	value.EnvNames = make(map[string]bool, len(p.Environment))
	for _, name := range p.Environment {
		value.EnvNames[name] = true
	}
	value.ProbeBuild = p.Build
	value.ProbedAt = p.Now().UTC()
	value.Version++
	return value, nil
}
