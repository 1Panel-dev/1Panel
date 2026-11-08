package aihttp

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/webhook_sender"
	"golang.org/x/net/idna"
)

const MaxResponseBytes = 1024 * 1024
const MaxDiscoveredModels = 4096

var (
	ErrTarget           = errors.New("AI endpoint is invalid or blocked by outbound policy")
	ErrRequest          = errors.New("AI endpoint request failed")
	ErrResponse         = errors.New("AI endpoint returned an invalid response")
	ErrResponseTooLarge = errors.New("AI endpoint response exceeded the size limit")
)

type ipResolver interface {
	LookupIPAddr(context.Context, string) ([]net.IPAddr, error)
}

type guardedTransport struct {
	resolver  ipResolver
	transport *http.Transport
}

func NewClient(timeout time.Duration) *http.Client {
	return newClient(timeout, nil, nil)
}

func newClient(timeout time.Duration, resolver ipResolver, transport *http.Transport) *http.Client {
	return &http.Client{
		Timeout:   timeout,
		Transport: &guardedTransport{resolver: resolver, transport: transport},
		CheckRedirect: func(next *http.Request, via []*http.Request) error {
			if len(via) >= 3 || len(via) == 0 || !sameOrigin(via[0].URL, next.URL) {
				return ErrTarget
			}
			return ValidateURL(next.URL.String())
		},
	}
}

func (guard *guardedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if request == nil || request.URL == nil {
		return nil, ErrTarget
	}
	canonical, err := normalizeTargetURL(request.URL.String())
	if err != nil {
		return nil, ErrTarget
	}
	hostname := canonical.Hostname()
	var addresses []net.IP
	if literal := net.ParseIP(hostname); literal != nil {
		addresses = append(addresses, literal)
	} else {
		resolver := guard.resolver
		if resolver == nil {
			resolver = net.DefaultResolver
		}
		resolved, err := resolver.LookupIPAddr(request.Context(), hostname)
		if err != nil {
			return nil, ErrTarget
		}
		for _, item := range resolved {
			if item.Zone != "" || !allowedIP(item.IP) {
				return nil, ErrTarget
			}
			addresses = append(addresses, item.IP)
		}
	}
	if len(addresses) == 0 {
		return nil, ErrTarget
	}
	for _, address := range addresses {
		if !allowedIP(address) {
			return nil, ErrTarget
		}
	}
	normalized := request.Clone(request.Context())
	normalized.URL = canonical
	normalized.Host = canonical.Host
	response, err := webhook_sender.RoundTripPinned(normalized, addresses, guard.transport)
	if err != nil {
		return nil, ErrRequest
	}
	return response, nil
}

func ValidateURL(rawURL string) error {
	_, err := normalizeTargetURL(rawURL)
	return err
}

func normalizeTargetURL(rawURL string) (*url.URL, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" || parsed.User != nil || parsed.Fragment != "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, ErrTarget
	}
	hostname := strings.TrimSuffix(parsed.Hostname(), ".")
	if hostname == "" || strings.Contains(hostname, "%") {
		return nil, ErrTarget
	}
	if net.ParseIP(hostname) == nil {
		hostname, err = idna.Lookup.ToASCII(hostname)
		if err != nil || hostname == "" {
			return nil, ErrTarget
		}
	}
	hostname = strings.ToLower(hostname)
	if address := net.ParseIP(hostname); address != nil {
		if !allowedIP(address) {
			return nil, ErrTarget
		}
		hostname = address.String()
	}
	if port := parsed.Port(); port != "" {
		parsed.Host = net.JoinHostPort(hostname, port)
	} else if strings.Contains(hostname, ":") {
		parsed.Host = "[" + hostname + "]"
	} else {
		parsed.Host = hostname
	}
	return parsed, nil
}

var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("240.0.0.0/4"),
	netip.MustParsePrefix("fd00:ec2::/64"),
	netip.MustParsePrefix("64:ff9b:1::/48"),
	netip.MustParsePrefix("2001::/32"),
}

func allowedIP(ip net.IP) bool {
	address, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	address = address.Unmap()
	if address.IsUnspecified() || address.IsMulticast() || address.IsLinkLocalUnicast() || address.IsLinkLocalMulticast() ||
		address == netip.MustParseAddr("100.100.100.200") {
		return false
	}
	if netip.MustParsePrefix("::/96").Contains(address) && address != netip.IPv6Loopback() {
		return false
	}
	if netip.MustParsePrefix("64:ff9b::/96").Contains(address) || netip.MustParsePrefix("::ffff:0:0:0/96").Contains(address) {
		value := address.As16()
		return allowedIP(net.IPv4(value[12], value[13], value[14], value[15]))
	}
	if netip.MustParsePrefix("2002::/16").Contains(address) {
		value := address.As16()
		return allowedIP(net.IPv4(value[2], value[3], value[4], value[5]))
	}
	for _, prefix := range blockedPrefixes {
		if prefix.Contains(address) {
			return false
		}
	}
	return true
}

func sameOrigin(first, second *url.URL) bool {
	if first == nil || second == nil {
		return false
	}
	first, firstErr := normalizeTargetURL(first.String())
	second, secondErr := normalizeTargetURL(second.String())
	if firstErr != nil || secondErr != nil || first.Scheme != second.Scheme || first.Hostname() != second.Hostname() {
		return false
	}
	port := func(value *url.URL) string {
		if value.Port() != "" {
			return value.Port()
		}
		if value.Scheme == "https" {
			return "443"
		}
		return "80"
	}
	return port(first) == port(second)
}

func SafeError(err error) error {
	if errors.Is(err, ErrTarget) {
		return ErrTarget
	}
	if errors.Is(err, ErrResponseTooLarge) {
		return ErrResponseTooLarge
	}
	if errors.Is(err, ErrResponse) {
		return ErrResponse
	}
	return ErrRequest
}

func ReadResponseBody(body io.Reader) ([]byte, error) {
	content, err := io.ReadAll(io.LimitReader(body, MaxResponseBytes+1))
	if err != nil {
		return nil, ErrResponse
	}
	if len(content) > MaxResponseBytes {
		return nil, ErrResponseTooLarge
	}
	return content, nil
}
