package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/subosito/gotenv"
)

type appUpgradePhase int

const (
	appUpgradePreparing appUpgradePhase = iota
	appUpgradePrepared
	appUpgradeStopped
	appUpgradeBackedUp
	appUpgradeDown
	appUpgradeMutated
	appUpgradeStarted
	appUpgradeReady
	appUpgradeCommitted
)

const composeServiceLabel = "com.docker.compose.service"

var appUpgradeLocks sync.Map

type appUpgradeSnapshot interface {
	Restore() error
	Cleanup()
}

type upgradeFileSnapshot struct {
	installPath string
	backupPath  string
	paths       []string
	existing    map[string]bool
}

type appUpgradeContext struct {
	req       request.AppInstallUpgrade
	original  model.AppInstall
	candidate model.AppInstall
	detail    model.AppDetail

	phase         appUpgradePhase
	stopAttempted bool
	downAttempted bool
	rollbackErr   error

	detailDir        string
	stageDir         string
	envContent       []byte
	oldEnvContent    []byte
	oldDockerCompose string
	oldImageIDs      []appImageID
	backupFile       string
	snapshot         appUpgradeSnapshot
	createdPaths     []string
}

func upgradeInstall(req request.AppInstallUpgrade) error {
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.InstallID))
	if err != nil {
		return err
	}
	if install.Status == constant.StatusUpgrading {
		return buserr.New("TaskIsExecuting")
	}
	if err = task.CheckScopeTaskIsExecuting(task.TaskScopeApp, install.ID); err != nil {
		return err
	}
	if _, loaded := appUpgradeLocks.LoadOrStore(install.ID, struct{}{}); loaded {
		return buserr.New("TaskIsExecuting")
	}
	releaseLock := true
	defer func() {
		if releaseLock {
			appUpgradeLocks.Delete(install.ID)
		}
	}()

	detail, err := appDetailRepo.GetFirst(repo.WithByID(req.DetailID))
	if err != nil {
		return err
	}
	if install.App.Key == vllmAppKeyForUpgrade && !isVllmUpgradeVersionAllowed(install.Version, detail.Version, loadVllmImageFromEnv(install.Env)) {
		return errors.New("vLLM can only upgrade within the same image type")
	}
	if install.Version == detail.Version {
		return errors.New("two version is same")
	}

	upgradeTask, err := task.NewTaskWithOps(install.Name, task.TaskUpgrade, task.TaskScopeApp, req.TaskID, install.ID)
	if err != nil {
		return err
	}
	ctx := &appUpgradeContext{
		req:              req,
		original:         install,
		candidate:        install,
		detail:           detail,
		phase:            appUpgradePreparing,
		oldDockerCompose: install.DockerCompose,
	}
	upgradeTask.AddSubTaskWithOps(i18n.GetMsgByKey("UpgradePrepare"), ctx.prepare, nil, 0, 0)
	upgradeTask.AddSubTaskWithOps(
		task.GetTaskName(install.Name, task.TaskUpgrade, task.TaskScopeApp),
		ctx.cutover,
		func(t *task.Task) {
			ctx.rollbackErr = ctx.rollback(t)
		},
		0,
		0,
	)

	upgradingInstall := install
	upgradingInstall.Status = constant.StatusUpgrading
	upgradingInstall.Message = ""
	if err = appInstallRepo.Save(context.Background(), &upgradingInstall); err != nil {
		return err
	}

	releaseLock = false
	go func() {
		defer appUpgradeLocks.Delete(install.ID)
		defer ctx.cleanup()
		taskErr := upgradeTask.Execute()
		if taskErr == nil {
			return
		}
		if ctx.rollbackErr != nil {
			taskErr = fmt.Errorf("%w; %s: %v", taskErr, i18n.GetMsgByKey("UpgradeRollbackFailed"), ctx.rollbackErr)
			upgradeTask.Task.ErrorMsg = taskErr.Error()
			_ = repo.NewITaskRepo().Update(context.Background(), upgradeTask.Task)
		}
		if !ctx.stopAttempted || ctx.rollbackErr == nil {
			restored := ctx.original
			_ = appInstallRepo.Save(context.Background(), &restored)
			return
		}
		failed := ctx.original
		failed.Status = constant.StatusUpgradeErr
		failed.Message = taskErr.Error()
		_ = appInstallRepo.Save(context.Background(), &failed)
	}()
	return nil
}

