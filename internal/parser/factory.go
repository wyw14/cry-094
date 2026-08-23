package parser

import (
	"fmt"
	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/parser/shell"
)

type Factory struct{ BuildID string }

func (f Factory) For(kind script.Shell) (ports.Parser, error) {
	switch kind {
	case script.ShellPOSIX, script.ShellPowerShell:
		return shell.New(kind, f.BuildID), nil
	default:
		return nil, fmt.Errorf("unsupported shell %s", kind)
	}
}
