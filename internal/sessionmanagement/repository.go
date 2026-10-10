package sessionmanagement

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Sessions(ctx context.Context, userID uint64) ([]Session, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,ip_address,user_agent,created_at,last_seen_at,expires_at FROM user_sessions WHERE user_id=? AND revoked_at IS NULL AND expires_at>CURRENT_TIMESTAMP(6) ORDER BY last_seen_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()
	var result []Session
	for rows.Next() {
		var item Session
		if err := rows.Scan(&item.ID, &item.IPAddress, &item.UserAgent, &item.CreatedAt, &item.LastSeenAt, &item.ExpiresAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Revoke(ctx context.Context, userID uint64, id, currentID string) error {
	result, err := r.db.ExecContext(ctx, `UPDATE user_sessions SET revoked_at=CURRENT_TIMESTAMP(6) WHERE id=? AND user_id=? AND id<>? AND revoked_at IS NULL`, id, userID, currentID)
	if err != nil {
		return err
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
func (r *Repository) RevokeOthers(ctx context.Context, userID uint64, currentID string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE user_sessions SET revoked_at=CURRENT_TIMESTAMP(6) WHERE user_id=? AND id<>? AND revoked_at IS NULL`, userID, currentID)
	return err
}
