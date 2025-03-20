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

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/homedir"
)

type ImageService struct{}

type IImageService interface {
	Page(req dto.SearchWithPage) (int64, interface{}, error)
	List() ([]dto.Options, error)
	ListAll() ([]dto.ImageInfo, error)
	ImageBuild(req dto.ImageBuild) error
	ImagePull(req dto.ImagePull) error
	ImageLoad(req dto.ImageLoad) error
	ImageSave(req dto.ImageSave) error
	ImagePush(req dto.ImagePush) error
	ImageRemove(req dto.BatchDelete) (dto.ContainerPruneReport, error)
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

func (u *ImageService) ImageBuild(req dto.ImageBuild) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	fileName := "Dockerfile"
	if req.From == "edit" {
		dir := fmt.Sprintf("%s/docker/build/%s", global.Dir.DataDir, strings.ReplaceAll(req.Name, ":", "_"))
		if _, err := os.Stat(dir); err != nil && os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return err
			}
		}

		pathItem := fmt.Sprintf("%s/Dockerfile", dir)
		file, err := os.OpenFile(pathItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, constant.FilePerm)
		if err != nil {
			return err
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
		return err
	}

	opts := types.ImageBuildOptions{
		Dockerfile: fileName,
		Tags:       []string{req.Name},
		Remove:     true,
		Labels:     stringsToMap(req.Tags),
	}
	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskBuild, task.TaskScopeImage, req.TaskID, 1)
	if err != nil {
		return fmt.Errorf("new task for image build failed, err: %v", err)
	}

	go func() {
		defer tar.Close()
		taskItem.AddSubTask(i18n.GetMsgByKey("ImageBuild"), func(t *task.Task) error {
			dockerCli := docker.NewClientWithExist(client)
			if err := dockerCli.BuildImageWithProcessAndOptions(taskItem, tar, opts); err != nil {
				return err
			}
			return nil
		}, nil)

		_ = taskItem.Execute()
	}()

	return nil
}

func (u *ImageService) ImagePull(req dto.ImagePull) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	imageItemName := strings.ReplaceAll(path.Base(req.ImageName), ":", "_")
	taskItem, err := task.NewTaskWithOps(imageItemName, task.TaskPull, task.TaskScopeImage, req.TaskID, 1)
	if err != nil {
		return fmt.Errorf("new task for image pull failed, err: %v", err)
	}
	go func() {
		taskItem.AddSubTask(i18n.GetWithName("ImagePull", req.ImageName), func(t *task.Task) error {
			options := image.PullOptions{}
			imageName := req.ImageName
			if req.RepoID == 0 {
				hasAuth, authStr := loadAuthInfo(req.ImageName)
				if hasAuth {
					options.RegistryAuth = authStr
				}
			} else {
				repo, err := imageRepoRepo.Get(repo.WithByID(req.RepoID))
				taskItem.LogWithStatus(i18n.GetMsgByKey("ImageRepoAuthFromDB"), err)
				if err != nil {
					return err
				}
				if repo.Auth {
					authConfig := registry.AuthConfig{
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
				imageName = repo.DownloadUrl + "/" + req.ImageName
			}
			dockerCli := docker.NewClientWithExist(client)
			err = dockerCli.PullImageWithProcessAndOptions(taskItem, imageName, options)
			taskItem.LogWithStatus(i18n.GetMsgByKey("TaskPull"), err)
			if err != nil {
				return err
			}
			return nil
		}, nil)
		_ = taskItem.Execute()
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
	file, err := os.OpenFile(fmt.Sprintf("%s/%s.tar", req.Path, req.Name), os.O_WRONLY|os.O_CREATE|os.O_EXCL, constant.FilePerm)
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

func (u *ImageService) ImagePush(req dto.ImagePush) error {
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()

	taskItem, err := task.NewTaskWithOps(req.Name, task.TaskPush, task.TaskScopeImage, req.TaskID, 1)
	if err != nil {
		return fmt.Errorf("new task for image push failed, err: %v", err)
	}

	go func() {
		options := image.PushOptions{All: true}
		var imageRepo model.ImageRepo
		newName := ""
		taskItem.AddSubTask(i18n.GetMsgByKey("ImagePush"), func(t *task.Task) error {
			imageRepo, err = imageRepoRepo.Get(repo.WithByID(req.RepoID))
			newName = fmt.Sprintf("%s/%s", imageRepo.DownloadUrl, req.Name)
			taskItem.LogWithStatus(i18n.GetMsgByKey("ImageRepoAuthFromDB"), err)
			if err != nil {
				return err
			}
			options = image.PushOptions{All: true}
			authConfig := registry.AuthConfig{
				Username: imageRepo.Username,
				Password: imageRepo.Password,
			}
			encodedJSON, _ := json.Marshal(authConfig)
			authStr := base64.URLEncoding.EncodeToString(encodedJSON)
			options.RegistryAuth = authStr
			return nil
		}, nil)
		taskItem.AddSubTask(i18n.GetMsgByKey("ImageRenameTag"), func(t *task.Task) error {
			taskItem.Log(i18n.GetWithName("ImageNewTag", newName))
			if newName != req.TagName {
				if err := client.ImageTag(context.TODO(), req.TagName, newName); err != nil {
					return err
				}
			}
			return nil
		}, nil)
		taskItem.AddSubTask(i18n.GetMsgByKey("TaskPush"), func(t *task.Task) error {
			dockerCli := docker.NewClientWithExist(client)
			if err := dockerCli.PushImageWithProcessAndOptions(taskItem, newName, options); err != nil {
				return err
			}
			return nil
		}, nil)
		_ = taskItem.Execute()
	}()

	return nil
}

func (u *ImageService) ImageRemove(req dto.BatchDelete) (dto.ContainerPruneReport, error) {
	report := dto.ContainerPruneReport{}
	client, err := docker.NewDockerClient()
	if err != nil {
		return report, err
	}
	defer client.Close()
	for _, id := range req.Names {
		imageItem, _, err := client.ImageInspectWithRaw(context.TODO(), id)
		if err != nil {
			return report, err
		}
		if _, err := client.ImageRemove(context.TODO(), id, image.RemoveOptions{Force: req.Force, PruneChildren: true}); err != nil {
			if strings.Contains(err.Error(), "image is being used") || strings.Contains(err.Error(), "is using") {
				if strings.Contains(id, "sha256:") {
					return report, buserr.New("ErrObjectInUsed")
				}
				return report, buserr.WithDetail("ErrInUsed", id, nil)
			}
			if strings.Contains(err.Error(), "image has dependent") {
				return report, buserr.New("ErrObjectBeDependent")
			}
			return report, err
		}
		report.DeletedNumber++
		report.SpaceReclaimed += int(imageItem.Size)
	}
	return report, nil
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
