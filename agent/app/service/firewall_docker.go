package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	agenti18n "github.com/1Panel-dev/1Panel/agent/i18n"
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
)

type dockerGuardRuntime interface {
	Initialize([]docker_guard.Policy) error
	Bind() error
	Reconcile([]docker_guard.Policy) error
	Unbind() error
	Cleanup() error
	Initialized(string) (bool, error)
	Status(string) docker_guard.FamilyStatus
	ListPolicies() ([]docker_guard.Policy, error)
}

type DockerPortGuardService struct {
	policies          repo.IDockerPortGuardRepo
	runtime           dockerGuardRuntime
	runtimeForBackend func(string) dockerGuardRuntime
	client            func() (*client.Client, error)
	version           func(string) string
}

type normalizedDockerGuardPolicy struct {
	Family   string
	HostIP   string
	HostPort uint16
	Protocol string
	Mode     string
}

var (
	dockerPortGuardServiceMu sync.Mutex
	dockerPortGuardSyncMu    sync.RWMutex
	dockerPortGuardSyncErr   error
	ErrDockerGuardInvalid    = errors.New("invalid Docker port guard request")
	ErrDockerUnavailable     = errors.New("Docker is unavailable")
)

type IDockerPortGuardService interface {
	LoadOverview(context.Context) (dto.DockerPortGuardList, error)
	Operate(context.Context, dto.DockerPortGuardOperation) error
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
	base.Backend = selectedDockerFirewallBackend(dockerFirewallBackend(info))
	base = s.runtimeStatus(s.guardRuntime(base.Backend), base.Backend)
	base.Version = s.loadFirewallVersion(base.Backend)
	if reconcileErr := lastDockerPortGuardReconcileError(); reconcileErr != nil {
		base.Message = reconcileErr.Error()
		markDockerGuardReconcileFailure(&base, reconcileErr)
	}
	endpoints, err := discoverDockerEndpoints(ctx, cli)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
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
		endpoints[i].PolicyUUID, endpoints[i].Mode, endpoints[i].Sources = policy.UUID, policy.Mode, decodeGuardSources(policy.Sources)
		endpoints[i].Description = policy.Description
		endpoints[i].Effective = (policy.Family == docker_guard.FamilyIPv4 && base.IPv4.Effective) || (policy.Family == docker_guard.FamilyIPv6 && base.IPv6.Effective)
		delete(byEndpoint, key)
	}
	orphanPolicies := make([]dto.DockerPortGuardEndpoint, 0, len(byEndpoint))
	for _, policy := range byEndpoint {
		orphanPolicies = append(orphanPolicies, dto.DockerPortGuardEndpoint{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
			PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: decodeGuardSources(policy.Sources), Description: policy.Description,
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

func (s *DockerPortGuardService) DeletePolicies(ctx context.Context, request dto.DockerPortGuardPolicyBatchDelete) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	uuids, err := normalizeDockerGuardPolicyUUIDs(request.UUIDs)
	if err != nil {
		return err
	}
	if err := s.policies.DeleteBatch(ctx, uuids); err != nil {
		return err
	}
	return s.reconcileLocked(ctx)
}

func normalizeDockerGuardPolicyUUIDs(values []string) ([]string, error) {
	uuids := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, policyUUID := range values {
		policyUUID = strings.TrimSpace(policyUUID)
		if policyUUID == "" {
			return nil, fmt.Errorf("%w: policy UUID cannot be empty", ErrDockerGuardInvalid)
		}
		if _, exists := seen[policyUUID]; exists {
			continue
		}
		seen[policyUUID] = struct{}{}
		uuids = append(uuids, policyUUID)
	}
	if len(uuids) == 0 {
		return nil, fmt.Errorf("%w: policy UUIDs cannot be empty", ErrDockerGuardInvalid)
	}
	return uuids, nil
}

func (s *DockerPortGuardService) UpsertPolicies(ctx context.Context, request dto.DockerPortGuardPolicyBatch) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	policies := make([]model.DockerPortGuardPolicy, 0, len(request.Endpoints))
	seen := make(map[string]struct{}, len(request.Endpoints))
	for _, endpoint := range request.Endpoints {
		normalized, sources, err := normalizeGuardPolicy(endpoint.Family, endpoint.HostIP, endpoint.HostPort, endpoint.Protocol, request.Mode, request.Sources)
		if err != nil {
			return err
		}
		key := guardEndpointKey(normalized.Family, normalized.HostIP, normalized.HostPort, normalized.Protocol)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		encoded, _ := json.Marshal(sources)
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
		Protocol: policy.Protocol, Mode: policy.Mode, Sources: decodeGuardSources(policy.Sources),
	}
}

func dockerGuardPolicySyncKey(policy docker_guard.Policy) string {
	mode := policy.Mode
	if mode == docker_guard.ModeAllow && len(policy.Sources) == 0 {
		mode = docker_guard.ModeAll
	}
	sources := append([]string(nil), policy.Sources...)
	sort.Strings(sources)
	return strings.Join([]string{
		policy.UUID, policy.Family, canonicalGuardHost(policy.HostIP), strconv.Itoa(int(policy.HostPort)),
		policy.Protocol, mode, strings.Join(sources, ","),
	}, "\x00")
}

func canonicalGuardHost(value string) string {
	if address, err := netip.ParseAddr(value); err == nil {
		return address.String()
	}
	return value
}

func verifyDockerGuardRuleSync(runtime dockerGuardRuntime, desired []docker_guard.Policy) error {
	actual, err := runtime.ListPolicies()
	if err != nil {
		return fmt.Errorf("verify synchronized Docker firewall policies: %w", err)
	}
	if !databaseSyncStatesEqual(actual, desired, dockerGuardPolicySyncKey) {
		return fmt.Errorf("verify synchronized Docker firewall policies: target policies do not match the database")
	}
	return nil
}

func reconcileDockerGuardSyncTarget(backend string, policies []docker_guard.Policy, runtime dockerGuardRuntime) error {
	families := make(map[string]struct{}, len(policies))
	needsInitialize, needsBind := false, false
	for _, policy := range policies {
		families[policy.Family] = struct{}{}
	}
	if len(families) == 0 {
		initialized := false
		for _, family := range []string{docker_guard.FamilyIPv4, docker_guard.FamilyIPv6} {
			status := runtime.Status(family)
			if status.Reason == docker_guard.ReasonInspectFailed {
				return fmt.Errorf("inspect Docker firewall target %s for %s failed", backend, family)
			}
			initialized = initialized || status.Initialized
		}
		if initialized {
			return runtime.Reconcile(nil)
		}
		return nil
	}
	for family := range families {
		status := runtime.Status(family)
		needsInitialize = needsInitialize || !status.Initialized
		needsBind = needsBind || !status.Bound || !status.Effective
	}
	var err error
	if needsInitialize {
		err = runtime.Initialize(policies)
	} else {
		if needsBind {
			err = runtime.Bind()
		}
		if err == nil {
			err = runtime.Reconcile(policies)
		}
	}
	if err != nil {
		return err
	}
	for family := range families {
		if !runtime.Status(family).Effective {
			return fmt.Errorf("Docker firewall target %s is not effective for %s", backend, family)
		}
	}
	return nil
}

func dockerGuardRuleSyncDTO(policy model.DockerPortGuardPolicy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
		PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: decodeGuardSources(policy.Sources), Description: policy.Description,
	}
}

