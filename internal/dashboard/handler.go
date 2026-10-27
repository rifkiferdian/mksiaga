package dashboard

import (
	"database/sql"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
)

type Handler struct {
	appName    string
	navigation *navigation.Service
	repository *Repository
	db         *sql.DB
}

func NewHandler(appName string, db *sql.DB, navigationService *navigation.Service) *Handler {
	return &Handler{appName: appName, navigation: navigationService, repository: NewRepository(db), db: db}
}

func (h *Handler) Index(c *gin.Context) {
	user, _ := auth.CurrentUser(c)
	storeID, err := strconv.ParseUint(user.StoreID, 10, 64)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	allStores := false
	for _, role := range strings.Split(user.RoleName, ",") {
		allStores = allStores || strings.TrimSpace(role) == "superadmin"
	}
	data, err := h.repository.Load(c.Request.Context(), storeID, allStores, time.Now())
	if err != nil {
		slog.Error("load dashboard data", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	canCreate, err := auth.HasPermission(c, h.db, "blmkc.create")
	if err != nil {
		slog.Error("load dashboard permission", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	menus, err := h.navigation.Menus(c.Request.Context(), user, "dashboard")
	if err != nil {
		slog.Error("prepare dashboard navigation", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	csrfToken, err := auth.CSRFToken(c)
	if err != nil {
		slog.Error("prepare dashboard csrf", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	c.HTML(http.StatusOK, "dashboard/index.html", gin.H{
		"Title":         "Dashboard",
		"Section":       "Ruang kerja",
		"AppName":       h.appName,
		"Authenticated": true,
		"IsDashboard":   true,
		"UserName":      user.Name,
		"UserInitials":  initials(user.Name),
		"StoreName":     user.StoreName,
		"RoleName":      user.RoleName,
		"DateLabel":     indonesianDate(time.Now()),
		"Menus":         menus,
		"CSRFToken":     csrfToken,
		"Data":          data,
		"CanCreateCase": canCreate,
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
