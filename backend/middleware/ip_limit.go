package middleware

import (
	"net"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
)

func WhiteAllow() gin.HandlerFunc {
	return func(c *gin.Context) {
		settingRepo := repo.NewISettingRepo()
		status, err := settingRepo.Get(settingRepo.WithByKey("AllowIPs"))
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrInternalServer, constant.ErrTypeInternalServer, err)
			return
		}

		if len(status.Value) == 0 {
			c.Next()
			return
		}
		clientIP := c.ClientIP()
		for _, ip := range strings.Split(status.Value, ",") {
			if len(ip) == 0 {
				continue
			}
			if ip == clientIP || (strings.Contains(ip, "/") && checkIpInCidr(ip, clientIP)) {
				c.Next()
				return
			}
		}
		code := LoadErrCode()
		helper.ErrWithHtml(c, code, "err_ip_limit")
	}
}

func checkIpInCidr(cidr, checkIP string) bool {
	ip, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		global.LOG.Errorf("parse CIDR %s failed, err: %v", cidr, err)
		return false
	}
	for ip := ip.Mask(ipNet.Mask); ipNet.Contains(ip); incIP(ip) {
		if ip.String() == checkIP {
			return true
		}
	}
	return false
}

func incIP(ip net.IP) {
	for j := len(ip) - 1; j >= 0; j-- {
		ip[j]++
		if ip[j] > 0 {
			break
		}
	}
}
