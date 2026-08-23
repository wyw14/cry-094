package analysisservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"

	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/analysis"
	"github.com/wyw14/cry-094/internal/domain/host"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
)

type ParserFactory interface {
	For(script.Shell) (ports.Parser, error)
}
type IDGenerator interface{ NewID() string }

type Service struct {
	teams     ports.TeamRepository
	artifacts ports.ArtifactRepository
	results   ports.AnalysisRepository
	store     ports.ObjectStore
	parsers   ParserFactory
	clock     ports.Clock
	ids       IDGenerator
}

func New(teams ports.TeamRepository, artifacts ports.ArtifactRepository, results ports.AnalysisRepository, store ports.ObjectStore, parsers ParserFactory, clock ports.Clock, ids IDGenerator) *Service {
	return &Service{teams: teams, artifacts: artifacts, results: results, store: store, parsers: parsers, clock: clock, ids: ids}
}

func (s *Service) Analyze(ctx context.Context, teamID, actorID, libraryID string, artifactIDs []string, inventory host.Inventory) (*analysis.Result, error) {
	owner, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if !owner.Can(actorID, team.PermissionRunAnalysis) {
		return nil, fmt.Errorf("analysis permission required")
	}
	library, err := s.artifacts.GetLibrary(ctx, libraryID)
	if err != nil {
		return nil, err
	}
	if library.TeamID != teamID {
		return nil, fmt.Errorf("library is outside actor team")
	}
	conclusions := make([]script.ParseConclusion, 0, len(artifactIDs))
	artifacts := make([]script.Artifact, 0, len(artifactIDs))
	for _, id := range artifactIDs {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		artifact, err := s.artifacts.GetArtifact(ctx, id)
		if err != nil {
			return nil, err
		}
		if artifact.LibraryID != libraryID {
			return nil, fmt.Errorf("artifact is outside requested library")
		}
		content, err := s.store.Get(ctx, artifact.StorageKey)
		if err != nil {
			return nil, err
		}
		if filesDigest(content) != artifact.Digest {
			return nil, fmt.Errorf("artifact digest changed in read-only storage")
		}
		parser, err := s.parsers.For(artifact.Shell)
		if err != nil {
			return nil, err
		}
		conclusion, err := parser.Parse(ctx, *artifact, content)
		if err != nil {
			return nil, err
		}
		if !conclusion.BoundTo(*artifact) {
			return nil, fmt.Errorf("parser conclusion is not bound to artifact digest")
		}
		artifact.ParserBuild = conclusion.ParserBuild
		artifacts = append(artifacts, *artifact)
		conclusions = append(conclusions, conclusion)
	}
	graph, issues, checks, err := buildGraph(artifacts, conclusions, inventory)
	if err != nil {
		return nil, err
	}
	result := analysis.NewResult(s.ids.NewID(), libraryID, graph, issues, aggregateDigest(artifacts), parserBuilds(conclusions), s.clock.Now())
	result.CapabilityChecks = checks
	if err := s.results.Save(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func buildGraph(artifacts []script.Artifact, conclusions []script.ParseConclusion, inventory host.Inventory) (analysis.Graph, []analysis.Issue, []analysis.CapabilityCheck, error) {
	nodes := make([]analysis.Node, 0, len(artifacts))
	outputOwners := map[string]string{}
	for index, artifact := range artifacts {
		nodes = append(nodes, analysis.Node{ArtifactID: artifact.ID, Digest: artifact.Digest})
		for _, out := range conclusions[index].Contract.Outputs {
			outputOwners[out] = artifact.ID
		}
	}
	edges := []analysis.Edge{}
	issues := []analysis.Issue{}
	checksByKey := map[string]analysis.CapabilityCheck{}
	for index, conclusion := range conclusions {
		artifact := artifacts[index]
		for _, input := range conclusion.Contract.Inputs {
			if owner, ok := outputOwners[input]; ok && owner != artifact.ID {
				edges = append(edges, analysis.Edge{From: owner, To: artifact.ID, Contract: input, Evidence: "producer output satisfies consumer input"})
			} else {
				issues = append(issues, analysis.Issue{Kind: analysis.IssueMissing, Severity: analysis.SeverityBlocking, Summary: "script input has no producer", Evidence: []analysis.EvidenceStep{{ArtifactID: artifact.ID, Dependency: input, Detail: "no artifact produces this input"}}, Resolution: "declare or upload a producing script"})
			}
		}
		for _, dependency := range conclusion.Dependencies {
			switch dependency.Kind {
			case script.DependencyTool:
				constraint := ""
				if dependency.Constraint != nil {
					constraint = dependency.Constraint.Operator + dependency.Constraint.Version.String()
				}
				checksByKey[dependency.Name+constraint] = analysis.CapabilityCheck{Name: dependency.Name, Constraint: constraint, Required: !dependency.Optional}
				tool, exists := inventory.Tools[dependency.Name]
				if !exists {
					issues = append(issues, analysis.Issue{Kind: analysis.IssueMissing, Severity: analysis.SeverityBlocking, Summary: "target host lacks a required tool", Evidence: []analysis.EvidenceStep{{ArtifactID: artifact.ID, Dependency: dependency.Name, Detail: dependency.Evidence, Line: dependency.Location.Line}}, Resolution: "select a host that declares the required tool"})
				} else if dependency.Constraint != nil && !dependency.Constraint.SatisfiedBy(tool.Version) {
					issues = append(issues, analysis.Issue{
						Kind:     analysis.IssueVersionConflict,
						Severity: analysis.SeverityBlocking,
						Summary:  fmt.Sprintf("target host %s does not satisfy required constraint %s", tool.Version.String(), constraint),
						Evidence: []analysis.EvidenceStep{{
							ArtifactID: artifact.ID,
							Dependency: dependency.Name,
							Detail:     dependency.Evidence + " (required " + constraint + ", host " + tool.Version.String() + ")",
							Line:       dependency.Location.Line,
						}},
						Resolution: "select a host whose " + dependency.Name + " satisfies " + constraint,
					})
				}
			case script.DependencyEnvironment:
				if !inventory.EnvNames[dependency.Name] {
					issues = append(issues, analysis.Issue{Kind: analysis.IssueMissing, Severity: analysis.SeverityWarning, Summary: "target inventory lacks a declared environment variable", Evidence: []analysis.EvidenceStep{{ArtifactID: artifact.ID, Dependency: dependency.Name, Detail: dependency.Evidence, Line: dependency.Location.Line}}, Resolution: "provide the variable through approved host configuration"})
				}
			}
		}
		for _, warning := range conclusion.Warnings {
			issues = append(issues, analysis.Issue{Kind: analysis.IssueUnsafeCommand, Severity: analysis.SeverityBlocking, Summary: "uploaded script contains an unsafe command", Evidence: []analysis.EvidenceStep{{ArtifactID: artifact.ID, Detail: warning}}, Resolution: "replace the command and upload a new immutable version"})
		}
	}
	checks := make([]analysis.CapabilityCheck, 0, len(checksByKey))
	for _, check := range checksByKey {
		checks = append(checks, check)
	}
	sort.Slice(checks, func(i, j int) bool {
		if checks[i].Name == checks[j].Name {
			return checks[i].Constraint < checks[j].Constraint
		}
		return checks[i].Name < checks[j].Name
	})
	graph, err := analysis.NewGraph(nodes, edges)
	return graph, issues, checks, err
}

func filesDigest(content []byte) string {
	sum := sha256.Sum256(content)
	return hex.EncodeToString(sum[:])
}
func aggregateDigest(artifacts []script.Artifact) string {
	values := make([]string, 0, len(artifacts))
	for _, a := range artifacts {
		values = append(values, a.Digest)
	}
	sort.Strings(values)
	sum := sha256.Sum256([]byte(strings.Join(values, ":")))
	return hex.EncodeToString(sum[:])
}
func parserBuilds(values []script.ParseConclusion) string {
	builds := make([]string, 0, len(values))
	for _, v := range values {
		builds = append(builds, v.ParserBuild)
	}
	sort.Strings(builds)
	return strings.Join(builds, ",")
}
