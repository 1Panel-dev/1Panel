package service

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/ntp"
)

const defaultNameServerPath = "/etc/resolv.conf"
const defaultHostPath = "/etc/hosts"

type DeviceService struct{}

type IDeviceService interface {
	LoadBaseInfo() (dto.DeviceBaseInfo, error)
	Update(key, value string) error
	UpdateHosts(req []dto.HostHelper) error
	LoadTimeZone() ([]dto.TimeZoneOptions, error)
}

func NewIDeviceService() IDeviceService {
	return &DeviceService{}
}

func (u *DeviceService) LoadBaseInfo() (dto.DeviceBaseInfo, error) {
	var baseInfo dto.DeviceBaseInfo
	baseInfo.LocalTime = time.Now().Format("2006-01-02 15:04:05 MST -0700")
	baseInfo.TimeZone = common.LoadTimeZoneByCmd()
	baseInfo.NameServers = loadNameServers()
	baseInfo.Hosts = loadHosts()

	return baseInfo, nil
}

func (u *DeviceService) LoadTimeZone() ([]dto.TimeZoneOptions, error) {
	std, err := cmd.Exec("timedatectl list-timezones")
	if err != nil {
		return nil, nil
	}

	optionMap := make(map[string][]string)
	zones := strings.Split(std, "\n")
	for _, zone := range zones {
		items := strings.Split(zone, "/")
		if len(items) < 2 {
			continue
		}
		optionMap[items[0]] = append(optionMap[items[0]], items[1])
	}

	var list []dto.TimeZoneOptions
	for k, v := range optionMap {
		list = append(list, dto.TimeZoneOptions{From: k, Zones: v})
	}
	return list, nil
}

func (u *DeviceService) Update(key, value string) error {
	switch key {
	case "TimeZone":
		if err := ntp.UpdateSystemTimeZone(value); err != nil {
			return err
		}
	case "NameServer":
		if err := updateNameServers(strings.Split(value, ",")); err != nil {
			return err
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
		if !strings.Contains(line, " ") {
			newFile += line + "\n"
			continue
		}
		for index, item := range req {
			if item.IP+item.Host == strings.TrimSpace(line) {
				newFile += line + "\n"
				req = append(req, req[:index]...)
				req = append(req, req[index+1:]...)
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

func loadNameServers() []string {
	var list []string
	nameServerConf, err := os.ReadFile(defaultNameServerPath)
	if err != nil {
		lines := strings.Split(string(nameServerConf), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "nameserver ") {
				list = append(list, strings.TrimPrefix(line, "nameserver "))
			}
		}
	}
	return list
}

func updateNameServers(list []string) error {
	conf, err := os.ReadFile(defaultNameServerPath)
	if err != nil {
		return fmt.Errorf("read namesever conf of %s failed, err: %v", defaultNameServerPath, err)
	}
	lines := strings.Split(string(conf), "\n")
	newFile := ""
	for _, line := range lines {
		if !strings.Contains(line, "nameserver ") {
			newFile += line + "\n"
			continue
		}
		itemNs := strings.Split(line, "nameserver  ")[1]
		for index, item := range list {
			if item == itemNs {
				newFile += line + "\n"
				list = append(list, list[:index]...)
				list = append(list, list[index+1:]...)
				break
			}
		}
	}
	for _, item := range list {
		newFile += fmt.Sprintf("nameserver %s \n", item)
	}
	file, err := os.OpenFile(defaultNameServerPath, os.O_WRONLY|os.O_TRUNC, 0640)
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
			if strings.Contains(line, " ") {
				items := strings.Split(line, " ")
				list = append(list, dto.HostHelper{IP: items[0], Host: items[1]})
			}
		}
	}
	return list
}
