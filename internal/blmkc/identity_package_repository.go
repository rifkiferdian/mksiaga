package blmkc

import (
	"context"
	"fmt"
	"strings"

	"mksiaga/internal/storescope"
)

func (r *Repository) IdentityPackageForm(ctx context.Context, caseID uint64) (IdentityPackageForm, error) {
	var item IdentityPackageForm
	var hasNIK bool
	err := r.db.QueryRowContext(ctx, `SELECT c.id,c.case_number,c.title,c.case_status,st.name,sub.id,sub.subject_number,
		COALESCE(sub.full_name,''),sub.nik_hash IS NOT NULL,
		COALESCE((SELECT GROUP_CONCAT(sa.alias_name ORDER BY sa.alias_name SEPARATOR ', ') FROM blmkc_subject_aliases sa WHERE sa.subject_id=sub.id),''),
		COALESCE(sub.birth_place,''),COALESCE(DATE_FORMAT(sub.birth_date,'%Y-%m-%d'),''),COALESCE(sub.gender,''),
		COALESCE(sub.address,''),COALESCE(sub.phone,''),COALESCE(sub.occupation,''),COALESCE(sub.identification_notes,'')
	FROM blmkc_cases c JOIN stores st ON st.id=c.store_id
	JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.is_primary=1
	JOIN blmkc_subjects sub ON sub.id=cs.subject_id
	WHERE c.id=? AND c.archived_at IS NULL AND c.case_status IN ('apprehended','resolved','reported_to_police')
		AND `+storescope.Clause(ctx, "c.store_id"), caseID).Scan(
		&item.CaseID, &item.CaseNumber, &item.CaseTitle, &item.CaseStatus, &item.StoreName, &item.SubjectID,
		&item.SubjectNumber, &item.FullName, &hasNIK, &item.Aliases, &item.BirthPlace, &item.BirthDate,
		&item.Gender, &item.Address, &item.Phone, &item.Occupation, &item.IdentificationNotes)
	item.HasNIK = hasNIK
	if hasNIK {
		item.NIKMasked = "••••••••••••••••"
	}
	return item, err
}

func (r *Repository) SaveIdentityPackage(ctx context.Context, caseID, userID uint64, input IdentityPackageInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var subjectID uint64
	var currentStatus string
	err = tx.QueryRowContext(ctx, `SELECT sub.id,c.case_status FROM blmkc_cases c
		JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.is_primary=1
		JOIN blmkc_subjects sub ON sub.id=cs.subject_id
		WHERE c.id=? AND c.archived_at IS NULL AND c.case_status IN ('apprehended','resolved','reported_to_police')
		AND `+storescope.Clause(ctx, "c.store_id")+` FOR UPDATE`, caseID).Scan(&subjectID, &currentStatus)
	if err != nil {
		return err
	}
	query := `UPDATE blmkc_subjects SET full_name=?,birth_place=?,birth_date=?,gender=?,address=?,phone=?,occupation=?,
		identity_status='verified',identification_notes=?,verified_by_user_id=?,verified_at=NOW(6)`
	args := []any{input.FullName, nullString(input.BirthPlace), nullString(input.BirthDate), nullString(input.Gender),
		nullString(input.Address), nullString(input.Phone), nullString(input.Occupation), nullString(input.IdentificationNotes), userID}
	if len(input.NIKEncrypted) > 0 {
		query += `,nik_encrypted=?,nik_hash=?`
		args = append(args, input.NIKEncrypted, input.NIKHash)
	}
	query += ` WHERE id=?`
	args = append(args, subjectID)
	if _, err := tx.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("update BLMKC identity: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM blmkc_subject_aliases WHERE subject_id=?`, subjectID); err != nil {
		return err
	}
	seen := map[string]bool{}
	for _, alias := range strings.Split(input.Aliases, ",") {
		alias = strings.TrimSpace(alias)
		if alias == "" || seen[strings.ToLower(alias)] {
			continue
		}
		if len([]rune(alias)) > 191 {
			return ErrInvalid
		}
		seen[strings.ToLower(alias)] = true
		if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_subject_aliases(subject_id,alias_name,notes) VALUES(?,?,'Dicatat melalui kelengkapan Identity')`, subjectID, alias); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE blmkc_case_subjects SET involvement_role='perpetrator',allegation_status='evidence_supported' WHERE case_id=? AND subject_id=?`, caseID, subjectID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,'identity_completed',?,?,?)`,
		caseID, userID, currentStatus, currentStatus, "Data identitas pelaku diperbarui."); err != nil {
		return err
	}
	return tx.Commit()
}
