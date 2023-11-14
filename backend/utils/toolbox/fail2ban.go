package toolbox

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
)

type Fail2ban struct{}

const defaultPath = "/etc/fail2ban/jail.local"

type FirewallClient interface {
	Status() (bool, bool, error)
	Version() (string, error)
	Operate(operate string) error
	OperateSSHD(operate, ip string) error
}

func NewFail2ban() (*Fail2ban, error) {
	if _, err := os.Stat(defaultPath); err != nil {
		if err := initLocalFile(); err != nil {
			return nil, err
		}
		stdout, err := cmd.Exec("fail2ban-client reload")
		if err != nil {
			global.LOG.Errorf("reload fail2ban failed, err: %s", stdout)
			return nil, err
		}
	}
	return &Fail2ban{}, nil
}

func (f *Fail2ban) Status() (bool, bool) {
	isEnable, _ := systemctl.IsEnable("fail2ban.service")
	isActive, _ := systemctl.IsActive("fail2ban.service")

	return isEnable, isActive
}

func (f *Fail2ban) Version() string {
	stdout, err := cmd.Exec("fail2ban-client version")
	if err != nil {
		global.LOG.Errorf("load the fail2ban version failed, err: %s", stdout)
		return "-"
	}
	return strings.ReplaceAll(stdout, "\n", "")
}

func (f *Fail2ban) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop", "enable", "disable":
		stdout, err := cmd.Execf("systemctl %s fail2ban.service", operate)
		if err != nil {
			return fmt.Errorf("%s the fail2ban.service failed, err: %s", operate, stdout)
		}
		return nil
	case "reload":
		stdout, err := cmd.Exec("fail2ban-client reload")
		if err != nil {
			return fmt.Errorf("fail2ban-client reload, err: %s", stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Fail2ban) OperateSSHD(operate, ip string) error {
	switch operate {
	case "banip", "unbanip":
		stdout, err := cmd.Execf("fail2ban-client set sshd %s %s", operate, ip)
		if err != nil {
			return fmt.Errorf("%s the fail2ban.service failed, err: %s", operate, stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Fail2ban) ListBanned() ([]string, error) {
	var lists []string
	stdout, err := cmd.Exec("fail2ban-client get sshd banned")
	if err != nil {
		return lists, err
	}
	addrs := strings.Split(stdout, "'")
	for _, addr := range addrs {
		if len(addr) == 0 || addr == "[" || addr == "]" {
			continue
		}
		lists = append(lists, addr)
	}
	return lists, nil
}

func (f *Fail2ban) ListIgnore() ([]string, error) {
	var lists []string
	stdout, err := cmd.Exec("fail2ban-client get sshd ignoreip")
	if err != nil {
		return lists, err
	}
	stdout = strings.ReplaceAll(stdout, "|", "")
	stdout = strings.ReplaceAll(stdout, "`", "")
	stdout = strings.ReplaceAll(stdout, "\n", "")
	addrs := strings.Split(stdout, "-")
	for _, addr := range addrs {
		if !strings.HasPrefix(addr, " ") {
			continue
		}
		lists = append(lists, addr)
	}
	return lists, nil
}

func initLocalFile() error {
	if _, err := os.Create(defaultPath); err != nil {
		return err
	}
	initFile := `#DEFAULT-START
[DEFAULT]
bantime = 600
findtime = 300
maxretry = 5
banaction = firewallcmd-ipset
action = %(action_mwl)s
#DEFAULT-END

[sshd]
ignoreip = 127.0.0.1/8
enabled = true
filter = sshd
port = 22
maxretry = 5
findtime = 300
bantime = 600
action = %(action_mwl)s
logpath = /var/log/secure`
	if err := os.WriteFile(defaultPath, []byte(initFile), 0640); err != nil {
		return err
	}
	return nil
}
