package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/dockerhub"
	"github.com/docker/cli/cli/config"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const imageCheckUpdateSetting = "ContainerImageUpdateEnabled"

type ImageCheckUpdateStatus struct {
	ProxyRequired bool   `json:"proxyRequired"`
	TaskID        string `json:"taskID"`
	Enabled       bool   `json:"enabled"`
	Running       bool   `json:"running"`
	Checked       int    `json:"checked"`
	Updates       int    `json:"updates"`
	Failed        int    `json:"failed"`
	Skipped       int    `json:"skipped"`
	Error         string `json:"error"`
	FinishedAt    int64  `json:"finishedAt"`
}

var imageCheckUpdateState struct {
	sync.Mutex
	status ImageCheckUpdateStatus
}

func GetContainerImageCheckUpdateStatus() (ImageCheckUpdateStatus, error) {
	var setting model.Setting
	if err := global.DB.Where("key = ?", imageCheckUpdateSetting).First(&setting).Error; err != nil {
		return ImageCheckUpdateStatus{}, err
	}
	httpClient, proxyErr := dockerhub.NewHTTPClient()
	proxyRequired := false
	if proxyErr != nil {
		if business, ok := proxyErr.(buserr.BusinessError); ok && business.Msg == "ErrImageCheckProxyRequired" {
			proxyRequired = true
		} else {
			return ImageCheckUpdateStatus{}, proxyErr
		}
	}
	if httpClient != nil && httpClient.Transport != nil {
		httpClient.CloseIdleConnections()
	}
	imageCheckUpdateState.Lock()
	defer imageCheckUpdateState.Unlock()
	status := imageCheckUpdateState.status
	status.Enabled = setting.Value == "Enable"
	status.ProxyRequired = proxyRequired
	return status, nil
}
func SetContainerImageCheckUpdateSchedule(enabled bool) error {
	value := "Disable"
	if enabled {
		httpClient, err := dockerhub.NewHTTPClient()
		if err != nil {
			return err
		}
		if httpClient.Transport != nil {
			httpClient.CloseIdleConnections()
		}
		value = "Enable"
	}
	return global.DB.Model(&model.Setting{}).Where("key = ?", imageCheckUpdateSetting).Update("value", value).Error
}

// StartContainerImageCheckUpdate coalesces manual and scheduled runs on this node.
func StartContainerImageCheckUpdate() (string, error) {
	imageCheckUpdateState.Lock()
	defer imageCheckUpdateState.Unlock()
	if imageCheckUpdateState.status.Running {
		return imageCheckUpdateState.status.TaskID, nil
	}
	taskItem, err := task.NewTask(i18n.GetMsgByKey("ContainerImageCheckUpdate"), task.TaskScan, task.TaskScopeImage, "", 0)
	if err != nil {
		return "", err
	}
	// Persist before returning the ID so the log drawer can immediately find it.
	if err := global.TaskDB.Save(taskItem.Task).Error; err != nil {
		_ = taskItem.Logger.Out.(io.Closer).Close()
		taskItem.DeleteLogFile()
		if cancel, ok := global.LoadTaskCancel(taskItem.TaskID); ok {
			cancel()
		}
		global.RemoveTaskCancel(taskItem.TaskID)
		return "", err
	}
	imageCheckUpdateState.status = ImageCheckUpdateStatus{Running: true, TaskID: taskItem.TaskID}
	go func() {
		result := ImageCheckUpdateStatus{TaskID: taskItem.TaskID}
		defer func() {
			result.FinishedAt = time.Now().UnixMilli()
			imageCheckUpdateState.Lock()
			imageCheckUpdateState.status = result
			imageCheckUpdateState.Unlock()
		}()
		// The action owns its deadline so the task runner waits for cleanup and logs.
		taskItem.AddSubTaskWithOps(i18n.GetMsgByKey("ContainerImageCheckUpdate"), func(t *task.Task) (err error) {
			defer func() {
				if recovered := recover(); recovered != nil {
					err = fmt.Errorf("image update check failed: %v", recovered)
				}
				t.Log(i18n.GetMsgWithMap("ContainerImageCheckSummary", map[string]interface{}{
					"checked": result.Checked, "updates": result.Updates, "failed": result.Failed, "skipped": result.Skipped,
				}))
			}()
			ctx, cancel := context.WithTimeout(t.TaskCtx, 30*time.Minute)
			defer cancel()
			if err = runContainerImageCheckUpdate(ctx, &result, t.Log); err != nil {
				return err
			}
			if result.Failed > 0 {
				return fmt.Errorf("%s", i18n.GetMsgWithMap("ContainerImageCheckFailures", map[string]interface{}{"count": result.Failed}))
			}
			return nil
		}, nil, 0, 0)
		if err := taskItem.Execute(); err != nil {
			result.Error = err.Error()
			global.LOG.Errorf("container image check: %s", err)
		}
	}()
	return taskItem.TaskID, nil
}