func (u *appUpgradeContext) prepare(t *task.Task) error {
	fileOp := files.NewFileOp()
	u.detailDir = path.Join(u.original.App.GetAppResourcePath(), u.detail.Version)
	if u.original.App.Resource == constant.AppResourceRemote {
		if err := downloadApp(u.original.App, u.detail, nil, t.Logger); err != nil {
			return err
		}
	}
	if !fileOp.Stat(u.detailDir) {
		return buserr.WithName("ErrFileNotFound", u.detailDir)
	}
	if u.detail.DockerCompose == "" {
		composeContent, err := fileOp.GetContent(path.Join(u.detailDir, "docker-compose.yml"))
		if err != nil {
			return err
		}
		u.detail.DockerCompose = string(composeContent)
		_ = appDetailRepo.Update(context.Background(), u.detail)
	}
	if strings.TrimSpace(u.detail.DockerCompose) == "" && strings.TrimSpace(u.req.DockerCompose) == "" {
		return buserr.WithName("ErrFileNotFound", "docker-compose.yml")
	}

	var err error
	u.oldEnvContent, err = fileOp.GetContent(u.original.GetEnvPath())
	if err != nil {
		return err
	}
	u.stageDir, err = os.MkdirTemp(u.original.GetAppPath(), "."+u.original.Name+"-upgrade-")
	if err != nil {
		return err
	}
	if err = fileOp.CopyDirWithNewName(u.detailDir, u.stageDir, "."); err != nil {
		return err
	}
	if err = copyUpgradeStageFile(u.original.GetPath(), u.stageDir, ".env"); err != nil {
		return err
	}
	if u.original.App.Key == constant.AppOpenclaw {
		if err = copyUpgradeStageFile(u.original.GetPath(), u.stageDir, path.Join("data", "conf", "openclaw.json")); err != nil {
			return err
		}
	}
	if u.original.App.Key == constant.AppOpenresty {
		for _, relativePath := range []string{
			nginxModuleBuildDir,
			nginxModuleModulesDir,
			path.Join(nginxModuleConfDir, nginxModuleEnabledConfDir),
		} {
			if err = copyUpgradeStageFile(u.original.GetPath(), u.stageDir, relativePath); err != nil {
				return err
			}
		}
	}

	stagedInstall := u.original
	stagedInstall.Name = path.Base(u.stageDir)
	stagedInstall.Version = u.detail.Version
	stagedInstall.AppDetailId = u.req.DetailID
	if stagedInstall.App.Key == vllmAppKeyForUpgrade {
		envs := make(map[string]interface{})
		if err = json.Unmarshal([]byte(stagedInstall.Env), &envs); err != nil {
			return err
		}
		image := buildVllmUpgradeImage(loadVllmImageFromEnv(stagedInstall.Env), u.original.Version, u.detail.Version)
		envs[vllmImageEnvKey] = image
		paramBytes, marshalErr := json.Marshal(envs)
		if marshalErr != nil {
			return marshalErr
		}
		stagedInstall.Env = string(paramBytes)
	}
	if err = migrateOpenclawProtocolUpgrade(&stagedInstall, u.original.Version, u.detail.Version); err != nil {
		return err
	}

	u.candidate = stagedInstall
	u.candidate.Name = u.original.Name
	u.candidate.DockerCompose, err = renderUpgradeCompose(u.candidate, u.detail, u.req.DockerCompose)
	if err != nil {
		return err
	}
	if strings.TrimSpace(u.candidate.DockerCompose) == "" {
		return buserr.WithName("ErrFileNotFound", "docker-compose.yml")
	}

	u.envContent, err = renderUpgradeEnv(u.candidate, u.oldEnvContent)
	if err != nil {
		return err
	}
	if err = writeUpgradeFile(path.Join(u.stageDir, ".env"), u.envContent, constant.FilePerm); err != nil {
		return err
	}
	if err = writeUpgradeFile(path.Join(u.stageDir, "docker-compose.yml"), []byte(u.candidate.DockerCompose), constant.FilePerm); err != nil {
		return err
	}
	project, err := docker.GetComposeProject(u.original.Name, u.stageDir, []byte(u.candidate.DockerCompose), u.envContent, false)
	if err != nil {
		return err
	}
	hasBuild := false
	for _, service := range project.Services {
		if service.Image == "" && service.Build == nil {
			return fmt.Errorf("compose service %s has neither image nor build configuration", service.Name)
		}
		hasBuild = hasBuild || service.Build != nil
	}
	if u.req.DeleteImage {
		dockerClient, clientErr := docker.NewClient()
		if clientErr != nil {
			return clientErr
		}
		u.oldImageIDs, err = getAppImageIDsByCompose(dockerClient, u.oldEnvContent, []byte(u.oldDockerCompose))
		dockerClient.Close()
		if err != nil {
			return err
		}
	}

	images := make([]string, 0, len(project.Services))
	for _, service := range project.Services {
		if service.Image != "" {
			images = append(images, service.Image)
		}
	}
	if err = prepareUpgradeImages(t, images, u.req.PullImage); err != nil {
		return err
	}
	if u.candidate.App.Key == constant.AppOpenresty {
		if err = u.prepareOpenresty(t, stagedInstall); err != nil {
			return err
		}
		if err = verifyUpgradeImages(images); err != nil {
			return err
		}
	} else if hasBuild {
		logStr := fmt.Sprintf("%s %s", i18n.GetMsgByKey("TaskBuild"), i18n.GetMsgByKey("Image"))
		t.LogStart(logStr)
		if err = compose.BuildWithTask(path.Join(u.stageDir, "docker-compose.yml"), project.Name, t); err != nil {
			t.LogFailedWithErr(logStr, err)
			return err
		}
		t.LogSuccess(logStr)
		if err = verifyUpgradeImages(images); err != nil {
			return err
		}
	}
	if u.original.App.Resource == constant.AppResourceRemote {
		go RequestDownloadCallBack(u.detail.DownloadCallBackUrl)
	}
	u.phase = appUpgradePrepared
	return nil
}

