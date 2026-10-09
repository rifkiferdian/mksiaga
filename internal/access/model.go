package access

type Role struct {
	ID              uint64
	Name            string
	GuardName       string
	Description     string
	PermissionCount int
	UserCount       int
}

type Permission struct {
	ID          uint64
	Name        string
	GuardName   string
	Description string
	RoleCount   int
	DirectCount int
	Assigned    bool
}

type PermissionGroup struct {
	Key           string
	Label         string
	Description   string
	Permissions   []Permission
	AssignedCount int
}
