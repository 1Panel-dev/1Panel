package terminal_session

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"time"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/gin-gonic/gin"
)

const (
	HeaderUserID         = "X-Panel-User-ID"
	HeaderAuthSessionID  = "X-Panel-Auth-Session-ID"
	HeaderAuthLeaseUntil = "X-Panel-Auth-Lease-Until"
)

type Identity struct {
	UserID         string
	AuthSessionID  string
	AuthLeaseUntil time.Time
}

func FromContext(c *gin.Context) (Identity, bool) {
	if c == nil {
		return Identity{}, false
	}
	value, ok := c.Get(psession.GinContextSessionUserKey)
	if !ok {
		return Identity{}, false
	}
	user, ok := value.(psession.SessionUser)
	if !ok || user.ID == "" {
		return Identity{}, false
	}
	if c.GetBool("API_AUTH") {
		identity := Identity{UserID: user.ID, AuthSessionID: APIAuthSessionID(user.ID), AuthLeaseUntil: time.Now().Add(90 * time.Second)}
		if keyID := c.GetString("API_AUTH_KEY_ID"); keyID != "" && c.GetString("API_AUTH_KEY_KIND") == "apiKey" {
			identity.AuthSessionID = APIKeyAuthSessionID(keyID)
			if value, exists := c.Get("API_AUTH_KEY_EXPIRES_AT"); exists {
				if expiresAt, ok := value.(time.Time); ok && !expiresAt.IsZero() && expiresAt.Before(identity.AuthLeaseUntil) {
					identity.AuthLeaseUntil = expiresAt
				}
			}
		}
		return identity, true
	}
	sessionID, err := c.Cookie(constant.SessionName)
	if err != nil || sessionID == "" {
		return Identity{}, false
	}
	return Identity{UserID: user.ID, AuthSessionID: HashAuthSessionID(sessionID)}, true
}

func APIAuthSessionID(userID string) string {
	return "api:" + userID
}

func APIKeyAuthSessionID(keyID string) string {
	return "api-key:" + keyID
}

func HashAuthSessionID(sessionID string) string {
	sum := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(sum[:])
}

func ClearForwardedHeaders(c *gin.Context) {
	c.Request.Header.Del(HeaderUserID)
	c.Request.Header.Del(HeaderAuthSessionID)
	c.Request.Header.Del(HeaderAuthLeaseUntil)
}

func SetForwardedHeaders(c *gin.Context, identity Identity) {
	c.Request.Header.Set(HeaderUserID, identity.UserID)
	c.Request.Header.Set(HeaderAuthSessionID, identity.AuthSessionID)
	c.Request.Header.Del(HeaderAuthLeaseUntil)
	if !identity.AuthLeaseUntil.IsZero() {
		c.Request.Header.Set(HeaderAuthLeaseUntil, strconv.FormatInt(identity.AuthLeaseUntil.UnixMilli(), 10))
	}
}
