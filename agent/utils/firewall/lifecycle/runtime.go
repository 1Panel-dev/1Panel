package lifecycle

import (
	"errors"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

const (
	ProviderFirewalld = "firewalld"
	ProviderUFW       = "ufw"
	ProviderIptables  = "iptables"
	ProviderNftables  = "nftables"
)

type IptablesCommands struct {
	IPv4     string
	IPv6     string
	Restore4 string
	Restore6 string
}

func (c IptablesCommands) IPv6Available() bool {
	return c.IPv6 != "" && c.Restore6 != ""
}

type Runtime struct {
	Provider string
	Iptables IptablesCommands
}

var which = cmd.Which

// DetectRuntime follows the host-firewall priority used by 1Panel. The default
// xtables command family wins over explicitly named xtables-nft tools; native
// nftables is selected only when neither xtables command family is usable.
func DetectRuntime() (Runtime, error) {
	hasFirewalld := which("firewalld")
	hasUFW := which("ufw")
	if hasFirewalld && hasUFW {
		return Runtime{}, errors.New("it is detected that the system has both firewalld and ufw services. To avoid conflicts, please uninstall and try again")
	}
	if hasFirewalld {
		return Runtime{Provider: ProviderFirewalld}, nil
	}
	if hasUFW {
		return Runtime{Provider: ProviderUFW}, nil
	}
	if commands, ok := detectIptablesCommands(""); ok {
		return Runtime{Provider: ProviderIptables, Iptables: commands}, nil
	}
	if commands, ok := detectIptablesCommands("-nft"); ok {
		return Runtime{Provider: ProviderIptables, Iptables: commands}, nil
	}
	if which("nft") {
		return Runtime{Provider: ProviderNftables}, nil
	}
	return Runtime{}, errors.New("no system firewall service detected (firewalld/ufw/iptables/iptables-nft/nft), please check and try again")
}

func detectIptablesCommands(suffix string) (IptablesCommands, bool) {
	ipv4 := "iptables" + suffix
	restore4 := "iptables" + suffix + "-restore"
	if !which(ipv4) || !which(restore4) {
		return IptablesCommands{}, false
	}
	commands := IptablesCommands{IPv4: ipv4, Restore4: restore4}
	ipv6 := "ip6tables" + suffix
	restore6 := "ip6tables" + suffix + "-restore"
	if which(ipv6) && which(restore6) {
		commands.IPv6 = ipv6
		commands.Restore6 = restore6
	}
	return commands, true
}

func ResolveIptablesCommands() (IptablesCommands, error) {
	if commands, ok := detectIptablesCommands(""); ok {
		return commands, nil
	}
	if commands, ok := detectIptablesCommands("-nft"); ok {
		return commands, nil
	}
	return IptablesCommands{}, fmt.Errorf("no complete iptables command family is available")
}
