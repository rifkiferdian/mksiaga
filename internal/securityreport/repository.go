package securityreport

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"strings"
	"time"

	"mksiaga/internal/storescope"
)

func (r *Repository) History(ctx context.Context, storeID uint64, f HistoryFilter) (HistoryData, error) {
	where := []string{storescope.Clause(ctx, "r.store_id"), "r.report_date BETWEEN ? AND ?"}
	args := []any{f.StartDate, f.EndDate}
	if f.ShiftID > 0 {
		where = append(where, "r.shift_id=?")
		args = append(args, f.ShiftID)
	}
	if f.UserID > 0 {
		where = append(where, "EXISTS(SELECT 1 FROM security_shift_report_members mm WHERE mm.report_id=r.id AND mm.user_id=?)")
		args = append(args, f.UserID)
	}
	if f.Status != "" {
		where = append(where, "r.status=?")
		args = append(args, f.Status)
	}
	if f.Level != "" {
		where = append(where, "EXISTS(SELECT 1 FROM security_shift_hourly_logs ll WHERE ll.report_id=r.id AND ll.level=?)")
		args = append(args, f.Level)
	}
	q := `SELECT r.id,r.representative_user_id,DATE_FORMAT(r.report_date,'%d-%m-%Y'),s.name,sh.name,u.name,r.status,DATE_FORMAT(r.work_started_at,'%d-%m-%Y %H:%i'),COALESCE(DATE_FORMAT(r.work_ended_at,'%d-%m-%Y %H:%i'),''),(SELECT COUNT(*) FROM security_shift_report_members m WHERE m.report_id=r.id),(SELECT COUNT(*) FROM security_shift_hourly_logs l WHERE l.report_id=r.id),COALESCE((SELECT CASE MAX(FIELD(l.level,'normal','waspada','urgent')) WHEN 3 THEN 'urgent' WHEN 2 THEN 'waspada' WHEN 1 THEN 'normal' END FROM security_shift_hourly_logs l WHERE l.report_id=r.id),'') FROM security_shift_reports r JOIN stores s ON s.id=r.store_id JOIN master_shift sh ON sh.id=r.shift_id JOIN users u ON u.id=r.representative_user_id WHERE ` + strings.Join(where, " AND ") + ` ORDER BY r.report_date DESC,r.id DESC LIMIT 500`
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return HistoryData{}, err
	}
	defer rows.Close()
	var reports []ListItem
	for rows.Next() {
		var v ListItem
		if err = rows.Scan(&v.ID, &v.RepresentativeID, &v.ReportDate, &v.StoreName, &v.ShiftName, &v.RepresentativeName, &v.Status, &v.WorkStartedAt, &v.WorkEndedAt, &v.MemberCount, &v.LogCount, &v.HighestLevel); err != nil {
			return HistoryData{}, err
		}
		reports = append(reports, v)
	}
	if err = rows.Err(); err != nil {
		return HistoryData{}, err
	}
	shifts, err := r.options(ctx, `SELECT id,name,code,COALESCE(TIME_FORMAT(start_time,'%H:%i'),''),COALESCE(TIME_FORMAT(end_time,'%H:%i'),'') FROM master_shift WHERE status='active' ORDER BY sort_order,name`)
	if err != nil {
		return HistoryData{}, err
	}
	users, err := r.options(ctx, `SELECT DISTINCT u.id,u.name,u.username,'','' FROM users u JOIN user_stores us ON us.user_id=u.id WHERE us.store_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL ORDER BY u.name`, storeID)
	return HistoryData{Filter: f, Reports: reports, Shifts: shifts, Users: users}, err
}

