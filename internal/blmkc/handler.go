package blmkc

import (
	"database/sql"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"mksiaga/internal/navigation"
)

type Handler struct {
	appName    string
	service    *Service
	navigation *navigation.Service
}

func NewHandler(appName string, service *Service, navigationService *navigation.Service) *Handler {
	return &Handler{appName: appName, service: service, navigation: navigationService}
}

func (h *Handler) Dashboard(c *gin.Context) {
	data, err := h.service.Dashboard(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/dashboard.html", "Dashboard BLMKC", "blmkc-dashboard", gin.H{"Data": data})
}

func (h *Handler) Cases(c *gin.Context) {
	storeID, _ := strconv.ParseUint(c.Query("store_id"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	filter := Filter{
		Search: strings.TrimSpace(c.Query("q")), StoreID: storeID,
		CaseStatus: strings.TrimSpace(c.Query("status")), VerificationStatus: strings.TrimSpace(c.Query("verification")),
		StartDate: strings.TrimSpace(c.Query("start_date")), EndDate: strings.TrimSpace(c.Query("end_date")), Page: page,
	}
	data, err := h.service.Cases(c.Request.Context(), filter)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/cases.html", "Daftar Kasus BLMKC", "blmkc-cases", gin.H{
		"Data": data, "SuccessMessage": caseMessage(c.Query("success")), "ErrorMessage": caseMessage(c.Query("error")),
	})
}

func (h *Handler) Wanted(c *gin.Context) {
	data, err := h.service.Wanted(c.Request.Context(), strings.TrimSpace(c.Query("q")), strings.TrimSpace(c.Query("status")))
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/wanted.html", "Wanted BLMKC", "blmkc-wanted", gin.H{"Data": data})
}

func (h *Handler) Identities(c *gin.Context) {
	data, err := h.service.Identities(c.Request.Context(), strings.TrimSpace(c.Query("q")))
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/identity.html", "Identity BLMKC", "blmkc-identity", gin.H{"Data": data})
}

func (h *Handler) Verifications(c *gin.Context) {
	data, err := h.service.Verifications(c.Request.Context(), strings.TrimSpace(c.Query("q")), strings.TrimSpace(c.Query("status")))
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/verification.html", "Verifikasi BLMKC", "blmkc-verification", gin.H{"Data": data})
}

func (h *Handler) NewCase(c *gin.Context) {
	data, err := h.service.CaseForm(c.Request.Context(), 0)
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/form.html", "Tambah Kasus BLMKC", "blmkc-cases", gin.H{"Data": data})
}

func (h *Handler) CreateCase(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectCases(c, "error", "csrf")
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	id, err := h.service.CreateCase(c.Request.Context(), caseFormInput(c), userID)
	if err != nil {
		h.redirectServiceError(c, "/blmkc/cases", err)
		return
	}
	c.Redirect(http.StatusSeeOther, "/blmkc/cases/"+strconv.FormatUint(id, 10)+"/evidence/new?success=created")
}

func (h *Handler) CaseDetail(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	data, err := h.service.CaseDetail(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/detail.html", "Detail Kasus BLMKC", "blmkc-cases", gin.H{
		"Data": data, "SuccessMessage": caseMessage(c.Query("success")), "ErrorMessage": caseMessage(c.Query("error")),
	})
}

func (h *Handler) EvidenceStep(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	data, err := h.service.CaseDetail(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/evidence_form.html", "Bukti CCTV Kasus", "blmkc-cases", gin.H{
		"Data": data, "SuccessMessage": caseMessage(c.Query("success")), "ErrorMessage": caseMessage(c.Query("error")),
	})
}

func (h *Handler) EditCase(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	data, err := h.service.CaseForm(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "blmkc/form.html", "Edit Kasus BLMKC", "blmkc-cases", gin.H{"Data": data})
}

