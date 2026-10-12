package area

import (
	"context"
	"database/sql"
	"fmt"
	"mksiaga/internal/storescope"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Areas(ctx context.Context) ([]Area, error) {
	query := `SELECT a.id,a.store_id,s.code,s.name,a.code,a.name,a.is_priority,a.sort_order,a.status,COUNT(k.id)
		FROM master_area a
		JOIN stores s ON s.id=a.store_id AND s.deleted_at IS NULL
		LEFT JOIN master_kamera k ON k.area_id=a.id AND k.store_id=a.store_id
		WHERE ` + storescope.Clause(ctx, "a.store_id") + `
 GROUP BY a.id,a.store_id,s.code,s.name,a.code,a.name,a.is_priority,a.sort_order,a.status
		ORDER BY s.name,a.sort_order,a.name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list areas: %w", err)
	}
	defer rows.Close()
	var result []Area
	for rows.Next() {
		var item Area
		if err := rows.Scan(&item.ID, &item.StoreID, &item.StoreCode, &item.StoreName, &item.Code, &item.Name, &item.IsPriority, &item.SortOrder, &item.Status, &item.CameraCount); err != nil {
			return nil, fmt.Errorf("scan area: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Stores(ctx context.Context) ([]StoreOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,code,name FROM stores WHERE deleted_at IS NULL AND `+storescope.Clause(ctx, "id")+` ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list area stores: %w", err)
	}
	defer rows.Close()
	var result []StoreOption
	for rows.Next() {
		var item StoreOption
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Area(ctx context.Context, id uint64) (Area, error) {
	var item Area
	err := r.db.QueryRowContext(ctx, `SELECT id,store_id,code,name,is_priority,sort_order,status FROM master_area WHERE id=? AND `+storescope.Clause(ctx, "store_id"), id).Scan(&item.ID, &item.StoreID, &item.Code, &item.Name, &item.IsPriority, &item.SortOrder, &item.Status)
	return item, err
}

func (r *Repository) ActiveStoreExists(ctx context.Context, id uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM stores WHERE id=? AND deleted_at IS NULL AND `+storescope.Clause(ctx, "id")+`)`, id).Scan(&exists)
	return exists, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	if !storescope.Allows(ctx, input.StoreID) {
		return sql.ErrNoRows
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO master_area(store_id,code,name,is_priority,sort_order,status) VALUES(?,?,?,?,?,?)`, input.StoreID, input.Code, input.Name, input.IsPriority, input.SortOrder, input.Status)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	if !storescope.Allows(ctx, input.StoreID) {
		return sql.ErrNoRows
	}
	result, err := r.db.ExecContext(ctx, `UPDATE master_area SET store_id=?,code=?,name=?,is_priority=?,sort_order=?,status=? WHERE id=? AND `+storescope.Clause(ctx, "store_id"), input.StoreID, input.Code, input.Name, input.IsPriority, input.SortOrder, input.Status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_area WHERE id=? AND `+storescope.Clause(ctx, "store_id")+`)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *Repository) CameraCount(ctx context.Context, id uint64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM master_kamera WHERE area_id=? AND `+storescope.Clause(ctx, "store_id"), id).Scan(&count)
	return count, err
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM master_area WHERE id=? AND `+storescope.Clause(ctx, "store_id"), id)
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
