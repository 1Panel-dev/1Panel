package firewall

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

// DetectProvider reuses the NewFirewallClient selection order without building a
// client, for the callers that only need the provider name:
// firewalld+ufw conflict -> firewalld -> ufw -> iptables.
func DetectProvider() (string, error) {
	return resolveProviderFromPresence(cmd.Which("firewalld"), cmd.Which("ufw"), cmd.Which("iptables"))
}

func resolveProviderFromPresence(firewalld, ufw, iptables bool) (string, error) {
	if firewalld && ufw {
		return "", errors.New("It is detected that the system has both firewalld and ufw services. To avoid conflicts, please uninstall and try again!")
	}
	if firewalld {
		return "firewalld", nil
	}
	if ufw {
		return "ufw", nil
	}
	if iptables {
		return "iptables", nil
	}
	return "", errors.New("No system firewall service detected (firewalld/ufw/iptables), please check and try again!")
}
