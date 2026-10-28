package router

import (
	"context"
	"database/sql"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"mksiaga/internal/access"
	actionManagement "mksiaga/internal/action"
	areaManagement "mksiaga/internal/area"
	"mksiaga/internal/auth"
	"mksiaga/internal/blmkc"
	cameraManagement "mksiaga/internal/camera"
	checklistManagement "mksiaga/internal/checklist"
	"mksiaga/internal/config"
	"mksiaga/internal/dailyreport"
	"mksiaga/internal/dashboard"
	"mksiaga/internal/httperror"
	incidentTypeManagement "mksiaga/internal/incidenttype"
	"mksiaga/internal/navigation"
	"mksiaga/internal/profile"
	"mksiaga/internal/securityreport"
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
	r.Use(func(c *gin.Context) {
		if c.Request.Body != nil && c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			limit := int64(2 << 20)
			if strings.HasPrefix(c.Request.URL.Path, "/blmkc/cases/") && (strings.HasSuffix(c.Request.URL.Path, "/evidence") || strings.HasSuffix(c.Request.URL.Path, "/identity-package") || strings.HasSuffix(c.Request.URL.Path, "/resolution")) {
				limit = 151 << 20
			}
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		}
		c.Next()
	})
	r.Use(auth.RequireCSRF())
	var authService *auth.Service
	if db != nil {
		authService = auth.NewService(auth.NewRepository(db))
	}
	authHandler := auth.NewHandler(cfg.AppName, authService, cfg.Session.CookieSecure)
	r.GET("/login", authHandler.ShowLogin)
	r.POST("/login", authHandler.Login)
	if db != nil {
		publicBLMKCHandler := blmkc.NewHandler(cfg.AppName, db, blmkc.NewService(blmkc.NewRepository(db), cfg.BLMKCDataKey), nil)
		r.GET("/wanted", publicBLMKCHandler.PublicWanted)
		r.GET("/wanted/cases/:id", publicBLMKCHandler.PublicWantedDetail)
		r.GET("/wanted/cases/:id/photos/:evidenceID", publicBLMKCHandler.PublicWantedPhoto)
	}
	authorized := r.Group("/")
	authorized.Use(auth.RequireLogin(db))
	var navigationService *navigation.Service
	if db != nil {
		navigationService = navigation.NewService(db)
	}
	authorized.GET("/", dashboard.NewHandler(cfg.AppName, db, navigationService).Index)
	authorized.POST("/logout", authHandler.Logout)
	if db != nil {
		accessHandler := access.NewHandler(cfg.AppName, access.NewService(access.NewRepository(db)), navigationService)
		actionHandler := actionManagement.NewHandler(cfg.AppName, actionManagement.NewService(actionManagement.NewRepository(db)), navigationService)
		areaHandler := areaManagement.NewHandler(cfg.AppName, areaManagement.NewService(areaManagement.NewRepository(db)), navigationService)
		cameraHandler := cameraManagement.NewHandler(cfg.AppName, cameraManagement.NewService(cameraManagement.NewRepository(db)), navigationService)
		checklistHandler := checklistManagement.NewHandler(cfg.AppName, checklistManagement.NewService(checklistManagement.NewRepository(db)), navigationService)
		dailyReportHandler := dailyreport.NewHandler(cfg.AppName, dailyreport.NewService(dailyreport.NewRepository(db)), navigationService)
		securityReportHandler := securityreport.NewHandler(cfg.AppName, securityreport.NewService(securityreport.NewRepository(db)), navigationService)
		blmkcHandler := blmkc.NewHandler(cfg.AppName, db, blmkc.NewService(blmkc.NewRepository(db), cfg.BLMKCDataKey), navigationService)
		incidentTypeHandler := incidentTypeManagement.NewHandler(cfg.AppName, incidentTypeManagement.NewService(incidentTypeManagement.NewRepository(db)), navigationService)
		profileHandler := profile.NewHandler(cfg.AppName, profile.NewService(profile.NewRepository(db)), navigationService)
		storeHandler := storeManagement.NewHandler(cfg.AppName, storeManagement.NewService(storeManagement.NewRepository(db)), navigationService)
		sessionHandler := sessionmanagement.NewHandler(cfg.AppName, sessionmanagement.NewService(sessionmanagement.NewRepository(db)), navigationService)
		shiftHandler := shiftManagement.NewHandler(cfg.AppName, shiftManagement.NewService(shiftManagement.NewRepository(db)), navigationService)
		userHandler := userManagement.NewHandler(cfg.AppName, userManagement.NewService(userManagement.NewRepository(db)), navigationService)
		reports := authorized.Group("/reports")
		reports.Use(storescope.Middleware(db))
		reports.GET("/daily", auth.RequirePermission(db, "daily-reports.view"), dailyReportHandler.Index)
		reports.GET("/recap", auth.RequirePermission(db, "daily-reports.view"), dailyReportHandler.Recap)
		reports.GET("/daily/new", auth.RequirePermission(db, "daily-reports.create"), dailyReportHandler.New)
		reports.POST("/daily", auth.RequirePermission(db, "daily-reports.create"), dailyReportHandler.Create)
		reports.GET("/daily/:id/workflow", auth.RequirePermission(db, "daily-reports.view"), dailyReportHandler.Workflow)
		reports.GET("/daily/:id/print", auth.RequirePermission(db, "daily-reports.view"), dailyReportHandler.Print)
		reports.POST("/daily/:id/receive", auth.RequirePermission(db, "daily-reports.view"), dailyReportHandler.Receive)
		reports.POST("/daily/:id/revise", auth.RequirePermission(db, "daily-reports.create"), dailyReportHandler.Revise)
		reports.POST("/daily/:id/review", auth.RequirePermission(db, "daily-reports.review"), dailyReportHandler.Review)
		reports.POST("/daily/:id/delete", auth.RequirePermission(db, "daily-reports.delete"), dailyReportHandler.Delete)
		reports.GET("/security", auth.RequirePermission(db, "security-shift-reports.view"), securityReportHandler.Index)
		reports.GET("/security/new", auth.RequirePermission(db, "security-shift-reports.create"), securityReportHandler.New)
		reports.POST("/security", auth.RequirePermission(db, "security-shift-reports.create"), securityReportHandler.Create)
		reports.GET("/security/:id", auth.RequirePermission(db, "security-shift-reports.view"), securityReportHandler.Show)
		reports.POST("/security/:id/hourly-logs", auth.RequirePermission(db, "security-shift-reports.create"), securityReportHandler.AddHourlyLog)
		blmkcRoutes := authorized.Group("/blmkc")
		blmkcRoutes.Use(storescope.Middleware(db))
		blmkcRoutes.GET("", auth.RequirePermission(db, "blmkc.view"), blmkcHandler.Dashboard)
		blmkcRoutes.GET("/cases", auth.RequirePermission(db, "blmkc.view"), blmkcHandler.Cases)
		blmkcRoutes.GET("/cases/new", auth.RequirePermission(db, "blmkc.create"), blmkcHandler.NewCase)
		blmkcRoutes.POST("/cases", auth.RequirePermission(db, "blmkc.create"), blmkcHandler.CreateCase)
		blmkcRoutes.GET("/cases/:id", auth.RequirePermission(db, "blmkc.view"), blmkcHandler.CaseDetail)
		blmkcRoutes.GET("/cases/:id/edit", auth.RequirePermission(db, "blmkc.update"), blmkcHandler.EditCase)
		blmkcRoutes.GET("/cases/:id/evidence/new", auth.RequirePermission(db, "blmkc.upload-evidence"), blmkcHandler.EvidenceStep)
		blmkcRoutes.POST("/cases/:id/evidence/complete", auth.RequirePermission(db, "blmkc.upload-evidence"), blmkcHandler.CompleteEvidenceStep)
		blmkcRoutes.GET("/cases/:id/identity-package", auth.RequirePermission(db, "blmkc.update"), blmkcHandler.IdentityPackageForm)
		blmkcRoutes.POST("/cases/:id/identity-package", auth.RequirePermission(db, "blmkc.update"), blmkcHandler.SaveIdentityPackage)
		blmkcRoutes.POST("/cases/:id/resolution", auth.RequirePermission(db, "blmkc.resolve"), blmkcHandler.SaveCaseResolution)
		blmkcRoutes.POST("/cases/:id/update", auth.RequirePermission(db, "blmkc.update"), blmkcHandler.UpdateCase)
		blmkcRoutes.POST("/cases/:id/advance-status", auth.RequirePermission(db, "blmkc.change-status"), blmkcHandler.AdvanceStatus)
		blmkcRoutes.POST("/cases/:id/evidence", auth.RequirePermission(db, "blmkc.upload-evidence"), blmkcHandler.UploadEvidence)
		blmkcRoutes.GET("/cases/:id/evidence/:evidenceID/preview", auth.RequirePermission(db, "blmkc.download-evidence"), blmkcHandler.PreviewEvidence)
		blmkcRoutes.GET("/cases/:id/evidence/:evidenceID/download", auth.RequirePermission(db, "blmkc.download-evidence"), blmkcHandler.DownloadEvidence)
		blmkcRoutes.GET("/wanted", auth.RequirePermission(db, "blmkc.view"), blmkcHandler.Wanted)
		blmkcRoutes.GET("/identity", auth.RequirePermission(db, "blmkc.view"), blmkcHandler.Identities)
		blmkcRoutes.GET("/identity/:id", auth.RequirePermission(db, "blmkc.view"), blmkcHandler.IdentityDetail)
		blmkcRoutes.GET("/verification", auth.RequirePermission(db, "blmkc.review"), blmkcHandler.Verifications)
		blmkcRoutes.GET("/verification/:id", auth.RequirePermission(db, "blmkc.review"), blmkcHandler.VerificationDetail)
		blmkcRoutes.POST("/verification/:id/review", auth.RequirePermission(db, "blmkc.review"), blmkcHandler.ReviewCase)
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
