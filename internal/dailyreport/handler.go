package dailyreport

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
	h.renderPage(c, "dailyreport/index.html", "Laporan Harian CCTV")
}

func (h *Handler) Recap(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	now := time.Now()
	startDate := c.DefaultQuery("start_date", now.Format("2006-01")+"-01")
	endDate := c.DefaultQuery("end_date", now.Format("2006-01-02"))
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	filter := RecapFilter{StartDate: startDate, EndDate: endDate, StoreID: storeID, Status: strings.TrimSpace(c.Query("status"))}
	data, err := h.service.Recap(c.Request.Context(), filter)
	if err != nil {
		h.internalError(c, err)
		return
	}
	menus, err := h.navigation.Menus(c.Request.Context(), u, "daily-recap")
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.HTML(http.StatusOK, "dailyreport/recap.html", gin.H{"AppName": h.appName, "IsDashboard": true, "Authenticated": true, "Title": "Rekap Laporan CCTV", "Section": "Laporan", "UserName": u.Name, "UserInitials": initials(u.Name), "StoreName": u.StoreName, "RoleName": u.RoleName, "Menus": menus, "Data": data})
}

func (h *Handler) New(c *gin.Context) {
	h.renderPage(c, "dailyreport/create.html", "Buat Laporan Harian CCTV")
}

