package blmkc

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"time"

	"mksiaga/internal/storescope"
)

func (r *Repository) CaseForm(ctx context.Context, id uint64) (CaseFormData, error) {
	data := CaseFormData{CaseID: id, IsEdit: id > 0}
	stores, err := storescope.Stores(ctx, r.db)
	if err != nil {
		return data, err
	}
	for _, store := range stores {
		data.Stores = append(data.Stores, FormOption{ID: store.ID, Code: store.Code, Name: store.Name})
	}
	data.IncidentTypes, err = r.formOptions(ctx, `SELECT id,0,code,name FROM master_jenis_insiden WHERE status='active' AND code='THEFT' ORDER BY sort_order,name`)
	if err != nil {
		return data, err
	}
	data.Areas, err = r.formOptions(ctx, `SELECT id,store_id,code,name FROM master_area WHERE status='active' AND `+storescope.Clause(ctx, "store_id")+` ORDER BY store_id,sort_order,name`)
	if err != nil {
		return data, err
	}
	data.Cameras, err = r.formOptions(ctx, `SELECT id,store_id,code,name FROM master_kamera WHERE status='active' AND `+storescope.Clause(ctx, "store_id")+` ORDER BY store_id,name`)
	if err != nil {
		return data, err
	}
	if id == 0 {
		if len(data.Stores) == 1 {
			data.Case.StoreID = data.Stores[0].ID
		}
		data.Case.OccurredAt = time.Now().Format("2006-01-02T15:04")
		data.Case.InitialStatus = "wanted"
		if len(data.IncidentTypes) == 1 {
			data.SelectedIncidentTypeID = data.IncidentTypes[0].ID
			data.Case.IncidentTypeID = &data.IncidentTypes[0].ID
		}
		return data, nil
	}
	var incidentID, areaID, cameraID sql.NullInt64
	var loss sql.NullFloat64
	err = r.db.QueryRowContext(ctx, `SELECT c.case_number,c.store_id,c.incident_type_id,c.area_id,c.camera_id,
		DATE_FORMAT(c.occurred_at,'%Y-%m-%dT%H:%i'),c.title,c.chronology,c.loss_amount,c.case_status,
		COALESCE((SELECT sub.full_name FROM blmkc_case_subjects cs JOIN blmkc_subjects sub ON sub.id=cs.subject_id WHERE cs.case_id=c.id ORDER BY cs.is_primary DESC,cs.id LIMIT 1),'')
		FROM blmkc_cases c WHERE c.id=? AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id"), id).Scan(
		&data.CaseNumber, &data.Case.StoreID, &incidentID, &areaID, &cameraID, &data.Case.OccurredAt,
		&data.Case.Title, &data.Case.Chronology, &loss, &data.Case.InitialStatus, &data.Case.SubjectName)
	if err != nil {
		return data, err
	}
	data.Case.IncidentTypeID = nullUint64(incidentID)
	data.Case.AreaID = nullUint64(areaID)
	data.Case.CameraID = nullUint64(cameraID)
	if data.Case.IncidentTypeID != nil {
		data.SelectedIncidentTypeID = *data.Case.IncidentTypeID
	}
	if data.Case.AreaID != nil {
		data.SelectedAreaID = *data.Case.AreaID
	}
	if data.Case.CameraID != nil {
		data.SelectedCameraID = *data.Case.CameraID
	}
	if loss.Valid {
		value := loss.Float64
		data.Case.LossAmount = &value
	}
	return data, nil
}

func (r *Repository) formOptions(ctx context.Context, query string) ([]FormOption, error) {
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []FormOption
	for rows.Next() {
		var item FormOption
		if err := rows.Scan(&item.ID, &item.StoreID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func (r *Repository) ValidCaseReferences(ctx context.Context, input CaseInput) (bool, error) {
	if !storescope.Allows(ctx, input.StoreID) {
		return false, nil
	}
	var storeOK, typeOK, areaOK, cameraOK bool
	err := r.db.QueryRowContext(ctx, `SELECT
		EXISTS(SELECT 1 FROM stores WHERE id=? AND status='active' AND deleted_at IS NULL),
		(? IS NOT NULL AND EXISTS(SELECT 1 FROM master_jenis_insiden WHERE id=? AND status='active' AND code='THEFT')),
		(? IS NULL OR EXISTS(SELECT 1 FROM master_area WHERE id=? AND store_id=? AND status='active')),
		(? IS NULL OR EXISTS(SELECT 1 FROM master_kamera WHERE id=? AND store_id=? AND status='active'))`,
		input.StoreID, input.IncidentTypeID, input.IncidentTypeID,
		input.AreaID, input.AreaID, input.StoreID,
		input.CameraID, input.CameraID, input.StoreID).Scan(&storeOK, &typeOK, &areaOK, &cameraOK)
	return storeOK && typeOK && areaOK && cameraOK, err
}

func (r *Repository) CreateCase(ctx context.Context, input CaseInput, userID uint64) (uint64, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	token, err := randomToken()
	if err != nil {
		return 0, err
	}
	result, err := tx.ExecContext(ctx, `INSERT INTO blmkc_cases(
		case_number,store_id,incident_type_id,area_id,camera_id,occurred_at,title,chronology,
		case_status,verification_status,legal_status,loss_amount,reported_by_user_id
	) VALUES(?,?,?,?,?,?,?,?,?,'draft','not_reported',?,?)`, "PENDING-"+token, input.StoreID,
		input.IncidentTypeID, input.AreaID, input.CameraID, input.OccurredAt, input.Title, input.Chronology,
		input.InitialStatus, input.LossAmount, userID)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	var storeCode string
	if err := tx.QueryRowContext(ctx, `SELECT code FROM stores WHERE id=?`, input.StoreID).Scan(&storeCode); err != nil {
		return 0, err
	}
	occurred, _ := time.Parse("2006-01-02 15:04:05", input.OccurredAt)
	caseNumber := fmt.Sprintf("BLMKC/%s/%d/%06d", storeCode, occurred.Year(), id)
	if _, err := tx.ExecContext(ctx, `UPDATE blmkc_cases SET case_number=? WHERE id=?`, caseNumber, id); err != nil {
		return 0, err
	}
	identityStatus := "unknown"
	if input.SubjectName != "" {
		identityStatus = "identified"
	}
	subjectResult, err := tx.ExecContext(ctx, `INSERT INTO blmkc_subjects(subject_number,full_name,identity_status,identification_notes,created_by_user_id)
		VALUES(?,?,?,?,?)`, "PENDING-"+token, nullString(input.SubjectName), identityStatus, "Dibuat bersama kasus "+caseNumber, userID)
	if err != nil {
		return 0, err
	}
	subjectID, err := subjectResult.LastInsertId()
	if err != nil {
		return 0, err
	}
	subjectNumber := fmt.Sprintf("BLMKC-SUB-%06d", subjectID)
	if _, err := tx.ExecContext(ctx, `UPDATE blmkc_subjects SET subject_number=? WHERE id=?`, subjectNumber, subjectID); err != nil {
		return 0, err
	}
	allegation := "unverified"
	if input.InitialStatus == "identified" {
		allegation = "identified"
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_subjects(case_id,subject_id,involvement_role,allegation_status,is_primary) VALUES(?,?,'suspected',?,1)`, id, subjectID, allegation); err != nil {
		return 0, err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,'created','',?,'Kasus dibuat')`, id, userID, input.InitialStatus); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return uint64(id), nil
}

func (r *Repository) UpdateCase(ctx context.Context, id, userID uint64, input CaseInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `UPDATE blmkc_cases SET store_id=?,incident_type_id=?,area_id=?,camera_id=?,occurred_at=?,title=?,chronology=?,loss_amount=?
		WHERE id=? AND archived_at IS NULL AND `+storescope.Clause(ctx, "store_id"), input.StoreID, input.IncidentTypeID,
		input.AreaID, input.CameraID, input.OccurredAt, input.Title, input.Chronology, input.LossAmount, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		var exists bool
		if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM blmkc_cases WHERE id=? AND archived_at IS NULL AND `+storescope.Clause(ctx, "store_id")+`)`, id).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return sql.ErrNoRows
		}
	}
	if input.SubjectName != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE blmkc_subjects sub JOIN blmkc_case_subjects cs ON cs.subject_id=sub.id
			SET sub.full_name=?,sub.identity_status=IF(sub.identity_status='unknown','identified',sub.identity_status)
			WHERE cs.case_id=? AND cs.is_primary=1`, input.SubjectName, id); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes)
		SELECT id,?,'updated',case_status,case_status,'Data kasus diperbarui' FROM blmkc_cases WHERE id=?`, userID, id); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) AdvanceStatus(ctx context.Context, id, userID uint64, subjectName, notes string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var current string
	var subjectID uint64
	var existingName sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT c.case_status,sub.id,sub.full_name FROM blmkc_cases c
		JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.is_primary=1
		JOIN blmkc_subjects sub ON sub.id=cs.subject_id
		WHERE c.id=? AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id")+` FOR UPDATE`, id).Scan(&current, &subjectID, &existingName)
	if err != nil {
		return err
	}
	next := map[string]string{"draft": "wanted", "wanted": "identified", "identified": "apprehended"}[current]
	if next == "" {
		return ErrInvalidTransition
	}
	if next == "identified" {
		if subjectName == "" && (!existingName.Valid || existingName.String == "") {
			return ErrSubjectRequired
		}
		if subjectName != "" {
			if _, err := tx.ExecContext(ctx, `UPDATE blmkc_subjects SET full_name=?,identity_status='identified' WHERE id=?`, subjectName, subjectID); err != nil {
				return err
			}
		}
		if _, err := tx.ExecContext(ctx, `UPDATE blmkc_case_subjects SET allegation_status='identified' WHERE case_id=? AND subject_id=?`, id, subjectID); err != nil {
			return err
		}
	}
	if next == "apprehended" {
		if _, err := tx.ExecContext(ctx, `UPDATE blmkc_subjects SET identity_status='verified',verified_by_user_id=?,verified_at=NOW(6) WHERE id=?`, userID, subjectID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `UPDATE blmkc_case_subjects SET involvement_role='perpetrator',allegation_status='evidence_supported' WHERE case_id=? AND subject_id=?`, id, subjectID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE blmkc_cases SET case_status=? WHERE id=?`, next, id); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,?,?,?,?)`, id, userID, "status_changed", current, next, nullString(notes)); err != nil {
		return err
	}
	return tx.Commit()
}

func randomToken() (string, error) {
	buffer := make([]byte, 12)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullUint64(value sql.NullInt64) *uint64 {
	if !value.Valid {
		return nil
	}
	result := uint64(value.Int64)
	return &result
}
