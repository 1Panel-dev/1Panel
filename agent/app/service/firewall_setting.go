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
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/forwarding"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"gorm.io/gorm"
)

type IFirewallSettingService interface {
	CreatePortWhitelist(context.Context, dto.FirewallPortWhitelistCreate) error
	UpdatePortWhitelist(context.Context, dto.FirewallPortWhitelistUpdate) error
	DeletePortWhitelist(context.Context, dto.FirewallPortWhitelistDelete) error
	Load(context.Context) (dto.FirewallSettings, error)
	Operate(context.Context, dto.FirewallBackendOperation) error
	OperateFamily(dto.FirewallFamilyOperation) (dto.FilterChainOperationResponse, error)
	OperateIPv6(dto.FirewallIPv6Operation) (dto.FilterChainOperationResponse, error)
}

type FirewallSettingService struct{}

var firewallWhitelistMu sync.Mutex

func (s *FirewallSettingService) CreatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistCreate) (result error) {
	firewallWhitelistMu.Lock()
	firewallRuleMutationMu.Lock()
	defer func() {
		firewallRuleMutationMu.Unlock()
		firewallWhitelistMu.Unlock()
		if result == nil {
			result = newFirewallService().SyncPortWhitelist(ctx)
		}
	}()
	current, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	current = append(current, request.Rule)
	current, err = firewall.ValidatePortWhitelist(current)
	if err != nil {
		return err
	}
	if err := validateFirewallWhitelistFamilies(current); err != nil {
		return err
	}
	value, err := json.Marshal(current)
	if err != nil {
		return err
	}
	return settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, string(value))
}

func (s *FirewallSettingService) UpdatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistUpdate) (result error) {
	firewallWhitelistMu.Lock()
	firewallRuleMutationMu.Lock()
	defer func() {
		firewallRuleMutationMu.Unlock()
		firewallWhitelistMu.Unlock()
		if result == nil {
			result = newFirewallService().SyncPortWhitelist(ctx)
		}
	}()
	current, err := loadFirewallPortWhiteList()
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
	if err := validateFirewallWhitelistFamilies(current); err != nil {
		return err
	}
	value, err := json.Marshal(current)
	if err != nil {
		return err
	}
	return settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, string(value))
}

func (s *FirewallSettingService) DeletePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistDelete) error {
	firewallWhitelistMu.Lock()
	firewallRuleMutationMu.Lock()
	defer func() {
		firewallRuleMutationMu.Unlock()
		firewallWhitelistMu.Unlock()
	}()
	current, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	if request.Rule == nil {
		return filter.ErrInvalidRule
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
	if err := validateFirewallWhitelistFamilies(current); err != nil {
		return err
	}
	value, err := json.Marshal(current)
	if err != nil {
		return err
	}
	return settingRepo.UpdateOrCreate(constant.FirewallPortWhiteList, string(value))
}

func (s *FirewallSettingService) Load(ctx context.Context) (dto.FirewallSettings, error) {
	families, err := loadFirewallFamilies()
	if err != nil {
		return dto.FirewallSettings{}, err
	}
	result := dto.FirewallSettings{PingStatus: firewall.LoadPingStatus(), IPv6Enabled: slices.Contains(families, constant.FirewallFamilyIPv6)}

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
				overview, err := loadSystemFirewallOverview(name, "base", families)
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
			manager, err := newForwardingAdapterFor(ctx, name)
			if err != nil {
				option.Message = err.Error()
			} else {
				status, statusErr := loadForwardingFirewallOverview(manager, families)
				option.IPv4, option.IPv6 = status.IPv4, status.IPv6
				if statusErr != nil {
					option.Message = statusErr.Error()
				} else {
					option.Initialized, option.Bound = status.IsInit, status.IsBind
				}
				if result.IPv6Enabled && name == constant.FirewallProviderIptables && !option.IPv6.Available {
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
			guard := newDockerFirewallRuntime(ctx, name)
			for _, family := range families {
				status := guard.Status(family)
				option.Initialized = option.Initialized || status.Initialized
				option.Bound = option.Bound || status.Bound
				info := dto.FirewallBackendFamilyStatus{
					Available:   status.Reason != dockerfirewall.ReasonCommandMissing,
					Initialized: status.Initialized, Bound: status.Bound, Reason: status.Reason,
				}
				if family == constant.FirewallFamilyIPv4 {
					option.IPv4 = info
				} else {
					option.IPv6 = info
				}
			}
		}
		result.Docker.Options = append(result.Docker.Options, option)
	}
	result.PortWhitelist, err = loadPortWhitelistSetting()
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
	if request.Operation == "cleanup" {
		_, err := newFirewallService().Reset(ctx, dto.FirewallRuleReset{Subsystem: request.Subsystem, Provider: filter.Provider(request.Backend)})
		return err
	}
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
			whitelistErr := service.SyncPortWhitelist(ctx)
			return whitelistErr
		}
		return nil
	case "forwarding":
		return s.operateForwarding(ctx, request)
	case "docker":
		return s.operateDocker(ctx, request)
	default:
		return fmt.Errorf("unsupported firewall subsystem %q", request.Subsystem)
	}
}

