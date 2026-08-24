package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
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
	for _, name := range []string{
		constant.FirewallProviderFirewalld,
		constant.FirewallProviderUFW,
		constant.FirewallProviderIptables,
		constant.FirewallProviderNftables,
	} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed {
			client, err := lifecycle.NewClientFor(name)
			if err != nil {
				option.Message = err.Error()
			} else if supportsManagedFilterChains(name) {
				option.Initialized, option.Bound, err = loadFirewallInitStatus(name, "base")
				if err != nil {
					option.Message = err.Error()
				}
				option.IPv4 = loadSystemFirewallFamilyInfo(name, constant.FirewallFamilyIPv4)
				option.IPv6 = loadSystemFirewallFamilyInfo(name, constant.FirewallFamilyIPv6)
			} else if option.Active, err = client.Status(); err != nil {
				option.Message = err.Error()
			}
		}
		if name == constant.FirewallProviderIptables {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.System.Options = append(result.System.Options, option)
	}
	result.System.Selected, _ = settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
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
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed {
			manager, err := newForwardingManagerFor(name)
			if err != nil {
				option.Message = err.Error()
			} else if status, err := manager.Status(); err != nil {
				option.Message = err.Error()
			} else {
				option.Initialized, option.Bound = status.IsInit, status.IsBind
				ipv4Init, ipv4Bound, ipv4Err := manager.FamilyStatus(constant.FirewallFamilyIPv4)
				ipv6Init, ipv6Bound, ipv6Err := manager.FamilyStatus(constant.FirewallFamilyIPv6)
				option.IPv4 = dto.FirewallBackendFamilyStatus{
					Available: ipv4Err == nil, Initialized: ipv4Init, Bound: ipv4Bound,
				}
				option.IPv6 = dto.FirewallBackendFamilyStatus{
					Available: ipv6Err == nil, Initialized: ipv6Init, Bound: ipv6Bound,
				}
				if name == constant.FirewallProviderIptables {
					if commands, commandErr := lifecycle.ResolveIptablesCommands(); commandErr == nil {
						option.IPv6.Available = option.IPv6.Available && commands.IPv6Available()
						if !commands.IPv6Available() {
							option.IPv6.Reason = docker_guard.ReasonCommandMissing
						}
					}
				}
			}
		}
		if option.IPv4.Initialized || option.IPv6.Initialized {
			initializedForwarding = append(initializedForwarding, name)
		}
		if name == constant.FirewallProviderIptables {
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
	result.Forwarding.Selected, _ = settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
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
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{
			Name: name, Installed: dockerInstalled && installed[name], Supported: true,
			Active: dockerInstalled && result.Docker.Selected == name,
		}
		var guard dockerGuardRuntime = docker_guard.NewManager()
		if name == constant.FirewallProviderNftables {
			guard = docker_guard.NewNftablesManager()
		}
		ipv4, ipv6 := guard.Status(docker_guard.FamilyIPv4), guard.Status(docker_guard.FamilyIPv6)
		option.Initialized = ipv4.Initialized || ipv6.Initialized
		option.Bound = ipv4.Bound || ipv6.Bound
		option.IPv4.Initialized, option.IPv4.Bound = ipv4.Initialized, ipv4.Bound
		option.IPv6.Initialized, option.IPv6.Bound = ipv6.Initialized, ipv6.Bound
		option.IPv4.Available = ipv4.Reason != docker_guard.ReasonCommandMissing
		option.IPv6.Available = ipv6.Reason != docker_guard.ReasonCommandMissing
		option.IPv4.Reason, option.IPv6.Reason = ipv4.Reason, ipv6.Reason
		result.Docker.Options = append(result.Docker.Options, option)
	}

	return result, nil
}

func loadSystemFirewallFamilyStatus(provider, family string) (bool, bool, error) {
	switch provider {
	case constant.FirewallProviderIptables:
		return iptables_helper.LoadFamilyInitStatus(family, "base")
	case constant.FirewallProviderNftables:
		return nftables_helper.LoadFamilyInitStatus(filter.Family(family), "base")
	default:
		return false, false, fmt.Errorf("unsupported firewall provider %q", provider)
	}
}

func loadSystemFirewallFamilyInfo(provider, family string) dto.FirewallBackendFamilyStatus {
	if provider == constant.FirewallProviderIptables && family == constant.FirewallFamilyIPv6 {
		commands, err := lifecycle.ResolveIptablesCommands()
		if err != nil || !commands.IPv6Available() {
			return dto.FirewallBackendFamilyStatus{Reason: docker_guard.ReasonCommandMissing}
		}
	}
	initialized, bound, err := loadSystemFirewallFamilyStatus(provider, family)
	return dto.FirewallBackendFamilyStatus{
		Available:   err == nil,
		Initialized: initialized,
		Bound:       bound,
	}
}

func (s *FirewallSettingService) Operate(ctx context.Context, request dto.FirewallBackendOperation) error {
	if request.Subsystem != "system" && request.Backend != constant.FirewallProviderIptables && request.Backend != constant.FirewallProviderNftables {
		return fmt.Errorf("%s only supports iptables or nftables", request.Subsystem)
	}
	if request.Subsystem == "system" && !supportsManagedFilterChains(request.Backend) && request.Operation != "select" {
		return fmt.Errorf("%s does not support initialization or cleanup", request.Backend)
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
	if request.Backend == constant.FirewallProviderNftables {
		guard = docker_guard.NewNftablesManager()
	}
	if request.Operation == "cleanup" {
		if err := guard.Cleanup(); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable)
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, request.Backend); err != nil {
		return err
	}
	if request.Operation == "initialize" {
		if err := NewIDockerPortGuardService().Operate(ctx, dto.DockerPortGuardOperation{Operation: "initialize"}); err != nil {
			_ = settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, previous)
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
	previous, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	if previous == "" {
		if client, err := lifecycle.NewClient(); err == nil {
			previous = client.Name()
		}
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallSystemBackendKey, request.Backend); err != nil {
		return err
	}
	rollback := func(err error) error {
		if err == nil {
			return nil
		}
		_ = settingRepo.UpdateOrCreate(constant.FirewallSystemBackendKey, previous)
		return err
	}
	if request.Operation == "select" {
		return nil
	}
	initErr := newFirewallService().OperateFilterChain(dto.FilterChainOperation{
		Name: constant.FirewallBasicChain, Operate: string(firewall.BaseOperationInit),
	})
	if initErr != nil {
		return rollback(initErr)
	}
	return settingRepo.UpdateOrCreate(constant.FirewallFilterInitializedKey, constant.StatusEnable)
}

func cleanupSystemBackend(backend string) error {
	switch backend {
	case constant.FirewallProviderIptables:
		return newIptablesHelperManager().Cleanup()
	case constant.FirewallProviderNftables:
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
		if err := settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusDisable); err != nil {
			return err
		}
		recordForwardingSyncError(nil)
		return nil
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, request.Backend); err != nil {
		return err
	}
	if request.Operation == "initialize" {
		return NewIForwardingService().Enable()
	}
	recordForwardingSyncError(nil)
	return nil
}
