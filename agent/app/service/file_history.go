package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"gorm.io/gorm"
	"strconv"
)

const (
	fileHistorySettingEnable      = "FileHistoryStatus"
	fileHistorySettingMaxPerPath  = "FileHistoryMaxPerPath"
	fileHistorySettingDiskQuotaMB = "FileHistoryDiskQuotaMB"
	fileHistoryOpSave             = "save"
	fileHistoryOpRestore          = "restore"
	fileHistoryOpRename           = "rename"
	fileHistoryOpMove             = "move"
	fileHistoryRootDirName        = "file-history"
	defaultFileHistoryMaxPerPath  = 20
	defaultFileHistoryDiskQuotaMB = 1024
)

var historyService = NewIFileHistoryService()
var fileHistoryPathLocks sync.Map

type FileHistoryService struct {
	repo repo.IFileHistoryRepo
}

type IFileHistoryService interface {
	GetSettingInfo() (*response.FileHistorySettingInfo, error)
	UpdateSetting(req request.FileHistorySettingUpdate) error
	RecordSave(path string, content []byte, fileMode os.FileMode) error
	HasRelatedHistory(path string) (bool, error)
	RecordOperation(operation string, filePath string, content []byte, fileMode os.FileMode, sourcePath string, targetPath string) error
	Restore(req request.FileHistoryRestoreReq) (response.FileInfo, error)
	Search(req request.FileHistorySearchReq) (int64, []response.FileHistoryInfo, error)
	GetContent(req request.FileHistoryContentReq) (response.FileHistoryInfo, error)
	Delete(req request.FileHistoryDeleteReq) error
	DeleteRelatedHistory(path string) error
}

func NewIFileHistoryService() IFileHistoryService {
	return &FileHistoryService{repo: repo.NewIFileHistoryRepo()}
}

func (s *FileHistoryService) GetSettingInfo() (*response.FileHistorySettingInfo, error) {
	info := &response.FileHistorySettingInfo{
		Enable:      constant.StatusEnable,
		MaxPerPath:  defaultFileHistoryMaxPerPath,
		DiskQuotaMB: defaultFileHistoryDiskQuotaMB,
	}
	if value, err := settingRepo.GetValueByKey(fileHistorySettingEnable); err == nil && value != "" {
		info.Enable = value
	}
	if value, err := settingRepo.GetValueByKey(fileHistorySettingMaxPerPath); err == nil && value != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed > 0 {
			info.MaxPerPath = parsed
		}
	}
	if value, err := settingRepo.GetValueByKey(fileHistorySettingDiskQuotaMB); err == nil && value != "" {
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil && parsed > 0 {
			info.DiskQuotaMB = parsed
		}
	}
	return info, nil
}

func (s *FileHistoryService) UpdateSetting(req request.FileHistorySettingUpdate) error {
	if req.MaxPerPath <= 0 {
		req.MaxPerPath = defaultFileHistoryMaxPerPath
	}
	if req.DiskQuotaMB <= 0 {
		req.DiskQuotaMB = defaultFileHistoryDiskQuotaMB
	}
	if err := settingRepo.UpdateOrCreate(fileHistorySettingEnable, req.Enable); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate(fileHistorySettingMaxPerPath, fmt.Sprintf("%d", req.MaxPerPath)); err != nil {
		return err
	}
	if err := settingRepo.UpdateOrCreate(fileHistorySettingDiskQuotaMB, fmt.Sprintf("%d", req.DiskQuotaMB)); err != nil {
		return err
	}
	return nil
}

func (s *FileHistoryService) RecordSave(filePath string, content []byte, fileMode os.FileMode) error {
	return s.RecordOperation(fileHistoryOpSave, filePath, content, fileMode, "", "")
}