func (r *Repository) Monitoring(ctx context.Context) (MonitoringData, error) {
	var d MonitoringData
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(r.status='active'),0),COALESCE(SUM(r.status='active' AND ((SELECT MAX(l.reported_at) FROM security_shift_hourly_logs l WHERE l.report_id=r.id) IS NULL OR (SELECT MAX(l.reported_at) FROM security_shift_hourly_logs l WHERE l.report_id=r.id)<CURRENT_TIMESTAMP-INTERVAL 60 MINUTE)),0),COALESCE(SUM(r.status='waiting_handover'),0),COALESCE(SUM(r.status='waiting_supervisor'),0) FROM security_shift_reports r WHERE `+storescope.Clause(ctx, "r.store_id")).Scan(&d.ActiveShifts, &d.LateReports, &d.WaitingReceiver, &d.WaitingSPV)
	if err != nil {
		return d, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM security_shift_hourly_logs l JOIN security_shift_reports r ON r.id=l.report_id WHERE l.level='urgent' AND l.follow_up_status IN ('monitoring','in_progress','forwarded') AND `+storescope.Clause(ctx, "r.store_id")).Scan(&d.UrgentOpen)
	if err != nil {
		return d, err
	}
	err = r.db.QueryRowContext(ctx, `SELECT GREATEST((SELECT COUNT(*) FROM master_shift WHERE status='active')-(SELECT COUNT(*) FROM security_shift_reports r WHERE r.report_date=CURRENT_DATE AND `+storescope.Clause(ctx, "r.store_id")+`),0)`).Scan(&d.MissingReports)
	if err != nil {
		return d, err
	}
	all, err := r.reports(ctx)
	if err != nil {
		return d, err
	}
	for _, v := range all {
		if v.Status == "active" {
			d.Active = append(d.Active, v)
		}
	}
	rows, err := r.db.QueryContext(ctx, `SELECT l.id,l.created_by_user_id,COALESCE(l.area_id,0),DATE_FORMAT(l.reported_at,'%d-%m-%Y %H:%i'),l.level,COALESCE(a.name,''),COALESCE(l.location_notes,''),l.description,COALESCE(l.action_taken,''),l.follow_up_status,u.name FROM security_shift_hourly_logs l JOIN security_shift_reports r ON r.id=l.report_id LEFT JOIN master_area a ON a.id=l.area_id JOIN users u ON u.id=l.created_by_user_id WHERE l.level='urgent' AND l.follow_up_status IN ('monitoring','in_progress','forwarded') AND `+storescope.Clause(ctx, "r.store_id")+` ORDER BY l.reported_at DESC LIMIT 20`)
	if err != nil {
		return d, err
	}
	defer rows.Close()
	for rows.Next() {
		var v HourlyLog
		if err = rows.Scan(&v.ID, &v.CreatedByID, &v.AreaID, &v.ReportedAt, &v.Level, &v.AreaName, &v.LocationNotes, &v.Description, &v.ActionTaken, &v.FollowUpStatus, &v.Creator); err != nil {
			return d, err
		}
		d.Urgent = append(d.Urgent, v)
	}
	return d, rows.Err()
}

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) PageData(ctx context.Context, storeID uint64) (PageData, error) {
	reports, err := r.reports(ctx)
	if err != nil {
		return PageData{}, err
	}
	shifts, err := r.options(ctx, `SELECT id,name,code,COALESCE(TIME_FORMAT(start_time,'%H:%i'),''),COALESCE(TIME_FORMAT(end_time,'%H:%i'),'') FROM master_shift WHERE status='active' ORDER BY sort_order,name`)
	if err != nil {
		return PageData{}, err
	}
	users, err := r.options(ctx, `SELECT DISTINCT u.id,u.name,u.username,'','' FROM users u JOIN user_stores us ON us.user_id=u.id JOIN user_store_roles usr ON usr.user_store_id=us.id JOIN roles ro ON ro.id=usr.role_id WHERE us.store_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL AND ro.name='security' ORDER BY u.name`, storeID)
	if err != nil {
		return PageData{}, err
	}
	supervisors, err := r.options(ctx, `SELECT DISTINCT u.id,u.name,u.username,'','' FROM users u JOIN user_stores us ON us.user_id=u.id JOIN user_store_roles usr ON usr.user_store_id=us.id JOIN roles ro ON ro.id=usr.role_id WHERE us.store_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL AND ro.name IN ('spv_security','superadmin') ORDER BY u.name`, storeID)
	if err != nil {
		return PageData{}, err
	}
	areas, err := r.options(ctx, `SELECT id,name,code,'','' FROM master_area WHERE store_id=? AND status='active' ORDER BY sort_order,name`, storeID)
	return PageData{Reports: reports, Shifts: shifts, Users: users, Supervisors: supervisors, Areas: areas}, err
}

