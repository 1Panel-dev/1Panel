package mcp

import (
	"encoding/json"
	"errors"
	"net/netip"
	"slices"
	"strings"
)

const MaxAllowedIPs = 64

func NormalizeAllowedIPs(entries []string) ([]string, error) {
	if len(entries) > MaxAllowedIPs {
		return nil, errors.New("too many MCP IP entries")
	}
	result := make([]string, 0, len(entries))
	for _, raw := range entries {
		entry := strings.TrimSpace(raw)
		if len(entry) == 0 || len(entry) > 64 {
			return nil, errors.New("invalid MCP IP entry")
		}
		if strings.Contains(entry, "/") {
			prefix, err := netip.ParsePrefix(entry)
			if err != nil || prefix.Addr().Zone() != "" {
				return nil, errors.New("invalid MCP CIDR")
			}
			if prefix.Addr().Is4In6() {
				if prefix.Bits() < 96 {
					return nil, errors.New("invalid mapped IPv4 CIDR")
				}
				prefix = netip.PrefixFrom(prefix.Addr().Unmap(), prefix.Bits()-96)
			}
			result = append(result, prefix.Masked().String())
		} else {
			address, err := netip.ParseAddr(entry)
			if err != nil || address.Zone() != "" {
				return nil, errors.New("invalid MCP address")
			}
			result = append(result, address.Unmap().String())
		}
	}
	slices.Sort(result)
	return slices.Compact(result), nil
}

func DecodeAllowedIPs(raw string) ([]string, error) {
	var result []string
	if json.Unmarshal([]byte(raw), &result) != nil || result == nil {
		return nil, errors.New("invalid stored MCP IP policy")
	}
	return NormalizeAllowedIPs(result)
}

func IPAllowed(entries []string, source string) bool {
	address, err := netip.ParseAddr(source)
	if err != nil || address.Zone() != "" {
		return false
	}
	address = address.Unmap()
	for _, entry := range entries {
		if prefix, err := netip.ParsePrefix(entry); err == nil {
			if prefix.Contains(address) {
				return true
			}
		} else if expected, err := netip.ParseAddr(entry); err == nil && expected.Unmap() == address {
			return true
		}
	}
	return false
}
