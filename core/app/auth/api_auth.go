package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/core/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/core/app/repo"
	"github.com/1Panel-dev/1Panel/core/buserr"
	"github.com/1Panel-dev/1Panel/core/constant"
	"github.com/1Panel-dev/1Panel/core/global"
	"github.com/1Panel-dev/1Panel/core/utils/common"
	"github.com/gin-gonic/gin"
)

type APIAuthConfig struct {
	ApiInterfaceStatus string
	ApiKey             string
	IpWhiteList        string
	ApiTrustedProxies  string
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
			var bizErr buserr.BusinessError
			if errors.As(err, &bizErr) && strings.HasPrefix(bizErr.Msg, "ErrApiConfig") {
				helper.BadAuth(c, bizErr.Msg, bizErr.Err)
				return
			}
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
		if !isIPInWhiteList(GetAPIClientIP(c, config.ApiTrustedProxies), config.IpWhiteList) {
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
	if config.ApiTrustedProxies, err = settingRepo.GetValueByKey("ApiTrustedProxies"); err != nil {
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

func GetAPIClientIP(c *gin.Context, trustedProxies string) string {
	remoteAddr := common.GetRealClientIP(c)
	remoteIP := net.ParseIP(remoteAddr)
	if remoteIP == nil {
		return remoteAddr
	}

	proxies, err := parseAPITrustedProxies(trustedProxies)
	if err != nil {
		if global.LOG != nil {
			global.LOG.Errorf("Failed to parse API trusted proxies: %v", err)
		}
		return remoteAddr
	}
	if !isIPInNetworks(remoteIP, proxies) {
		return remoteAddr
	}

	forwardedFor := strings.Join(c.Request.Header.Values("X-Forwarded-For"), ",")
	if strings.TrimSpace(forwardedFor) != "" {
		clientIP, ok := clientIPFromForwardedFor(forwardedFor, proxies)
		if !ok {
			return remoteAddr
		}
		return clientIP
	}

	realIPValue := strings.Join(c.Request.Header.Values("X-Real-IP"), ",")
	realIP := net.ParseIP(strings.TrimSpace(realIPValue))
	if realIP == nil {
		return remoteAddr
	}
	return realIP.String()
}

func NormalizeAPITrustedProxies(value string) (string, error) {
	lines := strings.Split(value, "\n")
	normalized := make([]string, 0, len(lines))
	for _, line := range lines {
		item := strings.TrimSpace(line)
		if item == "" {
			continue
		}
		if ip := net.ParseIP(item); ip != nil {
			normalized = append(normalized, ip.String())
			continue
		}
		_, ipNet, err := net.ParseCIDR(item)
		if err != nil {
			return "", fmt.Errorf("invalid API trusted proxy entry %q: %w", item, err)
		}
		ones, _ := ipNet.Mask.Size()
		if ones == 0 {
			return "", fmt.Errorf("invalid API trusted proxy entry %q: unrestricted CIDR is not allowed", item)
		}
		normalized = append(normalized, ipNet.String())
	}
	return strings.Join(normalized, "\n"), nil
}

func parseAPITrustedProxies(value string) ([]*net.IPNet, error) {
	normalized, err := NormalizeAPITrustedProxies(value)
	if err != nil {
		return nil, err
	}
	if normalized == "" {
		return []*net.IPNet{}, nil
	}

	lines := strings.Split(normalized, "\n")
	proxies := make([]*net.IPNet, 0, len(lines))
	for _, item := range lines {
		if ip := net.ParseIP(item); ip != nil {
			bits := 128
			if ip.To4() != nil {
				bits = 32
				ip = ip.To4()
			}
			proxies = append(proxies, &net.IPNet{IP: ip, Mask: net.CIDRMask(bits, bits)})
			continue
		}
		_, ipNet, err := net.ParseCIDR(item)
		if err != nil {
			return nil, err
		}
		proxies = append(proxies, ipNet)
	}
	return proxies, nil
}

func clientIPFromForwardedFor(value string, trustedProxies []*net.IPNet) (string, bool) {
	items := strings.Split(value, ",")
	ips := make([]net.IP, len(items))
	for i, item := range items {
		ip := net.ParseIP(strings.TrimSpace(item))
		if ip == nil {
			return "", false
		}
		ips[i] = ip
	}
	for i := len(ips) - 1; i >= 0; i-- {
		if i == 0 || !isIPInNetworks(ips[i], trustedProxies) {
			return ips[i].String(), true
		}
	}
	return "", false
}

func isIPInNetworks(ip net.IP, networks []*net.IPNet) bool {
	for _, network := range networks {
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

func IsValid1PanelTimestamp(panelTimestamp string, apiKeyValidityTime int) bool {
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

func IsValid1PanelToken(panelToken string, panelTimestamp string, apiKey string) bool {
	return IsValid1PanelTokenWithVersion(panelToken, panelTimestamp, apiKey, "")
}

func IsValid1PanelTokenWithVersion(panelToken string, panelTimestamp string, apiKey string, signatureVersion string) bool {
	panelToken = strings.ToLower(strings.TrimSpace(panelToken))
	version := strings.ToLower(strings.TrimSpace(signatureVersion))
	switch version {
	case "v1", "md5":
		return isValidMD5Token(panelToken, panelTimestamp, apiKey)
	case "hmac-sha256":
		return isValidHMACSHA256Token(panelToken, panelTimestamp, apiKey)
	default:
		return isValidMD5Token(panelToken, panelTimestamp, apiKey) || isValidHMACSHA256Token(panelToken, panelTimestamp, apiKey)
	}
}

func isValidMD5Token(panelToken string, panelTimestamp string, apiKey string) bool {
	return panelToken == GenerateMD5("1panel"+apiKey+panelTimestamp)
}

func isValidHMACSHA256Token(panelToken string, panelTimestamp string, apiKey string) bool {
	return panelToken == GenerateHMACSHA256(apiKey, "1panel:"+panelTimestamp)
}

func IsIPInWhiteList(clientIP string, ipWhiteString string) bool {
	if strings.TrimSpace(ipWhiteString) == "" {
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
		cidr = strings.TrimSpace(cidr)
		if cidr == "" {
			continue
		}
		if (iPv4 != nil && (cidr == "0.0.0.0" || cidr == "0.0.0.0/0" || iPv4.String() == cidr)) || (iPv6 != nil && (cidr == "::/0" || iPv6.String() == cidr)) {
			return true
		}
		whiteIP := net.ParseIP(cidr)
		if whiteIP != nil && whiteIP.Equal(clientParsedIP) {
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

func isValid1PanelTimestamp(panelTimestamp string, apiKeyValidityTime int) bool {
	return IsValid1PanelTimestamp(panelTimestamp, apiKeyValidityTime)
}

func isValid1PanelToken(panelToken string, panelTimestamp string, apiKey string) bool {
	return IsValid1PanelToken(panelToken, panelTimestamp, apiKey)
}

func isIPInWhiteList(clientIP string, ipWhiteString string) bool {
	return IsIPInWhiteList(clientIP, ipWhiteString)
}

func GenerateMD5(param string) string {
	hash := md5.New()
	hash.Write([]byte(param))
	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateHMACSHA256(secret string, message string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	return hex.EncodeToString(mac.Sum(nil))
}
