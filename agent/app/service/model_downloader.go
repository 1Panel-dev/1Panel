package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"gorm.io/gorm"
)

const (
	defaultModelDownloaderDir        = "/opt/1panel/ai/models"
	defaultHFEndpoint                = "https://huggingface.co"
	defaultModelScopeEndpoint        = "https://www.modelscope.cn"
	modelDownloaderSourceHuggingFace = "huggingface"
	modelDownloaderSourceModelScope  = "modelscope"
	statusDownloading                = "Downloading"
	modelDownloaderDirKey            = "ModelDownloaderDir"
	modelDownloaderEndpointKey       = "ModelDownloaderHFEndpoint"
	modelDownloaderTokenKey          = "ModelDownloaderHFToken"
	modelDownloaderMSEndpointKey     = "ModelDownloaderModelScopeEndpoint"
	modelDownloaderMSTokenKey        = "ModelDownloaderModelScopeToken"
)

type IModelDownloaderService interface {
	PageFeaturedModels(req dto.FeaturedModelSearch) (*dto.FeaturedModelPage, error)
	SyncFeaturedModels() error
	SearchHFModels(req dto.ModelSearch) (int64, []dto.ModelItem, error)
	GetHFModelInfo(req dto.ModelInfoReq) (*dto.ModelInfo, error)
	StartDownload(req dto.ModelDownload) (*dto.ModelDownloadTaskItem, error)
	SearchModelScopeModels(req dto.ModelSearch) (int64, []dto.ModelItem, error)
	GetModelScopeModelInfo(req dto.ModelInfoReq) (*dto.ModelInfo, error)
	StartModelScopeDownload(req dto.ModelDownload) (*dto.ModelDownloadTaskItem, error)
	GetSettings() (*dto.ModelDownloaderSetting, error)
	UpdateSettings(req dto.ModelDownloaderSetting) error
	UpdateFeaturedModelsSync(req dto.ModelDownloaderSyncSetting) error
	PageTasks(req dto.ModelDownloadTaskSearch) (int64, []dto.ModelDownloadTaskItem, error)
	CancelTask(req dto.ModelDownloadTaskOperate) error
	RetryTask(req dto.ModelDownloadTaskRetry) (*dto.ModelDownloadTaskItem, error)
	RemoveTask(req dto.ModelDownloadTaskOperate) error
	RecoverInterruptedTasks() error
	ListLocalModels(req dto.LocalModelSearch) (int64, []dto.LocalModelItem, error)
	DeleteLocalModel(req dto.LocalModelDelete) error
}

type ModelDownloaderService struct {
	client         *http.Client
	downloadClient *http.Client
	mu             sync.Mutex
	cancels        map[uint]context.CancelFunc
	downloadSem    chan struct{}
}

type modelDownloaderInputError struct {
	err error
}

func (m modelDownloaderInputError) Error() string {
	return m.err.Error()
}

func (m modelDownloaderInputError) Unwrap() error {
	return m.err
}

func newModelDownloaderInputError(format string, args ...interface{}) error {
	return modelDownloaderInputError{err: fmt.Errorf(format, args...)}
}

func newModelDownloaderRequestFailed(err error) error {
	if err == nil {
		err = errors.New("request failed")
	}
	if err.Error() == "" {
		err = errors.New("request failed")
	}
	return buserr.WithMap("ErrHttpReqFailed", map[string]interface{}{"err": err.Error()}, err)
}

func IsModelDownloaderInputError(err error) bool {
	var target modelDownloaderInputError
	return errors.As(err, &target)
}

func NewIModelDownloaderService() IModelDownloaderService {
	return &ModelDownloaderService{
		client:         &http.Client{Timeout: 30 * time.Second},
		downloadClient: &http.Client{},
		cancels:        make(map[uint]context.CancelFunc),
		downloadSem:    make(chan struct{}, 1),
	}
}

func (m *ModelDownloaderService) SearchHFModels(req dto.ModelSearch) (int64, []dto.ModelItem, error) {
	normalizeModelSearch(&req)

	var raw []hfModelResp
	if err := m.getJSON(context.Background(), buildHFModelSearchURL(hfEndpoint(), req), loadModelDownloaderHFToken(), &raw); err != nil {
		return 0, nil, err
	}
	items := make([]dto.ModelItem, 0, len(raw))
	for _, item := range raw {
		items = append(items, mapHFModel(item))
	}
	return estimateHFSearchTotal(req, len(items)), items, nil
}

func (m *ModelDownloaderService) GetHFModelInfo(req dto.ModelInfoReq) (*dto.ModelInfo, error) {
	repoID, err := normalizeModelDownloaderRepoID(req.RepoID)
	if err != nil {
		return nil, err
	}
	return m.loadHFModelInfo(context.Background(), repoID, req.Revision, loadModelDownloaderHFToken())
}

func (m *ModelDownloaderService) StartDownload(req dto.ModelDownload) (*dto.ModelDownloadTaskItem, error) {
	return m.startDownload(modelDownloaderSourceHuggingFace, req)
}

func (m *ModelDownloaderService) SearchModelScopeModels(req dto.ModelSearch) (int64, []dto.ModelItem, error) {
	normalizeModelSearch(&req)
	query := url.Values{}
	query.Set("page_number", fmt.Sprintf("%d", req.Page))
	query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	query.Set("sort", mapModelScopeSort(req.Sort))
	if keyword := strings.TrimSpace(req.Query); keyword != "" {
		query.Set("search", keyword)
	}

	var raw modelScopeSearchResp
	if err := m.getModelScopeOpenJSON(context.Background(), modelScopeEndpoint()+"/openapi/v1/models?"+query.Encode(), loadModelDownloaderModelScopeToken(), &raw); err != nil {
		return 0, nil, err
	}
	if !raw.Success && raw.Message != "" {
		return 0, nil, newModelDownloaderRequestFailed(errors.New(raw.Message))
	}
	items := make([]dto.ModelItem, 0, len(raw.Data.Models))
	for _, item := range raw.Data.Models {
		items = append(items, mapModelScopeModel(item))
	}
	return raw.Data.TotalCount, items, nil
}

