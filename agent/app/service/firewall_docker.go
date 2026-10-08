package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	dockerfirewall "github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

const (
	dockerGuardComposeProjectLabel  = "com.docker.compose.project"
	dockerGuardComposeCreatedBy     = "createdBy"
	dockerTrafficPathForward        = "forward"
	dockerTrafficPathInput          = "input"
	dockerTrafficPathUnknown        = "unknown"
	dockerManagementContainerGuard  = "container_guard"
	dockerManagementHostFirewall    = "host_firewall"
	dockerManagementNeedsDiagnosis  = "needs_diagnosis"
	dockerReasonNATInspectFailed    = "nat_inspect_failed"
	dockerReasonNATChainUnreachable = "nat_chain_unreachable"
	dockerReasonProxyInspectFailed  = "proxy_inspect_failed"
	dockerReasonNoMatchingPath      = "no_matching_path"
)

type DockerPortGuardService struct {
	runtime           dockerfirewall.Runtime
	runtimeForBackend func(context.Context, string) dockerfirewall.Runtime
	client            func() (*client.Client, error)
	version           func(string) string
}

var dockerPortGuardServiceMu sync.Mutex

type IDockerPortGuardService interface {
	LoadOverview(context.Context) (dto.DockerPortGuardList, error)
	ExportBackup(context.Context, filter.Provider) (dto.FirewallSubsystemBackup, error)
	LoadPublishedPorts(context.Context) ([]dto.DockerPortGuardContainer, error)
	Operate(context.Context, dto.DockerPortGuardOperation) error
	QueueInitialization(dto.DockerPortGuardOperation) (dto.FilterChainOperationResponse, error)
	DeletePolicies(dto.DockerPortGuardPolicyBatchDelete) (dto.FilterChainOperationResponse, error)
	UpsertPolicies(dto.DockerPortGuardPolicyBatch) (dto.FilterChainOperationResponse, error)
	Restore(context.Context) error
}

func NewIDockerPortGuardService() IDockerPortGuardService {
	return newDockerPortGuardService()
}

func (s *DockerPortGuardService) LoadOverview(ctx context.Context) (dto.DockerPortGuardList, error) {
	families, err := loadFirewallFamilies()
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	backend := selectedDockerFirewallBackend("")
	inventory, err := s.guardRuntime(ctx, backend).ListPolicies()
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	policies := dockerGuardInventoryEndpoints(inventory)
	unavailable := func() dto.DockerPortGuardList {
		backend := selectedDockerFirewallBackend("")
		base := s.runtimeStatus(s.guardRuntime(ctx, backend), backend, len(families) > 1)
		base.Message = i18n.Get("ErrDockerFailed")
		return dto.DockerPortGuardList{Base: base, Containers: []dto.DockerPortGuardContainer{}, OrphanPolicies: policies}
	}
	cli, err := s.client()
	if err != nil {
		return unavailable(), nil
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return unavailable(), nil
	}
	detectedBackend := dockerFirewallBackend(info)
	backend = selectedDockerFirewallBackend(detectedBackend)
	base := s.runtimeStatus(s.guardRuntime(ctx, backend), backend, len(families) > 1)
	endpoints, err := discoverDockerEndpoints(ctx, cli, true)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	annotateDockerEndpointManagement(ctx, endpoints, detectedBackend)
	endpoints, orphanPolicies := matchDockerGuardPolicies(base, policies, endpoints)
	sort.Slice(endpoints, func(i, j int) bool {
		return fmt.Sprintf("%s|%s|%d|%s", endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol) < fmt.Sprintf("%s|%s|%d|%s", endpoints[j].Family, endpoints[j].HostIP, endpoints[j].HostPort, endpoints[j].Protocol)
	})
	sort.Slice(orphanPolicies, func(i, j int) bool {
		return fmt.Sprintf("%s|%s|%d|%s", orphanPolicies[i].Family, orphanPolicies[i].HostIP, orphanPolicies[i].HostPort, orphanPolicies[i].Protocol) < fmt.Sprintf("%s|%s|%d|%s", orphanPolicies[j].Family, orphanPolicies[j].HostIP, orphanPolicies[j].HostPort, orphanPolicies[j].Protocol)
	})
	return dto.DockerPortGuardList{Base: base, Containers: groupDockerGuardContainers(endpoints), OrphanPolicies: orphanPolicies}, nil
}

