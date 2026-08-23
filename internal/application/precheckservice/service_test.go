package precheckservice

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/adapter/local"
	"github.com/wyw14/cry-094/internal/domain/analysis"
	"github.com/wyw14/cry-094/internal/domain/team"
	"github.com/wyw14/cry-094/internal/platform/clock"
	"github.com/wyw14/cry-094/internal/repository/memory"
)

type planIDs struct{ next int }

func (i *planIDs) NewID() string { i.next++; return fmt.Sprintf("plan-%d", i.next) }
func TestGenerateCreatesReadOnlyCapabilityChecks(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	teams := memory.NewTeamRepo()
	owner, _ := team.New("team", "Operations", "author", now)
	if err := owner.AddMember("author", "reviewer", team.RoleReviewer, now); err != nil {
		t.Fatal(err)
	}
	if err := teams.Create(ctx, owner); err != nil {
		t.Fatal(err)
	}
	analyses := memory.NewAnalysisRepo()
	result := &analysis.Result{ID: "analysis", LibraryID: "library", Status: analysis.StatusReady, CapabilityChecks: []analysis.CapabilityCheck{{Name: "jq", Constraint: ">=1.7", Required: true}}}
	if err := analyses.Save(ctx, result); err != nil {
		t.Fatal(err)
	}
	plans := memory.NewPlanRepo()
	signer, _ := local.NewSigner("key", []byte("0123456789abcdef"))
	service := New(analyses, plans, teams, signer, clock.Fixed{Value: now}, &planIDs{})
	plan, err := service.Generate(ctx, "analysis", "team", "author")
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Steps) != 2 || !plan.Steps[1].ReadOnly || !strings.Contains(plan.Steps[1].Description, ">=1.7") {
		t.Fatalf("unexpected plan steps: %#v", plan.Steps)
	}
	if err := service.Submit(ctx, plan.ID, "author"); err != nil {
		t.Fatal(err)
	}
	if err := service.Review(ctx, plan.ID, "reviewer", "commands only inspect capabilities", true); err != nil {
		t.Fatal(err)
	}
	if err := service.Sign(ctx, plan.ID, "reviewer"); err != nil {
		t.Fatal(err)
	}
	stored, err := plans.Get(ctx, plan.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Signature == "" {
		t.Fatal("approved digest must be signed")
	}
}
