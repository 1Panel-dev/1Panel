package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/docker"
	"github.com/docker/docker/api/types"
)

type ImageService struct{}

type IImageService interface {
	Page(req dto.PageInfo) (int64, interface{}, error)
	ImagePull(req dto.ImagePull) error
	ImageLoad(req dto.ImageLoad) error
	ImageSave(req dto.ImageSave) error
	ImagePush(req dto.ImagePush) error
	ImageRemove(req dto.ImageRemove) error
}

func NewIImageService() IImageService {
	return &ImageService{}
}
func (u *ImageService) Page(req dto.PageInfo) (int64, interface{}, error) {
	var (
		list      []types.ImageSummary
		records   []dto.ImageInfo
		backDatas []dto.ImageInfo
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	list, err = client.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return 0, nil, err
	}

	for _, image := range list {
		size := formatFileSize(image.Size)
		for _, item := range image.RepoTags {
			name := item[0:strings.LastIndex(item, ":")]
			tag := strings.ReplaceAll(item[strings.LastIndex(item, ":"):], ":", "")
			records = append(records, dto.ImageInfo{
				ID:        image.ID,
				Name:      name,
				Version:   tag,
				CreatedAt: time.Unix(image.Created, 0),
				Size:      size,
			})
		}
	}
	total, start, end := len(records), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backDatas = make([]dto.ImageInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		backDatas = records[start:end]
	}

	return int64(total), backDatas, nil
}

func (u *ImageService) ImagePull(req dto.ImagePull) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	if err != nil {
		return err
	}
	options := types.ImagePullOptions{}
	if repo.Auth {
		authConfig := types.AuthConfig{
			Username: repo.Username,
			Password: repo.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return err
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		options.RegistryAuth = authStr
	}
	image := repo.DownloadUrl + "/" + req.ImageName
	if len(repo.RepoName) != 0 {
		image = fmt.Sprintf("%s/%s/%s", repo.DownloadUrl, repo.RepoName, req.ImageName)
	}
	go func() {
		out, err := client.ImagePull(ctx, image, options)
		if err != nil {
			global.LOG.Errorf("image %s pull failed, err: %v", image, err)
			return
		}
		defer out.Close()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(out)
		global.LOG.Debugf("image %s pull stdout: %v", image, buf.String())
	}()
	return nil
}

func (u *ImageService) ImageLoad(req dto.ImageLoad) error {
	file, err := os.Open(req.Path)
	if err != nil {
		return err
	}
	defer file.Close()
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	if _, err := client.ImageLoad(context.TODO(), file, true); err != nil {
		return err
	}
	return nil
}

func (u *ImageService) ImageSave(req dto.ImageSave) error {
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.tar", req.Path, req.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}

	out, err := client.ImageSave(context.TODO(), []string{req.ImageName})
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err = io.Copy(file, out); err != nil {
		return err
	}
	return nil
}

func (u *ImageService) ImagePush(req dto.ImagePush) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	if err != nil {
		return err
	}
	options := types.ImagePushOptions{}
	if repo.Auth {
		authConfig := types.AuthConfig{
			Username: repo.Username,
			Password: repo.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return err
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		options.RegistryAuth = authStr
	}
	newName := fmt.Sprintf("%s/%s", repo.DownloadUrl, req.TagName)
	if err := client.ImageTag(context.TODO(), req.ImageName, newName); err != nil {
		return err
	}
	go func() {
		out, err := client.ImagePush(context.TODO(), newName, options)
		if err != nil {
			global.LOG.Errorf("image %s push failed, err: %v", req.ImageName, err)
			return
		}
		defer out.Close()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(out)
		global.LOG.Debugf("image %s push stdout: %v", req.ImageName, buf.String())
	}()

	return nil
}

func (u *ImageService) ImageRemove(req dto.ImageRemove) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	if _, err := client.ImageRemove(context.TODO(), req.ImageName, types.ImageRemoveOptions{Force: true}); err != nil {
		return err
	}
	return nil
}

func formatFileSize(fileSize int64) (size string) {
	if fileSize < 1024 {
		return fmt.Sprintf("%.2fB", float64(fileSize)/float64(1))
	} else if fileSize < (1024 * 1024) {
		return fmt.Sprintf("%.2fKB", float64(fileSize)/float64(1024))
	} else if fileSize < (1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fMB", float64(fileSize)/float64(1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fGB", float64(fileSize)/float64(1024*1024*1024))
	} else if fileSize < (1024 * 1024 * 1024 * 1024 * 1024) {
		return fmt.Sprintf("%.2fTB", float64(fileSize)/float64(1024*1024*1024*1024))
	} else {
		return fmt.Sprintf("%.2fEB", float64(fileSize)/float64(1024*1024*1024*1024*1024))
	}
}
