package access

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Roles(ctx context.Context) ([]Role, error) {
	const query = `SELECT r.id, r.name, r.guard_name, COALESCE(r.description, ''),
		COUNT(DISTINCT rp.permission_id), COUNT(DISTINCT usr.user_store_id)
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		LEFT JOIN user_store_roles usr ON usr.role_id = r.id
		GROUP BY r.id, r.name, r.guard_name, r.description ORDER BY r.name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	defer rows.Close()
	var result []Role
	for rows.Next() {
		var item Role
		if err := rows.Scan(&item.ID, &item.Name, &item.GuardName, &item.Description, &item.PermissionCount, &item.UserCount); err != nil {
			return nil, fmt.Errorf("scan role: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate roles: %w", err)
	}
	return result, nil
}

func (r *Repository) Role(ctx context.Context, id uint64) (Role, error) {
	var item Role
	err := r.db.QueryRowContext(ctx, `SELECT id,name,guard_name,COALESCE(description,'') FROM roles WHERE id=?`, id).Scan(&item.ID, &item.Name, &item.GuardName, &item.Description)
	if err != nil {
		return Role{}, err
	}
	return item, nil
}

func (r *Repository) CreateRole(ctx context.Context, role Role) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO roles(name,guard_name,description) VALUES(?,?,NULLIF(?,''))`, role.Name, role.GuardName, role.Description)
	return err
}

func (r *Repository) UpdateRole(ctx context.Context, role Role) error {
	_, err := r.db.ExecContext(ctx, `UPDATE roles SET name=?,guard_name=?,description=NULLIF(?,'') WHERE id=?`, role.Name, role.GuardName, role.Description, role.ID)
	return err
}

func (r *Repository) DeleteRole(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM roles WHERE id=?`, id)
	return err
}

func (r *Repository) Permissions(ctx context.Context) ([]Permission, error) {
	const query = `SELECT p.id,p.name,p.guard_name,COALESCE(p.description,''),
		COUNT(DISTINCT rp.role_id),COUNT(DISTINCT usp.user_store_id)
		FROM permissions p
		LEFT JOIN role_permissions rp ON rp.permission_id=p.id
		LEFT JOIN user_store_permissions usp ON usp.permission_id=p.id
		GROUP BY p.id,p.name,p.guard_name,p.description ORDER BY p.name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list permissions: %w", err)
	}
	defer rows.Close()
	var result []Permission
	for rows.Next() {
		var item Permission
		if err := rows.Scan(&item.ID, &item.Name, &item.GuardName, &item.Description, &item.RoleCount, &item.DirectCount); err != nil {
			return nil, fmt.Errorf("scan permission: %w", err)
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate permissions: %w", err)
	}
	return result, nil
}

func (r *Repository) PermissionsForRole(ctx context.Context, roleID uint64) ([]Permission, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT p.id,p.name,p.guard_name,COALESCE(p.description,''),rp.role_id IS NOT NULL
		FROM permissions p LEFT JOIN role_permissions rp ON rp.permission_id=p.id AND rp.role_id=? ORDER BY p.name`, roleID)
	if err != nil {
		return nil, fmt.Errorf("list role permissions: %w", err)
	}
	defer rows.Close()
	var result []Permission
	for rows.Next() {
		var item Permission
		if err := rows.Scan(&item.ID, &item.Name, &item.GuardName, &item.Description, &item.Assigned); err != nil {
			return nil, fmt.Errorf("scan role permission: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) CreatePermission(ctx context.Context, permission Permission) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO permissions(name,guard_name,description) VALUES(?,?,NULLIF(?,''))`, permission.Name, permission.GuardName, permission.Description)
	return err
}

func (r *Repository) UpdatePermission(ctx context.Context, permission Permission) error {
	_, err := r.db.ExecContext(ctx, `UPDATE permissions SET name=?,guard_name=?,description=NULLIF(?,'') WHERE id=?`, permission.Name, permission.GuardName, permission.Description, permission.ID)
	return err
}

func (r *Repository) DeletePermission(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM permissions WHERE id=?`, id)
	return err
}

func (r *Repository) SyncRolePermissions(ctx context.Context, roleID uint64, permissionIDs []uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin role permission transaction: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id=?`, roleID); err != nil {
		return fmt.Errorf("clear role permissions: %w", err)
	}
	for _, permissionID := range permissionIDs {
		if _, err := tx.ExecContext(ctx, `INSERT INTO role_permissions(role_id,permission_id) SELECT ?,id FROM permissions WHERE id=?`, roleID, permissionID); err != nil {
			return fmt.Errorf("assign role permission: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit role permissions: %w", err)
	}
	return nil
}