func (s *DockerPortGuardService) ExportBackup(ctx context.Context, provider filter.Provider) (dto.FirewallSubsystemBackup, error) {
	backend := string(provider)
	if backend == "" {
		backend = selectedDockerFirewallBackend("")
	}
	if backend != constant.FirewallProviderIptables && backend != constant.FirewallProviderNftables {
		return dto.FirewallSubsystemBackup{}, filter.ErrInvalidRule
	}
	inventory, err := s.guardRuntime(ctx, backend).ListPolicies()
	if err != nil {
		return dto.FirewallSubsystemBackup{}, err
	}
	if inventory.Policies == nil {
		inventory.Policies = []dockerfirewall.Policy{}
	}
	return dto.FirewallSubsystemBackup{Subsystem: "docker", Provider: filter.Provider(backend), Docker: &inventory}, nil
}

func (s *DockerPortGuardService) LoadPublishedPorts(ctx context.Context) ([]dto.DockerPortGuardContainer, error) {
	cli, err := s.client()
	if err != nil {
		return nil, buserr.WithDetail("ErrDockerFailed", err.Error(), err)
	}
	defer cli.Close()

	if socketPath, local := strings.CutPrefix(cli.DaemonHost(), "unix://"); local {
		if _, statErr := os.Stat(socketPath); errors.Is(statErr, os.ErrNotExist) {
			return []dto.DockerPortGuardContainer{}, nil
		}
	}

	endpoints, err := discoverDockerEndpoints(ctx, cli, false)
	if err != nil {
		return nil, err
	}
	backend := selectedDockerFirewallBackend("")
	if info, infoErr := cli.Info(ctx); infoErr == nil {
		backend = dockerFirewallBackend(info)
	}
	annotateDockerEndpointManagement(ctx, endpoints, backend)
	return groupDockerGuardContainers(endpoints), nil
}

func (s *DockerPortGuardService) Operate(ctx context.Context, request dto.DockerPortGuardOperation) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	switch request.Operation {
	case "initialize":
		return s.initialize(ctx, request, nil)
	case "bind":
		runtime, _, err := s.runtimeForDocker(ctx)
		if err != nil {
			return err
		}
		families, err := loadFirewallFamilies()
		if err != nil {
			return err
		}
		if err := errors.Join(runtime.Bind(families...), ctx.Err()); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable)
	case "unbind":
		var err error
		if s.runtime != nil {
			err = s.runtime.Unbind()
		} else {
			err = errors.Join(dockerfirewall.NewIptables(ctx).Unbind(), dockerfirewall.NewNftables(ctx).Unbind())
		}
		if err = errors.Join(err, ctx.Err()); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable)
	default:
		return fmt.Errorf("unsupported Docker port guard operation: %s", request.Operation)
	}
}

func (s *DockerPortGuardService) QueueInitialization(request dto.DockerPortGuardOperation) (dto.FilterChainOperationResponse, error) {
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	if request.Operation != "initialize" {
		return dto.FilterChainOperationResponse{}, filter.ErrInvalidRule
	}
	if request.BackupFile != "" {
		if _, err := readFirewallSubsystemBackup(request.BackupFile, "docker"); err != nil {
			return dto.FilterChainOperationResponse{}, err
		}
	}
	return queueFirewallRuleTask(firewallTaskDocker, task.TaskExec, request.TaskID, []string{firewallTaskDocker}, func(t *task.Task) error {
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		return s.initialize(t.TaskCtx, request, t)
	})
}

