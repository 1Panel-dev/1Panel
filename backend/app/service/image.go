package service

import (
	"bufio"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/homedir"
)

type ImageService struct{}

type IImageService interface {
	Page(req dto.SearchWithPage) (int64, interface{}, error)
	List() ([]dto.Options, error)
	ListAll() ([]dto.ImageInfo, error)
	ImageBuild(req dto.ImageBuild) (string, error)
	ImagePull(req dto.ImagePull) (string, error)
	ImageLoad(req dto.ImageLoad) error
	ImageSave(req dto.ImageSave) error
	ImagePush(req dto.ImagePush) (string, error)
	ImageRemove(req dto.BatchDelete) error
	ImageTag(req dto.ImageTag) error
}

func NewIImageService() IImageService {
	return &ImageService{}
}
func (u *ImageService) Page(req dto.SearchWithPage) (int64, interface{}, error) {
	var (
		list      []image.Summary
		records   []dto.ImageInfo
		backDatas []dto.ImageInfo
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return 0, nil, err
	}
	defer client.Close()
	list, err = client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return 0, nil, err
	}
	containers, _ := client.ContainerList(context.Background(), container.ListOptions{All: true})
	if len(req.Info) != 0 {
		length, count := len(list), 0
		for count < length {
			hasTag := false
			for _, tag := range list[count].RepoTags {
				if strings.Contains(tag, req.Info) {
					hasTag = true
					break
				}
			}
			if !hasTag {
				list = append(list[:count], list[(count+1):]...)
				length--
			} else {
				count++
			}
		}
	}

	for _, image := range list {
		size := formatFileSize(image.Size)
		records = append(records, dto.ImageInfo{
			ID:        image.ID,
			Tags:      image.RepoTags,
			IsUsed:    checkUsed(image.ID, containers),
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

func (u *ImageService) ListAll() ([]dto.ImageInfo, error) {
	var records []dto.ImageInfo
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err := client.ImageList(context.Background(), image.ListOptions{})
	if err != nil {
		return nil, err
	}
	containers, _ := client.ContainerList(context.Background(), container.ListOptions{All: true})
	for _, image := range list {
		size := formatFileSize(image.Size)
		records = append(records, dto.ImageInfo{
			ID:        image.ID,
			Tags:      image.RepoTags,
			IsUsed:    checkUsed(image.ID, containers),
			CreatedAt: time.Unix(image.Created, 0),
			Size:      size,
		})
	}
	return records, nil
}

func (u *ImageService) List() ([]dto.Options, error) {
	var (
		list      []image.Summary
		backDatas []dto.Options
	)
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	list, err = client.ImageList(context.Background(), image.ListOptions{})
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
	defer client.Close()
	fileName := "Dockerfile"
	if req.From == "edit" {
		dir := fmt.Sprintf("%s/docker/build/%s", constant.DataDir, strings.ReplaceAll(req.Name, ":", "_"))
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return "", err
			}
		}

		pathItem := fmt.Sprintf("%s/Dockerfile", dir)
		file, err := os.OpenFile(pathItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
		if err != nil {
			return "", err
		}
		defer file.Close()
		write := bufio.NewWriter(file)
		_, _ = write.WriteString(string(req.Dockerfile))
		write.Flush()
		req.Dockerfile = dir
	} else {
		fileName = path.Base(req.Dockerfile)
		req.Dockerfile = path.Dir(req.Dockerfile)
	}
	tar, err := archive.TarWithOptions(req.Dockerfile+"/", &archive.TarOptions{})
	if err != nil {
		return "", err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: fileName,
		Tags:       []string{req.Name},
		Remove:     true,
		Labels:     stringsToMap(req.Tags),
	}

	dockerLogDir := path.Join(global.CONF.System.TmpDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	logItem := fmt.Sprintf("%s/image_build_%s_%s.log", dockerLogDir, strings.ReplaceAll(req.Name, ":", "_"), time.Now().Format(constant.DateTimeSlimLayout))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	go func() {
		defer file.Close()
		defer tar.Close()
		res, err := client.ImageBuild(context.Background(), tar, opts)
		if err != nil {
			global.LOG.Errorf("build image %s failed, err: %v", req.Name, err)
			_, _ = file.WriteString("image build failed!")
			return
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			global.LOG.Errorf("build image %s failed, err: %v", req.Name, err)
			_, _ = file.WriteString(fmt.Sprintf("build image %s failed, err: %v", req.Name, err))
			_, _ = file.WriteString("image build failed!")
			return
		}

		if strings.Contains(string(body), "errorDetail") || strings.Contains(string(body), "error:") {
			global.LOG.Errorf("build image %s failed", req.Name)
			_, _ = file.Write(body)
			_, _ = file.WriteString("image build failed!")
			return
		}
		global.LOG.Infof("build image %s successful!", req.Name)
		_, _ = file.Write(body)
		_, _ = file.WriteString("image build successful!")
	}()

	return path.Base(logItem), nil
}

func (u *ImageService) ImagePull(req dto.ImagePull) (string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return "", err
	}
	defer client.Close()
	dockerLogDir := path.Join(global.CONF.System.TmpDir, "docker_logs")
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	imageItemName := strings.ReplaceAll(path.Base(req.ImageName), ":", "_")
	logItem := fmt.Sprintf("%s/image_pull_%s_%s.log", dockerLogDir, imageItemName, time.Now().Format(constant.DateTimeSlimLayout))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	options := image.PullOptions{}
	if req.RepoID == 0 {
		hasAuth, authStr := loadAuthInfo(req.ImageName)
		if hasAuth {
			options.RegistryAuth = authStr
		}
		go func() {
			defer file.Close()
			out, err := client.ImagePull(context.TODO(), req.ImageName, options)
			if err != nil {
				global.LOG.Errorf("image %s pull failed, err: %v", req.ImageName, err)
				return
			}
			defer out.Close()
			global.LOG.Infof("pull image %s successful!", req.ImageName)
			_, _ = io.Copy(file, out)
		}()
		return path.Base(logItem), nil
	}
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	if err != nil {
		return "", err
	}
	if repo.Auth {
		authConfig := registry.AuthConfig{
			Username: repo.Username,
			Password: repo.Password,
		}
		encodedJSON, err := json.Marshal(authConfig)
		if err != nil {
			return "", err
		}
		authStr := base64.StdEncoding.EncodeToString(encodedJSON)
		options.RegistryAuth = authStr
	}
	image := repo.DownloadUrl + "/" + req.ImageName
	go func() {
		defer file.Close()
		out, err := client.ImagePull(context.TODO(), image, options)
		if err != nil {
			_, _ = file.WriteString("image pull failed!")
			_, _ = file.WriteString(fmt.Sprintf("image %s pull failed, err: %v", image, err))
			return
		}
		defer out.Close()
		global.LOG.Infof("pull image %s successful!", req.ImageName)
		_, _ = io.Copy(file, out)
		_, _ = file.WriteString("image pull successful!")
	}()
	return path.Base(logItem), nil
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
	defer client.Close()
	res, err := client.ImageLoad(context.TODO(), file, true)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	content, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if strings.Contains(string(content), "Error") {
		return errors.New(string(content))
	}
	return nil
}

func (u *ImageService) ImageSave(req dto.ImageSave) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

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
	defer client.Close()

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
	defer client.Close()
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.RepoID))
	if err != nil {
		return "", err
	}
	options := image.PushOptions{All: true}
	authConfig := registry.AuthConfig{
		Username: repo.Username,
		Password: repo.Password,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return "", err
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	options.RegistryAuth = authStr
	newName := fmt.Sprintf("%s/%s", repo.DownloadUrl, req.Name)
	if newName != req.TagName {
		if err := client.ImageTag(context.TODO(), req.TagName, newName); err != nil {
			return "", err
		}
	}

	dockerLogDir := global.CONF.System.TmpDir + "/docker_logs"
	if _, err := os.Stat(dockerLogDir); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(dockerLogDir, os.ModePerm); err != nil {
			return "", err
		}
	}
	imageItemName := strings.ReplaceAll(path.Base(req.Name), ":", "_")
	logItem := fmt.Sprintf("%s/image_push_%s_%s.log", dockerLogDir, imageItemName, time.Now().Format(constant.DateTimeSlimLayout))
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return "", err
	}
	go func() {
		defer file.Close()
		out, err := client.ImagePush(context.TODO(), newName, options)
		if err != nil {
			global.LOG.Errorf("image %s push failed, err: %v", req.TagName, err)
			_, _ = file.WriteString("image push failed!")
			return
		}
		defer out.Close()
		global.LOG.Infof("push image %s successful!", req.Name)
		_, _ = io.Copy(file, out)
		_, _ = file.WriteString("image push successful!")
	}()

	return path.Base(logItem), nil
}