func (s *FirewallSettingService) OperateFamily(request dto.FirewallFamilyOperation) (dto.FilterChainOperationResponse, error) {
	if request.Family != constant.FirewallFamilyIPv4 && request.Family != constant.FirewallFamilyIPv6 {
		return dto.FilterChainOperationResponse{}, filter.ErrInvalidScope
	}
	if request.Backend != constant.FirewallProviderIptables && request.Backend != constant.FirewallProviderNftables {
		return dto.FilterChainOperationResponse{}, filter.ErrUnsupportedScope
	}
	if request.Operation != "initialize" && request.Operation != "repair" && request.Operation != "bind" {
		return dto.FilterChainOperationResponse{}, filter.ErrRuleOperation
	}
	subsystem := ""
	switch request.Subsystem {
	case "system":
		subsystem = firewallTaskHost
	case "forwarding":
		subsystem = firewallTaskForwarding
	case "docker":
		subsystem = firewallTaskDocker
	default:
		return dto.FilterChainOperationResponse{}, filter.ErrInvalidScope
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	return queueFirewallRuleTask(subsystem, task.TaskExec, "", nil, func(t *task.Task) error {
		if err := lockFirewallLifecycleIdle(); err != nil {
			return err
		}
		defer firewallLifecycleTaskMu.Unlock()
		t.Logf("backend=%s family=%s operation=%s", request.Backend, request.Family, request.Operation)
		families, err := loadFirewallFamilies()
		if err != nil {
			return err
		}
		if !slices.Contains(families, request.Family) {
			return fmt.Errorf("IPv6 firewall support is disabled")
		}
		initialize := request.Operation != "bind"
		switch request.Subsystem {
		case "system":
			firewallWhitelistMu.Lock()
			defer firewallWhitelistMu.Unlock()
			if err := newFirewallService().checkSelectedProvider(t.TaskCtx, filter.Provider(request.Backend)); err != nil {
				return err
			}
			ports, err := loadFirewallPortWhiteList()
			if err != nil {
				return err
			}
			required, err := firewall.RequiredPortWhitelist(ports)
			if err != nil {
				return err
			}
			firewallRuleMutationMu.Lock()
			if request.Backend == constant.FirewallProviderIptables {
				err = iptables_helper.OperateFamily(request.Family, initialize, required)
			} else {
				err = nftables_helper.OperateFamily(filter.Family(request.Family), initialize, required)
			}
			firewallRuleMutationMu.Unlock()
			if err != nil {
				return err
			}
			if initialize {
				if err := newFirewallService().applyPortWhitelist(t.TaskCtx, ports, nil, filter.Family(request.Family)); err != nil {
					return err
				}
			}
			return settingRepo.UpdateOrCreate("IptablesStatus", constant.StatusEnable)
		case "forwarding":
			forwardingMutationMu.Lock()
			defer forwardingMutationMu.Unlock()
			manager, err := newForwardingAdapter(t.TaskCtx)
			if err != nil {
				return err
			}
			if manager.Name() != request.Backend {
				return filter.ErrProviderUnavailable
			}
			if err := manager.OperateFamily(request.Family, initialize); err != nil {
				return err
			}
			rules, err := manager.List()
			if err != nil {
				return err
			}
			if err := persistForwardingRules(manager, rules); err != nil {
				return err
			}
			return settingRepo.UpdateOrCreate(constant.FirewallForwardingInitializedKey, constant.StatusEnable)
		default:
			dockerPortGuardServiceMu.Lock()
			defer dockerPortGuardServiceMu.Unlock()
			runtime, backend, err := newDockerPortGuardService().runtimeForDocker(t.TaskCtx)
			if err != nil {
				return err
			}
			if backend != request.Backend {
				return filter.ErrProviderUnavailable
			}
			if err := runtime.OperateFamily(request.Family, initialize); err != nil {
				return err
			}
			inventory, err := runtime.ListPolicies()
			if err != nil {
				return err
			}
			if err := persistDockerRules(backend, inventory); err != nil {
				return err
			}
			return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable)
		}
	})
}

