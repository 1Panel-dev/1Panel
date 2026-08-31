package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	agenti18n "github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	dockerGuardComposeProjectLabel = "com.docker.compose.project"
	dockerGuardComposeCreatedBy    = "createdBy"
	dockerTrafficPathForward       = "forward"
	dockerTrafficPathInput         = "input"
	dockerTrafficPathUnknown       = "unknown"
)

type dockerProxyEndpoint struct {
	protocol string
	hostIP   string
	hostPort uint16
}

type dockerForwardRules struct {
	output    string
	inspected bool
}

type dockerGuardRuntime = docker_guard.Runtime

type DockerPortGuardService struct {
	policies          repo.IDockerPortGuardRepo
	runtime           dockerGuardRuntime
	runtimeForBackend func(string) dockerGuardRuntime
	client            func() (*client.Client, error)
	version           func(string) string
}

var (
	dockerPortGuardServiceMu          sync.Mutex
	dockerPortGuardSyncMu             sync.RWMutex
	dockerPortGuardSyncErr            error
	ErrDockerGuardInvalid             = docker_guard.ErrInvalidPolicy
	ErrDockerUnavailable              = docker.ErrUnavailable
	ErrDockerIptablesChainUnavailable = docker_guard.ErrDockerIptablesChainUnavailable
	ErrDockerNftablesChainUnavailable = docker_guard.ErrDockerNftablesChainUnavailable
)

type IDockerPortGuardService interface {
	LoadOverview(context.Context) (dto.DockerPortGuardList, error)
	LoadPublishedPorts(context.Context) ([]dto.DockerPortGuardContainer, error)
	Operate(context.Context, dto.DockerPortGuardOperation) error
	QueueInitialization(dto.DockerPortGuardOperation) (dto.FilterChainOperationResponse, error)
	DeletePolicies(context.Context, dto.DockerPortGuardPolicyBatchDelete) error
	UpsertPolicies(context.Context, dto.DockerPortGuardPolicyBatch) error
	Reconcile(context.Context) error
}

func NewIDockerPortGuardService() IDockerPortGuardService {
	return newDockerPortGuardService()
}

func newDockerPortGuardService() *DockerPortGuardService {
	return &DockerPortGuardService{
		policies: repo.NewIDockerPortGuardRepo(),
		client:   docker.NewDockerClient,
		version:  dockerFirewallVersion,
	}
}

func ReconcileDockerPortGuard(ctx context.Context) error {
	if global.DB == nil {
		return nil
	}
	return NewIDockerPortGuardService().Reconcile(ctx)
}

func ReconcileDockerPortGuardBestEffort(ctx context.Context) {
	if err := ReconcileDockerPortGuard(ctx); err != nil {
		global.LOG.Warnf("reconcile Docker port guard failed, err: %v", err)
	}
}

func (s *DockerPortGuardService) LoadPublishedPorts(ctx context.Context) ([]dto.DockerPortGuardContainer, error) {
	cli, err := s.client()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	defer cli.Close()

	endpoints, err := discoverDockerEndpoints(ctx, cli)
	if err != nil {
		return nil, err
	}
	backend := selectedDockerFirewallBackend("")
	if info, infoErr := cli.Info(ctx); infoErr == nil {
		backend = dockerFirewallBackend(info)
	}
	annotateDockerEndpointTrafficPaths(endpoints, backend)
	return groupDockerGuardContainers(endpoints), nil
}

