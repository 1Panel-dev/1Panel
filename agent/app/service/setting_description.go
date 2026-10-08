package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/filter"
	"github.com/1Panel-dev/1Panel/agent/utils/firewall/lifecycle"
	"github.com/docker/docker/api/types/container"
)

func (u *SettingService) CleanupDescriptions(ctx context.Context) (int64, error) {
	var deleted int64
	var failures []error
	for _, kind := range []string{"container", "firewall", "firewall-docker"} {
		count, err := func() (int64, error) {
			switch kind {
			case "container":
				return cleanupUnusedDescriptions(ctx, kind, loadContainerDescriptionIDs)
			case "firewall":
				firewallRuleMutationMu.Lock()
				defer firewallRuleMutationMu.Unlock()
				return cleanupUnusedDescriptions(ctx, kind, loadHostFirewallDescriptionIDs)
			default:
				return cleanupUnusedDescriptions(ctx, kind, nil)
			}
		}()
		deleted += count
		if err != nil {
			failures = append(failures, fmt.Errorf("%s: %w", kind, err))
		}
	}
	return deleted, errors.Join(failures...)
}

func cleanupUnusedDescriptions(ctx context.Context, kind string, loadIDs func(context.Context) (map[string]bool, error)) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	descriptions, err := settingRepo.GetDescriptionList(repo.WithByType(kind))
	if err != nil || len(descriptions) == 0 {
		return 0, err
	}
	var active map[string]bool
	if loadIDs != nil {
		active, err = loadIDs(ctx)
		if err != nil {
			return 0, err
		}
	}
	ids, empty := make([]string, 0), make([]string, 0)
	for _, description := range descriptions {
		if !active[description.ID] {
			ids = append(ids, description.ID)
		} else if description.Description == "" && !description.IsPinned {
			empty = append(empty, description.ID)
		}
	}
	deleted, err := settingRepo.DeleteDescriptions(ctx, kind, ids, false)
	if err != nil {
		return deleted, err
	}
	count, err := settingRepo.DeleteDescriptions(ctx, kind, empty, true)
	return deleted + count, err
}

func loadContainerDescriptionIDs(ctx context.Context) (map[string]bool, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	containers, err := client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	ids := make(map[string]bool, len(containers))
	for _, item := range containers {
		ids[item.ID] = true
	}
	return ids, nil
}

func loadHostFirewallDescriptionIDs(ctx context.Context) (map[string]bool, error) {
	providers := lifecycle.InstalledProviders()
	if len(providers) == 0 {
		return nil, filter.ErrProviderUnavailable
	}
	service := newFirewallService()
	ids := make(map[string]bool)
	for _, name := range providers {
		provider := filter.Provider(name)
		inventory, err := service.readFirewallInventory(ctx, provider, filter.ManagedInputScopes(provider))
		if err != nil {
			return nil, err
		}
		for _, notice := range inventory.Notices {
			if notice.Code == filter.ScopeNoticeFamilyUnavailable || notice.Code == filter.ScopeNoticeManagedScopeInactive {
				return nil, fmt.Errorf("%w: %s %s", filter.ErrInventoryUnavailable, name, notice.Code)
			}
		}
		for _, item := range inventory.Items {
			id, err := filter.DescriptionID(*item.Observed)
			if err != nil {
				return nil, err
			}
			ids[id] = true
		}
	}
	return ids, nil
}
