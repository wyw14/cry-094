package precheck

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/common"
)

func TestPlanRequiresIndependentReviewBeforeSignature(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	plan, err := NewPlan("p", "a", "t", "author", "digest", []Step{{Description: "jq", Command: "command -v jq", ReadOnly: true}}, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := plan.ApplySignature("sig", "key", now); !errors.Is(err, common.ErrInvalidTransition) {
		t.Fatalf("expected transition error, got %v", err)
	}
	if err := plan.Submit(); err != nil {
		t.Fatal(err)
	}
	if err := plan.Approve("author", "looks safe", now); err == nil {
		t.Fatal("author must not self-approve")
	}
	if err := plan.Approve("reviewer", "commands are read-only", now); err != nil {
		t.Fatal(err)
	}
	if err := plan.ApplySignature("sig", "key", now); err != nil {
		t.Fatal(err)
	}
	if plan.State != StateSigned {
		t.Fatalf("got %s", plan.State)
	}
}
func TestPlanRejectsInstallerStep(t *testing.T) {
	_, err := NewPlan("p", "a", "t", "author", "digest", []Step{{Description: "install jq", Command: "apt install jq", ReadOnly: false}}, time.Now())
	if err == nil {
		t.Fatal("mutating plan must be rejected")
	}
}
