package service

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/1Panel-dev/1Panel/agent/utils/toolbox"
)

const defaultFail2BanPath = "/etc/fail2ban/jail.local"

var fail2BanConfigMu sync.Mutex

type Fail2BanService struct{}

type IFail2BanService interface {
	LoadBaseInfo() (dto.Fail2BanBaseInfo, error)
	Search(search dto.Fail2BanSearch) ([]string, error)
	Operate(operation string) error
	OperateSSHD(req dto.Fail2BanSet) error
	UpdateConf(req dto.Fail2BanUpdate) error
	UpdateConfByFile(req dto.UpdateByFile) error
}

func NewIFail2BanService() IFail2BanService {
	return &Fail2BanService{}
}

func (u *Fail2BanService) LoadBaseInfo() (dto.Fail2BanBaseInfo, error) {
	var baseInfo dto.Fail2BanBaseInfo
	client, err := toolbox.NewFail2Ban()
	if err != nil {
		return baseInfo, err
	}
	baseInfo.IsEnable, baseInfo.IsActive, baseInfo.IsExist = client.Status()
	if !baseInfo.IsExist {
		baseInfo.Version = "-"
	} else {
		baseInfo.Version = client.Version()
	}
	conf, err := os.ReadFile(defaultFail2BanPath)
	if err != nil {
		if baseInfo.IsActive {
			return baseInfo, fmt.Errorf("read fail2ban conf of %s failed, err: %v", defaultFail2BanPath, err)
		} else {
			return baseInfo, nil
		}
	}
	values := make(map[string]string)
	lines := strings.Split(string(conf), "\n")
	for _, target := range []string{"DEFAULT", "sshd"} {
		section := ""
		for _, line := range lines {
			nextSection, key, value := parseFail2BanLine(line)
			if nextSection != "" {
				section = nextSection
			}
			if section == target && key != "" {
				values[key] = value
			}
		}
	}
	baseInfo.Port, _ = strconv.Atoi(values["port"])
	baseInfo.MaxRetry, _ = strconv.Atoi(values["maxretry"])
	baseInfo.FindTime = values["findtime"]
	baseInfo.BanTime = values["bantime"]
	baseInfo.BanAction = values["banaction"]
	baseInfo.LogPath = values["logpath"]
	baseInfo.Mode = values["mode"]
	if match := re.GetRegex(re.Fail2BanModeValuePattern).FindStringSubmatch(values["filter"]); len(match) == 2 {
		baseInfo.Mode = strings.Trim(strings.TrimSpace(match[1]), `"'`)
	}
	if baseInfo.Mode == "" {
		baseInfo.Mode = "normal"
	}

	return baseInfo, nil
}

