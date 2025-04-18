package toolbox

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/systemctl"
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
	isExist, _ := systemctl.IsExist("fail2ban")
	if isExist {
		if _, err := os.Stat(defaultPath); err != nil {
			if err := initLocalFile(); err != nil {
				return nil, err
			}
			err := systemctl.Restart("fail2ban")
			if err != nil {
				global.LOG.Errorf("restart fail2ban failed, err: %s", err)
				return nil, err
			}
		}
	}
	return &Fail2ban{}, nil
}

func (f *Fail2ban) Status() (bool, bool, bool) {
	isEnable, _ := systemctl.IsEnable("fail2ban.service")
	isActive, _ := systemctl.IsActive("fail2ban.service")
	isExist, _ := systemctl.IsExist("fail2ban.service")

	return isEnable, isActive, isExist
}

func (f *Fail2ban) Version() string {
	stdout, err := cmd.Exec("fail2ban-client --version")
	if err != nil {
		global.LOG.Errorf("load the fail2ban version failed, err: %s", stdout)
		return "-"
	}
	versionRe := regexp.MustCompile(`(?i)fail2ban[:\s-]*v?(\d+\.\d+\.\d+)`)
	matches := versionRe.FindStringSubmatch(stdout)
	if len(matches) > 1 {
		return matches[1]
	}
	global.LOG.Errorf("Version regex failed to match output: %s", stdout)
	return "-"
}
func (f *Fail2ban) Operate(operate string) error {
	switch operate {
	case "start", "restart", "stop", "enable", "disable":
		stdout, err := systemctl.CustomAction(operate, "fail2ban")
		if err != nil {
			return fmt.Errorf("%s the fail2ban failed, err: %s", operate, stdout.Output)
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

func (f *Fail2ban) ReBanIPs(ips []string) error {
	ipItems, _ := f.ListBanned()
	stdout, err := cmd.Execf("fail2ban-client unban --all")
	if err != nil {
		stdout1, err := cmd.Execf("fail2ban-client set sshd banip %s", strings.Join(ipItems, " "))
		if err != nil {
			global.LOG.Errorf("rebanip after fail2ban-client unban --all failed, err: %s", stdout1)
		}
		return fmt.Errorf("fail2ban-client unban --all failed, err: %s", stdout)
	}
	stdout1, err := cmd.Execf("fail2ban-client set sshd banip %s", strings.Join(ips, " "))
	if err != nil {
		return fmt.Errorf("handle `fail2ban-client set sshd banip %s` failed, err: %s", strings.Join(ips, " "), stdout1)
	}
	return nil
}

func (f *Fail2ban) ListBanned() ([]string, error) {
	var lists []string
	stdout, err := cmd.Exec("fail2ban-client status sshd | grep 'Banned IP list:'")
	if err != nil {
		return lists, err
	}
	itemList := strings.Split(strings.Trim(stdout, "\n"), "Banned IP list:")
	if len(itemList) != 2 {
		return lists, nil
	}

	ips := strings.Fields(itemList[1])
	for _, item := range ips {
		if len(item) != 0 {
			lists = append(lists, item)
		}
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
	var initFile string
	if systemctl.GetGlobalManager().Name() == "openrc" {
		initFile = `[sshd]
enabled  = true
filter   = alpine-sshd
port     = $ssh_port
logpath  = $logpath
maxretry = 2
banaction = $banaction

[sshd-ddos]
enabled  = true
filter   = alpine-sshd-ddos
port     = $ssh_port
logpath  = /var/log/messages
maxretry = 2

[sshd-key]
enabled  = true
filter   = alpine-sshd-key
port     = $ssh_port
logpath  = /var/log/messages
maxretry = 2
`
	} else {
		initFile = `# DEFAULT-START
[DEFAULT]
bantime = 600
findtime = 300
maxretry = 5
banaction = $banaction
action = %(action_mwl)s
#DEFAULT-END

[sshd]
ignoreip = 127.0.0.1/8 ::1
enabled = true
filter = sshd
port = $ssh_port
maxretry = 5
findtime = 300
bantime = 600
banaction = $banaction
action = %(action_mwl)s
logpath = $logpath
maxmatches = 3
`
	}
	// 检测防火墙类型
	banaction := detectFirewall()

	// 检测SSH端口（支持Alpine的sshd_config位置）
	sshPort := detectSSHPort()

	// 检测日志路径（兼容Alpine的多种日志位置）
	logPath := detectAuthLogPath()

	// 执行变量替换
	initFile = strings.ReplaceAll(initFile, "$banaction", banaction)
	initFile = strings.ReplaceAll(initFile, "$logpath", logPath)
	initFile = strings.ReplaceAll(initFile, "$ssh_port", sshPort)

	if err := os.WriteFile(defaultPath, []byte(initFile), 0640); err != nil {
		return err
	}
	return nil
}

func detectFirewall() string {
	if active, _ := systemctl.IsActive("firewalld"); active {
		return "firewallcmd-ipset"
	}
	if active, _ := systemctl.IsActive("ufw"); active {
		return "ufw"
	}
	if active, _ := systemctl.IsActive("iptables"); active {
		return "iptables-allports"
	}
	if active, _ := systemctl.IsActive("nftables"); active {
		return "nftables-allports"
	}
	return "iptables-allports"
}

func detectAuthLogPath() string {
	paths := []string{
		"/var/log/auth.log", // 常见默认路径
		"/var/log/messages", // Alpine系统日志
		"/var/log/secure",   // RHEL风格路径
		"/var/log/syslog",   // Debian风格路径
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return "/var/log/auth.log"
}

func detectSSHPort() string {
	// 检查标准sshd_config路径
	configPaths := []string{
		"/etc/ssh/sshd_config",
		"/etc/sshd_config",
	}

	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "Port") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return parts[1]
				}
			}
		}
	}
	return "22" // 默认SSH端口
}
