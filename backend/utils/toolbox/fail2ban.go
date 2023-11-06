package toolbox

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
)

type Fail2Ban struct{}

const defaultPath = "/etc/fail2ban/jail.local"

type FirewallClient interface {
	Status() (bool, bool, error)
	Version() (string, error)
	Operation(operate string) error
	Init() error

	ListBanned() ([]string, error)
}

func NewFail2Ban() (*Fail2Ban, error) {
	if _, err := os.Stat(defaultPath); err != nil {
		if err := initLocalFile(); err != nil {
			return nil, err
		}
	}
	return &Fail2Ban{}, nil
}

func (f *Fail2Ban) Status() (bool, bool, error) {
	isEnable, _ := systemctl.IsEnable("fail2ban.service")
	isActive, _ := systemctl.IsActive("fail2ban.service")

	return isEnable, isActive, nil
}

func (f *Fail2Ban) Version() (string, error) {
	stdout, err := cmd.Exec("fail2ban-client version")
	if err != nil {
		return "", fmt.Errorf("load the fail2ban version failed, err: %s", stdout)
	}
	return strings.ReplaceAll(stdout, "\n ", ""), nil
}

func (f *Fail2Ban) Operation(operate string) error {
	switch operate {
	case "start", "restart", "stop", "enable", "disable":
		stdout, err := cmd.Execf("systemctl %s fail2ban.service", operate)
		if err != nil {
			return fmt.Errorf("%s the fail2ban.service failed, err: %s", operate, stdout)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Fail2Ban) ListBanned() ([]string, error) {
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

func (f *Fail2Ban) ListIgnore() ([]string, error) {
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
ignoreip = 127.0.0.1/8,172.16.10.114,172.16.10.116
bantime = 600
findtime = 300
maxretry = 5
banaction = firewallcmd-ipset
action = %(action_mwl)s
#DEFAULT-END

#sshd-START
[sshd]
enabled = true
filter = sshd
port = 22
maxretry = 5
findtime = 300
bantime = 86400
action = %(action_mwl)s
logpath = /var/log/secure
#sshd-END`

	if err := os.WriteFile(defaultPath, []byte(initFile), 0640); err != nil {
		return err
	}
	return nil
}
