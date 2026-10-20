package blmkc

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"
)

var (
	ErrInvalid           = errors.New("invalid BLMKC case")
	ErrInvalidTransition = errors.New("invalid BLMKC status transition")
	ErrSubjectRequired   = errors.New("subject identity is required")
)

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }

func (s *Service) Dashboard(ctx context.Context) (DashboardData, error) {
	return s.repository.Dashboard(ctx)
}

func (s *Service) Cases(ctx context.Context, filter Filter) (CaseListData, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	filter.PageSize = 20
	allowedCaseStatus := map[string]bool{"draft": true, "wanted": true, "identified": true, "apprehended": true, "resolved": true, "reported_to_police": true, "archived": true}
	if !allowedCaseStatus[filter.CaseStatus] {
		filter.CaseStatus = ""
	}
	allowedVerification := map[string]bool{"draft": true, "submitted": true, "revision": true, "verified": true, "rejected": true}
	if !allowedVerification[filter.VerificationStatus] {
		filter.VerificationStatus = ""
	}
	return s.repository.Cases(ctx, filter)
}

func (s *Service) Wanted(ctx context.Context, search, status string) (WantedData, error) {
	if status != "wanted" && status != "identified" {
		status = ""
	}
	return s.repository.Wanted(ctx, search, status)
}

func (s *Service) Identities(ctx context.Context, search string) (IdentityData, error) {
	return s.repository.Identities(ctx, search)
}

func (s *Service) Verifications(ctx context.Context, search, status string) (VerificationData, error) {
	allowed := map[string]bool{"submitted": true, "revision": true, "verified": true, "rejected": true}
	if !allowed[status] {
		status = "submitted"
	}
	return s.repository.Verifications(ctx, search, status)
}

func (s *Service) CaseForm(ctx context.Context, id uint64) (CaseFormData, error) {
	return s.repository.CaseForm(ctx, id)
}

func (s *Service) CaseDetail(ctx context.Context, id uint64) (CaseDetail, error) {
	return s.repository.CaseDetail(ctx, id)
}

func (s *Service) CreateCase(ctx context.Context, input CaseInput, userID uint64) (uint64, error) {
	input, err := s.normalizeCase(ctx, input, true)
	if err != nil {
		return 0, err
	}
	return s.repository.CreateCase(ctx, input, userID)
}

func (s *Service) UpdateCase(ctx context.Context, id, userID uint64, input CaseInput) error {
	input, err := s.normalizeCase(ctx, input, false)
	if err != nil {
		return err
	}
	return s.repository.UpdateCase(ctx, id, userID, input)
}

func (s *Service) AdvanceStatus(ctx context.Context, id, userID uint64, subjectName, notes string) error {
	subjectName = strings.TrimSpace(subjectName)
	notes = strings.TrimSpace(notes)
	if utf8.RuneCountInString(subjectName) > 191 || utf8.RuneCountInString(notes) > 2000 {
		return ErrInvalid
	}
	return s.repository.AdvanceStatus(ctx, id, userID, subjectName, notes)
}

func (s *Service) EvidenceUploadContext(ctx context.Context, caseID uint64) (uint64, error) {
	return s.repository.EvidenceUploadContext(ctx, caseID)
}

func (s *Service) CreateEvidence(ctx context.Context, input EvidenceInput) error {
	input.Description = strings.TrimSpace(input.Description)
	if input.CaseID == 0 || input.SubjectID == 0 || input.UserID == 0 || input.FileSize == 0 ||
		input.Category == "" || input.Filename == "" || input.StoragePath == "" || input.MIMEType == "" ||
		len(input.FileHash) != 32 || utf8.RuneCountInString(input.Description) > 500 {
		return ErrInvalid
	}
	return s.repository.CreateEvidence(ctx, input)
}

func (s *Service) Evidence(ctx context.Context, caseID, evidenceID uint64) (EvidenceFile, error) {
	return s.repository.Evidence(ctx, caseID, evidenceID)
}

func (s *Service) VerifyEvidence(ctx context.Context, caseID, evidenceID, userID uint64, approve bool, notes string) error {
	notes = strings.TrimSpace(notes)
	if caseID == 0 || evidenceID == 0 || userID == 0 || utf8.RuneCountInString(notes) > 500 || (!approve && notes == "") {
		return ErrInvalid
	}
	return s.repository.VerifyEvidence(ctx, caseID, evidenceID, userID, approve, notes)
}

func (s *Service) LogEvidenceAccess(ctx context.Context, userID, storeID, caseID uint64, action, ipAddress, userAgent string) error {
	return s.repository.LogEvidenceAccess(ctx, userID, storeID, caseID, action, ipAddress, userAgent)
}

func (s *Service) normalizeCase(ctx context.Context, input CaseInput, creating bool) (CaseInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Chronology = strings.TrimSpace(input.Chronology)
	input.SubjectName = strings.TrimSpace(input.SubjectName)
	parsed, err := time.Parse("2006-01-02T15:04", input.OccurredAt)
	if err != nil {
		return CaseInput{}, ErrInvalid
	}
	input.OccurredAt = parsed.Format("2006-01-02 15:04:05")
	if input.StoreID == 0 || input.IncidentTypeID == nil || input.Title == "" || input.Chronology == "" ||
		utf8.RuneCountInString(input.Title) > 255 || utf8.RuneCountInString(input.SubjectName) > 191 {
		return CaseInput{}, ErrInvalid
	}
	if input.LossAmount != nil && (*input.LossAmount < 0 || *input.LossAmount > 9999999999999.99) {
		return CaseInput{}, ErrInvalid
	}
	if creating {
		if input.InitialStatus != "wanted" && input.InitialStatus != "identified" {
			return CaseInput{}, ErrInvalid
		}
		if input.InitialStatus == "identified" && input.SubjectName == "" {
			return CaseInput{}, ErrSubjectRequired
		}
	}
	valid, err := s.repository.ValidCaseReferences(ctx, input)
	if err != nil {
		return CaseInput{}, err
	}
	if !valid {
		return CaseInput{}, ErrInvalid
	}
	return input, nil
}
