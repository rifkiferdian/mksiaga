package securityreport

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid security report")
var ErrForbidden = errors.New("not a report member")

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{repo: r} }
func (s *Service) PageData(ctx context.Context, storeID uint64) (PageData, error) {
	return s.repo.PageData(ctx, storeID)
}
func (s *Service) Detail(ctx context.Context, id uint64) (Detail, error) {
	if id == 0 {
		return Detail{}, ErrInvalid
	}
	detail, err := s.repo.Detail(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	if date, parseErr := time.Parse("2006-01-02", detail.ReportDate); parseErr == nil {
		months := [...]string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		detail.ReportDate = date.Format("02") + " " + months[date.Month()-1] + " " + date.Format("2006")
	}
	return detail, nil
}
func (s *Service) Create(ctx context.Context, in CreateInput) (uint64, error) {
	in.ReportDate = strings.TrimSpace(in.ReportDate)
	in.OpeningNotes = strings.TrimSpace(in.OpeningNotes)
	if in.StoreID == 0 || in.ShiftID == 0 || in.RepresentativeID == 0 || in.CreatorID == 0 || in.ReportDate == "" {
		return 0, ErrInvalid
	}
	return s.repo.Create(ctx, in)
}

func (s *Service) CreateHourlyLog(ctx context.Context, in HourlyLogInput) error {
	in.Level = strings.TrimSpace(in.Level)
	in.LocationNotes = strings.TrimSpace(in.LocationNotes)
	in.Description = strings.TrimSpace(in.Description)
	in.ActionTaken = strings.TrimSpace(in.ActionTaken)
	in.FollowUpStatus = strings.TrimSpace(in.FollowUpStatus)
	if in.ReportID == 0 || in.CreatorID == 0 || in.Description == "" {
		return ErrInvalid
	}
	if in.Level != "normal" && in.Level != "waspada" && in.Level != "urgent" {
		return ErrInvalid
	}
	validFollowUp := map[string]bool{"none": true, "monitoring": true, "in_progress": true, "completed": true, "forwarded": true}
	if !validFollowUp[in.FollowUpStatus] {
		return ErrInvalid
	}
	if in.Level != "normal" && in.ActionTaken == "" {
		return ErrInvalid
	}
	return s.repo.CreateHourlyLog(ctx, in)
}