func (u *appUpgradeContext) prepareOpenresty(t *task.Task, stagedInstall model.AppInstall) error {
	fileOp := files.NewFileOp()
	detailBuildDir := path.Join(u.detailDir, nginxModuleBuildDir)
	installBuildDir := path.Join(u.stageDir, nginxModuleBuildDir)
	if !fileOp.Stat(installBuildDir) {
		if err := fileOp.CreateDir(installBuildDir, constant.DirPerm); err != nil {
			return err
		}
	}
	if err := copyAppDetailMissing(fileOp, detailBuildDir, installBuildDir); err != nil {
		return err
	}
	if err := fileOp.DeleteDir(path.Join(installBuildDir, nginxModuleTmpDir)); err != nil {
		return err
	}
	if err := fileOp.CopyDir(path.Join(detailBuildDir, nginxModuleTmpDir), installBuildDir); err != nil {
		return err
	}
	for _, fileName := range []string{"Dockerfile", "nginx.conf", "nginx.vh.default.conf"} {
		if err := fileOp.CopyFile(path.Join(detailBuildDir, fileName), installBuildDir); err != nil {
			return err
		}
	}
	if err := syncNginxModuleBuilder(detailBuildDir, installBuildDir); err != nil {
		return err
	}
	targetCatalogSource := path.Join(detailBuildDir, nginxModuleCatalogFile)
	if !fileOp.Stat(targetCatalogSource) {
		return fmt.Errorf("target OpenResty module catalog not found: %s", targetCatalogSource)
	}
	targetCatalogPath := path.Join(installBuildDir, nginxModuleCatalogPendingFile)
	if err := stageNginxModuleCatalog(targetCatalogSource, targetCatalogPath); err != nil {
		return err
	}
	stagedInstall.Name = path.Base(u.stageDir)
	stagedInstall.Version = u.candidate.Version
	stagedInstall.Env = u.candidate.Env
	stagedInstall.DockerCompose = u.candidate.DockerCompose
	return buildNginx(t, stagedInstall, targetCatalogPath)
}

