package ping

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

func LoadStatus() string {
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
	}
	if strings.TrimSpace(string(data)) == "1" && strings.TrimSpace(string(v6Data)) == "1" {
		return constant.StatusEnable
	}
	return constant.StatusDisable
}

func UpdateStatus(enable string) error {
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

	hasIPv6 := false
	if _, err := os.Stat("/proc/sys/net/ipv6/icmp/echo_ignore_all"); err == nil {
		hasIPv6 = true
		if err := cmd.WriteFileWithOptionalSudo("/proc/sys/net/ipv6/icmp/echo_ignore_all", []byte(enable), constant.FilePerm); err != nil {
			global.LOG.Warnf("failed to apply ipv6 ping status temporarily: %v", err)
		}
	}

	var lines []string
	if err == nil {
		lines = strings.Split(string(lineBytes), "\n")
	}
	updated := make([]string, 0, len(lines)+2)
	hasIPv4Line, hasIPv6Line := false, false
	for _, line := range lines {
		switch {
		case strings.HasPrefix(strings.TrimSpace(line), "net.ipv4.icmp_echo_ignore_all"):
			updated = append(updated, "net.ipv4.icmp_echo_ignore_all="+enable)
			hasIPv4Line = true
		case strings.HasPrefix(strings.TrimSpace(line), "net.ipv6.icmp.echo_ignore_all"):
			updated = append(updated, "net.ipv6.icmp.echo_ignore_all="+enable)
			hasIPv6Line = true
		default:
			updated = append(updated, line)
		}
	}
	if !hasIPv4Line {
		updated = append(updated, "net.ipv4.icmp_echo_ignore_all="+enable)
	}
	if hasIPv6 && !hasIPv6Line {
		updated = append(updated, "net.ipv6.icmp.echo_ignore_all="+enable)
	}

	if err = cmd.WriteFileWithOptionalSudo(targetPath, []byte(strings.Join(updated, "\n")), constant.FilePerm); err != nil {
		return fmt.Errorf("failed to write to %s: %v", targetPath, err)
	}
	if err := cmd.NewCommandMgr().RunWithOptionalSudo("sysctl", applyArgs...); err != nil {
		global.LOG.Warnf("failed to apply persistent config: %v", err)
	}
	return nil
}