func (s *DockerPortGuardService) LoadOverview(ctx context.Context) (dto.DockerPortGuardList, error) {
	selectedBackend := selectedDockerFirewallBackend("")
	base := s.runtimeStatus(s.guardRuntime(selectedBackend), selectedBackend)
	base.Version = s.loadFirewallVersion(selectedBackend)
	policies, err := s.policies.List(ctx)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	cli, err := s.client()
	if err != nil {
		base.Message = agenti18n.Get("ErrDockerFailed")
		return dto.DockerPortGuardList{Base: base, Containers: []dto.DockerPortGuardContainer{}, OrphanPolicies: dockerGuardPolicyEndpoints(policies)}, nil
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		base.Message = agenti18n.Get("ErrDockerFailed")
		return dto.DockerPortGuardList{Base: base, Containers: []dto.DockerPortGuardContainer{}, OrphanPolicies: dockerGuardPolicyEndpoints(policies)}, nil
	}
	detectedBackend := dockerFirewallBackend(info)
	base.Backend = selectedDockerFirewallBackend(detectedBackend)
	base = s.runtimeStatus(s.guardRuntime(base.Backend), base.Backend)
	base.Version = s.loadFirewallVersion(base.Backend)
	if reconcileErr := lastDockerPortGuardReconcileError(); reconcileErr != nil {
		markDockerGuardReconcileFailure(&base, reconcileErr)
	}
	endpoints, err := discoverDockerEndpoints(ctx, cli)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	annotateDockerEndpointTrafficPaths(endpoints, detectedBackend)
	endpoints, orphanPolicies := matchDockerGuardPolicies(base, policies, endpoints)
	sort.Slice(endpoints, func(i, j int) bool {
		return guardEndpointKey(endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol) < guardEndpointKey(endpoints[j].Family, endpoints[j].HostIP, endpoints[j].HostPort, endpoints[j].Protocol)
	})
	sort.Slice(orphanPolicies, func(i, j int) bool {
		return guardEndpointKey(orphanPolicies[i].Family, orphanPolicies[i].HostIP, orphanPolicies[i].HostPort, orphanPolicies[i].Protocol) < guardEndpointKey(orphanPolicies[j].Family, orphanPolicies[j].HostIP, orphanPolicies[j].HostPort, orphanPolicies[j].Protocol)
	})
	return dto.DockerPortGuardList{Base: base, Containers: groupDockerGuardContainers(endpoints), OrphanPolicies: orphanPolicies}, nil
}

func matchDockerGuardPolicies(
	base dto.DockerPortGuardBase,
	policies []model.DockerPortGuardPolicy,
	endpoints []dto.DockerPortGuardEndpoint,
) ([]dto.DockerPortGuardEndpoint, []dto.DockerPortGuardEndpoint) {
	byEndpoint := make(map[string]model.DockerPortGuardPolicy, len(policies))
	for _, policy := range policies {
		byEndpoint[guardEndpointKey(policy.Family, policy.HostIP, policy.HostPort, policy.Protocol)] = policy
	}
	for i := range endpoints {
		key := guardEndpointKey(endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol)
		policy, ok := byEndpoint[key]
		if !ok {
			continue
		}
		endpoints[i].PolicyUUID, endpoints[i].Mode, endpoints[i].Sources = policy.UUID, policy.Mode, docker_guard.DecodeSources(policy.Sources)
		endpoints[i].Description = policy.Description
		endpoints[i].Effective = endpoints[i].TrafficPath == dockerTrafficPathForward &&
			((policy.Family == docker_guard.FamilyIPv4 && base.IPv4.Effective) || (policy.Family == docker_guard.FamilyIPv6 && base.IPv6.Effective))
		delete(byEndpoint, key)
	}
	orphanPolicies := make([]dto.DockerPortGuardEndpoint, 0, len(byEndpoint))
	for _, policy := range byEndpoint {
		orphanPolicies = append(orphanPolicies, dto.DockerPortGuardEndpoint{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
			PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: docker_guard.DecodeSources(policy.Sources), Description: policy.Description,
			TrafficPath: dockerTrafficPathUnknown,
		})
	}
	return endpoints, orphanPolicies
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
		if err := runtime.Initialize(policies); err != nil {
			recordDockerPortGuardReconcileError(err)
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, backend); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable); err != nil {
			return err
		}
		recordDockerPortGuardReconcileError(nil)
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
			err = errors.Join(docker_guard.NewManager().Unbind(), docker_guard.NewNftablesManager().Unbind())
		}
		if err != nil {
			return err
		}
		return settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable)
	default:
		return fmt.Errorf("unsupported Docker port guard operation: %s", request.Operation)
	}
}

