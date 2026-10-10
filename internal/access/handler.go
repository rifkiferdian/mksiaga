package access

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

func (h *Handler) Roles(c *gin.Context) {
	roles, err := h.service.Roles(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	permissionAssignments, userAssignments := 0, 0
	for _, role := range roles {
		permissionAssignments += role.PermissionCount
		userAssignments += role.UserCount
	}
	h.render(c, http.StatusOK, "access/roles.html", gin.H{
		"Title": "Role", "ActivePage": "roles", "Roles": roles,
		"RoleCount": len(roles), "RolePermissionCount": permissionAssignments, "RoleUserCount": userAssignments,
	})
}

func (h *Handler) CreateRole(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/roles", "error", "csrf")
		return
	}
	err := h.service.CreateRole(c.Request.Context(), c.PostForm("name"), c.PostForm("guard_name"), c.PostForm("description"))
	h.redirectResult(c, "/settings/roles", err, "role-created")
}

func (h *Handler) UpdateRole(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/roles", "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "/settings/roles", "error", "invalid")
		return
	}
	err := h.service.UpdateRole(c.Request.Context(), id, c.PostForm("name"), c.PostForm("guard_name"), c.PostForm("description"))
	h.redirectResult(c, "/settings/roles", err, "role-updated")
}

func (h *Handler) DeleteRole(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/roles", "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "/settings/roles", "error", "invalid")
		return
	}
	err := h.service.DeleteRole(c.Request.Context(), id)
	h.redirectResult(c, "/settings/roles", err, "role-deleted")
}

func (h *Handler) RolePermissions(c *gin.Context) {
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "/settings/roles", "error", "invalid")
		return
	}
	role, permissions, err := h.service.Role(c.Request.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		h.redirect(c, "/settings/roles", "error", "not-found")
		return
	}
	if err != nil {
		h.internalError(c, err)
		return
	}
	groups, assignedCount := groupPermissions(permissions)
	h.render(c, http.StatusOK, "access/role_permissions.html", gin.H{
		"Title": "Permission role", "ActivePage": "roles", "Role": role,
		"PermissionGroups": groups, "PermissionCount": len(permissions), "AssignedPermissionCount": assignedCount,
	})
}

func (h *Handler) UpdateRolePermissions(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/roles", "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "/settings/roles", "error", "invalid")
		return
	}
	var permissionIDs []uint64
	for _, raw := range c.PostFormArray("permission_ids") {
		permissionID, err := strconv.ParseUint(raw, 10, 64)
		if err != nil || permissionID == 0 {
			h.redirect(c, "/settings/roles", "error", "invalid")
			return
		}
		permissionIDs = append(permissionIDs, permissionID)
	}
	err := h.service.SyncRolePermissions(c.Request.Context(), id, permissionIDs)
	h.redirectResult(c, "/settings/roles", err, "role-permissions-updated")
}

func (h *Handler) Permissions(c *gin.Context) {
	permissions, err := h.service.Permissions(c.Request.Context())
	if err != nil {
		h.internalError(c, err)
		return
	}
	roleAssignments, directAssignments, unassigned := 0, 0, 0
	for _, permission := range permissions {
		roleAssignments += permission.RoleCount
		directAssignments += permission.DirectCount
		if permission.RoleCount == 0 && permission.DirectCount == 0 {
			unassigned++
		}
	}
	h.render(c, http.StatusOK, "access/permissions.html", gin.H{
		"Title": "Permission", "ActivePage": "permissions", "Permissions": permissions,
		"PermissionCount": len(permissions), "PermissionRoleCount": roleAssignments,
		"PermissionDirectCount": directAssignments, "PermissionUnassignedCount": unassigned,
	})
}

func (h *Handler) CreatePermission(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/permissions", "error", "csrf")
		return
	}
	err := h.service.CreatePermission(c.Request.Context(), c.PostForm("name"), c.PostForm("guard_name"), c.PostForm("description"))
	h.redirectResult(c, "/settings/permissions", err, "permission-created")
}

func (h *Handler) UpdatePermission(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/permissions", "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "/settings/permissions", "error", "invalid")
		return
	}
	err := h.service.UpdatePermission(c.Request.Context(), id, c.PostForm("name"), c.PostForm("guard_name"), c.PostForm("description"))
	h.redirectResult(c, "/settings/permissions", err, "permission-updated")
}

