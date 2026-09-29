package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"gorm.io/gorm"
)

type IFirewallSettingService interface {
	CreatePortWhitelist(context.Context, dto.FirewallPortWhitelistCreate) error
	UpdatePortWhitelist(context.Context, dto.FirewallPortWhitelistUpdate) error
	DeletePortWhitelist(context.Context, dto.FirewallPortWhitelistDelete) error
	Load(context.Context) (dto.FirewallSettings, error)
	Operate(context.Context, dto.FirewallBackendOperation) error
}

type FirewallSettingService struct{}

var firewallWhitelistMu sync.Mutex

func (s *FirewallSettingService) CreatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistCreate) error {
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := loadPortWhitelistSetting(tx)
		if err != nil {
			return err
		}
		current = append(current, request.Rule)
		current, err = firewall.ValidatePortWhitelist(current)
		if err != nil {
			return err
		}
		value, err := json.Marshal(current)
		if err != nil {
			return err
		}
		err = tx.Where("key = ?", constant.FirewallPortWhiteList).Assign(map[string]interface{}{"value": string(value)}).FirstOrCreate(&model.Setting{Key: constant.FirewallPortWhiteList}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *FirewallSettingService) UpdatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistUpdate) error {
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := loadPortWhitelistSetting(tx)
		if err != nil {
			return err
		}
		index, err := findPortWhitelistRule(current, request.OldRule)
		if err != nil {
			return err
		}
		current[index] = request.Rule
		current, err = firewall.ValidatePortWhitelist(current)
		if err != nil {
			return err
		}
		value, err := json.Marshal(current)
		if err != nil {
			return err
		}
		err = tx.Where("key = ?", constant.FirewallPortWhiteList).Assign(map[string]interface{}{"value": string(value)}).FirstOrCreate(&model.Setting{Key: constant.FirewallPortWhiteList}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *FirewallSettingService) DeletePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistDelete) error {
	if request.Rule == nil {
		return fmt.Errorf("select one firewall port whitelist rule to delete")
	}
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := loadPortWhitelistSetting(tx)
		if err != nil {
			return err
		}
		index, err := findPortWhitelistRule(current, *request.Rule)
		if err != nil {
			return err
		}
		current = slices.Delete(current, index, index+1)
		current, err = firewall.ValidatePortWhitelist(current)
		if err != nil {
			return err
		}
		value, err := json.Marshal(current)
		if err != nil {
			return err
		}
		err = tx.Where("key = ?", constant.FirewallPortWhiteList).Assign(map[string]interface{}{"value": string(value)}).FirstOrCreate(&model.Setting{Key: constant.FirewallPortWhiteList}).Error
		if err != nil {
			return err
		}
		return nil
	})
}

