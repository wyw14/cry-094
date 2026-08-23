package precheckservice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/precheck"
	"github.com/wyw14/cry-094/internal/domain/team"
)

type IDGenerator interface{ NewID() string }
type Service struct {
	analyses ports.AnalysisRepository
	plans    ports.PlanRepository
	teams    ports.TeamRepository
	signer   ports.Signer
	clock    ports.Clock
	ids      IDGenerator
}

func New(analyses ports.AnalysisRepository, plans ports.PlanRepository, teams ports.TeamRepository, signer ports.Signer, clock ports.Clock, ids IDGenerator) *Service {
	return &Service{analyses: analyses, plans: plans, teams: teams, signer: signer, clock: clock, ids: ids}
}

func (s *Service) Generate(ctx context.Context, analysisID, teamID, actorID string) (*precheck.Plan, error) {
	owner, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if !owner.Can(actorID, team.PermissionRunAnalysis) {
		return nil, fmt.Errorf("analysis permission required")
	}
	result, err := s.analyses.Get(ctx, analysisID)
	if err != nil {
		return nil, err
	}
	if !result.CanOrchestrate() {
		return nil, fmt.Errorf("dependency analysis blocks precheck generation")
	}
	steps := []precheck.Step{{Description: "display operating system identity", Command: "uname -s 2>/dev/null || $PSVersionTable.OS", ReadOnly: true}}
	for _, check := range result.CapabilityChecks {
		description := "inspect " + check.Name + " availability"
		if check.Constraint != "" {
			description += " and verify " + check.Constraint
		}
		steps = append(steps, precheck.Step{Description: description, Command: "command -v " + check.Name + " || Get-Command " + check.Name + " -ErrorAction SilentlyContinue", ReadOnly: true})
	}
	content := make([]string, 0, len(steps))
	for _, step := range steps {
		content = append(content, step.Command)
	}
	sum := sha256.Sum256([]byte(strings.Join(content, "\n")))
	plan, err := precheck.NewPlan(s.ids.NewID(), analysisID, teamID, actorID, hex.EncodeToString(sum[:]), steps, s.clock.Now())
	if err != nil {
		return nil, err
	}
	if err := s.plans.Save(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *Service) Submit(ctx context.Context, planID, actorID string) error {
	plan, err := s.plans.Get(ctx, planID)
	if err != nil {
		return err
	}
	if plan.CreatedBy != actorID {
		return fmt.Errorf("only the author can submit")
	}
	if err := plan.Submit(); err != nil {
		return err
	}
	return s.plans.Save(ctx, plan)
}
func (s *Service) Review(ctx context.Context, planID, reviewerID, reason string, approve bool) error {
	plan, err := s.plans.Get(ctx, planID)
	if err != nil {
		return err
	}
	owner, ownerErr := s.teams.Get(ctx, plan.TeamID)
	if ownerErr == nil && reviewerID == "" {
		return fmt.Errorf("reviewer identity required")
	}
	if ownerErr == nil {
		if _, exists := owner.Members[reviewerID]; !exists {
			return fmt.Errorf("reviewer is not a team member")
		}
	}
	if approve {
		err = plan.Approve(reviewerID, reason, s.clock.Now())
	} else {
		err = plan.Reject(reviewerID, reason, s.clock.Now())
	}
	if err != nil {
		return err
	}
	return s.plans.Save(ctx, plan)
}
func (s *Service) Sign(ctx context.Context, planID, reviewerID string) error {
	plan, err := s.plans.Get(ctx, planID)
	if err != nil {
		return err
	}
	if reviewerID == "" {
		return fmt.Errorf("signer identity required")
	}
	if plan.ReviewedBy != "" && plan.ReviewedBy != reviewerID {
		return fmt.Errorf("signer differs from reviewer")
	}
	if plan.State != precheck.StateApproved && plan.State != precheck.StateInReview {
		return fmt.Errorf("plan is not ready for signing")
	}
	signature, keyID, err := s.signer.Sign(ctx, plan.ContentDigest)
	if err != nil {
		return err
	}
	if err := plan.ApplySignature(signature, keyID, s.clock.Now()); err != nil {
		return err
	}
	return s.plans.Save(ctx, plan)
}