func (s *FileHistoryService) HasRelatedHistory(filePath string) (bool, error) {
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	if _, err = s.getLatestActiveRelatedByPath(absPath); err == nil {
		return true, nil
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	if _, err = s.getLatestRelatedByPath(absPath); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *FileHistoryService) RecordOperation(operation string, filePath string, content []byte, fileMode os.FileMode, sourcePath string, targetPath string) error {
	info, err := s.GetSettingInfo()
	if err != nil {
		return err
	}
	if info.Enable != constant.StatusEnable {
		return nil
	}
	if fileMode == 0 {
		fileMode = 0640
	}
	if content == nil {
		content = []byte{}
	}

	absPath, err := filepath.Abs(filePath)
	if err != nil {
		absPath = filePath
	}
	absSourcePath := normalizeAbsPath(sourcePath)
	absTargetPath := normalizeAbsPath(targetPath)
	resolvePaths := []string{absPath}
	recordPath := absPath
	switch operation {
	case fileHistoryOpRename, fileHistoryOpMove:
		resolvePaths = []string{absSourcePath, absTargetPath, absPath}
		if absTargetPath != "" {
			recordPath = absTargetPath
		}
	default:
		if absSourcePath != "" {
			resolvePaths = append(resolvePaths, absSourcePath)
		}
		if absTargetPath != "" {
			resolvePaths = append(resolvePaths, absTargetPath)
		}
	}

	fileID, latestChainRecord, chainErr := s.resolveFileChain(resolvePaths...)
	if chainErr != nil && !errors.Is(chainErr, gorm.ErrRecordNotFound) {
		return chainErr
	}
	if fileID == "" {
		fileID = common.GetUuid()
	}
	if operation == fileHistoryOpRename || operation == fileHistoryOpMove {
		if !s.isEditableHistorySnapshot(fileMode, content) {
			return nil
		}
	}

	previousID := uint(0)
	if latestVersion, err := s.getLatestVersionByFileID(fileID); err == nil {
		switch operation {
		case fileHistoryOpSave:
			contentHash := sha256HexBytes(content)
			if latestVersion.ContentSHA == contentHash {
				return nil
			}
			previousID = latestVersion.ID
		case fileHistoryOpRestore:
			previousID = latestVersion.ID
		default:
			if latestChainRecord.ID != 0 {
				previousID = latestChainRecord.ID
			}
		}
	} else if latestChainRecord.ID != 0 {
		previousID = latestChainRecord.ID
	}

	recordContent := content
	recordDeleted := false

	recordID, storagePath, err := s.recordSnapshot(fileID, recordDeleted, operation, recordPath, recordContent, fileMode, absSourcePath, absTargetPath, previousID)
	if err != nil {
		return err
	}
	if err := s.enforceRetention(fileID, recordID); err != nil {
		_ = s.repo.DeleteByIDs([]uint{recordID})
		_ = os.Remove(s.absStoragePath(storagePath))
		return err
	}
	return nil
}

func (s *FileHistoryService) recordSnapshot(fileID string, deleted bool, operation string, absPath string, content []byte, fileMode os.FileMode, sourcePath string, targetPath string, previousID uint) (uint, string, error) {
	if fileMode == 0 {
		fileMode = 0640
	}
	now := time.Now()
	pathHash := sha256Hex(absPath)
	contentHash := sha256HexBytes(content)
	fileName := filepath.Base(absPath)
	extension := filepath.Ext(absPath)
	storagePath := s.buildStoragePath(absPath, now)

	if err := os.MkdirAll(path.Dir(s.absStoragePath(storagePath)), os.ModePerm); err != nil {
		return 0, "", err
	}
	if err := os.WriteFile(s.absStoragePath(storagePath), content, 0640); err != nil {
		return 0, "", err
	}

	record := &model.FileHistory{
		FileID:      fileID,
		Path:        absPath,
		PathHash:    pathHash,
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
		FileName:    fileName,
		Extension:   extension,
		FileMode:    fmt.Sprintf("%04o", fileMode.Perm()),
		Operation:   operation,
		Deleted:     deleted,
		ContentSize: int64(len(content)),
		ContentSHA:  contentHash,
		StoragePath: storagePath,
		PreviousID:  previousID,
	}

	if err := s.repo.Create(record); err != nil {
		_ = os.Remove(s.absStoragePath(storagePath))
		return 0, "", err
	}
	return record.ID, storagePath, nil
}

func normalizeAbsPath(filePath string) string {
	filePath = strings.TrimSpace(filePath)
	if filePath == "" {
		return ""
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return filePath
	}
	return absPath
}

func (s *FileHistoryService) Search(req request.FileHistorySearchReq) (int64, []response.FileHistoryInfo, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	size := req.PageSize
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size
	opts := []repo.DBOption{}
	opts = append(opts, s.repo.WithNotOperation("init"))
	switch strings.ToLower(strings.TrimSpace(req.Scope)) {
	case "current":
		if strings.TrimSpace(req.Path) == "" {
			return 0, nil, buserr.New("ErrInvalidParams")
		}
		absPath, err := filepath.Abs(req.Path)
		if err != nil {
			absPath = req.Path
		}
		chain, err := s.getLatestActiveRelatedByPath(absPath)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				chain, err = s.getLatestRelatedByPath(absPath)
				if err != nil {
					if errors.Is(err, gorm.ErrRecordNotFound) {
						return 0, []response.FileHistoryInfo{}, nil
					}
					return 0, nil, err
				}
			} else {
				return 0, nil, err
			}
		}
		opts = append(opts, s.repo.WithByFileID(chain.FileID))
	case "all":
		if strings.TrimSpace(req.Path) != "" {
			opts = append(opts, s.repo.WithByRelatedPath(normalizeAbsPath(req.Path)))
		}
	default:
		return 0, nil, buserr.New("ErrInvalidParams")
	}
	if strings.TrimSpace(req.Operation) != "" {
		opts = append(opts, s.repo.WithByOperation(strings.TrimSpace(req.Operation)))
	}

	total, items, err := s.repo.Page(size, offset, opts...)
	if err != nil {
		return 0, nil, err
	}
	res := make([]response.FileHistoryInfo, 0, len(items))
	for _, item := range items {
		currentPath := item.Path
		if resolved, err := s.getCurrentPathByFileID(item.FileID); err == nil && strings.TrimSpace(resolved) != "" {
			currentPath = resolved
		}
		res = append(res, response.FileHistoryInfo{
			ID:          item.ID,
			FileID:      item.FileID,
			Path:        item.Path,
			CurrentPath: currentPath,
			PreviousID:  item.PreviousID,
			SourcePath:  item.SourcePath,
			TargetPath:  item.TargetPath,
			FileName:    item.FileName,
			Extension:   item.Extension,
			FileMode:    item.FileMode,
			Operation:   item.Operation,
			Deleted:     item.Deleted,
			ContentSize: item.ContentSize,
			ContentSHA:  item.ContentSHA,
			StoragePath: item.StoragePath,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}
	return total, res, nil
}

func (s *FileHistoryService) GetContent(req request.FileHistoryContentReq) (response.FileHistoryInfo, error) {
	record, err := s.repo.Get(s.repo.WithByID(req.ID))
	if err != nil {
		return response.FileHistoryInfo{}, err
	}
	content, err := os.ReadFile(s.absStoragePath(record.StoragePath))
	if err != nil {
		return response.FileHistoryInfo{}, err
	}
	currentContent := s.getCurrentContent(record)
	return response.FileHistoryInfo{
		ID:     record.ID,
		FileID: record.FileID,
		Path:   record.Path,
		CurrentPath: func() string {
			if currentPath, err := s.getCurrentPathByFileID(record.FileID); err == nil && strings.TrimSpace(currentPath) != "" {
				return currentPath
			}
			return record.Path
		}(),
		PreviousID:     record.PreviousID,
		SourcePath:     record.SourcePath,
		TargetPath:     record.TargetPath,
		FileName:       record.FileName,
		Extension:      record.Extension,
		FileMode:       record.FileMode,
		Operation:      record.Operation,
		Deleted:        record.Deleted,
		ContentSize:    record.ContentSize,
		ContentSHA:     record.ContentSHA,
		StoragePath:    record.StoragePath,
		Content:        string(content),
		CurrentContent: currentContent,
		CreatedAt:      record.CreatedAt,
		UpdatedAt:      record.UpdatedAt,
	}, nil
}

func (s *FileHistoryService) Restore(req request.FileHistoryRestoreReq) (response.FileInfo, error) {
	record, err := s.repo.Get(s.repo.WithByID(req.ID))
	if err != nil {
		return response.FileInfo{}, err
	}
	if !s.isVersionOperation(record.Operation) {
		return response.FileInfo{}, buserr.New("ErrInvalidParams")
	}
	content, err := os.ReadFile(s.absStoragePath(record.StoragePath))
	if err != nil {
		return response.FileInfo{}, err
	}

	currentPath := record.Path
	if chainPath, err := s.getCurrentPathByFileID(record.FileID); err == nil && strings.TrimSpace(chainPath) != "" {
		currentPath = chainPath
	}

	lock := s.getPathLock(currentPath)
	lock.Lock()
	defer lock.Unlock()

	var rollbackContent []byte
	var rollbackMode os.FileMode
	var existedBefore bool
	currentInfo, currentErr := files.NewFileInfo(files.FileOption{
		Path:   currentPath,
		Expand: false,
	})
	if currentErr == nil {
		existedBefore = true
		rollbackContent, err = os.ReadFile(currentPath)
		if err != nil {
			return response.FileInfo{}, err
		}
		rollbackMode = currentInfo.FileMode
	} else {
		rollbackMode = parseFileMode(record.FileMode)
		if rollbackMode == 0 {
			rollbackMode = 0640
		}
	}

	targetMode := rollbackMode
	if targetMode == 0 {
		targetMode = parseFileMode(record.FileMode)
	}
	if targetMode == 0 {
		targetMode = 0640
	}

	fo := files.NewFileOp()
	if err := fo.WriteFile(currentPath, strings.NewReader(string(content)), targetMode); err != nil {
		_ = s.rollbackRestore(currentPath, rollbackContent, rollbackMode, existedBefore)
		return response.FileInfo{}, err
	}
	if err := historyService.RecordOperation(fileHistoryOpRestore, currentPath, rollbackContent, rollbackMode, "", ""); err != nil {
		global.LOG.Warnf("record file restore history failed for %s: %v", currentPath, err)
	}

	info, err := files.NewFileInfo(files.FileOption{
		Path:   currentPath,
		Expand: true,
	})
	if err != nil {
		return response.FileInfo{}, err
	}
	return response.FileInfo{FileInfo: *info}, nil
}

func (s *FileHistoryService) Delete(req request.FileHistoryDeleteReq) error {
	if len(req.IDs) == 0 {
		return nil
	}
	for _, id := range req.IDs {
		record, err := s.repo.Get(s.repo.WithByID(id))
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return err
		}
		if err := s.repo.DeleteByIDs([]uint{id}); err != nil {
			return err
		}
		_ = os.Remove(s.absStoragePath(record.StoragePath))
	}
	return nil
}