// These caches live only for one check. Remote results do not depend on a
// container's local image ID, which is compared separately for every baseline.
type imageCheckUpdateRepository struct {
	client     imageCheckUpdateHubClient
	err        error
	tags       []string
	tagsLoaded bool
	tagsErr    error
}
type imageCheckUpdateTargetKey struct {
	repository, tag string
	platform        dockerhub.Platform
}
type imageCheckUpdateTargetResult struct {
	target     dockerhub.Target
	compatible bool
	err        error
}
type imageCheckUpdateResult struct {
	record model.ContainerImageCheckUpdate
	err    error
}
type imageCheckUpdateCache struct {
	username, password string
	httpClient         *http.Client
	repositories       map[string]*imageCheckUpdateRepository
	targets            map[imageCheckUpdateTargetKey]imageCheckUpdateTargetResult
}

func runContainerImageCheckUpdate(ctx context.Context, result *ImageCheckUpdateStatus, logger func(string)) error {
	httpClient, err := dockerhub.NewHTTPClient()
	if err != nil {
		return err
	}
	if httpClient.Transport != nil {
		defer httpClient.CloseIdleConnections()
	}
	log := func(key string, values map[string]interface{}) {
		if logger != nil {
			logger(i18n.GetMsgWithMap(key, values))
		}
	}
	cli, err := newImageCheckUpdateDockerClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return err
	}
	log("ContainerImageCheckScope", map[string]interface{}{"count": len(containers)})
	// Load old results once, rather than issuing one SELECT for every failed check.
	var previousRows []model.ContainerImageCheckUpdate
	if err := global.DB.WithContext(ctx).Find(&previousRows).Error; err != nil {
		return err
	}
	previous := make(map[string]model.ContainerImageCheckUpdate, len(previousRows))
	for _, row := range previousRows {
		previous[row.ContainerID] = row
	}
	cache := imageCheckUpdateCache{httpClient: httpClient, repositories: map[string]*imageCheckUpdateRepository{}, targets: map[imageCheckUpdateTargetKey]imageCheckUpdateTargetResult{}}
	cfg := config.LoadDefaultConfigFile(io.Discard)
	if cfg != nil {
		for _, host := range []string{"https://index.docker.io/v1/", "docker.io", "registry-1.docker.io"} {
			auth, e := cfg.GetAuthConfig(host)
			if e == nil && auth.Username != "" {
				cache.username, cache.password = auth.Username, auth.Password
				break
			}
		}
	}
	checks := map[string]imageCheckUpdateResult{}
	ids := make([]string, 0, len(containers))
	names := make(map[string]string, len(containers))
	pending := make([]model.ContainerImageCheckUpdate, 0, 100)
	for _, item := range containers {
		if err := ctx.Err(); err != nil {
			return err
		}
		ids = append(ids, item.ID)
		name := item.ID
		if len(item.Names) > 0 {
			name = strings.TrimPrefix(item.Names[0], "/")
		}
		names[item.ID] = name
		log("ContainerImageCheckItem", map[string]interface{}{"name": name, "image": item.Image})
		repository, tag, supported := dockerhub.Reference(item.Image)
		key := repository + ":" + tag + "@" + item.ImageID
		entry, found := checks[key]
		if !found {
			entry.record, entry.err = checkContainerImageUpdate(ctx, cli, item, repository, tag, supported, &cache)
			checks[key] = entry
		}
		record := entry.record
		record.ContainerID, record.ImageName = item.ID, item.Image
		if entry.err != nil {
			record.CheckStatus, record.CheckError = "failed", entry.err.Error()
			old := previous[item.ID]
			if old.ImageID == item.ImageID && old.ImageName == item.Image {
				record.LatestVersion, record.TargetImage, record.TargetImageID, record.HasUpdate = old.LatestVersion, old.TargetImage, old.TargetImageID, old.HasUpdate
			}
		}
		pending = append(pending, record)
		if len(pending) == 100 {
			if err := saveContainerImageCheckUpdates(ctx, cli, pending, names, result, log); err != nil {
				return err
			}
			pending = pending[:0]
		}
	}
	if err := saveContainerImageCheckUpdates(ctx, cli, pending, names, result, log); err != nil {
		return err
	}
	query := global.DB.WithContext(ctx).Where("1 = 1")
	if len(ids) > 0 {
		query = query.Where("container_id NOT IN ?", ids)
	}
	return query.Delete(&model.ContainerImageCheckUpdate{}).Error
}