func (s *DockerPortGuardService) QueueInitialization(
	request dto.DockerPortGuardOperation,
) (dto.FilterChainOperationResponse, error) {
	if request.Operation != "initialize" {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("only Docker port guard initialization can be queued")
	}
	if err := task.CheckScopeTaskIsExecuting(task.TaskScopeFirewall, 0); err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem, err := task.NewTaskWithOps("Docker port guard", task.TaskExec, task.TaskScopeFirewall, request.TaskID, 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("create Docker port guard initialization task: %w", err)
	}
	var runtime dockerGuardRuntime
	var backend string
	var policies []docker_guard.Policy
	taskItem.AddSubTask(agenti18n.GetMsgByKey("FirewallInspectDockerGuardStep"), func(t *task.Task) error {
		var err error
		runtime, backend, err = s.runtimeForDocker(t.TaskCtx)
		if err != nil {
			return err
		}
		policies, err = s.runtimePolicies(t.TaskCtx)
		if err != nil {
			return err
		}
		t.Logf("backend=%s", backend)
		return nil
	}, nil)
	taskItem.AddSubTask(agenti18n.GetWithName("FirewallInitializeDockerGuardStep", "Docker"), func(t *task.Task) error {
		dockerPortGuardServiceMu.Lock()
		defer dockerPortGuardServiceMu.Unlock()
		t.Logf("backend=%s", backend)
		err := runtime.Initialize(policies)
		recordDockerPortGuardReconcileError(err)
		return err
	}, nil)
	taskItem.AddSubTask(agenti18n.GetMsgByKey("FirewallPersistDockerGuardStep"), func(t *task.Task) error {
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, backend); err != nil {
			return err
		}
		if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable); err != nil {
			return err
		}
		recordDockerPortGuardReconcileError(nil)
		return nil
	}, nil)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save Docker port guard initialization task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func (s *DockerPortGuardService) DeletePolicies(ctx context.Context, request dto.DockerPortGuardPolicyBatchDelete) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	uuids, err := docker_guard.NormalizePolicyUUIDs(request.UUIDs)
	if err != nil {
		return err
	}
	if err := s.policies.DeleteBatch(ctx, uuids); err != nil {
		return err
	}
	return s.reconcileLocked(ctx)
}

func (s *DockerPortGuardService) UpsertPolicies(ctx context.Context, request dto.DockerPortGuardPolicyBatch) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	if err := s.rejectHostInputDockerGuardEndpoints(ctx, request.Endpoints); err != nil {
		return err
	}
	policies := make([]model.DockerPortGuardPolicy, 0, len(request.Endpoints))
	seen := make(map[string]struct{}, len(request.Endpoints))
	for _, endpoint := range request.Endpoints {
		normalized, err := docker_guard.NormalizePolicy(docker_guard.Policy{
			Family: endpoint.Family, HostIP: endpoint.HostIP, HostPort: endpoint.HostPort,
			Protocol: endpoint.Protocol, Mode: request.Mode, Sources: request.Sources,
		})
		if err != nil {
			return err
		}
		key := guardEndpointKey(normalized.Family, normalized.HostIP, normalized.HostPort, normalized.Protocol)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		encoded, _ := json.Marshal(normalized.Sources)
		policies = append(policies, model.DockerPortGuardPolicy{
			UUID: uuid.NewString(), Family: normalized.Family, HostIP: normalized.HostIP,
			HostPort: normalized.HostPort, Protocol: normalized.Protocol, Mode: normalized.Mode,
			Sources: string(encoded), Description: strings.TrimSpace(request.Description),
		})
	}
	if err := s.policies.UpsertBatch(ctx, policies); err != nil {
		return err
	}
	return s.reconcileLocked(ctx)
}

func (s *DockerPortGuardService) rejectHostInputDockerGuardEndpoints(
	ctx context.Context,
	requested []dto.DockerPortGuardEndpointIdentity,
) error {
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
	endpoints, err := discoverDockerEndpoints(ctx, cli)
	if err != nil {
		return nil
	}
	annotateDockerEndpointTrafficPaths(endpoints, dockerFirewallBackend(info))
	paths := make(map[string]string, len(endpoints))
	for _, endpoint := range endpoints {
		paths[guardEndpointKey(endpoint.Family, endpoint.HostIP, endpoint.HostPort, endpoint.Protocol)] = endpoint.TrafficPath
	}
	for _, endpoint := range requested {
		if paths[guardEndpointKey(endpoint.Family, endpoint.HostIP, endpoint.HostPort, endpoint.Protocol)] == dockerTrafficPathInput {
			return fmt.Errorf("%w: endpoint traffic is handled by the host input firewall", ErrDockerGuardInvalid)
		}
	}
	return nil
}

func (s *DockerPortGuardService) Reconcile(ctx context.Context) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	return s.reconcileLocked(ctx)
}

