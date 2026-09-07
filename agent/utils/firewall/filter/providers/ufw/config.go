package ufw

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/utils/ini_conf"
)

func IPv6Enabled() (bool, error) {
	return ipv6Enabled("/etc/default/ufw", "/proc/net/if_inet6")
}

func ipv6Enabled(configPath, kernelPath string) (bool, error) {
	value, err := ini_conf.GetIniValue(configPath, "", "IPV6")
	if err != nil {
		return false, fmt.Errorf("load UFW IPv6 configuration: %w", err)
	}
	if !strings.EqualFold(strings.TrimSpace(value), "yes") {
		return false, nil
	}
	if _, err := os.Stat(kernelPath); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("inspect kernel IPv6 support: %w", err)
	}
	return true, nil
}
