package toolbox

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

type Fail2ban struct{}

const defaultPath = "/etc/fail2ban/jail.local"

type FirewallClient interface {
	Status() (bool, bool, bool)
	Version() (string, error)
	Operate(operate string) error
	OperateSSHD(operate, ip string) error
}

func NewFail2Ban() (*Fail2ban, error) {
	isExist, _ := controller.CheckExist("fail2ban.service")
	if isExist {
		if _, err := os.Stat(defaultPath); err != nil {
			if err := initLocalFile(); err != nil {
				return nil, err
			}
			if err := controller.HandleRestart("fail2ban.service"); err != nil {
				global.LOG.Errorf("restart fail2ban failed, err: %v", err)
				return nil, err
			}
		}
	}
	return &Fail2ban{}, nil
}

func (f *Fail2ban) Status() (bool, bool, bool) {
	isEnable, _ := controller.CheckEnable("fail2ban.service")
	isActive, _ := controller.CheckActive("fail2ban.service")
	isExist, _ := controller.CheckExist("fail2ban.service")
	if !isActive && isFail2banAlive() {
		isActive = true
	}

	return isEnable, isActive, isExist
}

func isFail2banAlive() bool {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(5*time.Second)).RunWithStdout("fail2ban-client", "ping")
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(stdout), "pong")
}

func (f *Fail2ban) Version() string {
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("fail2ban-client", "version")
	if err != nil {
		global.LOG.Errorf("load the fail2ban version failed, %v", err)
		return "-"
	}
	return strings.ReplaceAll(stdout, "\n", "")
}

func (f *Fail2ban) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop", "enable", "disable":
		if err := controller.Handle(operate, "fail2ban.service"); err != nil {
			return fmt.Errorf("%s the fail2ban.service failed, err: %v", operate, err)
		}
		return nil
	case "reload":
		if err := cmd.NewCommandMgr().Run("fail2ban-client", "reload"); err != nil {
			return fmt.Errorf("fail2ban-client reload, %v", err)
		}
		return nil
	default:
		return fmt.Errorf("not support such operation: %v", operate)
	}
}

func (f *Fail2ban) ReBanIPs(ips []string) error {
	ipItems, _ := f.ListBanned()
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(10*time.Minute)).RunWithStdout("fail2ban-client", "unban", "--all")
	if err != nil {
		if len(ipItems) != 0 {
			args := append([]string{"set", "sshd", "banip"}, ipItems...)
			stdout1, err := cmd.NewCommandMgr(cmd.WithTimeout(10*time.Minute)).RunWithStdout("fail2ban-client", args...)
			if err != nil {
				global.LOG.Errorf("rebanip after fail2ban-client unban --all failed, err: %s", stdout1)
			}
		}
		return fmt.Errorf("fail2ban-client unban --all failed, err: %s", stdout)
	}
	if len(ips) == 0 {
		return nil
	}
	args := append([]string{"set", "sshd", "banip"}, ips...)
	stdout, err = cmd.NewCommandMgr(cmd.WithTimeout(10*time.Minute)).RunWithStdout("fail2ban-client", args...)
	if err != nil {
		return fmt.Errorf("handle `fail2ban-client set sshd banip %s` failed, err: %s", strings.Join(ips, " "), stdout)
	}
	return nil
}

func (f *Fail2ban) ListBanned() ([]string, error) {
	var lists []string
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("fail2ban-client", "status", "sshd")
	if err != nil {
		return lists, err
	}
	for _, line := range strings.Split(strings.Trim(stdout, "\n"), "\n") {
		itemList := strings.Split(line, "Banned IP list:")
		if len(itemList) != 2 {
			continue
		}

		ips := strings.Fields(itemList[1])
		for _, item := range ips {
			if len(item) != 0 {
				lists = append(lists, item)
			}
		}
		break
	}
	return lists, nil
}

func (f *Fail2ban) ListIgnore() ([]string, error) {
	var lists []string
	stdout, err := cmd.NewCommandMgr(cmd.WithTimeout(20*time.Second)).RunWithStdout("fail2ban-client", "get", "sshd", "ignoreip")
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
		lists = append(lists, strings.ReplaceAll(addr, " ", ""))
	}
	return lists, nil
}

func initLocalFile() error {
	f, err := os.Create(defaultPath)
	if err != nil {
		return err
	}
	defer f.Close()
	initFile := `#DEFAULT-START
[DEFAULT]
bantime = 600
findtime = 300
maxretry = 5
banaction = $banaction
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
banaction = $banaction
action = %(action_mwl)s
logpath = $logpath`

	banaction := ""
	if active, _ := controller.CheckActive("firewalld"); active {
		banaction = "firewallcmd-ipset"
	} else if active, _ := controller.CheckActive("ufw"); active {
		banaction = "ufw"
	} else {
		banaction = "iptables-allports"
	}
	initFile = strings.ReplaceAll(initFile, "$banaction", banaction)

	logPath := ""
	if _, err := os.Stat("/var/log/secure"); err == nil {
		logPath = "/var/log/secure"
	} else {
		logPath = "/var/log/auth.log"
	}
	initFile = strings.ReplaceAll(initFile, "$logpath", logPath)
	if err := os.WriteFile(defaultPath, []byte(initFile), 0640); err != nil {
		return err
	}
	return nil
}