func (h *Handler) DeletePermission(c *gin.Context) {
	if !auth.ValidCSRF(c, c.PostForm("csrf_token")) {
		h.redirect(c, "/settings/permissions", "error", "csrf")
		return
	}
	id, ok := routeID(c)
	if !ok {
		h.redirect(c, "/settings/permissions", "error", "invalid")
		return
	}
	err := h.service.DeletePermission(c.Request.Context(), id)
	h.redirectResult(c, "/settings/permissions", err, "permission-deleted")
}

func (h *Handler) render(c *gin.Context, status int, templateName string, data gin.H) {
	user, _ := auth.CurrentUser(c)
	activePage, _ := data["ActivePage"].(string)
	menus, err := h.navigation.Menus(c.Request.Context(), user, activePage)
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
	data["Section"] = "Pengelolaan"
	data["UserName"], data["UserInitials"] = user.Name, userInitials(user.Name)
	data["StoreName"], data["RoleName"], data["CSRFToken"] = user.StoreName, user.RoleName, token
	data["Menus"] = menus
	data["SuccessMessage"] = feedbackMessage(c.Query("success"))
	data["ErrorMessage"] = feedbackMessage(c.Query("error"))
	c.HTML(status, templateName, data)
}

func (h *Handler) redirectResult(c *gin.Context, path string, err error, success string) {
	if err == nil {
		h.redirect(c, path, "success", success)
		return
	}
	code := "failed"
	switch {
	case errors.Is(err, ErrInvalidName):
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
		slog.Error("access management operation failed", "error", err)
	}
	h.redirect(c, path, "error", code)
}

func (h *Handler) redirect(c *gin.Context, path, key, value string) {
	c.Redirect(http.StatusSeeOther, path+"?"+key+"="+url.QueryEscape(value))
}

func (h *Handler) internalError(c *gin.Context, err error) {
	slog.Error("render access management", "error", err)
	c.Redirect(http.StatusSeeOther, "/errors/500")
}

func routeID(c *gin.Context) (uint64, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	return id, err == nil && id > 0
}

func userInitials(name string) string {
	parts := strings.Fields(name)
	if len(parts) == 0 {
		return "US"
	}
	result := string([]rune(strings.ToUpper(parts[0]))[0])
	if len(parts) > 1 {
		result += string([]rune(strings.ToUpper(parts[len(parts)-1]))[0])
	}
	return result
}

func feedbackMessage(code string) string {
	return map[string]string{
		"role-created": "Role berhasil dibuat.", "role-updated": "Role berhasil diperbarui.", "role-deleted": "Role berhasil dihapus.", "role-permissions-updated": "Permission role berhasil disimpan.",
		"permission-created": "Permission berhasil dibuat.", "permission-updated": "Permission berhasil diperbarui.", "permission-deleted": "Permission berhasil dihapus.",
		"csrf": "Sesi form berakhir. Silakan muat ulang halaman.", "invalid": "Nama hanya boleh berisi huruf kecil, angka, titik, garis bawah, atau tanda hubung.",
		"duplicate": "Nama dengan guard tersebut sudah tersedia.", "protected": "Role superadmin tidak dapat dihapus.", "in-use": "Data masih digunakan dan tidak dapat dihapus.",
		"not-found": "Data tidak ditemukan.", "failed": "Operasi gagal diproses.",
	}[code]
}

func groupPermissions(permissions []Permission) ([]PermissionGroup, int) {
	groups := make([]PermissionGroup, 0)
	indexes := make(map[string]int)
	assignedTotal := 0
	for _, permission := range permissions {
		key, _, _ := strings.Cut(permission.Name, ".")
		index, exists := indexes[key]
		if !exists {
			label, description := permissionGroupMeta(key)
			index = len(groups)
			indexes[key] = index
			groups = append(groups, PermissionGroup{Key: key, Label: label, Description: description})
		}
		groups[index].Permissions = append(groups[index].Permissions, permission)
		if permission.Assigned {
			groups[index].AssignedCount++
			assignedTotal++
		}
	}
	return groups, assignedTotal
}

func permissionGroupMeta(key string) (string, string) {
	switch key {
	case "roles":
		return "Role", "Mengelola role dan assignment permission."
	case "permissions":
		return "Permission", "Mengelola daftar kemampuan dalam sistem."
	case "users":
		return "User", "Mengelola akun, store, dan akses staff."
	default:
		if key == "" {
			return "Lainnya", "Permission umum aplikasi."
		}
		runes := []rune(key)
		return strings.ToUpper(string(runes[0])) + string(runes[1:]), "Permission untuk modul " + key + "."
	}
}
