package dashboard

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
)

type Handler struct{ appName string }

func NewHandler(appName string) *Handler { return &Handler{appName: appName} }

func (h *Handler) Index(c *gin.Context) {
	user, _ := auth.CurrentUser(c)
	csrfToken, err := auth.CSRFToken(c)
	if err != nil {
		slog.Error("prepare dashboard csrf", "error", err)
		c.String(http.StatusInternalServerError, "Halaman tidak dapat disiapkan.")
		return
	}
	c.HTML(http.StatusOK, "dashboard/index.html", gin.H{
		"Title":         "Beranda",
		"AppName":       h.appName,
		"Authenticated": true,
		"IsDashboard":   true,
		"UserName":      user.Name,
		"UserInitials":  initials(user.Name),
		"StoreName":     user.StoreName,
		"RoleName":      user.RoleName,
		"DateLabel":     indonesianDate(time.Now()),
		"CSRFToken":     csrfToken,
	})
}

func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "US"
	}
	result := []rune(strings.ToUpper(parts[0]))[:1]
	if len(parts) > 1 {
		result = append(result, []rune(strings.ToUpper(parts[len(parts)-1]))[0])
	}
	return string(result)
}

func indonesianDate(value time.Time) string {
	days := [...]string{"Minggu", "Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu"}
	months := [...]string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}
	return strings.ToUpper(days[value.Weekday()] + ", " + value.Format("02") + " " + months[value.Month()-1] + " " + value.Format("2006"))
}
