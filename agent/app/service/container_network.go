package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
)

var networkCleanupSlot = make(chan struct{}, 1)

func (u *ContainerService) PageNetwork(req dto.SearchWithPage) (int64, interface{}, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	defer client.Close()
	list, err := client.NetworkList(context.TODO(), network.ListOptions{})
	if err != nil {
		return 0, nil, err
	}
	if len(req.Info) != 0 {
		length, count := len(list), 0
		for count < length {
			if !strings.Contains(list[count].Name, req.Info) {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}
	var (
		data    []dto.Network
		records []network.Inspect
	)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Created.Before(list[j].Created)
	})
	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]network.Inspect, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list[start:end]
	}

	for _, item := range records {
		tag := make([]string, 0)
		for key, val := range item.Labels {
			tag = append(tag, fmt.Sprintf("%s=%s", key, val))
		}
		var ipam network.IPAMConfig
		if len(item.IPAM.Config) > 0 {
			ipam = item.IPAM.Config[0]
		}
		data = append(data, dto.Network{
			ID:         item.ID,
			CreatedAt:  item.Created,
			Name:       item.Name,
			Driver:     item.Driver,
			IPAMDriver: item.IPAM.Driver,
			Subnet:     ipam.Subnet,
			Gateway:    ipam.Gateway,
			Attachable: item.Attachable,
			Labels:     tag,
		})
	}

	return int64(total), data, nil
}

func (u *ContainerService) ListNetwork() ([]dto.Options, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err := client.NetworkList(context.TODO(), network.ListOptions{})
	if err != nil {
		return nil, err
	}
	var datas []dto.Options
	for _, item := range list {
		datas = append(datas, dto.Options{Option: item.Name})
	}
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Option < datas[j].Option
	})
	return datas, nil
}

func (u *ContainerService) DeleteNetwork(req dto.BatchDelete) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	for _, id := range req.Names {
		if err := client.NetworkRemove(context.TODO(), id); err != nil {
			if strings.Contains(err.Error(), "has active endpoints") {
				return buserr.WithDetail("ErrInUsed", id, nil)
			}
			return err
		}
	}
	return nil
}
func (u *ContainerService) CreateNetwork(req dto.NetworkCreate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	var ipams []network.IPAMConfig

	if req.Ipv4 {
		var itemIpam network.IPAMConfig
		if len(req.AuxAddress) != 0 {
			itemIpam.AuxAddress = make(map[string]string)
		}
		if len(req.Subnet) != 0 {
			itemIpam.Subnet = req.Subnet
		}
		if len(req.Gateway) != 0 {
			itemIpam.Gateway = req.Gateway
		}
		if len(req.IPRange) != 0 {
			itemIpam.IPRange = req.IPRange
		}
		for _, addr := range req.AuxAddress {
			itemIpam.AuxAddress[addr.Key] = addr.Value
		}
		ipams = append(ipams, itemIpam)
	}
	if req.Ipv6 {
		var itemIpam network.IPAMConfig
		if len(req.AuxAddress) != 0 {
			itemIpam.AuxAddress = make(map[string]string)
		}
		if len(req.SubnetV6) != 0 {
			itemIpam.Subnet = req.SubnetV6
		}
		if len(req.GatewayV6) != 0 {
			itemIpam.Gateway = req.GatewayV6
		}
		if len(req.IPRangeV6) != 0 {
			itemIpam.IPRange = req.IPRangeV6
		}
		for _, addr := range req.AuxAddressV6 {
			itemIpam.AuxAddress[addr.Key] = addr.Value
		}
		ipams = append(ipams, itemIpam)
	}

	options := network.CreateOptions{
		EnableIPv6: &req.Ipv6,
		Driver:     req.Driver,
		Options:    stringsToMap(req.Options),
		Labels:     stringsToMap(req.Labels),
	}
	if len(ipams) != 0 {
		options.IPAM = &network.IPAM{Config: ipams}
	}
	if _, err := client.NetworkCreate(context.TODO(), req.Name, options); err != nil {
		return err
	}
	return nil
}

func cleanUnusedNetworks(t *task.Task, cli *client.Client) error {
	ctx := t.TaskCtx
	select {
	case networkCleanupSlot <- struct{}{}:
		defer func() { <-networkCleanupSlot }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	networks, err := cli.NetworkList(ctx, network.ListOptions{})
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
	deleted, skipped, failed := 0, 0, 0
	defer func() {
		t.Log(i18n.GetMsgWithMap("NetworkCleanupSummary", map[string]interface{}{
			"deleted": deleted, "skipped": skipped, "failed": failed,
		}))
	}()
	var used map[string]bool
	sort.Slice(networks, func(i, j int) bool { return networks[i].Name < networks[j].Name })
	for _, n := range networks {
		if err := ctx.Err(); err != nil {
			return err
		}
		switch n.Name {
		case "none", "host", "bridge", "1panel-network":
			skipped++
			continue
		}
		if n.Scope != "local" || n.Ingress || n.ConfigOnly {
			skipped++
			continue
		}
		values := map[string]interface{}{"name": n.Name, "id": n.ID}
		if used == nil {
			containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
			if err != nil {
				if ctx.Err() != nil {
					return ctx.Err()
				}
				return err
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
			skipped++
			t.Log(i18n.GetMsgWithMap("NetworkCleanupConnected", values))
			continue
		}
		inspected, err := cli.NetworkInspect(ctx, n.ID, network.InspectOptions{})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			if errdefs.IsNotFound(err) {
				skipped++
			} else {
				failed++
				t.Logf("Failed to inspect network [%s] (%s): %v", n.Name, n.ID, err)
			}
			continue
		}
		if len(inspected.Containers) > 0 {
			skipped++
			t.Log(i18n.GetMsgWithMap("NetworkCleanupConnected", values))
			continue
		}
		if err := cli.NetworkRemove(ctx, n.ID); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			switch {
			case errdefs.IsNotFound(err):
				skipped++
			case errdefs.IsConflict(err):
				skipped++
				t.Log(i18n.GetMsgWithMap("NetworkCleanupConnected", values))
			default:
				failed++
				t.Logf("Failed to remove network [%s] (%s): %v", n.Name, n.ID, err)
			}
			continue
		}
		deleted++
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if failed > 0 {
		return errors.New(i18n.GetMsgByKey("NetworkCleanupPartialFailure"))
	}
	return nil
}
