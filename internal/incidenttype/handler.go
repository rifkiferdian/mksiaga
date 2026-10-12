package incidenttype

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

func (h *Handler) Index(c *gin.Context) {
	items, err := h.service.IncidentTypes(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	active, inactive, descriptions := 0, 0, 0
	for _, item := range items {
		if item.Status == "active" {
			active++
		} else {
			inactive++
		}
		if item.RequiresDescription {
			descriptions++
		}
	}
	h.render(c, http.StatusOK, "incidenttype/index.html", gin.H{"Title": "Master Jenis Insiden", "IncidentTypes": items, "IncidentTypeCount": len(items), "ActiveIncidentTypeCount": active, "InactiveIncidentTypeCount": inactive, "DescriptionIncidentTypeCount": descriptions})
}

func (h *Handler) Create(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	h.redirectResult(c, h.service.Create(c.Request.Context(), formInput(c)), "incident-type-created")
}

func (h *Handler) Update(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.redirectResult(c, h.service.Update(c.Request.Context(), id, formInput(c)), "incident-type-updated")
}

func (h *Handler) Delete(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "error", "invalid")
		return
	}
	h.redirectResult(c, h.service.Delete(c.Request.Context(), id), "incident-type-deleted")
}

func formInput(c *gin.Context) Input {
	sortOrder, _ := strconv.ParseUint(c.PostForm("sort_order"), 10, 16)
	return Input{Code: c.PostForm("code"), Name: c.PostForm("name"), RequiresDescription: c.PostForm("requires_description") == "1", SortOrder: uint16(sortOrder), Status: c.PostForm("status")}
}

func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func (h *Handler) render(c *gin.Context, status int, templateName string, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, "incident-types")
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
	data["Section"] = "Master Aplikasi"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"], data["SuccessMessage"], data["ErrorMessage"] = menus, message(c.Query("success")), message(c.Query("error"))
	c.HTML(status, templateName, data)
}

func (h *Handler) redirectResult(c *gin.Context, err error, success string) {
	if err == nil {
		h.redirect(c, "success", success)
		return
	}
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalid):
		code = "invalid"
	case errors.Is(err, ErrDuplicate):
		code = "duplicate"
	case errors.Is(err, ErrInUse):
		code = "in-use"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("incident type operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}

func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/incident-types?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render incident types", "error", err)
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
func message(code string) string {
	return map[string]string{"incident-type-created": "Jenis insiden berhasil dibuat.", "incident-type-updated": "Jenis insiden berhasil diperbarui.", "incident-type-deleted": "Jenis insiden berhasil dihapus.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data jenis insiden belum lengkap atau tidak valid.", "duplicate": "Kode jenis insiden sudah digunakan.", "in-use": "Jenis insiden sudah digunakan dan tidak dapat dihapus.", "not-found": "Jenis insiden tidak ditemukan.", "failed": "Operasi jenis insiden gagal diproses."}[code]
}
