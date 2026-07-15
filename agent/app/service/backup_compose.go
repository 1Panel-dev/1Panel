package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
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
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	dockerUtils "github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
)

type composeBackupFile struct {
	OriginalPath string `json:"originalPath"`
	FileName     string `json:"fileName"`
	RelativePath string `json:"relativePath,omitempty"`
	BackupPath   string `json:"backupPath"`
}

type composeBackupMeta struct {
	ComposeName string              `json:"composeName"`
	ComposePath string              `json:"composePath"`
	CreatedAt   string              `json:"createdAt"`
	Files       []composeBackupFile `json:"files"`
	Containers  []string            `json:"containers"`
}

type composeBackupContext struct {
	req          dto.CommonBackup
	composeName  string
	composePath  string
	composeFiles []string
	composeDir   string
	fileOp       files.FileOp
	dockerClient *client.Client
	stopped      bool
	backupDir    string
	fileName     string
	filePath     string
	tmpDir       string
	meta         composeBackupMeta
}

type composeRecoverContext struct {
	req         dto.CommonRecover
	fileOp      files.FileOp
	tmpDir      string
	meta        composeBackupMeta
	composeName string
	targetDir   string
	composePath string
	enqueued    bool
}

func (u *BackupService) ComposeBackup(req dto.CommonBackup) error {
	timeNow := time.Now().Format(constant.DateTimeSlimLayout) + common.RandStrAndNum(5)
	fileName := req.FileName
	if fileName == "" {
		fileName = fmt.Sprintf("%s_%s.tar.gz", req.Name, timeNow)
	}
	if !strings.HasSuffix(fileName, ".tar.gz") {
		fileName += ".tar.gz"
	}
	itemDir := fmt.Sprintf("compose/%s", req.Name)
	backupDir := path.Join(global.Dir.LocalBackupDir, itemDir)
	record := &model.BackupRecord{
		Type:              req.Type,
		Name:              req.Name,
		SourceAccountIDs:  "1",
		DownloadAccountID: 1,
		FileDir:           itemDir,
		FileName:          fileName,
		TaskID:            req.TaskID,
		Status:            constant.StatusWaiting,
		Description:       req.Description,
	}
	if err := backupRepo.CreateRecord(record); err != nil {
		global.LOG.Errorf("save compose backup record failed, err: %v", err)
		return err
	}
	if err := handleComposeBackup(req, nil, record.ID, backupDir, fileName); err != nil {
		markBackupFailed(record.ID, err)
		return err
	}
	return nil
}

func (u *BackupService) ComposeRecover(req dto.CommonRecover) error {
	return handleComposeRecover(req, nil)
}

func handleComposeBackup(req dto.CommonBackup, parentTask *task.Task, recordID uint, backupDir, fileName string) error {
	composeCtx, err := newComposeBackupContext(req, backupDir, fileName)
	if err != nil {
		return err
	}
	containerNames, err := loadComposeContainerNames(composeCtx)
	if err != nil {
		composeCtx.close()
		return err
	}

	backupTask := parentTask
	if backupTask == nil {
		backupTask, err = task.NewTaskWithOps(composeCtx.composeName, task.TaskBackup, task.TaskScopeBackup, req.TaskID, 1)
		if err != nil {
			return err
		}
	}
	if req.StopBefore {
		backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupStop"), func(t *task.Task) error {
			return stepStopComposeForBackup(composeCtx)
		}, func(t *task.Task) {
			_ = stepStartComposeAfterBackup(composeCtx)
		}, 3, time.Hour)
	}
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupPrepare"), func(t *task.Task) error { return stepPrepareComposeBackup(composeCtx) }, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupFiles"), func(t *task.Task) error { return stepBackupComposeFiles(composeCtx) }, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupContainers"), func(t *task.Task) error { return nil }, nil, 3, time.Hour)
	for _, containerName := range containerNames {
		backupFileName := fmt.Sprintf("%s.tar.gz", sanitizeComposeFileName(containerName))
		backupFile := path.Join(composeCtx.tmpDir, "containers", backupFileName)
		if err := handleContainerBackup(containerName, backupTask, 0, path.Dir(backupFile), path.Base(backupFile), "", "", false); err != nil {
			return err
		}
		composeCtx.meta.Containers = append(composeCtx.meta.Containers, path.Join("containers", backupFileName))
	}
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupMeta"), func(t *task.Task) error { return stepWriteComposeBackupMeta(composeCtx) }, nil, 3, time.Hour)
	backupTask.AddSubTaskWithOps(task.GetTaskName(composeCtx.composeName, task.TaskBackup, task.TaskScopeBackup), func(t *task.Task) error {
		return stepPackComposeBackup(composeCtx)
	}, nil, 3, time.Hour)
	if req.StopBefore {
		backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupStart"), func(t *task.Task) error {
			return stepStartComposeAfterBackup(composeCtx)
		}, nil, 3, time.Hour)
	}
	backupTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeBackupCleanup"), func(t *task.Task) error {
		composeCtx.close()
		return nil
	}, nil, 0, time.Hour)
	if parentTask != nil {
		return nil
	}
	go func() {
		defer composeCtx.close()
		if err := backupTask.Execute(); err != nil {
			markBackupFailed(recordID, err)
			return
		}
		backupRepo.UpdateRecordByMap(recordID, map[string]interface{}{"status": constant.StatusSuccess})
	}()
	return nil
}

