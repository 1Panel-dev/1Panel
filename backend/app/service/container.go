package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerService struct{}

type IContainerService interface {
	Page(req dto.PageContainer) (int64, interface{}, error)
	PageNetwork(req dto.PageInfo) (int64, interface{}, error)
	PageVolume(req dto.PageInfo) (int64, interface{}, error)
	ContainerOperation(req dto.ContainerOperation) error
	ContainerLogs(param dto.ContainerLog) (string, error)
	ContainerInspect(id string) (string, error)
	DeleteNetwork(req dto.BatchDelete) error
	CreateNetwork(req dto.NetworkCreat) error
	DeleteVolume(req dto.BatchDelete) error
	CreateVolume(req dto.VolumeCreat) error
}

func NewIContainerService() IContainerService {
	return &ContainerService{}
}

func (u *ContainerService) Page(req dto.PageContainer) (int64, interface{}, error) {
	var (
		records   []types.Container
		list      []types.Container
		backDatas []dto.ContainerInfo
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	list, err = client.ContainerList(context.Background(), types.ContainerListOptions{All: req.Status == "all"})
	if err != nil {
		return 0, nil, err
	}
	total, start, end := len(list), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]types.Container, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list[start:end]
	}

	for _, container := range records {
		backDatas = append(backDatas, dto.ContainerInfo{
			ContainerID: container.ID,
			CreateTime:  time.Unix(container.Created, 0).Format("2006-01-02 15:04:05"),
			Name:        container.Names[0][1:],
			ImageId:     strings.Split(container.ImageID, ":")[1],
			ImageName:   container.Image,
			State:       container.State,
			RunTime:     container.Status,
		})
	}

	return int64(total), backDatas, nil
}

func (u *ContainerService) ContainerOperation(req dto.ContainerOperation) error {
	var err error
	ctx := context.Background()
	dc, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	switch req.Operation {
	case constant.ContainerOpStart:
		err = dc.ContainerStart(ctx, req.ContainerID, types.ContainerStartOptions{})
	case constant.ContainerOpStop:
		err = dc.ContainerStop(ctx, req.ContainerID, nil)
	case constant.ContainerOpRestart:
		err = dc.ContainerRestart(ctx, req.ContainerID, nil)
	case constant.ContainerOpKill:
		err = dc.ContainerKill(ctx, req.ContainerID, "SIGKILL")
	case constant.ContainerOpPause:
		err = dc.ContainerPause(ctx, req.ContainerID)
	case constant.ContainerOpUnpause:
		err = dc.ContainerUnpause(ctx, req.ContainerID)
	case constant.ContainerOpRename:
		err = dc.ContainerRename(ctx, req.ContainerID, req.NewName)
	case constant.ContainerOpRemove:
		err = dc.ContainerRemove(ctx, req.ContainerID, types.ContainerRemoveOptions{RemoveVolumes: true, RemoveLinks: true, Force: true})
	}
	return err
}

func (u *ContainerService) ContainerInspect(id string) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	inspect, err := client.ContainerInspect(context.Background(), id)
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(inspect)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

func (u *ContainerService) ContainerLogs(req dto.ContainerLog) (string, error) {
	var (
		options types.ContainerLogsOptions
		logs    io.ReadCloser
		buf     *bytes.Buffer
		err     error
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	options = types.ContainerLogsOptions{
		ShowStdout: true,
		Timestamps: true,
	}
	if req.Mode != "all" {
		options.Since = req.Mode
	}
	if logs, err = client.ContainerLogs(context.Background(), req.ContainerID, options); err != nil {
		return "", err
	}
	defer logs.Close()
	buf = new(bytes.Buffer)
	if _, err = stdcopy.StdCopy(buf, nil, logs); err != nil {
		return "", err
	}
	return buf.String(), nil
}

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
		var (
			ipv4 network.IPAMConfig
			ipv6 network.IPAMConfig
		)
		if len(item.IPAM.Config) > 1 {
			ipv4 = item.IPAM.Config[0]
			ipv6 = item.IPAM.Config[1]
		} else if len(item.IPAM.Config) > 0 {
			ipv4 = item.IPAM.Config[0]
		}
		data = append(data, dto.Network{
			ID:          item.ID,
			CreatedAt:   item.Created,
			Name:        item.Name,
			Driver:      item.Driver,
			IPAMDriver:  item.IPAM.Driver,
			IPV4Subnet:  ipv4.Subnet,
			IPV4Gateway: ipv4.Gateway,
			IPV6Subnet:  ipv6.Subnet,
			IPV6Gateway: ipv6.Gateway,
			Attachable:  item.Attachable,
			Labels:      tag,
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
	ipv4 := network.IPAMConfig{
		Subnet:  req.IPV4Subnet,
		Gateway: req.IPV4Gateway,
	}
	options := types.NetworkCreate{
		Driver: req.Driver,
		Scope:  req.Scope,
		IPAM: &network.IPAM{
			Config: []network.IPAMConfig{ipv4},
		},
		Options: stringsToMap(req.Options),
		Labels:  stringsToMap(req.Labels),
	}
	if _, err := client.NetworkCreate(context.TODO(), req.Name, options); err != nil {
		return err
	}
	return nil
}

func (u *ContainerService) PageVolume(req dto.PageInfo) (int64, interface{}, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	list, err := client.VolumeList(context.TODO(), filters.NewArgs())
	if err != nil {
		return 0, nil, err
	}
	var (
		data    []dto.Volume
		records []*types.Volume
	)
	total, start, end := len(list.Volumes), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]*types.Volume, 0)
	} else {
		if end >= total {
			end = total
		}
		records = list.Volumes[start:end]
	}

	for _, item := range records {
		tag := make([]string, 0)
		for _, val := range item.Labels {
			tag = append(tag, val)
		}
		createTime, _ := time.Parse("2006-01-02T15:04:05Z", item.CreatedAt)
		data = append(data, dto.Volume{
			CreatedAt:  createTime,
			Name:       item.Name,
			Driver:     item.Driver,
			Mountpoint: item.Mountpoint,
			Labels:     tag,
		})
	}

	return int64(total), data, nil
}
func (u *ContainerService) DeleteVolume(req dto.BatchDelete) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	for _, id := range req.Ids {
		if err := client.VolumeRemove(context.TODO(), id, true); err != nil {
			return err
		}
	}
	return nil
}
func (u *ContainerService) CreateVolume(req dto.VolumeCreat) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	options := volume.VolumeCreateBody{
		Name:       req.Name,
		Driver:     req.Driver,
		DriverOpts: stringsToMap(req.Options),
		Labels:     stringsToMap(req.Labels),
	}
	if _, err := client.VolumeCreate(context.TODO(), options); err != nil {
		return err
	}
	return nil
}

func stringsToMap(list []string) map[string]string {
	var lableMap = make(map[string]string)
	for _, label := range list {
		sps := strings.Split(label, "=")
		if len(sps) > 1 {
			lableMap[sps[0]] = sps[1]
		}
	}
	return lableMap
}