func checkContainerImageUpdate(ctx context.Context, cli imageCheckUpdateDockerClient, item container.Summary, repository, tag string, supported bool, cache *imageCheckUpdateCache) (model.ContainerImageCheckUpdate, error) {
	record := model.ContainerImageCheckUpdate{ImageID: item.ImageID, ImageName: item.Image, CheckStatus: "success", CheckedAt: time.Now()}
	if !supported {
		record.CheckStatus = "skipped"
		return record, nil
	}
	inspected, err := cli.ImageInspect(ctx, item.ImageID)
	if err != nil {
		return record, err
	}
	repo, loaded := cache.repositories[repository]
	if !loaded {
		repo = &imageCheckUpdateRepository{}
		repo.client, repo.err = newImageCheckUpdateHubClient(ctx, repository, cache.username, cache.password, cache.httpClient)
		cache.repositories[repository] = repo
	}
	if repo.err != nil {
		return record, repo.err
	}
	if dockerhub.NeedsTags(tag) {
		if !repo.tagsLoaded {
			repo.tags, repo.tagsErr = repo.client.Tags(ctx)
			repo.tagsLoaded = true
		}
		if repo.tagsErr != nil {
			return record, repo.tagsErr
		}
	}
	platform := dockerhub.Platform{OS: inspected.Os, Architecture: inspected.Architecture, Variant: inspected.Variant}
	for _, candidate := range dockerhub.Candidates(tag, repo.tags) {
		key := imageCheckUpdateTargetKey{repository: repository, tag: candidate, platform: platform}
		remote, loaded := cache.targets[key]
		if !loaded {
			remote.target, remote.compatible, remote.err = repo.client.Resolve(ctx, candidate, platform)
			cache.targets[key] = remote
		}
		if remote.err != nil {
			return record, remote.err
		}
		if !remote.compatible {
			continue
		}
		record.LatestVersion = remote.target.Tag
		record.TargetImage = "docker.io/" + repository + ":" + remote.target.Tag
		record.TargetImageID = remote.target.ConfigDigest
		// Both versioned and floating tags can be republished with new image content.
		record.HasUpdate = remote.target.Tag != tag || remote.target.ConfigDigest != inspected.ID
		return record, nil
	}
	return record, fmt.Errorf("no compatible image for %s/%s", platform.OS, platform.Architecture)
}

// Revalidate immediately before each small write batch. Never hold a database
// transaction open while waiting for Docker or Docker Hub.
func saveContainerImageCheckUpdates(ctx context.Context, cli imageCheckUpdateDockerClient, pending []model.ContainerImageCheckUpdate, names map[string]string, result *ImageCheckUpdateStatus, log func(string, map[string]interface{})) error {
	records := make([]model.ContainerImageCheckUpdate, 0, len(pending))
	for _, record := range pending {
		current, err := cli.ContainerInspect(ctx, record.ContainerID)
		if err != nil && !errdefs.IsNotFound(err) {
			return err
		}
		if err != nil || current.Image != record.ImageID {
			result.Skipped++
			log("ContainerImageCheckChanged", map[string]interface{}{"name": names[record.ContainerID]})
			continue
		}
		records = append(records, record)
	}
	if len(records) == 0 {
		return nil
	}
	if err := global.DB.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).CreateInBatches(&records, 100).Error; err != nil {
		return err
	}
	// Count and announce only results that were successfully persisted.
	for _, record := range records {
		values := map[string]interface{}{"name": names[record.ContainerID], "image": record.ImageName}
		switch record.CheckStatus {
		case "failed":
			result.Failed++
			values["error"] = record.CheckError
			log("ContainerImageCheckFailed", values)
		case "skipped":
			result.Skipped++
			log("ContainerImageCheckSkipped", values)
		default:
			result.Checked++
			if record.HasUpdate {
				result.Updates++
				values["target"] = record.TargetImage
				log("ContainerImageCheckAvailable", values)
			} else {
				log("ContainerImageCheckCurrent", values)
			}
		}
	}
	return nil
}

// sameContainerImage guards cached results after external image changes.
func sameContainerImage(stored, displayed string) bool {
	return strings.TrimPrefix(stored, "sha256:") == strings.TrimPrefix(displayed, "sha256:")
}

// Reconcile only after Docker confirms the replacement container uses the target.
func reconcileContainerImageCheckUpdate(oldID, newID, imageID, imageName string) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		var previous model.ContainerImageCheckUpdate
		if err := tx.Where("container_id = ?", oldID).Find(&previous).Error; err != nil {
			return err
		}
		if previous.ContainerID == "" {
			return nil
		}
		if err := tx.Where("container_id = ?", oldID).Delete(&model.ContainerImageCheckUpdate{}).Error; err != nil {
			return err
		}
		if previous.TargetImageID != imageID {
			return nil
		}
		previous.ContainerID = newID
		previous.ImageID = imageID
		previous.ImageName = imageName
		previous.HasUpdate = false
		previous.CheckStatus = "success"
		previous.CheckError = ""
		previous.CheckedAt = time.Now()
		return tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&previous).Error
	})
}

type imageCheckUpdateDockerClient interface {
	ContainerList(context.Context, container.ListOptions) ([]container.Summary, error)
	ContainerInspect(context.Context, string) (container.InspectResponse, error)
	ImageInspect(context.Context, string, ...client.ImageInspectOption) (image.InspectResponse, error)
	Close() error
}
type imageCheckUpdateHubClient interface {
	Tags(context.Context) ([]string, error)
	Resolve(context.Context, string, dockerhub.Platform) (dockerhub.Target, bool, error)
}

var newImageCheckUpdateDockerClient = func() (imageCheckUpdateDockerClient, error) { return docker.NewDockerClient() }
var newImageCheckUpdateHubClient = func(ctx context.Context, repository, username, password string, httpClient *http.Client) (imageCheckUpdateHubClient, error) {
	return dockerhub.New(ctx, repository, username, password, httpClient)
}