func loadComposeContainerNames(composeCtx *composeBackupContext) ([]string, error) {
	options := container.ListOptions{All: true}
	options.Filters = filters.NewArgs(filters.Arg("label", composeProjectLabel+"="+composeCtx.composeName))
	containers, err := composeCtx.dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(containers))
	for _, item := range containers {
		if len(item.Names) == 0 {
			continue
		}
		names = append(names, strings.TrimPrefix(item.Names[0], "/"))
	}
	sort.Strings(names)
	return names, nil
}

func handleComposeRecover(req dto.CommonRecover, parentTask *task.Task) error {
	var recoverCtx *composeRecoverContext

	recoverTask := parentTask
	var err error
	if recoverTask == nil {
		if isImportRecover(req) {
			taskName := i18n.GetMsgByKey("TaskImport") + i18n.GetMsgByKey("Compose")
			recoverTask, err = task.NewTask(taskName, task.TaskImport, task.TaskScopeBackup, req.TaskID, 1)
			if err != nil {
				return err
			}
		} else {
			taskName := req.Name
			if taskName == "" {
				taskName = "compose"
			}
			recoverTask, err = task.NewTaskWithOps(taskName, task.TaskRecover, task.TaskScopeBackup, req.TaskID, 1)
			if err != nil {
				return err
			}
		}
	}

	timeout := loadRecoverTimeout(req.Timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverPrepare"), func(t *task.Task) error {
		ctx, err := newComposeRecoverContext(req)
		if err != nil {
			return err
		}
		recoverCtx = ctx
		if err := stepPrepareComposeRecover(recoverCtx); err != nil {
			recoverCtx.close()
			recoverCtx = nil
			return err
		}
		return nil
	}, func(t *task.Task) {
		if recoverCtx != nil {
			recoverCtx.close()
			recoverCtx = nil
		}
	}, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverExtract"), func(t *task.Task) error { return stepExtractComposeRecover(recoverCtx) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverMeta"), func(t *task.Task) error {
		if err := stepLoadComposeRecoverMeta(recoverCtx); err != nil {
			return err
		}
		t.Log(i18n.GetMsgWithMap("ComposeRecoverMetaLogName", map[string]interface{}{
			"name": recoverCtx.composeName,
		}))
		t.Log(i18n.GetMsgWithMap("ComposeRecoverMetaLogPath", map[string]interface{}{
			"backupPath": recoverCtx.meta.ComposePath,
			"targetDir":  recoverCtx.targetDir,
		}))
		t.Log(i18n.GetMsgWithMap("ComposeRecoverMetaLogCount", map[string]interface{}{
			"files":      len(recoverCtx.meta.Files),
			"containers": len(recoverCtx.meta.Containers),
		}))
		return nil
	}, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverFiles"), func(t *task.Task) error { return stepRestoreComposeFiles(recoverCtx) }, nil, 3, timeout)
	recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverContainers"), func(t *task.Task) error {
		if recoverCtx.enqueued {
			return nil
		}
		containerItems := make([]string, 0, len(recoverCtx.meta.Containers))
		for _, item := range recoverCtx.meta.Containers {
			backupItem := item
			filePath, err := safeJoinWithinBase(recoverCtx.tmpDir, backupItem)
			if err != nil {
				return fmt.Errorf("invalid container backup path %q, err: %v", backupItem, err)
			}
			if !recoverCtx.fileOp.Stat(filePath) {
				return fmt.Errorf("container backup file not found: %s", backupItem)
			}
			containerItems = append(containerItems, backupItem)
		}
		for _, backupItem := range containerItems {
			filePath, err := safeJoinWithinBase(recoverCtx.tmpDir, backupItem)
			if err != nil {
				return fmt.Errorf("invalid container backup path %q, err: %v", backupItem, err)
			}
			containerLabel := strings.TrimSuffix(path.Base(backupItem), ".tar.gz")
			containerReq := recoverCtx.req
			containerReq.Type = "container"
			containerReq.Name = containerLabel
			containerReq.DetailName = ""
			containerReq.File = filePath
			containerReq.Secret = ""
			if err := handleContainerRecover(containerReq, recoverTask); err != nil {
				return err
			}
		}
		recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverRecord"), func(t *task.Task) error {
			return stepSaveComposeRecord(recoverCtx)
		}, nil, 3, timeout)
		recoverTask.AddSubTaskWithOps(i18n.GetMsgByKey("ComposeRecoverCleanup"), func(t *task.Task) error {
			if recoverCtx != nil {
				recoverCtx.close()
				recoverCtx = nil
			}
			return nil
		}, nil, 0, timeout)
		recoverCtx.enqueued = true
		return nil
	}, nil, 3, timeout)
	if parentTask != nil {
		return nil
	}
	go func() {
		_ = recoverTask.Execute()
	}()
	return nil
}

