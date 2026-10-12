package incidenttype

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) IncidentTypes(ctx context.Context) ([]IncidentType, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,code,name,requires_description,sort_order,status FROM master_jenis_insiden ORDER BY sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list incident types: %w", err)
	}
	defer rows.Close()
	var result []IncidentType
	for rows.Next() {
		var item IncidentType
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.RequiresDescription, &item.SortOrder, &item.Status); err != nil {
			return nil, fmt.Errorf("scan incident type: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) IncidentType(ctx context.Context, id uint64) (IncidentType, error) {
	var item IncidentType
	err := r.db.QueryRowContext(ctx, `SELECT id,code,name,requires_description,sort_order,status FROM master_jenis_insiden WHERE id=?`, id).Scan(&item.ID, &item.Code, &item.Name, &item.RequiresDescription, &item.SortOrder, &item.Status)
	return item, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO master_jenis_insiden(code,name,requires_description,sort_order,status) VALUES(?,?,?,?,?)`, input.Code, input.Name, input.RequiresDescription, input.SortOrder, input.Status)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE master_jenis_insiden SET code=?,name=?,requires_description=?,sort_order=?,status=? WHERE id=?`, input.Code, input.Name, input.RequiresDescription, input.SortOrder, input.Status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_jenis_insiden WHERE id=?)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM master_jenis_insiden WHERE id=?`, id)
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
