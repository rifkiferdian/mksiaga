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
	userID, _ := strconv.ParseUint(u.ID, 10, 64)
	for i := range data.Reports {
		data.Reports[i].CanEdit = data.Reports[i].RepresentativeID == userID && (data.Reports[i].Status == "active" || data.Reports[i].Status == "revision")
	}
	h.render(c, "securityreport/index.html", "Laporan Shift Security", gin.H{"Data": data})
}
func (h *Handler) History(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, e := strconv.ParseUint(u.StoreID, 10, 64)
	if e != nil {
		h.internalError(c, e)
		return
	}
	now := time.Now()
	start := c.DefaultQuery("start_date", now.Format("2006-01")+"-01")
	end := c.DefaultQuery("end_date", now.Format("2006-01-02"))
	shiftID, _ := strconv.ParseUint(c.Query("shift_id"), 10, 64)
	userID, _ := strconv.ParseUint(c.Query("user_id"), 10, 64)
	data, err := h.service.History(c.Request.Context(), storeID, HistoryFilter{StartDate: start, EndDate: end, ShiftID: shiftID, UserID: userID, Status: strings.TrimSpace(c.Query("status")), Level: strings.TrimSpace(c.Query("level"))})
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "securityreport/history.html", "Riwayat Laporan Shift", gin.H{"Data": data})
}
func (h *Handler) Monitoring(c *gin.Context) {
	data, err := h.service.Monitoring(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "securityreport/monitoring.html", "Monitoring SPV", gin.H{"Data": data})
}
func (h *Handler) Print(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	id, e := strconv.ParseUint(c.Param("id"), 10, 64)
	if e != nil {
		h.redirect(c, "/reports/security/history", "error", "not-found")
		return
	}
	d, err := h.service.Detail(c.Request.Context(), id)
	if err != nil {
		h.redirect(c, "/reports/security/history", "error", "not-found")
		return
	}
	c.HTML(http.StatusOK, "securityreport/print.html", gin.H{"AppName": h.appName, "Title": "Cetak Laporan Shift", "IsDashboard": true, "Authenticated": true, "Report": d, "UserName": u.Name})
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

func (h *Handler) Edit(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, e1 := strconv.ParseUint(u.StoreID, 10, 64)
	userID, e2 := strconv.ParseUint(u.ID, 10, 64)
	id, e3 := strconv.ParseUint(c.Param("id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil {
		h.redirect(c, "/reports/security", "error", "not-found")
		return
	}
	d, err := h.service.Detail(c.Request.Context(), id)
	if err != nil {
		h.redirect(c, "/reports/security", "error", "not-found")
		return
	}
	if d.RepresentativeID != userID || (d.Status != "active" && d.Status != "revision") {
		h.redirect(c, "/reports/security", "error", "forbidden-handover")
		return
	}
	data, err := h.service.PageData(c.Request.Context(), storeID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	selected := map[uint64]bool{}
	for _, member := range d.Members {
		selected[member.UserID] = true
	}
	h.render(c, "securityreport/edit.html", "Edit Laporan Shift", gin.H{"Report": d, "Data": data, "SelectedMembers": selected})
}

func (h *Handler) Update(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectToReport(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	reportID, e1 := strconv.ParseUint(c.Param("id"), 10, 64)
	actorID, e2 := strconv.ParseUint(u.ID, 10, 64)
	storeID, e3 := strconv.ParseUint(u.StoreID, 10, 64)
	shiftID, e4 := strconv.ParseUint(c.PostForm("shift_id"), 10, 64)
	representativeID, e5 := strconv.ParseUint(c.PostForm("representative_user_id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil || e5 != nil {
		h.redirectToReport(c, "error", "invalid")
		return
	}
	var supervisorID *uint64
	if raw := c.PostForm("supervisor_user_id"); raw != "" {
		v, e := strconv.ParseUint(raw, 10, 64)
		if e != nil {
			h.redirectToReport(c, "error", "invalid")
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
	err := h.service.Update(c.Request.Context(), UpdateInput{ReportID: reportID, ActorID: actorID, StoreID: storeID, ShiftID: shiftID, RepresentativeID: representativeID, SupervisorID: supervisorID, ReportDate: c.PostForm("report_date"), OpeningNotes: c.PostForm("opening_notes"), MemberIDs: members})
	if err != nil {
		code := "failed"
		var my *mysql.MySQLError
		if errors.Is(err, ErrInvalid) {
			code = "invalid"
		} else if errors.Is(err, ErrForbidden) {
			code = "forbidden-handover"
		} else if errors.As(err, &my) && my.Number == 1062 {
			code = "duplicate"
		} else {
			slog.Error("update security report", "error", err)
		}
		h.redirect(c, "/reports/security/"+strconv.FormatUint(reportID, 10)+"/edit", "error", code)
		return
	}
	h.redirect(c, "/reports/security/"+strconv.FormatUint(reportID, 10), "success", "report-updated")
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
	for i := range d.Logs {
		d.Logs[i].CanEdit = d.Status == "active" && (d.Logs[i].CreatedByID == userID || d.RepresentativeID == userID)
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
	canSubmit := d.RepresentativeID == userID && (d.Status == "active" || d.Status == "revision")
	canReceive := d.Handover != nil && d.Handover.ReceiverUserID == userID && d.Status == "waiting_handover"
	allowAny := hasRole(u.RoleName, "superadmin")
	canReview := d.Handover != nil && d.Status == "waiting_supervisor" && (allowAny || (hasRole(u.RoleName, "spv_security") && d.Handover.SupervisorUserID == userID))
	h.render(c, "securityreport/detail.html", "Detail Laporan Shift", gin.H{"Report": d, "Areas": data.Areas, "Users": data.Users, "Supervisors": data.Supervisors, "CanAddLog": canAdd && d.Status == "active", "CanSubmitHandover": canSubmit, "CanReceiveHandover": canReceive, "CanReviewHandover": canReview})
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

func (h *Handler) UpdateHourlyLog(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectToReport(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	reportID, e1 := strconv.ParseUint(c.Param("id"), 10, 64)
	logID, e2 := strconv.ParseUint(c.Param("logID"), 10, 64)
	actorID, e3 := strconv.ParseUint(u.ID, 10, 64)
	if e1 != nil || e2 != nil || e3 != nil {
		h.redirectToReport(c, "error", "invalid-log")
		return
	}
	var areaID *uint64
	if raw := c.PostForm("area_id"); raw != "" {
		v, e := strconv.ParseUint(raw, 10, 64)
		if e != nil {
			h.redirectToReport(c, "error", "invalid-log")
			return
		}
		areaID = &v
	}
	err := h.service.UpdateHourlyLog(c.Request.Context(), HourlyLogInput{ReportID: reportID, LogID: logID, CreatorID: actorID, AreaID: areaID, Level: c.PostForm("level"), LocationNotes: c.PostForm("location_notes"), Description: c.PostForm("description"), ActionTaken: c.PostForm("action_taken"), FollowUpStatus: c.PostForm("follow_up_status")})
	if err != nil {
		code := "failed-log"
		if errors.Is(err, ErrInvalid) {
			code = "invalid-log"
		} else if errors.Is(err, ErrForbidden) {
			code = "forbidden-log"
		} else if errors.Is(err, sql.ErrNoRows) {
			code = "not-found"
		} else {
			slog.Error("update hourly security log", "error", err)
		}
		h.redirectToReport(c, "error", code)
		return
	}
	h.redirectToReport(c, "success", "log-updated")
}

func (h *Handler) SubmitHandover(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectToReport(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	reportID, e1 := strconv.ParseUint(c.Param("id"), 10, 64)
	actorID, e2 := strconv.ParseUint(u.ID, 10, 64)
	receiverID, e3 := strconv.ParseUint(c.PostForm("receiver_user_id"), 10, 64)
	supervisorID, e4 := strconv.ParseUint(c.PostForm("supervisor_user_id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		h.redirectToReport(c, "error", "invalid-handover")
		return
	}
	err := h.service.SubmitHandover(c.Request.Context(), HandoverInput{ReportID: reportID, ActorID: actorID, ReceiverID: receiverID, SupervisorID: supervisorID, ClosingSummary: c.PostForm("closing_summary"), HandoverNotes: c.PostForm("handover_notes"), PendingItems: c.PostForm("pending_items"), InventoryNotes: c.PostForm("inventory_notes"), Signature: c.PostForm("giver_signature")})
	h.workflowResult(c, reportID, err, "handover-submitted")
}
func (h *Handler) ReceiveHandover(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectToReport(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	reportID, e1 := strconv.ParseUint(c.Param("id"), 10, 64)
	actorID, e2 := strconv.ParseUint(u.ID, 10, 64)
	if e1 != nil || e2 != nil {
		h.redirectToReport(c, "error", "invalid-handover")
		return
	}
	accept := c.PostForm("action") == "accept"
	err := h.service.ReceiveHandover(c.Request.Context(), ReceiveInput{ReportID: reportID, ActorID: actorID, Notes: c.PostForm("notes"), Signature: c.PostForm("receiver_signature"), Accept: accept})
	success := "handover-rejected"
	if accept {
		success = "handover-received"
	}
	h.workflowResult(c, reportID, err, success)
}
func (h *Handler) ReviewHandover(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectToReport(c, "error", "csrf")
		return
	}
	u, _ := auth.CurrentUser(c)
	reportID, e1 := strconv.ParseUint(c.Param("id"), 10, 64)
	actorID, e2 := strconv.ParseUint(u.ID, 10, 64)
	if e1 != nil || e2 != nil {
		h.redirectToReport(c, "error", "invalid-handover")
		return
	}
	approve := c.PostForm("action") == "approve"
	err := h.service.ReviewHandover(c.Request.Context(), ReviewInput{ReportID: reportID, ActorID: actorID, Notes: c.PostForm("notes"), Signature: c.PostForm("supervisor_signature"), Approve: approve, AllowAny: hasRole(u.RoleName, "superadmin")})
	success := "review-revision"
	if approve {
		success = "review-approved"
	}
	h.workflowResult(c, reportID, err, success)
}
func (h *Handler) workflowResult(c *gin.Context, reportID uint64, err error, success string) {
	if err == nil {
		h.redirect(c, "/reports/security/"+strconv.FormatUint(reportID, 10), "success", success)
		return
	}
	code := "failed-handover"
	if errors.Is(err, ErrInvalid) {
		code = "invalid-handover"
	} else if errors.Is(err, ErrForbidden) {
		code = "forbidden-handover"
	} else if errors.Is(err, ErrSupervisorRequired) {
		code = "supervisor-required"
	} else if errors.Is(err, sql.ErrNoRows) {
		code = "not-found"
	} else {
		slog.Error("security handover workflow", "error", err)
	}
	h.redirect(c, "/reports/security/"+strconv.FormatUint(reportID, 10), "error", code)
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
	if code == "report-updated" {
		return "Laporan shift berhasil diperbarui."
	}
	return map[string]string{"created": "Laporan shift berhasil dimulai.", "log-created": "Laporan per jam berhasil ditambahkan.", "log-updated": "Laporan per jam berhasil diperbarui.", "handover-submitted": "Laporan berhasil diserahkan dan menunggu penerima.", "handover-received": "Laporan diterima dan diteruskan kepada SPV.", "handover-rejected": "Laporan dikembalikan kepada penyerah untuk diperbaiki.", "review-approved": "Laporan disahkan oleh SPV dan dinyatakan selesai.", "review-revision": "SPV mengembalikan laporan untuk diperbaiki.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data laporan belum lengkap atau tidak valid.", "invalid-log": "Isi uraian laporan dan tindakan untuk level waspada atau urgent.", "invalid-handover": "Lengkapi data wajib dan tanda tangan sebelum melanjutkan.", "supervisor-required": "Tentukan SPV laporan terlebih dahulu sebelum serah terima.", "forbidden-handover": "Anda tidak berwenang menjalankan tahap serah terima ini.", "failed-handover": "Proses serah terima gagal disimpan.", "forbidden-log": "Anda tidak berwenang mengubah laporan ini atau shift sudah ditutup.", "failed-log": "Laporan per jam gagal disimpan.", "duplicate": "Laporan untuk tanggal dan shift tersebut sudah tersedia.", "not-found": "Laporan shift tidak ditemukan.", "failed": "Laporan shift gagal disimpan."}[code]
}
func hasRole(roles, required string) bool {
	for _, role := range strings.Split(roles, ",") {
		if strings.TrimSpace(role) == required {
			return true
		}
	}
	return false
}