func (r *Repository) reports(ctx context.Context) ([]ListItem, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,r.representative_user_id,DATE_FORMAT(r.report_date,'%d-%m-%Y'),s.name,sh.name,u.name,r.status,DATE_FORMAT(r.work_started_at,'%d-%m-%Y %H:%i'),COALESCE(DATE_FORMAT(r.work_ended_at,'%d-%m-%Y %H:%i'),''),COUNT(DISTINCT m.id),COUNT(DISTINCT l.id),COALESCE(CASE MAX(FIELD(l.level,'normal','waspada','urgent')) WHEN 3 THEN 'urgent' WHEN 2 THEN 'waspada' WHEN 1 THEN 'normal' END,'') FROM security_shift_reports r JOIN stores s ON s.id=r.store_id JOIN master_shift sh ON sh.id=r.shift_id JOIN users u ON u.id=r.representative_user_id LEFT JOIN security_shift_report_members m ON m.report_id=r.id LEFT JOIN security_shift_hourly_logs l ON l.report_id=r.id WHERE `+storescope.Clause(ctx, "r.store_id")+` GROUP BY r.id,s.name,sh.name,u.name ORDER BY r.report_date DESC,r.work_started_at DESC LIMIT 50`)
	if err != nil {
		return nil, fmt.Errorf("list security reports: %w", err)
	}
	defer rows.Close()
	var out []ListItem
	for rows.Next() {
		var v ListItem
		if err := rows.Scan(&v.ID, &v.RepresentativeID, &v.ReportDate, &v.StoreName, &v.ShiftName, &v.RepresentativeName, &v.Status, &v.WorkStartedAt, &v.WorkEndedAt, &v.MemberCount, &v.LogCount, &v.HighestLevel); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) options(ctx context.Context, query string, args ...any) ([]Option, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Option
	for rows.Next() {
		var v Option
		if err := rows.Scan(&v.ID, &v.Name, &v.Code, &v.StartTime, &v.EndTime); err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

func (r *Repository) Create(ctx context.Context, in CreateInput) (id uint64, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var configuredStart sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT TIME_FORMAT(start_time,'%H:%i:%s') FROM master_shift WHERE id=? AND status='active'`, in.ShiftID).Scan(&configuredStart); err != nil {
		return 0, err
	}
	validUser := func(userID uint64) bool {
		var ok bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN users u ON u.id=us.user_id WHERE us.store_id=? AND us.user_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL)`, in.StoreID, userID).Scan(&ok)
		return ok
	}
	if !storescope.Allows(ctx, in.StoreID) || !validUser(in.RepresentativeID) || !validUser(in.CreatorID) {
		return 0, ErrInvalid
	}
	if in.SupervisorID != nil && !validUser(*in.SupervisorID) {
		return 0, ErrInvalid
	}
	date, parseErr := time.Parse("2006-01-02", in.ReportDate)
	if parseErr != nil {
		return 0, ErrInvalid
	}
	startTime := time.Now().Format("15:04:05")
	if configuredStart.Valid && configuredStart.String != "" {
		startTime = configuredStart.String
	}
	startAt, parseErr := time.ParseInLocation("2006-01-02 15:04:05", date.Format("2006-01-02")+" "+startTime, time.Local)
	if parseErr != nil {
		return 0, ErrInvalid
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO security_shift_reports(store_id,shift_id,report_date,representative_user_id,supervisor_user_id,work_started_at,opening_notes,status,created_by_user_id) VALUES(?,?,?,?,?,?,NULLIF(?,''),'active',?)`, in.StoreID, in.ShiftID, in.ReportDate, in.RepresentativeID, in.SupervisorID, startAt, in.OpeningNotes, in.CreatorID)
	if err != nil {
		return 0, err
	}
	insertID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	id = uint64(insertID)
	seen := map[uint64]bool{}
	in.MemberIDs = append(in.MemberIDs, in.RepresentativeID)
	for _, uid := range in.MemberIDs {
		if uid == 0 || seen[uid] {
			continue
		}
		if !validUser(uid) {
			return 0, ErrInvalid
		}
		seen[uid] = true
		role := "member"
		if uid == in.RepresentativeID {
			role = "representative"
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_members(report_id,user_id,member_role,joined_at) VALUES(?,?,?,?)`, id, uid, role, startAt); err != nil {
			return 0, err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,'report_started','draft','active',NULLIF(?,''))`, id, in.CreatorID, in.OpeningNotes)
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	return id, err
}

