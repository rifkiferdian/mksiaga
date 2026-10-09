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
	"mksiaga/internal/access"
	"mksiaga/internal/auth"
	"mksiaga/internal/config"
	"mksiaga/internal/dashboard"
	"mksiaga/internal/navigation"
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
	fontAwesomeRoot := filepath.Join(filepath.Dir(webRoot), "node_modules", "@fortawesome", "fontawesome-free")
	r.StaticFile("/vendor/fontawesome/css/all.min.css", filepath.Join(fontAwesomeRoot, "css", "all.min.css"))
	r.Static("/vendor/fontawesome/webfonts", filepath.Join(fontAwesomeRoot, "webfonts"))
	var authService *auth.Service
	if db != nil {
		authService = auth.NewService(auth.NewRepository(db))
	}
	authHandler := auth.NewHandler(cfg.AppName, authService, cfg.Session.CookieSecure)
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	authorized := r.Group("/")
	authorized.Use(auth.RequireLogin())
	var navigationService *navigation.Service
	if db != nil {
		navigationService = navigation.NewService(db)
	}
	authorized.GET("/", dashboard.NewHandler(cfg.AppName, navigationService).Index)
	authorized.POST("/logout", authHandler.Logout)
	if db != nil {
		accessHandler := access.NewHandler(cfg.AppName, access.NewService(access.NewRepository(db)), navigationService)
		settings := authorized.Group("/settings")
		settings.GET("/roles", auth.RequirePermission(db, "roles.view"), accessHandler.Roles)
		settings.POST("/roles", auth.RequirePermission(db, "roles.create"), accessHandler.CreateRole)
		settings.POST("/roles/:id/update", auth.RequirePermission(db, "roles.update"), accessHandler.UpdateRole)
		settings.POST("/roles/:id/delete", auth.RequirePermission(db, "roles.delete"), accessHandler.DeleteRole)
		settings.GET("/roles/:id/permissions", auth.RequirePermission(db, "roles.update"), accessHandler.RolePermissions)
		settings.POST("/roles/:id/permissions", auth.RequirePermission(db, "roles.update"), accessHandler.UpdateRolePermissions)
		settings.GET("/permissions", auth.RequirePermission(db, "permissions.view"), accessHandler.Permissions)
		settings.POST("/permissions", auth.RequirePermission(db, "permissions.create"), accessHandler.CreatePermission)
		settings.POST("/permissions/:id/update", auth.RequirePermission(db, "permissions.update"), accessHandler.UpdatePermission)
		settings.POST("/permissions/:id/delete", auth.RequirePermission(db, "permissions.delete"), accessHandler.DeletePermission)
	}
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
