package security

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/cmd/server/res"
	"github.com/1Panel-dev/1Panel/core/cmd/server/web"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

var publicSharePagePattern = regexp.MustCompile(`^/s/[A-Za-z0-9]{10,16}$`)

func HandleNotRoute(c *gin.Context) bool {
	if !checkBindDomain(c) {
		HandleNotSecurity(c, "err_domain")
		return false
	}
	if !checkIPLimit(c) {
		HandleNotSecurity(c, "err_ip_limit")
		return false
	}
	if CanServeFrontendPath(c) {
		ToIndexHtml(c)
		return false
	}
	if isEntrancePath(c) {
		ToIndexHtml(c)
		return false
	}
	return true
}

func CheckSecurity(c *gin.Context) bool {
	authService := service.NewIAuthService()
	entrance := authService.GetSecurityEntrance()
	if entrance != "" && !checkEntrance(c) && !checkSession(c) {
		HandleNotSecurity(c, "")
		return false
	}
	if !checkBindDomain(c) {
		HandleNotSecurity(c, "err_domain")
		return false
	}
	if !checkIPLimit(c) {
		HandleNotSecurity(c, "err_ip_limit")
		return false
	}
	return true
}

func ToIndexHtml(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(http.StatusOK)
	data, err := web.IndexHtml.ReadFile("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "index.html not found")
		return
	}
	_, _ = c.Writer.Write(data)
	c.Writer.Flush()
}

func isEntrancePath(c *gin.Context) bool {
	entrance := service.NewIAuthService().GetSecurityEntrance()
	if entrance != "" && strings.TrimSuffix(c.Request.URL.Path, "/") == "/"+entrance {
		return true
	}
	return false
}

func checkEntrance(c *gin.Context) bool {
	authService := service.NewIAuthService()
	entrance := authService.GetSecurityEntrance()
	if entrance == "" {
		return true
	}

	cookieValue, err := c.Cookie("SecurityEntrance")
	if err != nil {
		return false
	}
	entranceValue, err := base64.StdEncoding.DecodeString(cookieValue)
	if err != nil {
		return false
	}
	return string(entranceValue) == entrance
}

func HandleNotSecurity(c *gin.Context, resType string) {
	code := LoadErrCode()
	if code == 444 {
		CloseDirectly(c)
		return
	}

	file := fmt.Sprintf("html/%d.html", code)
	if code == http.StatusOK && resType != "" {
		file = fmt.Sprintf("html/200_%s.html", resType)
	}
	data, err := res.ErrorMsg.ReadFile(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	c.Data(code, "text/html; charset=utf-8", data)
}

func LoadErrCode() int {
	settingRepo := repo.NewISettingRepo()
	codeVal, err := settingRepo.GetValueByKey("NoAuthSetting")
	if err != nil {
		return http.StatusInternalServerError
	}

	switch codeVal {
	case "400":
		return http.StatusBadRequest
	case "401":
		return http.StatusUnauthorized
	case "403":
		return http.StatusForbidden
	case "404":
		return http.StatusNotFound
	case "408":
		return http.StatusRequestTimeout
	case "416":
		return http.StatusRequestedRangeNotSatisfiable
	case "500":
		return http.StatusInternalServerError
	case "444":
		return 444
	default:
		return http.StatusOK
	}
}

func IsFrontendPath(path string) bool {
	reqUri := strings.TrimSuffix(path, "/")
	if _, ok := constant.WebUrlMap[reqUri]; ok {
		return true
	}
	for _, route := range constant.DynamicRoutes {
		if match, _ := regexp.MatchString(route, reqUri); match {
			return true
		}
	}
	return false
}

func CanServeFrontendPath(c *gin.Context) bool {
	if !IsFrontendPath(c.Request.URL.Path) {
		return false
	}
	if isPublicFileSharePagePath(c.Request.URL.Path) {
		return true
	}
	authService := service.NewIAuthService()
	if authService.GetSecurityEntrance() != "" {
		return checkEntrance(c) || authService.IsLogin(c)
	}
	return true
}

func isPublicFileSharePagePath(path string) bool {
	reqUri := strings.TrimSuffix(path, "/")
	return publicSharePagePattern.MatchString(reqUri)
}

func checkBindDomain(c *gin.Context) bool {
	settingRepo := repo.NewISettingRepo()
	status, _ := settingRepo.Get(repo.WithByKey("BindDomain"))
	if len(status.Value) == 0 {
		return true
	}
	domains := c.Request.Host
	parts := strings.Split(c.Request.Host, ":")
	if len(parts) > 0 {
		domains = parts[0]
	}
	return domains == status.Value
}

func checkIPLimit(c *gin.Context) bool {
	settingRepo := repo.NewISettingRepo()
	status, _ := settingRepo.Get(repo.WithByKey("AllowIPs"))
	if len(status.Value) == 0 {
		return true
	}
	trustedProxies, _ := settingRepo.Get(repo.WithByKey("AllowIPTrustedProxies"))
	clientIP := common.ResolveClientIP(c, trustedProxies.Value)

	for _, ip := range strings.Split(status.Value, ",") {
		if len(ip) == 0 {
			continue
		}
		if ip == clientIP || (strings.Contains(ip, "/") && common.CheckIpInCidr(ip, clientIP)) {
			return true
		}
	}
	return false
}

func checkSession(c *gin.Context) bool {
	_, err := global.SESSION.Get(c)
	return err == nil
}

func CloseDirectly(c *gin.Context) {
	hijacker, ok := c.Writer.(http.Hijacker)
	if !ok {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	conn, _, err := hijacker.Hijack()
	if err != nil {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	conn.Close()
	c.Abort()
}