func (u *appUpgradeContext) cutover(t *task.Task) error {
	u.stopAttempted = true
	t.LogStart(i18n.GetMsgByKey("UpgradeStop"))
	if out, err := compose.Stop(u.original.GetComposePath()); err != nil {
		if out != "" {
			err = fmt.Errorf("%s: %w", out, err)
		}
		t.LogFailedWithErr(i18n.GetMsgByKey("UpgradeStop"), err)
		return err
	}
	t.LogSuccess(i18n.GetMsgByKey("UpgradeStop"))
	u.phase = appUpgradeStopped

	var err error
	if u.original.App.Key == constant.AppOpenresty {
		u.snapshot, err = createOpenrestyUpgradeSnapshot(u.original.GetPath())
	} else {
		snapshotPaths := []string{".env", "docker-compose.yml", "scripts"}
		if u.original.App.Key == constant.AppOpenclaw {
			snapshotPaths = append(snapshotPaths, path.Join("data", "conf", "openclaw.json"))
		}
		u.snapshot, err = createUpgradeFileSnapshot(u.original.GetPath(), snapshotPaths)
	}
	if err != nil {
		return err
	}

	if u.req.Backup {
		if err = u.backup(t); err != nil {
			return err
		}
		u.phase = appUpgradeBackedUp
	} else {
		t.Log(i18n.GetMsgByKey("UpgradeBackupDisabled"))
	}

	u.downAttempted = true
	if out, downErr := compose.Down(u.original.GetComposePath()); downErr != nil {
		if out != "" {
			downErr = fmt.Errorf("%s: %w", out, downErr)
		}
		return downErr
	}
	u.phase = appUpgradeDown

	u.phase = appUpgradeMutated
	if err = u.applyStagedFiles(); err != nil {
		return err
	}
	if err = writeUpgradeFile(u.original.GetEnvPath(), u.envContent, constant.FilePerm); err != nil {
		return err
	}
	if err = runScript(t, &u.candidate, "upgrade"); err != nil {
		return err
	}
	if err = writeUpgradeFile(u.original.GetComposePath(), []byte(u.candidate.DockerCompose), constant.FilePerm); err != nil {
		return err
	}

	logStr := fmt.Sprintf("%s %s", i18n.GetMsgByKey("Run"), i18n.GetMsgByKey("App"))
	t.LogStart(logStr)
	if out, upErr := compose.UpWithoutPull(u.original.GetComposePath()); upErr != nil {
		if out != "" {
			upErr = fmt.Errorf("%s: %w", out, upErr)
		}
		t.LogFailedWithErr(logStr, upErr)
		return upErr
	}
	t.LogSuccess(logStr)
	u.phase = appUpgradeStarted

	t.LogStart(i18n.GetMsgByKey("UpgradeWaitReady"))
	containerNames, err := waitAppContainersReady(context.Background(), u.candidate)
	if err != nil {
		t.LogFailedWithErr(i18n.GetMsgByKey("UpgradeWaitReady"), err)
		return err
	}
	t.LogSuccess(i18n.GetMsgByKey("UpgradeWaitReady"))
	u.phase = appUpgradeReady
	u.candidate.ContainerName = strings.Join(containerNames, ",")
	u.candidate.Status = constant.StatusRunning
	u.candidate.Message = ""

	if u.candidate.App.Key == constant.AppOpenresty {
		liveCatalogPath := path.Join(u.candidate.GetPath(), nginxModuleBuildDir, nginxModuleCatalogPendingFile)
		if err = commitStaticNginxModuleBuilds(u.candidate, liveCatalogPath, t); err != nil {
			return err
		}
		activeCatalogPath := path.Join(u.candidate.GetPath(), nginxModuleBuildDir, nginxModuleCatalogFile)
		if err = activateNginxModuleCatalogAndCommit(liveCatalogPath, activeCatalogPath, func() error {
			return appInstallRepo.Save(context.Background(), &u.candidate)
		}); err != nil {
			return err
		}
	} else if err = appInstallRepo.Save(context.Background(), &u.candidate); err != nil {
		return err
	}
	u.phase = appUpgradeCommitted
	u.deleteOldImages(t)
	return nil
}

func (u *appUpgradeContext) backup(t *task.Task) error {
	fileName := fmt.Sprintf("upgrade_backup_%s_%s.tar.gz", u.original.Name, time.Now().Format(constant.DateTimeSlimLayout)+common.RandStrAndNum(5))
	record, err := backupAppWithParentTask(&u.original, t, fileName)
	if err != nil {
		return buserr.WithNameAndErr("ErrAppBackup", u.original.Name, err)
	}
	u.backupFile = path.Join(global.Dir.LocalBackupDir, record.FileDir, record.FileName)
	info, err := os.Stat(u.backupFile)
	if err != nil || info.Size() == 0 || record.Status != constant.StatusSuccess {
		if err == nil {
			err = errors.New("backup archive is empty or incomplete")
		}
		markBackupFailed(record.ID, err)
		return buserr.WithNameAndErr("ErrAppBackup", u.original.Name, err)
	}

	backupRecordService := NewIBackupRecordService()
	backups, _ := backupRecordService.ListAppRecords(u.original.App.Key, u.original.Name, "upgrade_backup")
	if len(backups) > 3 {
		deleteIDs := make([]uint, 0, len(backups)-3)
		for _, backup := range backups[:len(backups)-3] {
			deleteIDs = append(deleteIDs, backup.ID)
		}
		_ = backupRecordService.BatchDeleteRecord(deleteIDs)
	}
	return nil
}

