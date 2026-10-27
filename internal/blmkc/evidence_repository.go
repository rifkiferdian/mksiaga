package blmkc

import (
	"context"
	"fmt"
	"strings"

	"mksiaga/internal/storescope"
)

func (r *Repository) EvidenceUploadContext(ctx context.Context, caseID uint64) (uint64, error) {
	var subjectID uint64
	err := r.db.QueryRowContext(ctx, `SELECT sub.id FROM blmkc_cases c
		JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.is_primary=1
		JOIN blmkc_subjects sub ON sub.id=cs.subject_id
		WHERE c.id=? AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id"), caseID).Scan(&subjectID)
	return subjectID, err
}

func (r *Repository) CreateEvidence(ctx context.Context, input EvidenceInput) error {
	return r.CreateEvidenceWithVehicle(ctx, input, nil)
}

func (r *Repository) CreateEvidenceWithVehicle(ctx context.Context, input EvidenceInput, vehicle *VehicleInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	_, err = tx.ExecContext(ctx, `INSERT INTO blmkc_evidence_files(
		case_id,subject_id,file_category,original_filename,storage_path,mime_type,file_size,file_hash,captured_at,description,uploaded_by_user_id
	) VALUES(?,?,?,?,?,?,?,?,?,?,?)`, input.CaseID, input.SubjectID, input.Category, input.Filename,
		input.StoragePath, input.MIMEType, input.FileSize, input.FileHash, input.CapturedAt, nullString(input.Description), input.UserID)
	if err != nil {
		return fmt.Errorf("create BLMKC evidence: %w", err)
	}
	if vehicle != nil {
		normalizedPlate := strings.NewReplacer(" ", "", "-", "", ".", "").Replace(strings.ToUpper(vehicle.Plate))
		if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_vehicles(
			case_id,subject_id,vehicle_type,brand,model,color,license_plate,license_plate_normalized,notes
		) VALUES(?,?,?,?,?,?,?,?,?)`, input.CaseID, input.SubjectID, nullString(vehicle.Type), nullString(vehicle.Brand),
			nullString(vehicle.Model), nullString(vehicle.Color), nullString(vehicle.Plate), nullString(normalizedPlate), nullString(vehicle.Notes)); err != nil {
			return fmt.Errorf("create BLMKC vehicle: %w", err)
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes)
		SELECT id,?,'evidence_uploaded',case_status,case_status,? FROM blmkc_cases WHERE id=?`, input.UserID, "Bukti diunggah: "+input.Filename, input.CaseID); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) Evidence(ctx context.Context, caseID, evidenceID uint64) (EvidenceFile, error) {
	var item EvidenceFile
	err := r.db.QueryRowContext(ctx, `SELECT e.id,e.case_id,c.store_id,e.file_size,e.file_category,e.original_filename,e.storage_path,e.mime_type
		FROM blmkc_evidence_files e JOIN blmkc_cases c ON c.id=e.case_id
		WHERE e.id=? AND e.case_id=? AND e.deleted_at IS NULL AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id"), evidenceID, caseID).Scan(
		&item.ID, &item.CaseID, &item.StoreID, &item.FileSize, &item.Category, &item.Filename, &item.StoragePath, &item.MIMEType)
	return item, err
}

func (r *Repository) LogEvidenceAccess(ctx context.Context, userID, storeID, caseID uint64, action, ipAddress, userAgent string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO blmkc_access_logs(user_id,store_id,case_id,action,ip_address,user_agent) VALUES(?,?,?,?,?,?)`,
		userID, storeID, caseID, action, nullString(ipAddress), nullString(userAgent))
	return err
}
