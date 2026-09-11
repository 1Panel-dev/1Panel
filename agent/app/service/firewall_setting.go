package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/iptables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/nftables_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/ping"
	"gorm.io/gorm"
)

type IFirewallSettingService interface {
	QueuePortWhitelist(value string) (dto.FilterChainOperationResponse, error)
	Load(context.Context) (dto.FirewallSettings, error)
	Operate(context.Context, dto.FirewallBackendOperation) error
}

type FirewallSettingService struct{}

var firewallWhitelistTaskMu sync.Mutex

var ErrFirewallBackendCleanupRequired = errors.New("firewall backend cleanup required")

func firewallBackendCleanupRequired(current, target string) error {
	return fmt.Errorf(
		"%w: current backend %s still contains 1Panel runtime rules; clean it up before switching to %s",
		ErrFirewallBackendCleanupRequired,
		current,
		target,
	)
}

func NewIFirewallSettingService() IFirewallSettingService {
	return &FirewallSettingService{}
}

func (s *FirewallSettingService) QueuePortWhitelist(value string) (dto.FilterChainOperationResponse, error) {
	return s.queuePortWhitelist(value, newFirewallService())
}

func (s *FirewallSettingService) queuePortWhitelist(value string, firewallService *FirewallService) (dto.FilterChainOperationResponse, error) {
	firewallWhitelistTaskMu.Lock()
	defer firewallWhitelistTaskMu.Unlock()
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem, err := task.NewTask(i18n.GetMsgByKey("FirewallWhitelistTask"), task.TaskUpdate, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		succeeded, failed := 0, 0
		err := s.applyPortWhitelist(t.TaskCtx, value, firewallService, func(status, label string, err error) {
			switch status {
			case "applied":
				succeeded++
				t.LogSuccess(label)
			case "failed":
				failed++
				t.LogFailedWithErr(label, err)
			default:
				t.Log(i18n.GetWithName(status, label))
			}
		})
		t.Log(i18n.GetMsgWithMap("FirewallRuleOperationResult", map[string]interface{}{
			"succeeded": succeeded, "failed": failed,
		}))
		return err
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		closeUnstartedFirewallTask(taskItem)
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall whitelist task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

type whitelistReporter func(status, label string, err error)

func (s *FirewallSettingService) applyPortWhitelist(ctx context.Context, value string, firewallService *FirewallService, report whitelistReporter) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	ports, err := firewall.ParsePortWhitelist(value)
	if err != nil {
		return err
	}
	required, err := firewallService.requiredPorts()
	if err != nil {
		return err
	}
	removed, err := s.savePortWhitelist(ctx, ports, required, firewallService)
	if err != nil {
		return err
	}
	report("FirewallWhitelistSaved", "", nil)
	for _, port := range systemPorts(removed) {
		report("FirewallWhitelistReleased", whitelistPortLabel(port), nil)
	}

	ctx = context.WithValue(ctx, panelPortWhitelistKey{}, required)
	provider, providerErr := firewallService.selectedProvider(ctx)
	ready := s.portWhitelistReadiness(provider, providerErr, firewallService)
	return syncPortWhitelist(ctx, ports, required, ready, firewallService.ensureSystemPortLocked, report)
}

func (s *FirewallSettingService) savePortWhitelist(ctx context.Context, ports, required []firewall.PortWhitelist, firewallService *FirewallService) ([]firewall.PortWhitelist, error) {
	var removed []firewall.PortWhitelist
	err := global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var setting model.Setting
		err := tx.Where("key = ?", constant.FirewallPortWhiteList).First(&setting).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			setting.Value = constant.FirewallPortWhiteListValue
		} else if err != nil {
			return err
		}
		previous, err := firewall.ParsePortWhitelist(setting.Value)
		if err != nil {
			return err
		}
		removed = excludeFirewallPorts(excludeFirewallPorts(previous, ports), required)
		txCtx := context.WithValue(ctx, constant.DB, tx)
		if err := firewallService.releaseSystemPorts(txCtx, systemPorts(removed)); err != nil {
			return err
		}
		value, err := json.Marshal(ports)
		if err != nil {
			return err
		}
		return tx.Where("key = ?", constant.FirewallPortWhiteList).
			Assign(map[string]interface{}{"value": string(value)}).
			FirstOrCreate(&model.Setting{Key: constant.FirewallPortWhiteList}).Error
	})
	return removed, err
}

func (s *FirewallSettingService) portWhitelistReadiness(provider filter.Provider, providerErr error, firewallService *FirewallService) func(dto.FirewallSystemPort) (bool, error) {
	type state struct {
		ready bool
		err   error
	}
	states := make(map[string]state)
	return func(port dto.FirewallSystemPort) (bool, error) {
		if providerErr != nil {
			return false, providerErr
		}
		key := "service"
		if isDirectFirewallProvider(provider) {
			key = port.Family
		}
		if cached, ok := states[key]; ok {
			return cached.ready, cached.err
		}
		var result state
		if isDirectFirewallProvider(provider) {
			initialized, bound, err := loadSystemFirewallFamilyStatus(string(provider), port.Family)
			result = state{ready: initialized && bound, err: err}
		} else {
			client, err := firewallService.baseClient()
			result.err = err
			if err == nil {
				result.ready, result.err = client.Status()
			}
		}
		states[key] = result
		return result.ready, result.err
	}
}