func (s *DockerPortGuardService) loadRuleSyncCandidates(
	ctx context.Context,
	request dto.FirewallRuleSyncRequest,
) (string, []model.DockerPortGuardPolicy, dockerGuardRuntime, error) {
	targetProvider, err := databaseRuleSyncTarget(request, "Docker")
	if err != nil {
		return "", nil, nil, err
	}
	target := string(targetProvider)
	selected, err := s.selectedRuleSyncBackend(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	if target != selected {
		return "", nil, nil, fmt.Errorf(
			"%w: selected Docker firewall backend is %s, requested target is %s",
			filter.ErrProviderUnavailable, selected, target,
		)
	}
	policies, err := s.policies.List(ctx)
	if err != nil {
		return "", nil, nil, err
	}
	return target, policies, s.guardRuntime(target), nil
}

func (s *DockerPortGuardService) selectedRuleSyncBackend(ctx context.Context) (string, error) {
	if global.DB != nil {
		selected, _ := settingRepo.GetValueByKey(constant.FirewallDockerBackendKey)
		selected = strings.ToLower(strings.TrimSpace(selected))
		if selected == constant.FirewallProviderIptables || selected == constant.FirewallProviderNftables {
			return selected, nil
		}
	}
	if s.client == nil {
		return "", fmt.Errorf("%w: Docker firewall backend is unavailable", ErrDockerUnavailable)
	}
	cli, err := s.client()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	return selectedDockerFirewallBackend(dockerFirewallBackend(info)), nil
}

func dockerGuardPoliciesFromModels(policies []model.DockerPortGuardPolicy) []docker_guard.Policy {
	result := make([]docker_guard.Policy, 0, len(policies))
	for _, policy := range policies {
		result = append(result, dockerGuardPolicyFromModel(policy))
	}
	return result
}

func dockerGuardPolicyFromModel(policy model.DockerPortGuardPolicy) docker_guard.Policy {
	return docker_guard.Policy{
		UUID: policy.UUID, Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort,
		Protocol: policy.Protocol, Mode: policy.Mode, Sources: docker_guard.DecodeSources(policy.Sources),
	}
}

func dockerGuardRuleSyncDTO(policy model.DockerPortGuardPolicy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
		PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: docker_guard.DecodeSources(policy.Sources), Description: policy.Description,
		TrafficPath: dockerTrafficPathUnknown,
	}
}

func dockerGuardRuntimeRuleSyncDTO(policy docker_guard.Policy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
		PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: append([]string(nil), policy.Sources...),
		TrafficPath: dockerTrafficPathUnknown,
	}
}

func (s *DockerPortGuardService) reconcileLocked(ctx context.Context) (err error) {
	defer func() { recordDockerPortGuardReconcileError(err) }()
	persistedEnabled, err := dockerPortGuardPersistedEnabled()
	if err != nil {
		return fmt.Errorf("load Docker port guard persisted status: %w", err)
	}
	initialized, err := s.anyRuntimeInitialized()
	if err != nil {
		return &docker_guard.FamilyError{Family: docker_guard.FamilyIPv4, Err: fmt.Errorf("inspect initialization: %w", err)}
	}
	if !initialized && !persistedEnabled {
		return nil
	}
	runtime, _, err := s.runtimeForDocker(ctx)
	if err != nil {
		return err
	}
	initialized, err = runtime.Initialized(docker_guard.FamilyIPv4)
	if err != nil {
		return &docker_guard.FamilyError{Family: docker_guard.FamilyIPv4, Err: fmt.Errorf("inspect initialization: %w", err)}
	}
	if !initialized && !persistedEnabled {
		return nil
	}
	policies, err := s.runtimePolicies(ctx)
	if err != nil {
		return err
	}
	if !initialized {
		err = runtime.Initialize(policies)
	} else {
		err = runtime.Reconcile(policies)
	}
	return err
}

func dockerPortGuardPersistedEnabled() (bool, error) {
	status, err := settingRepo.GetValueByKey(constant.FirewallDockerPortGuardStatusKey)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return status == constant.StatusEnable, err
}

func (s *DockerPortGuardService) anyRuntimeInitialized() (bool, error) {
	if s.runtime != nil {
		return s.runtime.Initialized(docker_guard.FamilyIPv4)
	}
	for _, runtime := range []dockerGuardRuntime{docker_guard.NewManager(), docker_guard.NewNftablesManager()} {
		initialized, err := runtime.Initialized(docker_guard.FamilyIPv4)
		if err != nil || initialized {
			return initialized, err
		}
	}
	return false, nil
}

func recordDockerPortGuardReconcileError(err error) {
	dockerPortGuardSyncMu.Lock()
	dockerPortGuardSyncErr = err
	dockerPortGuardSyncMu.Unlock()
}

func lastDockerPortGuardReconcileError() error {
	dockerPortGuardSyncMu.RLock()
	defer dockerPortGuardSyncMu.RUnlock()
	return dockerPortGuardSyncErr
}

