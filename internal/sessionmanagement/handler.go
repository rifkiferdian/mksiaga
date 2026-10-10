package sessionmanagement

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
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	items, err := h.service.Sessions(c.Request.Context(), userID, auth.CurrentSessionID(c))
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.render(c, "session/index.html", gin.H{"Title": "Manajemen sesi", "Sessions": items, "SessionCount": len(items)})
}
func (h *Handler) Revoke(c *gin.Context) {
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	err = h.service.Revoke(c.Request.Context(), userID, c.Param("id"), auth.CurrentSessionID(c))
	if errors.Is(err, sql.ErrNoRows) {
		h.redirect(c, "error", "not-found")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	h.redirect(c, "success", "revoked")
}
func (h *Handler) RevokeOthers(c *gin.Context) {
	current, _ := auth.CurrentUser(c)
	userID, err := strconv.ParseUint(current.ID, 10, 64)
	if err != nil {
		h.internalError(c, err)
		return
	}
	if err := h.service.RevokeOthers(c.Request.Context(), userID, auth.CurrentSessionID(c)); err != nil {
		h.internalError(c, err)
		return
	}
	h.redirect(c, "success", "others-revoked")
}
func (h *Handler) render(c *gin.Context, templateName string, data gin.H) {
	current, _ := auth.CurrentUser(c)
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
	data["AppName"], data["IsDashboard"], data["Authenticated"] = h.appName, true, true
	data["Section"] = "Akun"
	data["UserName"], data["UserInitials"] = current.Name, initials(current.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = current.StoreName, current.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"] = message(c.Query("success"))
	data["ErrorMessage"] = message(c.Query("error"))
	c.HTML(http.StatusOK, templateName, data)
}
func (h *Handler) redirect(c *gin.Context, key, value string) {
	c.Redirect(http.StatusSeeOther, "/profile/sessions?"+key+"="+url.QueryEscape(value))
}
func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("session management", "error", err)
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
	return map[string]string{"revoked": "Perangkat berhasil dikeluarkan.", "others-revoked": "Semua perangkat lain berhasil dikeluarkan.", "not-found": "Sesi tidak ditemukan atau merupakan sesi yang sedang digunakan."}[code]
}