func syncPortWhitelist(
	ctx context.Context,
	ports, required []firewall.PortWhitelist,
	ready func(dto.FirewallSystemPort) (bool, error),
	ensure func(context.Context, dto.FirewallSystemPort) error,
	report whitelistReporter,
) error {
	var failures []error
	for _, port := range systemPorts(ports) {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(failures, err)...)
		}
		label := whitelistPortLabel(port)
		if containsFirewallPort(required, firewall.PortWhitelist{Family: port.Family, Port: port.Port, Protocol: port.Protocol}) {
			report("FirewallWhitelistRequired", label, nil)
			continue
		}
		active, err := ready(port)
		if err == nil && !active {
			report("FirewallWhitelistDeferred", label, nil)
			continue
		}
		if err == nil {
			err = ensure(ctx, port)
		}
		if err != nil {
			report("failed", label, err)
			failures = append(failures, fmt.Errorf("%s: %w", label, err))
			continue
		}
		report("applied", label, nil)
	}
	return errors.Join(failures...)
}

func whitelistPortLabel(port dto.FirewallSystemPort) string {
	return fmt.Sprintf("%s %s/%s", port.Family, port.Port, port.Protocol)
}

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
	result.System.Selected = configuredSystemFirewallBackend()
	if result.System.Selected == "" {
		if client, err := lifecycle.NewClient(); err == nil {
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
		if name == result.System.Selected && name == constant.FirewallProviderIptables {
			if commands, err := lifecycle.ResolveIptablesCommands(); err == nil {
				option.Implementation = commands.IPv4
			}
		}
		result.System.Options = append(result.System.Options, option)
	}

	result.Forwarding.Selected = configuredForwardingBackend()
	result.Forwarding.Current = result.Forwarding.Selected
	for _, name := range []string{constant.FirewallProviderIptables, constant.FirewallProviderNftables} {
		option := dto.FirewallBackendOption{Name: name, Installed: installed[name], Supported: true}
		if option.Installed && name == result.Forwarding.Selected {
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
	result.Docker.Selected = configuredDockerFirewallBackend()
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
			guard := docker_guard.NewRuntime(name)
			ipv4, ipv6 := guard.Status(docker_guard.FamilyIPv4), guard.Status(docker_guard.FamilyIPv6)
			option.Initialized = ipv4.Initialized || ipv6.Initialized
			option.Bound = ipv4.Bound || ipv6.Bound
			option.IPv4.Initialized, option.IPv4.Bound = ipv4.Initialized, ipv4.Bound
			option.IPv6.Initialized, option.IPv6.Bound = ipv6.Initialized, ipv6.Bound
			option.IPv4.Available = ipv4.Reason != docker_guard.ReasonCommandMissing
			option.IPv6.Available = ipv6.Reason != docker_guard.ReasonCommandMissing
			option.IPv4.Reason, option.IPv6.Reason = ipv4.Reason, ipv6.Reason
		}
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
		if err := s.operateSystem(request); err != nil {
			return err
		}
		if request.Operation == "initialize" {
			service := newFirewallService()
			if err := service.restoreStoredFirewallRules(ctx, filter.Provider(request.Backend)); err != nil {
				return err
			}
			return service.syncConfiguredFirewallPorts(ctx)
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

func (s *FirewallSettingService) operateDocker(ctx context.Context, request dto.FirewallBackendOperation) error {
	guard := docker_guard.NewRuntime(request.Backend)
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
			current = alternateDirectBackend(request.Backend)
		}
		initialized, err := dockerGuardBackendInitialized(current)
		if err != nil {
			return err
		}
		if current != request.Backend && initialized {
			return firewallBackendCleanupRequired(current, request.Backend)
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
	guard := docker_guard.NewRuntime(backend)
	for _, family := range []string{docker_guard.FamilyIPv4, docker_guard.FamilyIPv6} {
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

func (s *FirewallSettingService) operateSystem(request dto.FirewallBackendOperation) error {
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
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
	if request.Operation == "select" && previous != "" && previous != request.Backend {
		initialized, err := systemFirewallBackendInitialized(previous)
		if err != nil {
			return err
		}
		if initialized {
			return firewallBackendCleanupRequired(previous, request.Backend)
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
	return systemFirewallBackendInitializedWithClientFactory(backend, lifecycle.NewClientFor)
}

func systemFirewallBackendInitializedWithClientFactory(
	backend string,
	newClient func(string) (lifecycle.Client, error),
) (bool, error) {
	client, err := newClient(backend)
	if err != nil {
		if errors.Is(err, lifecycle.ErrNotInstalled) {
			return false, nil
		}
		return false, err
	}
	if supportsManagedFilterChains(backend) {
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

func cleanupInactiveSystemBackend(backend string) error {
	switch backend {
	case constant.FirewallProviderIptables:
		return (&iptables_helper.Manager{}).Cleanup()
	case constant.FirewallProviderNftables:
		return (&nftables_helper.Manager{}).Cleanup()
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
	previous, _ := settingRepo.GetValueByKey(constant.FirewallForwardingBackendKey)
	if request.Operation == "select" {
		current := previous
		if current == "" {
			detected, err := newForwardingManager()
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
			return firewallBackendCleanupRequired(current, request.Backend)
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
	manager, err := newForwardingManagerFor(backend)
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

func alternateDirectBackend(backend string) string {
	if backend == constant.FirewallProviderNftables {
		return constant.FirewallProviderIptables
	}
	return constant.FirewallProviderNftables
}
