package securityreport

import (
	"context"
	"errors"
	"strings"
	"time"
)

var ErrInvalid = errors.New("invalid security report")
var ErrForbidden = errors.New("not a report member")
var ErrSupervisorRequired = errors.New("supervisor required")

type Service struct{ repo *Repository }

func NewService(r *Repository) *Service { return &Service{repo: r} }
func (s *Service) PageData(ctx context.Context, storeID uint64) (PageData, error) {
	return s.repo.PageData(ctx, storeID)
}
func (s *Service) History(ctx context.Context, storeID uint64, f HistoryFilter) (HistoryData, error) {
	if _, e := time.Parse("2006-01-02", f.StartDate); e != nil {
		return HistoryData{}, ErrInvalid
	}
	if _, e := time.Parse("2006-01-02", f.EndDate); e != nil {
		return HistoryData{}, ErrInvalid
	}
	validStatus := map[string]bool{"": true, "active": true, "waiting_handover": true, "waiting_supervisor": true, "revision": true, "completed": true}
	validLevel := map[string]bool{"": true, "normal": true, "waspada": true, "urgent": true}
	if !validStatus[f.Status] || !validLevel[f.Level] {
		return HistoryData{}, ErrInvalid
	}
	return s.repo.History(ctx, storeID, f)
}
func (s *Service) Monitoring(ctx context.Context) (MonitoringData, error) {
	return s.repo.Monitoring(ctx)
}
func (s *Service) Detail(ctx context.Context, id uint64) (Detail, error) {
	if id == 0 {
		return Detail{}, ErrInvalid
	}
	detail, err := s.repo.Detail(ctx, id)
	if err != nil {
		return Detail{}, err
	}
	detail.ReportDate = detail.RawReportDate
	if date, parseErr := time.Parse("2006-01-02", detail.ReportDate); parseErr == nil {
		months := [...]string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
		detail.ReportDate = date.Format("02") + " " + months[date.Month()-1] + " " + date.Format("2006")
	}
	return detail, nil
}

func (s *Service) Update(ctx context.Context, in UpdateInput) error {
	in.ReportDate = strings.TrimSpace(in.ReportDate)
	in.OpeningNotes = strings.TrimSpace(in.OpeningNotes)
	if in.ReportID == 0 || in.ActorID == 0 || in.StoreID == 0 || in.ShiftID == 0 || in.RepresentativeID == 0 || in.ReportDate == "" {
		return ErrInvalid
	}
	return s.repo.Update(ctx, in)
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

func (s *Service) UpdateHourlyLog(ctx context.Context, in HourlyLogInput) error {
	in.Level = strings.TrimSpace(in.Level)
	in.LocationNotes = strings.TrimSpace(in.LocationNotes)
	in.Description = strings.TrimSpace(in.Description)
	in.ActionTaken = strings.TrimSpace(in.ActionTaken)
	in.FollowUpStatus = strings.TrimSpace(in.FollowUpStatus)
	if in.ReportID == 0 || in.LogID == 0 || in.CreatorID == 0 || in.Description == "" {
		return ErrInvalid
	}
	if in.Level != "normal" && in.Level != "waspada" && in.Level != "urgent" {
		return ErrInvalid
	}
	valid := map[string]bool{"none": true, "monitoring": true, "in_progress": true, "completed": true, "forwarded": true}
	if !valid[in.FollowUpStatus] || (in.Level != "normal" && in.ActionTaken == "") {
		return ErrInvalid
	}
	return s.repo.UpdateHourlyLog(ctx, in)
}

func (s *Service) SubmitHandover(ctx context.Context, in HandoverInput) error {
	in.ClosingSummary = strings.TrimSpace(in.ClosingSummary)
	in.HandoverNotes = strings.TrimSpace(in.HandoverNotes)
	in.PendingItems = strings.TrimSpace(in.PendingItems)
	in.InventoryNotes = strings.TrimSpace(in.InventoryNotes)
	in.Signature = strings.TrimSpace(in.Signature)
	if in.ReportID == 0 || in.ActorID == 0 || in.ReceiverID == 0 || in.SupervisorID == 0 || in.ClosingSummary == "" || in.HandoverNotes == "" || !validSignature(in.Signature) {
		return ErrInvalid
	}
	return s.repo.SubmitHandover(ctx, in)
}
func (s *Service) ReceiveHandover(ctx context.Context, in ReceiveInput) error {
	in.Notes = strings.TrimSpace(in.Notes)
	in.Signature = strings.TrimSpace(in.Signature)
	if in.ReportID == 0 || in.ActorID == 0 || (!in.Accept && in.Notes == "") || (in.Accept && !validSignature(in.Signature)) {
		return ErrInvalid
	}
	return s.repo.ReceiveHandover(ctx, in)
}
func (s *Service) ReviewHandover(ctx context.Context, in ReviewInput) error {
	in.Notes = strings.TrimSpace(in.Notes)
	in.Signature = strings.TrimSpace(in.Signature)
	if in.ReportID == 0 || in.ActorID == 0 || (!in.Approve && in.Notes == "") || (in.Approve && !validSignature(in.Signature)) {
		return ErrInvalid
	}
	return s.repo.ReviewHandover(ctx, in)
}
func validSignature(value string) bool {
	return strings.HasPrefix(value, "data:image/png;base64,") && len(value) > 100
}