func (h *Handler) UpdateCase(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectCase(c, id, "error", "csrf")
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	if err := h.service.UpdateCase(c.Request.Context(), id, userID, caseFormInput(c)); err != nil {
		h.redirectServiceError(c, "/blmkc/cases/"+strconv.FormatUint(id, 10), err)
		return
	}
	h.redirectCase(c, id, "success", "updated")
}

func (h *Handler) AdvanceStatus(c *gin.Context) {
	id, ok := caseRouteID(c)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/errors/404")
		return
	}
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirectCase(c, id, "error", "csrf")
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	err = h.service.AdvanceStatus(c.Request.Context(), id, userID, c.PostForm("subject_name"), c.PostForm("notes"))
	if err != nil {
		h.redirectServiceError(c, "/blmkc/cases/"+strconv.FormatUint(id, 10), err)
		return
	}
	h.redirectCase(c, id, "success", "status-updated")
}

func caseFormInput(c *gin.Context) CaseInput {
	storeID, _ := strconv.ParseUint(c.PostForm("store_id"), 10, 64)
	return CaseInput{
		StoreID: storeID, IncidentTypeID: optionalUint(c.PostForm("incident_type_id")),
		AreaID: optionalUint(c.PostForm("area_id")), CameraID: optionalUint(c.PostForm("camera_id")),
		OccurredAt: c.PostForm("occurred_at"), Title: c.PostForm("title"), Chronology: c.PostForm("chronology"),
		LossAmount: optionalFloat(c.PostForm("loss_amount")), SubjectName: c.PostForm("subject_name"), InitialStatus: c.PostForm("initial_status"),
	}
}

func optionalUint(value string) *uint64 {
	parsed, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	if err != nil || parsed == 0 {
		return nil
	}
	return &parsed
}

func optionalFloat(value string) *float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		invalid := -1.0
		return &invalid
	}
	return &parsed
}

func caseRouteID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *Handler) redirectServiceError(c *gin.Context, path string, err error) {
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalid):
		code = "invalid"
	case errors.Is(err, ErrInvalidTransition):
		code = "invalid-transition"
	case errors.Is(err, ErrSubjectRequired):
		code = "subject-required"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("BLMKC operation failed", "error", err)
	}
	c.Redirect(http.StatusSeeOther, path+"?error="+url.QueryEscape(code))
}

func (h *Handler) redirectCases(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/blmkc/cases?"+key+"="+url.QueryEscape(value))
}

func (h *Handler) redirectCase(c *gin.Context, id uint64, key, value string) {
	c.Redirect(http.StatusSeeOther, "/blmkc/cases/"+strconv.FormatUint(id, 10)+"?"+key+"="+url.QueryEscape(value))
}

func caseMessage(code string) string {
	return map[string]string{
		"created": "Kasus berhasil dibuat.", "updated": "Kasus berhasil diperbarui.", "status-updated": "Status kasus berhasil diperbarui.",
		"csrf": "Sesi form berakhir. Muat ulang halaman lalu coba kembali.", "invalid": "Data kasus belum lengkap atau tidak valid.",
		"invalid-transition": "Perubahan status tidak diizinkan.", "subject-required": "Nama atau identitas subjek wajib diisi sebelum menjadi Identity.",
		"evidence-uploaded": "Bukti CCTV berhasil diunggah dan menunggu verifikasi.", "evidence-verified": "Status verifikasi bukti berhasil diperbarui.",
		"invalid-evidence":   "File bukti tidak valid. Gunakan JPG, PNG, WebP, MP4, atau WebM.",
		"evidence-too-large": "Ukuran bukti melebihi batas: foto 10 MB atau video 150 MB.",
		"not-found":          "Kasus atau bukti tidak ditemukan.", "failed": "Operasi kasus gagal diproses.",
	}[code]
}

func (h *Handler) render(c *gin.Context, templateName, title, activePage string, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, activePage)
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
	data["Title"], data["Section"] = title, "BLMKC"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	c.HTML(http.StatusOK, templateName, data)
}

func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render BLMKC", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
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
