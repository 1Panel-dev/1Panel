package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

func (u *ContainerService) PageNetwork(req dto.SearchWithPage) (int64, interface{}, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	defer client.Close()
	list, err := client.NetworkList(context.TODO(), types.NetworkListOptions{})
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
		records []types.NetworkResource
	)
	sort.Slice(list, func(i, j int) bool {
		return list[i].Created.Before(list[j].Created)
	})
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

func (u *ContainerService) ListNetwork() ([]dto.Options, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err := client.NetworkList(context.TODO(), types.NetworkListOptions{})
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
				return buserr.WithDetail(constant.ErrInUsed, id, nil)
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
	var (
		ipams    []network.IPAMConfig
		enableV6 bool
	)
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
		enableV6 = true
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

	options := types.NetworkCreate{
		EnableIPv6: enableV6,
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
