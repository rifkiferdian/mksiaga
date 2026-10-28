package securityreport

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"mksiaga/internal/storescope"
)

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
	rows, err := r.db.QueryContext(ctx, `SELECT r.id,DATE_FORMAT(r.report_date,'%d-%m-%Y'),s.name,sh.name,u.name,r.status,DATE_FORMAT(r.work_started_at,'%d-%m-%Y %H:%i'),COALESCE(DATE_FORMAT(r.work_ended_at,'%d-%m-%Y %H:%i'),''),COUNT(DISTINCT m.id),COUNT(DISTINCT l.id),COALESCE(CASE MAX(FIELD(l.level,'normal','waspada','urgent')) WHEN 3 THEN 'urgent' WHEN 2 THEN 'waspada' WHEN 1 THEN 'normal' END,'') FROM security_shift_reports r JOIN stores s ON s.id=r.store_id JOIN master_shift sh ON sh.id=r.shift_id JOIN users u ON u.id=r.representative_user_id LEFT JOIN security_shift_report_members m ON m.report_id=r.id LEFT JOIN security_shift_hourly_logs l ON l.report_id=r.id WHERE `+storescope.Clause(ctx, "r.store_id")+` GROUP BY r.id,s.name,sh.name,u.name ORDER BY r.report_date DESC,r.work_started_at DESC LIMIT 50`)
	if err != nil {
		return nil, fmt.Errorf("list security reports: %w", err)
	}
	defer rows.Close()
	var out []ListItem
	for rows.Next() {
		var v ListItem
		if err := rows.Scan(&v.ID, &v.ReportDate, &v.StoreName, &v.ShiftName, &v.RepresentativeName, &v.Status, &v.WorkStartedAt, &v.WorkEndedAt, &v.MemberCount, &v.LogCount, &v.HighestLevel); err != nil {
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

func (r *Repository) Detail(ctx context.Context, id uint64) (Detail, error) {
	var d Detail
	err := r.db.QueryRowContext(ctx, `SELECT r.id,r.representative_user_id,DATE_FORMAT(r.report_date,'%Y-%m-%d'),s.name,sh.name,u.name,COALESCE(sp.name,''),TIME_FORMAT(r.work_started_at,'%H:%i'),COALESCE(TIME_FORMAT(r.work_ended_at,'%H:%i'),''),COALESCE(r.opening_notes,''),COALESCE(r.closing_summary,''),r.status,COALESCE(CASE MAX(FIELD(l.level,'normal','waspada','urgent')) WHEN 3 THEN 'urgent' WHEN 2 THEN 'waspada' WHEN 1 THEN 'normal' END,''),COALESCE(SUM(l.level='normal'),0),COALESCE(SUM(l.level='waspada'),0),COALESCE(SUM(l.level='urgent'),0) FROM security_shift_reports r JOIN stores s ON s.id=r.store_id JOIN master_shift sh ON sh.id=r.shift_id JOIN users u ON u.id=r.representative_user_id LEFT JOIN users sp ON sp.id=r.supervisor_user_id LEFT JOIN security_shift_hourly_logs l ON l.report_id=r.id WHERE r.id=? AND `+storescope.Clause(ctx, "r.store_id")+` GROUP BY r.id,s.name,sh.name,u.name,sp.name`, id).Scan(&d.ID, &d.RepresentativeID, &d.ReportDate, &d.StoreName, &d.ShiftName, &d.RepresentativeName, &d.SupervisorName, &d.WorkStartedAt, &d.WorkEndedAt, &d.OpeningNotes, &d.ClosingSummary, &d.Status, &d.HighestLevel, &d.NormalCount, &d.WaspadaCount, &d.UrgentCount)
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
	logs, err := r.db.QueryContext(ctx, `SELECT l.id,DATE_FORMAT(l.reported_at,'%d-%m-%Y %H:%i'),l.level,COALESCE(a.name,''),COALESCE(l.location_notes,''),l.description,COALESCE(l.action_taken,''),l.follow_up_status,u.name FROM security_shift_hourly_logs l LEFT JOIN master_area a ON a.id=l.area_id JOIN users u ON u.id=l.created_by_user_id WHERE l.report_id=? ORDER BY l.reported_at`, id)
	if err != nil {
		return d, err
	}
	defer logs.Close()
	for logs.Next() {
		var v HourlyLog
		if err = logs.Scan(&v.ID, &v.ReportedAt, &v.Level, &v.AreaName, &v.LocationNotes, &v.Description, &v.ActionTaken, &v.FollowUpStatus, &v.Creator); err != nil {
			return d, err
		}
		d.Logs = append(d.Logs, v)
	}
	return d, logs.Err()
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
