package service

import (
	"errors"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/docker_guard"
)

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
			if _, _, err := normalizeGuardPolicy(test.family, test.hostIP, 80, "tcp", test.mode, test.sources); err == nil {
				t.Fatal("expected validation error")
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
