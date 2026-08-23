package team

import (
	"fmt"
	"strings"
	"time"

	"github.com/wyw14/cry-094/internal/domain/common"
)

type Role string

const (
	RoleViewer   Role = "viewer"
	RoleAnalyst  Role = "analyst"
	RoleReviewer Role = "reviewer"
	RoleAdmin    Role = "admin"
)

type Member struct {
	UserID string    `json:"user_id"`
	Role   Role      `json:"role"`
	Joined time.Time `json:"joined_at"`
}

type Team struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Members   map[string]Member `json:"members"`
	CreatedAt time.Time         `json:"created_at"`
	Version   int64             `json:"version"`
}

func New(id, name, ownerID string, now time.Time) (*Team, error) {
	name = strings.TrimSpace(name)
	if id == "" || ownerID == "" || len(name) < 3 {
		return nil, common.NewCoded("TEAM_INVALID", "team id, owner and a descriptive name are required", nil)
	}
	return &Team{
		ID: id, Name: name, CreatedAt: now.UTC(), Version: 1,
		Members: map[string]Member{ownerID: {UserID: ownerID, Role: RoleAdmin, Joined: now.UTC()}},
	}, nil
}

func (t *Team) AddMember(actorID, userID string, role Role, now time.Time) error {
	if !t.Can(actorID, PermissionManageMembers) {
		return common.ErrForbidden
	}
	if userID == "" || !role.Valid() {
		return common.NewCoded("MEMBER_INVALID", "member and role are required", nil)
	}
	if _, exists := t.Members[userID]; exists {
		return fmt.Errorf("%w: member already exists", common.ErrConflict)
	}
	t.Members[userID] = Member{UserID: userID, Role: role, Joined: now.UTC()}
	t.Version++
	return nil
}

func (t *Team) ChangeRole(actorID, userID string, role Role) error {
	if !t.Can(actorID, PermissionManageMembers) {
		return common.ErrForbidden
	}
	member, ok := t.Members[userID]
	if !ok {
		return common.ErrNotFound
	}
	if !role.Valid() {
		return common.NewCoded("ROLE_INVALID", "unknown role", nil)
	}
	if member.Role == RoleAdmin && role != RoleAdmin && t.adminCount() == 1 {
		return common.NewCoded("LAST_ADMIN", "a team must keep at least one administrator", nil)
	}
	member.Role = role
	t.Members[userID] = member
	t.Version++
	return nil
}

func (t *Team) adminCount() int {
	count := 0
	for _, member := range t.Members {
		if member.Role == RoleAdmin {
			count++
		}
	}
	return count
}

func (r Role) Valid() bool {
	switch r {
	case RoleViewer, RoleAnalyst, RoleReviewer, RoleAdmin:
		return true
	default:
		return false
	}
}
