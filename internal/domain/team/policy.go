package team

type Permission string

const (
	PermissionViewLibrary   Permission = "library:view"
	PermissionUploadScript  Permission = "script:upload"
	PermissionRunAnalysis   Permission = "analysis:run"
	PermissionReviewPlan    Permission = "plan:review"
	PermissionManageHost    Permission = "host:manage"
	PermissionManageMembers Permission = "team:members"
)

var rolePermissions = map[Role]map[Permission]bool{
	RoleViewer: {
		PermissionViewLibrary: true,
	},
	RoleAnalyst: {
		PermissionViewLibrary: true, PermissionUploadScript: true, PermissionRunAnalysis: true,
	},
	RoleReviewer: {
		PermissionViewLibrary: true, PermissionUploadScript: true, PermissionRunAnalysis: true,
		PermissionReviewPlan: true,
	},
	RoleAdmin: {
		PermissionViewLibrary: true, PermissionUploadScript: true, PermissionRunAnalysis: true,
		PermissionReviewPlan: true, PermissionManageHost: true, PermissionManageMembers: true,
	},
}

func (t Team) Can(userID string, permission Permission) bool {
	member, ok := t.Members[userID]
	if !ok {
		return false
	}
	return rolePermissions[member.Role][permission]
}

func (t Team) RoleOf(userID string) (Role, bool) {
	member, ok := t.Members[userID]
	return member.Role, ok
}
