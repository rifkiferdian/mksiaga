package blmkc

import (
	"log/slog"
	"net/http"
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
	h.render(c, "blmkc/cases.html", "Daftar Kasus BLMKC", "blmkc-cases", gin.H{"Data": data})
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
