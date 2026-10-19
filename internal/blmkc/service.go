package blmkc

import "context"

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
