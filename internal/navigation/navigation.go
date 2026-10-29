package navigation

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"mksiaga/internal/auth"
)

type Item struct {
	Key      string
	Label    string
	URL      string
	Active   bool
	Open     bool
	Children []Item
}

type Group struct {
	Label string
	Items []Item
}

type definition struct {
	group      string
	key        string
	label      string
	url        string
	permission string
	parent     string
}

var definitions = []definition{
	{group: "Menu utama", key: "dashboard", label: "Dashboard", url: "/"},
	{group: "Laporan", key: "daily-reports", label: "Laporan Harian CCTV", url: "/reports/daily", permission: "daily-reports.view"},
	{group: "Laporan", key: "security-reports", label: "Laporan Shift Security", url: "/reports/security", permission: "security-shift-reports.view"},
	{group: "Laporan", key: "security-history", label: "Riwayat Laporan Shift", url: "/reports/security/history", permission: "security-shift-reports.view"},
	{group: "Laporan", key: "security-monitoring", label: "Monitoring SPV", url: "/reports/security/monitoring", permission: "security-shift-reports.review"},
	{group: "Laporan", key: "daily-recap", label: "Rekap Laporan CCTV", url: "/reports/recap", permission: "daily-reports.view"},
	{group: "Keamanan", parent: "blmkc", key: "blmkc-dashboard", label: "Dashboard", url: "/blmkc", permission: "blmkc.view"},
	{group: "Keamanan", parent: "blmkc", key: "blmkc-cases", label: "Daftar Kasus", url: "/blmkc/cases", permission: "blmkc.view"},
	{group: "Keamanan", parent: "blmkc", key: "blmkc-wanted", label: "Wanted / DPO", url: "/blmkc/wanted", permission: "blmkc.view"},
	{group: "Keamanan", parent: "blmkc", key: "blmkc-identity", label: "Pelaku", url: "/blmkc/identity", permission: "blmkc.view"},
	{group: "Keamanan", parent: "blmkc", key: "blmkc-verification", label: "Verifikasi", url: "/blmkc/verification", permission: "blmkc.review"},
	{group: "Master Data", parent: "master", key: "stores", label: "Store", url: "/master/stores", permission: "stores.view"},
	{group: "Master Aplikasi", parent: "appmaster", key: "areas", label: "Area", url: "/master/areas", permission: "areas.view"},
	{group: "Master Aplikasi", parent: "appmaster", key: "cameras", label: "Kamera", url: "/master/cameras", permission: "cameras.view"},
	{group: "Master Aplikasi", parent: "appmaster", key: "shifts", label: "Shift", url: "/master/shifts", permission: "shifts.view"},
	{group: "Master Aplikasi", parent: "appmaster", key: "checklists", label: "Checklist", url: "/master/checklists", permission: "checklists.view"},
	{group: "Master Aplikasi", parent: "appmaster", key: "incident-types", label: "Jenis Insiden", url: "/master/incident-types", permission: "incident-types.view"},
	{group: "Master Aplikasi", parent: "appmaster", key: "actions", label: "Tindakan", url: "/master/actions", permission: "actions.view"},
	{group: "Pengelolaan", parent: "access", key: "users", label: "User", url: "/settings/users", permission: "users.view"},
	{group: "Pengelolaan", parent: "access", key: "roles", label: "Role", url: "/settings/roles", permission: "roles.view"},
	{group: "Pengelolaan", parent: "access", key: "permissions", label: "Permission", url: "/settings/permissions", permission: "permissions.view"},
}

type Service struct{ db *sql.DB }

func NewService(db *sql.DB) *Service { return &Service{db: db} }

func (s *Service) Menus(ctx context.Context, user auth.SessionUser, activePage string) ([]Group, error) {
	allowed := make(map[string]bool)
	if hasRole(user.RoleName, "superadmin") {
		for _, item := range definitions {
			allowed[item.permission] = true
		}
	} else {
		userStoreID, err := strconv.ParseUint(user.UserStoreID, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid user store session: %w", err)
		}
		rows, err := s.db.QueryContext(ctx, `SELECT DISTINCT p.name
			FROM permissions p
			WHERE EXISTS(SELECT 1 FROM user_store_permissions usp WHERE usp.user_store_id=? AND usp.permission_id=p.id)
			   OR EXISTS(SELECT 1 FROM user_store_roles usr JOIN role_permissions rp ON rp.role_id=usr.role_id WHERE usr.user_store_id=? AND rp.permission_id=p.id)`, userStoreID, userStoreID)
		if err != nil {
			return nil, fmt.Errorf("load navigation permissions: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, err
			}
			allowed[name] = true
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	groups := make([]Group, 0, 2)
	indexes := make(map[string]int)
	children := make(map[string][]Item)
	for _, definition := range definitions {
		if definition.permission != "" && !allowed[definition.permission] {
			continue
		}
		item := Item{Key: definition.key, Label: definition.label, URL: definition.url, Active: definition.key == activePage}
		if definition.parent != "" {
			children[definition.parent] = append(children[definition.parent], item)
			continue
		}
		index, exists := indexes[definition.group]
		if !exists {
			index = len(groups)
			indexes[definition.group] = index
			groups = append(groups, Group{Label: definition.group})
		}
		groups[index].Items = append(groups[index].Items, item)
	}
	if accessItems := children["access"]; len(accessItems) > 0 {
		index, exists := indexes["Pengelolaan"]
		if !exists {
			index = len(groups)
			indexes["Pengelolaan"] = index
			groups = append(groups, Group{Label: "Pengelolaan"})
		}
		open := false
		for _, item := range accessItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "access", Label: "Manajemen Akses", Open: open, Active: open, Children: accessItems})
	}
	if masterItems := children["master"]; len(masterItems) > 0 {
		index, exists := indexes["Master Data"]
		if !exists {
			index = len(groups)
			indexes["Master Data"] = index
			groups = append(groups, Group{Label: "Master Data"})
		}
		open := false
		for _, item := range masterItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "master", Label: "Master Data", Open: open, Active: open, Children: masterItems})
	}
	if appMasterItems := children["appmaster"]; len(appMasterItems) > 0 {
		index, exists := indexes["Master Aplikasi"]
		if !exists {
			index = len(groups)
			indexes["Master Aplikasi"] = index
			groups = append(groups, Group{Label: "Master Aplikasi"})
		}
		open := false
		for _, item := range appMasterItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "appmaster", Label: "Master Aplikasi", Open: open, Active: open, Children: appMasterItems})
	}
	if blmkcItems := children["blmkc"]; len(blmkcItems) > 0 {
		index, exists := indexes["Keamanan"]
		if !exists {
			index = len(groups)
			indexes["Keamanan"] = index
			groups = append(groups, Group{Label: "Keamanan"})
		}
		open := false
		for _, item := range blmkcItems {
			open = open || item.Active
		}
		groups[index].Items = append(groups[index].Items, Item{Key: "blmkc", Label: "BLMKC", Open: open, Active: open, Children: blmkcItems})
	}
	return groups, nil
}

func hasRole(roleNames, required string) bool {
	for _, role := range strings.Split(roleNames, ",") {
		if strings.TrimSpace(role) == required {
			return true
		}
	}
	return false
}
