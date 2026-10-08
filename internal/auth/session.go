package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	SessionName       = "mksiaga_session"
	sessionUserID     = "user_id"
	sessionUserName   = "user_name"
	sessionStoreID    = "store_id"
	sessionStoreName  = "store_name"
	sessionRoleName   = "role_name"
	sessionCSRFToken  = "csrf_token"
	defaultSessionAge = 12 * 60 * 60
	rememberedAge     = 30 * 24 * 60 * 60
)

type SessionUser struct {
	ID        string
	Name      string
	StoreID   string
	StoreName string
	RoleName  string
}

func CurrentUser(c *gin.Context) (SessionUser, bool) {
	session := sessions.Default(c)
	userID, ok := session.Get(sessionUserID).(string)
	if !ok || userID == "" {
		return SessionUser{}, false
	}
	return SessionUser{
		ID:        userID,
		Name:      stringValue(session.Get(sessionUserName)),
		StoreID:   stringValue(session.Get(sessionStoreID)),
		StoreName: stringValue(session.Get(sessionStoreName)),
		RoleName:  stringValue(session.Get(sessionRoleName)),
	}, true
}

func RequireLogin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := CurrentUser(c); !ok {
			c.Redirect(http.StatusSeeOther, "/login")
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

func SaveLogin(c *gin.Context, user User, remember, secure bool) error {
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
	session.Set(sessionStoreName, user.StoreName)
	session.Set(sessionRoleName, user.RoleName)
	token, err := newToken()
	if err != nil {
		return err
	}
	session.Set(sessionCSRFToken, token)
	return session.Save()
}

func ClearLogin(c *gin.Context, secure bool) error {
	session := sessions.Default(c)
	session.Clear()
	session.Options(cookieOptions(-1, secure))
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
