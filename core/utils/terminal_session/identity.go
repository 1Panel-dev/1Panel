package terminal_session

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/gin-gonic/gin"
)

const (
	HeaderUserID        = "X-Panel-User-ID"
	HeaderAuthSessionID = "X-Panel-Auth-Session-ID"
)

type Identity struct {
	UserID        string
	AuthSessionID string
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
		return Identity{UserID: user.ID, AuthSessionID: APIAuthSessionID(user.ID)}, true
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

func HashAuthSessionID(sessionID string) string {
	sum := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(sum[:])
}

func ClearForwardedHeaders(c *gin.Context) {
	c.Request.Header.Del(HeaderUserID)
	c.Request.Header.Del(HeaderAuthSessionID)
}

func SetForwardedHeaders(c *gin.Context, identity Identity) {
	c.Request.Header.Set(HeaderUserID, identity.UserID)
	c.Request.Header.Set(HeaderAuthSessionID, identity.AuthSessionID)
}