func (u *ImageService) ImageRemove(req dto.BatchDelete) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	for _, id := range req.Names {
		if _, err := client.ImageRemove(context.TODO(), id, image.RemoveOptions{Force: req.Force, PruneChildren: true}); err != nil {
			if strings.Contains(err.Error(), "image is being used") || strings.Contains(err.Error(), "is using") {
				if strings.Contains(id, "sha256:") {
					return buserr.New(constant.ErrObjectInUsed)
				}
				return buserr.WithDetail(constant.ErrInUsed, id, nil)
			}
			if strings.Contains(err.Error(), "image has dependent") {
				return buserr.New(constant.ErrObjectBeDependent)
			}
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

func checkUsed(imageID string, containers []types.Container) bool {
	for _, container := range containers {
		if container.ImageID == imageID {
			return true
		}
	}
	return false
}

func loadAuthInfo(image string) (bool, string) {
	if !strings.Contains(image, "/") {
		return false, ""
	}
	homeDir := homedir.Get()
	confPath := path.Join(homeDir, ".docker/config.json")
	configFileBytes, err := os.ReadFile(confPath)
	if err != nil {
		return false, ""
	}
	var config dockerConfig
	if err = json.Unmarshal(configFileBytes, &config); err != nil {
		return false, ""
	}
	var (
		user   string
		passwd string
	)
	imagePrefix := strings.Split(image, "/")[0]
	if val, ok := config.Auths[imagePrefix]; ok {
		itemByte, _ := base64.StdEncoding.DecodeString(val.Auth)
		itemStr := string(itemByte)
		if strings.Contains(itemStr, ":") {
			user = strings.Split(itemStr, ":")[0]
			passwd = strings.Split(itemStr, ":")[1]
		}
	}
	authConfig := registry.AuthConfig{
		Username: user,
		Password: passwd,
	}
	encodedJSON, err := json.Marshal(authConfig)
	if err != nil {
		return false, ""
	}
	authStr := base64.URLEncoding.EncodeToString(encodedJSON)
	return true, authStr
}

type dockerConfig struct {
	Auths map[string]authConfig `json:"auths"`
}
type authConfig struct {
	Auth string `json:"auth"`
}