func (h *Handler) renderPage(c *gin.Context, templateName, title string) {
	u, _ := auth.CurrentUser(c)
	storeID, err := strconv.ParseUint(u.StoreID, 10, 64)
	userID, userErr := strconv.ParseUint(u.ID, 10, 64)
	if err != nil || userErr != nil {
		h.internalError(c, err)
		return
	}
	data, err := h.service.PageData(c.Request.Context(), storeID, userID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	if !hasRole(u.RoleName, "superadmin") {
		for i := range data.Reports {
			data.Reports[i].CanReview = false
		}
	}
	menus, err := h.navigation.Menus(c.Request.Context(), u, "daily-reports")
	if err != nil {
		h.internalError(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.HTML(http.StatusOK, templateName, gin.H{"AppName": h.appName, "IsDashboard": true, "Authenticated": true, "Title": title, "Section": "Laporan", "UserName": u.Name, "UserInitials": initials(u.Name), "StoreName": u.StoreName, "RoleName": u.RoleName, "CSRFToken": token, "Menus": menus, "Today": time.Now().Format("2006-01-02"), "MonitorSlots": []string{"08:00", "10:00", "12:00", "14:00", "16:00", "18:00", "20:00", "22:00"}, "Data": data, "SuccessMessage": message(c.Query("success")), "ErrorMessage": message(c.Query("error"))})
}

func (h *Handler) Workflow(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, e1 := strconv.ParseUint(u.StoreID, 10, 64)
	userID, e2 := strconv.ParseUint(u.ID, 10, 64)
	id, e3 := strconv.ParseUint(c.Param("id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	report, err := h.service.WorkflowReport(c.Request.Context(), id)
	if err != nil {
		h.redirect(c, "error", "not-found")
		return
	}
	data, err := h.service.PageData(c.Request.Context(), storeID, userID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	menus, err := h.navigation.Menus(c.Request.Context(), u, "daily-reports")
	if err != nil {
		h.internalError(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.HTML(http.StatusOK, "dailyreport/workflow.html", gin.H{"AppName": h.appName, "IsDashboard": true, "Authenticated": true, "Title": "Proses Laporan CCTV", "Section": "Laporan", "UserName": u.Name, "UserInitials": initials(u.Name), "StoreName": u.StoreName, "RoleName": u.RoleName, "CanReview": hasRole(u.RoleName, "superadmin"), "CurrentUserID": userID, "CSRFToken": token, "Menus": menus, "Report": report, "Data": data, "SuccessMessage": message(c.Query("success")), "ErrorMessage": message(c.Query("error"))})
}

func (h *Handler) Print(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	report, err := h.service.WorkflowReport(c.Request.Context(), id)
	if err != nil {
		h.redirect(c, "error", "not-found")
		return
	}
	c.HTML(http.StatusOK, "dailyreport/print.html", gin.H{
		"AppName": h.appName, "IsDashboard": true, "Authenticated": true,
		"Title":    "Cetak Laporan CCTV " + report.ReportNumber,
		"UserName": u.Name, "Report": report,
	})
}

func (h *Handler) Receive(c *gin.Context) { h.workflowAction(c, true) }
func (h *Handler) Review(c *gin.Context)  { h.workflowAction(c, false) }
func (h *Handler) Revise(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, e1 := strconv.ParseUint(u.StoreID, 10, 64)
	userID, e2 := strconv.ParseUint(u.ID, 10, 64)
	id, e3 := strconv.ParseUint(c.Param("id"), 10, 64)
	receiverID, e4 := strconv.ParseUint(c.PostForm("handover_to_user_id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil || e4 != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	if err := h.service.Revise(c.Request.Context(), id, storeID, userID, receiverID, c.PostForm("handover_at"), c.PostForm("handover_notes"), c.PostForm("handover_from_signature")); err != nil {
		code := "invalid"
		if errors.Is(err, sql.ErrNoRows) {
			code = "not-found"
		}
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/reports/daily/%d/workflow?error=%s", id, url.QueryEscape(code)))
		return
	}
	h.redirect(c, "success", "resubmitted")
}
func (h *Handler) workflowAction(c *gin.Context, receive bool) {
	u, _ := auth.CurrentUser(c)
	storeID, e1 := strconv.ParseUint(u.StoreID, 10, 64)
	userID, e2 := strconv.ParseUint(u.ID, 10, 64)
	id, e3 := strconv.ParseUint(c.Param("id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	action := c.PostForm("action")
	approve := action == "accept" || action == "approve"
	var err error
	if receive {
		err = h.service.Receive(c.Request.Context(), id, storeID, userID, c.PostForm("handover_to_signature"), c.PostForm("notes"), approve)
	} else {
		err = h.service.Review(c.Request.Context(), id, userID, c.PostForm("notes"), approve)
	}
	if err != nil {
		code := "invalid"
		if errors.Is(err, sql.ErrNoRows) {
			code = "not-found"
		}
		c.Redirect(http.StatusSeeOther, fmt.Sprintf("/reports/daily/%d/workflow?error=%s", id, url.QueryEscape(code)))
		return
	}
	result := "received"
	if !approve {
		result = "revision"
	} else if !receive {
		result = "approved"
	}
	h.redirect(c, "success", result)
}

func (h *Handler) Create(c *gin.Context) {
	u, _ := auth.CurrentUser(c)
	storeID, e1 := strconv.ParseUint(u.StoreID, 10, 64)
	officerID, e2 := strconv.ParseUint(u.ID, 10, 64)
	shiftID, e3 := strconv.ParseUint(c.PostForm("shift_id"), 10, 64)
	if e1 != nil || e2 != nil || e3 != nil {
		h.redirectForm(c, "invalid")
		return
	}
	in := Input{StoreID: storeID, OfficerID: officerID, SupervisorID: optionalID(c.PostForm("supervisor_id")), ShiftID: shiftID, ReportDate: c.PostForm("report_date"), WorkStart: c.PostForm("work_start"), WorkEnd: c.PostForm("work_end"), Location: c.PostForm("monitoring_location"), StartNotes: c.PostForm("start_notes"), EndNotes: c.PostForm("end_notes"), HandoverToUserID: optionalID(c.PostForm("handover_to_user_id")), HandoverAt: c.PostForm("handover_at"), HandoverNotes: c.PostForm("handover_notes"), HandoverFromSign: c.PostForm("handover_from_signature"), HandoverToSign: c.PostForm("handover_to_signature"), Status: c.PostForm("status")}
	for _, raw := range c.PostFormArray("checklist_ids[]") {
		id, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			continue
		}
		prefix := "checklist_" + raw
		answer := c.PostForm(prefix + "_answer")
		if answer == "" {
			continue
		}
		in.Checklists = append(in.Checklists, ChecklistAnswer{ChecklistID: id, Answer: answer, Active: optionalUint16(c.PostForm(prefix + "_active")), Total: optionalUint16(c.PostForm(prefix + "_total")), Notes: c.PostForm(prefix + "_notes")})
	}
	times := c.PostFormArray("monitor_time[]")
	areas := c.PostFormArray("monitor_area_id[]")
	conditions := c.PostFormArray("monitor_condition[]")
	findings := c.PostFormArray("monitor_finding[]")
	actions := c.PostFormArray("monitor_action[]")
	for i, v := range times {
		if strings.TrimSpace(v) == "" {
			continue
		}
		in.Monitoring = append(in.Monitoring, MonitoringLog{Time: v, AreaID: optionalID(at(areas, i)), Condition: at(conditions, i), Finding: at(findings, i), Action: at(actions, i)})
	}
	for _, rawIndex := range c.PostFormArray("incident_indices[]") {
		i, indexErr := strconv.Atoi(rawIndex)
		if indexErr != nil || i < 0 {
			continue
		}
		p := fmt.Sprintf("incident_%d_", i)
		incident := Incident{Time: c.PostForm(p + "time"), AreaID: optionalID(c.PostForm(p + "area_id")), CameraID: optionalID(c.PostForm(p + "camera_id")), IncidentTypeID: optionalID(c.PostForm(p + "type_id")), OtherType: c.PostForm(p + "other_type"), Description: c.PostForm(p + "description"), Gender: c.PostForm(p + "gender"), Age: c.PostForm(p + "age"), Clothing: c.PostForm(p + "clothing"), Features: c.PostForm(p + "features"), ResponseMinutes: optionalUint16(c.PostForm(p + "response_minutes")), Status: c.PostForm(p + "status"), OtherActionNotes: c.PostForm(p + "action_notes")}
		for _, raw := range c.PostFormArray(p + "action_ids[]") {
			if id, err := strconv.ParseUint(raw, 10, 64); err == nil {
				incident.ActionIDs = append(incident.ActionIDs, id)
			}
		}
		if incident.Time != "" || incident.IncidentTypeID != nil || incident.Description != "" {
			in.Incidents = append(in.Incidents, incident)
		}
	}
	if err := h.service.Create(c.Request.Context(), in); err != nil {
		code := "failed"
		if errors.Is(err, ErrInvalid) {
			code = "invalid"
		} else if errors.Is(err, ErrDuplicate) {
			code = "duplicate"
		} else {
			slog.Error("create daily report", "error", err)
		}
		h.redirectForm(c, code)
		return
	}
	h.redirect(c, "success", "created")
}

func (h *Handler) Delete(c *gin.Context) {
	id, e2 := strconv.ParseUint(c.Param("id"), 10, 64)
	if e2 != nil {
		h.redirect(c, "error", "invalid")
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			slog.Error("delete daily report", "error", err)
		}
		h.redirect(c, "error", "not-found")
		return
	}
	h.redirect(c, "success", "deleted")
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/reports/daily?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) redirectForm(c *gin.Context, value string) {
	c.Redirect(http.StatusSeeOther, "/reports/daily/new?error="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("daily report page", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}
func optionalID(v string) *uint64 {
	id, e := strconv.ParseUint(v, 10, 64)
	if e != nil || id == 0 {
		return nil
	}
	return &id
}
func optionalUint16(v string) *uint16 {
	n, e := strconv.ParseUint(v, 10, 16)
	if e != nil {
		return nil
	}
	value := uint16(n)
	return &value
}
func at(values []string, i int) string {
	if i >= 0 && i < len(values) {
		return values[i]
	}
	return ""
}
func initials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "US"
	}
	value := string([]rune(strings.ToUpper(parts[0]))[0])
	if len(parts) > 1 {
		value += string([]rune(strings.ToUpper(parts[len(parts)-1]))[0])
	}
	return value
}
func hasRole(roles, wanted string) bool {
	for _, role := range strings.Split(roles, ",") {
		if strings.TrimSpace(role) == wanted {
			return true
		}
	}
	return false
}
func message(code string) string {
	return map[string]string{"created": "Laporan harian berhasil disimpan.", "deleted": "Laporan berhasil dihapus.", "received": "Serah terima diterima dan diteruskan ke supervisor.", "approved": "Laporan disetujui dan dikunci.", "revision": "Laporan dikembalikan untuk diperbaiki.", "resubmitted": "Perbaikan laporan berhasil diajukan kembali.", "invalid": "Data wajib laporan belum lengkap atau tidak valid.", "duplicate": "Laporan untuk tanggal dan shift tersebut sudah tersedia.", "not-found": "Laporan tidak ditemukan atau statusnya sudah berubah.", "failed": "Laporan gagal disimpan."}[code]
}
