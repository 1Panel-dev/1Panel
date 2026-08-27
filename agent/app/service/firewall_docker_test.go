package service

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	agenti18n "github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/docker/docker/client"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type persistentDockerGuardRuntime struct {
	initialized bool
	initialize  int
	reconcile   int
	bind        int
	unbind      int
	policies    []docker_guard.Policy
	statuses    map[string]docker_guard.FamilyStatus
}

func (r *persistentDockerGuardRuntime) Initialize(policies []docker_guard.Policy) error {
	r.initialize++
	r.initialized = true
	r.policies = append([]docker_guard.Policy(nil), policies...)
	r.markFamiliesEffective(policies)
	return nil
}

func (r *persistentDockerGuardRuntime) Bind() error {
	r.bind++
	r.initialized = true
	for family, status := range r.statuses {
		status.Initialized, status.Bound, status.Effective = true, true, true
		r.statuses[family] = status
	}
	return nil
}

func (r *persistentDockerGuardRuntime) Reconcile(policies []docker_guard.Policy) error {
	r.reconcile++
	r.policies = append([]docker_guard.Policy(nil), policies...)
	return nil
}

func (r *persistentDockerGuardRuntime) markFamiliesEffective(policies []docker_guard.Policy) {
	if r.statuses == nil {
		return
	}
	for _, policy := range policies {
		r.statuses[policy.Family] = docker_guard.FamilyStatus{Initialized: true, Bound: true, Effective: true}
	}
}

func (r *persistentDockerGuardRuntime) Unbind() error {
	r.unbind++
	return nil
}

func (r *persistentDockerGuardRuntime) Cleanup() error { return nil }

func (r *persistentDockerGuardRuntime) Initialized(string) (bool, error) {
	return r.initialized, nil
}

func (r *persistentDockerGuardRuntime) Status(family string) docker_guard.FamilyStatus {
	if r.statuses != nil {
		return r.statuses[family]
	}
	return docker_guard.FamilyStatus{Initialized: r.initialized, Bound: r.initialized, Effective: r.initialized}
}

func (r *persistentDockerGuardRuntime) ListPolicies() ([]docker_guard.Policy, error) {
	return append([]docker_guard.Policy(nil), r.policies...), nil
}

type persistentDockerGuardPolicies struct{ items []model.DockerPortGuardPolicy }

func (r *persistentDockerGuardPolicies) List(context.Context) ([]model.DockerPortGuardPolicy, error) {
	return append([]model.DockerPortGuardPolicy(nil), r.items...), nil
}

func (r *persistentDockerGuardPolicies) DeleteBatch(context.Context, []string) error { return nil }

func (r *persistentDockerGuardPolicies) UpsertBatch(context.Context, []model.DockerPortGuardPolicy) error {
	return nil
}

func TestDockerGuardOverviewLocalizesUnavailableDocker(t *testing.T) {
	agenti18n.Init()
	service := &DockerPortGuardService{
		policies: &persistentDockerGuardPolicies{items: []model.DockerPortGuardPolicy{{
			UUID: "orphan-policy", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080,
			Protocol: "tcp", Mode: docker_guard.ModeAll, Sources: "[]",
		}}},
		runtime: &persistentDockerGuardRuntime{},
		version: func(string) string { return "1.8.10" },
		client: func() (*client.Client, error) {
			return nil, errors.New("Cannot connect to the Docker daemon at unix:///var/run/docker.sock")
		},
	}
	overview, err := service.LoadOverview(context.Background())
	if err != nil {
		t.Fatalf("load overview: %v", err)
	}
	if overview.Base.Message != agenti18n.Get("ErrDockerFailed") {
		t.Fatalf("message = %q, want localized Docker failure", overview.Base.Message)
	}
	if overview.Base.Version != "1.8.10" {
		t.Fatalf("version = %q, want 1.8.10", overview.Base.Version)
	}
	if len(overview.Containers) != 0 || len(overview.OrphanPolicies) != 1 {
		t.Fatalf("overview = %#v, want persisted policy returned separately from Docker containers", overview)
	}
	if overview.OrphanPolicies[0].PolicyUUID != "orphan-policy" {
		t.Fatalf("orphan policy = %#v", overview.OrphanPolicies[0])
	}
}