func dockerGuardRuntimeRuleSyncDTO(policy docker_guard.Policy) *dto.DockerPortGuardEndpoint {
	return &dto.DockerPortGuardEndpoint{
		Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
		PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: append([]string(nil), policy.Sources...),
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
		policies = append(policies, docker_guard.Policy{UUID: policy.UUID, Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol, Mode: policy.Mode, Sources: decodeGuardSources(policy.Sources)})
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
	if backend == constant.FirewallProviderNftables {
		return docker_guard.NewNftablesManager()
	}
	return docker_guard.NewManager()
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

func normalizeGuardPolicy(family, hostIP string, hostPort uint16, protocol, mode string, sources []string) (normalizedDockerGuardPolicy, []string, error) {
	family, hostIP, protocol, mode = strings.ToLower(strings.TrimSpace(family)), strings.TrimSpace(hostIP), strings.ToLower(strings.TrimSpace(protocol)), strings.ToLower(strings.TrimSpace(mode))
	if hostPort == 0 || (protocol != "tcp" && protocol != "udp") || (family != docker_guard.FamilyIPv4 && family != docker_guard.FamilyIPv6) || (mode != docker_guard.ModeAll && mode != docker_guard.ModeSources && mode != docker_guard.ModeAllow) {
		return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("%w: invalid policy fields", ErrDockerGuardInvalid)
	}
	addr, err := netip.ParseAddr(hostIP)
	if err != nil || (family == docker_guard.FamilyIPv4) != addr.Is4() {
		return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("%w: host IP does not match address family", ErrDockerGuardInvalid)
	}
	normalizedSources := make([]string, 0, len(sources))
	seen := map[string]struct{}{}
	for _, source := range sources {
		source = strings.TrimSpace(source)
		if source == "" {
			continue
		}
		prefix, err := netip.ParsePrefix(source)
		if err != nil {
			if sourceAddr, addrErr := netip.ParseAddr(source); addrErr == nil {
				bits := 128
				if sourceAddr.Is4() {
					bits = 32
				}
				prefix = netip.PrefixFrom(sourceAddr, bits)
			} else {
				return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("%w: invalid source address %q", ErrDockerGuardInvalid, source)
			}
		}
		if (family == docker_guard.FamilyIPv4) != prefix.Addr().Is4() {
			return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("%w: source %q does not match address family", ErrDockerGuardInvalid, source)
		}
		canonical := prefix.Masked().String()
		if _, ok := seen[canonical]; !ok {
			seen[canonical] = struct{}{}
			normalizedSources = append(normalizedSources, canonical)
		}
	}
	if mode == docker_guard.ModeSources && len(normalizedSources) == 0 {
		return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("%w: deny_sources requires at least one source", ErrDockerGuardInvalid)
	}
	if mode == docker_guard.ModeAll {
		normalizedSources = []string{}
	}
	sort.Strings(normalizedSources)
	return normalizedDockerGuardPolicy{Family: family, HostIP: hostIP, HostPort: hostPort, Protocol: protocol, Mode: mode}, normalizedSources, nil
}

func decodeGuardSources(value string) []string {
	result := []string{}
	_ = json.Unmarshal([]byte(value), &result)
	return result
}

func dockerGuardPolicyEndpoints(policies []model.DockerPortGuardPolicy) []dto.DockerPortGuardEndpoint {
	endpoints := make([]dto.DockerPortGuardEndpoint, 0, len(policies))
	for _, policy := range policies {
		endpoints = append(endpoints, dto.DockerPortGuardEndpoint{
			Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol,
			PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: decodeGuardSources(policy.Sources),
			Description: policy.Description,
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
			policyKey := fmt.Sprintf("%t|%s|%s|%t|%s", endpoint.PolicyUUID != "", endpoint.Mode, strings.Join(sources, ","), endpoint.Effective, endpoint.Description)
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
