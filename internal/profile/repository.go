package profile

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Profile(ctx context.Context, userID, userStoreID uint64) (Profile, error) {
	const query = `SELECT u.id,u.name,u.username,COALESCE(u.email,''),COALESCE(u.phone,''),
		COALESCE(s.name,''),COALESCE(GROUP_CONCAT(DISTINCT ro.name ORDER BY ro.name SEPARATOR ', '),''),
		COALESCE(DATE_FORMAT(u.last_login_at,'%d %b %Y %H:%i'),'Belum pernah')
		FROM users u
		LEFT JOIN user_stores us ON us.id=? AND us.user_id=u.id
		LEFT JOIN stores s ON s.id=us.store_id
		LEFT JOIN user_store_roles usr ON usr.user_store_id=us.id
		LEFT JOIN roles ro ON ro.id=usr.role_id
		WHERE u.id=? AND u.deleted_at IS NULL
		GROUP BY u.id,u.name,u.username,u.email,u.phone,s.name,u.last_login_at`
	var item Profile
	err := r.db.QueryRowContext(ctx, query, userStoreID, userID).Scan(&item.ID, &item.Name, &item.Username, &item.Email, &item.Phone, &item.StoreName, &item.RoleNames, &item.LastLogin)
	if err != nil {
		return Profile{}, fmt.Errorf("load profile: %w", err)
	}
	return item, nil
}

func (r *Repository) UpdateName(ctx context.Context, userID uint64, name string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE users SET name=? WHERE id=? AND deleted_at IS NULL`, name, userID)
	if err != nil {
		return fmt.Errorf("update profile name: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) PasswordHash(ctx context.Context, userID uint64) (string, error) {
	var hash string
	err := r.db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id=? AND deleted_at IS NULL`, userID).Scan(&hash)
	return hash, err
}

func (r *Repository) UpdatePassword(ctx context.Context, userID uint64, passwordHash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash=? WHERE id=? AND deleted_at IS NULL`, passwordHash, userID)
	return err
}