func (u *appUpgradeContext) applyStagedFiles() error {
	fileOp := files.NewFileOp()
	if err := copyAppDetailMissingTracked(fileOp, u.detailDir, u.original.GetPath(), &u.createdPaths); err != nil {
		return err
	}
	if err := replaceUpgradePath(u.stageDir, u.original.GetPath(), "scripts"); err != nil {
		return err
	}
	if u.original.App.Key == constant.AppOpenclaw {
		if err := replaceUpgradePath(u.stageDir, u.original.GetPath(), path.Join("data", "conf", "openclaw.json")); err != nil {
			return err
		}
	}
	if u.original.App.Key == constant.AppOpenresty {
		for _, relativePath := range []string{
			nginxModuleBuildDir,
			nginxModuleModulesDir,
			path.Join(nginxModuleConfDir, nginxModuleEnabledConfDir),
			path.Join(nginxModuleConfDir, "nginx.conf"),
		} {
			if err := replaceUpgradePath(u.stageDir, u.original.GetPath(), relativePath); err != nil {
				return err
			}
		}
	}
	return nil
}

func (u *appUpgradeContext) rollback(t *task.Task) (rollbackErr error) {
	if !u.stopAttempted {
		return nil
	}
	logStr := i18n.GetWithName("AppRecover", u.original.Name)
	t.LogStart(logStr)
	defer func() {
		if rollbackErr != nil {
			t.LogFailedWithErr(logStr, rollbackErr)
		} else {
			t.LogSuccess(logStr)
		}
	}()
	if !u.downAttempted {
		if out, err := compose.Operate(u.original.GetComposePath(), "start"); err != nil {
			if out != "" {
				err = fmt.Errorf("%s: %w", out, err)
			}
			return err
		}
		return u.finishRollback()
	}
	if u.phase < appUpgradeMutated {
		if out, err := compose.UpWithoutPull(u.original.GetComposePath()); err != nil {
			if out != "" {
				err = fmt.Errorf("%s: %w", out, err)
			}
			return err
		}
		return u.finishRollback()
	}

	if out, err := compose.Down(u.original.GetComposePath()); err != nil {
		if out != "" {
			err = fmt.Errorf("%s: %w", out, err)
		}
		rollbackErr = err
	}
	if u.backupFile != "" {
		_ = u.restoreManagedFiles()
		if err := handleAppRecover(&u.original, t, u.backupFile, true, "", ""); err != nil {
			_, _ = compose.UpWithoutPull(u.original.GetComposePath())
			return errors.Join(rollbackErr, err)
		}
	} else {
		if err := u.restoreManagedFiles(); err != nil {
			return errors.Join(rollbackErr, err)
		}
		if out, err := compose.UpWithoutPull(u.original.GetComposePath()); err != nil {
			if out != "" {
				err = fmt.Errorf("%s: %w", out, err)
			}
			return errors.Join(rollbackErr, err)
		}
	}
	return errors.Join(rollbackErr, u.finishRollback())
}

func (u *appUpgradeContext) finishRollback() error {
	if _, err := waitAppContainersReady(context.Background(), u.original); err != nil {
		return err
	}
	restored := u.original
	if err := appInstallRepo.Save(context.Background(), &restored); err != nil {
		return err
	}
	return nil
}

func (u *appUpgradeContext) restoreManagedFiles() error {
	var restoreErr error
	if u.snapshot != nil {
		restoreErr = u.snapshot.Restore()
	}
	for index := len(u.createdPaths) - 1; index >= 0; index-- {
		if err := os.RemoveAll(u.createdPaths[index]); err != nil {
			restoreErr = errors.Join(restoreErr, err)
		}
	}
	return restoreErr
}

