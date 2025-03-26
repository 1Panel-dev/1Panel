package security

import (
	"encoding/base64"
	"fmt"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/app/service"
	"github.com/1Panel-dev/1Panel/core/cmd/server/res"
	"github.com/1Panel-dev/1Panel/core/cmd/server/web"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func HandleNotRoute(c *gin.Context) bool {
	if !checkBindDomain(c) {
		HandleNotSecurity(c, "err_domain")
		return false
	}
	if !checkIPLimit(c) {
		HandleNotSecurity(c, "err_ip_limit")
		return false
	}
	if checkFrontendPath(c) {
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
	if !checkEntrance(c) && !checkSession(c) {
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
	_, _ = c.Writer.Write(web.IndexByte)
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
	resPage, err := service.NewIAuthService().GetResponsePage()
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if resPage == "444" {
		c.String(444, "")
		return
	}

	file := fmt.Sprintf("html/%s.html", resPage)
	if resPage == "200" && resType != "" {
		file = fmt.Sprintf("html/200_%s.html", resType)
	}
	data, err := res.ErrorMsg.ReadFile(file)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	statusCode, err := strconv.Atoi(resPage)
	if err != nil {
		c.String(http.StatusInternalServerError, "Internal Server Error")
		return
	}
	c.Data(statusCode, "text/html; charset=utf-8", data)
}

func isFrontendPath(c *gin.Context) bool {
	reqUri := strings.TrimSuffix(c.Request.URL.Path, "/")
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

func checkFrontendPath(c *gin.Context) bool {
	if !isFrontendPath(c) {
		return false
	}
	authService := service.NewIAuthService()
	if authService.GetSecurityEntrance() != "" {
		return authService.IsLogin(c)
	}
	return true
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
	clientIP := c.ClientIP()
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
