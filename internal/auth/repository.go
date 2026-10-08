package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

var ErrUserNotFound = errors.New("user not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindActiveByLogin(ctx context.Context, login string) (User, error) {
	const query = `
		SELECT
			u.id,
			u.name,
			u.username,
			u.password_hash,
			us.id,
			s.id,
			s.name,
			COALESCE((
				SELECT GROUP_CONCAT(r.name ORDER BY r.name SEPARATOR ', ')
				FROM user_store_roles usr
				JOIN roles r ON r.id = usr.role_id
				WHERE usr.user_store_id = us.id
			), '')
		FROM users u
		JOIN user_stores us ON us.user_id = u.id
		JOIN stores s ON s.id = us.store_id
		WHERE (u.username = ? OR u.email = ?)
			AND u.status = 'active'
			AND u.deleted_at IS NULL
			AND us.status = 'active'
			AND s.status = 'active'
			AND s.deleted_at IS NULL
		ORDER BY us.is_default DESC, us.id ASC
		LIMIT 1`

	var user User
	err := r.db.QueryRowContext(ctx, query, login, login).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.PasswordHash,
		&user.UserStoreID,
		&user.StoreID,
		&user.StoreName,
		&user.RoleName,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find active user: %w", err)
	}

	return user, nil
}

func (r *Repository) RecordLogin(ctx context.Context, userID uint64) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE users SET last_login_at = CURRENT_TIMESTAMP(6) WHERE id = ?`, userID); err != nil {
		return fmt.Errorf("record login: %w", err)
	}
	return nil
}