func markDockerGuardReconcileFailure(base *dto.DockerPortGuardBase, err error) {
	var familyErr *docker_guard.FamilyError
	if errors.As(err, &familyErr) {
		markDockerGuardFamilyNotEffective(base, familyErr.Family)
		if familyErr.Family == docker_guard.FamilyIPv4 {
			markDockerGuardFamilyNotEffective(base, docker_guard.FamilyIPv6)
		}
		return
	}
	markDockerGuardFamilyNotEffective(base, docker_guard.FamilyIPv4)
	markDockerGuardFamilyNotEffective(base, docker_guard.FamilyIPv6)
}

func markDockerGuardFamilyNotEffective(base *dto.DockerPortGuardBase, family string) {
	var status *dto.DockerPortGuardFamilyStatus
	switch family {
	case docker_guard.FamilyIPv4:
		status = &base.IPv4
	case docker_guard.FamilyIPv6:
		status = &base.IPv6
	default:
		return
	}
	if !status.Initialized {
		return
	}
	status.State = docker_guard.StatusNotEffective
	status.Reason = docker_guard.ReasonInspectFailed
	status.Effective = false
}

func (s *DockerPortGuardService) runtimePolicies(ctx context.Context) ([]docker_guard.Policy, error) {
	stored, err := s.policies.List(ctx)
	if err != nil {
		return nil, err
	}
	policies := make([]docker_guard.Policy, 0, len(stored))
	for _, policy := range stored {
		policies = append(policies, docker_guard.Policy{UUID: policy.UUID, Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol, Mode: policy.Mode, Sources: docker_guard.DecodeSources(policy.Sources)})
	}
	return policies, nil
}

func (s *DockerPortGuardService) runtimeStatus(runtime dockerGuardRuntime, backend string) dto.DockerPortGuardBase {
	ipv4 := runtime.Status(docker_guard.FamilyIPv4)
	ipv6 := runtime.Status(docker_guard.FamilyIPv6)
	return dto.DockerPortGuardBase{
		Name:        dockerFirewallDisplayName(backend),
		Backend:     backend,
		IsExist:     ipv4.Reason != docker_guard.ReasonCommandMissing || ipv6.Reason != docker_guard.ReasonCommandMissing,
		Initialized: ipv4.Initialized || ipv6.Initialized,
		Bound:       ipv4.Bound || ipv6.Bound,
		IPv4:        dto.DockerPortGuardFamilyStatus{State: ipv4.State, Reason: ipv4.Reason, Initialized: ipv4.Initialized, Bound: ipv4.Bound, Effective: ipv4.Effective},
		IPv6:        dto.DockerPortGuardFamilyStatus{State: ipv6.State, Reason: ipv6.Reason, Initialized: ipv6.Initialized, Bound: ipv6.Bound, Effective: ipv6.Effective},
	}
}

func (s *DockerPortGuardService) guardRuntime(backend string) dockerGuardRuntime {
	if s.runtimeForBackend != nil {
		return s.runtimeForBackend(backend)
	}
	if s.runtime != nil {
		return s.runtime
	}
	return docker_guard.NewRuntime(backend)
}

func (s *DockerPortGuardService) runtimeForDocker(ctx context.Context) (dockerGuardRuntime, string, error) {
	if s.runtime != nil {
		return s.runtime, selectedDockerFirewallBackend(constant.FirewallProviderIptables), nil
	}
	cli, err := s.client()
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %v", ErrDockerUnavailable, err)
	}
	backend := selectedDockerFirewallBackend(dockerFirewallBackend(info))
	if backend != constant.FirewallProviderIptables && backend != constant.FirewallProviderNftables {
		return nil, backend, fmt.Errorf("Docker firewall backend %q is not supported", backend)
	}
	return s.guardRuntime(backend), backend, nil
}

func dockerFirewallBackend(info system.Info) string {
	if info.FirewallBackend == nil || info.FirewallBackend.Driver == "" {
		return constant.FirewallProviderIptables
	}
	return strings.ToLower(info.FirewallBackend.Driver)
}

func dockerFirewallDisplayName(backend string) string {
	switch strings.ToLower(strings.TrimSpace(backend)) {
	case constant.FirewallProviderNftables:
		return "nftables-docker"
	default:
		return "iptables-docker"
	}
}

func (s *DockerPortGuardService) loadFirewallVersion(backend string) string {
	if s.version == nil {
		return "-"
	}
	return s.version(backend)
}

func dockerFirewallVersion(backend string) string {
	client, err := lifecycle.NewClientFor(backend)
	if err != nil {
		return "-"
	}
	version, err := client.Version()
	if err != nil || strings.TrimSpace(version) == "" {
		return "-"
	}
	return version
}

