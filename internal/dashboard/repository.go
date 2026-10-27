package dashboard

import (
	"context"
	"database/sql"
	"fmt"
	"html/template"
	"math"
	"strings"
	"time"
)

type Repository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *Repository { return &Repository{db: db} }

func (r *Repository) Load(ctx context.Context, storeID uint64, allStores bool, now time.Time) (Data, error) {
	var data Data
	scope, args := "c.store_id=?", []any{storeID}
	if allStores {
		scope, args = "1=1", nil
	}
	query := `SELECT COUNT(*),COALESCE(SUM(c.verification_status<>'submitted' AND c.case_status IN ('wanted','identified')),0),
		COALESCE(SUM(c.verification_status<>'submitted' AND c.case_status IN ('apprehended','resolved','reported_to_police')),0),
		COALESCE(SUM(c.verification_status='submitted'),0),
		COALESCE(SUM(c.case_status='draft'),0),COALESCE(SUM(c.case_status='wanted'),0),
		COALESCE(SUM(c.case_status='identified'),0),COALESCE(SUM(c.case_status='apprehended'),0),
		COALESCE(SUM(c.case_status='resolved'),0),COALESCE(SUM(c.case_status='reported_to_police'),0)
		FROM blmkc_cases c WHERE c.archived_at IS NULL AND ` + scope
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&data.Summary.Total, &data.Summary.Wanted, &data.Summary.Apprehended, &data.Summary.Pending,
		&data.Composition.Draft, &data.Composition.Wanted, &data.Composition.Identified, &data.Composition.Apprehended,
		&data.Composition.Resolved, &data.Composition.Police); err != nil {
		return data, fmt.Errorf("load dashboard summary: %w", err)
	}
	data.Summary.Other = data.Summary.Total - data.Summary.Wanted - data.Summary.Apprehended - data.Summary.Pending
	if data.Summary.Other < 0 {
		data.Summary.Other = 0
	}

	recordQuery := `SELECT c.id,COALESCE(NULLIF(sub.full_name,''),sub.subject_number),
		COALESCE((SELECT GROUP_CONCAT(sa.alias_name ORDER BY sa.alias_name SEPARATOR ', ') FROM blmkc_subject_aliases sa WHERE sa.subject_id=sub.id),'-'),
		c.case_status,c.verification_status,s.name,DATE_FORMAT(c.occurred_at,'%d %b %Y'),c.title
		FROM blmkc_cases c JOIN stores s ON s.id=c.store_id
		JOIN blmkc_case_subjects cs ON cs.case_id=c.id AND cs.is_primary=1
		JOIN blmkc_subjects sub ON sub.id=cs.subject_id
		WHERE c.archived_at IS NULL AND ` + scope + ` ORDER BY c.occurred_at DESC,c.id DESC LIMIT 5`
	rows, err := r.db.QueryContext(ctx, recordQuery, args...)
	if err != nil {
		return data, fmt.Errorf("load dashboard records: %w", err)
	}
	for rows.Next() {
		var item Record
		var caseStatus, verification string
		if err := rows.Scan(&item.ID, &item.Name, &item.Alias, &caseStatus, &verification, &item.Location, &item.Date, &item.Note); err != nil {
			rows.Close()
			return data, err
		}
		item.Status, item.StatusClass = dashboardStatus(caseStatus)
		item.Verification, item.VerificationClass = verificationStatus(verification)
		item.Initials = initials(item.Name)
		item.Search = strings.ToLower(item.Name + " " + item.Alias + " " + item.Location + " " + item.Note + " " + item.Status + " " + item.Verification)
		data.Records = append(data.Records, item)
	}
	if err := rows.Close(); err != nil {
		return data, err
	}

	activityQuery := `SELECT e.event_type,e.to_status,s.name,DATE_FORMAT(e.created_at,'%d/%m/%Y %H:%i')
		FROM blmkc_case_events e JOIN blmkc_cases c ON c.id=e.case_id JOIN stores s ON s.id=c.store_id
		WHERE c.archived_at IS NULL AND ` + scope + ` ORDER BY e.created_at DESC,e.id DESC LIMIT 3`
	rows, err = r.db.QueryContext(ctx, activityQuery, args...)
	if err != nil {
		return data, fmt.Errorf("load dashboard activities: %w", err)
	}
	for rows.Next() {
		var eventType, toStatus string
		var item Activity
		if err := rows.Scan(&eventType, &toStatus, &item.Location, &item.Time); err != nil {
			rows.Close()
			return data, err
		}
		item.Label, item.Color = activityLabel(eventType, toStatus)
		data.Activities = append(data.Activities, item)
	}
	if err := rows.Close(); err != nil {
		return data, err
	}

	monthCounts := make(map[string][2]int)
	trendQuery := `SELECT DATE_FORMAT(c.occurred_at,'%Y-%m'),COUNT(*),
		COALESCE(SUM(c.case_status IN ('apprehended','resolved','reported_to_police')),0)
		FROM blmkc_cases c WHERE c.archived_at IS NULL AND c.occurred_at>=? AND ` + scope + ` GROUP BY DATE_FORMAT(c.occurred_at,'%Y-%m')`
	start := time.Date(now.Year(), now.Month()-5, 1, 0, 0, 0, 0, now.Location())
	trendArgs := append([]any{start.Format("2006-01-02")}, args...)
	rows, err = r.db.QueryContext(ctx, trendQuery, trendArgs...)
	if err != nil {
		return data, fmt.Errorf("load dashboard trend: %w", err)
	}
	for rows.Next() {
		var key string
		var total, handled int
		if err := rows.Scan(&key, &total, &handled); err != nil {
			rows.Close()
			return data, err
		}
		monthCounts[key] = [2]int{total, handled}
	}
	if err := rows.Close(); err != nil {
		return data, err
	}
	var totals, handled []int
	for index := 0; index < 6; index++ {
		month := start.AddDate(0, index, 0)
		data.MonthLabels = append(data.MonthLabels, monthLabel(month.Month()))
		count := monthCounts[month.Format("2006-01")]
		totals, handled = append(totals, count[0]), append(handled, count[1])
	}
	data.TotalPoints = chartPoints(totals)
	data.HandledPoints = chartPoints(handled)
	data.AreaPath = "0,160 " + data.TotalPoints + " 700,160"
	data.Gradient = statusGradient(data.Summary.Total, data.Composition)
	return data, nil
}

