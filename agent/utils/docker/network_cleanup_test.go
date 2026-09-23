package docker

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
)

type cleanupFake struct {
	networks                    []network.Inspect
	containers                  []container.Summary
	inspected                   map[string]network.Inspect
	inspectErrors, removeErrors map[string]error
	listError                   error
	removed                     []string
}

func (f *cleanupFake) NetworkList(context.Context, network.ListOptions) ([]network.Inspect, error) {
	return f.networks, nil
}
func (f *cleanupFake) ContainerList(_ context.Context, o container.ListOptions) ([]container.Summary, error) {
	if !o.All {
		panic("must include stopped containers")
	}
	return f.containers, f.listError
}
func (f *cleanupFake) NetworkInspect(_ context.Context, id string, _ network.InspectOptions) (network.Inspect, error) {
	return f.inspected[id], f.inspectErrors[id]
}
func (f *cleanupFake) NetworkRemove(_ context.Context, id string) error {
	f.removed = append(f.removed, id)
	return f.removeErrors[id]
}

func TestNetworkCleanupProtectsAndReports(t *testing.T) {
	f := &cleanupFake{inspected: map[string]network.Inspect{"attached": {Containers: map[string]network.EndpointResource{"container": {}}}}, inspectErrors: map[string]error{"unknown": errors.New("inspect failure")}, removeErrors: map[string]error{"race": errdefs.Conflict(errors.New("has active endpoints")), "gone": errdefs.NotFound(errors.New("gone")), "failed": errors.New("denied")}}
	for _, name := range []string{"none", "host", "bridge", "1panel-network", "configured", "attached", "stopped", "free", "unknown", "race", "gone", "failed"} {
		f.networks = append(f.networks, network.Inspect{ID: name, Name: name, Scope: "local"})
	}
	// Compose ownership labels do not protect an otherwise unused network.
	for i := range f.networks {
		if f.networks[i].Name == "configured" {
			f.networks[i].Labels = map[string]string{"com.docker.compose.project": "demo", "com.docker.compose.network": "default"}
		}
	}
	f.containers = []container.Summary{{State: "exited", NetworkSettings: &container.NetworkSettingsSummary{Networks: map[string]*network.EndpointSettings{"stopped": {NetworkID: "stopped"}}}}}
	report, err := CleanUnusedNetworks(context.Background(), f)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(f.removed, []string{"configured", "failed", "free", "gone", "race"}) {
		t.Fatal(f.removed)
	}
	if len(report.Deleted) != 2 || report.Deleted[0].Name != "configured" || report.Deleted[1].Name != "free" || len(report.Failed) != 2 || len(report.Skipped) != 8 {
		t.Fatalf("%+v", report)
	}
}

func TestNetworkCleanupDiscoveryFailureDeletesNothing(t *testing.T) {
	f := &cleanupFake{networks: []network.Inspect{{ID: "free", Name: "free", Scope: "local"}}, listError: errors.New("cannot list containers")}
	if _, err := CleanUnusedNetworks(context.Background(), f); err == nil {
		t.Fatal("expected error")
	}
	if len(f.removed) > 0 {
		t.Fatal(f.removed)
	}
}
