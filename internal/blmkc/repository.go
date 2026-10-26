package blmkc

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"mksiaga/internal/storescope"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Dashboard(ctx context.Context) (DashboardData, error) {
	var data DashboardData
	row := r.db.QueryRowContext(ctx, `SELECT
		COUNT(*),
		COALESCE(SUM(case_status='wanted'),0),
		COALESCE(SUM(case_status='identified'),0),
		COALESCE(SUM(case_status='apprehended'),0),
		COALESCE(SUM(case_status='resolved'),0),
		COALESCE(SUM(case_status='reported_to_police'),0),
		COALESCE(SUM(verification_status='submitted'),0),
		COALESCE(SUM(loss_amount),0)
	FROM blmkc_cases WHERE archived_at IS NULL AND `+storescope.Clause(ctx, "store_id"))
	if err := row.Scan(&data.Summary.Total, &data.Summary.Wanted, &data.Summary.Identified,
		&data.Summary.Apprehended, &data.Summary.Resolved, &data.Summary.PoliceReported,
		&data.Summary.PendingReview, &data.Summary.LossAmount); err != nil {
		return data, fmt.Errorf("load BLMKC summary: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `SELECT case_status,COUNT(*) FROM blmkc_cases
		WHERE archived_at IS NULL AND `+storescope.Clause(ctx, "store_id")+`
		GROUP BY case_status ORDER BY COUNT(*) DESC,case_status`)
	if err != nil {
		return data, fmt.Errorf("load BLMKC status counts: %w", err)
	}
	for rows.Next() {
		var item StatusCount
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			rows.Close()
			return data, fmt.Errorf("scan BLMKC status count: %w", err)
		}
		data.StatusCounts = append(data.StatusCounts, item)
	}
	if err := rows.Close(); err != nil {
		return data, err
	}
	if err := rows.Err(); err != nil {
		return data, err
	}

	data.RecentCases, err = r.queryCases(ctx, "", nil, 8, 0)
	if err != nil {
		return data, err
	}
	return data, nil
}

func (r *Repository) Cases(ctx context.Context, filter Filter) (CaseListData, error) {
	where, args := filterClause(ctx, filter)
	var data CaseListData
	data.Filter = filter
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM blmkc_cases c WHERE c.archived_at IS NULL AND `+where, args...).Scan(&data.Total); err != nil {
		return data, fmt.Errorf("count BLMKC cases: %w", err)
	}
	offset := (filter.Page - 1) * filter.PageSize
	items, err := r.queryCases(ctx, " AND "+where, args, filter.PageSize, offset)
	if err != nil {
		return data, err
	}
	data.Cases = items
	data.TotalPages = (data.Total + filter.PageSize - 1) / filter.PageSize
	if data.TotalPages == 0 {
		data.TotalPages = 1
	}
	if filter.Page > 1 {
		data.PrevPage = filter.Page - 1
	}
	if filter.Page < data.TotalPages {
		data.NextPage = filter.Page + 1
	}
	stores, err := storescope.Stores(ctx, r.db)
	if err != nil {
		return data, fmt.Errorf("load BLMKC stores: %w", err)
	}
	for _, store := range stores {
		data.Stores = append(data.Stores, Option{ID: store.ID, Code: store.Code, Name: store.Name})
	}
	return data, nil
}

func (r *Repository) Wanted(ctx context.Context, search, status string) (WantedData, error) {
	data := WantedData{Search: search, Status: status}
	if err := r.db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(case_status='wanted'),0),COALESCE(SUM(case_status='identified'),0)
		FROM blmkc_cases WHERE archived_at IS NULL AND case_status IN ('wanted','identified') AND `+storescope.Clause(ctx, "store_id")).Scan(&data.WantedCount, &data.IdentifiedCount); err != nil {
		return data, fmt.Errorf("count BLMKC wanted: %w", err)
	}
	clauses := []string{"c.case_status IN ('wanted','identified')"}
	args := make([]any, 0, 8)
	if status == "wanted" || status == "identified" {
		clauses = append(clauses, "c.case_status=?")
		args = append(args, status)
	}
	if search != "" {
		term := "%" + search + "%"
		clauses = append(clauses, `(c.case_number LIKE ? OR c.title LIKE ? OR EXISTS(
			SELECT 1 FROM blmkc_case_subjects cs JOIN blmkc_subjects sub ON sub.id=cs.subject_id
			WHERE cs.case_id=c.id AND (sub.full_name LIKE ? OR sub.subject_number LIKE ?)) OR EXISTS(
			SELECT 1 FROM blmkc_vehicles v WHERE v.case_id=c.id AND v.license_plate_normalized LIKE ?))`)
		args = append(args, term, term, term, term, term)
	}
	items, err := r.queryCases(ctx, " AND "+strings.Join(clauses, " AND "), args, 100, 0)
	if err != nil {
		return data, err
	}
	data.Cases = items
	return data, nil
}

func (r *Repository) Identities(ctx context.Context, search string) (IdentityData, error) {
	data := IdentityData{Search: search}
	query := `SELECT sub.id,sub.subject_number,COALESCE(NULLIF(sub.full_name,''),'Belum diketahui'),
		COALESCE((SELECT GROUP_CONCAT(sa.alias_name ORDER BY sa.alias_name SEPARATOR ', ') FROM blmkc_subject_aliases sa WHERE sa.subject_id=sub.id),'-'),
		sub.identity_status,COUNT(DISTINCT c.id),COUNT(DISTINCT c.store_id),
		DATE_FORMAT(MAX(c.occurred_at),'%d/%m/%Y'),
		SUBSTRING_INDEX(GROUP_CONCAT(c.case_status ORDER BY c.occurred_at DESC,c.id DESC),',',1)
	FROM blmkc_subjects sub
	JOIN blmkc_case_subjects cs ON cs.subject_id=sub.id
	JOIN blmkc_cases c ON c.id=cs.case_id
	WHERE sub.archived_at IS NULL AND c.archived_at IS NULL
		AND c.case_status IN ('apprehended','resolved','reported_to_police')
		AND ` + storescope.Clause(ctx, "c.store_id")
	args := make([]any, 0, 3)
	if search != "" {
		term := "%" + search + "%"
		query += ` AND (sub.full_name LIKE ? OR sub.subject_number LIKE ? OR EXISTS(SELECT 1 FROM blmkc_subject_aliases sa2 WHERE sa2.subject_id=sub.id AND sa2.alias_name LIKE ?))`
		args = append(args, term, term, term)
	}
	query += ` GROUP BY sub.id,sub.subject_number,sub.full_name,sub.identity_status ORDER BY MAX(c.occurred_at) DESC,sub.id DESC LIMIT 100`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return data, fmt.Errorf("list BLMKC identities: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item Identity
		if err := rows.Scan(&item.ID, &item.SubjectNumber, &item.FullName, &item.Aliases, &item.IdentityStatus,
			&item.CaseCount, &item.StoreCount, &item.LastCaseDate, &item.LastCaseStatus); err != nil {
			return data, fmt.Errorf("scan BLMKC identity: %w", err)
		}
		data.Subjects = append(data.Subjects, item)
	}
	if err := rows.Err(); err != nil {
		return data, err
	}
	data.Total = len(data.Subjects)
	return data, nil
}

func (r *Repository) Verifications(ctx context.Context, search, status string) (VerificationData, error) {
	data := VerificationData{Search: search, Status: status}
	if err := r.db.QueryRowContext(ctx, `SELECT
		COALESCE(SUM(verification_status='submitted'),0),COALESCE(SUM(verification_status='revision'),0),
		COALESCE(SUM(verification_status='verified'),0),COALESCE(SUM(verification_status='rejected'),0)
		FROM blmkc_cases WHERE archived_at IS NULL AND `+storescope.Clause(ctx, "store_id")).Scan(&data.Submitted, &data.Revision, &data.Verified, &data.Rejected); err != nil {
		return data, fmt.Errorf("count BLMKC verifications: %w", err)
	}
	clauses := []string{"c.verification_status=?"}
	args := []any{status}
	if search != "" {
		term := "%" + search + "%"
		clauses = append(clauses, "(c.case_number LIKE ? OR c.title LIKE ?)")
		args = append(args, term, term)
	}
	items, err := r.queryCases(ctx, " AND "+strings.Join(clauses, " AND "), args, 100, 0)
	if err != nil {
		return data, err
	}
	data.Cases = items
	return data, nil
}

func (r *Repository) queryCases(ctx context.Context, extraWhere string, args []any, limit, offset int) ([]Case, error) {
	query := `SELECT c.id,
		COALESCE((SELECT bcs.subject_id FROM blmkc_case_subjects bcs
			WHERE bcs.case_id=c.id ORDER BY bcs.is_primary DESC,bcs.id LIMIT 1),0),
		c.case_number,s.name,c.title,
		DATE_FORMAT(c.occurred_at,'%d/%m/%Y %H:%i'),c.case_status,c.verification_status,
		COALESCE(j.name,'Belum ditentukan'),
		COALESCE((SELECT COALESCE(NULLIF(bs.full_name,''),bs.subject_number)
			FROM blmkc_case_subjects bcs JOIN blmkc_subjects bs ON bs.id=bcs.subject_id
			WHERE bcs.case_id=c.id ORDER BY bcs.is_primary DESC,bcs.id LIMIT 1),'Belum ada subjek'),
		(SELECT COUNT(*) FROM blmkc_case_items ci WHERE ci.case_id=c.id),
		(SELECT COUNT(*) FROM blmkc_evidence_files ef WHERE ef.case_id=c.id AND ef.deleted_at IS NULL),
		COALESCE((SELECT ef.id FROM blmkc_evidence_files ef
			WHERE ef.case_id=c.id AND ef.deleted_at IS NULL
			AND ef.file_category IN ('stolen_item_photo','incident_photo','subject_photo','vehicle_photo')
			ORDER BY ef.created_at DESC,ef.id DESC LIMIT 1),0)
	FROM blmkc_cases c
	JOIN stores s ON s.id=c.store_id
	LEFT JOIN master_jenis_insiden j ON j.id=c.incident_type_id
	WHERE c.archived_at IS NULL AND ` + storescope.Clause(ctx, "c.store_id") + extraWhere + `
	ORDER BY c.occurred_at DESC,c.id DESC LIMIT ? OFFSET ?`
	queryArgs := append(append([]any{}, args...), limit, offset)
	rows, err := r.db.QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("list BLMKC cases: %w", err)
	}
	defer rows.Close()
	var result []Case
	for rows.Next() {
		var item Case
		if err := rows.Scan(&item.ID, &item.PrimarySubjectID, &item.CaseNumber, &item.StoreName, &item.Title, &item.OccurredAt,
			&item.CaseStatus, &item.VerificationStatus, &item.IncidentType, &item.PrimarySubject,
			&item.ItemCount, &item.EvidenceCount, &item.PrimaryEvidenceID); err != nil {
			return nil, fmt.Errorf("scan BLMKC case: %w", err)
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func filterClause(ctx context.Context, filter Filter) (string, []any) {
	clauses := []string{storescope.Clause(ctx, "c.store_id")}
	args := make([]any, 0, 8)
	if filter.Search != "" {
		term := "%" + filter.Search + "%"
		clauses = append(clauses, `(c.case_number LIKE ? OR c.title LIKE ? OR c.chronology LIKE ?
			OR EXISTS(SELECT 1 FROM blmkc_case_subjects cs JOIN blmkc_subjects sub ON sub.id=cs.subject_id WHERE cs.case_id=c.id AND (sub.full_name LIKE ? OR sub.subject_number LIKE ?))
			OR EXISTS(SELECT 1 FROM blmkc_vehicles v WHERE v.case_id=c.id AND v.license_plate_normalized LIKE ?)
			OR EXISTS(SELECT 1 FROM blmkc_case_items i WHERE i.case_id=c.id AND i.item_name LIKE ?))`)
		args = append(args, term, term, term, term, term, term, term)
	}
	if filter.StoreID > 0 && storescope.Allows(ctx, filter.StoreID) {
		clauses = append(clauses, "c.store_id=?")
		args = append(args, filter.StoreID)
	}
	if filter.CaseStatus != "" {
		clauses = append(clauses, "c.case_status=?")
		args = append(args, filter.CaseStatus)
	}
	if filter.VerificationStatus != "" {
		clauses = append(clauses, "c.verification_status=?")
		args = append(args, filter.VerificationStatus)
	}
	if filter.StartDate != "" {
		clauses = append(clauses, "DATE(c.occurred_at)>=?")
		args = append(args, filter.StartDate)
	}
	if filter.EndDate != "" {
		clauses = append(clauses, "DATE(c.occurred_at)<=?")
		args = append(args, filter.EndDate)
	}
	return strings.Join(clauses, " AND "), args
}
