package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
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
	policies          repo.IDockerPortGuardRepo
	runtime           dockerfirewall.Runtime
	runtimeForBackend func(string) dockerfirewall.Runtime
	client            func() (*client.Client, error)
	version           func(string) string
}

var dockerPortGuardServiceMu sync.Mutex

type IDockerPortGuardService interface {
	LoadOverview(context.Context) (dto.DockerPortGuardList, error)
	LoadPublishedPorts(context.Context) ([]dto.DockerPortGuardContainer, error)
	Operate(context.Context, dto.DockerPortGuardOperation) error
	QueueInitialization(dto.DockerPortGuardOperation) (dto.FilterChainOperationResponse, error)
	DeletePolicies(dto.DockerPortGuardPolicyBatchDelete) (dto.FilterChainOperationResponse, error)
	UpsertPolicies(dto.DockerPortGuardPolicyBatch) (dto.FilterChainOperationResponse, error)
	Reconcile(context.Context) error
}

func (s *DockerPortGuardService) LoadOverview(ctx context.Context) (dto.DockerPortGuardList, error) {
	policies, err := s.policies.ListManaged(ctx)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	unavailable := func() dto.DockerPortGuardList {
		backend := selectedDockerFirewallBackend("")
		base := s.runtimeStatus(s.guardRuntime(backend), backend)
		base.Message = i18n.Get("ErrDockerFailed")
		return dto.DockerPortGuardList{Base: base, Containers: []dto.DockerPortGuardContainer{}, OrphanPolicies: dockerGuardPolicyEndpoints(policies)}
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
	backend := selectedDockerFirewallBackend(detectedBackend)
	base := s.runtimeStatus(s.guardRuntime(backend), backend)
	endpoints, err := discoverDockerEndpoints(ctx, cli, true)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	annotateDockerEndpointManagement(endpoints, detectedBackend)
	endpoints, orphanPolicies := matchDockerGuardPolicies(base, policies, endpoints)
	sort.Slice(endpoints, func(i, j int) bool {
		return fmt.Sprintf("%s|%s|%d|%s", endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol) < fmt.Sprintf("%s|%s|%d|%s", endpoints[j].Family, endpoints[j].HostIP, endpoints[j].HostPort, endpoints[j].Protocol)
	})
	sort.Slice(orphanPolicies, func(i, j int) bool {
		return fmt.Sprintf("%s|%s|%d|%s", orphanPolicies[i].Family, orphanPolicies[i].HostIP, orphanPolicies[i].HostPort, orphanPolicies[i].Protocol) < fmt.Sprintf("%s|%s|%d|%s", orphanPolicies[j].Family, orphanPolicies[j].HostIP, orphanPolicies[j].HostPort, orphanPolicies[j].Protocol)
	})
	return dto.DockerPortGuardList{Base: base, Containers: groupDockerGuardContainers(endpoints), OrphanPolicies: orphanPolicies}, nil
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
	annotateDockerEndpointManagement(endpoints, backend)
	return groupDockerGuardContainers(endpoints), nil
}

func (s *DockerPortGuardService) Operate(ctx context.Context, request dto.DockerPortGuardOperation) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	switch request.Operation {
	case "initialize":
		runtime, backend, err := s.runtimeForDocker(ctx)
		if err != nil {
			return err
		}
		policies, err := s.runtimePolicies(ctx)
		if err != nil {
			return err
		}
		inventory, err := runtime.ListPolicies()
		if err != nil {
			return err
		}
		if err := runtime.Initialize(policies, inventory); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, backend); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable); err != nil {
			return err
		}
		return nil
	case "bind":
		runtime, _, err := s.runtimeForDocker(ctx)
		if err != nil {
			return err
		}
		if err := runtime.Bind(); err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable)
	case "unbind":
		var err error
		if s.runtime != nil {
			err = s.runtime.Unbind()
		} else {
			err = errors.Join(dockerfirewall.NewIptables().Unbind(), dockerfirewall.NewNftables().Unbind())
		}
		if err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable)
	default:
		return fmt.Errorf("unsupported Docker port guard operation: %s", request.Operation)
	}
}