func (m *ModelDownloaderService) GetModelScopeModelInfo(req dto.ModelInfoReq) (*dto.ModelInfo, error) {
	repoID, err := normalizeModelDownloaderRepoID(req.RepoID)
	if err != nil {
		return nil, err
	}
	return m.loadModelScopeModelInfo(context.Background(), repoID, req.Revision, loadModelDownloaderModelScopeToken())
}

func (m *ModelDownloaderService) StartModelScopeDownload(req dto.ModelDownload) (*dto.ModelDownloadTaskItem, error) {
	return m.startDownload(modelDownloaderSourceModelScope, req)
}

func (m *ModelDownloaderService) startDownload(source string, req dto.ModelDownload) (*dto.ModelDownloadTaskItem, error) {
	repoID, err := normalizeModelDownloaderRepoID(req.RepoID)
	if err != nil {
		return nil, err
	}
	if req.DownloadType == "" {
		req.DownloadType = "all"
	}
	if req.DownloadType != "all" && req.DownloadType != "tokenizer" {
		return nil, newModelDownloaderInputError("invalid download type")
	}
	modelName := modelNameFromRepo(repoID)
	targetDir := filepath.Join(loadModelDownloaderDir(), repoID)
	if req.DownloadType == "tokenizer" {
		digest := sha256.Sum256([]byte(source + "\x00" + repoID + "\x00" + resolveRevision(source, req.Revision)))
		targetDir = filepath.Join(global.CONF.Base.InstallDir, "1panel/ai/tokenizer", fmt.Sprintf("%s-%x", repoID, digest[:3]))
		modelName = filepath.Base(targetDir)
	}
	m.mu.Lock()
	if err := ensureTargetDirAvailable(targetDir); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	if err := m.ensureNoActiveDownload(source, repoID, targetDir, req.DownloadType); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	task := &model.ModelDownloadTask{
		DownloadType: req.DownloadType,
		Source:       source,
		RepoID:       repoID,
		Revision:     resolveRevision(source, req.Revision),
		ModelName:    modelName,
		TargetDir:    targetDir,
		Status:       constant.StatusWaiting,
	}
	if err := modelDownloadTaskRepo.Create(task); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	m.mu.Unlock()
	m.startTask(task.ID)
	return mapDownloadTask(*task), nil
}

func (m *ModelDownloaderService) GetSettings() (*dto.ModelDownloaderSetting, error) {
	syncStatus, err := settingRepo.GetValueByKey("FeaturedModelsSync")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		syncStatus = constant.StatusEnable
	} else if err != nil {
		return nil, err
	}
	return &dto.ModelDownloaderSetting{
		FeaturedModelsSync: syncStatus,
		ModelDir:           loadModelDownloaderDir(),
		HFEndpoint:         loadModelDownloaderHFEndpointSetting(),
		HFToken:            loadModelDownloaderHFToken(),
		ModelScopeEndpoint: modelScopeEndpoint(),
		ModelScopeToken:    loadModelDownloaderModelScopeToken(),
	}, nil
}

func (m *ModelDownloaderService) UpdateSettings(req dto.ModelDownloaderSetting) error {
	if err := settingRepo.UpdateOrCreate(modelDownloaderDirKey, resolveDownloaderModelDir(req.ModelDir)); err != nil {
		return err
	}
	if err := updateModelDownloaderEndpointSetting(modelDownloaderEndpointKey, req.HFEndpoint); err != nil {
		return err
	}
	if err := updateModelDownloaderSetting(modelDownloaderTokenKey, req.HFToken); err != nil {
		return err
	}
	if err := updateModelDownloaderEndpointSetting(modelDownloaderMSEndpointKey, req.ModelScopeEndpoint); err != nil {
		return err
	}
	return updateModelDownloaderSetting(modelDownloaderMSTokenKey, req.ModelScopeToken)
}

func (m *ModelDownloaderService) UpdateFeaturedModelsSync(req dto.ModelDownloaderSyncSetting) error {
	if req.Value != constant.StatusEnable && req.Value != constant.StatusDisable {
		return newModelDownloaderInputError("invalid sync setting")
	}
	return settingRepo.UpdateOrCreate("FeaturedModelsSync", req.Value)
}

func (m *ModelDownloaderService) PageTasks(req dto.ModelDownloadTaskSearch) (int64, []dto.ModelDownloadTaskItem, error) {
	if req.IncludeLocal {
		return pageModelDownloads(req)
	}
	var opts []repo.DBOption
	if req.DownloadType != "" {
		opts = append(opts, modelDownloadTaskRepo.WithByDownloadType(req.DownloadType))
	}
	total, list, err := modelDownloadTaskRepo.Page(req.Page, req.PageSize, req.Info, req.Status, opts...)
	if err != nil {
		return 0, nil, err
	}
	items := make([]dto.ModelDownloadTaskItem, 0, len(list))
	for _, item := range list {
		items = append(items, *mapDownloadTask(item))
	}
	return total, items, nil
}

func (m *ModelDownloaderService) CancelTask(req dto.ModelDownloadTaskOperate) error {
	task, err := modelDownloadTaskRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if task.Status != constant.StatusWaiting && task.Status != statusDownloading {
		return newModelDownloaderInputError("task is not cancellable")
	}
	m.mu.Lock()
	cancel := m.cancels[req.ID]
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	now := time.Now()
	return modelDownloadTaskRepo.Update(req.ID, map[string]interface{}{
		"status":       constant.StatusCanceled,
		"completed_at": &now,
	})
}

