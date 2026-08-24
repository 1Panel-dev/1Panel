package clientip

import (
	"net"
	"net/http"
	"strings"
)

const (
	forwardedHeader     = "Forwarded"
	forwardedForHeader  = "X-Forwarded-For"
	realIPHeader        = "X-Real-IP"
	panelClientIPHeader = "X-Panel-Client-IP"
)

func FromRemoteAddr(remoteAddr string) string {
	host, _, err := net.SplitHostPort(strings.TrimSpace(remoteAddr))
	if err != nil {
		return ""
	}
	ip := net.ParseIP(strings.TrimSpace(host))
	if ip == nil {
		return ""
	}
	return ip.String()
}

func ReplaceForwardingHeaders(header http.Header, clientIP string) {
	if header == nil {
		return
	}
	RemoveForwardingHeaders(header)

	ip := net.ParseIP(strings.TrimSpace(clientIP))
	if ip == nil {
		return
	}
	normalized := ip.String()
	header.Set(forwardedForHeader, normalized)
	header.Set(realIPHeader, normalized)
}

func RemoveForwardingHeaders(header http.Header) {
	if header == nil {
		return
	}
	header.Del(forwardedHeader)
	header.Del(forwardedForHeader)
	header.Del(realIPHeader)
	header.Del(panelClientIPHeader)
}
