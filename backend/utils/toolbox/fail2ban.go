package toolbox

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
)

type Fail2ban struct {
	serviceConfig *systemctl.ServiceConfig
	fwConfigs     map[string]*systemctl.ServiceConfig
}

const defaultPath = "/etc/fail2ban/jail.local"

var fail2banService = &systemctl.ServiceConfig{
	ID:          "fail2ban",
	DisplayName: "Fail2ban Service",
	ServiceName: map[string]string{
		"systemd":  "fail2ban.service",
		"openrc":   "fail2ban",
		"sysvinit": "fail2ban",
	},
	UseSocket:   false,
	Description: "Fail2ban intrusion prevention",
}

func NewFail2Ban() (*Fail2ban, error) {
	fwConfigs := map[string]*systemctl.ServiceConfig{
		"firewalld": {
			ID:          "firewalld",
			ServiceName: map[string]string{"systemd": "firewalld.service"},
		},
		"ufw": {
			ID:          "ufw",
			ServiceName: map[string]string{"sysvinit": "ufw"},
		},
	}

	f := &Fail2ban{
		serviceConfig: fail2banService,
		fwConfigs:     fwConfigs,
	}

	exist, err := systemctl.IsExist(f.serviceConfig)
	if err != nil || !exist {
		return nil, fmt.Errorf("fail2ban service not found: %v", err)
	}

	if _, err := os.Stat(defaultPath); err != nil {
		if err := f.initLocalFile(); err != nil {
			return nil, err
		}
		if err := f.Operate("restart"); err != nil {
			global.LOG.Errorf("restart fail2ban failed: %v", err)
			return nil, err
		}
	}
	return f, nil
}

func (f *Fail2ban) Status() (bool, bool, bool) {
	enable, _ := systemctl.IsEnabled(f.serviceConfig)
	active, _ := systemctl.IsActive(f.serviceConfig)
	exist, _ := systemctl.IsExist(f.serviceConfig)
	return enable, active, exist
}

func (f *Fail2ban) Version() string {
	stdout, err := cmd.Exec("fail2ban-client version")
	if err != nil {
		global.LOG.Errorf("get version failed: %v", err)
		return "-"
	}
	return strings.TrimSpace(stdout)
}

func (f *Fail2ban) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop", "enable", "disable":
		return systemctl.Operate(operate, f.serviceConfig)
	case "reload":
		stdout, err := cmd.Exec("fail2ban-client reload")
		if err != nil {
			return fmt.Errorf("reload failed: %s", stdout)
		}
		return nil
	default:
		return fmt.Errorf("unsupported operation: %s", operate)
	}
}

func (f *Fail2ban) ReBanIPs(ips []string) error {
	if _, err := cmd.Exec("fail2ban-client unban --all"); err != nil {
		return fmt.Errorf("unban failed: %v", err)
	}

	ipList := strings.Join(ips, " ")
	stdout, err := cmd.Execf("fail2ban-client set sshd banip %s", ipList)
	if err != nil {
		return fmt.Errorf("ban failed: %s", stdout)
	}
	return nil
}

func (f *Fail2ban) ListBanned() ([]string, error) {
	stdout, err := cmd.Exec("fail2ban-client status sshd")
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, line := range strings.Split(stdout, "\n") {
		if strings.Contains(line, "Banned IP list:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				ips = strings.Fields(parts[1])
			}
		}
	}
	return ips, nil
}

func (f *Fail2ban) ListIgnore() ([]string, error) {
	stdout, err := cmd.Exec("fail2ban-client get sshd ignoreip")
	if err != nil {
		return nil, err
	}

	cleaned := strings.NewReplacer(
		"|", "",
		"`", "",
		"\n", "",
	).Replace(stdout)

	var addrs []string
	for _, part := range strings.Split(cleaned, " ") {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			addrs = append(addrs, trimmed)
		}
	}
	return addrs, nil
}

func (f *Fail2ban) initLocalFile() error {
	initFile := `[DEFAULT]
bantime = 600
findtime = 300
maxretry = 5
banaction = %s
action = %%(action_mwl)s

[sshd]
ignoreip = 127.0.0.1/8
enabled = true
filter = sshd
port = 22
maxretry = 5
findtime = 300
bantime = 600
logpath = %s`

	banAction := f.detectBanAction()
	logPath := f.detectLogPath()

	config := fmt.Sprintf(initFile, banAction, logPath)
	if err := os.WriteFile(defaultPath, []byte(config), 0640); err != nil {
		return fmt.Errorf("write config failed: %v", err)
	}
	return nil
}

func (f *Fail2ban) detectBanAction() string {
	if active, _ := systemctl.IsActive(f.fwConfigs["firewalld"]); active {
		return "firewallcmd-ipset"
	}
	if active, _ := systemctl.IsActive(f.fwConfigs["ufw"]); active {
		return "ufw"
	}
	return "iptables-allports"
}

func (f *Fail2ban) detectLogPath() string {
	if _, err := os.Stat("/var/log/secure"); err == nil {
		return "/var/log/secure"
	}
	return "/var/log/auth.log"
}
