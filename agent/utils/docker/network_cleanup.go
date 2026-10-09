package docker

import (
	"context"
	"sort"
	"time"

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

type NetworkCleanupReport struct {
	Deleted int
	Skipped int
	Failed  int
}

type NetworkCleanupItem struct {
	ID     string
	Name   string
	Reason string
}

func CleanUnusedNetworks(ctx context.Context, cli NetworkCleanupClient, until time.Time, onResult func(NetworkCleanupItem)) (*NetworkCleanupReport, error) {
	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, err
	}
	var used map[string]bool
	report := &NetworkCleanupReport{}
	record := func(item NetworkCleanupItem) {
		switch item.Reason {
		case "":
			report.Deleted++
		case "inspect_failed", "remove_failed":
			report.Failed++
		default:
			report.Skipped++
		}
		if onResult != nil {
			onResult(item)
		}
	}
	sort.Slice(networks, func(i, j int) bool { return networks[i].Name < networks[j].Name })
	for _, n := range networks {
		if err := ctx.Err(); err != nil {
			return report, err
		}
		item := NetworkCleanupItem{ID: n.ID, Name: n.Name}
		switch {
		case n.Name == "none" || n.Name == "host" || n.Name == "bridge" || n.Name == "1panel-network":
			item.Reason = "protected"
		case n.Scope != "local" || n.Ingress || n.ConfigOnly:
			item.Reason = "unsupported_network"
		case !until.IsZero() && (n.Created.IsZero() || n.Created.After(until)):
			item.Reason = "recent"
		}
		if item.Reason != "" {
			record(item)
			continue
		}
		if used == nil {
			containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
			if err != nil {
				if ctx.Err() != nil {
					return report, ctx.Err()
				}
				return report, err
			}
			used = make(map[string]bool)
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
		}
		if used[n.Name] || used[n.ID] {
			item.Reason = "container_connected"
			record(item)
			continue
		}
		inspected, err := cli.NetworkInspect(ctx, n.ID, network.InspectOptions{})
		if err != nil {
			if ctx.Err() != nil {
				return report, ctx.Err()
			}
			if errdefs.IsNotFound(err) {
				item.Reason = "already_removed"
			} else {
				item.Reason = "inspect_failed"
			}
			record(item)
			continue
		}
		if len(inspected.Containers) > 0 {
			item.Reason = "container_connected"
			record(item)
			continue
		}
		if err := cli.NetworkRemove(ctx, n.ID); err != nil {
			if ctx.Err() != nil {
				return report, ctx.Err()
			}
			switch {
			case errdefs.IsNotFound(err):
				item.Reason = "already_removed"
			case errdefs.IsConflict(err):
				item.Reason = "container_connected"
			default:
				item.Reason = "remove_failed"
			}
			record(item)
			continue
		}
		record(item)
	}
	return report, ctx.Err()
}
