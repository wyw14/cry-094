package shell

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/wyw14/cry-094/internal/domain/script"
)

var commandPattern = regexp.MustCompile(`(?m)(?:^|[;&|])\s*(jq|yq|ansible(?:-playbook)?|bash|pwsh|powershell)\b`)
var envPattern = regexp.MustCompile(`\$\{?([A-Z][A-Z0-9_]*)\}?`)
var psEnvPattern = regexp.MustCompile(`\$env:([A-Za-z_][A-Za-z0-9_]*)`)
var requirementPattern = regexp.MustCompile(`(?i)^\s*#\s*requires:\s*([a-z][a-z0-9-]*)\s*([<>=]{1,2}\s*[0-9]+(?:\.[0-9]+){0,2})\s*$`)
var inputPattern = regexp.MustCompile(`<\s*([A-Za-z0-9_.-]+)`)
var outputPattern = regexp.MustCompile(`>\s*([A-Za-z0-9_.-]+)`)
var sourcePattern = regexp.MustCompile(`(?:^|;)\s*(?:source|\.)\s+([A-Za-z0-9_./-]+)`)

type Parser struct {
	shell script.Shell
	build string
}

func New(shellType script.Shell, build string) Parser { return Parser{shell: shellType, build: build} }
func (p Parser) Name() string                         { return string(p.shell) }
func (p Parser) Build() string                        { return p.build }
func (p Parser) Parse(ctx context.Context, artifact script.Artifact, content []byte) (script.ParseConclusion, error) {
	detached, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	return p.parseDetached(detached, artifact, content)
}

func (p Parser) parseDetached(ctx context.Context, artifact script.Artifact, content []byte) (script.ParseConclusion, error) {
	if int64(len(content)) > 1024*1024 {
		return script.ParseConclusion{}, fmt.Errorf("script exceeds parser limit")
	}
	conclusion := script.ParseConclusion{ArtifactID: artifact.ID, Digest: artifact.Digest, ParserName: p.Name(), ParserBuild: p.build}
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	scanner.Buffer(make([]byte, 1024), 1024*1024)
	line := 0
	depth := 0
	for scanner.Scan() {
		select {
		case <-ctx.Done():
			return script.ParseConclusion{}, ctx.Err()
		default:
		}
		line++
		text := scanner.Text()
		depth += strings.Count(text, "{") + strings.Count(text, "(") - strings.Count(text, "}") - strings.Count(text, ")")
		if depth > 64 {
			return script.ParseConclusion{}, fmt.Errorf("script nesting exceeds parser depth limit")
		}
		if depth < 0 {
			depth = 0
		}
		if match := requirementPattern.FindStringSubmatch(text); len(match) == 3 {
			constraint, err := script.ParseConstraint(strings.ReplaceAll(match[2], " ", ""))
			if err != nil {
				return script.ParseConclusion{}, fmt.Errorf("parse requirement at line %d: %w", line, err)
			}
			conclusion.Dependencies = append(conclusion.Dependencies, script.Dependency{Kind: script.DependencyTool, Name: strings.ToLower(match[1]), Constraint: &constraint, Location: script.Location{Line: line}, Evidence: strings.TrimSpace(text)})
			continue
		}
		matches := commandPattern.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			name := match[1]
			kind := script.DependencyTool
			evidence := strings.TrimSpace(text)
			conclusion.Dependencies = append(conclusion.Dependencies, script.Dependency{Kind: kind, Name: name, Location: script.Location{Line: line}, Evidence: evidence})
		}
		for _, match := range envPattern.FindAllStringSubmatch(text, -1) {
			conclusion.Dependencies = append(conclusion.Dependencies, script.Dependency{Kind: script.DependencyEnvironment, Name: match[1], Location: script.Location{Line: line}, Evidence: strings.TrimSpace(text)})
		}
		for _, match := range psEnvPattern.FindAllStringSubmatch(text, -1) {
			conclusion.Dependencies = append(conclusion.Dependencies, script.Dependency{Kind: script.DependencyEnvironment, Name: strings.ToUpper(match[1]), Location: script.Location{Line: line}, Evidence: strings.TrimSpace(text)})
		}
		for _, match := range sourcePattern.FindAllStringSubmatch(text, -1) {
			conclusion.Dependencies = append(conclusion.Dependencies, script.Dependency{Kind: script.DependencyLibrary, Name: match[1], Implicit: true, Location: script.Location{Line: line}, Evidence: strings.TrimSpace(text)})
		}
		if strings.Contains(text, "curl ") || strings.Contains(text, "rm -rf") {
			conclusion.Warnings = append(conclusion.Warnings, fmt.Sprintf("unsafe command at line %d", line))
		}
		for _, match := range inputPattern.FindAllStringSubmatch(text, -1) {
			conclusion.Contract.Inputs = append(conclusion.Contract.Inputs, match[1])
		}
		for _, match := range outputPattern.FindAllStringSubmatch(text, -1) {
			conclusion.Contract.Outputs = append(conclusion.Contract.Outputs, match[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return script.ParseConclusion{}, err
	}
	conclusion.Dependencies = mergeDependencies(conclusion.Dependencies)
	return conclusion, nil
}

func mergeDependencies(values []script.Dependency) []script.Dependency {
	result := make([]script.Dependency, 0, len(values))
	toolIndex := map[string]int{}
	seen := map[string]bool{}
	for _, value := range values {
		if value.Kind == script.DependencyTool {
			if index, ok := toolIndex[value.Name]; ok {
				if result[index].Constraint == nil && value.Constraint != nil {
					result[index] = value
				}
				continue
			}
			toolIndex[value.Name] = len(result)
			result = append(result, value)
			continue
		}
		key := string(value.Kind) + ":" + value.Name
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, value)
	}
	return result
}
