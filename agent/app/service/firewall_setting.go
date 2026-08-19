package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
)

type IFirewallSettingService interface {
	Load(context.Context) (dto.FirewallSettings, error)
	Operate(context.Context, dto.FirewallBackendOperation) error
}

type FirewallSettingService struct{}

func NewIFirewallSettingService() IFirewallSettingService { return &FirewallSettingService{} }

func (s *FirewallSettingService) Load(ctx context.Context) (dto.FirewallSettings, error) {
	result := dto.FirewallSettings{PingStatus: ping.LoadStatus()}
	if ports, err := settingRepo.GetValueByKey(constant.FirewallPortWhiteList); err == nil {
		result.PortWhitelist = ports
	} else {
		result.PortWhitelist = constant.FirewallPortWhiteListValue
	}

	installed := make(map[string]bool)
	for _, name := range lifecycle.InstalledProviders() {
		installed[name] = true
	}
	for _, name := range []string{"firewalld", "ufw", "iptables", "nftables"} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed {
			client, err := lifecycle.NewClientFor(name)
			if err != nil {
				option.Message = err.Error()
			} else {
				option.Active, _ = client.Status()
				option.Initialized, option.Bound, _ = loadFirewallInitStatus(name, "base")
				option.IPv4.Initialized, option.IPv4.Bound, _ = loadSystemFirewallFamilyStatus(name, "ipv4")
				option.IPv6.Initialized, option.IPv6.Bound, _ = loadSystemFirewallFamilyStatus(name, "ipv6")
			}
		}
		if name == "iptables" {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.System.Options = append(result.System.Options, option)
	}
	result.System.Selected, _ = settingRepo.GetValueByKey(settingSystemFirewallBackend)
	if !installed[result.System.Selected] {
		result.System.Selected = ""
	}
	if result.System.Selected == "" {
		if client, err := lifecycle.NewClient(); err == nil {
			result.System.Selected = client.Name()
		}
	}
	result.System.Current = result.System.Selected

	initializedForwarding := make([]string, 0, 2)
	for _, name := range []string{"iptables", "nftables"} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed {
			manager, err := newForwardingManagerFor(name)
			if err != nil {
				option.Message = err.Error()
			} else if status, err := manager.Status(); err != nil {
				option.Message = err.Error()
			} else {
				option.Active, option.Initialized, option.Bound = status.IsActive, status.IsInit, status.IsBind
				option.IPv4.Initialized, option.IPv4.Bound, _ = manager.FamilyStatus("ipv4")
				option.IPv6.Initialized, option.IPv6.Bound, _ = manager.FamilyStatus("ipv6")
			}
		}
		if option.IPv4.Initialized || option.IPv6.Initialized {
			initializedForwarding = append(initializedForwarding, name)
		}
		if name == "iptables" {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.Forwarding.Options = append(result.Forwarding.Options, option)
	}
	if len(initializedForwarding) == 1 {
		result.Forwarding.Current = initializedForwarding[0]
	} else if len(initializedForwarding) > 1 {
		result.Forwarding.Current = strings.Join(initializedForwarding, " + ")
	}
	result.Forwarding.Selected, _ = settingRepo.GetValueByKey(settingForwardingBackend)
	if !installed[result.Forwarding.Selected] {
		result.Forwarding.Selected = ""
	}
	if result.Forwarding.Selected == "" {
		for _, option := range result.Forwarding.Options {
			if option.Initialized && result.Forwarding.Selected == "" {
				result.Forwarding.Selected = option.Name
			} else if option.Initialized {
				result.Forwarding.Selected = ""
				break
			}
		}
	}

	dockerInstalled := cmd.Which("docker")
	detectedDockerBackend := ""
	if dockerInstalled {
		if client, err := docker.NewDockerClient(); err == nil {
			defer client.Close()
			if info, err := client.Info(ctx); err == nil {
				detectedDockerBackend = dockerFirewallBackend(info)
			}
		}
	}
	result.Docker.Selected = selectedDockerFirewallBackend(detectedDockerBackend)
	result.Docker.Current = result.Docker.Selected
	for _, name := range []string{"iptables", "nftables"} {
		option := dto.FirewallBackendOption{
			Name: name, Installed: dockerInstalled && installed[name], Supported: true,
			Active: dockerInstalled && result.Docker.Selected == name,
		}
		var guard dockerGuardRuntime = docker_guard.NewManager()
		if name == "nftables" {
			guard = docker_guard.NewNftablesManager()
		}
		ipv4, ipv6 := guard.Status(docker_guard.FamilyIPv4), guard.Status(docker_guard.FamilyIPv6)
		option.Initialized = ipv4.Initialized || ipv6.Initialized
		option.Bound = ipv4.Bound || ipv6.Bound
		option.IPv4.Initialized, option.IPv4.Bound = ipv4.Initialized, ipv4.Bound
		option.IPv6.Initialized, option.IPv6.Bound = ipv6.Initialized, ipv6.Bound
		result.Docker.Options = append(result.Docker.Options, option)
	}

	return result, nil
}