func TestDockerGuardRuntimeStatusAggregatesAvailableFamilies(t *testing.T) {
	runtime := &persistentDockerGuardRuntime{statuses: map[string]docker_guard.FamilyStatus{
		docker_guard.FamilyIPv6: {Initialized: true, Bound: true, Effective: true},
	}}
	base := (&DockerPortGuardService{}).runtimeStatus(runtime, constant.FirewallProviderNftables)
	if !base.Initialized || !base.Bound || base.IPv4.Initialized || !base.IPv6.Initialized {
		t.Fatalf("unexpected aggregate Docker guard status: %#v", base)
	}
}

func TestDockerGuardRuntimeStatusReportsMissingBackend(t *testing.T) {
	runtime := &persistentDockerGuardRuntime{statuses: map[string]docker_guard.FamilyStatus{
		docker_guard.FamilyIPv4: {Reason: docker_guard.ReasonCommandMissing},
		docker_guard.FamilyIPv6: {Reason: docker_guard.ReasonCommandMissing},
	}}
	base := (&DockerPortGuardService{}).runtimeStatus(runtime, constant.FirewallProviderIptables)
	if base.IsExist {
		t.Fatalf("missing Docker firewall backend reported as installed: %#v", base)
	}
}

func TestMatchDockerGuardPoliciesReturnsUnmatchedDatabaseRules(t *testing.T) {
	policies := []model.DockerPortGuardPolicy{
		{UUID: "matched", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080, Protocol: "tcp", Mode: docker_guard.ModeAll, Sources: "[]"},
		{UUID: "orphan", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 9090, Protocol: "tcp", Mode: docker_guard.ModeSources, Sources: `["203.0.113.0/24"]`},
	}
	endpoints := []dto.DockerPortGuardEndpoint{{
		Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080, Protocol: "tcp", ContainerID: "container-1",
	}}
	matched, orphanPolicies := matchDockerGuardPolicies(dto.DockerPortGuardBase{}, policies, endpoints)
	if len(matched) != 1 || matched[0].PolicyUUID != "matched" {
		t.Fatalf("matched endpoints = %#v", matched)
	}
	if len(orphanPolicies) != 1 || orphanPolicies[0].PolicyUUID != "orphan" || orphanPolicies[0].HostPort != 9090 {
		t.Fatalf("orphan policies = %#v", orphanPolicies)
	}
}