func (m *ModelDownloaderService) RetryTask(req dto.ModelDownloadTaskRetry) (*dto.ModelDownloadTaskItem, error) {
	task, err := modelDownloadTaskRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return nil, err
	}
	if task.Status != constant.StatusFailed && task.Status != constant.StatusCanceled {
		return nil, newModelDownloaderInputError("task is not retryable")
	}
	m.mu.Lock()
	if err := ensureTargetDirAvailable(task.TargetDir); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	if err := m.ensureNoActiveDownload(task.Source, task.RepoID, task.TargetDir, task.DownloadType); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	if err := modelDownloadTaskRepo.Update(task.ID, map[string]interface{}{
		"status":          constant.StatusWaiting,
		"progress":        0,
		"downloaded_size": int64(0),
		"error_message":   "",
		"started_at":      nil,
		"completed_at":    nil,
	}); err != nil {
		m.mu.Unlock()
		return nil, err
	}
	m.mu.Unlock()
	m.startTask(task.ID)
	task.Status = constant.StatusWaiting
	task.Progress = 0
	task.DownloadedSize = 0
	task.ErrorMessage = ""
	task.StartedAt = nil
	task.CompletedAt = nil
	return mapDownloadTask(task), nil
}

func (m *ModelDownloaderService) RemoveTask(req dto.ModelDownloadTaskOperate) error {
	task, err := modelDownloadTaskRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if task.Status == constant.StatusWaiting || task.Status == statusDownloading {
		return newModelDownloaderInputError("active task cannot be removed")
	}
	if req.WithFile {
		if err := m.deleteLocalModel(task.TargetDir, task.DownloadType, true); err != nil {
			return err
		}
		if task.DownloadType == "tokenizer" {
			return nil
		}
	}
	if err := os.RemoveAll(taskTempDir(task)); err != nil {
		return err
	}
	return modelDownloadTaskRepo.DeleteBy(repo.WithByID(req.ID))
}

func (m *ModelDownloaderService) RecoverInterruptedTasks() error {
	now := time.Now()
	return modelDownloadTaskRepo.UpdateBy(
		map[string]interface{}{
			"status":        constant.StatusFailed,
			"error_message": constant.InterruptedMsg,
			"completed_at":  &now,
		},
		modelDownloadTaskRepo.WithByStatusIn([]string{constant.StatusWaiting, statusDownloading}),
	)
}

func (m *ModelDownloaderService) ListLocalModels(req dto.LocalModelSearch) (int64, []dto.LocalModelItem, error) {
	items, err := listLocalModels(req)
	if err != nil {
		return 0, nil, err
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt > items[j].CreatedAt
	})
	return int64(len(items)), pageLocalModels(items, req.Page, req.PageSize), nil
}

func (m *ModelDownloaderService) DeleteLocalModel(req dto.LocalModelDelete) error {
	if req.Name != strings.TrimSpace(req.Name) {
		return newModelDownloaderInputError("invalid model directory")
	}
	if _, err := normalizeModelDownloaderRepoID(req.Name); err != nil {
		return err
	}
	root := loadModelDownloaderDir()
	if req.DownloadType == "tokenizer" {
		root = filepath.Join(global.CONF.Base.InstallDir, "1panel/ai/tokenizer")
	}
	items, err := listModelDownloadDirectories(root)
	if err != nil {
		return err
	}
	for _, item := range items {
		if item.Name == req.Name {
			return m.deleteLocalModel(item.Path, req.DownloadType, req.WithFile)
		}
	}
	return nil
}

func (m *ModelDownloaderService) deleteLocalModel(target, downloadType string, withFile bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	target = filepath.Clean(target)
	if target == "." || target == ".." || target == string(filepath.Separator) {
		return newModelDownloaderInputError("invalid model directory")
	}
	for _, path := range []string{filepath.Dir(target), target} {
		stat, err := os.Lstat(path)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if err == nil && (!stat.IsDir() || stat.Mode()&os.ModeSymlink != 0) {
			return newModelDownloaderInputError("invalid model directory")
		}
	}
	tasks, err := modelDownloadTaskRepo.List(modelDownloadTaskRepo.WithByTargetDir(target))
	if err != nil {
		return err
	}
	for _, task := range tasks {
		if task.Status == constant.StatusWaiting || task.Status == statusDownloading {
			return newModelDownloaderInputError("download target directory is still in use by an active download")
		}
	}
	if downloadType != "tokenizer" {
		return os.RemoveAll(target)
	}
	for _, task := range tasks {
		if task.DownloadType != "tokenizer" {
			continue
		}
		if err := os.RemoveAll(taskTempDir(task)); err != nil {
			return err
		}
	}
	if withFile {
		if err := os.RemoveAll(target); err != nil {
			return err
		}
	} else if err := os.Remove(filepath.Join(target, ".1panel-tokenizer")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return modelDownloadTaskRepo.DeleteBy(modelDownloadTaskRepo.WithByTargetDir(target), modelDownloadTaskRepo.WithByDownloadType("tokenizer"))
}

func pageModelDownloads(req dto.ModelDownloadTaskSearch) (int64, []dto.ModelDownloadTaskItem, error) {
	var opts []repo.DBOption
	if req.DownloadType != "" {
		opts = append(opts, modelDownloadTaskRepo.WithByDownloadType(req.DownloadType))
	}
	tasks, err := modelDownloadTaskRepo.List(opts...)
	if err != nil {
		return 0, nil, err
	}
	locals, err := listLocalModels(dto.LocalModelSearch{DownloadType: req.DownloadType})
	if err != nil {
		return 0, nil, err
	}
	localByPath := make(map[string]dto.LocalModelItem, len(locals))
	for _, item := range locals {
		localByPath[item.DownloadType+":"+filepath.Clean(item.Path)] = item
	}
	sort.Slice(tasks, func(i, j int) bool {
		iActive := tasks[i].Status == constant.StatusWaiting || tasks[i].Status == statusDownloading
		jActive := tasks[j].Status == constant.StatusWaiting || tasks[j].Status == statusDownloading
		if iActive != jActive {
			return iActive
		}
		return tasks[i].ID > tasks[j].ID
	})
	items := make([]dto.ModelDownloadTaskItem, 0, len(tasks)+len(locals))
	for _, task := range tasks {
		item := *mapDownloadTask(task)
		key := task.DownloadType + ":" + filepath.Clean(task.TargetDir)
		if local, ok := localByPath[key]; ok {
			item.LocalModel = &local
			delete(localByPath, key)
		}
		items = append(items, item)
	}
	for _, local := range localByPath {
		repoID := local.RepoID
		if repoID == "" {
			repoID = local.Name
		}
		items = append(items, dto.ModelDownloadTaskItem{
			DownloadType:   local.DownloadType,
			Source:         local.Source,
			RepoID:         repoID,
			TargetDir:      local.Path,
			Status:         constant.StatusSuccess,
			Progress:       100,
			TotalSize:      local.Size,
			DownloadedSize: local.Size,
			CreatedAt:      local.CreatedAt,
			LocalModel:     &local,
		})
	}
	filtered := items[:0]
	for _, item := range items {
		if req.Status != "" && item.Status != req.Status {
			continue
		}
		if req.Info != "" && !strings.Contains(strings.ToLower(item.RepoID), strings.ToLower(req.Info)) {
			continue
		}
		filtered = append(filtered, item)
	}
	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].CreatedAt != filtered[j].CreatedAt {
			return filtered[i].CreatedAt > filtered[j].CreatedAt
		}
		if filtered[i].ID != filtered[j].ID {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].TargetDir < filtered[j].TargetDir
	})
	total := int64(len(filtered))
	page, size := req.Page, req.PageSize
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	start := (page - 1) * size
	if start >= len(filtered) {
		return total, []dto.ModelDownloadTaskItem{}, nil
	}
	end := start + size
	if end > len(filtered) {
		end = len(filtered)
	}
	return total, filtered[start:end], nil
}

