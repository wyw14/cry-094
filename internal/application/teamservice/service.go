package teamservice

import (
	"context"
	"fmt"

	"github.com/wyw14/cry-094/internal/application/ports"
	"github.com/wyw14/cry-094/internal/domain/audit"
	"github.com/wyw14/cry-094/internal/domain/team"
)

type IDGenerator interface{ NewID() string }

type Service struct {
	teams ports.TeamRepository
	audit ports.AuditRepository
	clock ports.Clock
	ids   IDGenerator
}

func New(teams ports.TeamRepository, audits ports.AuditRepository, clock ports.Clock, ids IDGenerator) *Service {
	return &Service{teams: teams, audit: audits, clock: clock, ids: ids}
}

func (s *Service) Create(ctx context.Context, name, ownerID, source string) (*team.Team, error) {
	now := s.clock.Now()
	value, err := team.New(s.ids.NewID(), name, ownerID, now)
	if err != nil {
		return nil, err
	}
	if err := s.teams.Create(ctx, value); err != nil {
		return nil, err
	}
	event, err := audit.NewEvent(s.ids.NewID(), value.ID, ownerID, source, "team.created", "team", value.ID, "initial team setup", nil, value, now)
	if err != nil {
		return nil, err
	}
	if err := s.audit.Append(ctx, event); err != nil {
		return nil, fmt.Errorf("append team audit: %w", err)
	}
	return value, nil
}

func (s *Service) AddMember(ctx context.Context, teamID, actorID, userID string, role team.Role, source string) error {
	value, err := s.teams.Get(ctx, teamID)
	if err != nil {
		return err
	}
	before := *value
	version := value.Version
	if err := value.AddMember(actorID, userID, role, s.clock.Now()); err != nil {
		return err
	}
	if err := s.teams.Save(ctx, value, version); err != nil {
		return err
	}
	event, err := audit.NewEvent(s.ids.NewID(), teamID, actorID, source, "member.added", "team", teamID, "explicit membership update", before, value, s.clock.Now())
	if err != nil {
		return err
	}
	return s.audit.Append(ctx, event)
}