func (s *DockerPortGuardService) QueueInitialization(request dto.DockerPortGuardOperation) (dto.FilterChainOperationResponse, error) {
	if request.Operation != "initialize" {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("only Docker port guard initialization can be queued")
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem, err := task.NewTask(firewallTaskName(task.TaskExec, firewallTaskDocker, ""), task.TaskExec, task.TaskScopeFirewall, request.TaskID, 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("create Docker port guard initialization task: %w", err)
	}
	var runtime dockerfirewall.Runtime
	var backend string
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallInspectDockerGuardStep"), func(t *task.Task) error {
		var err error
		runtime, backend, err = s.runtimeForDocker(t.TaskCtx)
		if err != nil {
			return err
		}
		t.Logf("backend=%s", backend)
		return nil
	}, nil)
	taskItem.AddSubTask(i18n.GetWithName("FirewallInitializeDockerGuardStep", "Docker"), func(t *task.Task) error {
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		policies, err := s.runtimePolicies(t.TaskCtx)
		if err != nil {
			return err
		}
		t.Logf("backend=%s", backend)
		inventory, err := runtime.ListPolicies()
		if err != nil {
			return err
		}
		return runtime.Initialize(policies, inventory)
	}, nil)
	taskItem.AddSubTask(i18n.GetMsgByKey("FirewallPersistDockerGuardStep"), func(t *task.Task) error {
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, backend); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable); err != nil {
			return err
		}
		return nil
	}, nil)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save Docker port guard initialization task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *DockerPortGuardService) DeletePolicies(request dto.DockerPortGuardPolicyBatchDelete) (dto.FilterChainOperationResponse, error) {
	uuids, err := normalizeDockerFirewallUUIDs(request.UUIDs)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	labels := make([]string, len(uuids))
	for i, id := range uuids {
		labels[i] = fmt.Sprintf("[%d/%d] %s", i+1, len(uuids), id)
	}
	return queueFirewallRuleTask(firewallTaskDocker, task.TaskDelete, labels, func(ctx context.Context) error {
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.policies.DeleteBatch(ctx, uuids); err != nil {
			return err
		}
		return s.reconcileLocked(ctx)
	})
}

func (s *DockerPortGuardService) UpsertPolicies(request dto.DockerPortGuardPolicyBatch) (dto.FilterChainOperationResponse, error) {
	if len(request.Policies) > filter.MaxAtomicExpansion {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("create or import at most %d rules per batch (after expansion)", filter.MaxAtomicExpansion)
	}
	labels := make([]string, len(request.Policies))
	policies := make([]model.DockerPortGuardPolicy, 0, len(request.Policies))
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
		if normalized.Mode == dockerfirewall.ModeAll {
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
		encoded, err := json.Marshal(normalized.Sources)
		if err != nil {
			return dto.FilterChainOperationResponse{}, fmt.Errorf("%s: %w", labels[i], err)
		}
		policies = append(policies, model.DockerPortGuardPolicy{
			UUID: uuid.NewString(), Family: normalized.Family, HostIP: normalized.HostIP,
			HostPort: normalized.HostPort, Protocol: normalized.Protocol, Mode: normalized.Mode,
			Sources: string(encoded), Description: strings.TrimSpace(policy.Description),
		})
		endpoints = append(endpoints, dto.DockerPortGuardEndpointIdentity{
			Family: normalized.Family, HostIP: normalized.HostIP, HostPort: normalized.HostPort, Protocol: normalized.Protocol,
		})
	}
	return queueFirewallRuleTask(firewallTaskDocker, task.TaskUpdate, labels, func(ctx context.Context) error {
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := s.rejectHostInputDockerGuardEndpoints(ctx, endpoints); err != nil {
			return err
		}
		if err := s.policies.UpsertBatch(ctx, policies); err != nil {
			return err
		}
		return s.reconcileLocked(ctx)
	})
}

func NewIDockerPortGuardService() IDockerPortGuardService {
	return newDockerPortGuardService()
}