func discoverDockerEndpoints(ctx context.Context, cli *client.Client) ([]dto.DockerPortGuardEndpoint, error) {
	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	endpoints := make([]dto.DockerPortGuardEndpoint, 0)
	for _, item := range containers {
		name := strings.TrimPrefix(firstGuardString(item.Names), "/")
		compose := item.Labels[dockerGuardComposeProjectLabel]
		application := ""
		if created, ok := item.Labels[dockerGuardComposeCreatedBy]; ok && created == "Apps" {
			application = compose
		}
		for _, port := range item.Ports {
			if port.PublicPort == 0 || (port.Type != "tcp" && port.Type != "udp") {
				continue
			}
			family := docker_guard.FamilyIPv4
			hostIP := port.IP
			if addr, err := netip.ParseAddr(hostIP); err == nil && addr.Is6() {
				family = docker_guard.FamilyIPv6
			} else if hostIP == "" {
				hostIP = "0.0.0.0"
			}
			endpoints = append(endpoints, dto.DockerPortGuardEndpoint{Family: family, HostIP: hostIP, HostPort: port.PublicPort, Protocol: port.Type, ContainerID: item.ID, ContainerName: name, ContainerPort: port.PrivatePort, Compose: compose, Application: application, Sources: []string{}})
		}
	}
	return endpoints, nil
}

func dockerGuardPolicyEndpoints(policies []model.DockerPortGuardPolicy) []dto.DockerPortGuardEndpoint {
	endpoints := make([]dto.DockerPortGuardEndpoint, 0, len(policies))
	for _, policy := range policies {
		endpoints = append(endpoints, dto.DockerPortGuardEndpoint{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
			PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: docker_guard.DecodeSources(policy.Sources),
			Description: policy.Description, TrafficPath: dockerTrafficPathUnknown,
		})
	}
	return endpoints
}

func groupDockerGuardContainers(endpoints []dto.DockerPortGuardEndpoint) []dto.DockerPortGuardContainer {
	containers := make(map[string]*dto.DockerPortGuardContainer)
	order := make([]string, 0)
	for _, endpoint := range endpoints {
		key := endpoint.ContainerID
		if key == "" {
			key = "__orphan__"
		}
		container, ok := containers[key]
		if !ok {
			container = &dto.DockerPortGuardContainer{
				Key: key, Name: endpoint.ContainerName, Compose: endpoint.Compose,
				Application: endpoint.Application, Endpoints: []dto.DockerPortGuardEndpoint{},
			}
			containers[key] = container
			order = append(order, key)
		}
		container.Endpoints = append(container.Endpoints, endpoint)
	}
	sort.Slice(order, func(i, j int) bool {
		return containers[order[i]].Name < containers[order[j]].Name
	})

	result := make([]dto.DockerPortGuardContainer, 0, len(order))
	for _, key := range order {
		container := containers[key]
		items := make([]docker.PortRangeItem, 0, len(container.Endpoints))
		for i, endpoint := range container.Endpoints {
			sources := append([]string(nil), endpoint.Sources...)
			sort.Strings(sources)
			policyKey := fmt.Sprintf("%t|%s|%s|%t|%s|%s", endpoint.PolicyUUID != "", endpoint.Mode, strings.Join(sources, ","), endpoint.Effective, endpoint.Description, endpoint.TrafficPath)
			items = append(items, docker.PortRangeItem{
				Key:        endpoint.Family + "|" + endpoint.HostIP + "|" + endpoint.Protocol + "|" + policyKey,
				PublicPort: endpoint.HostPort, PrivatePort: endpoint.ContainerPort,
				HasPrivatePort: endpoint.ContainerPort != 0, Position: i,
			})
		}
		container.PortGroups = make([]dto.DockerPortGuardPortGroup, 0, len(items))
		for _, portRange := range docker.MergePortRanges(items) {
			start := container.Endpoints[portRange.Start.Position]
			address := start.HostIP
			if strings.Contains(address, ":") {
				address = "[" + address + "]"
			}
			ports := fmt.Sprintf("%d", portRange.Start.PublicPort)
			if portRange.Start.PublicPort != portRange.End.PublicPort {
				ports = fmt.Sprintf("%d-%d", portRange.Start.PublicPort, portRange.End.PublicPort)
			}
			container.PortGroups = append(container.PortGroups, dto.DockerPortGuardPortGroup{
				Key:   fmt.Sprintf("%s|%d-%d", portRange.Start.Key, portRange.Start.PublicPort, portRange.End.PublicPort),
				Label: fmt.Sprintf("%s:%s/%s", address, ports, start.Protocol), Endpoint: start,
				Endpoints: func() []dto.DockerPortGuardEndpoint {
					members := make([]dto.DockerPortGuardEndpoint, 0, len(portRange.Items))
					for _, item := range portRange.Items {
						members = append(members, container.Endpoints[item.Position])
					}
					return members
				}(),
			})
		}
		result = append(result, *container)
	}
	return result
}

