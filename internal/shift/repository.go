package shift

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Shifts(ctx context.Context) ([]Shift, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,code,name,COALESCE(TIME_FORMAT(start_time,'%H:%i'),''),COALESCE(TIME_FORMAT(end_time,'%H:%i'),''),sort_order,status FROM master_shift ORDER BY sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list shifts: %w", err)
	}
	defer rows.Close()
	var result []Shift
	for rows.Next() {
		var item Shift
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.StartTime, &item.EndTime, &item.SortOrder, &item.Status); err != nil {
			return nil, fmt.Errorf("scan shift: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Shift(ctx context.Context, id uint64) (Shift, error) {
	var item Shift
	err := r.db.QueryRowContext(ctx, `SELECT id,code,name,COALESCE(TIME_FORMAT(start_time,'%H:%i'),''),COALESCE(TIME_FORMAT(end_time,'%H:%i'),''),sort_order,status FROM master_shift WHERE id=?`, id).Scan(&item.ID, &item.Code, &item.Name, &item.StartTime, &item.EndTime, &item.SortOrder, &item.Status)
	return item, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO master_shift(code,name,start_time,end_time,sort_order,status) VALUES(?,?,NULLIF(?,''),NULLIF(?,''),?,?)`, input.Code, input.Name, input.StartTime, input.EndTime, input.SortOrder, input.Status)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE master_shift SET code=?,name=?,start_time=NULLIF(?,''),end_time=NULLIF(?,''),sort_order=?,status=? WHERE id=?`, input.Code, input.Name, input.StartTime, input.EndTime, input.SortOrder, input.Status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_shift WHERE id=?)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM master_shift WHERE id=?`, id)
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