func (u *Fail2BanService) Search(req dto.Fail2BanSearch) ([]string, error) {
	var list []string
	client, err := toolbox.NewFail2Ban()
	if err != nil {
		return nil, err
	}
	if req.Status == "banned" {
		list, err = client.ListBanned()

	} else {
		list, err = client.ListIgnore()
	}
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (u *Fail2BanService) Operate(operation string) error {
	client, err := toolbox.NewFail2Ban()
	if err != nil {
		return err
	}
	return client.Operate(operation)
}

func (u *Fail2BanService) OperateSSHD(req dto.Fail2BanSet) error {
	if req.Operate == "ignore" {
		return u.UpdateConf(dto.Fail2BanUpdate{Key: "ignoreip", Value: strings.Join(req.IPs, ",")})
	}
	client, err := toolbox.NewFail2Ban()
	if err != nil {
		return err
	}
	return client.ReBanIPs(req.IPs)
}

func (u *Fail2BanService) UpdateConf(req dto.Fail2BanUpdate) error {
	fail2BanConfigMu.Lock()
	defer fail2BanConfigMu.Unlock()
	req.Value = strings.TrimSpace(req.Value)
	if req.Key == "mode" {
		switch req.Value {
		case "normal", "ddos", "extra", "aggressive":
		default:
			return fmt.Errorf("invalid Fail2ban mode: %s", req.Value)
		}
		client, err := toolbox.NewFail2Ban()
		if err != nil {
			return err
		}
		if common.CompareVersion("0.10.0", client.Version()) {
			return fmt.Errorf("Fail2ban mode requires version 0.10.0 or later")
		}
	}

	if req.Key == "banaction" {
		if req.Value == "firewallcmd-ipset" || req.Value == "ufw" {
			itemName := "ufw"
			if req.Value == "firewallcmd-ipset" {
				itemName = "firewalld"
			}
			client, err := NewSelectedSystemFirewallClient()
			if err != nil {
				return err
			}
			if client.Name() != itemName {
				return buserr.WithName("ErrBanAction", itemName)
			}
			isActive, _ := client.Status()
			if !isActive {
				return buserr.WithName("ErrBanAction", itemName)
			}
		}
	}
	conf, err := os.ReadFile(defaultFail2BanPath)
	if err != nil {
		return fmt.Errorf("read fail2ban conf of %s failed, err: %v", defaultFail2BanPath, err)
	}
	if req.Key == "logpath" {
		if _, err := os.Stat(req.Value); err != nil {
			return err
		}
	}
	ending := "\n"
	if strings.Contains(string(conf), "\r\n") {
		ending = "\r\n"
	}
	lines := strings.SplitAfter(string(conf), "\n")
	section := ""
	hasSection, found := false, false
	insert := len(lines)
	for index, line := range lines {
		nextSection, key, _ := parseFail2BanLine(line)
		if nextSection != "" {
			if section == "sshd" && insert == len(lines) {
				insert = index
			}
			section = nextSection
			if section == "sshd" {
				hasSection = true
			}
			continue
		}
		if section == "sshd" && req.Key == "mode" && key == "filter" {
			pattern := re.GetRegex(re.Fail2BanModeAssignmentPattern)
			if pattern.MatchString(line) {
				lines[index] = pattern.ReplaceAllString(line, "${1}"+req.Value)
			} else {
				pattern = re.GetRegex(re.Fail2BanFilterNamePattern)
				lines[index] = pattern.ReplaceAllString(line, "${1}[mode="+req.Value+"]")
			}
		}
		if section != "sshd" || key != req.Key {
			continue
		}
		lines[index] = ""
		if !found {
			indent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
			lines[index] = indent + req.Key + " = " + req.Value
			if strings.HasSuffix(line, "\n") {
				lines[index] += ending
			}
			found = true
		}
	}
	newFile := strings.Join(lines, "")
	if !found {
		before := strings.Join(lines[:insert], "")
		if before != "" && !strings.HasSuffix(before, "\n") {
			before += ending
		}
		if !hasSection {
			before += "[sshd]" + ending
		}
		newFile = before + req.Key + " = " + req.Value + ending + strings.Join(lines[insert:], "")
	}
	if err := saveFail2BanConfig(defaultFail2BanPath, newFile); err != nil {
		return err
	}

	client, err := toolbox.NewFail2Ban()
	if err != nil {
		return err
	}
	return client.Operate("restart")
}

func (u *Fail2BanService) UpdateConfByFile(req dto.UpdateByFile) error {
	file, err := os.OpenFile(defaultFail2BanPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()

	client, err := toolbox.NewFail2Ban()
	if err != nil {
		return err
	}
	if err := client.Operate("restart"); err != nil {
		return err
	}
	return nil
}

func parseFail2BanLine(line string) (section, key, value string) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		return "", "", ""
	}
	if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
		return strings.TrimSpace(line[1 : len(line)-1]), "", ""
	}
	key, value, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", ""
	}
	return "", strings.ToLower(strings.TrimSpace(key)), strings.TrimSpace(value)
}

func saveFail2BanConfig(path, content string) error {
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(target), ".fail2ban-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	defer file.Close()
	if owner, ok := info.Sys().(*syscall.Stat_t); ok {
		if err := file.Chown(int(owner.Uid), int(owner.Gid)); err != nil {
			return err
		}
	}
	if err := file.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	return os.Rename(file.Name(), target)
}
