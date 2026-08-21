package common

import (
	"fmt"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// ResolveClientIP returns the TCP peer address unless that peer is trusted.
// Forwarded client address headers are never used for untrusted peers.
func ResolveClientIP(c *gin.Context, trustedProxies string) string {
	remoteAddr := GetRealClientIP(c)
	remoteIP := net.ParseIP(remoteAddr)
	if remoteIP == nil {
		return remoteAddr
	}

	proxies, err := parseTrustedProxies(trustedProxies)
	if err != nil || !isIPInNetworks(remoteIP, proxies) {
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

func NormalizeTrustedProxies(value string) (string, error) {
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
			return "", fmt.Errorf("invalid trusted proxy entry %q: %w", item, err)
		}
		ones, _ := ipNet.Mask.Size()
		if ones == 0 {
			return "", fmt.Errorf("invalid trusted proxy entry %q: unrestricted CIDR is not allowed", item)
		}
		normalized = append(normalized, ipNet.String())
	}
	return strings.Join(normalized, "\n"), nil
}

func parseTrustedProxies(value string) ([]*net.IPNet, error) {
	normalized, err := NormalizeTrustedProxies(value)
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
