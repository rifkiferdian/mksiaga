package dailyreport

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

var ErrInvalid = errors.New("invalid daily report")
var ErrDuplicate = errors.New("duplicate daily report")

type Service struct{ repository *Repository }

func NewService(r *Repository) *Service { return &Service{repository: r} }
func (s *Service) PageData(ctx context.Context, storeID uint64) (PageData, error) {
	return s.repository.PageData(ctx, storeID)
}
func (s *Service) Create(ctx context.Context, in Input) error {
	in.Location = strings.TrimSpace(in.Location)
	in.StartNotes = strings.TrimSpace(in.StartNotes)
	in.EndNotes = strings.TrimSpace(in.EndNotes)
	in.HandoverNotes = strings.TrimSpace(in.HandoverNotes)
	in.HandoverFromSign = strings.TrimSpace(in.HandoverFromSign)
	in.HandoverToSign = strings.TrimSpace(in.HandoverToSign)
	if in.StoreID == 0 || in.OfficerID == 0 || in.ShiftID == 0 || in.Location == "" || (in.Status != "draft" && in.Status != "submitted") {
		return ErrInvalid
	}
	if !validSignature(in.HandoverFromSign) || !validSignature(in.HandoverToSign) || (in.Status == "submitted" && (in.HandoverToUserID == nil || in.HandoverAt == "" || in.HandoverFromSign == "" || in.HandoverToSign == "")) {
		return ErrInvalid
	}
	if _, e := time.Parse("2006-01-02", in.ReportDate); e != nil {
		return ErrInvalid
	}
	if !validTime(in.WorkStart) || !validTime(in.WorkEnd) {
		return ErrInvalid
	}
	if in.HandoverAt != "" && !validTime(in.HandoverAt) {
		return ErrInvalid
	}
	checklistCount, err := s.repository.ActiveChecklistCount(ctx)
	if err != nil {
		return err
	}
	if len(in.Checklists) != checklistCount {
		return ErrInvalid
	}
	seen := make(map[uint64]bool, len(in.Checklists))
	for _, v := range in.Checklists {
		if v.ChecklistID == 0 || seen[v.ChecklistID] || (v.Answer != "yes" && v.Answer != "no") || (v.Active != nil && v.Total != nil && *v.Active > *v.Total) {
			return ErrInvalid
		}
		seen[v.ChecklistID] = true
	}
	for _, v := range in.Monitoring {
		if !validTime(v.Time) || (v.Condition != "normal" && v.Condition != "finding") {
			return ErrInvalid
		}
	}
	err = s.repository.Create(ctx, in)
	var me *mysql.MySQLError
	if errors.As(err, &me) && me.Number == 1062 {
		return ErrDuplicate
	}
	return err
}
func validTime(v string) bool { _, e := time.Parse("15:04", v); return e == nil }
func validSignature(v string) bool {
	return v == "" || (len(v) <= 700000 && strings.HasPrefix(v, "data:image/png;base64,"))
}
func (s *Service) Delete(ctx context.Context, id, storeID uint64) error {
	return s.repository.Delete(ctx, id, storeID)
}
