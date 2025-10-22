package controller

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/controller/manager"
)

type Controller interface {
	Name() string
	IsActive(serviceName string) (bool, error)
	IsEnable(serviceName string) (bool, error)
	IsExist(serviceName string) (bool, error)
	Status(serviceName string) (string, error)

	Operate(operate, serviceName string) error

	Reload() error
}

func New() (Controller, error) {
	managerOptions := []string{"systemd", "openrc", "sysvinit"}
	for _, item := range managerOptions {
		if _, err := exec.LookPath(item); err != nil {
			continue
		}
		switch item {
		case "systemd":
			return manager.NewSystemd(), nil
		case "openrc":
			return manager.NewOpenrc(), nil
		case "sysvinit":
			return manager.NewSysvinit(), nil
		}
	}
	return nil, errors.New("not support such manager initializatio")
}

func Handle(operate, serviceName string) error {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return err
	}
	client, err := New()
	if err != nil {
		return err
	}
	return client.Operate(operate, service)
}
func HandleStart(serviceName string) error {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return err
	}
	return Handle("start", service)
}
func HandleStop(serviceName string) error {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return err
	}
	return Handle("stop", service)
}
func HandleRestart(serviceName string) error {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return err
	}
	return Handle("restart", service)
}

func CheckExist(serviceName string) (bool, error) {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return false, err
	}
	client, err := New()
	if err != nil {
		return false, err
	}
	b, er := client.IsExist(service)
	return b, er
}
func CheckActive(serviceName string) (bool, error) {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return false, err
	}
	client, err := New()
	if err != nil {
		return false, err
	}
	return client.IsActive(service)
}
func CheckEnable(serviceName string) (bool, error) {
	service, err := LoadServiceName(serviceName)
	if err != nil {
		return false, err
	}
	client, err := New()
	if err != nil {
		return false, err
	}
	return client.IsEnable(service)
}

func Reload() error {
	client, err := New()
	if err != nil {
		return err
	}
	return client.Reload()
}

func RestartPanel(core, agent, reload bool) {
	client, err := New()
	if err != nil {
		global.LOG.Errorf("load client for controller failed, err: %v", err)
		return
	}
	if reload {
		if err := client.Reload(); err != nil {
			global.LOG.Errorf("restart 1panel service failed, err: %v", err)
			return
		}
	}
	if agent {
		if err := client.Operate("restart", "1panel-agent"); err != nil {
			global.LOG.Errorf("restart 1panel agent service failed, err: %v", err)
			return
		}
	}
	if core {
		if err := client.Operate("restart", "1panel-core"); err != nil {
			global.LOG.Errorf("restart 1panel core service failed, err: %v", err)
			return
		}
	}
}

func LoadServiceName(keyword string) (string, error) {
	client, err := New()
	if err != nil {
		return "", err
	}

	processedName := loadProcessedName(client.Name(), keyword)
	exist, err := client.IsExist(processedName)
	if exist {
		return processedName, nil
	}
	alistName := loadFromPredefined(client, keyword)
	if len(alistName) != 0 {
		return alistName, nil
	}
	return "", fmt.Errorf("find such service for %s failed", keyword)
}

func loadProcessedName(mgr, keyword string) string {
	keyword = strings.ToLower(keyword)
	if strings.HasSuffix(keyword, ".service.socket") {
		keyword = strings.TrimSuffix(keyword, ".service.socket") + ".socket"
	}
	if mgr != "systemd" {
		keyword = strings.TrimSuffix(keyword, ".service")
		return keyword
	}
	if !strings.HasSuffix(keyword, ".service") && !strings.HasSuffix(keyword, ".socket") {
		keyword += ".service"
	}
	return keyword
}

func loadFromPredefined(mgr Controller, keyword string) string {
	predefinedMap := map[string][]string{
		"clam":         {"clamav-daemon.service", "clamd@scan.service", "clamd"},
		"freshclam":    {"clamav-freshclam.service", "freshclam.service"},
		"fail2ban":     {"fail2ban.service", "fail2ban"},
		"supervisor":   {"supervisord.service", "supervisor.service", "supervisord", "supervisor"},
		"ssh":          {"sshd.service", "ssh.service", "sshd", "ssh"},
		"1panel-core":  {"1panel-core.service", "1panel-cored"},
		"1panel-agent": {"1panel-agent.service", "1panel-agentd"},
		"docker":       {"docker.service", "dockerd"},
	}
	if val, ok := predefinedMap[keyword]; ok {
		for _, item := range val {
			if exist, _ := mgr.IsExist(item); exist {
				return item
			}
		}
	}
	return ""
}