func listLocalModels(req dto.LocalModelSearch) ([]dto.LocalModelItem, error) {
	items := make([]dto.LocalModelItem, 0)
	if req.DownloadType != "all" {
		tokenizers, err := listDownloadedTokenizers(req.Info)
		if err != nil {
			return nil, err
		}
		items = append(items, tokenizers...)
	}
	if req.DownloadType != "tokenizer" {
		models, err := listModelDownloadDirectories(loadModelDownloaderDir())
		if err != nil {
			return nil, err
		}
		for _, item := range models {
			if req.Info != "" && !strings.Contains(strings.ToLower(item.Name), strings.ToLower(req.Info)) {
				continue
			}
			info, err := os.Stat(item.Path)
			if err != nil {
				continue
			}
			item.DownloadType = "all"
			item.Size, _ = files.NewFileOp().GetDirSize(item.Path)
			item.SizeFormatted = formatBytes(item.Size)
			item.CreatedAt = info.ModTime().Format("2006-01-02 15:04:05")
			items = append(items, item)
		}
	}

	return items, nil
}

func (m *ModelDownloaderService) ensureNoActiveDownload(source, repoID, targetDir, downloadType string) error {
	_, err := modelDownloadTaskRepo.GetFirst(
		modelDownloadTaskRepo.WithBySource(source),
		modelDownloadTaskRepo.WithByRepoID(repoID),
		modelDownloadTaskRepo.WithByDownloadType(downloadType),
		modelDownloadTaskRepo.WithByStatusIn([]string{constant.StatusWaiting, statusDownloading}),
	)
	if err == nil {
		return newModelDownloaderInputError("download for %s is already in progress", repoID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	_, err = modelDownloadTaskRepo.GetFirst(
		modelDownloadTaskRepo.WithByTargetDir(targetDir),
		modelDownloadTaskRepo.WithByStatusIn([]string{constant.StatusWaiting, statusDownloading}),
	)
	if err == nil {
		return newModelDownloaderInputError("download target directory is already in progress: %s", targetDir)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return nil
}

func (m *ModelDownloaderService) startTask(taskID uint) {
	ctx, cancel := context.WithCancel(context.Background())
	m.mu.Lock()
	m.cancels[taskID] = cancel
	m.mu.Unlock()
	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.cancels, taskID)
			m.mu.Unlock()
		}()
		m.downloadTask(ctx, taskID)
	}()
}

func (m *ModelDownloaderService) downloadTask(ctx context.Context, taskID uint) {
	select {
	case m.downloadSem <- struct{}{}:
		defer func() { <-m.downloadSem }()
	case <-ctx.Done():
		return
	}
	task, err := modelDownloadTaskRepo.GetFirst(repo.WithByID(taskID))
	if err != nil {
		return
	}
	if ctx.Err() != nil {
		return
	}
	now := time.Now()
	_ = modelDownloadTaskRepo.Update(task.ID, map[string]interface{}{
		"status":     statusDownloading,
		"started_at": &now,
	})

	if err := m.runDownload(ctx, task); err != nil {
		if ctx.Err() != nil {
			done := time.Now()
			_ = modelDownloadTaskRepo.Update(task.ID, map[string]interface{}{
				"status":       constant.StatusCanceled,
				"completed_at": &done,
			})
			return
		}
		done := time.Now()
		_ = modelDownloadTaskRepo.Update(task.ID, map[string]interface{}{
			"status":        constant.StatusFailed,
			"error_message": err.Error(),
			"completed_at":  &done,
		})
	}
}

func (m *ModelDownloaderService) runDownload(ctx context.Context, task model.ModelDownloadTask) error {
	info, err := m.loadModelInfo(ctx, task)
	if err != nil {
		return err
	}
	var downloadFiles []dto.ModelFile
	if task.DownloadType == "tokenizer" {
		downloadFiles = filterTokenizerDownloadFiles(info.Files)
		if err := validateTokenizerDownloadFiles(downloadFiles); err != nil {
			return err
		}
	} else {
		downloadFiles = filterModelDownloadFiles(info.Files)
	}
	if len(downloadFiles) == 0 {
		return newModelDownloaderInputError("no downloadable model files found: %s", task.RepoID)
	}
	totalSize := int64(0)
	for _, item := range downloadFiles {
		totalSize += item.Size
	}
	if err := os.MkdirAll(filepath.Dir(task.TargetDir), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(task.TargetDir); err == nil {
		return newModelDownloaderInputError("model directory already exists: %s", task.TargetDir)
	}
	tmpDir := taskTempDir(task)
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		return err
	}
	downloaded := int64(0)
	for _, file := range downloadFiles {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := m.downloadModelFile(ctx, task, file, tmpDir, &downloaded, totalSize); err != nil {
			return err
		}
	}
	if task.DownloadType == "tokenizer" {
		if err := validateDownloadedTokenizer(tmpDir); err != nil {
			return err
		}
		metadata, err := json.Marshal(tokenizerDownloadMetadata{Source: task.Source, RepoID: task.RepoID, Revision: task.Revision})
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(tmpDir, ".1panel-tokenizer"), metadata, 0600); err != nil {
			return err
		}
		totalSize = downloaded
	}
	if err := os.Rename(tmpDir, task.TargetDir); err != nil {
		return err
	}
	done := time.Now()
	return modelDownloadTaskRepo.Update(task.ID, map[string]interface{}{
		"status":          constant.StatusSuccess,
		"progress":        float64(100),
		"total_size":      totalSize,
		"downloaded_size": totalSize,
		"completed_at":    &done,
		"error_message":   "",
	})
}

