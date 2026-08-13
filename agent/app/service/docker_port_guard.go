package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/netip"
	"sort"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/google/uuid"
)

type dockerGuardRuntime interface {
	Initialize([]docker_guard.Policy) error
	Bind() error
	Reconcile([]docker_guard.Policy) error
	Unbind() error
	Initialized(string) (bool, error)
	Status(string) docker_guard.FamilyStatus
}

type DockerPortGuardService struct {
	policies repo.IDockerPortGuardRepo
	runtime  dockerGuardRuntime
	client   func() (*client.Client, error)
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
)

type IDockerPortGuardService interface {
	List(context.Context) (dto.DockerPortGuardList, error)
	Operate(context.Context, dto.DockerPortGuardOperation) error
	DeleteBatch(context.Context, dto.DockerPortGuardPolicyBatchDelete) error
	UpsertBatch(context.Context, dto.DockerPortGuardPolicyBatch) error
	Reconcile(context.Context) error
}

func NewIDockerPortGuardService() IDockerPortGuardService {
	return &DockerPortGuardService{policies: repo.NewIDockerPortGuardRepo(), runtime: docker_guard.NewManager(), client: docker.NewDockerClient}
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

func (s *DockerPortGuardService) List(ctx context.Context) (dto.DockerPortGuardList, error) {
	base := s.runtimeStatus()
	policies, err := s.policies.List(ctx)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	cli, err := s.client()
	if err != nil {
		base.Message = err.Error()
		return dto.DockerPortGuardList{Base: base, Containers: groupDockerGuardContainers(dockerGuardPolicyEndpoints(policies))}, nil
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		base.Message = err.Error()
		return dto.DockerPortGuardList{Base: base, Containers: groupDockerGuardContainers(dockerGuardPolicyEndpoints(policies))}, nil
	}
	base.Backend = dockerFirewallBackend(info)
	if reconcileErr := lastDockerPortGuardReconcileError(); reconcileErr != nil {
		base.Message = reconcileErr.Error()
		markDockerGuardReconcileFailure(&base, reconcileErr)
	}
	endpoints, err := discoverDockerEndpoints(ctx, cli)
	if err != nil {
		return dto.DockerPortGuardList{}, err
	}
	byEndpoint := make(map[string]model.DockerPortGuardPolicy, len(policies))
	for _, policy := range policies {
		byEndpoint[guardEndpointKey(policy.Family, policy.HostIP, policy.HostPort, policy.Protocol)] = policy
	}
	for i := range endpoints {
		if policy, ok := byEndpoint[guardEndpointKey(endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol)]; ok {
			endpoints[i].PolicyUUID, endpoints[i].Mode, endpoints[i].Sources = policy.UUID, policy.Mode, decodeGuardSources(policy.Sources)
			endpoints[i].Description = policy.Description
			endpoints[i].Effective = (policy.Family == docker_guard.FamilyIPv4 && base.IPv4.Effective) || (policy.Family == docker_guard.FamilyIPv6 && base.IPv6.Effective)
			delete(byEndpoint, guardEndpointKey(policy.Family, policy.HostIP, policy.HostPort, policy.Protocol))
		}
	}
	// Keep policies visible after a container is stopped or an endpoint is temporarily absent.
	for _, policy := range byEndpoint {
		endpoints = append(endpoints, dto.DockerPortGuardEndpoint{Family: policy.Family, HostIP: policy.HostIP, HostPort: policy.HostPort, Protocol: policy.Protocol, PolicyUUID: policy.UUID, Mode: policy.Mode, Sources: decodeGuardSources(policy.Sources), Effective: false, Description: policy.Description})
	}
	sort.Slice(endpoints, func(i, j int) bool {
		return guardEndpointKey(endpoints[i].Family, endpoints[i].HostIP, endpoints[i].HostPort, endpoints[i].Protocol) < guardEndpointKey(endpoints[j].Family, endpoints[j].HostIP, endpoints[j].HostPort, endpoints[j].Protocol)
	})
	return dto.DockerPortGuardList{Base: base, Containers: groupDockerGuardContainers(endpoints)}, nil
}

func (s *DockerPortGuardService) Operate(ctx context.Context, request dto.DockerPortGuardOperation) error {
	dockerPortGuardServiceMu.Lock()
	defer dockerPortGuardServiceMu.Unlock()
	switch request.Operation {
	case "initialize":
		if err := s.checkBackend(ctx); err != nil {
			return err
		}
		policies, err := s.runtimePolicies(ctx)
		if err != nil {
			return err
		}
		if err := s.runtime.Initialize(policies); err != nil {
			recordDockerPortGuardReconcileError(err)
			return err
		}
		recordDockerPortGuardReconcileError(nil)
		return nil
	case "bind":
		if err := s.checkBackend(ctx); err != nil {
			return err
		}
		return s.runtime.Bind()
	case "unbind":
		return s.runtime.Unbind()
	default:
		return fmt.Errorf("unsupported Docker port guard operation: %s", request.Operation)
	}
}

func (s *DockerPortGuardService) DeleteBatch(ctx context.Context, request dto.DockerPortGuardPolicyBatchDelete) error {
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
			return nil, errors.New("Docker port guard policy UUID cannot be empty")
		}
		if _, exists := seen[policyUUID]; exists {
			continue
		}
		seen[policyUUID] = struct{}{}
		uuids = append(uuids, policyUUID)
	}
	if len(uuids) == 0 {
		return nil, errors.New("Docker port guard policy UUIDs cannot be empty")
	}
	return uuids, nil
}