func (s *DockerPortGuardService) DeletePolicies(request dto.DockerPortGuardPolicyBatchDelete) (dto.FilterChainOperationResponse, error) {
	uuids, err := normalizeDockerFirewallUUIDs(request.UUIDs)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	return queueFirewallRuleTask(firewallTaskDocker, task.TaskDelete, "", uuids, func(t *task.Task) error {
		ctx := t.TaskCtx
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		runtime, backend, err := s.runtimeForDocker(ctx)
		if err != nil {
			return err
		}
		inventory, err := runtime.ListPolicies()
		if err != nil {
			return err
		}
		wanted := make(map[string]bool, len(uuids))
		for _, id := range uuids {
			wanted[id] = true
		}
		remaining := make([]dockerfirewall.Policy, 0, len(inventory.Policies))
		for _, policy := range inventory.Policies {
			if wanted[policy.UUID] {
				delete(wanted, policy.UUID)
			} else {
				remaining = append(remaining, policy)
			}
		}
		if len(wanted) > 0 {
			return filter.ErrRuleStale
		}
		return applyDockerPolicies(ctx, runtime, backend, inventory, remaining)
	})
}

func (s *DockerPortGuardService) UpsertPolicies(request dto.DockerPortGuardPolicyBatch) (dto.FilterChainOperationResponse, error) {
	if len(request.Policies) > filter.MaxAtomicExpansion {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
	}
	labels := make([]string, len(request.Policies))
	policies := make([]dockerfirewall.Policy, 0, len(request.Policies))
	endpoints := make([]dto.DockerPortGuardEndpointIdentity, 0, len(request.Policies))
	count := 0
	for i, policy := range request.Policies {
		labels[i] = fmt.Sprintf("[%d/%d] %s %s %s:%d %s", i+1, len(request.Policies), policy.Family, policy.Protocol, policy.HostIP, policy.HostPort, policy.Mode)
		normalized, err := normalizeDockerFirewallPolicy(dockerfirewall.Policy{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort,
			Protocol: policy.Protocol, Mode: policy.Mode, Sources: policy.Sources,
		})
		if err != nil {
			return dto.FilterChainOperationResponse{}, fmt.Errorf("%s: %w", labels[i], err)
		}
		if normalized.Mode == dockerfirewall.ModeAll || normalized.Mode == dockerfirewall.ModeAcceptAll {
			count++
		} else {
			count += len(normalized.Sources)
			if normalized.Mode == dockerfirewall.ModeAllow {
				count++
			}
		}
		if count > filter.MaxAtomicExpansion {
			return dto.FilterChainOperationResponse{}, fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
		}
		normalized.UUID = uuid.NewString()
		policies = append(policies, normalized)
	}
	return queueFirewallRuleTask(firewallTaskDocker, task.TaskUpdate, "", labels, func(t *task.Task) error {
		ctx := t.TaskCtx
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		if err := ctx.Err(); err != nil {
			return err
		}
		runtime, backend, err := s.runtimeForDocker(ctx)
		if err != nil {
			return err
		}
		inventory, err := runtime.ListPolicies()
		if err != nil {
			return err
		}
		current := append([]dockerfirewall.Policy(nil), inventory.Policies...)
		if request.Import {
			backup := dto.FirewallSubsystemBackup{Provider: filter.Provider(backend), Docker: &dockerfirewall.PolicyInventory{Policies: policies}}
			merged, err := mergeDockerBackup(inventory, backup, backend, t)
			if err != nil {
				return err
			}
			current = merged.Policies
		}
		byEndpoint := make(map[string]int, len(current))
		for i, policy := range current {
			byEndpoint[dockerPolicyEndpointKey(policy)] = i
		}
		for i := range policies {
			key := dockerPolicyEndpointKey(policies[i])
			index, exists := byEndpoint[key]
			if request.Import {
				if !exists || current[index].UUID != policies[i].UUID {
					labels[i] = ""
					continue
				}
			} else if exists {
				policies[i].UUID = current[index].UUID
				current[index] = policies[i]
			} else {
				byEndpoint[key] = len(current)
				current = append(current, policies[i])
			}
			endpoints = append(endpoints, dto.DockerPortGuardEndpointIdentity{
				Family: policies[i].Family, HostIP: policies[i].HostIP, HostPort: policies[i].HostPort, Protocol: policies[i].Protocol,
			})
		}
		if len(endpoints) == 0 {
			return nil
		}
		if err := s.rejectHostInputDockerGuardEndpoints(ctx, endpoints); err != nil {
			return err
		}
		if err := applyDockerPolicies(ctx, runtime, backend, inventory, current); err != nil {
			return err
		}
		return nil
	})
}

