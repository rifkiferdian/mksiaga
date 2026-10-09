package profile

import (
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
	current, _ := auth.CurrentUser(c)
	userID, userStoreID, ok := ids(current)
	if !ok {
		c.Redirect(http.StatusSeeOther, "/login")
		return
	}
	item, err := h.service.Profile(c.Request.Context(), userID, userStoreID)
	if err != nil {
		h.internalError(c, err)
		return
	}
	menus, err := h.navigation.Menus(c.Request.Context(), current, "")
	if err != nil {
		h.internalError(c, err)
		return
	}
	token, err := auth.CSRFToken(c)
	if err != nil {
		h.internalError(c, err)
		return
	}
	c.HTML(http.StatusOK, "profile/index.html", gin.H{"Title": "Profile", "Section": "Akun", "AppName": h.appName, "IsDashboard": true, "Authenticated": true, "UserName": current.Name, "UserInitials": initials(current.Name), "StoreName": current.StoreName, "RoleName": current.RoleName, "CSRFToken": token, "Menus": menus, "Profile": item, "ProfileInitials": initials(item.Name), "SuccessMessage": message(c.Query("success")), "ErrorMessage": message(c.Query("error"))})
}
func (h *Handler) UpdateName(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, _, ok := ids(current)
	if !ok {
		h.redirect(c, "error", "failed")
		return
	}
	name, err := h.service.UpdateName(c.Request.Context(), userID, c.PostForm("name"))
	if err != nil {
		h.redirectError(c, err)
		return
	}
	if err := auth.UpdateSessionName(c, name); err != nil {
		h.internalError(c, err)
		return
	}
	h.redirect(c, "success", "name-updated")
}
func (h *Handler) UpdatePassword(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "error", "csrf")
		return
	}
	current, _ := auth.CurrentUser(c)
	userID, _, ok := ids(current)
	if !ok {
		h.redirect(c, "error", "failed")
		return
	}
	err := h.service.UpdatePassword(c.Request.Context(), userID, c.PostForm("current_password"), c.PostForm("new_password"), c.PostForm("password_confirmation"))
	if err != nil {
		h.redirectError(c, err)
		return
	}
	h.redirect(c, "success", "password-updated")
}
func (h *Handler) redirectError(c *gin.Context, err error) {
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalidName):
		code = "invalid-name"
	case errors.Is(err, ErrInvalidPassword):
		code = "invalid-password"
	case errors.Is(err, ErrWrongPassword):
		code = "wrong-password"
	default:
		slog.Error("profile operation failed", "error", err)
	}
	h.redirect(c, "error", code)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/profile?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render profile", "error", err)
	c.String(http.StatusInternalServerError, "Halaman tidak dapat diproses.")
}
func ids(current auth.SessionUser) (uint64, uint64, bool) {
	userID, e1 := strconv.ParseUint(current.ID, 10, 64)
	userStoreID, e2 := strconv.ParseUint(current.UserStoreID, 10, 64)
	return userID, userStoreID, e1 == nil && e2 == nil
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
	return map[string]string{"name-updated": "Nama profile berhasil diperbarui.", "password-updated": "Password berhasil diganti.", "csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid-name": "Nama tidak boleh kosong dan maksimal 150 karakter.", "invalid-password": "Password baru minimal 8 karakter, harus sama dengan konfirmasi, dan berbeda dari password saat ini.", "wrong-password": "Password saat ini tidak sesuai.", "failed": "Perubahan profile gagal diproses."}[code]
}
