package sessionmanagement

import (
	"context"
	"fmt"
	"strings"
	"time"
)

var jakartaLocation = func() *time.Location {
	location, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return location
}()

type Service struct{ repository *Repository }

func NewService(repository *Repository) *Service { return &Service{repository: repository} }
func (s *Service) Sessions(ctx context.Context, userID uint64, currentID string) ([]Session, error) {
	items, err := s.repository.Sessions(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	for index := range items {
		items[index].Current = items[index].ID == currentID
		items[index].Browser, items[index].Device, items[index].Icon = agentInfo(items[index].UserAgent)
		items[index].Created = formatTime(items[index].CreatedAt)
		items[index].LastSeen = formatTime(items[index].LastSeenAt)
		items[index].Expires = formatTime(items[index].ExpiresAt)
		items[index].Remaining = remaining(now, items[index].ExpiresAt)
	}
	return items, nil
}
func (s *Service) Revoke(ctx context.Context, userID uint64, id, currentID string) error {
	return s.repository.Revoke(ctx, userID, id, currentID)
}
func (s *Service) RevokeOthers(ctx context.Context, userID uint64, currentID string) error {
	return s.repository.RevokeOthers(ctx, userID, currentID)
}
func agentInfo(agent string) (string, string, string) {
	browser := "Browser tidak dikenal"
	switch {
	case strings.Contains(agent, "Edg/"):
		browser = "Microsoft Edge"
	case strings.Contains(agent, "Chrome/"):
		browser = "Google Chrome"
	case strings.Contains(agent, "Firefox/"):
		browser = "Mozilla Firefox"
	case strings.Contains(agent, "Safari/"):
		browser = "Safari"
	}
	device, icon := "Perangkat tidak dikenal", "fa-desktop"
	switch {
	case strings.Contains(agent, "Android"):
		device, icon = "Perangkat Android", "fa-mobile-screen-button"
	case strings.Contains(agent, "iPhone"):
		device, icon = "iPhone", "fa-mobile-screen-button"
	case strings.Contains(agent, "iPad"):
		device, icon = "iPad", "fa-tablet-screen-button"
	case strings.Contains(agent, "Windows"):
		device, icon = "Komputer Windows", "fa-desktop"
	case strings.Contains(agent, "Macintosh"):
		device, icon = "Komputer Mac", "fa-laptop"
	case strings.Contains(agent, "Linux"):
		device, icon = "Komputer Linux", "fa-desktop"
	}
	return browser, device, icon
}
func formatTime(value time.Time) string {
	return value.In(jakartaLocation).Format("02 Jan 2006, 15:04") + " WIB"
}
func remaining(now, expiry time.Time) string {
	duration := expiry.Sub(now)
	if duration <= 0 {
		return "Berakhir"
	}
	days := int(duration.Hours() / 24)
	if days > 0 {
		return fmt.Sprintf("%d hari lagi", days)
	}
	hours := int(duration.Hours())
	if hours > 0 {
		return fmt.Sprintf("%d jam lagi", hours)
	}
	return fmt.Sprintf("%d menit lagi", int(duration.Minutes()))
}
