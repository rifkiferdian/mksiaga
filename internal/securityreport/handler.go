package securityreport

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
)

type Handler struct {
	appName    string
	service    *Service
	navigation *navigation.Service
}

func NewHandler(appName string, s *Service, n *navigation.Service) *Handler {
	return &Handler{appName: appName, service: s, navigation: n}
}

func (h *Handler) Index(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, err := strconv.ParseUint(u.StoreID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	data, err := h.service.PageData(c.Request.Context(), storeID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "securityreport/index.html", "Laporan Shift Security", gin.H{"Data": data})
}
func (h *Handler) New(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, err := strconv.ParseUint(u.StoreID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	userID, err := strconv.ParseUint(u.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	data, err := h.service.PageData(c.Request.Context(), storeID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "securityreport/create.html", "Mulai Laporan Shift", gin.H{"Data": data, "Today": time.Now().Format("2006-01-02"), "CurrentUserID": userID})
}
func (h *Handler) Create(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/reports/security/new", "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	storeID, e1 := strconv.ParseUint(u.StoreID, 10, 64)
	creatorID, e2 := strconv.ParseUint(u.ID, 10, 64)
	shiftID, e3 := strconv.ParseUint(c.PostForm("shift_id"), 10, 64)
	representativeID, e4 := strconv.ParseUint(c.PostForm("representative_user_id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		h.redirect(c, "/reports/security/new", "error", "invalid")
		return
	}
	var supervisorID *uint64
	if raw := c.PostForm("supervisor_user_id"); raw != "" {
		v, e := strconv.ParseUint(raw, 10, 64)
		if e != nil {
			h.redirect(c, "/reports/security/new", "error", "invalid")
			return
		}
		supervisorID = &v
	}
	var members []uint64
	for _, raw := range c.PostFormArray("member_ids[]") {
		v, e := strconv.ParseUint(raw, 10, 64)
		if e == nil && v > 0 {
			members = append(members, v)
		}
	}
	id, err := h.service.Create(c.Request.Context(), CreateInput{StoreID: storeID, ShiftID: shiftID, RepresentativeID: representativeID, CreatorID: creatorID, SupervisorID: supervisorID, ReportDate: c.PostForm("report_date"), OpeningNotes: c.PostForm("opening_notes"), MemberIDs: members})
	if err != nil {
		code := "failed"
		var mysqlErr *mysql.MySQLError
		if errors.Is(err, ErrInvalid) {
			code = "invalid"
		} else if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			code = "duplicate"
		} else {
			slog.Error("create security report", "error", err)
		}
		h.redirect(c, "/reports/security/new", "error", code)
		return
	}
	h.redirect(c, "/reports/security/"+strconv.FormatUint(id, 10), "success", "created")
}
func (h *Handler) Show(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, storeErr := strconv.ParseUint(u.StoreID, 10, 64)
	userID, userErr := strconv.ParseUint(u.ID, 10, 64)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || storeErr != nil || userErr != nil {
		h.redirect(c, "/reports/security", "error", "not-found")
		return
	}
	d, err := h.service.Detail(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			h.redirect(c, "/reports/security", "error", "not-found")
			return
		}
		h.internalError(c, err)
		return
	}
	data, err := h.service.PageData(c.Request.Context(), storeID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	canAdd := false
	for _, member := range d.Members {
		if member.UserID == userID {
			canAdd = true
			break
		}
	}
	h.render(c, "securityreport/detail.html", "Detail Laporan Shift", gin.H{"Report": d, "Areas": data.Areas, "CanAddLog": canAdd && d.Status == "active"})
}

func (h *Handler) AddHourlyLog(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectToReport(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	reportID, e1 := strconv.ParseUint(c.Param("id"), 10, 64)
	creatorID, e2 := strconv.ParseUint(u.ID, 10, 64)
	if e1 != nil || e2 != nil {
		h.redirectToReport(c, "error", "invalid-log")
		return
	}
	var areaID *uint64
	if raw := c.PostForm("area_id"); raw != "" {
		value, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			h.redirectToReport(c, "error", "invalid-log")
			return
		}
		areaID = &value
	}
	err := h.service.CreateHourlyLog(c.Request.Context(), HourlyLogInput{ReportID: reportID, CreatorID: creatorID, AreaID: areaID, Level: c.PostForm("level"), LocationNotes: c.PostForm("location_notes"), Description: c.PostForm("description"), ActionTaken: c.PostForm("action_taken"), FollowUpStatus: c.PostForm("follow_up_status")})
	if err != nil {
		code := "failed-log"
		if errors.Is(err, ErrInvalid) {
			code = "invalid-log"
		} else if errors.Is(err, ErrForbidden) {
			code = "forbidden-log"
		} else if errors.Is(err, sql.ErrNoRows) {
			code = "not-found"
		} else {
			slog.Error("create hourly security log", "error", err)
		}
		h.redirect(c, "/reports/security/"+strconv.FormatUint(reportID, 10), "error", code)
		return
	}
	h.redirect(c, "/reports/security/"+strconv.FormatUint(reportID, 10), "success", "log-created")
}

func (h *Handler) redirectToReport(c *gin.Context, key, value string) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	h.redirect(c, "/reports/security/"+strconv.FormatUint(id, 10), key, value)
}

func (h *Handler) render(c *gin.Context, template, title string, data gin.H) {
	u, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), u, "security-reports")
	if err != nil {
		h.internalError(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.internalError(c, err)
		return
	}
	data["AppName"], data["IsDashboard"], data["Authenticated"] = h.appName, true, true
	data["Title"], data["Section"] = title, "Laporan"
	data["UserName"], data["UserInitials"], data["StoreName"], data["RoleName"] = u.Name, initials(u.Name), u.StoreName, u.RoleName
	data["Menus"], data["CSRFToken"] = menus, token
	data["SuccessMessage"], data["ErrorMessage"] = message(c.Query("success")), message(c.Query("error"))
	c.HTML(http.StatusOK, template, data)
}
func (h *Handler) redirect(c *gin.Context, path, key, value string) {
	c.Redirect(http.StatusSeeOther, path+"?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("security report page", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}
func initials(name string) string {
	p := strings.Fields(name)
	if len(p) == 0 {
		return "US"
	}
	v := string([]rune(strings.ToUpper(p[0]))[0])
	if len(p) > 1 {
		v += string([]rune(strings.ToUpper(p[len(p)-1]))[0])
	}
	return v
}
func message(code string) string {
	return map[string]string{"created": "Laporan shift berhasil dimulai.", "log-created": "Laporan per jam berhasil ditambahkan.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data laporan belum lengkap atau tidak valid.", "invalid-log": "Isi uraian laporan dan tindakan untuk level waspada atau urgent.", "forbidden-log": "Anda bukan anggota shift ini dan tidak dapat menambahkan laporan.", "failed-log": "Laporan per jam gagal disimpan.", "duplicate": "Laporan untuk tanggal dan shift tersebut sudah tersedia.", "not-found": "Laporan shift tidak ditemukan.", "failed": "Laporan shift gagal disimpan."}[code]
}
