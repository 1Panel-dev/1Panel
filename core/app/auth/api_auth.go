package auth

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
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

type APIAuthConfig struct {
	ApiInterfaceStatus string
	ApiKey             string
	IpWhiteList        string
	ApiKeyValidityTime int
}

type APIAuthConfigLoader func(c *gin.Context) (APIAuthConfig, error)
type APIAuthSuccessHandler func(c *gin.Context, config APIAuthConfig)

func APIAuthMiddleware(loadConfig APIAuthConfigLoader, onSuccess APIAuthSuccessHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api/v2/core/auth") {
			c.Next()
			return
		}

		panelToken := c.GetHeader("1Panel-Token")
		panelTimestamp := c.GetHeader("1Panel-Timestamp")
		if panelToken == "" && panelTimestamp == "" {
			c.Next()
			return
		}

		config, err := loadConfig(c)
		if err != nil {
			helper.InternalServer(c, err)
			return
		}
		if config.ApiInterfaceStatus != constant.StatusEnable {
			helper.BadAuth(c, "ErrApiConfigStatusInvalid", nil)
			return
		}
		if !isValid1PanelTimestamp(panelTimestamp, config.ApiKeyValidityTime) {
			helper.BadAuth(c, "ErrApiConfigKeyTimeInvalid", nil)
			return
		}
		if !isValid1PanelToken(panelToken, panelTimestamp, config.ApiKey) {
			helper.BadAuth(c, "ErrApiConfigKeyInvalid", nil)
			return
		}
		if !isIPInWhiteList(c.ClientIP(), config.IpWhiteList) {
			helper.BadAuth(c, "ErrApiConfigIPInvalid", nil)
			return
		}

		c.Set("API_AUTH", true)
		if onSuccess != nil {
			onSuccess(c, config)
		}
		c.Next()
	}
}

func LoadAPIAuthConfig(_ *gin.Context) (APIAuthConfig, error) {
	settingRepo := repo.NewISettingRepo()
	config := APIAuthConfig{}
	var err error
	if config.ApiInterfaceStatus, err = settingRepo.GetValueByKey("ApiInterfaceStatus"); err != nil {
		return config, err
	}
	if config.ApiKey, err = settingRepo.GetValueByKey("ApiKey"); err != nil {
		return config, err
	}
	if config.IpWhiteList, err = settingRepo.GetValueByKey("IpWhiteList"); err != nil {
		return config, err
	}
	apiValidity, err := settingRepo.GetValueByKey("ApiKeyValidityTime")
	if err != nil {
		return config, err
	}
	if config.ApiKeyValidityTime, err = strconv.Atoi(apiValidity); err != nil {
		return config, err
	}
	return config, nil
}

func isValid1PanelTimestamp(panelTimestamp string, apiKeyValidityTime int) bool {
	apiTime := apiKeyValidityTime
	if apiTime < 0 {
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

func isValid1PanelToken(panelToken string, panelTimestamp string, apiKey string) bool {
	return panelToken == GenerateMD5("1panel"+apiKey+panelTimestamp)
}

func isIPInWhiteList(clientIP string, ipWhiteString string) bool {
	if len(ipWhiteString) == 0 {
		global.LOG.Error("IP whitelist is empty")
		return false
	}
	ipWhiteList, ipErr := common.HandleIPList(ipWhiteString)
	if ipErr != nil {
		global.LOG.Errorf("Failed to handle IP list: %v", ipErr)
		return false
	}
	clientParsedIP := net.ParseIP(clientIP)
	if clientParsedIP == nil {
		return false
	}
	iPv4 := clientParsedIP.To4()
	iPv6 := clientParsedIP.To16()
	for _, cidr := range ipWhiteList {
		if (iPv4 != nil && (cidr == "0.0.0.0" || cidr == "0.0.0.0/0" || iPv4.String() == cidr)) || (iPv6 != nil && (cidr == "::/0" || iPv6.String() == cidr)) {
			return true
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if (iPv4 != nil && ipNet.Contains(iPv4)) || (iPv6 != nil && ipNet.Contains(iPv6)) {
			return true
		}
	}
	return false
}

func GenerateMD5(param string) string {
	hash := md5.New()
	hash.Write([]byte(param))
	return hex.EncodeToString(hash.Sum(nil))
}
