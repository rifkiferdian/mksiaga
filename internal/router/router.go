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
	actionManagement "mksiaga/internal/action"
	areaManagement "mksiaga/internal/area"
	"mksiaga/internal/auth"
	cameraManagement "mksiaga/internal/camera"
	checklistManagement "mksiaga/internal/checklist"
	"mksiaga/internal/config"
	"mksiaga/internal/dailyreport"
	"mksiaga/internal/dashboard"
	"mksiaga/internal/httperror"
	incidentTypeManagement "mksiaga/internal/incidenttype"
	"mksiaga/internal/navigation"
	"mksiaga/internal/profile"
	"mksiaga/internal/sessionmanagement"
	shiftManagement "mksiaga/internal/shift"
	storeManagement "mksiaga/internal/store"
	"mksiaga/internal/storescope"
	userManagement "mksiaga/internal/user"
	"mksiaga/internal/view"
)

func New(cfg config.Config, db *sql.DB, webRoot string) (*gin.Engine, error) {
	renderer, err := view.New(filepath.Join(webRoot, "templates"))
	if err != nil {
		return nil, err
	}
	gin.SetMode(cfg.GinMode)
	r := gin.New()
	r.Use(gin.Logger(), httperror.Recovery(cfg.AppName))
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
	sweetAlertRoot := filepath.Join(filepath.Dir(webRoot), "node_modules", "sweetalert2", "dist")
	r.StaticFile("/vendor/sweetalert2/sweetalert2.all.min.js", filepath.Join(sweetAlertRoot, "sweetalert2.all.min.js"))
	r.GET("/errors/403", httperror.Handler(cfg.AppName, "403"))
	r.GET("/errors/404", httperror.Handler(cfg.AppName, "404"))
	r.GET("/errors/419", httperror.Handler(cfg.AppName, "419"))
	r.GET("/errors/500", httperror.Handler(cfg.AppName, "500"))
	r.GET("/errors/maintenance", httperror.Handler(cfg.AppName, "maintenance"))
	if cfg.Maintenance.Enabled {
		r.Use(httperror.Maintenance(cfg.AppName, cfg.Maintenance.Message))
	}
	r.Use(auth.RequireCSRF())
	var authService *auth.Service
	if db != nil {
		authService = auth.NewService(auth.NewRepository(db))
	}
	authHandler := auth.NewHandler(cfg.AppName, authService, cfg.Session.CookieSecure)
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	authorized := r.Group("/")
	authorized.Use(auth.RequireLogin(db))
	var navigationService *navigation.Service
	if db != nil {
		navigationService = navigation.NewService(db)
	}
	authorized.GET("/", dashboard.NewHandler(cfg.AppName, navigationService).Index)
	authorized.POST("/logout", authHandler.Logout)
	if db != nil {
		accessHandler := access.NewHandler(cfg.AppName, access.NewService(access.NewRepository(db)), navigationService)
		actionHandler := actionManagement.NewHandler(cfg.AppName, actionManagement.NewService(actionManagement.NewRepository(db)), navigationService)
		areaHandler := areaManagement.NewHandler(cfg.AppName, areaManagement.NewService(areaManagement.NewRepository(db)), navigationService)
		cameraHandler := cameraManagement.NewHandler(cfg.AppName, cameraManagement.NewService(cameraManagement.NewRepository(db)), navigationService)
		checklistHandler := checklistManagement.NewHandler(cfg.AppName, checklistManagement.NewService(checklistManagement.NewRepository(db)), navigationService)
		dailyReportHandler := dailyreport.NewHandler(cfg.AppName, dailyreport.NewService(dailyreport.NewRepository(db)), navigationService)
		incidentTypeHandler := incidentTypeManagement.NewHandler(cfg.AppName, incidentTypeManagement.NewService(incidentTypeManagement.NewRepository(db)), navigationService)
		profileHandler := profile.NewHandler(cfg.AppName, profile.NewService(profile.NewRepository(db)), navigationService)
		storeHandler := storeManagement.NewHandler(cfg.AppName, storeManagement.NewService(storeManagement.NewRepository(db)), navigationService)
		sessionHandler := sessionmanagement.NewHandler(cfg.AppName, sessionmanagement.NewService(sessionmanagement.NewRepository(db)), navigationService)
		shiftHandler := shiftManagement.NewHandler(cfg.AppName, shiftManagement.NewService(shiftManagement.NewRepository(db)), navigationService)
		userHandler := userManagement.NewHandler(cfg.AppName, userManagement.NewService(userManagement.NewRepository(db)), navigationService)
		reports := authorized.Group("/reports")
		reports.Use(storescope.Middleware(db))
		reports.GET("/daily", auth.RequirePermission(db, "daily-reports.view"), dailyReportHandler.Index)
		reports.GET("/daily/new", auth.RequirePermission(db, "daily-reports.create"), dailyReportHandler.New)
		reports.POST("/daily", auth.RequirePermission(db, "daily-reports.create"), dailyReportHandler.Create)
		reports.POST("/daily/:id/delete", auth.RequirePermission(db, "daily-reports.delete"), dailyReportHandler.Delete)
		authorized.GET("/profile", profileHandler.Index)
		authorized.POST("/profile/name", profileHandler.UpdateName)
		authorized.POST("/profile/password", profileHandler.UpdatePassword)
		authorized.GET("/profile/sessions", sessionHandler.Index)
		authorized.POST("/profile/sessions/revoke-others", sessionHandler.RevokeOthers)
		authorized.POST("/profile/sessions/:id/revoke", sessionHandler.Revoke)
		settings := authorized.Group("/settings")
		settings.GET("/users", auth.RequirePermission(db, "users.view"), userHandler.Index)
		settings.POST("/users", auth.RequirePermission(db, "users.create"), userHandler.Create)
		settings.POST("/users/:id/update", auth.RequirePermission(db, "users.update"), userHandler.Update)
		settings.POST("/users/:id/delete", auth.RequirePermission(db, "users.delete"), userHandler.Delete)
		settings.GET("/users/:id/stores", auth.RequirePermission(db, "users.update"), userHandler.StoreAssignments)
		settings.POST("/users/:id/stores", auth.RequirePermission(db, "users.update"), userHandler.UpdateStoreAssignments)
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
		master := authorized.Group("/master")
		master.Use(storescope.Middleware(db))
		master.GET("/stores", auth.RequirePermission(db, "stores.view"), storeHandler.Index)
		master.POST("/stores", auth.RequirePermission(db, "stores.create"), storeHandler.Create)
		master.POST("/stores/:id/update", auth.RequirePermission(db, "stores.update"), storeHandler.Update)
		master.POST("/stores/:id/delete", auth.RequirePermission(db, "stores.delete"), storeHandler.Delete)
		master.GET("/areas", auth.RequirePermission(db, "areas.view"), areaHandler.Index)
		master.POST("/areas", auth.RequirePermission(db, "areas.create"), areaHandler.Create)
		master.POST("/areas/:id/update", auth.RequirePermission(db, "areas.update"), areaHandler.Update)
		master.POST("/areas/:id/delete", auth.RequirePermission(db, "areas.delete"), areaHandler.Delete)
		master.GET("/checklists", auth.RequirePermission(db, "checklists.view"), checklistHandler.Index)
		master.POST("/checklists", auth.RequirePermission(db, "checklists.create"), checklistHandler.Create)
		master.POST("/checklists/:id/update", auth.RequirePermission(db, "checklists.update"), checklistHandler.Update)
		master.POST("/checklists/:id/delete", auth.RequirePermission(db, "checklists.delete"), checklistHandler.Delete)
		master.GET("/incident-types", auth.RequirePermission(db, "incident-types.view"), incidentTypeHandler.Index)
		master.POST("/incident-types", auth.RequirePermission(db, "incident-types.create"), incidentTypeHandler.Create)
		master.POST("/incident-types/:id/update", auth.RequirePermission(db, "incident-types.update"), incidentTypeHandler.Update)
		master.POST("/incident-types/:id/delete", auth.RequirePermission(db, "incident-types.delete"), incidentTypeHandler.Delete)
		master.GET("/cameras", auth.RequirePermission(db, "cameras.view"), cameraHandler.Index)
		master.POST("/cameras", auth.RequirePermission(db, "cameras.create"), cameraHandler.Create)
		master.POST("/cameras/:id/update", auth.RequirePermission(db, "cameras.update"), cameraHandler.Update)
		master.POST("/cameras/:id/delete", auth.RequirePermission(db, "cameras.delete"), cameraHandler.Delete)
		master.GET("/shifts", auth.RequirePermission(db, "shifts.view"), shiftHandler.Index)
		master.POST("/shifts", auth.RequirePermission(db, "shifts.create"), shiftHandler.Create)
		master.POST("/shifts/:id/update", auth.RequirePermission(db, "shifts.update"), shiftHandler.Update)
		master.POST("/shifts/:id/delete", auth.RequirePermission(db, "shifts.delete"), shiftHandler.Delete)
		master.GET("/actions", auth.RequirePermission(db, "actions.view"), actionHandler.Index)
		master.POST("/actions", auth.RequirePermission(db, "actions.create"), actionHandler.Create)
		master.POST("/actions/:id/update", auth.RequirePermission(db, "actions.update"), actionHandler.Update)
		master.POST("/actions/:id/delete", auth.RequirePermission(db, "actions.delete"), actionHandler.Delete)
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
	r.NoRoute(httperror.Handler(cfg.AppName, "404"))
	return r, nil
}