func (s *FirewallSettingService) Load(ctx context.Context) (dto.FirewallSettings, error) {
	result := dto.FirewallSettings{PingStatus: firewall.LoadPingStatus()}

	installed := make(map[string]bool)
	for _, name := range lifecycle.InstalledProviders() {
		installed[name] = true
	}
	systemBackend, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	result.System.Selected = strings.TrimSpace(systemBackend)
	if result.System.Selected == "" {
		if client, err := lifecycle.NewClient(""); err == nil {
			result.System.Selected = client.Name()
		}
	}
	result.System.Current = result.System.Selected
	for _, name := range []string{
		constant.FirewallProviderFirewalld,
		constant.FirewallProviderUFW,
		constant.FirewallProviderIptables,
		constant.FirewallProviderNftables,
	} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed && name == result.System.Selected {
			client, err := lifecycle.NewClient(name)
			if err != nil {
				option.Message = err.Error()
			} else if name == constant.FirewallProviderIptables || name == constant.FirewallProviderNftables {
				overview, err := loadSystemFirewallOverview(name, "base")
				if err != nil {
					option.Message = err.Error()
				}
				option.Initialized, option.Bound = overview.IsInit, overview.IsBind
				option.IPv4, option.IPv6 = overview.IPv4, overview.IPv6
			} else if option.Active, err = client.Status(); err != nil {
				option.Message = err.Error()
			}
		}
		if name == result.System.Selected && name == constant.FirewallProviderIptables {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.System.Options = append(result.System.Options, option)
	}

	forwardingBackend, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	result.Forwarding.Selected = strings.TrimSpace(forwardingBackend)
	if result.Forwarding.Selected == "" {
		result.Forwarding.Selected = constant.FirewallProviderIptables
	}
	result.Forwarding.Current = result.Forwarding.Selected
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed && name == result.Forwarding.Selected {
			manager, err := newForwardingAdapterFor(name)
			if err != nil {
				option.Message = err.Error()
			} else {
				status, statusErr := loadForwardingFirewallOverview(manager)
				option.IPv4, option.IPv6 = status.IPv4, status.IPv6
				if statusErr != nil {
					option.Message = statusErr.Error()
				} else {
					option.Initialized, option.Bound = status.IsInit, status.IsBind
				}
				if name == constant.FirewallProviderIptables && !option.IPv6.Available {
					if commands, err := lifecycle.ResolveIptablesCommands(); err == nil && !commands.IPv6Available() {
						option.IPv6.Reason = dockerfirewall.ReasonCommandMissing
					}
				}
			}
		}
		if name == result.Forwarding.Selected && name == constant.FirewallProviderIptables {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.Forwarding.Options = append(result.Forwarding.Options, option)
	}

	dockerInstalled := cmd.Which("docker")
	dockerVersion := ""
	if dockerInstalled {
		dockerVersion = loadDockerEngineVersion(ctx)
	}
	dockerBackend, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	dockerBackend = strings.ToLower(strings.TrimSpace(dockerBackend))
	if dockerBackend == constant.FirewallProviderIptables || dockerBackend == constant.FirewallProviderNftables {
		result.Docker.Selected = dockerBackend
	}
	result.Docker.Current = result.Docker.Selected
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{
			Name: name, Installed: installed[name], Supported: dockerInstalled,
			Active: dockerInstalled && installed[name] && result.Docker.Selected == name,
		}
		if name == constant.FirewallProviderNftables && dockerInstalled && !dockerNftablesSupported(dockerVersion) {
			option.Supported = false
			option.SupportReason = "docker_version_unsupported"
			option.Active = false
		}
		if option.Active {
			guard := newDockerFirewallRuntime(name)
			ipv4, ipv6 := guard.Status(dockerfirewall.FamilyIPv4), guard.Status(dockerfirewall.FamilyIPv6)
			option.Initialized = ipv4.Initialized || ipv6.Initialized
			option.Bound = ipv4.Bound || ipv6.Bound
			option.IPv4.Initialized, option.IPv4.Bound = ipv4.Initialized, ipv4.Bound
			option.IPv6.Initialized, option.IPv6.Bound = ipv6.Initialized, ipv6.Bound
			option.IPv4.Available = ipv4.Reason != dockerfirewall.ReasonCommandMissing
			option.IPv6.Available = ipv6.Reason != dockerfirewall.ReasonCommandMissing
			option.IPv4.Reason, option.IPv6.Reason = ipv4.Reason, ipv6.Reason
		}
		result.Docker.Options = append(result.Docker.Options, option)
	}
	var err error
	result.PortWhitelist, err = loadPortWhitelistSetting(global.DB.WithContext(ctx))
	if err != nil {
		return result, err
	}
	result.PanelPort = LoadPanelPort()
	sshPort, sshErr := loadSSHWhitelistPortFrom(sshPath)
	if sshErr != nil {
		global.LOG.Warnf("load SSH port for firewall settings: %v", sshErr)
	} else {
		result.SSHPort = sshPort
	}
	return result, err
}

func (s *FirewallSettingService) Operate(ctx context.Context, request dto.FirewallBackendOperation) error {
	if err := lockFirewallLifecycleIdle(); err != nil {
		return err
	}
	defer firewallLifecycleTaskMu.Unlock()
	if request.Subsystem != "system" && request.Backend != constant.FirewallProviderIptables && request.Backend != constant.FirewallProviderNftables {
		return fmt.Errorf("%s only supports iptables or nftables", request.Subsystem)
	}
	if request.Subsystem == "system" && (request.Backend != constant.FirewallProviderIptables && request.Backend != constant.FirewallProviderNftables) && request.Operation != "select" {
		return fmt.Errorf("%s does not support initialization or cleanup", request.Backend)
	}
	switch request.Subsystem {
	case "system":
		if err := s.operateSystem(request); err != nil {
			return err
		}
		if request.Operation == "initialize" {
			service := newFirewallService()
			rulesErr := service.restoreStoredFirewallRules(ctx, filter.Provider(request.Backend), nil)
			whitelistErr := service.SyncPortWhitelist(ctx)
			return errors.Join(rulesErr, whitelistErr)
		}
		return nil
	case "forwarding":
		return s.operateForwarding(request)
	case "docker":
		return s.operateDocker(ctx, request)
	default:
		return fmt.Errorf("unsupported firewall subsystem %q", request.Subsystem)
	}
}

func NewIFirewallSettingService() IFirewallSettingService {
	return &FirewallSettingService{}
}

