package blmkc

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"mksiaga/internal/storescope"
)

func (s *Service) SaveCaseResolution(ctx context.Context, caseID, userID uint64, input CaseResolutionInput) error {
	input.ResolutionType = strings.TrimSpace(input.ResolutionType)
	input.ResolutionSummary = strings.TrimSpace(input.ResolutionSummary)
	input.PoliceReportNumber = strings.TrimSpace(input.PoliceReportNumber)
	input.PoliceStation = strings.TrimSpace(input.PoliceStation)
	allowed := map[string]bool{"warning": true, "statement": true, "settlement": true, "compensation": true, "handed_to_police": true, "other": true}
	if caseID == 0 || userID == 0 || !allowed[input.ResolutionType] || input.ResolutionDate == "" || input.ResolutionSummary == "" ||
		utf8.RuneCountInString(input.PoliceReportNumber) > 150 || utf8.RuneCountInString(input.PoliceStation) > 191 {
		return ErrInvalid
	}
	parsed, err := time.Parse("2006-01-02T15:04", input.ResolutionDate)
	if err != nil {
		return ErrInvalid
	}
	input.ResolutionDate = parsed.Format("2006-01-02 15:04:05")
	if input.CompensationAmount != nil && (*input.CompensationAmount < 0 || *input.CompensationAmount > 9999999999999.99) {
		return ErrInvalid
	}
	return s.repository.SaveCaseResolution(ctx, caseID, userID, input)
}

func (r *Repository) SaveCaseResolution(ctx context.Context, caseID, userID uint64, input CaseResolutionInput) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var currentStatus string
	err = tx.QueryRowContext(ctx, `SELECT case_status FROM blmkc_cases WHERE id=? AND archived_at IS NULL
		AND case_status IN ('apprehended','resolved','reported_to_police') AND `+storescope.Clause(ctx, "store_id")+` FOR UPDATE`, caseID).Scan(&currentStatus)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_resolutions(case_id,resolution_type,resolution_date,summary,police_report_number,police_station,compensation_amount,created_by_user_id,verified_by_user_id) VALUES(?,?,?,?,?,?,?,?,?)`,
		caseID, input.ResolutionType, input.ResolutionDate, input.ResolutionSummary, nullString(input.PoliceReportNumber), nullString(input.PoliceStation), input.CompensationAmount, userID, userID); err != nil {
		return err
	}
	newStatus := currentStatus
	legalStatus := ""
	switch input.ResolutionType {
	case "settlement", "compensation":
		newStatus, legalStatus = "resolved", "completed"
	case "handed_to_police":
		newStatus, legalStatus = "reported_to_police", "reported"
	}
	if legalStatus != "" {
		if _, err := tx.ExecContext(ctx, `UPDATE blmkc_cases SET case_status=?,legal_status=?,closed_at=CASE WHEN ?='resolved' THEN NOW(6) ELSE closed_at END WHERE id=?`, newStatus, legalStatus, newStatus, caseID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO blmkc_case_events(case_id,actor_user_id,event_type,from_status,to_status,notes) VALUES(?,?,'resolution_recorded',?,?,?)`,
		caseID, userID, currentStatus, newStatus, "Penyelesaian kasus dicatat: "+input.ResolutionType); err != nil {
		return err
	}
	return tx.Commit()
}

func (h *Handler) SaveCaseResolution(c *gin.Context) {
	caseID, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	redirect := "/blmkc/cases/" + strconv.FormatUint(caseID, 10)
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		c.Redirect(http.StatusSeeOther, redirect+"?error=csrf")
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	subjectID, err := h.service.EvidenceUploadContext(c.Request.Context(), caseID)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	input := CaseResolutionInput{
		ResolutionType: c.PostForm("resolution_type"), ResolutionDate: c.PostForm("resolution_date"),
		ResolutionSummary: c.PostForm("resolution_summary"), PoliceReportNumber: c.PostForm("police_report_number"),
		PoliceStation: c.PostForm("police_station"), CompensationAmount: optionalFloat(c.PostForm("compensation_amount")),
	}
	if err := h.service.SaveCaseResolution(c.Request.Context(), caseID, userID, input); err != nil {
		code := "failed"
		if errors.Is(err, ErrInvalid) {
			code = "invalid-resolution"
		} else if errors.Is(err, sql.ErrNoRows) {
			code = "not-found"
		}
		c.Redirect(http.StatusSeeOther, redirect+"?error="+code)
		return
	}
	for _, document := range resolutionDocuments {
		header, err := c.FormFile(document.Field)
		if errors.Is(err, http.ErrMissingFile) {
			continue
		}
		if err != nil {
			c.Redirect(http.StatusSeeOther, redirect+"?error=invalid-identity-document")
			return
		}
		if err := h.saveCaseDocument(c, caseID, subjectID, userID, document, header); err != nil {
			code := "invalid-identity-document"
			if errors.Is(err, errEvidenceTooLarge) {
				code = "identity-document-too-large"
			}
			c.Redirect(http.StatusSeeOther, redirect+"?error="+code)
			return
		}
	}
	c.Redirect(http.StatusSeeOther, redirect+"?success=resolution-saved")
}