func (s *DockerPortGuardService) Restore(ctx context.Context) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	enabled, err := dockerPortGuardPersistedEnabled()
	if err != nil || !enabled {
		return err
	}
	runtime, backend, err := s.runtimeForDocker(ctx)
	if err != nil {
		return err
	}
	backup, err := readFirewallSubsystemBackup("docker-"+backend+".rules", "docker")
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	families, err := loadFirewallFamilies()
	if err != nil {
		return err
	}
	if !slices.Contains(families, constant.FirewallFamilyIPv6) {
		before := len(backup.Docker.Policies)
		backup.Docker.Policies = slices.DeleteFunc(backup.Docker.Policies, func(policy dockerfirewall.Policy) bool { return policy.Family == constant.FirewallFamilyIPv6 })
		logFirewallIPv6Skipped(nil, "Docker startup", before-len(backup.Docker.Policies))
	}
	missing := make(map[string]bool)
	needsBind := false
	for _, family := range families {
		status := runtime.Status(family)
		if status.Reason == dockerfirewall.ReasonInspectFailed {
			return fmt.Errorf("inspect Docker guard %s failed", family)
		}
		if status.Reason == dockerfirewall.ReasonCommandMissing {
			continue
		}
		missing[family] = !status.Initialized
		needsBind = needsBind || (status.Initialized && !status.Effective)
	}
	if !missing[dockerfirewall.FamilyIPv4] && !missing[dockerfirewall.FamilyIPv6] {
		if needsBind {
			return runtime.Bind(families...)
		}
		return nil
	}
	inventory, err := runtime.ListPolicies()
	if err != nil {
		return err
	}
	for _, policy := range inventory.Policies {
		missing[policy.Family] = false
	}
	if inventory.RuleOrders == nil {
		inventory.RuleOrders = make(map[string][]int64)
	}
	for _, policy := range backup.Docker.Policies {
		if !missing[policy.Family] {
			continue
		}
		inventory.Policies = append(inventory.Policies, policy)
		key := policy.Family + "\x00" + policy.UUID
		inventory.RuleOrders[key] = backup.Docker.RuleOrders[key]
	}
	return runtime.Initialize(inventory.Policies, inventory, families...)
}

func (s *DockerPortGuardService) runtimeStatus(runtime dockerfirewall.Runtime, backend string, ipv6Enabled bool) dto.DockerPortGuardBase {
	ipv4 := runtime.Status(dockerfirewall.FamilyIPv4)
	var ipv6 dockerfirewall.FamilyStatus
	if ipv6Enabled {
		ipv6 = runtime.Status(dockerfirewall.FamilyIPv6)
	}
	version := "-"
	if s.version != nil {
		version = s.version(backend)
	}
	name := "iptables-docker"
	if strings.EqualFold(strings.TrimSpace(backend), constant.FirewallProviderNftables) {
		name = "nftables-docker"
	}
	return dto.DockerPortGuardBase{
		IPv6Enabled: ipv6Enabled,
		Name:        name,
		Version:     version,
		Backend:     backend,
		IsExist:     ipv4.Reason != dockerfirewall.ReasonCommandMissing || (ipv6Enabled && ipv6.Reason != dockerfirewall.ReasonCommandMissing),
		Initialized: ipv4.Initialized || ipv6.Initialized,
		Bound:       ipv4.Bound || ipv6.Bound,
		IPv4:        dto.DockerPortGuardFamilyStatus{Partial: ipv4.Partial, State: ipv4.State, Reason: ipv4.Reason, Initialized: ipv4.Initialized, Bound: ipv4.Bound, Effective: ipv4.Effective},
		IPv6:        dto.DockerPortGuardFamilyStatus{Partial: ipv6.Partial, State: ipv6.State, Reason: ipv6.Reason, Initialized: ipv6.Initialized, Bound: ipv6.Bound, Effective: ipv6.Effective},
	}
}

