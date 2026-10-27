package blmkc

import (
	"context"
	"fmt"

	"mksiaga/internal/storescope"
)

func (r *Repository) CaseDetail(ctx context.Context, id uint64) (CaseDetail, error) {
	var item CaseDetail
	err := r.db.QueryRowContext(ctx, `SELECT c.id,c.case_number,c.store_id,st.name,c.title,c.chronology,
		DATE_FORMAT(c.occurred_at,'%d/%m/%Y %H:%i'),DATE_FORMAT(c.reported_at,'%d/%m/%Y %H:%i'),
		c.case_status,c.verification_status,c.legal_status,
		CASE WHEN c.loss_amount IS NULL THEN '-' ELSE CONCAT('Rp ',FORMAT(c.loss_amount,0,'id_ID')) END,
		COALESCE(it.name,'Belum ditentukan'),COALESCE(a.name,'-'),COALESCE(cam.name,'-'),
		reporter.name,COALESCE(reviewer.name,'-'),COALESCE(DATE_FORMAT(c.reviewed_at,'%d/%m/%Y %H:%i'),'-'),COALESCE(c.review_notes,''),sub.id,sub.subject_number,
		COALESCE(NULLIF(sub.full_name,''),'Belum diketahui'),sub.identity_status,
		COALESCE((SELECT GROUP_CONCAT(sa.alias_name ORDER BY sa.alias_name SEPARATOR ', ') FROM blmkc_subject_aliases sa WHERE sa.subject_id=sub.id),'-')
	FROM blmkc_cases c
	JOIN stores st ON st.id=c.store_id
	JOIN users reporter ON reporter.id=c.reported_by_user_id
	LEFT JOIN users reviewer ON reviewer.id=c.reviewed_by_user_id
	LEFT JOIN master_jenis_insiden it ON it.id=c.incident_type_id
	LEFT JOIN master_area a ON a.id=c.area_id
	LEFT JOIN master_kamera cam ON cam.id=c.camera_id
	JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.is_primary=1
	JOIN blmkc_subjects sub ON sub.id=cs.subject_id
	WHERE c.id=? AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id"), id).Scan(
		&item.ID, &item.CaseNumber, &item.StoreID, &item.StoreName, &item.Title, &item.Chronology,
		&item.OccurredAt, &item.ReportedAt, &item.CaseStatus, &item.VerificationStatus, &item.LegalStatus,
		&item.LossAmount, &item.IncidentType, &item.AreaName, &item.CameraName, &item.ReporterName,
		&item.ReviewerName, &item.ReviewedAt, &item.ReviewNotes, &item.SubjectID, &item.SubjectNumber, &item.SubjectName,
		&item.SubjectIdentityStatus, &item.SubjectAliases)
	if err != nil {
		return item, err
	}
	item.NextStatus = map[string]string{"draft": "wanted", "wanted": "identified", "identified": "apprehended"}[item.CaseStatus]
	item.CanAdvance = item.NextStatus != ""

	rows, err := r.db.QueryContext(ctx, `SELECT item_name,COALESCE(brand,'-'),FORMAT(quantity,2),
		CASE WHEN estimated_value IS NULL THEN '-' ELSE CONCAT('Rp ',FORMAT(estimated_value,0,'id_ID')) END,item_status,COALESCE(notes,'')
		FROM blmkc_case_items WHERE case_id=? ORDER BY id`, id)
	if err != nil {
		return item, fmt.Errorf("load case items: %w", err)
	}
	for rows.Next() {
		var value CaseItemDetail
		if err := rows.Scan(&value.Name, &value.Brand, &value.Quantity, &value.Value, &value.Status, &value.Notes); err != nil {
			rows.Close()
			return item, err
		}
		item.Items = append(item.Items, value)
	}
	if err := rows.Close(); err != nil {
		return item, err
	}

	rows, err = r.db.QueryContext(ctx, `SELECT COALESCE(vehicle_type,'-'),COALESCE(brand,'-'),COALESCE(model,'-'),COALESCE(color,'-'),COALESCE(license_plate,'-'),COALESCE(notes,'') FROM blmkc_vehicles WHERE case_id=? ORDER BY id`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var value VehicleDetail
		if err := rows.Scan(&value.Type, &value.Brand, &value.Model, &value.Color, &value.Plate, &value.Notes); err != nil {
			rows.Close()
			return item, err
		}
		item.Vehicles = append(item.Vehicles, value)
	}
	if err := rows.Close(); err != nil {
		return item, err
	}

	rows, err = r.db.QueryContext(ctx, `SELECT id,file_size,file_category,original_filename,mime_type,COALESCE(description,''),
		DATE_FORMAT(created_at,'%d/%m/%Y %H:%i')
		FROM blmkc_evidence_files WHERE case_id=? AND deleted_at IS NULL ORDER BY created_at DESC,id DESC`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var value EvidenceDetail
		if err := rows.Scan(&value.ID, &value.FileSize, &value.Category, &value.Filename, &value.MIMEType,
			&value.Description, &value.CreatedAt); err != nil {
			rows.Close()
			return item, err
		}
		item.Evidence = append(item.Evidence, value)
		item.Evidence[len(item.Evidence)-1].SizeLabel = evidenceSizeLabel(value.FileSize)
	}
	if err := rows.Close(); err != nil {
		return item, err
	}

	rows, err = r.db.QueryContext(ctx, `SELECT resolution_type,DATE_FORMAT(resolution_date,'%d/%m/%Y %H:%i'),summary,
		COALESCE(police_report_number,'-'),COALESCE(police_station,'-'),
		CASE WHEN compensation_amount IS NULL THEN '-' ELSE CONCAT('Rp ',FORMAT(compensation_amount,0,'id_ID')) END
		FROM blmkc_case_resolutions WHERE case_id=? ORDER BY resolution_date DESC,id DESC`, id)
	if err != nil {
		return item, err
	}
	for rows.Next() {
		var value ResolutionDetail
		if err := rows.Scan(&value.Type, &value.Date, &value.Summary, &value.PoliceNumber, &value.PoliceStation, &value.Compensation); err != nil {
			rows.Close()
			return item, err
		}
		item.Resolutions = append(item.Resolutions, value)
	}
	if err := rows.Close(); err != nil {
		return item, err
	}

	rows, err = r.db.QueryContext(ctx, `SELECT u.name,e.event_type,COALESCE(e.from_status,''),COALESCE(e.to_status,''),COALESCE(e.notes,''),DATE_FORMAT(e.created_at,'%d/%m/%Y %H:%i') FROM blmkc_case_events e JOIN users u ON u.id=e.actor_user_id WHERE e.case_id=? ORDER BY e.created_at DESC,e.id DESC`, id)
	if err != nil {
		return item, err
	}
	defer rows.Close()
	for rows.Next() {
		var value EventDetail
		if err := rows.Scan(&value.Actor, &value.EventType, &value.FromStatus, &value.ToStatus, &value.Notes, &value.CreatedAt); err != nil {
			return item, err
		}
		item.Events = append(item.Events, value)
	}
	return item, rows.Err()
}

func evidenceSizeLabel(size uint64) string {
	if size >= 1<<20 {
		return fmt.Sprintf("%.1f MB", float64(size)/float64(1<<20))
	}
	if size >= 1<<10 {
		return fmt.Sprintf("%.1f KB", float64(size)/float64(1<<10))
	}
	return fmt.Sprintf("%d B", size)
}
