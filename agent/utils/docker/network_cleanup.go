package docker

import (
	"context"
	"sort"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/errdefs"
)

type NetworkCleanupClient interface {
	NetworkList(context.Context, network.ListOptions) ([]network.Inspect, error)
	ContainerList(context.Context, container.ListOptions) ([]container.Summary, error)
	NetworkInspect(context.Context, string, network.InspectOptions) (network.Inspect, error)
	NetworkRemove(context.Context, string) error
}

func CleanUnusedNetworks(ctx context.Context, cli NetworkCleanupClient, onResult ...func(string, dto.NetworkCleanupItem)) (*dto.NetworkCleanupReport, error) {
	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		return nil, err
	}
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	used := make(map[string]bool)
	for _, c := range containers {
		if c.NetworkSettings == nil {
			continue
		}
		for name, endpoint := range c.NetworkSettings.Networks {
			used[name] = true
			if endpoint != nil {
				used[endpoint.NetworkID] = true
			}
		}
	}
	report := &dto.NetworkCleanupReport{Deleted: []dto.NetworkCleanupItem{}, Skipped: []dto.NetworkCleanupItem{}, Failed: []dto.NetworkCleanupItem{}}
	record := func(status string, item dto.NetworkCleanupItem) {
		switch status {
		case "deleted":
			report.Deleted = append(report.Deleted, item)
		case "skipped":
			report.Skipped = append(report.Skipped, item)
		case "failed":
			report.Failed = append(report.Failed, item)
		}
		for _, notify := range onResult {
			if notify != nil {
				notify(status, item)
			}
		}
	}
	sort.Slice(networks, func(i, j int) bool { return networks[i].Name < networks[j].Name })
	for _, n := range networks {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		item := dto.NetworkCleanupItem{ID: n.ID, Name: n.Name}
		switch {
		case n.Name == "none" || n.Name == "host" || n.Name == "bridge" || n.Name == "1panel-network":
			item.Reason = "protected"
		case n.Scope != "local" || n.Ingress || n.ConfigOnly:
			item.Reason = "unsupported_network"
		case used[n.Name] || used[n.ID]:
			item.Reason = "container_connected"
		}
		if item.Reason != "" {
			record("skipped", item)
			continue
		}
		inspected, err := cli.NetworkInspect(ctx, n.ID, network.InspectOptions{})
		if err != nil {
			if errdefs.IsNotFound(err) {
				item.Reason = "already_removed"
				record("skipped", item)
			} else {
				item.Reason = "inspect_failed"
				record("failed", item)
			}
			continue
		}
		if len(inspected.Containers) > 0 {
			item.Reason = "container_connected"
			record("skipped", item)
			continue
		}
		if err := cli.NetworkRemove(ctx, n.ID); err != nil {
			switch {
			case errdefs.IsNotFound(err):
				item.Reason = "already_removed"
				record("skipped", item)
			case errdefs.IsConflict(err):
				item.Reason = "network_in_use"
				record("skipped", item)
			default:
				item.Reason = "remove_failed"
				record("failed", item)
			}
			continue
		}
		record("deleted", item)
	}
	return report, nil
}