func matchDockerGuardPolicies(base dto.DockerPortGuardBase, policies []dto.DockerPortGuardEndpoint, endpoints []dto.DockerPortGuardEndpoint) ([]dto.DockerPortGuardEndpoint, []dto.DockerPortGuardEndpoint) {
	matched := make(map[int]bool, len(policies))
	byEndpoint := make(map[string]int, len(policies))
	for index, policy := range policies {
		key := strings.Join([]string{policy.Family, policy.HostIP, strconv.Itoa(int(policy.HostPort)), policy.Protocol}, "\x00")
		if _, exists := byEndpoint[key]; !exists {
			byEndpoint[key] = index
		}
	}
	for i := range endpoints {
		key := strings.Join([]string{endpoints[i].Family, endpoints[i].HostIP, strconv.Itoa(int(endpoints[i].HostPort)), endpoints[i].Protocol}, "\x00")
		index, exists := byEndpoint[key]
		if !exists {
			continue
		}
		policy := policies[index]
		endpoints[i].PolicyUUID, endpoints[i].Mode, endpoints[i].Sources = policy.PolicyUUID, policy.Mode, policy.Sources
		endpoints[i].Effective = endpoints[i].ManagementTarget == dockerManagementContainerGuard && ((policy.Family == dockerfirewall.FamilyIPv4 && base.IPv4.Effective) || (policy.Family == dockerfirewall.FamilyIPv6 && base.IPv6.Effective))
		matched[index] = true
	}
	orphans := make([]dto.DockerPortGuardEndpoint, 0)
	for i, policy := range policies {
		if !matched[i] {
			orphans = append(orphans, policy)
		}
	}
	return endpoints, orphans
}

func (s *DockerPortGuardService) rejectHostInputDockerGuardEndpoints(ctx context.Context, requested []dto.DockerPortGuardEndpointIdentity) error {
	if s.client == nil || len(requested) == 0 {
		return ctx.Err()
	}
	cli, err := s.client()
	if err != nil {
		return ctx.Err()
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return ctx.Err()
	}
	endpoints, err := discoverDockerEndpoints(ctx, cli, true)
	if err != nil {
		return ctx.Err()
	}
	annotateDockerEndpointManagement(ctx, endpoints, dockerFirewallBackend(info))
	if err := ctx.Err(); err != nil {
		return err
	}
	targets := make(map[string]string, len(endpoints))
	for _, endpoint := range endpoints {
		targets[fmt.Sprintf("%s|%s|%d|%s", endpoint.Family, endpoint.HostIP, endpoint.HostPort, endpoint.Protocol)] = endpoint.ManagementTarget
	}
	for _, endpoint := range requested {
		target := targets[fmt.Sprintf("%s|%s|%d|%s", endpoint.Family, endpoint.HostIP, endpoint.HostPort, endpoint.Protocol)]
		if target == dockerManagementHostFirewall {
			return buserr.WithDetail("ErrInvalidParams", "endpoint traffic is handled by the host input firewall", nil)
		}
		if target == dockerManagementNeedsDiagnosis {
			return buserr.WithDetail("ErrInvalidParams", "endpoint traffic management target requires diagnosis", nil)
		}
	}
	return nil
}