func (u *appUpgradeContext) deleteOldImages(t *task.Task) {
	if !u.req.DeleteImage {
		return
	}
	excludeImages, err := docker.GetImagesFromDockerCompose(u.envContent, []byte(u.candidate.DockerCompose))
	if err != nil {
		t.LogFailedWithErr(i18n.GetMsgByKey("TaskDelete")+i18n.GetMsgByKey("Image"), err)
		return
	}
	dockerClient, err := docker.NewClient()
	if err != nil {
		t.LogFailedWithErr(i18n.GetMsgByKey("TaskDelete")+i18n.GetMsgByKey("Image"), err)
		return
	}
	defer dockerClient.Close()
	if err = deleteAppImagesByIDs(t, dockerClient, u.oldImageIDs, excludeImages); err != nil {
		t.LogFailedWithErr(i18n.GetMsgByKey("TaskDelete")+i18n.GetMsgByKey("Image"), err)
	}
}

func (u *appUpgradeContext) cleanup() {
	if u.snapshot != nil {
		u.snapshot.Cleanup()
	}
	if u.stageDir != "" {
		_ = os.RemoveAll(u.stageDir)
	}
}

type upgradeImageClient interface {
	PullImageWithProcess(*task.Task, string) error
	ImageExists(string) (bool, error)
	Close()
}

func prepareUpgradeImages(t *task.Task, images []string, pull bool) error {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return err
	}
	return prepareUpgradeImagesWithClient(t, dockerClient, images, pull)
}

func prepareUpgradeImagesWithClient(t *task.Task, dockerClient upgradeImageClient, images []string, pull bool) error {
	defer dockerClient.Close()
	seen := make(map[string]struct{}, len(images))
	for _, image := range images {
		image = strings.TrimSpace(image)
		if image == "" {
			continue
		}
		if _, ok := seen[image]; ok {
			continue
		}
		seen[image] = struct{}{}
		if pull {
			if t != nil {
				t.Log(i18n.GetWithName("PullImageStart", image))
			}
			if pullErr := dockerClient.PullImageWithProcess(t, image); pullErr != nil {
				if exists, _ := dockerClient.ImageExists(image); exists {
					if t != nil {
						t.Log(i18n.GetMsgByKey("UseExistImage"))
					}
					continue
				}
				return buserr.WithNameAndErr("ErrDockerPullImage", "", pullErr)
			}
		}
		exists, inspectErr := dockerClient.ImageExists(image)
		if inspectErr != nil || !exists {
			return buserr.WithNameAndErr("ErrDockerPullImage", "", fmt.Errorf("image %s is not available locally: %v", image, inspectErr))
		}
		if pull && t != nil {
			t.LogSuccess(i18n.GetMsgByKey("PullImage"))
		}
	}
	return nil
}

func verifyUpgradeImages(images []string) error {
	dockerClient, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer dockerClient.Close()
	for _, image := range images {
		exists, inspectErr := dockerClient.ImageExists(image)
		if inspectErr != nil || !exists {
			return buserr.WithNameAndErr("ErrDockerPullImage", "", fmt.Errorf("image %s is not available locally: %v", image, inspectErr))
		}
	}
	return nil
}

func renderUpgradeEnv(install model.AppInstall, original []byte) ([]byte, error) {
	envs := make(map[string]interface{})
	if err := json.Unmarshal([]byte(install.Env), &envs); err != nil {
		return nil, err
	}
	params := make(map[string]string, len(envs))
	handleMap(envs, params)
	if install.App.Key == constant.AppOpenresty {
		originalEnv, _ := gotenv.Unmarshal(string(original))
		for _, key := range []string{"CONTAINER_PACKAGE_URL", "RESTY_ADD_PACKAGE_BUILDDEPS", "RESTY_CONFIG_OPTIONS_MORE"} {
			params[key] = originalEnv[key]
		}
	}
	content, err := gotenv.Marshal(params)
	if err != nil {
		return nil, err
	}
	return []byte(content), nil
}

func renderUpgradeCompose(install model.AppInstall, detail model.AppDetail, customCompose string) (string, error) {
	if customCompose != "" {
		return customCompose, nil
	}
	if install.App.Key == vllmAppKeyForUpgrade {
		return install.DockerCompose, nil
	}
	return getUpgradeCompose(install, detail)
}

func writeUpgradeFile(filePath string, content []byte, mode os.FileMode) error {
	tmp, err := os.CreateTemp(path.Dir(filePath), "."+path.Base(filePath)+".*")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if err = tmp.Chmod(mode); err == nil {
		_, err = tmp.Write(content)
	}
	if err == nil {
		err = tmp.Sync()
	}
	if closeErr := tmp.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, filePath)
}

