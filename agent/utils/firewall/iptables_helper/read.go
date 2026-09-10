package iptables_helper

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
)

func ReadTable(ctx context.Context, table string, ipv6 bool) (string, error) {
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return "", err
	}
	executable := commands.IPv4
	if ipv6 {
		if !commands.IPv6Available() {
			return "", fmt.Errorf("%w: ip6tables/ip6tables-restore are not installed", filter.ErrFamilyUnavailable)
		}
		executable = commands.IPv6
	}
	output, err := runTables(ctx, executable, table, false, true, "-S")
	if err != nil && ipv6 && (strings.Contains(err.Error(), "Address family not supported") || strings.Contains(err.Error(), "Protocol not supported")) {
		return output, fmt.Errorf("%w: %v", filter.ErrFamilyUnavailable, err)
	}
	return output, err
}