func newComposeBackupContext(req dto.CommonBackup, backupDir, fileName string) (*composeBackupContext, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("compose name is required")
	}
	dockerClient, err := dockerUtils.NewDockerClient()
	if err != nil {
		return nil, err
	}
	composePath, composeFiles, err := loadComposePathAndFiles(req.Name, dockerClient)
	if err != nil {
		_ = dockerClient.Close()
		return nil, err
	}
	filePath := path.Join(backupDir, fileName)
	tmpDir := path.Join(path.Dir(filePath), strings.TrimSuffix(path.Base(filePath), ".tar.gz"))
	ctx := &composeBackupContext{
		req:          req,
		composeName:  req.Name,
		composePath:  composePath,
		composeFiles: composeFiles,
		composeDir:   path.Dir(composeFiles[0]),
		fileOp:       files.NewFileOp(),
		dockerClient: dockerClient,
		backupDir:    backupDir,
		fileName:     fileName,
		filePath:     filePath,
		tmpDir:       tmpDir,
		meta: composeBackupMeta{
			ComposeName: req.Name,
			ComposePath: composePath,
			CreatedAt:   time.Now().Format(constant.DateTimeLayout),
			Files:       make([]composeBackupFile, 0),
			Containers:  make([]string, 0),
		},
	}
	return ctx, nil
}

func loadComposePathAndFiles(composeName string, dockerClient *client.Client) (string, []string, error) {
	composeRecord, _ := composeRepo.GetRecord(repo.WithByName(composeName))
	if composeRecord.ID == 0 {
		composeRecord, _ = composeRepo.GetRecord(repo.WithByName(strings.ToLower(composeName)))
	}
	composePath := composeRecord.Path
	if composePath == "" {
		options := container.ListOptions{All: true}
		options.Filters = filters.NewArgs(filters.Arg("label", composeProjectLabel))
		list, err := dockerClient.ContainerList(context.Background(), options)
		if err != nil {
			return "", nil, err
		}
		if len(list) == 0 {
			return "", nil, fmt.Errorf("compose %s not found", composeName)
		}
		var targetContainer *container.Summary
		for i := range list {
			if strings.EqualFold(list[i].Labels[composeProjectLabel], composeName) {
				targetContainer = &list[i]
				break
			}
		}
		if targetContainer == nil {
			return "", nil, fmt.Errorf("compose %s not found", composeName)
		}
		config := targetContainer.Labels[composeConfigLabel]
		workdir := targetContainer.Labels[composeWorkdirLabel]
		if len(config) != 0 && len(workdir) != 0 && strings.Contains(config, workdir) {
			composePath = config
		} else {
			composePath = workdir
		}
	}
	composeFiles := normalizeComposeFiles(composePath)
	if len(composeFiles) == 0 {
		return "", nil, fmt.Errorf("compose file not found for %s", composeName)
	}
	return composePath, composeFiles, nil
}

