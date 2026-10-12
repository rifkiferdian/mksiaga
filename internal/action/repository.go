package action

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Actions(ctx context.Context) ([]Action, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,code,name,requires_description,sort_order,status FROM master_tindakan ORDER BY sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list actions: %w", err)
	}
	defer rows.Close()
	var result []Action
	for rows.Next() {
		var item Action
		if err := rows.Scan(&item.ID, &item.Code, &item.Name, &item.RequiresDescription, &item.SortOrder, &item.Status); err != nil {
			return nil, fmt.Errorf("scan action: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Action(ctx context.Context, id uint64) (Action, error) {
	var item Action
	err := r.db.QueryRowContext(ctx, `SELECT id,code,name,requires_description,sort_order,status FROM master_tindakan WHERE id=?`, id).Scan(&item.ID, &item.Code, &item.Name, &item.RequiresDescription, &item.SortOrder, &item.Status)
	return item, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO master_tindakan(code,name,requires_description,sort_order,status) VALUES(?,?,?,?,?)`, input.Code, input.Name, input.RequiresDescription, input.SortOrder, input.Status)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	result, err := r.db.ExecContext(ctx, `UPDATE master_tindakan SET code=?,name=?,requires_description=?,sort_order=?,status=? WHERE id=?`, input.Code, input.Name, input.RequiresDescription, input.SortOrder, input.Status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_tindakan WHERE id=?)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM master_tindakan WHERE id=?`, id)
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