func (m *ModelDownloaderService) loadModelInfo(ctx context.Context, task model.ModelDownloadTask) (*dto.ModelInfo, error) {
	switch task.Source {
	case modelDownloaderSourceHuggingFace:
		return m.loadHFModelInfo(ctx, task.RepoID, task.Revision, loadModelDownloaderHFToken())
	case modelDownloaderSourceModelScope:
		return m.loadModelScopeModelInfo(ctx, task.RepoID, task.Revision, loadModelDownloaderModelScopeToken())
	default:
		return nil, newModelDownloaderInputError("unsupported model source: %s", task.Source)
	}
}

func (m *ModelDownloaderService) downloadModelFile(ctx context.Context, task model.ModelDownloadTask, file dto.ModelFile, tmpDir string, downloaded *int64, totalSize int64) error {
	target, err := safeJoinModelFile(tmpDir, file.Name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	existing := int64(0)
	if stat, err := os.Stat(target); err == nil {
		existing = stat.Size()
	}
	if task.DownloadType == "tokenizer" && existing > tokenizerDownloadMaxSize-*downloaded {
		return newModelDownloaderInputError("tokenizer files exceed 256 MiB")
	}
	if file.Size > 0 && existing == file.Size {
		*downloaded += existing
		return m.updateProgress(task.ID, *downloaded, totalSize)
	}
	fileURL, err := modelFileResolveURL(task.Source, task.RepoID, task.Revision, file.Name)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return err
	}
	applyDownloadAuth(req, task.Source)
	if existing > 0 && (file.Size == 0 || existing < file.Size) {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", existing))
	}
	resp, err := m.downloadClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("download %s failed: %s", file.Name, resp.Status)
	}
	flag := os.O_CREATE | os.O_WRONLY
	if resp.StatusCode == http.StatusPartialContent {
		flag |= os.O_APPEND
		*downloaded += existing
	} else {
		flag |= os.O_TRUNC
	}
	out, err := os.OpenFile(target, flag, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	writer := &downloadProgressWriter{
		writer: out,
		onProgress: func(n int64) {
			*downloaded += n
			_ = m.updateProgress(task.ID, *downloaded, totalSize)
		},
	}
	var input io.Reader = resp.Body
	if task.DownloadType == "tokenizer" {
		remaining := tokenizerDownloadMaxSize - *downloaded
		if remaining < 0 || resp.ContentLength > remaining {
			return newModelDownloaderInputError("tokenizer files exceed 256 MiB")
		}
		input = io.LimitReader(resp.Body, remaining+1)
	}
	_, err = io.Copy(writer, input)
	if writer.updated > 0 {
		writer.onProgress(writer.updated)
		writer.updated = 0
	}
	if task.DownloadType == "tokenizer" && *downloaded > tokenizerDownloadMaxSize {
		return newModelDownloaderInputError("tokenizer files exceed 256 MiB")
	}
	return err
}

func (m *ModelDownloaderService) updateProgress(taskID uint, downloadedSize, totalSize int64) error {
	progress := float64(0)
	if totalSize > 0 {
		progress = float64(downloadedSize) / float64(totalSize) * 100
		if progress > 99 {
			progress = 99
		}
	}
	return modelDownloadTaskRepo.Update(taskID, map[string]interface{}{
		"total_size":      totalSize,
		"downloaded_size": downloadedSize,
		"progress":        progress,
	})
}

func (m *ModelDownloaderService) loadHFModelInfo(ctx context.Context, repoID, revision, hfToken string) (*dto.ModelInfo, error) {
	var raw hfModelResp
	if err := m.getJSON(ctx, hfModelInfoURL(repoID, revision), hfToken, &raw); err != nil {
		return nil, err
	}
	item := mapHFModel(raw)
	filesList := make([]dto.ModelFile, 0, len(raw.Siblings))
	totalSize := int64(0)
	for _, sibling := range raw.Siblings {
		if sibling.RFilename == "" {
			continue
		}
		size := sibling.Size
		totalSize += size
		filesList = append(filesList, dto.ModelFile{
			Name:          sibling.RFilename,
			Size:          size,
			SizeFormatted: formatBytes(size),
		})
	}
	if item.Size == 0 {
		item.Size = totalSize
		item.SizeFormatted = formatBytes(totalSize)
	}
	card := m.loadHFModelCard(ctx, repoID, resolveRevision(modelDownloaderSourceHuggingFace, revision), hfToken)
	return &dto.ModelInfo{
		ModelItem: item,
		ModelCard: card,
		Files:     filesList,
	}, nil
}