func copyUpgradeStageFile(sourceRoot, targetRoot, relativePath string) error {
	source := path.Join(sourceRoot, relativePath)
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	target := path.Join(targetRoot, relativePath)
	_ = os.RemoveAll(target)
	return copyOpenrestyUpgradeSnapshotEntry(source, target)
}

func replaceUpgradePath(sourceRoot, targetRoot, relativePath string) error {
	source := path.Join(sourceRoot, relativePath)
	if _, err := os.Stat(source); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	target := path.Join(targetRoot, relativePath)
	if err := os.RemoveAll(target); err != nil {
		return err
	}
	return copyOpenrestyUpgradeSnapshotEntry(source, target)
}

func createUpgradeFileSnapshot(installPath string, paths []string) (*upgradeFileSnapshot, error) {
	backupPath, err := os.MkdirTemp("", "1panel-app-upgrade-*")
	if err != nil {
		return nil, err
	}
	snapshot := &upgradeFileSnapshot{
		installPath: installPath,
		backupPath:  backupPath,
		paths:       paths,
		existing:    make(map[string]bool, len(paths)),
	}
	for _, relativePath := range paths {
		source := path.Join(installPath, relativePath)
		if _, err = os.Stat(source); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			snapshot.Cleanup()
			return nil, err
		}
		snapshot.existing[relativePath] = true
		if err = copyOpenrestyUpgradeSnapshotEntry(source, path.Join(backupPath, relativePath)); err != nil {
			snapshot.Cleanup()
			return nil, err
		}
	}
	return snapshot, nil
}

func (s *upgradeFileSnapshot) Restore() error {
	for _, relativePath := range s.paths {
		target := path.Join(s.installPath, relativePath)
		if err := os.RemoveAll(target); err != nil {
			return err
		}
		if !s.existing[relativePath] {
			continue
		}
		if err := copyOpenrestyUpgradeSnapshotEntry(path.Join(s.backupPath, relativePath), target); err != nil {
			return err
		}
	}
	return nil
}

func (s *upgradeFileSnapshot) Cleanup() {
	if s != nil && s.backupPath != "" {
		_ = os.RemoveAll(s.backupPath)
	}
}

type appContainerReadinessClient interface {
	ContainerList(context.Context, container.ListOptions) ([]container.Summary, error)
	ContainerInspect(context.Context, string) (container.InspectResponse, error)
}

func waitAppContainersReady(ctx context.Context, install model.AppInstall) ([]string, error) {
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	return waitAppContainersReadyWithClient(ctx, client, install)
}

func waitAppContainersReadyWithClient(ctx context.Context, client appContainerReadinessClient, install model.AppInstall) ([]string, error) {
	envContent, err := os.ReadFile(install.GetEnvPath())
	if err != nil {
		envContent, err = renderUpgradeEnv(install, nil)
		if err != nil {
			return nil, err
		}
	}
	project, err := docker.GetComposeProject(install.Name, install.GetPath(), []byte(install.DockerCompose), envContent, false)
	if err != nil {
		return nil, err
	}
	expectedServices := make(map[string]struct{})
	for _, service := range project.Services {
		if !skipCheckStatus(service) {
			expectedServices[service.Name] = struct{}{}
		}
	}
	if len(expectedServices) == 0 {
		return strings.Split(install.ContainerName, ","), nil
	}
	options := container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", composeWorkdirLabel+"="+install.GetPath()),
		),
	}
	containers, err := client.ContainerList(ctx, options)
	if err != nil {
		return nil, err
	}
	foundServices := make(map[string]bool, len(expectedServices))
	containerNames := make([]string, 0, len(containers))
	for _, item := range containers {
		serviceName := item.Labels[composeServiceLabel]
		if _, ok := expectedServices[serviceName]; !ok {
			continue
		}
		if err = waitContainerReady(ctx, client, item.ID); err != nil {
			return nil, fmt.Errorf("container %s is not ready: %w", serviceName, err)
		}
		foundServices[serviceName] = true
		if len(item.Names) > 0 {
			containerNames = append(containerNames, strings.TrimPrefix(item.Names[0], "/"))
		}
	}
	for serviceName := range expectedServices {
		if !foundServices[serviceName] {
			return nil, fmt.Errorf("container for service %s was not created", serviceName)
		}
	}
	sort.Strings(containerNames)
	return containerNames, nil
}
