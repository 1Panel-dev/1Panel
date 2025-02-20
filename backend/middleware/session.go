package middleware

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/gin-gonic/gin"
	"net"
	"strconv"
	"strings"
	"time"
)

func SessionAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if method, exist := c.Get("authMethod"); exist && method == constant.AuthMethodJWT {
			c.Next()
			return
		}
		panelToken := c.GetHeader("1Panel-Token")
		panelTimestamp := c.GetHeader("1Panel-Timestamp")
		if panelToken != "" || panelTimestamp != "" {
			if global.CONF.System.ApiInterfaceStatus == "enable" {
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

		sId, err := c.Cookie(constant.SessionName)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeNotLogin, nil)
			return
		}
		psession, err := global.SESSION.Get(sId)
		if err != nil {
			helper.ErrorWithDetail(c, constant.CodeErrUnauthorized, constant.ErrTypeNotLogin, nil)
			return
		}
		settingRepo := repo.NewISettingRepo()
		setting, err := settingRepo.Get(settingRepo.WithByKey("SessionTimeout"))
		if err != nil {
			global.LOG.Errorf("create operation record failed, err: %v", err)
		}
		lifeTime, _ := strconv.Atoi(setting.Value)
		_ = global.SESSION.Set(sId, psession, lifeTime)
		c.Next()
	}
}

func isValid1PanelTimestamp(panelTimestamp string) bool {
	apiKeyValidityTime := global.CONF.System.ApiKeyValidityTime
	apiTime, err := strconv.Atoi(apiKeyValidityTime)
	if err != nil || apiTime < 0 {
		global.LOG.Errorf("apiTime %d, err: %v", apiTime, err)
		return false
	}
	if apiTime == 0 {
		return true
	}
	panelTime, err := strconv.ParseInt(panelTimestamp, 10, 64)
	if err != nil {
		global.LOG.Errorf("panelTimestamp %s, panelTime %d, apiTime %d, err: %v", panelTimestamp, apiTime, panelTime, err)
		return false
	}
	nowTime := time.Now().Unix()
	tolerance := int64(60)
	if panelTime > nowTime+tolerance {
		global.LOG.Errorf("Valid Panel Timestamp, apiTime %d, panelTime %d, nowTime %d, err: %v", apiTime, panelTime, nowTime, err)
		return false
	}
	return nowTime-panelTime <= int64(apiTime)*60+tolerance
}

func isValid1PanelToken(panelToken string, panelTimestamp string) bool {
	system1PanelToken := global.CONF.System.ApiKey
	if panelToken == GenerateMD5("1panel"+system1PanelToken+panelTimestamp) {
		return true
	}
	return false
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
