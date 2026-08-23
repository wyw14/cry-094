package audit

import (
	"encoding/json"
	"time"
)

type Event struct {
	ID           string          `json:"id"`
	TeamID       string          `json:"team_id"`
	ActorID      string          `json:"actor_id"`
	Source       string          `json:"source"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	Before       json.RawMessage `json:"before"`
	After        json.RawMessage `json:"after"`
	Reason       string          `json:"reason"`
	OccurredAt   time.Time       `json:"occurred_at"`
}

func NewEvent(id, teamID, actorID, source, action, resourceType, resourceID, reason string, before, after any, now time.Time) (Event, error) {
	beforeJSON, err := json.Marshal(before)
	if err != nil {
		return Event{}, err
	}
	afterJSON, err := json.Marshal(after)
	if err != nil {
		return Event{}, err
	}
	return Event{ID: id, TeamID: teamID, ActorID: actorID, Source: source, Action: action,
		ResourceType: resourceType, ResourceID: resourceID, Before: beforeJSON, After: afterJSON,
		Reason: reason, OccurredAt: now.UTC()}, nil
}
