package shell

import (
	"context"
	"strings"
	"testing"

	"github.com/wyw14/cry-094/internal/domain/script"
)

func TestParserExtractsCommandsEnvironmentAndUnsafeEvidence(t *testing.T) {
	content := []byte("#!/bin/sh\njq '.items' \"$INPUT_FILE\" > output.json\ncurl https://example.invalid/run\n")
	artifact := script.Artifact{ID: "a", Digest: "digest", Shell: script.ShellPOSIX}
	result, err := New(script.ShellPOSIX, "build-v1").Parse(context.Background(), artifact, content)
	if err != nil {
		t.Fatal(err)
	}
	if result.ParserBuild != "build-v1" || len(result.Dependencies) != 2 {
		t.Fatalf("unexpected conclusion: %#v", result)
	}
	if len(result.Warnings) != 1 || !strings.Contains(result.Warnings[0], "line 3") {
		t.Fatalf("unsafe evidence missing: %#v", result.Warnings)
	}
}
func TestParserHonorsCancellationAndSizeLimit(t *testing.T) {
	parser := New(script.ShellPOSIX, "v1")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := parser.Parse(ctx, script.Artifact{}, []byte("jq .")); err == nil {
		t.Fatal("cancelled context must stop parser")
	}
	if _, err := parser.Parse(context.Background(), script.Artifact{}, make([]byte, 1024*1024+1)); err == nil {
		t.Fatal("oversized script must fail")
	}
}
