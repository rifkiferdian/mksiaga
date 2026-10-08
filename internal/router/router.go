package router

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
	"mksiaga/internal/config"
	"mksiaga/internal/dashboard"
	"mksiaga/internal/view"
)

func New(cfg config.Config, db *sql.DB, webRoot string) (*gin.Engine, error) {
	renderer, err := view.New(filepath.Join(webRoot, "templates"))
	if err != nil {
		return nil, err
	}
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	store := cookie.NewStore([]byte(cfg.Session.Secret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   12 * 60 * 60,
		HttpOnly: true,
		Secure:   cfg.Session.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions(auth.SessionName, store))
	if err := r.SetTrustedProxies(nil); err != nil {
		return nil, err
	}
	r.HTMLRender = renderer
	r.Static("/static", filepath.Join(webRoot, "static"))
	var authService *auth.Service
	if db != nil {
		authService = auth.NewService(auth.NewRepository(db))
	}
	authHandler := auth.NewHandler(cfg.AppName, authService, cfg.Session.CookieSecure)
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	authorized := r.Group("/")
	authorized.Use(auth.RequireLogin())
	authorized.GET("/", dashboard.NewHandler(cfg.AppName).Index)
	authorized.POST("/logout", authHandler.Logout)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready", "database": "disabled"})
			return
		}
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		if err := db.PingContext(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not_ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
	return r, nil
}