func (s *DockerPortGuardService) UpsertBatch(ctx context.Context, request dto.DockerPortGuardPolicyBatch) error {
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

func (s *DockerPortGuardService) reconcileLocked(ctx context.Context) (err error) {
	defer func() { recordDockerPortGuardReconcileError(err) }()
	initialized, err := s.runtime.Initialized(docker_guard.FamilyIPv4)
	if err != nil {
		return &docker_guard.FamilyError{Family: docker_guard.FamilyIPv4, Err: fmt.Errorf("inspect initialization: %w", err)}
	}
	if !initialized {
		return nil
	}
	if err := s.checkBackend(ctx); err != nil {
		return err
	}
	policies, err := s.runtimePolicies(ctx)
	if err != nil {
		return err
	}
	return s.runtime.Reconcile(policies)
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
		// Families are reconciled in IPv4/IPv6 order. If IPv4 fails, IPv6 was not
		// attempted and cannot be reported as synchronized with the desired state.
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

func (s *DockerPortGuardService) runtimeStatus() dto.DockerPortGuardBase {
	ipv4 := s.runtime.Status(docker_guard.FamilyIPv4)
	ipv6 := s.runtime.Status(docker_guard.FamilyIPv6)
	return dto.DockerPortGuardBase{
		Initialized: ipv4.Initialized,
		Bound:       ipv4.Bound,
		IPv4:        dto.DockerPortGuardFamilyStatus{State: ipv4.State, Reason: ipv4.Reason, Initialized: ipv4.Initialized, Bound: ipv4.Bound, Effective: ipv4.Effective},
		IPv6:        dto.DockerPortGuardFamilyStatus{State: ipv6.State, Reason: ipv6.Reason, Initialized: ipv6.Initialized, Bound: ipv6.Bound, Effective: ipv6.Effective},
	}
}

func (s *DockerPortGuardService) checkBackend(ctx context.Context) error {
	cli, err := s.client()
	if err != nil {
		return fmt.Errorf("Docker is not running: %w", err)
	}
	defer cli.Close()
	info, err := cli.Info(ctx)
	if err != nil {
		return fmt.Errorf("Docker is not running: %w", err)
	}
	backend := dockerFirewallBackend(info)
	if backend == "nftables" {
		return errors.New("Docker native nftables backend is not supported")
	}
	return nil
}

func dockerFirewallBackend(info system.Info) string {
	if info.FirewallBackend == nil || info.FirewallBackend.Driver == "" {
		return "iptables"
	}
	return strings.ToLower(info.FirewallBackend.Driver)
}

func discoverDockerEndpoints(ctx context.Context, cli *client.Client) ([]dto.DockerPortGuardEndpoint, error) {
	containers, err := cli.ContainerList(ctx, containertypes.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	endpoints := make([]dto.DockerPortGuardEndpoint, 0)
	for _, item := range containers {
		name := strings.TrimPrefix(firstGuardString(item.Names), "/")
		compose := item.Labels[composeProjectLabel]
		application := ""
		if created, ok := item.Labels[composeCreatedBy]; ok && created == "Apps" {
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
		return normalizedDockerGuardPolicy{}, nil, errors.New("invalid Docker port guard policy")
	}
	addr, err := netip.ParseAddr(hostIP)
	if err != nil || (family == docker_guard.FamilyIPv4) != addr.Is4() {
		return normalizedDockerGuardPolicy{}, nil, errors.New("host IP does not match address family")
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
				return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("invalid source address %q", source)
			}
		}
		if (family == docker_guard.FamilyIPv4) != prefix.Addr().Is4() {
			return normalizedDockerGuardPolicy{}, nil, fmt.Errorf("source %q does not match address family", source)
		}
		canonical := prefix.Masked().String()
		if _, ok := seen[canonical]; !ok {
			seen[canonical] = struct{}{}
			normalizedSources = append(normalizedSources, canonical)
		}
	}
	if mode == docker_guard.ModeSources && len(normalizedSources) == 0 {
		return normalizedDockerGuardPolicy{}, nil, errors.New("deny_sources requires at least one source")
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
