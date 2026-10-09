package access

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
)

var (
	ErrInvalidName = errors.New("invalid name")
	ErrDuplicate   = errors.New("duplicate name")
	ErrProtected   = errors.New("protected record")
	ErrInUse       = errors.New("record in use")
)

var accessNamePattern = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9._-]{0,123}[a-z0-9])?$`)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) Roles(ctx context.Context) ([]Role, error) { return s.repository.Roles(ctx) }
func (s *Service) Permissions(ctx context.Context) ([]Permission, error) {
	return s.repository.Permissions(ctx)
}
func (s *Service) Role(ctx context.Context, id uint64) (Role, []Permission, error) {
	role, err := s.repository.Role(ctx, id)
	if err != nil {
		return Role{}, nil, err
	}
	permissions, err := s.repository.PermissionsForRole(ctx, id)
	return role, permissions, err
}

func (s *Service) CreateRole(ctx context.Context, name, guard, description string) error {
	role, err := normalizedRole(0, name, guard, description)
	if err != nil {
		return err
	}
	return translateDuplicate(s.repository.CreateRole(ctx, role))
}

func (s *Service) UpdateRole(ctx context.Context, id uint64, name, guard, description string) error {
	role, err := normalizedRole(id, name, guard, description)
	if err != nil {
		return err
	}
	existing, err := s.repository.Role(ctx, id)
	if err != nil {
		return err
	}
	if existing.Name == "superadmin" && (role.Name != "superadmin" || role.GuardName != "web") {
		return ErrProtected
	}
	return translateDuplicate(s.repository.UpdateRole(ctx, role))
}

func (s *Service) DeleteRole(ctx context.Context, id uint64) error {
	role, err := s.repository.Role(ctx, id)
	if err != nil {
		return err
	}
	if role.Name == "superadmin" {
		return ErrProtected
	}
	roles, err := s.repository.Roles(ctx)
	if err != nil {
		return err
	}
	for _, item := range roles {
		if item.ID == id && item.UserCount > 0 {
			return ErrInUse
		}
	}
	return s.repository.DeleteRole(ctx, id)
}

func (s *Service) CreatePermission(ctx context.Context, name, guard, description string) error {
	permission, err := normalizedPermission(0, name, guard, description)
	if err != nil {
		return err
	}
	return translateDuplicate(s.repository.CreatePermission(ctx, permission))
}

func (s *Service) UpdatePermission(ctx context.Context, id uint64, name, guard, description string) error {
	permission, err := normalizedPermission(id, name, guard, description)
	if err != nil {
		return err
	}
	return translateDuplicate(s.repository.UpdatePermission(ctx, permission))
}

func (s *Service) DeletePermission(ctx context.Context, id uint64) error {
	permissions, err := s.repository.Permissions(ctx)
	if err != nil {
		return err
	}
	for _, item := range permissions {
		if item.ID == id && (item.RoleCount > 0 || item.DirectCount > 0) {
			return ErrInUse
		}
	}
	return s.repository.DeletePermission(ctx, id)
}

func (s *Service) SyncRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	if _, err := s.repository.Role(ctx, roleID); err != nil {
		return err
	}
	unique := make([]uint64, 0, len(permissionIDs))
	seen := make(map[uint64]struct{}, len(permissionIDs))
	for _, id := range permissionIDs {
		if _, exists := seen[id]; exists {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return s.repository.SyncRolePermissions(ctx, roleID, unique)
}

func normalizedRole(id uint64, name, guard, description string) (Role, error) {
	name, guard = strings.ToLower(strings.TrimSpace(name)), strings.ToLower(strings.TrimSpace(guard))
	description = strings.TrimSpace(description)
	if len(name) > 125 || len(guard) > 50 || len(description) > 255 || !accessNamePattern.MatchString(name) || !accessNamePattern.MatchString(guard) {
		return Role{}, ErrInvalidName
	}
	return Role{ID: id, Name: name, GuardName: guard, Description: description}, nil
}

func normalizedPermission(id uint64, name, guard, description string) (Permission, error) {
	name, guard = strings.ToLower(strings.TrimSpace(name)), strings.ToLower(strings.TrimSpace(guard))
	description = strings.TrimSpace(description)
	if len(name) > 125 || len(guard) > 50 || len(description) > 255 || !accessNamePattern.MatchString(name) || !accessNamePattern.MatchString(guard) {
		return Permission{}, ErrInvalidName
	}
	return Permission{ID: id, Name: name, GuardName: guard, Description: description}, nil
}

func translateDuplicate(err error) error {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrDuplicate
	}
	return err
}