func (r *Repository) Update(ctx context.Context, in UpdateInput) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var currentRepresentative uint64
	var status string
	err = tx.QueryRowContext(ctx, `SELECT representative_user_id,status FROM security_shift_reports WHERE id=? AND store_id=? AND `+storescope.Clause(ctx, "store_id")+` FOR UPDATE`, in.ReportID, in.StoreID).Scan(&currentRepresentative, &status)
	if err != nil {
		return err
	}
	if currentRepresentative != in.ActorID || (status != "active" && status != "revision") {
		return ErrForbidden
	}
	var start sql.NullString
	if err = tx.QueryRowContext(ctx, `SELECT TIME_FORMAT(start_time,'%H:%i:%s') FROM master_shift WHERE id=? AND status='active'`, in.ShiftID).Scan(&start); err != nil {
		return ErrInvalid
	}
	validUser := func(id uint64) bool {
		var ok bool
		_ = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN users u ON u.id=us.user_id WHERE us.store_id=? AND us.user_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL)`, in.StoreID, id).Scan(&ok)
		return ok
	}
	if !validUser(in.RepresentativeID) {
		return ErrInvalid
	}
	if in.SupervisorID != nil && !validUser(*in.SupervisorID) {
		return ErrInvalid
	}
	date, e := time.Parse("2006-01-02", in.ReportDate)
	if e != nil {
		return ErrInvalid
	}
	startValue := time.Now().Format("15:04:05")
	if start.Valid && start.String != "" {
		startValue = start.String
	}
	startAt, e := time.ParseInLocation("2006-01-02 15:04:05", date.Format("2006-01-02")+" "+startValue, time.Local)
	if e != nil {
		return ErrInvalid
	}
	_, err = tx.ExecContext(ctx, `UPDATE security_shift_reports SET shift_id=?,report_date=?,representative_user_id=?,supervisor_user_id=?,work_started_at=?,opening_notes=NULLIF(?,'') WHERE id=?`, in.ShiftID, in.ReportDate, in.RepresentativeID, in.SupervisorID, startAt, in.OpeningNotes, in.ReportID)
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM security_shift_report_members WHERE report_id=?`, in.ReportID); err != nil {
		return err
	}
	seen := map[uint64]bool{}
	in.MemberIDs = append(in.MemberIDs, in.RepresentativeID)
	for _, uid := range in.MemberIDs {
		if uid == 0 || seen[uid] {
			continue
		}
		if !validUser(uid) {
			return ErrInvalid
		}
		seen[uid] = true
		role := "member"
		if uid == in.RepresentativeID {
			role = "representative"
		}
		if _, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_members(report_id,user_id,member_role,joined_at) VALUES(?,?,?,?)`, in.ReportID, uid, role, startAt); err != nil {
			return err
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,notes) VALUES(?,?,'report_updated','Identitas dan anggota shift diperbarui')`, in.ReportID, in.ActorID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Detail(ctx context.Context, id uint64) (Detail, error) {
	var d Detail
	err := r.db.QueryRowContext(ctx, `SELECT r.id,r.shift_id,r.representative_user_id,COALESCE(r.supervisor_user_id,0),DATE_FORMAT(r.report_date,'%Y-%m-%d'),s.name,sh.name,u.name,COALESCE(sp.name,''),TIME_FORMAT(r.work_started_at,'%H:%i'),COALESCE(TIME_FORMAT(r.work_ended_at,'%H:%i'),''),COALESCE(r.opening_notes,''),COALESCE(r.closing_summary,''),r.status,COALESCE(CASE MAX(FIELD(l.level,'normal','waspada','urgent')) WHEN 3 THEN 'urgent' WHEN 2 THEN 'waspada' WHEN 1 THEN 'normal' END,''),COALESCE(SUM(l.level='normal'),0),COALESCE(SUM(l.level='waspada'),0),COALESCE(SUM(l.level='urgent'),0) FROM security_shift_reports r JOIN stores s ON s.id=r.store_id JOIN master_shift sh ON sh.id=r.shift_id JOIN users u ON u.id=r.representative_user_id LEFT JOIN users sp ON sp.id=r.supervisor_user_id LEFT JOIN security_shift_hourly_logs l ON l.report_id=r.id WHERE r.id=? AND `+storescope.Clause(ctx, "r.store_id")+` GROUP BY r.id,s.name,sh.name,u.name,sp.name`, id).Scan(&d.ID, &d.ShiftID, &d.RepresentativeID, &d.SupervisorID, &d.RawReportDate, &d.StoreName, &d.ShiftName, &d.RepresentativeName, &d.SupervisorName, &d.WorkStartedAt, &d.WorkEndedAt, &d.OpeningNotes, &d.ClosingSummary, &d.Status, &d.HighestLevel, &d.NormalCount, &d.WaspadaCount, &d.UrgentCount)
	if err != nil {
		return d, err
	}
	rows, err := r.db.QueryContext(ctx, `SELECT m.id,m.user_id,u.name,m.member_role,COALESCE(DATE_FORMAT(m.joined_at,'%H:%i'),'') FROM security_shift_report_members m JOIN users u ON u.id=m.user_id WHERE m.report_id=? ORDER BY m.member_role='representative' DESC,u.name`, id)
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var m Member
		if err = rows.Scan(&m.ID, &m.UserID, &m.Name, &m.Role, &m.JoinedAt); err != nil {
			rows.Close()
			return d, err
		}
		d.Members = append(d.Members, m)
	}
	if err = rows.Close(); err != nil {
		return d, err
	}
	logs, err := r.db.QueryContext(ctx, `SELECT l.id,l.created_by_user_id,COALESCE(l.area_id,0),DATE_FORMAT(l.reported_at,'%d-%m-%Y %H:%i'),l.level,COALESCE(a.name,''),COALESCE(l.location_notes,''),l.description,COALESCE(l.action_taken,''),l.follow_up_status,u.name FROM security_shift_hourly_logs l LEFT JOIN master_area a ON a.id=l.area_id JOIN users u ON u.id=l.created_by_user_id WHERE l.report_id=? ORDER BY l.reported_at`, id)
	if err != nil {
		return d, err
	}
	defer logs.Close()
	for logs.Next() {
		var v HourlyLog
		if err = logs.Scan(&v.ID, &v.CreatedByID, &v.AreaID, &v.ReportedAt, &v.Level, &v.AreaName, &v.LocationNotes, &v.Description, &v.ActionTaken, &v.FollowUpStatus, &v.Creator); err != nil {
			return d, err
		}
		d.Logs = append(d.Logs, v)
	}
	if err = logs.Err(); err != nil {
		return d, err
	}
	var h Handover
	var hasGiver, hasReceiver, hasSPV bool
	var giverSignature, receiverSignature, spvSignature string
	err = r.db.QueryRowContext(ctx, `SELECT h.id,COALESCE(h.received_by_user_id,0),COALESCE(h.supervisor_user_id,0),giver.name,COALESCE(receiver.name,''),COALESCE(supervisor.name,''),COALESCE(h.handover_notes,''),COALESCE(h.pending_items,''),COALESCE(h.inventory_notes,''),h.status,COALESCE(DATE_FORMAT(h.handed_over_at,'%d-%m-%Y %H:%i'),''),COALESCE(DATE_FORMAT(h.received_at,'%d-%m-%Y %H:%i'),''),COALESCE(DATE_FORMAT(h.supervisor_signed_at,'%d-%m-%Y %H:%i'),''),h.giver_signature IS NOT NULL,h.receiver_signature IS NOT NULL,h.supervisor_signature IS NOT NULL,COALESCE(h.giver_signature,''),COALESCE(h.receiver_signature,''),COALESCE(h.supervisor_signature,'') FROM security_shift_handovers h JOIN users giver ON giver.id=h.handed_over_by_user_id LEFT JOIN users receiver ON receiver.id=h.received_by_user_id LEFT JOIN users supervisor ON supervisor.id=h.supervisor_user_id WHERE h.from_report_id=?`, id).Scan(&h.ID, &h.ReceiverUserID, &h.SupervisorUserID, &h.GiverName, &h.ReceiverName, &h.SupervisorName, &h.HandoverNotes, &h.PendingItems, &h.InventoryNotes, &h.Status, &h.HandedOverAt, &h.ReceivedAt, &h.SupervisorSignedAt, &hasGiver, &hasReceiver, &hasSPV, &giverSignature, &receiverSignature, &spvSignature)
	if err == nil {
		h.HasGiverSignature, h.HasReceiverSignature, h.HasSPVSignature = hasGiver, hasReceiver, hasSPV
		h.GiverSignature, h.ReceiverSignature, h.SPVSignature = template.URL(giverSignature), template.URL(receiverSignature), template.URL(spvSignature)
		d.Handover = &h
	} else if err != sql.ErrNoRows {
		return d, err
	}
	return d, nil
}

