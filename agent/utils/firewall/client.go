package firewall

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/client"
)

// FilterClient is the filter capability surface. Rule expansion lives behind
// Expand*/Apply* so that callers never build provider native commands themselves.
type FilterClient interface {
	Name() string // ufw firewalld iptables
	Start() error
	Stop() error
	Restart() error
	Reload() error
	Status() (bool, error)
	Version() (string, error)

	ListPort() ([]client.FireInfo, error)
	ListAddress() ([]client.FireInfo, error)

	Port(port client.FireInfo, operation string) error

	// ExpandPortRule turns one logical rule into the ordered native operations
	// this provider needs. It runs no command.
	ExpandPortRule(rule client.FireInfo) []client.PortUnit
	ApplyPortUnit(unit client.PortUnit, operation string) error
	ExpandAddressRule(rule client.FireInfo) []client.AddressUnit
	ApplyAddressUnit(unit client.AddressUnit, operation string) error

	// AddPortWhiteList re-adds the whole whitelist after the provider has been
	// started, SyncPortWhiteList applies the difference against PortWhiteList.Previous.
	AddPortWhiteList(list client.PortWhiteList) error
	SyncPortWhiteList(list client.PortWhiteList) error
}

func NewFirewallClient() (FilterClient, error) {
	provider, err := DetectProvider()
	if err != nil {
		return nil, err
	}
	return newClientByName(provider)
}

func newClientByName(name string) (FilterClient, error) {
	switch name {
	case "firewalld":
		return client.NewFirewalld()
	case "ufw":
		return client.NewUfw()
	case "iptables":
		return client.NewIptables()
	default:
		return nil, errors.New("unsupported firewall provider: " + name)
	}
}

func LoadPingStatus() string {
	data, err := os.ReadFile("/proc/sys/net/ipv4/icmp_echo_ignore_all")
	if err != nil {
		return constant.StatusNone
	}
	v6Data, v6err := os.ReadFile("/proc/sys/net/ipv6/icmp/echo_ignore_all")
	if v6err != nil {
		if strings.TrimSpace(string(data)) == "1" {
			return constant.StatusEnable
		}
		return constant.StatusDisable
	} else {
		if strings.TrimSpace(string(data)) == "1" && strings.TrimSpace(string(v6Data)) == "1" {
			return constant.StatusEnable
		}
		return constant.StatusDisable
	}
}

func UpdatePingStatus(enable string) error {
	const confPath = "/etc/sysctl.conf"
	const panelSysctlPath = "/etc/sysctl.d/98-onepanel.conf"

	var targetPath string
	applyArgs := []string{"-p"}

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		targetPath = panelSysctlPath
		applyArgs = []string{"--system"}
		if err := cmd.NewCommandMgr().RunWithOptionalSudo("mkdir", "-p", "/etc/sysctl.d"); err != nil {
			return fmt.Errorf("failed to create directory /etc/sysctl.d: %v", err)
		}
	} else {
		targetPath = confPath
	}

	lineBytes, err := os.ReadFile(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %v", targetPath, err)
	}

	if err := cmd.WriteFileWithOptionalSudo("/proc/sys/net/ipv4/icmp_echo_ignore_all", []byte(enable), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to apply ipv4 ping status temporarily: %v", err)
	}

	var hasIpv6 bool
	if _, err := os.Stat("/proc/sys/net/ipv6/icmp/echo_ignore_all"); err == nil {
		hasIpv6 = true
		if err := cmd.WriteFileWithOptionalSudo("/proc/sys/net/ipv6/icmp/echo_ignore_all", []byte(enable), constant.FilePerm); err != nil {
			global.LOG.Warnf("failed to apply ipv6 ping status temporarily: %v", err)
		}
	}

	var files []string
	if err == nil {
		files = strings.Split(string(lineBytes), "\n")
	}

	var newFiles []string
	hasIPv4Line, hasIPv6Line := false, false

	for _, line := range files {
		if strings.HasPrefix(strings.TrimSpace(line), "net.ipv4.icmp_echo_ignore_all") {
			newFiles = append(newFiles, "net.ipv4.icmp_echo_ignore_all="+enable)
			hasIPv4Line = true
			continue
		}
		if strings.HasPrefix(strings.TrimSpace(line), "net.ipv6.icmp.echo_ignore_all") {
			newFiles = append(newFiles, "net.ipv6.icmp.echo_ignore_all="+enable)
			hasIPv6Line = true
			continue
		}
		newFiles = append(newFiles, line)
	}

	if !hasIPv4Line {
		newFiles = append(newFiles, "net.ipv4.icmp_echo_ignore_all="+enable)
	}
	if hasIpv6 && !hasIPv6Line {
		newFiles = append(newFiles, "net.ipv6.icmp.echo_ignore_all="+enable)
	}

	if err = cmd.WriteFileWithOptionalSudo(targetPath, []byte(strings.Join(newFiles, "\n")), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to write to %s: %v", targetPath, err)
	}

	if err := cmd.NewCommandMgr().RunWithOptionalSudo("sysctl", applyArgs...); err != nil {
		global.LOG.Warnf("failed to apply persistent config: %v", err)
	}

	return nil
}
