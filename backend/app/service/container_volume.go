package service

import (
	"context"
	"sort"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

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
	sort.Slice(list.Volumes, func(i, j int) bool {
		return list.Volumes[i].CreatedAt > list.Volumes[j].CreatedAt
	})
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
func (u *ContainerService) ListVolume() ([]dto.Options, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	list, err := client.VolumeList(context.TODO(), filters.NewArgs())
	if err != nil {
		return nil, err
	}
	var data []dto.Options
	for _, item := range list.Volumes {
		data = append(data, dto.Options{
			Option: item.Name,
		})
	}
	return data, nil
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
