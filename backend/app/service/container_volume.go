package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
)

func (u *ContainerService) PageVolume(req dto.SearchWithPage) (int64, interface{}, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	list, err := client.VolumeList(context.TODO(), volume.ListOptions{})
	if err != nil {
		return 0, nil, err
	}
	if len(req.Info) != 0 {
		length, count := len(list.Volumes), 0
		for count < length {
			if !strings.Contains(list.Volumes[count].Name, req.Info) {
				list.Volumes = append(list.Volumes[:count], list.Volumes[(count+1):]...)
				length--
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

	nyc, _ := time.LoadLocation(common.LoadTimeZoneByCmd())
	for _, item := range records {
		tag := make([]string, 0)
		for _, val := range item.Labels {
			tag = append(tag, val)
		}
		var createTime time.Time
		if strings.Contains(item.CreatedAt, "Z") {
			createTime, _ = time.ParseInLocation("2006-01-02T15:04:05Z", item.CreatedAt, nyc)
		} else if strings.Contains(item.CreatedAt, "+") {
			createTime, _ = time.ParseInLocation("2006-01-02T15:04:05+08:00", item.CreatedAt, nyc)
		} else {
			createTime, _ = time.ParseInLocation("2006-01-02T15:04:05", item.CreatedAt, nyc)
		}
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
	defer client.Close()
	list, err := client.VolumeList(context.TODO(), volume.ListOptions{})
	if err != nil {
		return nil, err
	}
	var datas []dto.Options
	for _, item := range list.Volumes {
		datas = append(datas, dto.Options{
			Option: item.Name,
		})
	}
	sort.Slice(datas, func(i, j int) bool {
		return datas[i].Option < datas[j].Option
	})
	return datas, nil
}
func (u *ContainerService) DeleteVolume(req dto.BatchDelete) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
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
func (u *ContainerService) CreateVolume(req dto.VolumeCreate) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	arg := filters.NewArgs()
	arg.Add("name", req.Name)
	vos, _ := client.VolumeList(context.TODO(), volume.ListOptions{Filters: arg})
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