func (m *ModelDownloaderService) loadModelScopeModelInfo(ctx context.Context, repoID, revision, token string) (*dto.ModelInfo, error) {
	revision = resolveRevision(modelDownloaderSourceModelScope, revision)
	var detail modelScopeDetailResp
	detailURL := modelScopeEndpoint() + "/api/v1/models/" + escapedRepoPath(repoID) + "?Revision=" + url.QueryEscape(revision)
	if err := m.getModelScopeAPIJSON(ctx, detailURL, token, &detail); err != nil {
		return nil, err
	}
	if detail.Code != 0 && detail.Code != http.StatusOK {
		return nil, newModelDownloaderRequestFailed(errors.New(detail.Message))
	}

	var fileResp modelScopeFileListResp
	fileURL := modelScopeEndpoint() + "/api/v1/models/" + escapedRepoPath(repoID) + "/repo/files?Revision=" + url.QueryEscape(revision) + "&Recursive=true"
	if err := m.getModelScopeAPIJSON(ctx, fileURL, token, &fileResp); err != nil {
		return nil, err
	}
	if fileResp.Code != 0 && fileResp.Code != http.StatusOK {
		return nil, newModelDownloaderRequestFailed(errors.New(fileResp.Message))
	}

	filesList := make([]dto.ModelFile, 0, len(fileResp.Data.Files))
	totalSize := int64(0)
	for _, file := range fileResp.Data.Files {
		if file.Path == "" || (file.Type != "" && file.Type != "blob") {
			continue
		}
		name := strings.TrimLeft(file.Path, "/")
		if name == ".gitattributes" || name == ".gitignore" {
			continue
		}
		totalSize += file.Size
		filesList = append(filesList, dto.ModelFile{
			Name:          name,
			Size:          file.Size,
			SizeFormatted: formatBytes(file.Size),
		})
	}

	item := mapModelScopeModel(modelScopeModelResp{
		ID:          repoID,
		Downloads:   detail.Data.Downloads,
		Likes:       detail.Data.Stars,
		FileSize:    totalSize,
		DisplayName: detail.Data.Name,
	})
	item.Size = totalSize
	item.SizeFormatted = formatBytes(totalSize)
	return &dto.ModelInfo{
		ModelItem: item,
		ModelCard: strings.TrimSpace(detail.Data.ReadMeContent),
		Files:     filesList,
	}, nil
}

func (m *ModelDownloaderService) loadHFModelCard(ctx context.Context, repoID, revision, hfToken string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, hfResolveURL(repoID, revision, "README.md"), nil)
	if err != nil {
		return ""
	}
	if hfToken != "" {
		req.Header.Set("Authorization", "Bearer "+hfToken)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return ""
	}
	card := string(data)
	if strings.HasPrefix(card, "---") {
		if index := strings.Index(card[3:], "---"); index >= 0 {
			card = strings.TrimSpace(card[index+6:])
		}
	}
	return card
}

func (m *ModelDownloaderService) getJSON(ctx context.Context, target, token string, out interface{}) error {
	return m.getModelDownloaderJSON(ctx, target, out, func(req *http.Request) {
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	})
}

func (m *ModelDownloaderService) getModelScopeOpenJSON(ctx context.Context, target, token string, out interface{}) error {
	return m.getModelDownloaderJSON(ctx, target, out, func(req *http.Request) {
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	})
}

func (m *ModelDownloaderService) getModelScopeAPIJSON(ctx context.Context, target, token string, out interface{}) error {
	return m.getModelDownloaderJSON(ctx, target, out, func(req *http.Request) {
		applyModelScopeAPIToken(req, token)
	})
}

func (m *ModelDownloaderService) getModelDownloaderJSON(ctx context.Context, target string, out interface{}, configure func(*http.Request)) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return err
	}
	if configure != nil {
		configure(req)
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return newModelDownloaderRequestFailed(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return newModelDownloaderRequestFailed(errors.New(resp.Status))
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

type downloadProgressWriter struct {
	writer     io.Writer
	updated    int64
	onProgress func(int64)
}

func (d *downloadProgressWriter) Write(p []byte) (int, error) {
	n, err := d.writer.Write(p)
	if n > 0 {
		d.updated += int64(n)
		if d.updated >= 2*1024*1024 {
			d.onProgress(d.updated)
			d.updated = 0
		}
	}
	return n, err
}

type hfModelResp struct {
	ID          string        `json:"id"`
	ModelID     string        `json:"modelId"`
	Downloads   int64         `json:"downloads"`
	Likes       int64         `json:"likes"`
	Safetensors hfSafetensors `json:"safetensors"`
	Siblings    []struct {
		RFilename string `json:"rfilename"`
		Size      int64  `json:"size"`
	} `json:"siblings"`
}

type hfSafetensors struct {
	Parameters map[string]float64 `json:"parameters"`
}

type modelScopeSearchResp struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Models     []modelScopeModelResp `json:"models"`
		TotalCount int64                 `json:"total_count"`
	} `json:"data"`
}

type modelScopeModelResp struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	ModelID     string `json:"model_id"`
	DisplayName string `json:"display_name"`
	Downloads   int64  `json:"downloads"`
	Likes       int64  `json:"likes"`
	FileSize    int64  `json:"file_size"`
	Params      int64  `json:"params"`
}

type modelScopeDetailResp struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    struct {
		Name          string `json:"Name"`
		Downloads     int64  `json:"Downloads"`
		Stars         int64  `json:"Stars"`
		ReadMeContent string `json:"ReadMeContent"`
	} `json:"Data"`
}

type modelScopeFileListResp struct {
	Code    int    `json:"Code"`
	Message string `json:"Message"`
	Data    struct {
		Files []modelScopeFileResp `json:"Files"`
	} `json:"Data"`
}

type modelScopeFileResp struct {
	Path string `json:"Path"`
	Type string `json:"Type"`
	Size int64  `json:"Size"`
}

func listModelDownloadDirectories(root string) ([]dto.LocalModelItem, error) {
	entries, err := os.ReadDir(root)
	if os.IsNotExist(err) {
		return []dto.LocalModelItem{}, nil
	}
	if err != nil {
		return nil, err
	}
	var items []dto.LocalModelItem
	for _, entry := range entries {
		if !entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		children, err := os.ReadDir(filepath.Join(root, entry.Name()))
		if err != nil {
			return nil, err
		}
		for _, child := range children {
			if child.IsDir() && !strings.HasPrefix(child.Name(), ".") {
				items = append(items, dto.LocalModelItem{
					Name: entry.Name() + "/" + child.Name(),
					Path: filepath.Join(root, entry.Name(), child.Name()),
				})
			}
		}
	}
	return items, nil
}