func (s *FileHistoryService) DeleteRelatedHistory(filePath string) error {
	absPath := normalizeAbsPath(filePath)
	if strings.TrimSpace(absPath) == "" {
		return nil
	}

	fileID, latestRecord, err := s.resolveFileChain(absPath)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(fileID) == "" {
		fileID = latestRecord.FileID
	}
	if strings.TrimSpace(fileID) == "" {
		return nil
	}

	var records []model.FileHistory
	if err := global.DB.Model(&model.FileHistory{}).Where("file_id = ?", fileID).Find(&records).Error; err != nil {
		return err
	}
	if len(records) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(records))
	for _, record := range records {
		ids = append(ids, record.ID)
		_ = os.Remove(s.absStoragePath(record.StoragePath))
	}
	return s.repo.DeleteByIDs(ids)
}

func (s *FileHistoryService) rollbackRestore(filePath string, rollbackContent []byte, rollbackMode os.FileMode, existedBefore bool) error {
	fo := files.NewFileOp()
	if !existedBefore {
		return os.Remove(filePath)
	}
	if rollbackMode == 0 {
		rollbackMode = 0640
	}
	return fo.WriteFile(filePath, strings.NewReader(string(rollbackContent)), rollbackMode)
}

func (s *FileHistoryService) enforceRetention(fileID string, keepID uint) error {
	info, err := s.GetSettingInfo()
	if err != nil {
		return err
	}
	if info.Enable != constant.StatusEnable {
		return nil
	}

	maxPerPath := info.MaxPerPath
	if maxPerPath <= 0 {
		maxPerPath = 20
	}
	quotaBytes := int64(info.DiskQuotaMB) * 1024 * 1024

	if maxPerPath > 0 {
		total, records, err := s.listFileOldest(fileID)
		if err != nil {
			return err
		}
		if int(total) > maxPerPath {
			removeCount := int(total) - maxPerPath
			for i := 0; i < removeCount && i < len(records); i++ {
				if err := s.repo.DeleteByIDs([]uint{records[i].ID}); err != nil {
					return err
				}
				_ = os.Remove(s.absStoragePath(records[i].StoragePath))
			}
		}
	}

	if quotaBytes <= 0 {
		return nil
	}

	totalSize, err := s.totalSize()
	if err != nil {
		return err
	}
	if totalSize <= quotaBytes {
		return nil
	}
	return buserr.New("ErrHistoryQuotaExceeded")
}