func guardEndpointKey(family, hostIP string, hostPort uint16, protocol string) string {
	return fmt.Sprintf("%s|%s|%d|%s", family, hostIP, hostPort, protocol)
}

func firstGuardString(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func annotateDockerEndpointTrafficPaths(endpoints []dto.DockerPortGuardEndpoint, backend string) {
	rules := map[string]dockerForwardRules{
		constant.FirewallFamilyIPv4: loadDockerDNATRules(backend, constant.FirewallFamilyIPv4),
		constant.FirewallFamilyIPv6: loadDockerDNATRules(backend, constant.FirewallFamilyIPv6),
	}
	proxies := loadDockerProxyEndpoints()
	for i := range endpoints {
		familyRules := rules[endpoints[i].Family]
		endpoints[i].TrafficPath = dockerEndpointTrafficPath(backend, familyRules.output, familyRules.inspected, proxies, endpoints[i])
	}
}

func dockerEndpointTrafficPath(backend, rules string, inspected bool, proxies []dockerProxyEndpoint, endpoint dto.DockerPortGuardEndpoint) string {
	if !inspected {
		return dockerTrafficPathUnknown
	}
	if dockerDNATRuleMatches(backend, rules, endpoint) {
		return dockerTrafficPathForward
	}
	if dockerProxyEndpointMatches(proxies, endpoint) {
		return dockerTrafficPathInput
	}
	return dockerTrafficPathUnknown
}

func loadDockerDNATRules(backend, family string) dockerForwardRules {
	manager := cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second), cmd.WithEnv("LC_ALL=C"))
	if backend == constant.FirewallProviderNftables {
		tableFamily := "ip"
		if family == constant.FirewallFamilyIPv6 {
			tableFamily = "ip6"
		}
		tables, err := manager.RunWithOptionalSudoAndStdout("nft", "list", "tables")
		if err != nil {
			return dockerForwardRules{}
		}
		if !strings.Contains(tables, "table "+tableFamily+" docker-bridges") {
			return dockerForwardRules{inspected: true}
		}
		output, err := manager.RunWithOptionalSudoAndStdout("nft", "list", "table", tableFamily, "docker-bridges")
		return dockerForwardRules{output: output, inspected: err == nil}
	}
	commands, err := lifecycle.ResolveIptablesCommands()
	if err != nil {
		return dockerForwardRules{}
	}
	executable := commands.IPv4
	if family == constant.FirewallFamilyIPv6 {
		executable = commands.IPv6
	}
	if executable == "" {
		return dockerForwardRules{}
	}
	output, err := manager.RunWithOptionalSudoAndStdout(executable, "-w", "-t", "nat", "-S")
	return dockerForwardRules{output: output, inspected: err == nil}
}

func loadDockerProxyEndpoints() []dockerProxyEndpoint {
	manager := cmd.NewCommandMgr(cmd.WithTimeout(10*time.Second), cmd.WithEnv("LC_ALL=C"))
	output, err := manager.RunWithStdout("ps", "-ww", "-eo", "args=")
	if err != nil {
		return nil
	}
	return parseDockerProxyEndpoints(output)
}

func parseDockerProxyEndpoints(output string) []dockerProxyEndpoint {
	result := make([]dockerProxyEndpoint, 0)
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || !dockerProxyCommand(fields) {
			continue
		}
		protocol := commandFlagValue(fields, "-proto")
		hostIP := commandFlagValue(fields, "-host-ip")
		hostPortValue := commandFlagValue(fields, "-host-port")
		hostPort, err := strconv.ParseUint(hostPortValue, 10, 16)
		if err != nil || (protocol != "tcp" && protocol != "udp") || hostIP == "" {
			continue
		}
		result = append(result, dockerProxyEndpoint{protocol: protocol, hostIP: canonicalAddress(hostIP), hostPort: uint16(hostPort)})
	}
	return result
}

func dockerProxyCommand(fields []string) bool {
	for _, field := range fields {
		if filepath.Base(field) == "docker-proxy" {
			return true
		}
	}
	return false
}

