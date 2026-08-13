package docker

import (
	"reflect"
	"testing"

	containertypes "github.com/docker/docker/api/types/container"
)

func TestSimplifyPortsMergesConsecutiveMappings(t *testing.T) {
	ports := []containertypes.Port{
		{IP: "0.0.0.0", PublicPort: 8002, PrivatePort: 82, Type: "tcp"},
		{IP: "0.0.0.0", PublicPort: 8000, PrivatePort: 80, Type: "tcp"},
		{IP: "0.0.0.0", PublicPort: 8001, PrivatePort: 81, Type: "tcp"},
		{IP: "0.0.0.0", PublicPort: 9000, PrivatePort: 90, Type: "udp"},
	}
	want := []string{"0.0.0.0:8000-8002->80-82/tcp", "0.0.0.0:9000->90/udp"}
	if got := SimplifyPorts(ports); !reflect.DeepEqual(got, want) {
		t.Fatalf("SimplifyPorts() = %#v, want %#v", got, want)
	}
}

func TestMergePortRangesKeepsDifferentIdentitiesSeparate(t *testing.T) {
	items := []PortRangeItem{
		{Key: "policy-a", PublicPort: 80},
		{Key: "policy-b", PublicPort: 81},
	}
	if got := MergePortRanges(items); len(got) != 2 {
		t.Fatalf("MergePortRanges() returned %d ranges, want 2", len(got))
	}
}