func (s *FirewallSettingService) OperateIPv6(request dto.FirewallIPv6Operation) (dto.FilterChainOperationResponse, error) {
	if request.Status != constant.StatusEnable && request.Status != constant.StatusDisable {
		return dto.FilterChainOperationResponse{}, filter.ErrInvalidRule
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	return queueFirewallRuleTask(firewallTaskHost, task.TaskExec, "", nil, func(t *task.Task) error {
		if err := lockFirewallLifecycleIdle(); err != nil {
			return err
		}
		defer firewallLifecycleTaskMu.Unlock()
		firewallWhitelistMu.Lock()
		defer firewallWhitelistMu.Unlock()
		firewallRuleMutationMu.Lock()
		defer firewallRuleMutationMu.Unlock()
		forwardingMutationMu.Lock()
		defer forwardingMutationMu.Unlock()
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		if request.Status == constant.StatusEnable {
			return settingRepo.UpdateOrCreate(constant.FirewallIPv6SupportKey, request.Status)
		}
		ports, err := loadFirewallPortWhiteList()
		if err != nil {
			return err
		}
		value, err := json.Marshal(ipv4PortWhitelist(ports))
		if err != nil {
			return err
		}
		installed := lifecycle.InstalledProviders()
		for _, selection := range []struct{ subsystem, key string }{
			{"system", constant.FirewallSystemBackendKey},
			{"forwarding", constant.FirewallForwardingBackendKey},
			{"docker", constant.FirewallDockerBackendKey},
		} {
			backend, err := settingRepo.GetValueByKey(selection.key)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
			backend = strings.ToLower(strings.TrimSpace(backend))
			if backend == "" {
				if selection.subsystem == "system" {
					if len(installed) == 0 {
						continue
					}
					client, err := NewSelectedSystemFirewallClient()
					if err != nil {
						return err
					}
					backend = client.Name()
				} else {
					backend = constant.FirewallProviderIptables
				}
			}
			if !slices.Contains(installed, backend) || (backend != constant.FirewallProviderIptables && backend != constant.FirewallProviderNftables) {
				continue
			}
			t.Logf("disable IPv6 firewall bindings: subsystem=%s backend=%s", selection.subsystem, backend)
			switch selection.subsystem {
			case "system":
				if backend == constant.FirewallProviderIptables {
					err = iptables_helper.UnbindIPv6BaseChains()
				} else {
					err = nftables_helper.SetTableDormant(t.TaskCtx, "ip6", nftables_helper.TableName)
					if err == nil {
						err = nftables_helper.PersistRuleset(t.TaskCtx)
					}
				}
			case "forwarding":
				var manager forwarding.Adapter
				manager, err = newForwardingAdapterFor(t.TaskCtx, backend)
				if err == nil {
					err = manager.UnbindFamily(constant.FirewallFamilyIPv6)
				}
			case "docker":
				err = newDockerFirewallRuntime(t.TaskCtx, backend).Unbind(constant.FirewallFamilyIPv6)
			}
			if err != nil && !errors.Is(err, filter.ErrFamilyUnavailable) {
				return err
			}
		}
		if err := t.TaskCtx.Err(); err != nil {
			return err
		}
		return settingRepo.UpdateValues(map[string]string{constant.FirewallIPv6SupportKey: request.Status, constant.FirewallPortWhiteList: string(value)})
	})
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

func (s *FirewallSettingService) operateForwarding(ctx context.Context, request dto.FirewallBackendOperation) error {
	if _, err := newForwardingAdapterFor(ctx, request.Backend); err != nil {
		return err
	}
	previous, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			detected, err := newForwardingAdapter(ctx)
			if err != nil {
				return err
			}
			current = detected.Name()
		}
		initialized, err := forwardingBackendInitialized(ctx, current)
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
		return newForwardingService().Enable(ctx)
	}
	return nil
}

func forwardingBackendInitialized(ctx context.Context, backend string) (bool, error) {
	manager, err := newForwardingAdapterFor(ctx, backend)
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

	previous, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			current = constant.FirewallProviderNftables
			if request.Backend == constant.FirewallProviderNftables {
				current = constant.FirewallProviderIptables
			}
		}
		initialized, err := dockerGuardBackendInitialized(ctx, current)
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

func dockerGuardBackendInitialized(ctx context.Context, backend string) (bool, error) {
	guard := newDockerFirewallRuntime(ctx, backend)
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