func commandFlagValue(fields []string, name string) string {
	for i := 0; i < len(fields); i++ {
		if fields[i] == name && i+1 < len(fields) {
			return fields[i+1]
		}
		if strings.HasPrefix(fields[i], name+"=") {
			return strings.TrimPrefix(fields[i], name+"=")
		}
	}
	return ""
}

func dockerProxyEndpointMatches(proxies []dockerProxyEndpoint, endpoint dto.DockerPortGuardEndpoint) bool {
	for _, proxy := range proxies {
		if proxy.protocol == endpoint.Protocol && proxy.hostPort == endpoint.HostPort && hostAddressMatches(proxy.hostIP, endpoint.HostIP, endpoint.Family) {
			return true
		}
	}
	return false
}

func dockerDNATRuleMatches(backend, output string, endpoint dto.DockerPortGuardEndpoint) bool {
	if strings.TrimSpace(output) == "" {
		return false
	}
	if backend == constant.FirewallProviderNftables {
		return nftDNATRuleMatches(output, endpoint)
	}
	return iptablesDNATRuleMatches(output, endpoint)
}

func iptablesDNATRuleMatches(output string, endpoint dto.DockerPortGuardEndpoint) bool {
	port := strconv.Itoa(int(endpoint.HostPort))
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if commandFlagValue(fields, "-p") != endpoint.Protocol || commandFlagValue(fields, "--dport") != port || commandFlagValue(fields, "-j") != "DNAT" {
			continue
		}
		if destinationAddressMatches(commandFlagValue(fields, "-d"), endpoint) {
			return true
		}
	}
	return false
}

func nftDNATRuleMatches(output string, endpoint dto.DockerPortGuardEndpoint) bool {
	port := strconv.Itoa(int(endpoint.HostPort))
	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(strings.NewReplacer("{", " ", "}", " ", ",", " ", ";", " ").Replace(line))
		if !containsToken(fields, "dnat") || !nftProtocolPortMatches(fields, endpoint.Protocol, port) {
			continue
		}
		destination := nftDestinationAddress(fields, endpoint.Family)
		if destinationAddressMatches(destination, endpoint) {
			return true
		}
	}
	return false
}

func nftProtocolPortMatches(fields []string, protocol, port string) bool {
	for i := 0; i+2 < len(fields); i++ {
		if fields[i] == protocol && fields[i+1] == "dport" && fields[i+2] == port {
			return true
		}
		if fields[i] == "th" && fields[i+1] == "dport" && fields[i+2] == port && nftMetaProtocolMatches(fields, protocol) {
			return true
		}
	}
	return false
}

func nftMetaProtocolMatches(fields []string, protocol string) bool {
	for i := 0; i+2 < len(fields); i++ {
		if fields[i] == "meta" && fields[i+1] == "l4proto" && fields[i+2] == protocol {
			return true
		}
	}
	return false
}

func nftDestinationAddress(fields []string, family string) string {
	token := "ip"
	if family == constant.FirewallFamilyIPv6 {
		token = "ip6"
	}
	for i := 0; i+2 < len(fields); i++ {
		if fields[i] == token && fields[i+1] == "daddr" {
			return fields[i+2]
		}
	}
	return ""
}

func destinationAddressMatches(ruleAddress string, endpoint dto.DockerPortGuardEndpoint) bool {
	ruleAddress = strings.TrimSpace(strings.Split(ruleAddress, "/")[0])
	if isWildcardHostAddress(endpoint.HostIP, endpoint.Family) {
		return ruleAddress == ""
	}
	return ruleAddress == "" || canonicalAddress(ruleAddress) == canonicalAddress(endpoint.HostIP)
}

func hostAddressMatches(left, right, family string) bool {
	if isWildcardHostAddress(left, family) && isWildcardHostAddress(right, family) {
		return true
	}
	return canonicalAddress(left) == canonicalAddress(right)
}

func isWildcardHostAddress(value, family string) bool {
	value = strings.TrimSpace(value)
	if family == constant.FirewallFamilyIPv6 {
		return value == "" || value == "::"
	}
	return value == "" || value == "0.0.0.0"
}

func canonicalAddress(value string) string {
	if address, err := netip.ParseAddr(strings.TrimSpace(value)); err == nil {
		return address.String()
	}
	return strings.TrimSpace(value)
}

func containsToken(fields []string, value string) bool {
	for _, field := range fields {
		if field == value {
			return true
		}
	}
	return false
}
