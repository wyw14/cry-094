package precheck

import (
	"fmt"
	"time"

	"github.com/wyw14/cry-094/internal/domain/common"
)

type State string

const (
	StateDraft    State = "draft"
	StateInReview State = "in_review"
	StateApproved State = "approved"
	StateRejected State = "rejected"
	StateSigned   State = "signed"
)

type Step struct {
	Description string `json:"description"`
	Command     string `json:"command"`
	ReadOnly    bool   `json:"read_only"`
}

type Plan struct {
	ID            string    `json:"id"`
	AnalysisID    string    `json:"analysis_id"`
	TeamID        string    `json:"team_id"`
	State         State     `json:"state"`
	Steps         []Step    `json:"steps"`
	CreatedBy     string    `json:"created_by"`
	ReviewedBy    string    `json:"reviewed_by,omitempty"`
	ReviewReason  string    `json:"review_reason,omitempty"`
	ContentDigest string    `json:"content_digest"`
	Signature     string    `json:"signature,omitempty"`
	SignerKeyID   string    `json:"signer_key_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	ReviewedAt    time.Time `json:"reviewed_at,omitempty"`
	SignedAt      time.Time `json:"signed_at,omitempty"`
	Version       int64     `json:"version"`
}

func NewPlan(id, analysisID, teamID, actorID, digest string, steps []Step, now time.Time) (*Plan, error) {
	if id == "" || analysisID == "" || teamID == "" || actorID == "" || digest == "" || len(steps) == 0 {
		return nil, common.NewCoded("PLAN_INVALID", "plan identity, owner, digest and steps are required", nil)
	}
	for _, step := range steps {
		if !step.ReadOnly {
			return nil, common.NewCoded("PLAN_MUTATING", "precheck steps must be read-only", nil)
		}
	}
	return &Plan{ID: id, AnalysisID: analysisID, TeamID: teamID, State: StateDraft,
		Steps: append([]Step(nil), steps...), CreatedBy: actorID, ContentDigest: digest,
		CreatedAt: now.UTC(), Version: 1}, nil
}

func (p *Plan) Submit() error {
	if p.State != StateDraft && p.State != StateRejected {
		return common.ErrInvalidTransition
	}
	p.State = StateInReview
	p.Version++
	return nil
}

func (p *Plan) Approve(reviewerID, reason string, now time.Time) error {
	if p.State != StateInReview {
		return common.ErrInvalidTransition
	}
	if reviewerID == "" || reviewerID == p.CreatedBy || reason == "" {
		return common.NewCoded("REVIEW_INVALID", "review requires a different reviewer and a reason", nil)
	}
	p.State = StateApproved
	p.ReviewedBy = reviewerID
	p.ReviewReason = reason
	p.ReviewedAt = now.UTC()
	p.Version++
	return nil
}

func (p *Plan) Reject(reviewerID, reason string, now time.Time) error {
	if p.State != StateInReview {
		return common.ErrInvalidTransition
	}
	if reviewerID == "" || reason == "" {
		return fmt.Errorf("reviewer and rejection reason are required")
	}
	p.State = StateRejected
	p.ReviewedBy = reviewerID
	p.ReviewReason = reason
	p.ReviewedAt = now.UTC()
	p.Version++
	return nil
}

func (p *Plan) ApplySignature(signature, keyID string, now time.Time) error {
	if p.State != StateApproved {
		return common.ErrInvalidTransition
	}
	if signature == "" || keyID == "" {
		return common.NewCoded("SIGNATURE_INVALID", "signature and key id are required", nil)
	}
	p.State = StateSigned
	p.Signature = signature
	p.SignerKeyID = keyID
	p.SignedAt = now.UTC()
	p.Version++
	return nil
}