func normalizeComposeFiles(composePath string) []string {
	items := strings.Split(composePath, ",")
	result := make([]string, 0)
	seen := make(map[string]struct{})
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		stat, err := os.Stat(item)
		if err == nil && stat.IsDir() {
			item = path.Join(item, "docker-compose.yml")
		}
		if _, err := os.Stat(item); err != nil {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func (c *composeBackupContext) close() {
	if c.dockerClient != nil {
		_ = c.dockerClient.Close()
		c.dockerClient = nil
	}
	if c.tmpDir != "" {
		_ = os.RemoveAll(c.tmpDir)
		c.tmpDir = ""
	}
}

func stepPrepareComposeBackup(composeCtx *composeBackupContext) error {
	if err := os.MkdirAll(composeCtx.backupDir, os.ModePerm); err != nil {
		return fmt.Errorf("mkdir %s failed, err: %v", composeCtx.backupDir, err)
	}
	_ = os.RemoveAll(composeCtx.tmpDir)
	if err := os.MkdirAll(path.Join(composeCtx.tmpDir, "compose_files"), os.ModePerm); err != nil {
		return err
	}
	if err := os.MkdirAll(path.Join(composeCtx.tmpDir, "containers"), os.ModePerm); err != nil {
		return err
	}
	return nil
}

func stepStopComposeForBackup(composeCtx *composeBackupContext) error {
	if composeCtx.stopped {
		return nil
	}
	options := container.ListOptions{All: false}
	options.Filters = filters.NewArgs(filters.Arg("label", composeProjectLabel+"="+composeCtx.composeName))
	runningList, err := composeCtx.dockerClient.ContainerList(context.Background(), options)
	if err != nil {
		return err
	}
	if len(runningList) == 0 {
		return nil
	}
	if stdout, err := compose.Operate(composeCtx.composePath, "stop"); err != nil {
		return fmt.Errorf("docker-compose stop failed, std: %s, err: %v", stdout, err)
	}
	composeCtx.stopped = true
	return nil
}

func stepStartComposeAfterBackup(composeCtx *composeBackupContext) error {
	if !composeCtx.stopped {
		return nil
	}
	if stdout, err := compose.Up(composeCtx.composePath); err != nil {
		return fmt.Errorf("docker-compose up failed, std: %s, err: %v", stdout, err)
	}
	composeCtx.stopped = false
	return nil
}

func stepBackupComposeFiles(composeCtx *composeBackupContext) error {
	for i, filePath := range composeCtx.composeFiles {
		backupName := fmt.Sprintf("%02d_%s", i, path.Base(filePath))
		backupPath := path.Join(composeCtx.tmpDir, "compose_files", backupName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		if err := composeCtx.fileOp.SaveFile(backupPath, string(content), fs.ModePerm); err != nil {
			return err
		}
		relativePath := path.Base(filePath)
		if composeCtx.composeDir != "" {
			rel, relErr := filepath.Rel(composeCtx.composeDir, filePath)
			if relErr == nil {
				rel = filepath.ToSlash(rel)
				if rel != "" && rel != "." && !strings.HasPrefix(rel, "../") {
					relativePath = rel
				}
			}
		}
		composeCtx.meta.Files = append(composeCtx.meta.Files, composeBackupFile{
			OriginalPath: filePath,
			FileName:     path.Base(filePath),
			RelativePath: relativePath,
			BackupPath:   path.Join("compose_files", backupName),
		})
	}
	if len(composeCtx.composeFiles) != 0 {
		envPath := path.Join(path.Dir(composeCtx.composeFiles[0]), ".env")
		if composeCtx.fileOp.Stat(envPath) {
			envContent, err := os.ReadFile(envPath)
			if err != nil {
				return err
			}
			if err := composeCtx.fileOp.SaveFile(path.Join(composeCtx.tmpDir, "compose_files", ".env"), string(envContent), fs.ModePerm); err != nil {
				return err
			}
		}
	}
	return nil
}

func stepWriteComposeBackupMeta(composeCtx *composeBackupContext) error {
	metaBytes, err := json.MarshalIndent(composeCtx.meta, "", "  ")
	if err != nil {
		return err
	}
	return composeCtx.fileOp.SaveFile(path.Join(composeCtx.tmpDir, "compose_meta.json"), string(metaBytes), fs.ModePerm)
}

func stepPackComposeBackup(composeCtx *composeBackupContext) error {
	return composeCtx.fileOp.TarGzCompressPro(true, composeCtx.tmpDir, composeCtx.filePath, composeCtx.req.Secret, "")
}

func newComposeRecoverContext(req dto.CommonRecover) (*composeRecoverContext, error) {
	tmpDir := path.Join(path.Dir(req.File), strings.TrimSuffix(path.Base(req.File), ".tar.gz"))
	ctx := &composeRecoverContext{
		req:    req,
		fileOp: files.NewFileOp(),
		tmpDir: tmpDir,
		meta: composeBackupMeta{
			Files:      make([]composeBackupFile, 0),
			Containers: make([]string, 0),
		},
	}
	return ctx, nil
}

func (c *composeRecoverContext) close() {
	if c.tmpDir != "" {
		_ = os.RemoveAll(c.tmpDir)
	}
}

func stepPrepareComposeRecover(recoverCtx *composeRecoverContext) error {
	if !recoverCtx.fileOp.Stat(recoverCtx.req.File) {
		return buserr.WithName("ErrFileNotFound", recoverCtx.req.File)
	}
	_ = os.RemoveAll(recoverCtx.tmpDir)
	return nil
}

func stepExtractComposeRecover(recoverCtx *composeRecoverContext) error {
	return recoverCtx.fileOp.TarGzExtractPro(recoverCtx.req.File, path.Dir(recoverCtx.req.File), recoverCtx.req.Secret)
}

func stepLoadComposeRecoverMeta(recoverCtx *composeRecoverContext) error {
	metaPath := path.Join(recoverCtx.tmpDir, "compose_meta.json")
	if !recoverCtx.fileOp.Stat(metaPath) {
		return fmt.Errorf("compose_meta.json not found in backup file")
	}
	metaBytes, err := os.ReadFile(metaPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(metaBytes, &recoverCtx.meta); err != nil {
		return fmt.Errorf("unmarshal compose_meta.json failed, err: %v", err)
	}
	recoverCtx.composeName = strings.TrimSpace(recoverCtx.req.Name)
	if recoverCtx.composeName == "" {
		recoverCtx.composeName = strings.TrimSpace(recoverCtx.meta.ComposeName)
	}
	if recoverCtx.composeName == "" {
		return fmt.Errorf("compose name not found in recover request or backup file")
	}
	recoverCtx.targetDir = resolveComposeRecoverTargetDir(recoverCtx.meta, recoverCtx.composeName)
	return nil
}

func resolveComposeRecoverTargetDir(meta composeBackupMeta, composeName string) string {
	composePath := strings.TrimSpace(meta.ComposePath)
	if composePath != "" {
		items := strings.Split(composePath, ",")
		for _, item := range items {
			p := strings.TrimSpace(item)
			if p == "" {
				continue
			}
			ext := strings.ToLower(path.Ext(p))
			if ext == ".yml" || ext == ".yaml" {
				return path.Dir(p)
			}
			return p
		}
	}
	return path.Join(global.Dir.DataDir, "docker/compose", composeName)
}

func safeJoinWithinBase(baseDir, name string) (string, error) {
	base := filepath.Clean(baseDir)
	candidate := strings.TrimSpace(name)
	candidate = strings.ReplaceAll(candidate, "\\", "/")
	candidate = filepath.Clean(filepath.FromSlash(candidate))
	if candidate == "" || candidate == "." {
		return "", fmt.Errorf("invalid path: empty")
	}
	if filepath.IsAbs(candidate) {
		return "", fmt.Errorf("invalid path %q: absolute path is not allowed", name)
	}
	if candidate == ".." || strings.HasPrefix(candidate, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path %q: path escapes base directory", name)
	}
	resolved := filepath.Clean(filepath.Join(base, candidate))
	rel, err := filepath.Rel(base, resolved)
	if err != nil {
		return "", fmt.Errorf("resolve path %q failed, err: %v", name, err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid path %q: path escapes base directory", name)
	}
	return resolved, nil
}

func stepRestoreComposeFiles(recoverCtx *composeRecoverContext) error {
	if recoverCtx.targetDir != "" {
		_ = os.RemoveAll(recoverCtx.targetDir)
	}
	if err := os.MkdirAll(recoverCtx.targetDir, os.ModePerm); err != nil {
		return err
	}
	restored := make([]string, 0, len(recoverCtx.meta.Files))
	for _, item := range recoverCtx.meta.Files {
		backupPath, err := safeJoinWithinBase(recoverCtx.tmpDir, item.BackupPath)
		if err != nil {
			return fmt.Errorf("invalid compose backup path %q, err: %v", item.BackupPath, err)
		}
		if !recoverCtx.fileOp.Stat(backupPath) {
			continue
		}
		targetName := item.FileName
		if item.RelativePath != "" {
			targetName = item.RelativePath
		}
		if targetName == "" {
			targetName = path.Base(item.OriginalPath)
		}
		if targetName == "" {
			targetName = "docker-compose.yml"
		}
		targetPath, err := safeJoinWithinBase(recoverCtx.targetDir, targetName)
		if err != nil {
			return fmt.Errorf("invalid compose target path %q, err: %v", targetName, err)
		}
		if err := os.MkdirAll(path.Dir(targetPath), os.ModePerm); err != nil {
			return err
		}
		content, err := os.ReadFile(backupPath)
		if err != nil {
			return err
		}
		if err := recoverCtx.fileOp.SaveFile(targetPath, string(content), fs.ModePerm); err != nil {
			return err
		}
		restored = append(restored, targetPath)
	}
	envPath := path.Join(recoverCtx.tmpDir, "compose_files", ".env")
	if recoverCtx.fileOp.Stat(envPath) {
		envContent, err := os.ReadFile(envPath)
		if err != nil {
			return err
		}
		if err := recoverCtx.fileOp.SaveFile(path.Join(recoverCtx.targetDir, ".env"), string(envContent), fs.ModePerm); err != nil {
			return err
		}
	}
	if len(restored) == 0 {
		defaultPath := path.Join(recoverCtx.targetDir, "docker-compose.yml")
		if !recoverCtx.fileOp.Stat(defaultPath) {
			return fmt.Errorf("compose file not found in backup data")
		}
		restored = append(restored, defaultPath)
	}
	sort.Strings(restored)
	recoverCtx.composePath = strings.Join(restored, ",")
	return nil
}

func stepSaveComposeRecord(recoverCtx *composeRecoverContext) error {
	if recoverCtx.composePath == "" {
		recoverCtx.composePath = path.Join(recoverCtx.targetDir, "docker-compose.yml")
	}
	recordName := strings.ToLower(recoverCtx.composeName)
	record, _ := composeRepo.GetRecord(repo.WithByName(recordName))
	if record.ID == 0 {
		return composeRepo.CreateRecord(&model.Compose{Name: recordName, Path: recoverCtx.composePath})
	}
	return composeRepo.UpdateRecord(recordName, map[string]interface{}{"path": recoverCtx.composePath})
}

func sanitizeComposeFileName(in string) string {
	name := strings.TrimSpace(in)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ":", "_")
	if name == "" {
		return "container"
	}
	return name
}