func (s *FirewallSettingService) operateSystem(request dto.FirewallBackendOperation) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if _, err := lifecycle.NewClient(request.Backend); err != nil {
		return err
	}
	if request.Operation == "cleanup" {
		return cleanupSystemBackend(request.Backend)
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallSystemBackendKey)
	if previous == "" {
		if client, err := lifecycle.NewClient(""); err == nil {
			previous = client.Name()
		}
	}
	if request.Operation == "select" && previous != "" && previous != request.Backend {
		initialized, err := systemFirewallBackendInitialized(previous)
		if err != nil {
			return err
		}
		if initialized {
			return buserr.WithMap("ErrFirewallBackendCleanupRequired", map[string]interface{}{"current": previous, "target": request.Backend}, nil)
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
	initErr := newFirewallService().operateFilterChainBaseLocked(request.Backend, dto.FilterChainOperation{
		Name: constant.FirewallBasicChain, Operate: string(firewall.BaseOperationInit),
	})
	if initErr != nil {
		return rollback(initErr)
	}
	return settingRepo.UpdateOrCreate(constant.FirewallFilterInitializedKey, constant.StatusEnable)
}

func systemFirewallBackendInitialized(backend string) (bool, error) {
	client, err := lifecycle.NewClient(backend)
	if err != nil {
		if errors.Is(err, lifecycle.ErrNotInstalled) {
			return false, nil
		}
		return false, err
	}
	if backend == constant.FirewallProviderIptables || backend == constant.FirewallProviderNftables {
		for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
			initialized, _, err := loadSystemFirewallFamilyStatus(backend, family)
			if family == constant.FirewallFamilyIPv6 && errors.Is(err, filter.ErrFamilyUnavailable) {
				continue
			}
			if err != nil {
				return false, err
			}
			if initialized {
				return true, nil
			}
		}
		return false, nil
	}
	return client.Status()
}

func (s *FirewallSettingService) operateForwarding(request dto.FirewallBackendOperation) error {
	manager, err := newForwardingAdapterFor(request.Backend)
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
	previous, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			detected, err := newForwardingAdapter()
			if err != nil {
				return err
			}
			current = detected.Name()
		}
		initialized, err := forwardingBackendInitialized(current)
		if err != nil {
			return err
		}
		if current != request.Backend && initialized {
			return buserr.WithMap("ErrFirewallBackendCleanupRequired", map[string]interface{}{"current": current, "target": request.Backend}, nil)
		}
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallForwardingBackendKey, request.Backend); err != nil {
		return err
	}
	if request.Operation == "initialize" {
		return newForwardingService().Enable()
	}
	recordForwardingSyncError(nil)
	return nil
}

func forwardingBackendInitialized(backend string) (bool, error) {
	manager, err := newForwardingAdapterFor(backend)
	if err != nil {
		if errors.Is(err, lifecycle.ErrNotInstalled) {
			return false, nil
		}
		return false, err
	}
	for _, family := range []string{constant.FirewallFamilyIPv4, constant.FirewallFamilyIPv6} {
		initialized, _, err := manager.FamilyStatus(family)
		if err != nil {
			return false, err
		}
		if initialized {
			return true, nil
		}
	}
	return false, nil
}

func (s *FirewallSettingService) operateDocker(ctx context.Context, request dto.FirewallBackendOperation) error {
	guard := newDockerFirewallRuntime(request.Backend)
	if request.Operation == "cleanup" {
		if err := guard.Cleanup(); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable)
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			current = constant.FirewallProviderNftables
			if request.Backend == constant.FirewallProviderNftables {
				current = constant.FirewallProviderIptables
			}
		}
		initialized, err := dockerGuardBackendInitialized(current)
		if err != nil {
			return err
		}
		if current != request.Backend && initialized {
			return buserr.WithMap("ErrFirewallBackendCleanupRequired", map[string]interface{}{"current": current, "target": request.Backend}, nil)
		}
	}
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, request.Backend); err != nil {
		return err
	}
	if request.Operation == "select" {
		if err := (&DockerService{}).UpdateFirewallBackend(request.Backend); err != nil {
			_ = settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, previous)
			return err
		}
	}
	if request.Operation == "initialize" {
		if err := newDockerPortGuardService().Operate(ctx, dto.DockerPortGuardOperation{Operation: "initialize"}); err != nil {
			_ = settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, previous)
			return err
		}
	}
	return nil
}

func dockerGuardBackendInitialized(backend string) (bool, error) {
	guard := newDockerFirewallRuntime(backend)
	for _, family := range []string{dockerfirewall.FamilyIPv4, dockerfirewall.FamilyIPv6} {
		initialized, err := guard.Initialized(family)
		if err != nil {
			return false, err
		}
		if initialized {
			return true, nil
		}
	}
	return false, nil
}