func (r *Repository) CreateHourlyLog(ctx context.Context, in HourlyLogInput) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var storeID uint64
	var status string
	err = tx.QueryRowContext(ctx, `SELECT store_id,status FROM security_shift_reports WHERE id=? AND `+storescope.Clause(ctx, "store_id")+` FOR UPDATE`, in.ReportID).Scan(&storeID, &status)
	if err != nil {
		return err
	}
	if status != "active" {
		return ErrInvalid
	}
	var member bool
	if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM security_shift_report_members WHERE report_id=? AND user_id=?)`, in.ReportID, in.CreatorID).Scan(&member); err != nil {
		return err
	}
	if !member {
		return ErrForbidden
	}
	if in.AreaID != nil {
		var validArea bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_area WHERE id=? AND store_id=? AND status='active')`, *in.AreaID, storeID).Scan(&validArea); err != nil {
			return err
		}
		if !validArea {
			return ErrInvalid
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_hourly_logs(report_id,reported_at,level,area_id,location_notes,description,action_taken,follow_up_status,created_by_user_id) VALUES(?,CURRENT_TIMESTAMP(6),?,?,NULLIF(?,''),?,NULLIF(?,''),?,?)`, in.ReportID, in.Level, in.AreaID, in.LocationNotes, in.Description, in.ActionTaken, in.FollowUpStatus, in.CreatorID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,notes) VALUES(?,?,'hourly_log_created',?)`, in.ReportID, in.CreatorID, "Laporan per jam level "+in.Level)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) UpdateHourlyLog(ctx context.Context, in HourlyLogInput) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var storeID, representativeID, createdByID uint64
	var status string
	err = tx.QueryRowContext(ctx, `SELECT r.store_id,r.representative_user_id,r.status,l.created_by_user_id FROM security_shift_hourly_logs l JOIN security_shift_reports r ON r.id=l.report_id WHERE l.id=? AND l.report_id=? AND `+storescope.Clause(ctx, "r.store_id")+` FOR UPDATE`, in.LogID, in.ReportID).Scan(&storeID, &representativeID, &status, &createdByID)
	if err != nil {
		return err
	}
	if status != "active" || (in.CreatorID != createdByID && in.CreatorID != representativeID) {
		return ErrForbidden
	}
	if in.AreaID != nil {
		var valid bool
		if err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM master_area WHERE id=? AND store_id=? AND status='active')`, *in.AreaID, storeID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return ErrInvalid
		}
	}
	_, err = tx.ExecContext(ctx, `UPDATE security_shift_hourly_logs SET level=?,area_id=?,location_notes=NULLIF(?,''),description=?,action_taken=NULLIF(?,''),follow_up_status=? WHERE id=? AND report_id=?`, in.Level, in.AreaID, in.LocationNotes, in.Description, in.ActionTaken, in.FollowUpStatus, in.LogID, in.ReportID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,notes) VALUES(?,?,'hourly_log_updated',?)`, in.ReportID, in.CreatorID, "Laporan per jam #"+fmt.Sprint(in.LogID)+" diperbarui")
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) SubmitHandover(ctx context.Context, in HandoverInput) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var representativeID, storeID uint64
	var status string
	err = tx.QueryRowContext(ctx, `SELECT representative_user_id,store_id,status FROM security_shift_reports WHERE id=? AND `+storescope.Clause(ctx, "store_id")+` FOR UPDATE`, in.ReportID).Scan(&representativeID, &storeID, &status)
	if err != nil {
		return err
	}
	if representativeID != in.ActorID || (status != "active" && status != "revision") {
		return ErrForbidden
	}
	var validReceiver bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN users u ON u.id=us.user_id JOIN user_store_roles usr ON usr.user_store_id=us.id JOIN roles ro ON ro.id=usr.role_id WHERE us.store_id=? AND us.user_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL AND ro.name='security')`, storeID, in.ReceiverID).Scan(&validReceiver)
	if err != nil {
		return err
	}
	if !validReceiver || in.ReceiverID == in.ActorID {
		return ErrInvalid
	}
	var validSupervisor bool
	err = tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN users u ON u.id=us.user_id JOIN user_store_roles usr ON usr.user_store_id=us.id JOIN roles ro ON ro.id=usr.role_id WHERE us.store_id=? AND us.user_id=? AND us.status='active' AND u.status='active' AND u.deleted_at IS NULL AND ro.name IN ('spv_security','superadmin'))`, storeID, in.SupervisorID).Scan(&validSupervisor)
	if err != nil {
		return err
	}
	if !validSupervisor {
		return ErrInvalid
	}
	if status == "active" {
		_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_handovers(from_report_id,handed_over_by_user_id,received_by_user_id,supervisor_user_id,handover_notes,pending_items,inventory_notes,giver_signature,handed_over_at,status) VALUES(?,?,?,?,NULLIF(?,''),NULLIF(?,''),NULLIF(?,''),?,CURRENT_TIMESTAMP(6),'waiting_receiver')`, in.ReportID, in.ActorID, in.ReceiverID, in.SupervisorID, in.HandoverNotes, in.PendingItems, in.InventoryNotes, in.Signature)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE security_shift_handovers SET received_by_user_id=?,supervisor_user_id=?,handover_notes=NULLIF(?,''),pending_items=NULLIF(?,''),inventory_notes=NULLIF(?,''),giver_signature=?,receiver_signature=NULL,supervisor_signature=NULL,handed_over_at=CURRENT_TIMESTAMP(6),received_at=NULL,supervisor_signed_at=NULL,status='waiting_receiver' WHERE from_report_id=?`, in.ReceiverID, in.SupervisorID, in.HandoverNotes, in.PendingItems, in.InventoryNotes, in.Signature, in.ReportID)
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE security_shift_reports SET supervisor_user_id=?,closing_summary=NULLIF(?,''),work_ended_at=CURRENT_TIMESTAMP(6),status='waiting_handover' WHERE id=?`, in.SupervisorID, in.ClosingSummary, in.ReportID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,'handover_submitted',?,'waiting_handover',NULLIF(?,''))`, in.ReportID, in.ActorID, status, in.HandoverNotes)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ReceiveHandover(ctx context.Context, in ReceiveInput) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var receiverID uint64
	var handoverStatus, reportStatus string
	err = tx.QueryRowContext(ctx, `SELECT h.received_by_user_id,h.status,r.status FROM security_shift_handovers h JOIN security_shift_reports r ON r.id=h.from_report_id WHERE h.from_report_id=? AND `+storescope.Clause(ctx, "r.store_id")+` FOR UPDATE`, in.ReportID).Scan(&receiverID, &handoverStatus, &reportStatus)
	if err != nil {
		return err
	}
	if receiverID != in.ActorID || handoverStatus != "waiting_receiver" || reportStatus != "waiting_handover" {
		return ErrForbidden
	}
	toStatus, eventType := "revision", "handover_rejected"
	if in.Accept {
		toStatus, eventType = "waiting_supervisor", "handover_received"
		_, err = tx.ExecContext(ctx, `UPDATE security_shift_handovers SET receiver_signature=?,received_at=CURRENT_TIMESTAMP(6),status='waiting_supervisor' WHERE from_report_id=?`, in.Signature, in.ReportID)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE security_shift_handovers SET status='revision' WHERE from_report_id=?`, in.ReportID)
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE security_shift_reports SET status=? WHERE id=?`, toStatus, in.ReportID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,?,'waiting_handover',?,NULLIF(?,''))`, in.ReportID, in.ActorID, eventType, toStatus, in.Notes)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) ReviewHandover(ctx context.Context, in ReviewInput) (err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()
	var supervisorID uint64
	var handoverStatus, reportStatus string
	err = tx.QueryRowContext(ctx, `SELECT h.supervisor_user_id,h.status,r.status FROM security_shift_handovers h JOIN security_shift_reports r ON r.id=h.from_report_id WHERE h.from_report_id=? AND `+storescope.Clause(ctx, "r.store_id")+` FOR UPDATE`, in.ReportID).Scan(&supervisorID, &handoverStatus, &reportStatus)
	if err != nil {
		return err
	}
	if (!in.AllowAny && supervisorID != in.ActorID) || handoverStatus != "waiting_supervisor" || reportStatus != "waiting_supervisor" {
		return ErrForbidden
	}
	toStatus, eventType := "revision", "supervisor_revision_requested"
	if in.Approve {
		toStatus, eventType = "completed", "supervisor_approved"
		_, err = tx.ExecContext(ctx, `UPDATE security_shift_handovers SET supervisor_user_id=?,supervisor_signature=?,supervisor_signed_at=CURRENT_TIMESTAMP(6),status='completed' WHERE from_report_id=?`, in.ActorID, in.Signature, in.ReportID)
	} else {
		_, err = tx.ExecContext(ctx, `UPDATE security_shift_handovers SET status='revision' WHERE from_report_id=?`, in.ReportID)
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE security_shift_reports SET status=?,completed_at=IF(?='completed',CURRENT_TIMESTAMP(6),NULL) WHERE id=?`, toStatus, toStatus, in.ReportID)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO security_shift_report_events(report_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,?,'waiting_supervisor',?,NULLIF(?,''))`, in.ReportID, in.ActorID, eventType, toStatus, in.Notes)
	if err != nil {
		return err
	}
	return tx.Commit()
}
