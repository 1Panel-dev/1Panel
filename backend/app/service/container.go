package service

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerService struct{}

type IContainerService interface {
	Page(req dto.PageContainer) (int64, interface{}, error)
	ContainerOperation(req dto.ContainerOperation) error
	ContainerLogs(param dto.ContainerLog) (string, error)
	ContainerInspect(id string) (string, error)
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
