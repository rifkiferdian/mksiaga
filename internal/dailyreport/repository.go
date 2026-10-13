package dailyreport

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) PageData(ctx context.Context, storeID uint64) (PageData, error) {
	var data PageData
	var err error
	if data.Reports, err = r.reports(ctx, storeID); err != nil {
		return data, err
	}
	if data.Shifts, err = r.options(ctx, `SELECT id,name,code FROM master_shift WHERE status='active' ORDER BY sort_order,name`); err != nil {
		return data, err
	}
	if data.Areas, err = r.options(ctx, `SELECT id,name,code FROM master_area WHERE store_id=? AND status='active' ORDER BY sort_order,name`, storeID); err != nil {
		return data, err
	}
	if data.PriorityAreas, err = r.options(ctx, `SELECT id,name,code FROM master_area WHERE store_id=? AND status='active' AND is_priority=1 ORDER BY sort_order,name`, storeID); err != nil {
		return data, err
	}
	if data.Cameras, err = r.options(ctx, `SELECT id,name,code FROM master_kamera WHERE store_id=? AND status='active' ORDER BY name`, storeID); err != nil {
		return data, err
	}
	if data.Users, err = r.options(ctx, `SELECT u.id,u.name,COALESCE(u.employee_number,'') FROM users u JOIN user_stores us ON us.user_id=u.id WHERE us.store_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL ORDER BY u.name`, storeID); err != nil {
		return data, err
	}
	if data.IncidentTypes, err = r.options(ctx, `SELECT id,name,code FROM master_jenis_insiden WHERE status='active' ORDER BY sort_order,name`); err != nil {
		return data, err
	}
	if data.Actions, err = r.options(ctx, `SELECT id,name,code FROM master_tindakan WHERE status='active' ORDER BY sort_order,name`); err != nil {
		return data, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,phase,requires_camera_count FROM master_checklist WHERE status='active' ORDER BY FIELD(phase,'start','end'),sort_order,name`)
	if err != nil {
		return data, fmt.Errorf("list checklist options: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item ChecklistOption
		if err := rows.Scan(&item.ID, &item.Name, &item.Phase, &item.RequiresCameraCount); err != nil {
			return data, err
		}
		data.Checklists = append(data.Checklists, item)
	}
	return data, rows.Err()
}

func (r *Repository) reports(ctx context.Context, storeID uint64) ([]Report, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,DATE_FORMAT(r.report_date,'%d/%m/%Y'),s.name,u.name,r.status,DATE_FORMAT(r.created_at,'%d/%m/%Y %H:%i') FROM daily_cctv_reports r JOIN master_shift s ON s.id=r.shift_id JOIN users u ON u.id=r.officer_id WHERE r.store_id=? ORDER BY r.report_date DESC,r.id DESC LIMIT 20`, storeID)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()
	var result []Report
	for rows.Next() {
		var item Report
		if err := rows.Scan(&item.ID, &item.ReportDate, &item.ShiftName, &item.OfficerName, &item.Status, &item.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) options(ctx context.Context, query string, args ...any) ([]Option, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Option
	for rows.Next() {
		var item Option
		if err := rows.Scan(&item.ID, &item.Name, &item.Code); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) Create(ctx context.Context, in Input) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM master_shift WHERE id=? AND status='active')`, in.ShiftID) ||
		!exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN users u ON u.id=us.user_id WHERE us.store_id=? AND us.user_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL)`, in.StoreID, in.OfficerID) {
		return ErrInvalid
	}
	for _, id := range []*uint64{in.SupervisorID, in.HandoverToUserID} {
		if id != nil && !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM user_stores WHERE store_id=? AND user_id=? AND status='active')`, in.StoreID, *id) {
			return ErrInvalid
		}
	}
	for _, item := range in.Monitoring {
		if item.AreaID != nil && !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM master_area WHERE id=? AND store_id=? AND status='active')`, *item.AreaID, in.StoreID) {
			return ErrInvalid
		}
	}
	for _, item := range in.Incidents {
		if item.AreaID != nil && !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM master_area WHERE id=? AND store_id=? AND status='active')`, *item.AreaID, in.StoreID) {
			return ErrInvalid
		}
		if item.CameraID != nil && !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM master_kamera WHERE id=? AND store_id=? AND status='active')`, *item.CameraID, in.StoreID) {
			return ErrInvalid
		}
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO daily_cctv_reports(store_id,officer_id,supervisor_id,shift_id,report_date,work_start,work_end,monitoring_location,start_notes,end_notes,handover_to_user_id,handover_at,handover_notes,handover_from_signature,handover_to_signature,status,submitted_at) VALUES(?,?,?,?,?,?,?,?,NULLIF(?,''),NULLIF(?,''),?,NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),?,IF(?='submitted',CURRENT_TIMESTAMP(6),NULL))`, in.StoreID, in.OfficerID, in.SupervisorID, in.ShiftID, in.ReportDate, in.WorkStart, in.WorkEnd, in.Location, in.StartNotes, in.EndNotes, in.HandoverToUserID, in.HandoverAt, in.HandoverNotes, in.HandoverFromSign, in.HandoverToSign, in.Status, in.Status)
	if err != nil {
		return err
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	for _, v := range in.Checklists {
		_, err = tx.ExecContext(ctx, `INSERT INTO daily_cctv_report_checklists(report_id,checklist_id,answer,camera_active,camera_total,notes) SELECT ?,id,?,?,?,NULLIF(?,'') FROM master_checklist WHERE id=? AND status='active'`, reportID, v.Answer, v.Active, v.Total, v.Notes, v.ChecklistID)
		if err != nil {
			return err
		}
	}
	for _, v := range in.Monitoring {
		_, err = tx.ExecContext(ctx, `INSERT INTO daily_cctv_monitoring_logs(report_id,log_time,area_id,condition_status,finding,action_taken) VALUES(?,?,?,?,NULLIF(?,''),NULLIF(?,''))`, reportID, v.Time, v.AreaID, v.Condition, v.Finding, v.Action)
		if err != nil {
			return err
		}
	}
	for _, v := range in.Incidents {
		res, e := tx.ExecContext(ctx, `INSERT INTO daily_cctv_incidents(report_id,incident_time,area_id,camera_id,incident_type_id,other_type,description,perpetrator_gender,perpetrator_age,perpetrator_clothing,perpetrator_features,response_minutes,incident_status) VALUES(?,NULLIF(?,''),?,?,?,NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),?,NULLIF(?,''))`, reportID, v.Time, v.AreaID, v.CameraID, v.IncidentTypeID, v.OtherType, v.Description, v.Gender, v.Age, v.Clothing, v.Features, v.ResponseMinutes, v.Status)
		if e != nil {
			return e
		}
		incidentID, e := res.LastInsertId()
		if e != nil {
			return e
		}
		for _, actionID := range v.ActionIDs {
			_, e = tx.ExecContext(ctx, `INSERT INTO daily_cctv_incident_actions(incident_id,action_id,notes) SELECT ?,id,NULLIF(?,'') FROM master_tindakan WHERE id=? AND status='active'`, incidentID, v.OtherActionNotes, actionID)
			if e != nil {
				return e
			}
		}
	}
	err = tx.Commit()
	return err
}

func (r *Repository) Delete(ctx context.Context, id, storeID uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM daily_cctv_reports WHERE id=? AND store_id=?`, id, storeID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err == nil && n == 0 {
		return sql.ErrNoRows
	}
	return err
}

func (r *Repository) ActiveChecklistCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM master_checklist WHERE status='active'`).Scan(&count)
	return count, err
}

func exists(tx *sql.Tx, ctx context.Context, query string, args ...any) bool {
	var ok bool
	return tx.QueryRowContext(ctx, query, args...).Scan(&ok) == nil && ok
}
