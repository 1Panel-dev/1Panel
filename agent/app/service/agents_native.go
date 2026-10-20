package service

import (
	"net"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

var deepSeekUsernamePattern = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
var deepSeekHostLabelPattern = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?$`)

func isDeepSeekAccessHost(value string) bool {
	if ip := net.ParseIP(value); ip != nil {
		return ip.To4() != nil && !strings.Contains(value, ":")
	}
	if value == "" || len(value) > 253 || strings.Trim(value, "0123456789.") == "" {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if !deepSeekHostLabelPattern.MatchString(label) {
			return false
		}
	}
	return true
}

func validateNativeAgentCreate(req dto.AgentCreateReq) error {
	if req.AgentType != constant.AppDeepSeekHarness && req.AgentType != constant.AppOctop {
		return nil
	}
	if strings.TrimSpace(req.DashboardUsername) == "" || strings.ContainsAny(req.DashboardUsername, "\r\n\x00") {
		return buserr.WithDetail("ErrInvalidParams", "dashboardUsername", nil)
	}
	if strings.ContainsAny(req.DashboardPassword, "\r\n\x00") || req.DashboardPassword != strings.TrimSpace(req.DashboardPassword) {
		return buserr.WithDetail("ErrInvalidParams", "dashboardPassword", nil)
	}
	if req.AgentType == constant.AppDeepSeekHarness {
		if !isDeepSeekAccessHost(req.AccessHost) {
			return buserr.WithDetail("ErrInvalidParams", "accessHost", nil)
		}
		if !deepSeekUsernamePattern.MatchString(req.DashboardUsername) {
			return buserr.WithDetail("ErrInvalidParams", "dashboardUsername", nil)
		}
		if utf8.RuneCountInString(req.DashboardPassword) < 12 {
			return buserr.WithDetail("ErrInvalidParams", "dashboardPassword", nil)
		}
		return nil
	}
	hasLetter, hasDigit := false, false
	for _, char := range req.DashboardPassword {
		hasLetter = hasLetter || unicode.IsLetter(char)
		hasDigit = hasDigit || unicode.IsDigit(char)
	}
	if utf8.RuneCountInString(req.DashboardPassword) < 8 || !hasLetter || !hasDigit {
		return buserr.WithDetail("ErrInvalidParams", "dashboardPassword", nil)
	}
	switch strings.ToLower(req.DashboardPassword) {
	case "password", "password1", "password12", "password123", "12345678", "123456789", "qwerty123",
		"admin123", "welcome1", "letmein1", "changeme1", "octop123", "abc12345", "iloveyou1":
		return buserr.WithDetail("ErrInvalidParams", "dashboardPassword", nil)
	}
	return nil
}
