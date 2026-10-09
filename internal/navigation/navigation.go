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
	Key    string
	Label  string
	URL    string
	Active bool
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
}

var definitions = []definition{
	{group: "Menu utama", key: "dashboard", label: "Dashboard", url: "/"},
	{group: "Pengelolaan", key: "roles", label: "Role", url: "/settings/roles", permission: "roles.view"},
	{group: "Pengelolaan", key: "permissions", label: "Permission", url: "/settings/permissions", permission: "permissions.view"},
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
	for _, definition := range definitions {
		if definition.permission != "" && !allowed[definition.permission] {
			continue
		}
		index, exists := indexes[definition.group]
		if !exists {
			index = len(groups)
			indexes[definition.group] = index
			groups = append(groups, Group{Label: definition.group})
		}
		groups[index].Items = append(groups[index].Items, Item{Key: definition.key, Label: definition.label, URL: definition.url, Active: definition.key == activePage})
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