func loadSystemFirewallFamilyStatus(provider, family string) (bool, bool, error) {
	switch provider {
	case "firewalld", "ufw":
		return true, true, nil
	case "iptables":
		return iptables_helper.LoadFamilyInitStatus(family, "base")
	case "nftables":
		return nftables_helper.LoadFamilyInitStatus(filter.Family(family), "base")
	default:
		return false, false, fmt.Errorf("unsupported firewall provider %q", provider)
	}
}

func (s *FirewallSettingService) Operate(ctx context.Context, request dto.FirewallBackendOperation) error {
	if request.Subsystem != "system" && request.Backend != "iptables" && request.Backend != "nftables" {
		return fmt.Errorf("%s only supports iptables or nftables", request.Subsystem)
	}
	switch request.Subsystem {
	case "system":
		return s.operateSystem(request)
	case "forwarding":
		return s.operateForwarding(request)
	case "docker":
		return s.operateDocker(ctx, request)
	default:
		return fmt.Errorf("unsupported firewall subsystem %q", request.Subsystem)
	}
}

func (s *FirewallSettingService) operateDocker(ctx context.Context, request dto.FirewallBackendOperation) error {
	var guard dockerGuardRuntime = docker_guard.NewManager()
	if request.Backend == "nftables" {
		guard = docker_guard.NewNftablesManager()
	}
	if request.Operation == "cleanup" {
		if err := guard.Cleanup(); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(settingDockerPortGuardStatus, constant.StatusDisable)
	}
	previous, _ := settingRepo.GetValueByKey(settingDockerFirewallBackend)
	if err := settingRepo.UpdateOrCreate(settingDockerFirewallBackend, request.Backend); err != nil {
		return err
	}
	if request.Operation == "initialize" {
		if err := NewIDockerPortGuardService().Operate(ctx, dto.DockerPortGuardOperation{Operation: "initialize"}); err != nil {
			_ = settingRepo.UpdateOrCreate(settingDockerFirewallBackend, previous)
			return err
		}
	}
	return nil
}

func (s *FirewallSettingService) operateSystem(request dto.FirewallBackendOperation) error {
	if _, err := lifecycle.NewClientFor(request.Backend); err != nil {
		return err
	}
	if request.Operation == "cleanup" {
		return cleanupSystemBackend(request.Backend)
	}
	previous, _ := settingRepo.GetValueByKey(settingSystemFirewallBackend)
	if previous == "" {
		if client, err := lifecycle.NewClient(); err == nil {
			previous = client.Name()
		}
	}
	if err := settingRepo.UpdateOrCreate(settingSystemFirewallBackend, request.Backend); err != nil {
		return err
	}
	rollback := func(err error) error {
		if err == nil {
			return nil
		}
		_ = settingRepo.UpdateOrCreate(settingSystemFirewallBackend, previous)
		return err
	}
	if request.Operation == "select" {
		return nil
	}
	var initErr error
	if request.Backend == "iptables" || request.Backend == "nftables" {
		initErr = newFirewallService().OperateFilterChain(dto.FilterChainOperation{Name: "1PANEL_BASIC", Operate: "init-base"})
	} else {
		initErr = newFirewallService().OperateFirewall(dto.FirewallLifecycleOperation{Operation: "start"})
	}
	if initErr != nil {
		return rollback(initErr)
	}
	if request.Backend == "iptables" || request.Backend == "nftables" {
		return settingRepo.UpdateOrCreate(settingFilterInitialized, constant.StatusEnable)
	}
	return nil
}

func cleanupSystemBackend(backend string) error {
	switch backend {
	case "iptables":
		return newIptablesHelperManager().Cleanup()
	case "nftables":
		return newNftablesHelperManager().Cleanup()
	default:
		return fmt.Errorf("cleanup is only available for 1Panel-owned iptables and nftables resources")
	}
}

func (s *FirewallSettingService) operateForwarding(request dto.FirewallBackendOperation) error {
	manager, err := newForwardingManagerFor(request.Backend)
	if err != nil {
		return err
	}
	if request.Operation == "cleanup" {
		if err := manager.Cleanup(); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(settingForwardingInitialized, constant.StatusDisable); err != nil {
			return err
		}
		recordForwardingSyncError(nil)
		return nil
	}
	if err := settingRepo.UpdateOrCreate(settingForwardingBackend, request.Backend); err != nil {
		return err
	}
	if request.Operation == "initialize" {
		return NewIForwardingService().Enable()
	}
	recordForwardingSyncError(nil)
	return nil
}
