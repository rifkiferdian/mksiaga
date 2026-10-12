package checklist

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Checklists(ctx context.Context) ([]Checklist, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,code,phase,name,requires_camera_count,sort_order,status FROM master_checklist ORDER BY FIELD(phase,'start','end'),sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list checklists: %w", err)
	}
	defer rows.Close()
	var result []Checklist
	for rows.Next() {
		var item Checklist
		if err := rows.Scan(&item.ID, &item.Code, &item.Phase, &item.Name, &item.RequiresCameraCount, &item.SortOrder, &item.Status); err != nil {
			return nil, fmt.Errorf("scan checklist: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Checklist(ctx context.Context, id uint64) (Checklist, error) {
	var item Checklist
	err := r.db.QueryRowContext(ctx, `SELECT id,code,phase,name,requires_camera_count,sort_order,status FROM master_checklist WHERE id=?`, id).Scan(&item.ID, &item.Code, &item.Phase, &item.Name, &item.RequiresCameraCount, &item.SortOrder, &item.Status)
	return item, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO master_checklist(code,phase,name,requires_camera_count,sort_order,status) VALUES(?,?,?,?,?,?)`, input.Code, input.Phase, input.Name, input.RequiresCameraCount, input.SortOrder, input.Status)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE master_checklist SET code=?,phase=?,name=?,requires_camera_count=?,sort_order=?,status=? WHERE id=?`, input.Code, input.Phase, input.Name, input.RequiresCameraCount, input.SortOrder, input.Status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_checklist WHERE id=?)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM master_checklist WHERE id=?`, id)
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
