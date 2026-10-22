package dailyreport

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"mksiaga/internal/storescope"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Recap(ctx context.Context, filter RecapFilter) (RecapData, error) {
	data := RecapData{Filter: filter}
	stores, err := storescope.Stores(ctx, r.db)
	if err != nil {
		return data, err
	}
	data.Stores = stores
	where, args := recapWhere(ctx, filter)
	if where == "" {
		return data, ErrInvalid
	}
	summaryQuery := `SELECT COUNT(*),COALESCE(SUM(r.status='approved'),0),COALESCE(SUM(r.status='waiting_handover'),0),COALESCE(SUM(r.status='waiting_review'),0),COALESCE(SUM(r.status='revision'),0),COALESCE(SUM(r.status='draft'),0),COALESCE(SUM((SELECT COUNT(*) FROM daily_cctv_monitoring_logs ml WHERE ml.report_id=r.id AND ml.condition_status='finding')),0),COALESCE(SUM((SELECT COUNT(*) FROM daily_cctv_incidents di WHERE di.report_id=r.id)),0) FROM daily_cctv_reports r WHERE ` + where
	if err := r.db.QueryRowContext(ctx, summaryQuery, args...).Scan(&data.Summary.Total, &data.Summary.Approved, &data.Summary.WaitingHandover, &data.Summary.WaitingReview, &data.Summary.Revision, &data.Summary.Draft, &data.Summary.Findings, &data.Summary.Incidents); err != nil {
		return data, fmt.Errorf("recap summary: %w", err)
	}
	groupQuery := `SELECT st.name,COUNT(*),COALESCE(SUM(r.status='approved'),0),COALESCE(SUM(r.status IN ('waiting_handover','waiting_review')),0),COALESCE(SUM(r.status='revision'),0),COALESCE(SUM((SELECT COUNT(*) FROM daily_cctv_monitoring_logs ml WHERE ml.report_id=r.id AND ml.condition_status='finding')),0),COALESCE(SUM((SELECT COUNT(*) FROM daily_cctv_incidents di WHERE di.report_id=r.id)),0) FROM daily_cctv_reports r JOIN stores st ON st.id=r.store_id WHERE ` + where + ` GROUP BY st.id,st.name ORDER BY st.name`
	rows, err := r.db.QueryContext(ctx, groupQuery, args...)
	if err != nil {
		return data, fmt.Errorf("recap stores: %w", err)
	}
	for rows.Next() {
		var item RecapStore
		if err := rows.Scan(&item.StoreName, &item.Total, &item.Approved, &item.Pending, &item.Revision, &item.Findings, &item.Incidents); err != nil {
			rows.Close()
			return data, err
		}
		data.ByStore = append(data.ByStore, item)
	}
	if err := rows.Close(); err != nil {
		return data, err
	}
	detailQuery := `SELECT r.id,DATE_FORMAT(r.report_date,'%d/%m/%Y'),st.name,sh.name,u.name,r.status,(SELECT COUNT(*) FROM daily_cctv_monitoring_logs ml WHERE ml.report_id=r.id AND ml.condition_status='finding'),(SELECT COUNT(*) FROM daily_cctv_incidents di WHERE di.report_id=r.id) FROM daily_cctv_reports r JOIN stores st ON st.id=r.store_id JOIN master_shift sh ON sh.id=r.shift_id JOIN users u ON u.id=r.officer_id WHERE ` + where + ` ORDER BY r.report_date DESC,r.id DESC LIMIT 200`
	rows, err = r.db.QueryContext(ctx, detailQuery, args...)
	if err != nil {
		return data, fmt.Errorf("recap rows: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item RecapRow
		if err := rows.Scan(&item.ID, &item.ReportDate, &item.StoreName, &item.ShiftName, &item.OfficerName, &item.Status, &item.Findings, &item.Incidents); err != nil {
			return data, err
		}
		item.ReportNumber = fmt.Sprintf("CCTV-%06d", item.ID)
		data.Rows = append(data.Rows, item)
	}
	return data, rows.Err()
}

func recapWhere(ctx context.Context, filter RecapFilter) (string, []any) {
	parts := []string{storescope.Clause(ctx, "r.store_id"), "r.report_date BETWEEN ? AND ?"}
	args := []any{filter.StartDate, filter.EndDate}
	if filter.StoreID > 0 {
		if !storescope.Allows(ctx, filter.StoreID) {
			return "", nil
		}
		parts = append(parts, "r.store_id=?")
		args = append(args, filter.StoreID)
	}
	if filter.Status != "" {
		parts = append(parts, "r.status=?")
		args = append(args, filter.Status)
	}
	return strings.Join(parts, " AND "), args
}

func (r *Repository) PageData(ctx context.Context, storeID, userID uint64) (PageData, error) {
	var data PageData
	var err error
	if data.Reports, err = r.reports(ctx, userID); err != nil {
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
	if data.Users, err = r.options(ctx, `SELECT u.id,u.name,COALESCE(u.employee_number,'') FROM users u JOIN user_stores us ON us.user_id=u.id WHERE us.store_id=? AND u.id<>? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL ORDER BY u.name`, storeID, userID); err != nil {
		return data, err
	}
	if data.Supervisors, err = r.options(ctx, `SELECT DISTINCT u.id,u.name,COALESCE(u.employee_number,'') FROM users u JOIN user_stores us ON us.user_id=u.id JOIN user_store_roles usr ON usr.user_store_id=us.id JOIN roles r ON r.id=usr.role_id WHERE us.store_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL AND r.name IN ('spv_security','superadmin') AND r.guard_name='web' ORDER BY u.name`, storeID); err != nil {
		return data, err
	}
	if data.IncidentTypes, err = r.options(ctx, `SELECT id,name,code FROM master_jenis_insiden WHERE status='active' ORDER BY sort_order,name`); err != nil {
		return data, err
	}
	if data.Actions, err = r.options(ctx, `SELECT id,name,code FROM master_tindakan WHERE status='active' ORDER BY sort_order,name`); err != nil {
		return data, err
	}
	_ = r.db.QueryRowContext(ctx, `SELECT COALESCE(default_shift_id,0) FROM user_stores WHERE store_id=? AND user_id=? AND status='active'`, storeID, userID).Scan(&data.DefaultShiftID)
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

func (r *Repository) reports(ctx context.Context, userID uint64) ([]Report, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,st.name,DATE_FORMAT(r.report_date,'%d/%m/%Y'),s.name,u.name,r.status,DATE_FORMAT(r.created_at,'%d/%m/%Y %H:%i'),r.handover_to_user_id,(r.status='waiting_handover' AND r.handover_to_user_id=?),(r.status='waiting_review' AND r.supervisor_id=?) FROM daily_cctv_reports r JOIN stores st ON st.id=r.store_id AND st.deleted_at IS NULL JOIN master_shift s ON s.id=r.shift_id JOIN users u ON u.id=r.officer_id WHERE `+storescope.Clause(ctx, "r.store_id")+` ORDER BY r.report_date DESC,r.id DESC LIMIT 20`, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("list reports: %w", err)
	}
	defer rows.Close()
	var result []Report
	for rows.Next() {
		var item Report
		if err := rows.Scan(&item.ID, &item.StoreName, &item.ReportDate, &item.ShiftName, &item.OfficerName, &item.Status, &item.CreatedAt, &item.HandoverToUserID, &item.CanReceive, &item.CanReview); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) WorkflowReport(ctx context.Context, id uint64) (WorkflowReport, error) {
	var item WorkflowReport
	err := r.db.QueryRowContext(ctx, `SELECT r.id,st.name,r.officer_id,DATE_FORMAT(r.report_date,'%d/%m/%Y'),s.name,o.name,COALESCE(r.supervisor_id,0),COALESCE(supervisor.name,''),TIME_FORMAT(r.work_start,'%H:%i'),TIME_FORMAT(r.work_end,'%H:%i'),r.monitoring_location,COALESCE(r.start_notes,''),COALESCE(r.end_notes,''),COALESCE(receiver.name,''),COALESCE(r.handover_to_user_id,0),COALESCE(TIME_FORMAT(r.handover_at,'%H:%i'),''),COALESCE(r.handover_notes,''),r.status,COALESCE(r.review_notes,'') FROM daily_cctv_reports r JOIN stores st ON st.id=r.store_id AND st.deleted_at IS NULL JOIN master_shift s ON s.id=r.shift_id JOIN users o ON o.id=r.officer_id LEFT JOIN users supervisor ON supervisor.id=r.supervisor_id LEFT JOIN users receiver ON receiver.id=r.handover_to_user_id WHERE r.id=? AND `+storescope.Clause(ctx, "r.store_id"), id).Scan(&item.ID, &item.StoreName, &item.OfficerID, &item.ReportDate, &item.ShiftName, &item.OfficerName, &item.SupervisorID, &item.SupervisorName, &item.WorkStart, &item.WorkEnd, &item.Location, &item.StartNotes, &item.EndNotes, &item.HandoverToName, &item.HandoverToUserID, &item.HandoverAt, &item.HandoverNotes, &item.Status, &item.ReviewNotes)
	if err != nil {
		return item, err
	}
	item.ReportNumber = fmt.Sprintf("CCTV-%06d", item.ID)
	rows, err := r.db.QueryContext(ctx, `SELECT mc.phase,mc.name,rc.answer,CASE WHEN rc.camera_total IS NULL THEN '' ELSE CONCAT(COALESCE(rc.camera_active,0),' / ',rc.camera_total) END,COALESCE(rc.notes,'') FROM daily_cctv_report_checklists rc JOIN master_checklist mc ON mc.id=rc.checklist_id WHERE rc.report_id=? ORDER BY FIELD(mc.phase,'start','end'),mc.sort_order,mc.name`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var v WorkflowChecklist
		if err = rows.Scan(&v.Phase, &v.Name, &v.Answer, &v.CameraCount, &v.Notes); err != nil {
			rows.Close()
			return item, err
		}
		item.Checklists = append(item.Checklists, v)
		if v.Answer == "no" {
			item.ChecklistIssueCount++
		}
	}
	if err = rows.Close(); err != nil {
		return item, err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT TIME_FORMAT(l.log_time,'%H:%i'),COALESCE(a.name,'-'),l.condition_status,COALESCE(l.finding,''),COALESCE(l.action_taken,'') FROM daily_cctv_monitoring_logs l LEFT JOIN master_area a ON a.id=l.area_id WHERE l.report_id=? ORDER BY l.log_time,l.id`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var v WorkflowMonitoring
		if err = rows.Scan(&v.Time, &v.AreaName, &v.Condition, &v.Finding, &v.Action); err != nil {
			rows.Close()
			return item, err
		}
		item.Monitoring = append(item.Monitoring, v)
		if v.Condition == "finding" {
			item.FindingCount++
		}
	}
	if err = rows.Close(); err != nil {
		return item, err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT COALESCE(TIME_FORMAT(i.incident_time,'%H:%i'),'-'),COALESCE(ar.name,'-'),COALESCE(cam.name,'-'),COALESCE(NULLIF(i.other_type,''),it.name,'-'),COALESCE(i.description,''),CONCAT_WS(' · ',NULLIF(CONCAT('JK: ',COALESCE(i.perpetrator_gender,'')),'JK: '),NULLIF(CONCAT('Usia: ',COALESCE(i.perpetrator_age,'')),'Usia: '),NULLIF(CONCAT('Pakaian: ',COALESCE(i.perpetrator_clothing,'')),'Pakaian: '),NULLIF(CONCAT('Ciri: ',COALESCE(i.perpetrator_features,'')),'Ciri: ')),CASE WHEN i.response_minutes IS NULL THEN '-' ELSE CONCAT(i.response_minutes,' menit') END,COALESCE(i.incident_status,'-'),COALESCE(GROUP_CONCAT(DISTINCT CONCAT(act.name,IF(ia.notes IS NULL OR ia.notes='','',CONCAT(' (',ia.notes,')'))) ORDER BY act.sort_order SEPARATOR ', '),'-') FROM daily_cctv_incidents i LEFT JOIN master_area ar ON ar.id=i.area_id LEFT JOIN master_kamera cam ON cam.id=i.camera_id LEFT JOIN master_jenis_insiden it ON it.id=i.incident_type_id LEFT JOIN daily_cctv_incident_actions ia ON ia.incident_id=i.id LEFT JOIN master_tindakan act ON act.id=ia.action_id WHERE i.report_id=? GROUP BY i.id ORDER BY i.incident_time,i.id`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var v WorkflowIncident
		if err = rows.Scan(&v.Time, &v.AreaName, &v.CameraName, &v.TypeName, &v.Description, &v.PersonDetails, &v.Response, &v.Status, &v.Actions); err != nil {
			rows.Close()
			return item, err
		}
		item.Incidents = append(item.Incidents, v)
		item.IncidentCount++
	}
	if err = rows.Close(); err != nil {
		return item, err
	}
	rows, err = r.db.QueryContext(ctx, `SELECT u.name,e.from_status,e.to_status,COALESCE(e.notes,''),DATE_FORMAT(e.created_at,'%d/%m/%Y %H:%i') FROM daily_cctv_report_events e JOIN users u ON u.id=e.actor_user_id WHERE e.report_id=? ORDER BY e.created_at,e.id`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var v WorkflowEvent
		if err = rows.Scan(&v.ActorName, &v.FromStatus, &v.ToStatus, &v.Notes, &v.CreatedAt); err != nil {
			rows.Close()
			return item, err
		}
		item.Events = append(item.Events, v)
	}
	if err = rows.Close(); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) Receive(ctx context.Context, id, storeID, userID uint64, signature, notes string, accept bool) error {
	toStatus := "revision"
	if accept {
		toStatus = "waiting_review"
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	result, err := tx.ExecContext(ctx, `UPDATE daily_cctv_reports SET status=?,handover_to_signature=NULLIF(?,''),accepted_at=IF(?,CURRENT_TIMESTAMP(6),NULL),review_notes=NULLIF(?,'') WHERE id=? AND store_id=? AND handover_to_user_id=? AND status='waiting_handover'`, toStatus, signature, accept, notes, id, storeID, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO daily_cctv_report_events(report_id,actor_user_id,from_status,to_status,notes) VALUES(?,?,'waiting_handover',?,NULLIF(?,''))`, id, userID, toStatus, notes)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func (r *Repository) Review(ctx context.Context, id, userID uint64, notes string, approve, allowAnySupervisor bool) error {
	toStatus := "revision"
	if approve {
		toStatus = "approved"
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	result, err := tx.ExecContext(ctx, `UPDATE daily_cctv_reports SET status=?,reviewed_by_user_id=?,reviewed_at=IF(?,CURRENT_TIMESTAMP(6),NULL),review_notes=NULLIF(?,'') WHERE id=? AND `+storescope.Clause(ctx, "store_id")+` AND status='waiting_review' AND (supervisor_id=? OR ?=1)`, toStatus, userID, approve, notes, id, userID, allowAnySupervisor)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO daily_cctv_report_events(report_id,actor_user_id,from_status,to_status,notes) VALUES(?,?,'waiting_review',?,NULLIF(?,''))`, id, userID, toStatus, notes)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
}

func (r *Repository) Revise(ctx context.Context, id, storeID, userID, receiverID uint64, handoverAt, notes, signature string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	if !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM user_stores WHERE store_id=? AND user_id=? AND status='active')`, storeID, receiverID) {
		return ErrInvalid
	}
	result, err := tx.ExecContext(ctx, `UPDATE daily_cctv_reports SET status='waiting_handover',handover_to_user_id=?,handover_at=?,handover_notes=NULLIF(?,''),handover_from_signature=?,handover_to_signature=NULL,accepted_at=NULL,reviewed_by_user_id=NULL,reviewed_at=NULL,review_notes=NULL,submitted_at=CURRENT_TIMESTAMP(6) WHERE id=? AND store_id=? AND officer_id=? AND status='revision'`, receiverID, handoverAt, notes, signature, id, storeID, userID)
	if err != nil {
		return err
	}
	n, _ := result.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO daily_cctv_report_events(report_id,actor_user_id,from_status,to_status,notes) VALUES(?,?,'revision','waiting_handover',NULLIF(?,''))`, id, userID, notes)
	if err != nil {
		return err
	}
	err = tx.Commit()
	return err
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
	if in.SupervisorID != nil && !exists(tx, ctx, `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN user_store_roles usr ON usr.user_store_id=us.id JOIN roles r ON r.id=usr.role_id WHERE us.store_id=? AND us.user_id=? AND us.status='active' AND r.name IN ('spv_security','superadmin') AND r.guard_name='web')`, in.StoreID, *in.SupervisorID) {
		return ErrInvalid
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
	result, err := tx.ExecContext(ctx, `INSERT INTO daily_cctv_reports(store_id,officer_id,supervisor_id,shift_id,report_date,work_start,work_end,monitoring_location,start_notes,end_notes,handover_to_user_id,handover_at,handover_notes,handover_from_signature,status,submitted_at) VALUES(?,?,?,?,?,?,?,?,NULLIF(?,''),NULLIF(?,''),?,NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),?,IF(?='waiting_handover',CURRENT_TIMESTAMP(6),NULL))`, in.StoreID, in.OfficerID, in.SupervisorID, in.ShiftID, in.ReportDate, in.WorkStart, in.WorkEnd, in.Location, in.StartNotes, in.EndNotes, in.HandoverToUserID, in.HandoverAt, in.HandoverNotes, in.HandoverFromSign, in.Status, in.Status)
	if err != nil {
		return err
	}
	reportID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if in.Status == "waiting_handover" {
		if _, err = tx.ExecContext(ctx, `INSERT INTO daily_cctv_report_events(report_id,actor_user_id,from_status,to_status,notes) VALUES(?,?,'draft','waiting_handover',NULLIF(?,''))`, reportID, in.OfficerID, in.HandoverNotes); err != nil {
			return err
		}
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

func (r *Repository) Delete(ctx context.Context, id uint64) error {
	result, err := r.db.ExecContext(ctx, `DELETE FROM daily_cctv_reports WHERE id=? AND `+storescope.Clause(ctx, "store_id")+` AND status IN ('draft','revision')`, id)
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
