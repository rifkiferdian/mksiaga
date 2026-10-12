package camera

import (
	"context"
	"database/sql"
	"fmt"
	"mksiaga/internal/storescope"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Cameras(ctx context.Context) ([]Camera, error) {
	query := `SELECT k.id,k.store_id,s.code,s.name,k.area_id,a.code,a.name,k.code,k.name,COALESCE(k.description,''),k.status
		FROM master_kamera k
		JOIN stores s ON s.id=k.store_id AND s.deleted_at IS NULL
		JOIN master_area a ON a.id=k.area_id AND a.store_id=k.store_id
		WHERE ` + storescope.Clause(ctx, "k.store_id") + ` ORDER BY s.name,a.sort_order,a.name,k.code`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list cameras: %w", err)
	}
	defer rows.Close()
	var result []Camera
	for rows.Next() {
		var item Camera
		if err := rows.Scan(&item.ID, &item.StoreID, &item.StoreCode, &item.StoreName, &item.AreaID, &item.AreaCode, &item.AreaName, &item.Code, &item.Name, &item.Description, &item.Status); err != nil {
			return nil, fmt.Errorf("scan camera: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Stores(ctx context.Context) ([]StoreOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,code,name FROM stores WHERE deleted_at IS NULL AND `+storescope.Clause(ctx, "id")+` ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list camera stores: %w", err)
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

func (r *Repository) Areas(ctx context.Context) ([]AreaOption, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,store_id,code,name FROM master_area WHERE `+storescope.Clause(ctx, "store_id")+` ORDER BY store_id,sort_order,name`)
	if err != nil {
		return nil, fmt.Errorf("list camera areas: %w", err)
	}
	defer rows.Close()
	var result []AreaOption
	for rows.Next() {
		var item AreaOption
		if err := rows.Scan(&item.ID, &item.StoreID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Camera(ctx context.Context, id uint64) (Camera, error) {
	var item Camera
	err := r.db.QueryRowContext(ctx, `SELECT id,store_id,area_id,code,name,COALESCE(description,''),status FROM master_kamera WHERE id=? AND `+storescope.Clause(ctx, "store_id"), id).Scan(&item.ID, &item.StoreID, &item.AreaID, &item.Code, &item.Name, &item.Description, &item.Status)
	return item, err
}

func (r *Repository) AreaBelongsToStore(ctx context.Context, areaID, storeID uint64) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_area a JOIN stores s ON s.id=a.store_id AND s.deleted_at IS NULL WHERE a.id=? AND a.store_id=? AND `+storescope.Clause(ctx, "a.store_id")+`)`, areaID, storeID).Scan(&exists)
	return exists, err
}

func (r *Repository) Create(ctx context.Context, input Input) error {
	if !storescope.Allows(ctx, input.StoreID) {
		return sql.ErrNoRows
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO master_kamera(store_id,area_id,code,name,description,status) VALUES(?,?,?,?,NULLIF(?,''),?)`, input.StoreID, input.AreaID, input.Code, input.Name, input.Description, input.Status)
	return err
}

func (r *Repository) Update(ctx context.Context, id uint64, input Input) error {
	if !storescope.Allows(ctx, input.StoreID) {
		return sql.ErrNoRows
	}
	result, err := r.db.ExecContext(ctx, `UPDATE master_kamera SET store_id=?,area_id=?,code=?,name=?,description=NULLIF(?,''),status=? WHERE id=? AND `+storescope.Clause(ctx, "store_id"), input.StoreID, input.AreaID, input.Code, input.Name, input.Description, input.Status, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_kamera WHERE id=? AND `+storescope.Clause(ctx, "store_id")+`)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM master_kamera WHERE id=? AND `+storescope.Clause(ctx, "store_id"), id)
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
