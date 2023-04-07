package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

func (u *ContainerService) PageVolume(req dto.SearchWithPage) (int64, interface{}, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	list, err := client.VolumeList(context.TODO(), filters.NewArgs())
	if err != nil {
		return 0, nil, err
	}
	if len(req.Info) != 0 {
		lenth, count := len(list.Volumes), 0
		for count < lenth {
			if !strings.Contains(list.Volumes[count].Name, req.Info) {
				list.Volumes = append(list.Volumes[:count], list.Volumes[(count+1):]...)
				lenth--
			} else {
				count++
			}
		}
	}
	var (
		data    []dto.Volume
		records []*volume.Volume
	)
	sort.Slice(list.Volumes, func(i, j int) bool {
		return list.Volumes[i].CreatedAt > list.Volumes[j].CreatedAt
	})
	total, start, end := len(list.Volumes), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		records = make([]*volume.Volume, 0)
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
		if len(item.CreatedAt) > 19 {
			item.CreatedAt = item.CreatedAt[0:19]
		}
		createTime, _ := time.Parse("2006-01-02T15:04:05", item.CreatedAt)
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
	for _, id := range req.Names {
		if err := client.VolumeRemove(context.TODO(), id, true); err != nil {
			if strings.Contains(err.Error(), "volume is in use") {
				return buserr.WithDetail(constant.ErrInUsed, id, nil)
			}
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
	var array []filters.KeyValuePair
	array = append(array, filters.Arg("name", req.Name))
	vos, _ := client.VolumeList(context.TODO(), filters.NewArgs(array...))
	if len(vos.Volumes) != 0 {
		for _, v := range vos.Volumes {
			if v.Name == req.Name {
				return constant.ErrRecordExist
			}
		}
	}
	options := volume.CreateOptions{
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
