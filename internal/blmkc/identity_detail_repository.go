package blmkc

import (
	"context"
	"fmt"

	"mksiaga/internal/storescope"
)

func (r *Repository) IdentityDetail(ctx context.Context, id uint64) (IdentityDetail, error) {
	var item IdentityDetail
	err := r.db.QueryRowContext(ctx, `SELECT sub.id,sub.subject_number,
		COALESCE(NULLIF(sub.full_name,''),'Belum diketahui'),
		COALESCE((SELECT GROUP_CONCAT(sa.alias_name ORDER BY sa.alias_name SEPARATOR ', ') FROM blmkc_subject_aliases sa WHERE sa.subject_id=sub.id),'-'),
		sub.identity_status,COALESCE(NULLIF(sub.gender,''),'-'),
		COALESCE(NULLIF(CONCAT_WS(', ',NULLIF(sub.birth_place,''),DATE_FORMAT(sub.birth_date,'%d/%m/%Y')),''),'-'),
		COALESCE(NULLIF(sub.address,''),'-'),COALESCE(NULLIF(sub.phone,''),'-'),COALESCE(NULLIF(sub.occupation,''),'-'),
		COALESCE(NULLIF(sub.identification_notes,''),'-'),COALESCE(verifier.name,'-'),
		COALESCE(DATE_FORMAT(sub.verified_at,'%d/%m/%Y %H:%i'),'-'),DATE_FORMAT(sub.created_at,'%d/%m/%Y %H:%i'),
		COUNT(DISTINCT c.id),COUNT(DISTINCT c.store_id),
		SUBSTRING_INDEX(GROUP_CONCAT(c.id ORDER BY c.occurred_at DESC,c.id DESC),',',1),
		SUBSTRING_INDEX(GROUP_CONCAT(c.case_status ORDER BY c.occurred_at DESC,c.id DESC),',',1)
	FROM blmkc_subjects sub
	JOIN blmkc_case_subjects cs ON cs.subject_id=sub.id
	JOIN blmkc_cases c ON c.id=cs.case_id
	LEFT JOIN users verifier ON verifier.id=sub.verified_by_user_id
	WHERE sub.id=? AND sub.archived_at IS NULL AND c.archived_at IS NULL
		AND `+storescope.Clause(ctx, "c.store_id")+`
	GROUP BY sub.id,sub.subject_number,sub.full_name,sub.identity_status,sub.gender,sub.birth_place,sub.birth_date,
		sub.address,sub.phone,sub.occupation,sub.identification_notes,verifier.name,sub.verified_at,sub.created_at`, id).Scan(
		&item.ID, &item.SubjectNumber, &item.FullName, &item.Aliases, &item.IdentityStatus, &item.Gender,
		&item.BirthInformation, &item.Address, &item.Phone, &item.Occupation, &item.IdentificationNotes,
		&item.VerifiedBy, &item.VerifiedAt, &item.CreatedAt, &item.CaseCount, &item.StoreCount, &item.PackageCaseID,
		&item.PackageCaseStatus)
	if err != nil {
		return item, err
	}

	item.Cases, err = r.queryCases(ctx, ` AND EXISTS(SELECT 1 FROM blmkc_case_subjects ids WHERE ids.case_id=c.id AND ids.subject_id=?)`, []any{id}, 100, 0)
	if err != nil {
		return item, err
	}
	if err := r.identityEvidence(ctx, id, &item); err != nil {
		return item, err
	}
	if err := r.identityCasePhotos(ctx, id, &item); err != nil {
		return item, err
	}
	return item, nil
}

func (r *Repository) identityCasePhotos(ctx context.Context, id uint64, item *IdentityDetail) error {
	rows, err := r.db.QueryContext(ctx, `SELECT e.id,e.case_id,COALESCE(e.description,'')
		FROM blmkc_evidence_files e
		JOIN blmkc_cases c ON c.id=e.case_id
		JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.subject_id=?
		WHERE e.deleted_at IS NULL AND c.archived_at IS NULL
		AND e.mime_type LIKE 'image/%'
		AND e.file_category IN ('stolen_item_photo','incident_photo','subject_photo','vehicle_photo')
		AND `+storescope.Clause(ctx, "c.store_id")+`
		ORDER BY e.created_at DESC,e.id DESC`, id)
	if err != nil {
		return fmt.Errorf("load BLMKC identity case photos: %w", err)
	}
	defer rows.Close()
	item.CasePhotos = make(map[uint64][]IdentityCasePhoto)
	for rows.Next() {
		var photo IdentityCasePhoto
		if err := rows.Scan(&photo.ID, &photo.CaseID, &photo.Description); err != nil {
			return err
		}
		item.CasePhotos[photo.CaseID] = append(item.CasePhotos[photo.CaseID], photo)
	}
	return rows.Err()
}

func (r *Repository) identityVehicles(ctx context.Context, id uint64, item *IdentityDetail) error {
	rows, err := r.db.QueryContext(ctx, `SELECT c.id,c.case_number,v.vehicle_type,COALESCE(v.brand,'-'),COALESCE(v.model,'-'),
		COALESCE(v.color,'-'),COALESCE(v.license_plate,'-')
	FROM blmkc_vehicles v JOIN blmkc_cases c ON c.id=v.case_id
	WHERE v.subject_id=? AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id")+`
	ORDER BY c.occurred_at DESC,v.id DESC`, id)
	if err != nil {
		return fmt.Errorf("load BLMKC identity vehicles: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var value IdentityVehicle
		if err := rows.Scan(&value.CaseID, &value.CaseNumber, &value.Type, &value.Brand, &value.Model, &value.Color, &value.Plate); err != nil {
			return err
		}
		item.Vehicles = append(item.Vehicles, value)
	}
	return rows.Err()
}

func (r *Repository) identityEvidence(ctx context.Context, id uint64, item *IdentityDetail) error {
	rows, err := r.db.QueryContext(ctx, `SELECT e.id,c.id,c.case_number,e.original_filename,e.mime_type,COALESCE(e.description,''),
		e.file_size,DATE_FORMAT(e.created_at,'%d/%m/%Y %H:%i')
	FROM blmkc_evidence_files e JOIN blmkc_cases c ON c.id=e.case_id
	WHERE e.subject_id=? AND e.deleted_at IS NULL AND c.archived_at IS NULL AND `+storescope.Clause(ctx, "c.store_id")+`
	ORDER BY e.created_at DESC,e.id DESC`, id)
	if err != nil {
		return fmt.Errorf("load BLMKC identity evidence: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var value IdentityEvidence
		var size uint64
		if err := rows.Scan(&value.ID, &value.CaseID, &value.CaseNumber, &value.Filename, &value.MIMEType,
			&value.Description, &size, &value.CreatedAt); err != nil {
			return err
		}
		value.SizeLabel = evidenceSizeLabel(size)
		item.Evidence = append(item.Evidence, value)
	}
	return rows.Err()
}
