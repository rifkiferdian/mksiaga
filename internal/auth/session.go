package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionName        = "mksiaga_session"
	sessionUserID      = "user_id"
	sessionUserName    = "user_name"
	sessionStoreID     = "store_id"
	sessionUserStoreID = "user_store_id"
	sessionStoreName   = "store_name"
	sessionRoleName    = "role_name"
	sessionCSRFToken   = "csrf_token"
	sessionID          = "session_id"
	defaultSessionAge  = 12 * 60 * 60
	rememberedAge      = 30 * 24 * 60 * 60
)

type SessionUser struct {
	ID          string
	Name        string
	StoreID     string
	UserStoreID string
	StoreName   string
	RoleName    string
}

func CurrentUser(c *gin.Context) (SessionUser, bool) {
	session := sessions.Default(c)
	userID, ok := session.Get(sessionUserID).(string)
	if !ok || userID == "" {
		return SessionUser{}, false
	}
	return SessionUser{
		ID:          userID,
		Name:        stringValue(session.Get(sessionUserName)),
		StoreID:     stringValue(session.Get(sessionStoreID)),
		UserStoreID: stringValue(session.Get(sessionUserStoreID)),
		StoreName:   stringValue(session.Get(sessionStoreName)),
		RoleName:    stringValue(session.Get(sessionRoleName)),
	}, true
}

func RequireLogin(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUser(c); !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		if db != nil {
			id := CurrentSessionID(c)
			var active bool
			if id == "" || db.QueryRowContext(c.Request.Context(), `SELECT EXISTS(SELECT 1 FROM user_sessions WHERE id=? AND revoked_at IS NULL AND expires_at>CURRENT_TIMESTAMP(6))`, id).Scan(&active) != nil || !active {
				session := sessions.Default(c)
				session.Clear()
				_ = session.Save()
				c.Redirect(http.StatusSeeOther, "/login")
				c.Abort()
				return
			}
			_, _ = db.ExecContext(c.Request.Context(), `UPDATE user_sessions SET last_seen_at=CURRENT_TIMESTAMP(6) WHERE id=? AND last_seen_at<CURRENT_TIMESTAMP(6)-INTERVAL 1 MINUTE`, id)
		}
		c.Next()
	}
}

func RequireCSRF() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if !ValidCSRF(c, c.PostForm("csrf_token")) {
				c.Redirect(http.StatusSeeOther, "/errors/419")
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func RequireRole(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		for _, role := range strings.Split(user.RoleName, ",") {
			if strings.TrimSpace(role) == required {
				c.Next()
				return
			}
		}
		c.Redirect(http.StatusSeeOther, "/errors/403")
		c.Abort()
	}
}

func RequirePermission(db *sql.DB, permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.Redirect(http.StatusSeeOther, "/login")
			c.Abort()
			return
		}
		for _, role := range strings.Split(user.RoleName, ",") {
			if strings.TrimSpace(role) == "superadmin" {
				c.Next()
				return
			}
		}
		userStoreID, err := strconv.ParseUint(user.UserStoreID, 10, 64)
		if err != nil {
			c.Redirect(http.StatusSeeOther, "/errors/403")
			c.Abort()
			return
		}
		const query = `SELECT EXISTS(
			SELECT 1 FROM permissions p
			WHERE p.name=? AND p.guard_name='web' AND (
				EXISTS(SELECT 1 FROM user_store_permissions usp WHERE usp.user_store_id=? AND usp.permission_id=p.id)
				OR EXISTS(SELECT 1 FROM user_store_roles usr JOIN role_permissions rp ON rp.role_id=usr.role_id WHERE usr.user_store_id=? AND rp.permission_id=p.id)
			)
		)`
		var allowed bool
		if err := db.QueryRowContext(c.Request.Context(), query, permission, userStoreID, userStoreID).Scan(&allowed); err != nil || !allowed {
			c.Redirect(http.StatusSeeOther, "/errors/403")
			c.Abort()
			return
		}
		c.Next()
	}
}

func CSRFToken(c *gin.Context) (string, error) {
	session := sessions.Default(c)
	if token, ok := session.Get(sessionCSRFToken).(string); ok && token != "" {
		return token, nil
	}
	token, err := newToken()
	if err != nil {
		return "", err
	}
	session.Set(sessionCSRFToken, token)
	if err := session.Save(); err != nil {
		return "", fmt.Errorf("save csrf session: %w", err)
	}
	return token, nil
}

func ValidCSRF(c *gin.Context, submitted string) bool {
	expected, ok := sessions.Default(c).Get(sessionCSRFToken).(string)
	if !ok || expected == "" || submitted == "" || len(expected) != len(submitted) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(submitted)) == 1
}

func SaveLogin(c *gin.Context, user User, id string, remember, secure bool) error {
	session := sessions.Default(c)
	session.Clear()
	maxAge := defaultSessionAge
	if remember {
		maxAge = rememberedAge
	}
	session.Options(cookieOptions(maxAge, secure))
	session.Set(sessionUserID, strconv.FormatUint(user.ID, 10))
	session.Set(sessionUserName, user.Name)
	session.Set(sessionStoreID, strconv.FormatUint(user.StoreID, 10))
	session.Set(sessionUserStoreID, strconv.FormatUint(user.UserStoreID, 10))
	session.Set(sessionStoreName, user.StoreName)
	session.Set(sessionRoleName, user.RoleName)
	session.Set(sessionID, id)
	token, err := newToken()
	if err != nil {
		return err
	}
	session.Set(sessionCSRFToken, token)
	return session.Save()
}

func CurrentSessionID(c *gin.Context) string { return stringValue(sessions.Default(c).Get(sessionID)) }
func NewSessionID() (string, error)          { return newToken() }
func SessionLifetime(remember bool) time.Duration {
	if remember {
		return rememberedAge * time.Second
	}
	return defaultSessionAge * time.Second
}

func ClearLogin(c *gin.Context, secure bool) error {
	session := sessions.Default(c)
	session.Clear()
	session.Options(cookieOptions(-1, secure))
	return session.Save()
}

func UpdateSessionName(c *gin.Context, name string) error {
	session := sessions.Default(c)
	session.Set(sessionUserName, name)
	return session.Save()
}

func cookieOptions(maxAge int, secure bool) sessions.Options {
	return sessions.Options{
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
	}
}

func newToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate csrf token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
