package middleware

import (
	"net/http"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/auth"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/global"
	psessionUtils "github.com/1Panel-dev/1Panel/core/init/session/psession"
	"github.com/1Panel-dev/1Panel/core/utils/xpack"
	"github.com/gin-gonic/gin"
)

func PasswordExpired() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !strings.HasPrefix(c.Request.URL.Path, "/api/v2/") {
			c.Next()
			return
		}
		if c.GetBool("LOCAL_REQUEST") {
			c.Next()
			return
		}
		if IsPublicFileShareAPI(c.Request.URL.Path) {
			c.Next()
			return
		}
		if passwordExpirationRecoveryRoute(c.Request.Method, c.Request.URL.Path) {
			c.Next()
			return
		}
		if c.GetBool("API_AUTH") {
			c.Next()
			return
		}
		var err error
		sessionUser, ok := c.Get(psessionUtils.GinContextSessionUserKey)
		psession, typeOK := sessionUser.(psessionUtils.SessionUser)
		if !ok || !typeOK {
			psession, err = global.SESSION.Get(c)
		}
		if err != nil {
			errItem := err.Error()
			if errItem == "ErrSessionDataFormat" || errItem == "ErrSessionDataNotFound" {
				helper.BadAuth(c, "ErrNotLogin", buserr.New(errItem))
				return
			}
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		c.Set(psessionUtils.GinContextSessionUserKey, psession)
		if len(psession.Name) == 0 {
			helper.BadAuth(c, "ErrNotLogin", err)
			return
		}
		expired, err := auth.IsPasswordExpired(c, xpack.AuthProvider)
		if err != nil {
			helper.ErrorWithDetail(c, http.StatusInternalServerError, "ErrPasswordExpired", err)
			return
		}

		if expired {
			helper.ErrorWithDetail(c, 313, "ErrPasswordExpired", err)
			return
		}
		c.Next()
	}
}

func passwordExpirationRecoveryRoute(method, path string) bool {
	const core = "/api/v2/core"
	switch method {
	case http.MethodGet:
		switch path {
		case core + "/auth/captcha", core + "/auth/setting", core + "/auth/current", core + "/auth/ldap/status",
			core + "/auth/oidc/status", core + "/auth/oidc/callback",
			core + "/auth/saml2/status", core + "/auth/saml2/metadata", core + "/auth/saml2/logout",
			core + "/xpack/settings/search", core + "/enterprise/users/info",
			core + "/enterprise/licenses/info", core + "/enterprise/licenses/status":
			return true
		}
	case http.MethodPost:
		switch path {
		case core + "/auth/login", core + "/auth/mfalogin", core + "/auth/logout",
			core + "/auth/passkey/begin", core + "/auth/passkey/finish",
			core + "/auth/current/update", core + "/auth/expired/reset",
			core + "/auth/oidc/begin", core + "/auth/oidc/finish",
			core + "/auth/saml2/begin", core + "/auth/saml2/acs", core + "/auth/saml2/finish", core + "/auth/saml2/logout",
			core + "/settings/search/base", core + "/enterprise/licenses/upload", core + "/xapp/verifyQRCode",
			core + "/auth/api/keys/search", core + "/auth/api/keys/status",
			core + "/auth/api/keys/revoke", core + "/auth/api/keys/delete":
			// The status service independently rejects Enable while allowing
			// Disable. Metadata and credential revocation remain available.
			return true
		}
	}
	return false
}
