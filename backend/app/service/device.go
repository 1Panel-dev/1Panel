package service

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/ntp"
)

const defaultDNSPath = "/etc/resolv.conf"
const defaultHostPath = "/etc/hosts"

type DeviceService struct{}

type IDeviceService interface {
	LoadBaseInfo() (dto.DeviceBaseInfo, error)
	Update(key, value string) error
	UpdateHosts(req []dto.HostHelper) error
	UpdatePasswd(req dto.ChangePasswd) error
	UpdateByConf(req dto.UpdateByNameAndFile) error
	LoadTimeZone() ([]string, error)
	CheckDNS(key, value string) (bool, error)
	LoadConf(name string) (string, error)
}

func NewIDeviceService() IDeviceService {
	return &DeviceService{}
}

func (u *DeviceService) LoadBaseInfo() (dto.DeviceBaseInfo, error) {
	var baseInfo dto.DeviceBaseInfo
	baseInfo.LocalTime = time.Now().Format("2006-01-02 15:04:05 MST -0700")
	baseInfo.TimeZone = common.LoadTimeZoneByCmd()
	baseInfo.DNS = loadDNS()
	baseInfo.Hosts = loadHosts()
	baseInfo.Hostname = loadHostname()
	baseInfo.User = loadUser()
	ntp, _ := settingRepo.Get(settingRepo.WithByKey("NtpSite"))
	baseInfo.Ntp = ntp.Value

	return baseInfo, nil
}

func (u *DeviceService) LoadTimeZone() ([]string, error) {
	std, err := cmd.Exec("timedatectl list-timezones")
	if err != nil {
		return []string{}, err
	}
	return strings.Split(std, "\n"), nil
}

func (u *DeviceService) CheckDNS(key, value string) (bool, error) {
	content, err := os.ReadFile(defaultDNSPath)
	if err != nil {
		return false, err
	}
	defer func() { _ = u.UpdateByConf(dto.UpdateByNameAndFile{Name: "DNS", File: string(content)}) }()
	if key == "form" {
		if err := u.Update("DNS", value); err != nil {
			return false, err
		}
	} else {
		if err := u.UpdateByConf(dto.UpdateByNameAndFile{Name: "DNS", File: value}); err != nil {
			return false, err
		}
	}

	conn, err := net.DialTimeout("ip4:icmp", "www.baidu.com", time.Second*2)
	if err != nil {
		return false, err
	}
	defer conn.Close()

	return true, nil
}

func (u *DeviceService) Update(key, value string) error {
	switch key {
	case "TimeZone":
		if err := ntp.UpdateSystemTimeZone(value); err != nil {
			return err
		}
		go func() {
			_, err := cmd.Exec("systemctl restart 1panel.service")
			if err != nil {
				global.LOG.Errorf("restart system for new time zone failed, err: %v", err)
			}
		}()
	case "DNS":
		if err := updateDNS(strings.Split(value, ",")); err != nil {
			return err
		}
	case "Hostname":
		std, err := cmd.Execf("%s hostnamectl set-hostname %s", cmd.SudoHandleCmd(), value)
		if err != nil {
			return errors.New(std)
		}
	case "LocalTime":
		if err := settingRepo.Update("NtpSite", value); err != nil {
			return err
		}
		ntime, err := ntp.GetRemoteTime(value)
		if err != nil {
			return err
		}
		ts := ntime.Format("2006-01-02 15:04:05")
		if err := ntp.UpdateSystemTime(ts); err != nil {
			return err
		}
	default:
		return fmt.Errorf("not support such key %s", key)
	}
	return nil
}

func (u *DeviceService) UpdateHosts(req []dto.HostHelper) error {
	conf, err := os.ReadFile(defaultHostPath)
	if err != nil {
		return fmt.Errorf("read namesever conf of %s failed, err: %v", defaultHostPath, err)
	}
	lines := strings.Split(string(conf), "\n")
	newFile := ""
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			newFile += line + "\n"
			continue
		}
		for index, item := range req {
			if item.IP == parts[0] && item.Host == strings.Join(parts[1:], " ") {
				newFile += line + "\n"
				req = append(req[:index], req[index+1:]...)
				break
			}
		}
	}
	for _, item := range req {
		newFile += fmt.Sprintf("%s   %s \n", item.IP, item.Host)
	}
	file, err := os.OpenFile(defaultHostPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(newFile)
	write.Flush()
	return nil
}

func (u *DeviceService) UpdatePasswd(req dto.ChangePasswd) error {
	std, err := cmd.Execf("%s echo '%s:%s' | %s chpasswd", cmd.SudoHandleCmd(), req.User, req.Passwd, cmd.SudoHandleCmd())
	if err != nil {
		return errors.New(std)
	}
	return nil
}

func (u *DeviceService) LoadConf(name string) (string, error) {
	pathItem := ""
	switch name {
	case "DNS":
		pathItem = defaultDNSPath
	case "Hosts":
		pathItem = defaultHostPath
	default:
		return "", fmt.Errorf("not support such name %s", name)
	}
	if _, err := os.Stat(pathItem); err != nil {
		return "", err
	}
	content, err := os.ReadFile(pathItem)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func (u *DeviceService) UpdateByConf(req dto.UpdateByNameAndFile) error {
	if req.Name != "DNS" && req.Name != "Hosts" {
		return fmt.Errorf("not support such name %s", req.Name)
	}
	path := defaultDNSPath
	if req.Name == "Hosts" {
		path = defaultHostPath
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(req.File)
	write.Flush()
	return nil
}

func loadDNS() []string {
	var list []string
	dnsConf, err := os.ReadFile(defaultDNSPath)
	if err == nil {
		lines := strings.Split(string(dnsConf), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "nameserver ") {
				list = append(list, strings.TrimPrefix(line, "nameserver "))
			}
		}
	}
	return list
}

func updateDNS(list []string) error {
	conf, err := os.ReadFile(defaultDNSPath)
	if err != nil {
		return fmt.Errorf("read nameserver conf of %s failed, err: %v", defaultDNSPath, err)
	}
	lines := strings.Split(string(conf), "\n")
	newFile := ""
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 || parts[0] != "nameserver" {
			newFile += line + "\n"
			continue
		}
		itemNs := strings.Join(parts[1:], " ")
		for index, item := range list {
			if item == itemNs {
				newFile += line + "\n"
				list = append(list[:index], list[index+1:]...)
				break
			}
		}
	}
	for _, item := range list {
		newFile += fmt.Sprintf("nameserver %s \n", item)
	}
	file, err := os.OpenFile(defaultDNSPath, os.O_WRONLY|os.O_TRUNC, 0640)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, _ = write.WriteString(newFile)
	write.Flush()
	return nil
}

func loadHosts() []dto.HostHelper {
	var list []dto.HostHelper
	hostConf, err := os.ReadFile(defaultHostPath)
	if err == nil {
		lines := strings.Split(string(hostConf), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) < 2 {
				continue
			}
			list = append(list, dto.HostHelper{IP: parts[0], Host: strings.Join(parts[1:], " ")})
		}
	}
	return list
}

func loadHostname() string {
	std, err := cmd.Exec("hostname")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(std, "\n", "")
}

func loadUser() string {
	std, err := cmd.Exec("whoami")
	if err != nil {
		return ""
	}
	return strings.ReplaceAll(std, "\n", "")
}
