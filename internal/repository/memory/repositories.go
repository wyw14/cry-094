package memory

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/wyw14/cry-094/internal/domain/analysis"
	"github.com/wyw14/cry-094/internal/domain/audit"
	"github.com/wyw14/cry-094/internal/domain/precheck"
	"github.com/wyw14/cry-094/internal/domain/script"
	"github.com/wyw14/cry-094/internal/domain/team"
)

type TeamRepo struct {
	mu     sync.RWMutex
	values map[string]*team.Team
}

func NewTeamRepo() *TeamRepo { return &TeamRepo{values: make(map[string]*team.Team)} }
func (r *TeamRepo) Create(_ context.Context, value *team.Team) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.values[value.ID]; ok {
		return fmt.Errorf("team exists")
	}
	r.values[value.ID] = cloneTeam(value)
	return nil
}
func (r *TeamRepo) Get(_ context.Context, id string) (*team.Team, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.values[id]
	if !ok {
		return nil, fmt.Errorf("team not found")
	}
	return cloneTeam(v), nil
}
func (r *TeamRepo) Save(_ context.Context, value *team.Team, version int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	old, ok := r.values[value.ID]
	if !ok {
		return fmt.Errorf("team not found")
	}
	if old.Version != version {
		return fmt.Errorf("version conflict")
	}
	r.values[value.ID] = cloneTeam(value)
	return nil
}
func cloneTeam(v *team.Team) *team.Team {
	c := *v
	c.Members = make(map[string]team.Member, len(v.Members))
	for id, m := range v.Members {
		c.Members[id] = m
	}
	return &c
}

type ArtifactRepo struct {
	mu        sync.RWMutex
	libraries map[string]*script.Library
	artifacts map[string]*script.Artifact
}

func NewArtifactRepo() *ArtifactRepo {
	return &ArtifactRepo{libraries: make(map[string]*script.Library), artifacts: make(map[string]*script.Artifact)}
}
func (r *ArtifactRepo) CreateLibrary(_ context.Context, value *script.Library) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.libraries[value.ID]; ok {
		return fmt.Errorf("library exists")
	}
	r.libraries[value.ID] = cloneLibrary(value)
	return nil
}
func (r *ArtifactRepo) GetLibrary(_ context.Context, id string) (*script.Library, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	value, ok := r.libraries[id]
	if !ok {
		return nil, fmt.Errorf("library not found")
	}
	return cloneLibrary(value), nil
}
func (r *ArtifactRepo) CreateArtifact(_ context.Context, value *script.Artifact) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.artifacts[value.ID]; ok {
		return fmt.Errorf("artifact exists")
	}
	r.artifacts[value.ID] = cloneArtifact(value)
	return nil
}
func (r *ArtifactRepo) GetArtifact(_ context.Context, id string) (*script.Artifact, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.artifacts[id]
	if !ok {
		return nil, fmt.Errorf("artifact not found")
	}
	return cloneArtifact(v), nil
}
func (r *ArtifactRepo) ListArtifacts(_ context.Context, libraryID, cursor string, limit int) ([]script.Artifact, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	values := make([]script.Artifact, 0)
	for _, v := range r.artifacts {
		if libraryID != "" && v.LibraryID != libraryID {
			continue
		}
		values = append(values, *cloneArtifact(v))
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].LibraryID == values[j].LibraryID {
			return values[i].Number < values[j].Number
		}
		return values[i].LibraryID < values[j].LibraryID
	})
	start := 0
	for i := range values {
		if values[i].ID == cursor {
			start = i + 1
			break
		}
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	end := start + limit
	if end > len(values) {
		end = len(values)
	}
	next := ""
	if end < len(values) {
		next = values[end-1].ID
	}
	page := make([]script.Artifact, end-start)
	copy(page, values[start:end])
	return page, next, nil
}
func cloneLibrary(v *script.Library) *script.Library {
	c := *v
	c.Tags = map[string]string{}
	for k, val := range v.Tags {
		c.Tags[k] = val
	}
	return &c
}
func cloneArtifact(v *script.Artifact) *script.Artifact { c := *v; return &c }

type AnalysisRepo struct {
	mu     sync.RWMutex
	values map[string]*analysis.Result
}

func NewAnalysisRepo() *AnalysisRepo { return &AnalysisRepo{values: make(map[string]*analysis.Result)} }
func (r *AnalysisRepo) Save(_ context.Context, value *analysis.Result) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *value
	r.values[value.ID] = &c
	return nil
}
func (r *AnalysisRepo) Get(_ context.Context, id string) (*analysis.Result, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.values[id]
	if !ok {
		return nil, fmt.Errorf("analysis not found")
	}
	c := *v
	return &c, nil
}

type PlanRepo struct {
	mu     sync.RWMutex
	values map[string]*precheck.Plan
}

func NewPlanRepo() *PlanRepo { return &PlanRepo{values: make(map[string]*precheck.Plan)} }
func (r *PlanRepo) Save(_ context.Context, value *precheck.Plan) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	c := *value
	c.Steps = append([]precheck.Step(nil), value.Steps...)
	r.values[value.ID] = &c
	return nil
}
func (r *PlanRepo) Get(_ context.Context, id string) (*precheck.Plan, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	v, ok := r.values[id]
	if !ok {
		return nil, fmt.Errorf("plan not found")
	}
	c := *v
	c.Steps = append([]precheck.Step(nil), v.Steps...)
	return &c, nil
}

type AuditRepo struct {
	mu     sync.RWMutex
	values []audit.Event
}

func NewAuditRepo() *AuditRepo { return &AuditRepo{} }
func (r *AuditRepo) Append(_ context.Context, value audit.Event) error {
	r.mu.Lock()
	r.values = append(r.values, value)
	r.mu.Unlock()
	return nil
}
func (r *AuditRepo) List() []audit.Event {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return append([]audit.Event(nil), r.values...)
}