func normalizeModelDownloaderRepoID(repoID string) (string, error) {
	repoID = strings.TrimSpace(repoID)
	if !re.GetRegex(re.ModelDownloaderRepoIDPattern).MatchString(repoID) {
		return "", newModelDownloaderInputError("invalid repository ID: %s", repoID)
	}
	if strings.Contains(repoID, "..") {
		return "", newModelDownloaderInputError("invalid repository ID: %s", repoID)
	}
	return repoID, nil
}

func resolveDownloaderModelDir(dir string) string {
	if strings.TrimSpace(dir) == "" {
		return defaultModelDownloaderDir
	}
	return filepath.Clean(strings.TrimSpace(dir))
}

func ensureTargetDirAvailable(targetDir string) error {
	if _, err := os.Stat(targetDir); err == nil {
		return newModelDownloaderInputError("model directory already exists: %s", targetDir)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func filterModelDownloadFiles(items []dto.ModelFile) []dto.ModelFile {
	hasSafeTensors := false
	for _, item := range items {
		if strings.HasSuffix(strings.ToLower(item.Name), ".safetensors") {
			hasSafeTensors = true
			break
		}
	}
	filtered := make([]dto.ModelFile, 0, len(items))
	for _, item := range items {
		name := strings.ToLower(filepath.ToSlash(item.Name))
		if hasSafeTensors {
			if strings.HasSuffix(name, ".bin") || strings.HasPrefix(name, "original/") {
				continue
			}
			if strings.HasPrefix(filepath.Base(name), "consolidated.") && strings.HasSuffix(name, ".pth") {
				continue
			}
		}
		filtered = append(filtered, item)
	}
	return filtered
}

func mapHFModel(item hfModelResp) dto.ModelItem {
	repoID := item.ID
	if repoID == "" {
		repoID = item.ModelID
	}
	size := calcSafetensorsSize(item.Safetensors)
	return dto.ModelItem{
		RepoID:        repoID,
		Name:          modelNameFromRepo(repoID),
		Downloads:     item.Downloads,
		Likes:         item.Likes,
		Size:          size,
		SizeFormatted: formatBytes(size),
	}
}

func mapModelScopeModel(item modelScopeModelResp) dto.ModelItem {
	repoID := item.ID
	if repoID == "" {
		repoID = item.ModelID
	}
	return dto.ModelItem{
		RepoID:        repoID,
		Name:          modelNameFromRepo(repoID),
		Downloads:     item.Downloads,
		Likes:         item.Likes,
		Size:          item.FileSize,
		SizeFormatted: formatBytes(item.FileSize),
	}
}

func calcSafetensorsSize(safetensors hfSafetensors) int64 {
	dtypeBytes := map[string]int64{
		"F64": 8, "F32": 4, "F16": 2, "BF16": 2,
		"I64": 8, "I32": 4, "I16": 2, "I8": 1,
		"U64": 8, "U32": 4, "U16": 2, "U8": 1,
		"BOOL": 1,
	}
	total := int64(0)
	for dtype, num := range safetensors.Parameters {
		width := dtypeBytes[strings.ToUpper(dtype)]
		if width == 0 {
			width = 1
		}
		total += int64(num) * width
	}
	return total
}

func formatBytes(size int64) string {
	if size <= 0 {
		return ""
	}
	const unit = 1024
	if size < unit*unit {
		return fmt.Sprintf("%.1f KB", float64(size)/unit)
	}
	if size < unit*unit*unit {
		return fmt.Sprintf("%.1f MB", float64(size)/(unit*unit))
	}
	return fmt.Sprintf("%.1f GB", float64(size)/(unit*unit*unit))
}

func mapDownloadTask(item model.ModelDownloadTask) *dto.ModelDownloadTaskItem {
	status := item.Status
	return &dto.ModelDownloadTaskItem{
		DownloadType:    item.DownloadType,
		ID:              item.ID,
		Source:          item.Source,
		RepoID:          item.RepoID,
		Revision:        item.Revision,
		ModelName:       item.ModelName,
		TargetDir:       item.TargetDir,
		Status:          status,
		Progress:        item.Progress,
		TotalSize:       item.TotalSize,
		DownloadedSize:  item.DownloadedSize,
		ErrorMessage:    item.ErrorMessage,
		CreatedAt:       item.CreatedAt.Format("2006-01-02 15:04:05"),
		StartedAt:       formatOptionalTime(item.StartedAt),
		CompletedAt:     formatOptionalTime(item.CompletedAt),
		CanCancel:       status == constant.StatusWaiting || status == statusDownloading,
		CanRetry:        status == constant.StatusFailed || status == constant.StatusCanceled,
		CanRemoveRecord: status == constant.StatusSuccess || status == constant.StatusFailed || status == constant.StatusCanceled,
	}
}

func formatOptionalTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func pageLocalModels(items []dto.LocalModelItem, page, pageSize int) []dto.LocalModelItem {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start >= len(items) {
		return []dto.LocalModelItem{}
	}
	end := start + pageSize
	if end > len(items) {
		end = len(items)
	}
	return items[start:end]
}

func modelNameFromRepo(repoID string) string {
	parts := strings.Split(repoID, "/")
	if len(parts) == 0 {
		return repoID
	}
	return parts[len(parts)-1]
}

func hfEndpoint() string {
	endpoint := loadModelDownloaderHFEndpointSetting()
	if endpoint == "" {
		endpoint = strings.TrimRight(os.Getenv("HF_ENDPOINT"), "/")
	}
	if endpoint == "" {
		endpoint = defaultHFEndpoint
	}
	return endpoint
}

func modelScopeEndpoint() string {
	endpoint := loadModelDownloaderModelScopeEndpointSetting()
	if endpoint == "" {
		endpoint = defaultModelScopeEndpoint
	}
	return endpoint
}

func normalizeModelSearch(req *dto.ModelSearch) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		if req.Limit > 0 {
			req.PageSize = req.Limit
		} else {
			req.PageSize = 50
		}
	}
	req.Limit = req.PageSize
	if req.Sort == "" {
		req.Sort = "trending"
	}
}

