package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"slices"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	filterruntime "github.com/1Panel-dev/1Panel/agent/utils/firewall/filter/runtime"
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
}

type FirewallSettingService struct{}

var firewallWhitelistMu sync.Mutex

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

func (s *FirewallSettingService) CreatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistCreate) error {
	return savePortWhitelist(ctx, func(current []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		return append(current, request.Rule), nil
	})
}

func (s *FirewallSettingService) UpdatePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistUpdate) error {
	return savePortWhitelist(ctx, func(current []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		index, err := findPortWhitelistRule(current, request.OldRule)
		if err != nil {
			return nil, err
		}
		current[index] = request.Rule
		return current, nil
	})
}

func (s *FirewallSettingService) DeletePortWhitelist(ctx context.Context, request dto.FirewallPortWhitelistDelete) error {
	if request.Rule == nil {
		return fmt.Errorf("select one firewall port whitelist rule to delete")
	}
	return savePortWhitelist(ctx, func(current []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		index, err := findPortWhitelistRule(current, *request.Rule)
		if err != nil {
			return nil, err
		}
		return slices.Delete(current, index, index+1), nil
	})
}

func findPortWhitelistRule(rules []firewall.PortWhitelist, target firewall.PortWhitelist) (int, error) {
	index := slices.IndexFunc(rules, func(rule firewall.PortWhitelist) bool {
		return samePortWhitelistRule(rule, target)
	})
	if index < 0 {
		return -1, fmt.Errorf("firewall port whitelist rule has changed or no longer exists; refresh and retry")
	}
	return index, nil
}

func samePortWhitelistRule(left, right firewall.PortWhitelist) bool {
	if reflect.DeepEqual(left, right) {
		return true
	}
	normalizedLeft, err := firewall.ValidatePortWhitelist([]firewall.PortWhitelist{left})
	if err != nil {
		return false
	}
	normalizedRight, err := firewall.ValidatePortWhitelist([]firewall.PortWhitelist{right})
	if err != nil {
		return false
	}
	slices.Sort(normalizedLeft[0].Sources)
	slices.Sort(normalizedRight[0].Sources)
	return reflect.DeepEqual(normalizedLeft[0], normalizedRight[0])
}

func savePortWhitelist(ctx context.Context, change func([]firewall.PortWhitelist) ([]firewall.PortWhitelist, error)) error {
	firewallWhitelistMu.Lock()
	defer firewallWhitelistMu.Unlock()
	firewallRuleMutationMu.Lock()
	defer firewallRuleMutationMu.Unlock()
	defer filterruntime.InvalidateInventory()
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := loadPortWhitelistSetting(tx)
		if err != nil {
			return err
		}
		desired, err := change(current)
		if err != nil {
			return err
		}
		desired, err = firewall.ValidatePortWhitelist(desired)
		if err != nil {
			return err
		}
		value, err := json.Marshal(desired)
		if err != nil {
			return err
		}
		return tx.Where("key = ?", constant.FirewallPortWhiteList).Assign(map[string]interface{}{"value": string(value)}).
			FirstOrCreate(&model.Setting{Key: constant.FirewallPortWhiteList}).Error
	})
}

func checkFirewallRuleWhitelistProtection(provider filter.Provider, record model.FirewallRule) error {
	ports, err := loadFirewallPortWhiteList()
	if err != nil {
		return err
	}
	rules, err := record.RulesForProvider(provider)
	if err != nil {
		return err
	}
	for _, rule := range rules {
		if filter.RuleMatchesPortWhitelist(rule, ports) {
			return filter.ErrProtectedRule
		}
	}
	return nil
}

func loadPortWhitelistSetting(db *gorm.DB) ([]firewall.PortWhitelist, error) {
	var setting model.Setting
	if err := db.Where("key = ?", constant.FirewallPortWhiteList).First(&setting).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		setting.Value = constant.FirewallPortWhiteListValue
	} else if err != nil {
		return nil, err
	}
	var rules []firewall.PortWhitelist
	err := json.Unmarshal([]byte(setting.Value), &rules)
	return rules, err
}

func loadSSHWhitelistPortFrom(path string) (string, error) {
	directives, _, err := parseSSHConfigTree(path)
	if errors.Is(err, os.ErrNotExist) {
		return defaultSSHPort, nil
	}
	if err != nil {
		return "", err
	}
	return loadSSHPortValues(directives)[0], nil
}

func customWhitelist(entries []firewall.PortWhitelist) []firewall.PortWhitelist {
	result := make([]firewall.PortWhitelist, 0, len(entries))
	for _, entry := range entries {
		if entry.Type == "" {
			result = append(result, entry)
		}
	}
	return result
}

func InitializeFirewallWhitelistPorts(entries []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
	entries = slices.Clone(entries)
	var sshPort string
	for i := range entries {
		entry := &entries[i]
		if entry.Type == "" || entry.Port != "" {
			continue
		}
		switch entry.Type {
		case firewall.PortWhitelistTypePanel:
			entry.Port = LoadPanelPort()
		case firewall.PortWhitelistTypeSSH:
			if sshPort == "" {
				var err error
				sshPort, err = loadSSHWhitelistPortFrom(sshPath)
				if err != nil {
					return nil, err
				}
			}
			entry.Port = sshPort
		}
	}
	return firewall.ValidatePortWhitelist(entries)
}

func updateSystemAccessPortWhitelist(ctx context.Context, serviceType string, ports []string) error {
	return savePortWhitelist(ctx, func(entries []firewall.PortWhitelist) ([]firewall.PortWhitelist, error) {
		for i := range entries {
			if entries[i].Type == serviceType {
				if len(ports) == 0 {
					return nil, fmt.Errorf("firewall whitelist %s requires a port", serviceType)
				}
				entries[i].Port = ports[0]
			}
		}
		return entries, nil
	})
}

func (s *FirewallSettingService) Load(ctx context.Context) (dto.FirewallSettings, error) {
	result := dto.FirewallSettings{PingStatus: firewall.LoadPingStatus()}

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
	if err := lockFirewallLifecycleIdle(); err != nil {
		return err
	}
	defer firewallLifecycleTaskMu.Unlock()
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