func (s *FileHistoryService) getPathLock(absPath string) *sync.Mutex {
	actual, _ := fileHistoryPathLocks.LoadOrStore(absPath, &sync.Mutex{})
	return actual.(*sync.Mutex)
}

func (s *FileHistoryService) listFileOldest(fileID string) (int64, []model.FileHistory, error) {
	var total int64
	var items []model.FileHistory
	db := global.DB.Model(&model.FileHistory{}).Where("file_id = ?", fileID)
	if err := db.Count(&total).Error; err != nil {
		return 0, nil, err
	}
	if err := db.Order("created_at asc").Find(&items).Error; err != nil {
		return 0, nil, err
	}
	return total, items, nil
}

func (s *FileHistoryService) getLatestRelatedByPath(absPath string) (model.FileHistory, error) {
	var item model.FileHistory
	db := global.DB.Model(&model.FileHistory{}).Where("path = ? OR source_path = ? OR target_path = ?", absPath, absPath, absPath).Order("created_at desc")
	err := db.First(&item).Error
	return item, err
}

func (s *FileHistoryService) getLatestActiveRelatedByPath(absPath string) (model.FileHistory, error) {
	var item model.FileHistory
	db := global.DB.Model(&model.FileHistory{}).Where("deleted = ? AND (path = ? OR source_path = ? OR target_path = ?)", false, absPath, absPath, absPath).Order("created_at desc")
	err := db.First(&item).Error
	return item, err
}