func buildHFModelSearchURL(endpoint string, req dto.ModelSearch) string {
	query := url.Values{}
	if keyword := strings.TrimSpace(req.Query); keyword != "" {
		query.Set("search", keyword)
	}
	query.Set("sort", mapHFSort(req.Sort))
	query.Set("direction", "-1")
	if req.Sort == "least_params" {
		query.Set("direction", "1")
	}
	query.Set("limit", fmt.Sprintf("%d", req.PageSize))
	query.Set("skip", fmt.Sprintf("%d", (req.Page-1)*req.PageSize))
	query.Set("full", "true")
	return endpoint + "/api/models?" + query.Encode()
}

func estimateHFSearchTotal(req dto.ModelSearch, itemCount int) int64 {
	offset := int64((req.Page - 1) * req.PageSize)
	total := offset + int64(itemCount)
	if itemCount == req.PageSize {
		total++
	}
	return total
}

func mapHFSort(sort string) string {
	switch sort {
	case "downloads":
		return "downloads"
	case "likes":
		return "likes"
	case "updated":
		return "lastModified"
	case "created":
		return "createdAt"
	case "most_params", "least_params":
		return "num_parameters"
	default:
		return "trendingScore"
	}
}

func mapModelScopeSort(sort string) string {
	switch sort {
	case "downloads":
		return "downloads"
	case "likes":
		return "likes"
	case "updated":
		return "last_modified"
	default:
		return "default"
	}
}

func resolveRevision(source, revision string) string {
	if strings.TrimSpace(revision) == "" {
		if source == modelDownloaderSourceModelScope {
			return "master"
		}
		return "main"
	}
	return strings.TrimSpace(revision)
}

func escapedRepoPath(repoID string) string {
	parts := strings.Split(repoID, "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func hfResolveURL(repoID, revision, filename string) string {
	return hfEndpoint() + "/" + escapedRepoPath(repoID) + "/resolve/" + url.PathEscape(resolveRevision(modelDownloaderSourceHuggingFace, revision)) + "/" + pathEscapeFile(filename)
}

func hfModelInfoURL(repoID, revision string) string {
	query := url.Values{}
	query.Set("blobs", "true")
	query.Set("revision", resolveRevision(modelDownloaderSourceHuggingFace, revision))
	return hfEndpoint() + "/api/models/" + escapedRepoPath(repoID) + "?" + query.Encode()
}

func modelFileResolveURL(source, repoID, revision, filename string) (string, error) {
	switch source {
	case modelDownloaderSourceHuggingFace:
		return hfResolveURL(repoID, revision, filename), nil
	case modelDownloaderSourceModelScope:
		return buildModelScopeResolveURL(modelScopeEndpoint(), repoID, revision, filename), nil
	default:
		return "", newModelDownloaderInputError("unsupported model source: %s", source)
	}
}

func buildModelScopeResolveURL(endpoint, repoID, revision, filename string) string {
	query := url.Values{}
	query.Set("Revision", resolveRevision(modelDownloaderSourceModelScope, revision))
	query.Set("FilePath", filename)
	return strings.TrimRight(endpoint, "/") + "/api/v1/models/" + escapedRepoPath(repoID) + "/repo?" + query.Encode()
}

func applyDownloadAuth(req *http.Request, source string) {
	if source == modelDownloaderSourceModelScope {
		applyModelScopeAPIToken(req, loadModelDownloaderModelScopeToken())
		return
	}
	if token := loadModelDownloaderHFToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func applyModelScopeAPIToken(req *http.Request, token string) {
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "m_session_id", Value: token})
	}
}

func pathEscapeFile(filename string) string {
	parts := strings.Split(filepath.ToSlash(filename), "/")
	for i := range parts {
		parts[i] = url.PathEscape(parts[i])
	}
	return strings.Join(parts, "/")
}

func safeJoinModelFile(baseDir, filename string) (string, error) {
	if strings.TrimSpace(filename) == "" {
		return "", fmt.Errorf("invalid model file path")
	}
	normalized := filepath.FromSlash(strings.ReplaceAll(filename, "\\", "/"))
	if filepath.IsAbs(normalized) {
		return "", fmt.Errorf("invalid model file path: %s", filename)
	}
	cleanName := filepath.Clean(normalized)
	if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid model file path: %s", filename)
	}
	base := filepath.Clean(baseDir)
	target := filepath.Join(base, cleanName)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("invalid model file path: %s", filename)
	}
	return target, nil
}

func taskTempDir(task model.ModelDownloadTask) string {
	return filepath.Join(filepath.Dir(task.TargetDir), fmt.Sprintf(".%s-%d.download", task.ModelName, task.ID))
}

func loadModelDownloaderDir() string {
	value, err := settingRepo.GetValueByKey(modelDownloaderDirKey)
	if err != nil {
		return defaultModelDownloaderDir
	}
	return resolveDownloaderModelDir(value)
}

func loadModelDownloaderHFEndpointSetting() string {
	return loadModelDownloaderEndpointSetting(modelDownloaderEndpointKey)
}

func loadModelDownloaderHFToken() string {
	return loadModelDownloaderSetting(modelDownloaderTokenKey)
}

func loadModelDownloaderModelScopeEndpointSetting() string {
	return loadModelDownloaderEndpointSetting(modelDownloaderMSEndpointKey)
}

func loadModelDownloaderModelScopeToken() string {
	return loadModelDownloaderSetting(modelDownloaderMSTokenKey)
}

func updateModelDownloaderEndpointSetting(key, value string) error {
	return updateModelDownloaderSetting(key, cleanModelDownloaderEndpoint(value))
}

func updateModelDownloaderSetting(key, value string) error {
	return settingRepo.UpdateOrCreate(key, strings.TrimSpace(value))
}

func loadModelDownloaderEndpointSetting(key string) string {
	return cleanModelDownloaderEndpoint(loadModelDownloaderSetting(key))
}

func loadModelDownloaderSetting(key string) string {
	value, err := settingRepo.GetValueByKey(key)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(value)
}

func cleanModelDownloaderEndpoint(value string) string {
	return strings.TrimRight(strings.TrimSpace(value), "/")
}
