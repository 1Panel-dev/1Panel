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

type FirewallClient interface {
	Name() string // ufw firewalld
	Start() error
	Stop() error
	Restart() error
	Reload() error
	Status() (bool, error)
	Version() (string, error)

	ListPort() ([]client.FireInfo, error)
	ListForward() ([]client.FireInfo, error)
	ListAddress() ([]client.FireInfo, error)

	Port(port client.FireInfo, operation string) error
	RichRules(rule client.FireInfo, operation string) error
	PortForward(info client.Forward, operation string) error

	EnableForward() error
}

func NewFirewallClient() (FirewallClient, error) {
	firewalld := cmd.Which("firewalld")
	ufw := cmd.Which("ufw")

	if firewalld && ufw {
		return nil, errors.New("It is detected that the system has both firewalld and ufw services. To avoid conflicts, please uninstall and try again!")
	}
	if firewalld {
		return client.NewFirewalld()
	}
	if ufw {
		return client.NewUfw()
	}

	iptables := cmd.Which("iptables")
	if iptables {
		return client.NewIptables()
	}
	return nil, errors.New("No system firewall service detected (firewalld/ufw/iptables), please check and try again!")
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
	var applyCmd string

	if _, err := os.Stat(confPath); os.IsNotExist(err) {
		targetPath = panelSysctlPath
		applyCmd = fmt.Sprintf("%s sysctl --system", cmd.SudoHandleCmd())
		if err := cmd.RunDefaultBashCf("%s mkdir -p /etc/sysctl.d", cmd.SudoHandleCmd()); err != nil {
			return fmt.Errorf("failed to create directory /etc/sysctl.d: %v", err)
		}
	} else {
		targetPath = confPath
		applyCmd = fmt.Sprintf("%s sysctl -p", cmd.SudoHandleCmd())
	}

	lineBytes, err := os.ReadFile(targetPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %v", targetPath, err)
	}

	if err := cmd.RunDefaultBashCf("echo %s | %s tee /proc/sys/net/ipv4/icmp_echo_ignore_all > /dev/null", enable, cmd.SudoHandleCmd()); err != nil {
		return fmt.Errorf("failed to apply ipv4 ping status temporarily: %v", err)
	}

	var hasIpv6 bool
	if _, err := os.Stat("/proc/sys/net/ipv6/icmp/echo_ignore_all"); err == nil {
		hasIpv6 = true
		if err := cmd.RunDefaultBashCf("echo %s | %s tee /proc/sys/net/ipv6/icmp/echo_ignore_all > /dev/null", enable, cmd.SudoHandleCmd()); err != nil {
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

	file, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		return fmt.Errorf("failed to open %s: %v", targetPath, err)
	}
	defer file.Close()

	if _, err = file.WriteString(strings.Join(newFiles, "\n")); err != nil {
		return fmt.Errorf("failed to write to %s: %v", targetPath, err)
	}

	if err := cmd.RunDefaultBashC(applyCmd); err != nil {
		global.LOG.Warnf("failed to apply persistent config with '%s': %v", applyCmd, err)
	}

	return nil
}
