// Package storescope limits operational masters to the active store.
package storescope

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"mksiaga/internal/auth"
)

type contextKey struct{}
type scope struct {
	StoreID uint64
	All     bool
}
type Store struct {
	ID         uint64
	Code, Name string
}

// ForStore is also used by integration tests and background operations.
func ForStore(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, contextKey{}, scope{StoreID: id})
}
func ForAll(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKey{}, scope{All: true})
}
func Allows(ctx context.Context, id uint64) bool {
	s, ok := ctx.Value(contextKey{}).(scope)
	return ok && id > 0 && (s.All || s.StoreID == id)
}

// column must be a constant SQL identifier supplied by repository code.
// Missing scope denies access. Only a parsed numeric ID is interpolated.
func Clause(ctx context.Context, column string) string {
	s, ok := ctx.Value(contextKey{}).(scope)
	if !ok || (!s.All && s.StoreID == 0) {
		return "1=0"
	}
	if s.All {
		return "1=1"
	}
	return column + "=" + strconv.FormatUint(s.StoreID, 10)
}

func Middleware(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := auth.CurrentUser(c)
		if !ok {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		for _, role := range strings.Split(user.RoleName, ",") {
			if strings.TrimSpace(role) == "superadmin" {
				c.Request = c.Request.WithContext(ForAll(c.Request.Context()))
				c.Next()
				return
			}
		}
		id, err := strconv.ParseUint(user.StoreID, 10, 64)
		var allowed bool
		if err == nil && id > 0 {
			err = db.QueryRowContext(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM user_stores us JOIN stores s ON s.id=us.store_id WHERE us.id=? AND us.user_id=? AND us.store_id=? AND us.status='active' AND s.status='active' AND s.deleted_at IS NULL)`, user.UserStoreID, user.ID, id).Scan(&allowed)
		}
		if err != nil || !allowed {
			c.Redirect(http.StatusSeeOther, "/errors/403")
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(ForStore(c.Request.Context(), id))
		c.Next()
	}
}

func Stores(ctx context.Context, db *sql.DB) ([]Store, error) {
	rows, err := db.QueryContext(ctx, `SELECT id,code,name FROM stores WHERE deleted_at IS NULL AND `+Clause(ctx, "id")+` ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []Store
	for rows.Next() {
		var item Store
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, rows.Err()
}

func ValidStore(ctx context.Context, db *sql.DB, id uint64) (bool, error) {
	if !Allows(ctx, id) {
		return false, nil
	}
	var valid bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM stores WHERE id=? AND deleted_at IS NULL)`, id).Scan(&valid)
	return valid, err
}