func (s *FileHistoryService) getLatestVersionByFileID(fileID string) (model.FileHistory, error) {
	var item model.FileHistory
	db := global.DB.Model(&model.FileHistory{}).Where("file_id = ? AND operation IN ?", fileID, []string{fileHistoryOpSave, fileHistoryOpRestore}).Order("created_at desc")
	err := db.First(&item).Error
	return item, err
}

func (s *FileHistoryService) getCurrentPathByFileID(fileID string) (string, error) {
	record, err := s.getLatestRelatedByFileID(fileID)
	if err != nil {
		return "", err
	}
	return record.Path, nil
}

func (s *FileHistoryService) getCurrentContent(record model.FileHistory) string {
	currentPath := ""
	if resolved, err := s.getCurrentPathByFileID(record.FileID); err == nil && strings.TrimSpace(resolved) != "" {
		currentPath = resolved
	}
	candidates := []string{}
	if strings.TrimSpace(currentPath) != "" {
		candidates = append(candidates, currentPath)
	}
	if strings.TrimSpace(record.Path) != "" && record.Path != currentPath {
		candidates = append(candidates, record.Path)
	}
	for _, candidate := range candidates {
		content, err := os.ReadFile(candidate)
		if err == nil {
			return string(content)
		}
	}
	content, err := os.ReadFile(s.absStoragePath(record.StoragePath))
	if err == nil {
		return string(content)
	}
	return ""
}

