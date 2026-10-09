package store

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
	stores, err := h.service.Stores(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	active, inactive, users := 0, 0, 0
	for _, item := range stores {
		if item.Status == "active" {
			active++
		} else {
			inactive++
		}
		users += item.UserCount
	}
	h.render(c, http.StatusOK, "store/index.html", gin.H{"Title": "Store", "ActivePage": "stores", "Stores": stores, "StoreCount": len(stores), "ActiveStoreCount": active, "InactiveStoreCount": inactive, "StoreUserCount": users})
}
func (h *Handler) Create(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	h.redirectResult(c, h.service.Create(c.Request.Context(), formInput(c)), "store-created")
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
	h.redirectResult(c, h.service.Update(c.Request.Context(), id, formInput(c)), "store-updated")
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
	current, _ := auth.CurrentUser(c)
	currentStoreID, _ := strconv.ParseUint(current.StoreID, 10, 64)
	h.redirectResult(c, h.service.Delete(c.Request.Context(), id, currentStoreID), "store-deleted")
}
func formInput(c *gin.Context) Input {
	return Input{Code: c.PostForm("code"), Name: c.PostForm("name"), Address: c.PostForm("address"), Phone: c.PostForm("phone"), Timezone: c.PostForm("timezone"), Status: c.PostForm("status")}
}
func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}
func (h *Handler) render(c *gin.Context, status int, templateName string, data gin.H) {
	current, _ := auth.CurrentUser(c)
	menus, err := h.navigation.Menus(c.Request.Context(), current, "stores")
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
	data["Section"] = "Master Data"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"] = message(c.Query("success"))
	data["ErrorMessage"] = message(c.Query("error"))
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
	case errors.Is(err, ErrProtected):
		code = "protected"
	case errors.Is(err, ErrInUse):
		code = "in-use"
	case errors.Is(err, sql.ErrNoRows):
		code = "not-found"
	default:
		slog.Error("store operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/master/stores?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render stores", "error", err)
	c.String(http.StatusInternalServerError, "Halaman tidak dapat diproses.")
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
	return map[string]string{"store-created": "Store berhasil dibuat.", "store-updated": "Store berhasil diperbarui.", "store-deleted": "Store berhasil dihapus.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Data store belum lengkap atau tidak valid.", "duplicate": "Kode store sudah digunakan.", "protected": "Store yang sedang digunakan tidak dapat dihapus.", "in-use": "Store masih memiliki user aktif dan tidak dapat dihapus.", "not-found": "Store tidak ditemukan.", "failed": "Operasi store gagal diproses."}[code]
}
