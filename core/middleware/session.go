package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/gin-gonic/gin"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") {
			c.Next()
			return
		}
		if method, exist := c.Get("authMethod"); exist && method == constant.AuthMethodJWT {
			c.Next()
			return
		}

		panelToken := c.GetHeader("1Panel-Token")
		panelTimestamp := c.GetHeader("1Panel-Timestamp")
		if panelToken != "" || panelTimestamp != "" {
			if global.CONF.System.ApiInterfaceStatus == constant.StatusEnable {
				clientIP := c.ClientIP()
				if !isValid1PanelTimestamp(panelTimestamp) {
					helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrApiConfigKeyTimeInvalid, nil)
					return
				}
				if !isValid1PanelToken(panelToken, panelTimestamp) {
					helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrApiConfigKeyInvalid, nil)
					return
				}

				if !isIPInWhiteList(clientIP) {
					helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrApiConfigIPInvalid, nil)
					return
				}
				c.Next()
				return
			} else {
				helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrApiConfigStatusInvalid, nil)
				return
			}
		}

		psession, err := global.SESSION.Get(c)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeNotLogin, err)
			return
		}
		settingRepo := repo.NewISettingRepo()
		setting, err := settingRepo.Get(repo.WithByKey("SessionTimeout"))
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		lifeTime, _ := strconv.Atoi(setting.Value)
		httpsSetting, err := settingRepo.Get(repo.WithByKey("SSL"))
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
			return
		}
		_ = global.SESSION.Set(c, psession, httpsSetting.Value == constant.StatusEnable, lifeTime)
		c.Next()
	}
}

func isValid1PanelTimestamp(panelTimestamp string) bool {
	apiKeyValidityTime := global.CONF.System.ApiKeyValidityTime
	apiTime, err := strconv.Atoi(apiKeyValidityTime)
	if err != nil {
		return false
	}
	panelTime, err := strconv.ParseInt(panelTimestamp, 10, 64)
	if err != nil {
		return false
	}
	nowTime := time.Now().Unix()
	if panelTime > nowTime {
		return false
	}
	return apiTime == 0 || nowTime-panelTime <= int64(apiTime*60)
}

func isValid1PanelToken(panelToken string, panelTimestamp string) bool {
	system1PanelToken := global.CONF.System.ApiKey
	return panelToken == GenerateMD5("1panel"+system1PanelToken+panelTimestamp)
}

func isIPInWhiteList(clientIP string) bool {
	ipWhiteString := global.CONF.System.IpWhiteList
	ipWhiteList := strings.Split(ipWhiteString, "\n")
	for _, cidr := range ipWhiteList {
		if cidr == "0.0.0.0" {
			return true
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			if cidr == clientIP {
				return true
			}
			continue
		}
		if ipNet.Contains(net.ParseIP(clientIP)) {
			return true
		}
	}
	return false
}

func GenerateMD5(input string) string {
	hash := md5.New()
	hash.Write([]byte(input))
	return hex.EncodeToString(hash.Sum(nil))
}
