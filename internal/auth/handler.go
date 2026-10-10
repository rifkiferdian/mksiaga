package auth

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	appName      string
	service      *Service
	cookieSecure bool
}

func NewHandler(appName string, service *Service, cookieSecure bool) *Handler {
	return &Handler{appName: appName, service: service, cookieSecure: cookieSecure}
}

func (h *Handler) ShowLogin(c *gin.Context) {
	if _, ok := CurrentUser(c); ok {
		c.Redirect(http.StatusSeeOther, "/")
		return
	}
	h.renderLogin(c, http.StatusOK, "", "")
}

func (h *Handler) Login(c *gin.Context) {
	login := c.PostForm("username")
	if !ValidCSRF(c, c.PostForm("csrf_token")) {
		h.renderLogin(c, http.StatusBadRequest, "Sesi form telah berakhir. Silakan coba kembali.", login)
		return
	}
	if h.service == nil {
		h.renderLogin(c, http.StatusServiceUnavailable, "Database belum tersedia. Hubungi administrator aplikasi.", login)
		return
	}

	user, err := h.service.Authenticate(c.Request.Context(), login, c.PostForm("password"))
	if errors.Is(err, ErrInvalidCredentials) {
		h.renderLogin(c, http.StatusUnauthorized, "Username/email atau password tidak sesuai.", login)
		return
	}
	if err != nil {
		slog.Error("login failed", "error", err)
		h.renderLogin(c, http.StatusInternalServerError, "Terjadi gangguan saat memproses login. Silakan coba kembali.", login)
		return
	}

	remember := c.PostForm("remember") == "1"
	sessionID, err := NewSessionID()
	if err == nil {
		err = h.service.CreateSession(c.Request.Context(), sessionID, user, clipped(c.ClientIP(), 45), clipped(c.Request.UserAgent(), 512), time.Now().Add(SessionLifetime(remember)))
	}
	if err != nil {
		slog.Error("create tracked login session", "error", err)
		h.renderLogin(c, http.StatusInternalServerError, "Sesi login tidak dapat dibuat. Silakan coba kembali.", login)
		return
	}
	if err := SaveLogin(c, user, sessionID, remember, h.cookieSecure); err != nil {
		_ = h.service.RevokeSession(c.Request.Context(), sessionID)
		slog.Error("save login session", "error", err)
		h.renderLogin(c, http.StatusInternalServerError, "Sesi login tidak dapat dibuat. Silakan coba kembali.", login)
		return
	}
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *Handler) Logout(c *gin.Context) {
	if !ValidCSRF(c, c.PostForm("csrf_token")) {
		c.String(http.StatusBadRequest, "Permintaan tidak valid.")
		return
	}
	if h.service != nil {
		_ = h.service.RevokeSession(c.Request.Context(), CurrentSessionID(c))
	}
	if err := ClearLogin(c, h.cookieSecure); err != nil {
		slog.Error("clear login session", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	c.Redirect(http.StatusSeeOther, "/login")
}

func clipped(value string, limit int) string {
	if len(value) > limit {
		return value[:limit]
	}
	return value
}

func (h *Handler) renderLogin(c *gin.Context, status int, message, login string) {
	token, err := CSRFToken(c)
	if err != nil {
		slog.Error("prepare login csrf", "error", err)
		c.Redirect(http.StatusSeeOther, "/errors/500")
		return
	}
	c.HTML(status, "auth/login.html", gin.H{
		"Title":      "Masuk",
		"AppName":    h.appName,
		"IsAuthPage": true,
		"Error":      message,
		"Login":      login,
		"CSRFToken":  token,
	})
}
