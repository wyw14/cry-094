package team

import (
	"errors"
	"testing"
	"time"

	"github.com/wyw14/cry-094/internal/domain/common"
)

func TestTeamPermissionsAndLastAdmin(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	value, err := New("team-1", "Operations", "owner", now)
	if err != nil {
		t.Fatal(err)
	}
	if !value.Can("owner", PermissionManageMembers) {
		t.Fatal("owner should be administrator")
	}
	if err := value.AddMember("owner", "reviewer", RoleReviewer, now); err != nil {
		t.Fatal(err)
	}
	if value.Can("reviewer", PermissionManageMembers) {
		t.Fatal("reviewer must not manage members")
	}
	err = value.ChangeRole("owner", "owner", RoleAnalyst)
	if err == nil {
		t.Fatal("last administrator demotion must fail")
	}
	var coded *common.CodedError
	if !errors.As(err, &coded) {
		t.Fatalf("expected coded error, got %T", err)
	}
}
