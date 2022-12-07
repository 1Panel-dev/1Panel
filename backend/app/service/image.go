package service

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/archive"
)

const dockerLogDir = constant.TmpDir + "/docker_logs"

type ImageService struct{}

type IImageService interface {
	Page(req dto.PageInfo) (int64, interface{}, error)
	List() ([]dto.Options, error)
	ImageBuild(req dto.ImageBuild) (string, error)
	ImagePull(req dto.ImagePull) (string, error)
	ImageLoad(req dto.ImageLoad) error
	ImageSave(req dto.ImageSave) error
	ImagePush(req dto.ImagePush) (string, error)
	ImageRemove(req dto.BatchDelete) error
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
		records = append(records, dto.ImageInfo{
			ID:        image.ID,
			Tags:      image.RepoTags,
			CreatedAt: time.Unix(image.Created, 0),
			Size:      size,
		})
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

func (u *ImageService) List() ([]dto.Options, error) {
	var (
		list      []types.ImageSummary
		backDatas []dto.Options
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	list, err = client.ImageList(context.Background(), types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	for _, image := range list {
		for _, tag := range image.RepoTags {
			backDatas = append(backDatas, dto.Options{
				Option: tag,
			})
		}
	}
	return backDatas, nil
}

func (u *ImageService) ImageBuild(req dto.ImageBuild) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	if req.From == "edit" {
		dir := fmt.Sprintf("%s/%s", constant.TmpDockerBuildDir, req.Name)
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		}

		path := fmt.Sprintf("%s/Dockerfile", dir)
		file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return "", err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(string(req.Dockerfile))
		write.Flush()
		req.Dockerfile = dir
	}
	tar, err := archive.TarWithOptions(req.Dockerfile+"/", &archive.TarOptions{})
	if err != nil {
		return "", err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: "Dockerfile",
		Tags:       []string{req.Name},
		Remove:     true,
		Labels:     stringsToMap(req.Tags),
	}
	logName := fmt.Sprintf("%s/build.log", req.Dockerfile)

	path := logName
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	go func() {
		defer file.Close()
		defer tar.Close()
		res, err := client.ImageBuild(context.TODO(), tar, opts)
		if err != nil {
			global.LOG.Errorf("build image %s failed, err: %v", req.Name, err)
			return
		}
		defer res.Body.Close()
		global.LOG.Debugf("build image %s successful!", req.Name)
		_, _ = io.Copy(file, res.Body)
	}()

	return logName, nil
}

func (u *ImageService) ImagePull(req dto.ImagePull) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	path := fmt.Sprintf("%s/image_pull_%s_%s.log", dockerLogDir, strings.ReplaceAll(req.ImageName, ":", "_"), time.Now().Format("20060102150405"))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	if req.RepoID == 0 {
		go func() {
			defer file.Close()
			out, err := client.ImagePull(context.TODO(), req.ImageName, types.ImagePullOptions{})
			if err != nil {
				global.LOG.Errorf("image %s pull failed, err: %v", req.ImageName, err)
				return
			}
			defer out.Close()
			global.LOG.Debugf("pull image %s successful!", req.ImageName)
			_, _ = io.Copy(file, out)
		}()
		return path, nil
	}
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	if err != nil {
		return "", err
	}
	options := types.ImagePullOptions{}
	if repo.Auth {
		authConfig := types.AuthConfig{
			Username: repo.Username,
			Password: repo.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return "", err
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		options.RegistryAuth = authStr
	}
	image := repo.DownloadUrl + "/" + req.ImageName
	go func() {
		defer file.Close()
		out, err := client.ImagePull(context.TODO(), image, options)
		if err != nil {
			global.LOG.Errorf("image %s pull failed, err: %v", image, err)
			return
		}
		defer out.Close()
		global.LOG.Debugf("pull image %s successful!", req.ImageName)
		_, _ = io.Copy(file, out)
	}()
	return path, nil
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
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}

	out, err := client.ImageSave(context.TODO(), []string{req.TagName})
	if err != nil {
		return err
	}
	defer out.Close()
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.tar", req.Path, req.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err = io.Copy(file, out); err != nil {
		return err
	}
	return nil
}

func (u *ImageService) ImageTag(req dto.ImageTag) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}

	if err := client.ImageTag(context.TODO(), req.SourceID, req.TargetName); err != nil {
		return err
	}
	return nil
}

func (u *ImageService) ImagePush(req dto.ImagePush) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	if err != nil {
		return "", err
	}
	options := types.ImagePushOptions{}
	if repo.Auth {
		authConfig := types.AuthConfig{
			Username: repo.Username,
			Password: repo.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return "", err
		}
		authStr := base64.URLEncoding.EncodeToString(encodedJSON)
		options.RegistryAuth = authStr
	}
	newName := fmt.Sprintf("%s/%s", repo.DownloadUrl, req.Name)
	if newName != req.TagName {
		if err := client.ImageTag(context.TODO(), req.TagName, newName); err != nil {
			return "", err
		}
	}

	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	path := fmt.Sprintf("%s/image_push_%s_%s.log", dockerLogDir, strings.ReplaceAll(req.Name, ":", "_"), time.Now().Format("20060102150405"))
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	go func() {
		defer file.Close()
		out, err := client.ImagePush(context.TODO(), newName, options)
		if err != nil {
			global.LOG.Errorf("image %s push failed, err: %v", req.TagName, err)
			return
		}
		defer out.Close()
		global.LOG.Debugf("push image %s successful!", req.Name)
		_, _ = io.Copy(file, out)
	}()

	return path, nil
}

func (u *ImageService) ImageRemove(req dto.BatchDelete) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	for _, ids := range req.Ids {
		if _, err := client.ImageRemove(context.TODO(), ids, types.ImageRemoveOptions{Force: true, PruneChildren: true}); err != nil {
			return err
		}
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