func (s *DockerPortGuardService) runtimeStatus(runtime dockerfirewall.Runtime, backend string) dto.DockerPortGuardBase {
	ipv4 := runtime.Status(dockerfirewall.FamilyIPv4)
	ipv6 := runtime.Status(dockerfirewall.FamilyIPv6)
	version := "-"
	if s.version != nil {
		version = s.version(backend)
	}
	name := "iptables-docker"
	if strings.EqualFold(strings.TrimSpace(backend), constant.FirewallProviderNftables) {
		name = "nftables-docker"
	}
	return dto.DockerPortGuardBase{
		Name:        name,
		Version:     version,
		Backend:     backend,
		IsExist:     ipv4.Reason != dockerfirewall.ReasonCommandMissing || ipv6.Reason != dockerfirewall.ReasonCommandMissing,
		Initialized: ipv4.Initialized || ipv6.Initialized,
		Bound:       ipv4.Bound || ipv6.Bound,
		IPv4:        dto.DockerPortGuardFamilyStatus{State: ipv4.State, Reason: ipv4.Reason, Initialized: ipv4.Initialized, Bound: ipv4.Bound, Effective: ipv4.Effective},
		IPv6:        dto.DockerPortGuardFamilyStatus{State: ipv6.State, Reason: ipv6.Reason, Initialized: ipv6.Initialized, Bound: ipv6.Bound, Effective: ipv6.Effective},
	}
}

func dockerGuardPolicyEndpoints(policies []model.DockerPortGuardPolicy) []dto.DockerPortGuardEndpoint {
	endpoints := make([]dto.DockerPortGuardEndpoint, 0, len(policies))
	for _, policy := range policies {
		sources := []string{}
		_ = json.Unmarshal([]byte(policy.Sources), &sources)
		endpoints = append(endpoints, dto.DockerPortGuardEndpoint{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
			PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: sources,
			Description: policy.Description, TrafficPath: dockerTrafficPathUnknown,
			ManagementTarget: dockerManagementNeedsDiagnosis, ManagementReason: dockerReasonNoMatchingPath,
		})
	}
	return endpoints
}

func matchDockerGuardPolicies(base dto.DockerPortGuardBase, policies []model.DockerPortGuardPolicy, endpoints []dto.DockerPortGuardEndpoint) ([]dto.DockerPortGuardEndpoint, []dto.DockerPortGuardEndpoint) {
	byEndpoint := make(map[string]model.DockerPortGuardPolicy, len(policies))
	for _, policy := range policies {
		byEndpoint[fmt.Sprintf("%s|%s|%d|%s", policy.Family, policy.HostIP, policy.HostPort, policy.Protocol)] = policy
	}
	for i := range endpoints {
		key := fmt.Sprintf("%s|%s|%d|%s", endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol)
		policy, ok := byEndpoint[key]
		if !ok {
			continue
		}
		sources := []string{}
		_ = json.Unmarshal([]byte(policy.Sources), &sources)
		endpoints[i].PolicyUUID, endpoints[i].Mode, endpoints[i].Sources = policy.UUID, policy.Mode, sources
		endpoints[i].Description = policy.Description
		endpoints[i].Effective = endpoints[i].ManagementTarget == dockerManagementContainerGuard &&
			((policy.Family == dockerfirewall.FamilyIPv4 && base.IPv4.Effective) || (policy.Family == dockerfirewall.FamilyIPv6 && base.IPv6.Effective))
		delete(byEndpoint, key)
	}
	orphanPolicies := make([]dto.DockerPortGuardEndpoint, 0, len(byEndpoint))
	for _, policy := range byEndpoint {
		sources := []string{}
		_ = json.Unmarshal([]byte(policy.Sources), &sources)
		orphanPolicies = append(orphanPolicies, dto.DockerPortGuardEndpoint{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
			PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: sources, Description: policy.Description,
			TrafficPath: dockerTrafficPathUnknown, ManagementTarget: dockerManagementNeedsDiagnosis,
			ManagementReason: dockerReasonNoMatchingPath,
		})
	}
	return endpoints, orphanPolicies
}

func (s *DockerPortGuardService) rejectHostInputDockerGuardEndpoints(ctx context.Context, requested []dto.DockerPortGuardEndpointIdentity) error {
	if s.client == nil || len(requested) == 0 {
		return nil
	}
	cli, err := s.client()
	if err != nil {
		return nil
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return nil
	}
	endpoints, err := discoverDockerEndpoints(ctx, cli, true)
	if err != nil {
		return nil
	}
	annotateDockerEndpointManagement(endpoints, dockerFirewallBackend(info))
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