func TestDockerGuardRuleSyncInitializesTargetWithPersistedPolicies(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	selectDockerGuardBackend(t, constant.FirewallProviderNftables)
	target := &persistentDockerGuardRuntime{}
	policies := &persistentDockerGuardPolicies{items: []model.DockerPortGuardPolicy{{
		UUID: "policy-1", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080,
		Protocol: "tcp", Mode: docker_guard.ModeSources, Sources: `["203.0.113.0/24"]`,
	}}}
	service := &DockerPortGuardService{
		policies:          policies,
		runtimeForBackend: func(string) dockerGuardRuntime { return target },
	}
	request := dto.FirewallRuleSyncRequest{Subsystem: "docker", TargetProvider: filter.ProviderNftables}
	preview, err := service.previewRuleSync(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Total != 1 || preview.Ready != 1 || preview.TargetProvider != filter.ProviderNftables || preview.Items[0].DockerRule == nil {
		t.Fatalf("unexpected preview: %#v", preview)
	}
	result, err := service.syncRules(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if result.Succeeded != 1 || result.Failed != 0 || target.initialize != 1 || len(target.policies) != 1 {
		t.Fatalf("unexpected sync result=%#v target=%#v", result, target)
	}
	retry, err := service.previewRuleSync(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if retry.Ready != 0 || retry.Existing != 1 || retry.Removed != 0 {
		t.Fatalf("synchronized Docker policy was not recognized: %#v", retry)
	}
	retryResult, err := service.syncRules(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if retryResult.Succeeded != 0 || retryResult.Skipped != 1 || retryResult.Removed != 0 {
		t.Fatalf("existing Docker policy was not counted as skipped: %#v", retryResult)
	}
}

func TestDockerGuardRuleSyncReconcilesInitializedEffectiveTarget(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	selectDockerGuardBackend(t, constant.FirewallProviderNftables)
	policy := model.DockerPortGuardPolicy{
		UUID: "policy-1", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080,
		Protocol: "tcp", Mode: docker_guard.ModeAll, Sources: "[]",
	}
	target := &persistentDockerGuardRuntime{initialized: true}
	service := &DockerPortGuardService{
		policies:          &persistentDockerGuardPolicies{items: []model.DockerPortGuardPolicy{policy}},
		runtimeForBackend: func(string) dockerGuardRuntime { return target },
	}
	result, err := service.syncRules(context.Background(), dto.FirewallRuleSyncRequest{
		Subsystem: "docker", TargetProvider: filter.ProviderNftables,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Succeeded != 1 || result.Skipped != 0 || target.reconcile != 1 || len(target.policies) != 1 {
		t.Fatalf("initialized target skipped runtime reconciliation: result=%#v target=%#v", result, target)
	}
}

func TestDockerGuardRuleSyncRebindsInitializedIneffectiveTarget(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	selectDockerGuardBackend(t, constant.FirewallProviderNftables)
	policy := model.DockerPortGuardPolicy{
		UUID: "policy-1", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080,
		Protocol: "tcp", Mode: docker_guard.ModeAll, Sources: "[]",
	}
	target := &persistentDockerGuardRuntime{
		initialized: true,
		statuses: map[string]docker_guard.FamilyStatus{
			docker_guard.FamilyIPv4: {Initialized: true},
		},
	}
	service := &DockerPortGuardService{
		policies:          &persistentDockerGuardPolicies{items: []model.DockerPortGuardPolicy{policy}},
		runtimeForBackend: func(string) dockerGuardRuntime { return target },
	}
	result, err := service.syncRules(context.Background(), dto.FirewallRuleSyncRequest{
		Subsystem: "docker", TargetProvider: filter.ProviderNftables,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Succeeded != 1 || target.bind != 1 || target.reconcile != 1 ||
		!target.Status(docker_guard.FamilyIPv4).Effective {
		t.Fatalf("initialized target was not rebound and reconciled: result=%#v target=%#v", result, target)
	}
}

func TestDockerGuardRuleSyncClearsInitializedTargetWhenDatabaseIsEmpty(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	selectDockerGuardBackend(t, constant.FirewallProviderNftables)
	target := &persistentDockerGuardRuntime{
		initialized: true,
		policies: []docker_guard.Policy{{
			UUID: "stale-policy", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0", HostPort: 8080,
			Protocol: "tcp", Mode: docker_guard.ModeAll,
		}},
	}
	service := &DockerPortGuardService{
		policies:          &persistentDockerGuardPolicies{},
		runtimeForBackend: func(string) dockerGuardRuntime { return target },
	}
	preview, err := service.previewRuleSync(context.Background(), dto.FirewallRuleSyncRequest{
		Subsystem: "docker", TargetProvider: filter.ProviderNftables,
	})
	if err != nil {
		t.Fatal(err)
	}
	if preview.Total != 0 || preview.Removed != 1 || len(preview.Items) != 1 || preview.Items[0].Status != "remove" {
		t.Fatalf("stale runtime policy was not included in preview: %#v", preview)
	}
	result, err := service.syncRules(context.Background(), dto.FirewallRuleSyncRequest{
		Subsystem: "docker", TargetProvider: filter.ProviderNftables,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 || result.Removed != 1 || target.initialize != 0 || target.reconcile != 1 || len(target.policies) != 0 {
		t.Fatalf("empty database did not clear initialized target: result=%#v target=%#v", result, target)
	}
}

func TestDockerGuardRuleSyncLeavesUninitializedTargetEmpty(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	selectDockerGuardBackend(t, constant.FirewallProviderNftables)
	target := &persistentDockerGuardRuntime{}
	service := &DockerPortGuardService{
		policies:          &persistentDockerGuardPolicies{},
		runtimeForBackend: func(string) dockerGuardRuntime { return target },
	}
	result, err := service.syncRules(context.Background(), dto.FirewallRuleSyncRequest{
		Subsystem: "docker", TargetProvider: filter.ProviderNftables,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Total != 0 || target.initialize != 0 || target.reconcile != 0 {
		t.Fatalf("empty database initialized an unused target: result=%#v target=%#v", result, target)
	}
}

func TestDockerGuardRuleSyncRejectsUnselectedTarget(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	selectDockerGuardBackend(t, constant.FirewallProviderIptables)
	target := &persistentDockerGuardRuntime{}
	service := &DockerPortGuardService{
		policies:          &persistentDockerGuardPolicies{},
		runtimeForBackend: func(string) dockerGuardRuntime { return target },
	}
	request := dto.FirewallRuleSyncRequest{Subsystem: "docker", TargetProvider: filter.ProviderNftables}
	if _, err := service.previewRuleSync(context.Background(), request); !errors.Is(err, filter.ErrProviderUnavailable) {
		t.Fatalf("preview error = %v, want provider unavailable", err)
	}
	if _, err := service.syncRules(context.Background(), request); !errors.Is(err, filter.ErrProviderUnavailable) {
		t.Fatalf("sync error = %v, want provider unavailable", err)
	}
	if target.initialize != 0 || target.bind != 0 || target.reconcile != 0 {
		t.Fatalf("unselected target was modified: %#v", target)
	}
}

func selectDockerGuardBackend(t *testing.T, backend string) {
	t.Helper()
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerBackendKey, backend); err != nil {
		t.Fatal(err)
	}
}

func setupDockerGuardSettingsDB(t *testing.T) {
	t.Helper()
	previousDB := global.DB
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.NewString())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatalf("open settings database: %v", err)
	}
	if err := db.AutoMigrate(&model.Setting{}); err != nil {
		t.Fatalf("migrate settings database: %v", err)
	}
	global.DB = db
	t.Cleanup(func() { global.DB = previousDB })
}

func TestNormalizeDockerPortGuardPolicy(t *testing.T) {
	policy, sources, err := normalizeGuardPolicy("ipv4", "0.0.0.0", 8080, "TCP", "deny_sources", []string{"203.0.113.10", "192.0.2.7/24", "203.0.113.10/32"})
	if err != nil {
		t.Fatal(err)
	}
	if policy.Protocol != "tcp" {
		t.Fatalf("protocol was not normalized: %#v", policy)
	}
	want := []string{"192.0.2.0/24", "203.0.113.10/32"}
	if !reflect.DeepEqual(sources, want) {
		t.Fatalf("sources = %#v, want %#v", sources, want)
	}
}

func TestNormalizeDockerPortGuardPolicyRejectsInvalidSources(t *testing.T) {
	for _, test := range []struct {
		name    string
		family  string
		hostIP  string
		mode    string
		sources []string
	}{
		{name: "empty deny sources", family: "ipv4", hostIP: "0.0.0.0", mode: "deny_sources"},
		{name: "mixed source family", family: "ipv4", hostIP: "0.0.0.0", mode: "deny_sources", sources: []string{"2001:db8::/64"}},
		{name: "mixed host family", family: "ipv6", hostIP: "0.0.0.0", mode: "deny_all"},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, _, err := normalizeGuardPolicy(test.family, test.hostIP, 80, "tcp", test.mode, test.sources); !errors.Is(err, ErrDockerGuardInvalid) {
				t.Fatalf("expected typed validation error, got %v", err)
			}
		})
	}
}

func TestNormalizeDockerPortGuardPolicyAllowsEmptyAllowList(t *testing.T) {
	policy, sources, err := normalizeGuardPolicy("ipv4", "0.0.0.0", 5432, "tcp", "allow_sources", nil)
	if err != nil {
		t.Fatal(err)
	}
	if policy.Mode != "allow_sources" || len(sources) != 0 {
		t.Fatalf("policy = %#v, sources = %#v", policy, sources)
	}
}

func TestDockerGuardEndpointKeyIncludesAddressFamilyAndProtocol(t *testing.T) {
	first := guardEndpointKey("ipv4", "0.0.0.0", 53, "udp")
	if first == guardEndpointKey("ipv6", "::", 53, "udp") || first == guardEndpointKey("ipv4", "0.0.0.0", 53, "tcp") {
		t.Fatal("endpoint identity collapsed distinct endpoint dimensions")
	}
}

func TestNormalizeDockerGuardPolicyUUIDs(t *testing.T) {
	got, err := normalizeDockerGuardPolicyUUIDs([]string{" first ", "second", "first"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"first", "second"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("UUIDs = %#v, want %#v", got, want)
	}
	if _, err := normalizeDockerGuardPolicyUUIDs([]string{""}); err == nil {
		t.Fatal("expected empty UUID to be rejected")
	}
}

func TestGroupDockerGuardContainersMergesCompatiblePorts(t *testing.T) {
	endpoints := []dto.DockerPortGuardEndpoint{
		{Family: "ipv4", HostIP: "0.0.0.0", HostPort: 8001, Protocol: "tcp", ContainerID: "container-1", ContainerName: "demo", ContainerPort: 81, Mode: "deny_all", PolicyUUID: "policy-2", Effective: true, Sources: []string{}},
		{Family: "ipv4", HostIP: "0.0.0.0", HostPort: 8000, Protocol: "tcp", ContainerID: "container-1", ContainerName: "demo", ContainerPort: 80, Mode: "deny_all", PolicyUUID: "policy-1", Effective: true, Sources: []string{}},
		{Family: "ipv4", HostIP: "0.0.0.0", HostPort: 8002, Protocol: "tcp", ContainerID: "container-1", ContainerName: "demo", ContainerPort: 82, Mode: "allow_sources", PolicyUUID: "policy-3", Effective: true, Sources: []string{"192.0.2.1/32"}},
	}
	containers := groupDockerGuardContainers(endpoints)
	if len(containers) != 1 || len(containers[0].PortGroups) != 2 {
		t.Fatalf("containers = %#v, want one container with two port groups", containers)
	}
	foundRange := false
	for _, group := range containers[0].PortGroups {
		foundRange = foundRange || group.Label == "0.0.0.0:8000-8001/tcp"
	}
	if !foundRange {
		t.Fatalf("port groups = %#v, expected merged range", containers[0].PortGroups)
	}
	for _, group := range containers[0].PortGroups {
		if group.Label == "0.0.0.0:8000-8001/tcp" && len(group.Endpoints) != 2 {
			t.Fatalf("merged group endpoints = %#v, want 2 endpoints", group.Endpoints)
		}
	}
}

func TestMarkDockerGuardReconcileFailureOnlyAffectsFailedFamily(t *testing.T) {
	base := dto.DockerPortGuardBase{
		IPv4: dto.DockerPortGuardFamilyStatus{State: docker_guard.StatusEffective, Initialized: true, Bound: true, Effective: true},
		IPv6: dto.DockerPortGuardFamilyStatus{State: docker_guard.StatusEffective, Initialized: true, Bound: true, Effective: true},
	}
	markDockerGuardReconcileFailure(&base, &docker_guard.FamilyError{Family: docker_guard.FamilyIPv6, Err: errors.New("restore failed")})
	if !base.IPv4.Effective || base.IPv4.State != docker_guard.StatusEffective {
		t.Fatalf("IPv4 status changed unexpectedly: %#v", base.IPv4)
	}
	if base.IPv6.Effective || base.IPv6.State != docker_guard.StatusNotEffective || base.IPv6.Reason != docker_guard.ReasonInspectFailed {
		t.Fatalf("IPv6 status = %#v, want not effective", base.IPv6)
	}
}

func TestMarkDockerGuardIPv4FailureAlsoMarksUnattemptedIPv6(t *testing.T) {
	base := dto.DockerPortGuardBase{
		IPv4: dto.DockerPortGuardFamilyStatus{State: docker_guard.StatusEffective, Initialized: true, Bound: true, Effective: true},
		IPv6: dto.DockerPortGuardFamilyStatus{State: docker_guard.StatusEffective, Initialized: true, Bound: true, Effective: true},
	}
	markDockerGuardReconcileFailure(&base, &docker_guard.FamilyError{Family: docker_guard.FamilyIPv4, Err: errors.New("restore failed")})
	if base.IPv4.Effective || base.IPv6.Effective {
		t.Fatalf("statuses = IPv4 %#v, IPv6 %#v; both must be not effective", base.IPv4, base.IPv6)
	}
}

func TestDockerGuardReconcileErrorState(t *testing.T) {
	t.Cleanup(func() { recordDockerPortGuardReconcileError(nil) })
	want := errors.New("restore failed")
	recordDockerPortGuardReconcileError(want)
	if got := lastDockerPortGuardReconcileError(); !errors.Is(got, want) {
		t.Fatalf("last reconcile error = %v, want %v", got, want)
	}
	recordDockerPortGuardReconcileError(nil)
	if got := lastDockerPortGuardReconcileError(); got != nil {
		t.Fatalf("last reconcile error = %v, want nil", got)
	}
}

func TestDockerFirewallDisplayName(t *testing.T) {
	for backend, want := range map[string]string{
		"iptables": "iptables-docker",
		"nftables": "nftables-docker",
		"":         "iptables-docker",
	} {
		if got := dockerFirewallDisplayName(backend); got != want {
			t.Fatalf("dockerFirewallDisplayName(%q) = %q, want %q", backend, got, want)
		}
	}
}

func TestDockerGuardRuntimeMatchesDockerBackend(t *testing.T) {
	service := &DockerPortGuardService{}
	if _, ok := service.guardRuntime("iptables").(*docker_guard.Manager); !ok {
		t.Fatal("iptables backend did not select the iptables Docker guard runtime")
	}
	if _, ok := service.guardRuntime("nftables").(*docker_guard.NftablesManager); !ok {
		t.Fatal("nftables backend did not select the nftables Docker guard runtime")
	}
}

func TestDockerGuardInitializeAndUnbindPersistStatus(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	runtime := &persistentDockerGuardRuntime{}
	service := &DockerPortGuardService{
		policies: &persistentDockerGuardPolicies{},
		runtime:  runtime,
	}
	if err := service.Operate(context.Background(), dto.DockerPortGuardOperation{Operation: "initialize"}); err != nil {
		t.Fatalf("initialize Docker port guard: %v", err)
	}
	status, err := settingRepo.GetValueByKey(constant.FirewallDockerPortGuardStatusKey)
	if err != nil || status != constant.StatusEnable {
		t.Fatalf("persisted status = %q, %v; want %q", status, err, constant.StatusEnable)
	}
	if runtime.initialize != 1 {
		t.Fatalf("initialize calls = %d, want 1", runtime.initialize)
	}

	if err := service.Operate(context.Background(), dto.DockerPortGuardOperation{Operation: "unbind"}); err != nil {
		t.Fatalf("unbind Docker port guard: %v", err)
	}
	status, err = settingRepo.GetValueByKey(constant.FirewallDockerPortGuardStatusKey)
	if err != nil || status != constant.StatusDisable {
		t.Fatalf("persisted status = %q, %v; want %q", status, err, constant.StatusDisable)
	}
}

func TestDockerGuardReconcileRestoresPersistedInitialization(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusEnable); err != nil {
		t.Fatal(err)
	}
	runtime := &persistentDockerGuardRuntime{}
	service := &DockerPortGuardService{
		policies: &persistentDockerGuardPolicies{items: []model.DockerPortGuardPolicy{{
			UUID: "policy-1", Family: docker_guard.FamilyIPv4, HostIP: "0.0.0.0",
			HostPort: 8080, Protocol: "tcp", Mode: docker_guard.ModeAll, Sources: "[]",
		}}},
		runtime: runtime,
	}
	if err := service.Reconcile(context.Background()); err != nil {
		t.Fatalf("restore Docker port guard: %v", err)
	}
	if runtime.initialize != 1 || !runtime.initialized {
		t.Fatalf("runtime was not initialized: %#v", runtime)
	}
	if len(runtime.policies) != 1 || runtime.policies[0].UUID != "policy-1" {
		t.Fatalf("restored policies = %#v", runtime.policies)
	}
}

func TestDockerGuardReconcileLeavesDisabledGuardUninitialized(t *testing.T) {
	setupDockerGuardSettingsDB(t)
	if err := settingRepo.UpdateOrCreate(constant.FirewallDockerPortGuardStatusKey, constant.StatusDisable); err != nil {
		t.Fatal(err)
	}
	runtime := &persistentDockerGuardRuntime{}
	service := &DockerPortGuardService{policies: &persistentDockerGuardPolicies{}, runtime: runtime}
	if err := service.Reconcile(context.Background()); err != nil {
		t.Fatalf("reconcile disabled Docker port guard: %v", err)
	}
	if runtime.initialize != 0 || runtime.initialized {
		t.Fatalf("disabled guard was initialized: %#v", runtime)
	}
}
