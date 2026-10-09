package store

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Stores(ctx context.Context) ([]Store, error) {
	const query = `SELECT s.id,s.code,s.name,COALESCE(s.address,''),COALESCE(s.phone,''),s.timezone,s.status,COUNT(DISTINCT us.user_id)
		FROM stores s LEFT JOIN user_stores us ON us.store_id=s.id AND us.status='active'
		WHERE s.deleted_at IS NULL GROUP BY s.id,s.code,s.name,s.address,s.phone,s.timezone,s.status ORDER BY s.name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}
	defer rows.Close()
	var result []Store
	for rows.Next() {
		var item Store
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.Address, &item.Phone, &item.Timezone, &item.Status, &item.UserCount); err != nil {
			return nil, fmt.Errorf("scan store: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}
func (r *Repository) Store(ctx context.Context, id uint64) (Store, error) {
	var item Store
	err := r.db.QueryRowContext(ctx, `SELECT id,code,name,COALESCE(address,''),COALESCE(phone,''),timezone,status FROM stores WHERE id=? AND deleted_at IS NULL`, id).Scan(&item.ID, &item.Code, &item.Name, &item.Address, &item.Phone, &item.Timezone, &item.Status)
	return item, err
}
func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO stores(code,name,address,phone,timezone,status) VALUES(?,?,NULLIF(?,''),NULLIF(?,''),?,?)`, input.Code, input.Name, input.Address, input.Phone, input.Timezone, input.Status)
	return err
}
func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE stores SET code=?,name=?,address=NULLIF(?,''),phone=NULLIF(?,''),timezone=?,status=? WHERE id=? AND deleted_at IS NULL`, input.Code, input.Name, input.Address, input.Phone, input.Timezone, input.Status, id)
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
func (r *Repository) Delete(ctx context.Context, id uint64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE stores SET status='inactive',deleted_at=CURRENT_TIMESTAMP(6) WHERE id=? AND deleted_at IS NULL`, id)
	return err
}