func dashboardStatus(status string) (string, string) {
	values := map[string][2]string{
		"draft":              {"Draft", "bg-slate-100 text-slate-600"},
		"wanted":             {"Wanted", "bg-red-50 text-red-700"},
		"identified":         {"Teridentifikasi", "bg-blue-50 text-blue-700"},
		"apprehended":        {"Diamankan", "bg-purple-50 text-purple-700"},
		"resolved":           {"Selesai", "bg-emerald-50 text-emerald-700"},
		"reported_to_police": {"Dilaporkan ke polisi", "bg-indigo-50 text-indigo-700"},
	}
	value, ok := values[status]
	if !ok {
		return status, "bg-slate-100 text-slate-600"
	}
	return value[0], value[1]
}

func verificationStatus(status string) (string, string) {
	values := map[string][2]string{
		"draft":     {"Belum diajukan", "text-slate-500"},
		"submitted": {"Menunggu", "text-amber-700"},
		"revision":  {"Revisi", "text-orange-700"},
		"verified":  {"Terverifikasi", "text-emerald-700"},
		"rejected":  {"Ditolak", "text-red-700"},
	}
	value, ok := values[status]
	if !ok {
		return status, "text-slate-500"
	}
	return value[0], value[1]
}

func activityLabel(eventType, status string) (string, string) {
	switch eventType {
	case "created":
		return "Kasus baru ditambahkan", "bg-orange-500"
	case "verification_review":
		return "Verifikasi kasus: " + status, "bg-amber-500"
	default:
		return "Status diperbarui: " + status, "bg-emerald-500"
	}
}

func chartPoints(values []int) string {
	maximum := 1
	for _, value := range values {
		if value > maximum {
			maximum = value
		}
	}
	points := make([]string, len(values))
	for index, value := range values {
		x := float64(index) * 140
		y := 145 - (float64(value)/float64(maximum))*115
		points[index] = fmt.Sprintf("%.0f,%.0f", x, math.Max(20, y))
	}
	return strings.Join(points, " ")
}

func statusGradient(total int, composition Composition) template.CSS {
	if total == 0 {
		return "conic-gradient(#e2e8f0 0 100%)"
	}
	values := []int{composition.Draft, composition.Wanted, composition.Identified, composition.Apprehended, composition.Resolved, composition.Police}
	colors := []string{"#cbd5e1", "#ef4444", "#3b82f6", "#a855f7", "#10b981", "#6366f1"}
	parts := make([]string, 0, len(values))
	start := 0.0
	for index, value := range values {
		end := start + float64(value)*100/float64(total)
		if value > 0 {
			parts = append(parts, fmt.Sprintf("%s %.2f%% %.2f%%", colors[index], start, end))
		}
		start = end
	}
	if start < 100 {
		parts = append(parts, fmt.Sprintf("#64748b %.2f%% 100%%", start))
	}
	return template.CSS("conic-gradient(" + strings.Join(parts, ",") + ")")
}

func monthLabel(month time.Month) string {
	return [...]string{"Jan", "Feb", "Mar", "Apr", "Mei", "Jun", "Jul", "Agu", "Sep", "Okt", "Nov", "Des"}[month-1]
}
