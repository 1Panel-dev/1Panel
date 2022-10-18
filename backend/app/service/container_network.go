package service

import (
	"context"
	"fmt"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

func (u *ContainerService) PageNetwork(req dto.PageInfo) (int64, interface{}, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	list, err := client.NetworkList(context.TODO(), types.NetworkListOptions{})
	if err != nil {
		return 0, nil, err
	}
	var (
		data    []dto.Network
		records []types.NetworkResource
	)
	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]types.NetworkResource, 0)
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
func (u *ContainerService) DeleteNetwork(req dto.BatchDelete) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	for _, id := range req.Ids {
		if err := client.NetworkRemove(context.TODO(), id); err != nil {
			return err
		}
	}
	return nil
}
func (u *ContainerService) CreateNetwork(req dto.NetworkCreat) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	var (
		ipam    network.IPAMConfig
		hasConf bool
	)
	if len(req.Subnet) != 0 {
		ipam.Subnet = req.Subnet
		hasConf = true
	}
	if len(req.Gateway) != 0 {
		ipam.Gateway = req.Gateway
		hasConf = true
	}
	if len(req.IPRange) != 0 {
		ipam.IPRange = req.IPRange
		hasConf = true
	}

	options := types.NetworkCreate{
		Driver:  req.Driver,
		Options: stringsToMap(req.Options),
		Labels:  stringsToMap(req.Labels),
	}
	if hasConf {
		options.IPAM = &network.IPAM{Config: []network.IPAMConfig{ipam}}
	}
	if _, err := client.NetworkCreate(context.TODO(), req.Name, options); err != nil {
		return err
	}
	return nil
}