func (s *FileHistoryService) isEditableHistorySnapshot(fileMode os.FileMode, content []byte) bool {
	if fileMode.IsDir() || files.IsBlockDevice(fileMode) {
		return false
	}
	if len(content) == 0 {
		return true
	}
	return !files.DetectBinary(content)
}

func (s *FileHistoryService) getLatestRelatedByFileID(fileID string) (model.FileHistory, error) {
	var item model.FileHistory
	db := global.DB.Model(&model.FileHistory{}).Where("file_id = ?", fileID).Order("created_at desc")
	err := db.First(&item).Error
	return item, err
}

func (s *FileHistoryService) resolveFileChain(paths ...string) (string, model.FileHistory, error) {
	for _, itemPath := range paths {
		itemPath = strings.TrimSpace(itemPath)
		if itemPath == "" {
			continue
		}
		record, err := s.getLatestActiveRelatedByPath(itemPath)
		if err == nil {
			return record.FileID, record, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return "", model.FileHistory{}, err
		}
	}
	return "", model.FileHistory{}, gorm.ErrRecordNotFound
}

func (s *FileHistoryService) isVersionOperation(operation string) bool {
	return operation == fileHistoryOpSave || operation == fileHistoryOpRestore
}

func (s *FileHistoryService) totalSize() (int64, error) {
	var total int64
	err := global.DB.Model(&model.FileHistory{}).Select("coalesce(sum(content_size),0)").Scan(&total).Error
	return total, err
}

func (s *FileHistoryService) listOldest() ([]model.FileHistory, error) {
	var items []model.FileHistory
	err := global.DB.Model(&model.FileHistory{}).Order("created_at asc").Find(&items).Error
	return items, err
}

func (s *FileHistoryService) buildStoragePath(absPath string, now time.Time) string {
	fileName := filepath.Base(absPath)
	extension := filepath.Ext(fileName)
	baseName := strings.TrimSuffix(fileName, extension)
	if baseName == "" {
		baseName = fileName
		extension = ""
	}
	baseName = sanitizeHistoryFileName(baseName)
	if baseName == "" {
		baseName = "file"
	}
	dayDir := now.Format("2006-01-02")
	shortCode := strings.ReplaceAll(common.GetUuid(), "-", "")
	if len(shortCode) > 8 {
		shortCode = shortCode[:8]
	}
	return path.Join(fileHistoryRootDirName, dayDir, fmt.Sprintf("%s__%s%s", baseName, shortCode, extension))
}

func (s *FileHistoryService) absStoragePath(rel string) string {
	return path.Join(global.Dir.LocalBackupDir, rel)
}

func sanitizeHistoryFileName(name string) string {
	name = strings.TrimSpace(name)
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(name)
}

func parseFileMode(mode string) os.FileMode {
	if strings.TrimSpace(mode) == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(strings.TrimSpace(mode), 8, 32)
	if err != nil {
		return 0
	}
	return os.FileMode(parsed)
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func sha256HexBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
