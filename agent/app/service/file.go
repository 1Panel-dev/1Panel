package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/convert"
	"github.com/1Panel-dev/1Panel/agent/utils/ini_conf"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/jinzhu/copier"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"

	"github.com/1Panel-dev/1Panel/agent/app/repo"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"golang.org/x/net/html/charset"
	"golang.org/x/sys/unix"
	"golang.org/x/text/transform"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/pkg/errors"
)

type FileService struct {
}

type IFileService interface {
	GetFileList(op request.FileOption) (response.FileInfo, error)
	SearchUploadWithPage(req request.SearchUploadWithPage) (int64, interface{}, error)
	GetFileTree(op request.FileOption) ([]response.FileTree, error)
	Create(op request.FileCreate) error
	Delete(op request.FileDelete) error
	BatchDelete(op request.FileBatchDelete) error
	Compress(c request.FileCompress) error
	StopCompress(taskID string) error
	DeCompress(c request.FileDeCompress) error
	GetContent(op request.FileContentReq) (response.FileInfo, error)
	GetPreviewContent(op request.FileContentReq) (response.FileInfo, error)
	SaveContent(edit request.FileEdit) error
	FileDownload(d request.FileDownload) (string, error)
	DirSize(req request.DirSizeReq) (response.DirSizeRes, error)
	DepthDirSize(req request.DirSizeReq) ([]response.DepthDirSizeRes, error)
	ChangeName(req request.FileRename) error
	Wget(w request.FileWget) (string, error)
	MvFile(m request.FileMove) error
	ChangeOwner(req request.FileRoleUpdate) error
	ChangeMode(op request.FileCreate) error
	BatchChangeModeAndOwner(op request.FileRoleReq) error
	ReadLogByLine(req request.FileReadByLineReq) (*response.FileLineContent, error)

	GetPathByType(pathType string) string
	BatchCheckFiles(req request.FilePathsCheck) []response.ExistFileInfo
	GetHostMount() []dto.DiskInfo
	GetUsersAndGroups() (*response.UserGroupResponse, error)
	Convert(req request.FileConvertRequest)
	ConvertLog(req dto.PageInfo) (int64, []response.FileConvertLog, error)
	BatchGetRemarks(req request.FileRemarkBatch) map[string]string
	SetRemark(req request.FileRemarkUpdate) error
	AISearch(req request.FileAISearch) (*response.FileAISearchResult, error)
}

const (
	fileRemarkXattr         = "user.1panel.remark"
	fileRemarkEncodedMaxLen = 256
)

func NewIFileService() IFileService {
	return &FileService{}
}

func (f *FileService) GetFileList(op request.FileOption) (response.FileInfo, error) {
	var fileInfo response.FileInfo
	data, err := os.Stat(op.Path)
	if err != nil && os.IsNotExist(err) {
		return fileInfo, nil
	}
	if !data.IsDir() {
		op.FileOption.Path = filepath.Dir(op.FileOption.Path)
	}
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return fileInfo, err
	}
	shareMap, err := NewIFileShareService().SharePathCodeMap()
	if err != nil {
		return fileInfo, err
	}
	applyFileShares(info, shareMap)
	fileInfo.FileInfo = *info
	return fileInfo, nil
}

func applyFileShares(info *files.FileInfo, shareMap map[string]string) {
	if info == nil {
		return
	}
	if code, ok := shareMap[info.Path]; ok {
		info.ShareCode = code
	} else {
		info.ShareCode = ""
	}
	for _, item := range info.Items {
		applyFileShares(item, shareMap)
	}
}

func (f *FileService) SearchUploadWithPage(req request.SearchUploadWithPage) (int64, interface{}, error) {
	var (
		files    []response.UploadInfo
		backData []response.UploadInfo
	)
	fileList, err := os.ReadDir(req.Path)
	if err != nil {
		return 0, files, nil
	}
	for _, item := range fileList {
		if item.IsDir() {
			continue
		}
		fileItem, err := item.Info()
		if err != nil {
			continue
		}
		files = append(files, response.UploadInfo{
			CreatedAt: fileItem.ModTime().Format(constant.DateTimeLayout),
			Size:      int(fileItem.Size()),
			Name:      item.Name(),
		})
	}

	total, start, end := len(files), (req.Page-1)*req.PageSize, req.Page*req.PageSize
	if start > total {
		backData = make([]response.UploadInfo, 0)
	} else {
		if end >= total {
			end = total
		}
		backData = files[start:end]
	}
	return int64(total), backData, nil
}

func (f *FileService) GetFileTree(op request.FileOption) ([]response.FileTree, error) {
	var treeArray []response.FileTree
	if _, err := os.Stat(op.Path); err != nil && os.IsNotExist(err) {
		return treeArray, nil
	}
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return nil, err
	}
	node := response.FileTree{
		ID:        common.GetUuid(),
		Name:      info.Name,
		Path:      info.Path,
		IsDir:     info.IsDir,
		Extension: info.Extension,
	}
	err = f.buildFileTree(&node, info.Items, op, 2)
	if err != nil {
		return nil, err
	}
	return append(treeArray, node), nil
}

func shouldFilterPath(path string) bool {
	return files.ShouldFilterSensitivePath(path)
}

func (f *FileService) buildFileTree(node *response.FileTree, items []*files.FileInfo, op request.FileOption, level int) error {
	for _, v := range items {
		if shouldFilterPath(v.Path) {
			global.LOG.Infof("File Tree: Skipping %s due to filter\n", v.Path)
			continue
		}
		childNode := response.FileTree{
			ID:        common.GetUuid(),
			Name:      v.Name,
			Path:      v.Path,
			IsDir:     v.IsDir,
			Extension: v.Extension,
		}
		if level > 1 && v.IsDir {
			if err := f.buildChildNode(&childNode, v, op, level); err != nil {
				return err
			}
		}

		node.Children = append(node.Children, childNode)
	}
	return nil
}

func (f *FileService) buildChildNode(childNode *response.FileTree, fileInfo *files.FileInfo, op request.FileOption, level int) error {
	op.Path = fileInfo.Path
	subInfo, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		if os.IsPermission(err) || errors.Is(err, unix.EACCES) {
			global.LOG.Infof("File Tree: Skipping %s due to permission denied\n", fileInfo.Path)
			return nil
		}
		global.LOG.Errorf("File Tree: Skipping %s due to error: %s\n", fileInfo.Path, err.Error())
		return nil
	}

	return f.buildFileTree(childNode, subInfo.Items, op, level-1)
}

func hasInvalidFileName(fullPath string) bool {
	return files.IsInvalidChar(filepath.Base(fullPath))
}

func (f *FileService) Create(op request.FileCreate) error {
	if hasInvalidFileName(op.Path) {
		return buserr.New("ErrInvalidChar")
	}
	fo := files.NewFileOp()
	if fo.Stat(op.Path) {
		return buserr.New("ErrFileIsExist")
	}
	mode := op.Mode
	if mode == 0 {
		fileInfo, err := os.Stat(filepath.Dir(op.Path))
		if err == nil {
			mode = int64(fileInfo.Mode().Perm())
		} else {
			mode = constant.DirPerm
		}
	}
	if op.IsDir {
		if err := fo.CreateDirWithMode(op.Path, fs.FileMode(mode)); err != nil {
			return err
		}
		handleDefaultOwn(op.Path)
		return nil
	}
	if op.IsLink {
		if !fo.Stat(op.LinkPath) {
			return buserr.New("ErrLinkPathNotFound")
		}
		if err := fo.LinkFile(op.LinkPath, op.Path, op.IsSymlink); err != nil {
			return err
		}
		handleDefaultOwn(op.Path)
		return nil
	}
	if err := fo.CreateFileWithMode(op.Path, fs.FileMode(mode)); err != nil {
		return err
	}
	handleDefaultOwn(op.Path)
	return nil
}

func (f *FileService) Delete(op request.FileDelete) error {
	if op.IsDir {
		excludeDir := global.Dir.DataDir
		if path.Base(op.Path) == ".1panel_clash" || op.Path == excludeDir {
			return buserr.New("ErrPathNotDelete")
		}
	}
	fo := files.NewFileOp()
	recycleBinStatus, _ := settingRepo.Get(settingRepo.WithByKey("FileRecycleBin"))
	if recycleBinStatus.Value == "Disable" {
		op.ForceDelete = true
	}
	var historyTargets []string
	if op.ForceDelete {
		var err error
		historyTargets, err = f.collectPermanentDeleteTargets(op.Path, op.IsDir)
		if err != nil {
			return err
		}
	}
	if op.ForceDelete {
		var err error
		if op.IsDir {
			err = fo.DeleteDir(op.Path)
		} else {
			err = fo.DeleteFile(op.Path)
		}
		if err != nil {
			return err
		}
		if err := cleanupTrashInfoByEntryPath(op.Path); err != nil {
			global.LOG.Warnf("cleanup trashinfo failed for %s: %v", op.Path, err)
		}
		f.cleanupPermanentDeleteHistory(historyTargets)
		return nil
	}
	info, _ := fo.Fs.Stat(op.Path)
	if info == nil || files.IsSymlink(info.Mode()) {
		if err := os.Remove(op.Path); err != nil {
			return err
		}
		f.cleanupPermanentDeleteHistory([]string{op.Path})
		return nil
	}

	if err := NewIRecycleBinService().Create(request.RecycleBinCreate{SourcePath: op.Path}); err != nil {
		return err
	}
	return favoriteRepo.Delete(favoriteRepo.WithByPath(op.Path))
}

func (f *FileService) BatchDelete(op request.FileBatchDelete) error {
	fo := files.NewFileOp()
	if op.IsDir {
		for _, file := range op.Paths {
			targets, err := f.collectPermanentDeleteTargets(file, true)
			if err != nil {
				return err
			}
			if err := fo.DeleteDir(file); err != nil {
				return err
			}
			if err := cleanupTrashInfoByEntryPath(file); err != nil {
				global.LOG.Warnf("cleanup trashinfo failed for %s: %v", file, err)
			}
			f.cleanupPermanentDeleteHistory(targets)
		}
	} else {
		for _, file := range op.Paths {
			if err := fo.DeleteFile(file); err != nil {
				return err
			}
			if err := cleanupTrashInfoByEntryPath(file); err != nil {
				global.LOG.Warnf("cleanup trashinfo failed for %s: %v", file, err)
			}
			f.cleanupPermanentDeleteHistory([]string{file})
		}
	}
	return nil
}

func (f *FileService) ChangeMode(op request.FileCreate) error {
	fo := files.NewFileOp()
	if err := fo.ChmodR(op.Path, op.Mode, op.Sub); err != nil {
		return err
	}
	return nil
}

func (f *FileService) BatchChangeModeAndOwner(op request.FileRoleReq) error {
	fo := files.NewFileOp()
	for _, p := range op.Paths {
		if !fo.Stat(p) {
			return buserr.New("ErrPathNotFound")
		}
	}
	if err := fo.ChownRPaths(op.Paths, op.User, op.Group, op.Sub); err != nil {
		return err
	}
	if err := fo.ChmodRPaths(op.Paths, op.Mode, op.Sub); err != nil {
		return err
	}
	return nil
}

func (f *FileService) collectPermanentDeleteTargets(targetPath string, isDir bool) ([]string, error) {
	if !isDir {
		return []string{targetPath}, nil
	}

	var targets []string
	if err := filepath.WalkDir(targetPath, func(currentPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d == nil || d.IsDir() {
			return nil
		}
		targets = append(targets, currentPath)
		return nil
	}); err != nil {
		return nil, err
	}
	return targets, nil
}

func (f *FileService) cleanupPermanentDeleteHistory(targets []string) {
	if len(targets) == 0 {
		return
	}
	for _, target := range targets {
		if err := historyService.DeleteRelatedHistory(target); err != nil {
			global.LOG.Warnf("cleanup file history failed for %s: %v", target, err)
		}
	}
}

func (f *FileService) ChangeOwner(req request.FileRoleUpdate) error {
	fo := files.NewFileOp()
	if err := fo.ChownR(req.Path, req.User, req.Group, req.Sub); err != nil {
		return err
	}
	return nil
}

func (f *FileService) Compress(c request.FileCompress) error {
	fo := files.NewFileOp()
	if !c.Replace && fo.Stat(filepath.Join(c.Dst, c.Name)) {
		return buserr.New("ErrFileIsExist")
	}
	if err := preflightCompressTool(files.CompressType(c.Type)); err != nil {
		return err
	}
	taskItem, err := task.NewTask(c.Name, task.TaskExec, task.TaskScopeTask, c.TaskID, 1)
	if err != nil {
		return err
	}
	go func() {
		taskItem.AddSubTask(c.Name, func(t *task.Task) error {
			t.LogStart(c.Name)
			compressType := files.CompressType(c.Type)
			dstFile := filepath.Join(c.Dst, c.Name)
			success := false
			defer func() {
				if !success {
					_ = os.Remove(dstFile)
				}
			}()
			if err := fo.Compress(t.TaskCtx, c.Files, c.Dst, c.Name, compressType, c.Secret, nil); err != nil {
				return err
			}
			info, err := os.Stat(dstFile)
			if err != nil {
				return err
			}
			if info.Size() == 0 {
				return fmt.Errorf("compressed file not generated: %s", dstFile)
			}
			success = true
			return nil
		}, nil)
		_ = taskItem.Execute()
	}()
	return nil
}

func preflightCompressTool(compressType files.CompressType) error {
	switch compressType {
	case files.TarGz, files.Rar, files.X7z:
		_, err := files.NewShellArchiver(compressType)
		return err
	default:
		return nil
	}
}

func (f *FileService) StopCompress(taskID string) error {
	if cancel, ok := global.TaskCtxMap[taskID]; ok {
		cancel()
		return nil
	}
	return buserr.New("TaskNotFound")
}

func (f *FileService) DeCompress(c request.FileDeCompress) error {
	fo := files.NewFileOp()
	if c.Type == "tar" && len(c.Secret) != 0 {
		c.Type = "tar.gz"
	}
	return fo.Decompress(c.Path, c.Dst, files.CompressType(c.Type), c.Secret)
}

func (f *FileService) GetContent(op request.FileContentReq) (response.FileInfo, error) {
	info, err := files.NewFileInfo(files.FileOption{
		Path:     op.Path,
		Expand:   true,
		IsDetail: op.IsDetail,
	})
	if err != nil {
		return response.FileInfo{}, err
	}

	content := []byte(info.Content)
	if len(content) > 1024 {
		content = content[:1024]
	}
	if !utf8.Valid(content) {
		_, decodeName, _ := charset.DetermineEncoding(content, "")
		decoder := files.GetDecoderByName(decodeName)
		if decoder != nil {
			reader := strings.NewReader(info.Content)
			var dec *encoding.Decoder
			if decodeName == "windows-1252" {
				dec = simplifiedchinese.GBK.NewDecoder()
			} else {
				dec = decoder.NewDecoder()
			}
			decodedReader := transform.NewReader(reader, dec)
			contents, err := io.ReadAll(decodedReader)
			if err != nil {
				return response.FileInfo{}, err
			}
			info.Content = string(contents)
		}
	}
	return response.FileInfo{FileInfo: *info}, nil
}

func (f *FileService) GetPreviewContent(op request.FileContentReq) (response.FileInfo, error) {
	info, err := files.NewFileInfo(files.FileOption{
		Path:     op.Path,
		Expand:   false,
		IsDetail: op.IsDetail,
	})
	if err != nil {
		return response.FileInfo{}, err
	}

	if files.IsBlockDevice(info.FileMode) {
		return response.FileInfo{FileInfo: *info}, nil
	}

	file, err := os.Open(op.Path)
	if err != nil {
		return response.FileInfo{}, err
	}
	defer file.Close()

	headBuf := make([]byte, 1024)
	n, err := file.Read(headBuf)
	if err != nil && err != io.EOF {
		return response.FileInfo{}, err
	}
	headBuf = headBuf[:n]

	if len(headBuf) > 0 && files.DetectBinary(headBuf) {
		return response.FileInfo{FileInfo: *info}, nil
	}

	const maxSize = 10 * 1024 * 1024
	if info.Size <= maxSize {
		if _, err := file.Seek(0, 0); err != nil {
			return response.FileInfo{}, err
		}
		content, err := io.ReadAll(file)
		if err != nil {
			return response.FileInfo{}, err
		}
		info.Content = string(content)
	} else {
		lines, err := files.TailFromEnd(op.Path, 300)
		if err != nil {
			return response.FileInfo{}, err
		}
		info.Content = strings.Join(lines, "\n")
	}

	content := []byte(info.Content)
	if len(content) > 1024 {
		content = content[:1024]
	}
	if !utf8.Valid(content) {
		_, decodeName, _ := charset.DetermineEncoding(content, "")
		decoder := files.GetDecoderByName(decodeName)
		if decoder != nil {
			reader := strings.NewReader(info.Content)
			var dec *encoding.Decoder
			if decodeName == "windows-1252" {
				dec = simplifiedchinese.GBK.NewDecoder()
			} else {
				dec = decoder.NewDecoder()
			}
			decodedReader := transform.NewReader(reader, dec)
			contents, err := io.ReadAll(decodedReader)
			if err != nil {
				return response.FileInfo{}, err
			}
			info.Content = string(contents)
		}
	}

	return response.FileInfo{FileInfo: *info}, nil
}

func (f *FileService) SaveContent(edit request.FileEdit) error {
	info, err := files.NewFileInfo(files.FileOption{
		Path:   edit.Path,
		Expand: false,
	})
	if err != nil {
		return err
	}

	fo := files.NewFileOp()
	oldContent, _ := os.ReadFile(edit.Path)
	if bytes.Equal(oldContent, []byte(edit.Content)) {
		return nil
	}

	if err := fo.WriteFile(edit.Path, strings.NewReader(edit.Content), info.FileMode); err != nil {
		return err
	}
	if err := historyService.RecordSave(edit.Path, oldContent, info.FileMode); err != nil {
		global.LOG.Warnf("record file save history failed for %s: %v", edit.Path, err)
	}
	return nil
}

func (f *FileService) ChangeName(req request.FileRename) error {
	if hasInvalidFileName(req.NewName) {
		return buserr.New("ErrInvalidChar")
	}
	fo := files.NewFileOp()
	info, _ := files.NewFileInfo(files.FileOption{Path: req.OldName, Expand: false})
	content, _ := os.ReadFile(req.OldName)
	if err := fo.Rename(req.OldName, req.NewName); err != nil {
		return err
	}
	if info != nil && !info.IsDir {
		if histErr := historyService.RecordOperation(fileHistoryOpRename, req.OldName, content, info.FileMode, req.OldName, req.NewName); histErr != nil {
			global.LOG.Warnf("record file rename history failed for %s: %v", req.OldName, histErr)
		}
	}
	return nil
}

func (f *FileService) Wget(w request.FileWget) (string, error) {
	fo := files.NewFileOp()
	key := "file-wget-" + common.GetUuid()
	return key, fo.DownloadFileWithProcess(w.Url, filepath.Join(w.Path, w.Name), key, w.IgnoreCertificate)
}

func (f *FileService) MvFile(m request.FileMove) error {
	fo := files.NewFileOp()
	if !fo.Stat(m.NewPath) {
		return buserr.New("ErrPathNotFound")
	}
	for _, oldPath := range m.OldPaths {
		if !fo.Stat(oldPath) {
			return buserr.WithName("ErrFileNotFound", oldPath)
		}
		if oldPath == m.NewPath || strings.Contains(m.NewPath, filepath.Clean(oldPath)+"/") {
			return buserr.New("ErrMovePathFailed")
		}
	}
	type moveSnapshot struct {
		path    string
		content []byte
		mode    os.FileMode
		isDir   bool
	}
	snapshots := make([]moveSnapshot, 0, len(m.OldPaths))
	for _, oldPath := range m.OldPaths {
		content, _ := os.ReadFile(oldPath)
		mode := os.FileMode(0640)
		isDir := false
		if info, err := files.NewFileInfo(files.FileOption{Path: oldPath, Expand: false}); err == nil {
			mode = info.FileMode
			isDir = info.IsDir
		}
		snapshots = append(snapshots, moveSnapshot{path: oldPath, content: content, mode: mode, isDir: isDir})
	}
	var errs []error
	if m.Type == "cut" {
		if len(m.CoverPaths) > 0 {
			for _, src := range m.CoverPaths {
				if err := fo.CopyAndReName(src, m.NewPath, "", true); err != nil {
					errs = append(errs, err)
					global.LOG.Errorf("cut copy file [%s] to [%s] failed, err: %s", src, m.NewPath, err.Error())
				}
			}
		}
		if err := fo.Cut(m.OldPaths, m.NewPath, m.Name, m.Cover); err != nil {
			return err
		}
		for _, snapshot := range snapshots {
			if !snapshot.isDir {
				targetPath := buildHistoryMoveTargetPath(m.NewPath, m.Name, snapshot.path, len(m.OldPaths))
				if histErr := historyService.RecordOperation(fileHistoryOpMove, snapshot.path, snapshot.content, snapshot.mode, snapshot.path, targetPath); histErr != nil {
					global.LOG.Warnf("record file move history failed for %s: %v", snapshot.path, histErr)
				}
			}
		}
		return nil
	}
	if m.Type == "copy" {
		for _, src := range m.OldPaths {
			if err := fo.CopyAndReName(src, m.NewPath, m.Name, m.Cover); err != nil {
				errs = append(errs, err)
				global.LOG.Errorf("copy file [%s] to [%s] failed, err: %s", src, m.NewPath, err.Error())
			}
		}
		if len(m.CoverPaths) > 0 {
			for _, src := range m.CoverPaths {
				if err := fo.CopyAndReName(src, m.NewPath, "", true); err != nil {
					errs = append(errs, err)
					global.LOG.Errorf("copy file [%s] to [%s] failed, err: %s", src, m.NewPath, err.Error())
				}
			}
		}
	}

	var errString string
	for _, err := range errs {
		errString += err.Error() + "\n"
	}
	if errString != "" {
		return errors.New(errString)
	}
	return nil
}

func buildHistoryMoveTargetPath(dst, name, sourcePath string, sourceCount int) string {
	if strings.TrimSpace(dst) == "" {
		return sourcePath
	}
	if strings.TrimSpace(name) != "" && sourceCount == 1 {
		return filepath.Join(dst, name)
	}
	return filepath.Join(dst, filepath.Base(sourcePath))
}

func (f *FileService) FileDownload(d request.FileDownload) (string, error) {
	filePath := d.Paths[0]
	if d.Compress {
		tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().UnixNano()))
		if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
			return "", err
		}
		fo := files.NewFileOp()
		if err := fo.Compress(context.Background(), d.Paths, tempPath, d.Name, files.CompressType(d.Type), "", nil); err != nil {
			return "", err
		}
		filePath = filepath.Join(tempPath, d.Name)
	}
	return filePath, nil
}

func (f *FileService) DirSize(req request.DirSizeReq) (response.DirSizeRes, error) {
	var (
		res response.DirSizeRes
	)
	if req.Path == "/proc" {
		return res, nil
	}
	fo := files.NewFileOp()
	size, err := fo.GetDirSize(req.Path)
	if err != nil {
		return res, err
	}
	res.Size = size
	return res, nil
}

func (f *FileService) DepthDirSize(req request.DirSizeReq) ([]response.DepthDirSizeRes, error) {
	var (
		res []response.DepthDirSizeRes
	)
	if req.Path == "/proc" {
		return res, nil
	}
	fo := files.NewFileOp()
	dirSizes, err := fo.GetDepthDirSize(req.Path)
	_ = copier.Copy(&res, &dirSizes)
	if err != nil {
		return res, err
	}
	return res, nil
}

func (f *FileService) ReadLogByLine(req request.FileReadByLineReq) (*response.FileLineContent, error) {
	logFilePath := ""
	taskStatus := ""
	if len(req.Name) != 0 {
		safeName := path.Base(req.Name)
		if safeName != req.Name || strings.Contains(safeName, "..") {
			return nil, buserr.New("ErrInvalidParams")
		}
	}
	switch req.Type {
	case constant.TypeWebsite:
		website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		logFilePath = GetSitePath(website, req.Name)
	case constant.TypePhp:
		php, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		logFilePath = php.GetLogPath()
	case constant.TypeSSL:
		ssl, err := websiteSSLRepo.GetFirst(repo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		logFilePath = ssl.GetLogPath()
	case constant.TypeSystem:
		fileName := ""
		if len(req.Name) == 0 {
			fileName = "1Panel.log"
		} else {
			if strings.HasSuffix(req.Name, time.Now().Format("2006-01-02")) {
				fileName = "1Panel.log"
				if strings.HasPrefix(req.Name, "Core-") {
					fileName = "1Panel-Core.log"
				}
			} else {
				fileName = "1Panel-" + req.Name + ".log"
			}
		}
		logFilePath = path.Join(global.Dir.DataDir, "log", fileName)
		if _, err := os.Stat(logFilePath); err != nil {
			fileGzPath := path.Join(global.Dir.DataDir, "log", fileName+".gz")
			if _, err := os.Stat(fileGzPath); err != nil {
				return nil, buserr.New("ErrHttpReqNotFound")
			}
			if err := handleGunzip(fileGzPath); err != nil {
				return nil, fmt.Errorf("handle ungzip file %s failed, err: %v", fileGzPath, err)
			}
		}
	case constant.TypeTask:
		var opts []repo.DBOption
		if req.TaskID != "" {
			opts = append(opts, taskRepo.WithByID(req.TaskID))
		} else {
			opts = append(opts, repo.WithOrderRuleBy("created_at", "desc"), repo.WithByType(req.TaskType), taskRepo.WithOperate(req.TaskOperate), taskRepo.WithResourceID(req.ResourceID))
		}
		taskModel, err := taskRepo.GetFirst(opts...)
		if err != nil {
			return nil, err
		}
		logFilePath = taskModel.LogFile
		taskStatus = taskModel.Status
	case "mysql-slow-logs":
		logFilePath = path.Join(global.Dir.DataDir, "apps", "mysql", req.Name, "data", "1Panel-slow.log")
	case "mariadb-slow-logs":
		logFilePath = path.Join(global.Dir.DataDir, "apps", "mariadb", req.Name, "db", "data", "1Panel-slow.log")
	case "php-fpm-slow-logs":
		php, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		logFilePath = php.GetSlowLogPath()
	case constant.Supervisord:
		configPath := "/etc/supervisord.conf"
		pathSet, _ := settingRepo.Get(settingRepo.WithByKey(constant.SupervisorConfigPath))
		if pathSet.ID != 0 || pathSet.Value != "" {
			configPath = pathSet.Value
		}
		logFilePath, _ = ini_conf.GetIniValue(configPath, "supervisord", "logfile")
	case constant.Supervisor:
		logFilePath = path.Join(global.Dir.DataDir, "tools", "supervisord", "log", req.Name)
	case "ai-proxy":
		safeName := path.Base(req.Name)
		if safeName != req.Name || strings.Contains(safeName, "..") {
			return nil, buserr.New("ErrInvalidParams")
		}
		logFilePath = path.Join(global.Dir.LogDir, "ai", safeName)
	}

	file, err := os.Open(logFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	var (
		lines       []string
		isEndOfFile bool
		scope       string
		logFileRes  *dto.LogFileRes
	)
	if stat.Size() > files.MaxReadFileSize {
		lines, _ = files.TailFromEnd(logFilePath, req.PageSize)
		isEndOfFile = true
		scope = "tail"
	} else {
		logFileRes, err = files.ReadFileByLine(logFilePath, req.Page, req.PageSize, req.Latest)
		if err != nil {
			return nil, err
		}
		scope = "page"
		lines = logFileRes.Lines
	}

	res := &response.FileLineContent{
		End:        isEndOfFile,
		Path:       logFilePath,
		TaskStatus: taskStatus,
		Lines:      lines,
		Scope:      scope,
	}
	if logFileRes != nil {
		res.TotalLines = logFileRes.TotalLines
		res.Total = logFileRes.TotalPages
		res.End = logFileRes.IsEndOfFile
	}
	return res, nil
}

func (f *FileService) GetPathByType(pathType string) string {
	if pathType == "websiteDir" {
		value, _ := settingRepo.GetValueByKey("WEBSITE_DIR")
		if value == "" {
			return path.Join(global.Dir.BaseDir, "1panel", "www")
		}
		return value
	}
	return ""
}

func (f *FileService) BatchCheckFiles(req request.FilePathsCheck) []response.ExistFileInfo {
	fileList := make([]response.ExistFileInfo, 0, len(req.Paths))
	for _, filePath := range req.Paths {
		if info, err := os.Stat(filePath); err == nil {
			fileList = append(fileList, response.ExistFileInfo{
				Size:    info.Size(),
				Name:    info.Name(),
				Path:    filePath,
				ModTime: info.ModTime(),
				IsDir:   info.IsDir(),
			})
		}
	}
	return fileList
}

func (f *FileService) GetHostMount() []dto.DiskInfo {
	return loadDiskInfo()
}

func (f *FileService) GetUsersAndGroups() (*response.UserGroupResponse, error) {
	groupMap, err := getValidGroups()
	if err != nil {
		return nil, err
	}

	users, groupSet, err := getValidUsers(groupMap)
	if err != nil {
		return nil, err
	}

	var groups []string
	for group := range groupSet {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	return &response.UserGroupResponse{
		Users:  users,
		Groups: groups,
	}, nil
}

func (f *FileService) BatchGetRemarks(req request.FileRemarkBatch) map[string]string {
	remarks := make(map[string]string)
	for _, filePath := range req.Paths {
		remark, err := getFileRemark(filePath)
		if err != nil {
			if isXattrNotSupported(err) {
				return map[string]string{}
			}
			continue
		}
		if remark == "" {
			continue
		}
		remarks[filePath] = remark
	}

	return remarks
}

func (f *FileService) SetRemark(req request.FileRemarkUpdate) error {
	if req.Remark == "" {
		if err := unix.Lremovexattr(req.Path, fileRemarkXattr); err != nil {
			if isXattrNotFound(err) {
				return nil
			}
			if isXattrNotSupported(err) {
				return buserr.WithDetail("ErrInvalidParams", "xattr not supported", err)
			}
			return err
		}
		return nil
	}

	encoded := base64.StdEncoding.EncodeToString([]byte(req.Remark))
	if len(encoded) >= fileRemarkEncodedMaxLen {
		return buserr.WithDetail("ErrInvalidParams", "remark length must be less than 256", nil)
	}
	if err := unix.Lsetxattr(req.Path, fileRemarkXattr, []byte(encoded), 0); err != nil {
		if isXattrNotSupported(err) {
			return buserr.WithDetail("ErrInvalidParams", "xattr not supported", err)
		}
		return err
	}
	return nil
}

func getValidGroups() (map[string]bool, error) {
	groupFile, err := os.Open("/etc/group")
	if err != nil {
		return nil, fmt.Errorf("failed to open /etc/group: %w", err)
	}
	defer groupFile.Close()

	groupMap := make(map[string]bool)
	scanner := bufio.NewScanner(groupFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) < 3 {
			continue
		}
		groupName := parts[0]
		gid, _ := strconv.Atoi(parts[2])
		if groupName == "root" || gid >= 1000 {
			groupMap[groupName] = true
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan /etc/group: %w", err)
	}
	return groupMap, nil
}

func getFileRemark(filePath string) (string, error) {
	size, err := unix.Lgetxattr(filePath, fileRemarkXattr, nil)
	if err != nil {
		if isXattrNotFound(err) {
			return "", nil
		}
		return "", err
	}
	if size == 0 {
		return "", nil
	}
	buf := make([]byte, size)
	n, err := unix.Lgetxattr(filePath, fileRemarkXattr, buf)
	if err != nil {
		return "", err
	}
	decoded, err := base64.StdEncoding.DecodeString(string(buf[:n]))
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func isXattrNotSupported(err error) bool {
	return errors.Is(err, unix.ENOTSUP) || errors.Is(err, unix.EOPNOTSUPP)
}

func isXattrNotFound(err error) bool {
	return errors.Is(err, unix.ENODATA)
}

func getValidUsers(validGroups map[string]bool) ([]response.UserInfo, map[string]struct{}, error) {
	passwdFile, err := os.Open("/etc/passwd")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open /etc/passwd: %w", err)
	}
	defer passwdFile.Close()

	var users []response.UserInfo
	groupSet := make(map[string]struct{})
	scanner := bufio.NewScanner(passwdFile)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), ":")
		if len(parts) < 4 {
			continue
		}

		username := parts[0]
		uid, _ := strconv.Atoi(parts[2])
		gid := parts[3]

		if username != "root" && uid < 1000 {
			continue
		}

		groupName := gid
		if g, err := user.LookupGroupId(gid); err == nil {
			groupName = g.Name
		}

		if !validGroups[groupName] {
			continue
		}

		users = append(users, response.UserInfo{
			Username: username,
			Group:    groupName,
		})
		groupSet[groupName] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed to scan /etc/passwd: %w", err)
	}
	return users, groupSet, nil
}

func (f *FileService) Convert(req request.FileConvertRequest) {
	convertTask, err := task.NewTaskWithOps(i18n.GetMsgByKey("FileConvert"), task.TaskExec, task.TaskScopeFileConvert, req.TaskID, 1)
	if err != nil {
		global.LOG.Errorf("Create convert task failed %v", err)
		return
	}
	convertTask.AddSubTask(task.GetTaskName(i18n.GetMsgByKey("FileConvert"), task.TaskExec, task.TaskScopeFileConvert), func(t *task.Task) (err error) {
		for _, file := range req.Files {
			input := filepath.Join(file.Path, file.InputFile)
			nameOnly := file.InputFile[0 : len(file.InputFile)-len(file.Extension)]
			output := filepath.Join(req.OutputPath, nameOnly+"."+file.OutputFormat)
			status, errMsg := convert.MediaFile(input, output, file.OutputFormat, req.DeleteSource)
			if status == "FAILED" {
				convertTask.Log(fmt.Sprintf("%s -> %s [%s]: %s\n",
					input, output, status, errMsg))
			} else {
				convertTask.Log(fmt.Sprintf("%s -> %s [%s]: %s\n",
					input, output, status, "SUCCESS"))
			}
		}
		return nil
	}, nil)
	go func() {
		_ = convertTask.Execute()
	}()
}

func (f *FileService) ConvertLog(req dto.PageInfo) (total int64, data []response.FileConvertLog, err error) {
	logFilePath := filepath.Join(global.Dir.ConvertLogDir, "convert.log")
	file, err := os.Open(logFilePath)
	if err != nil {
		return 0, nil, err
	}
	defer file.Close()

	const chunkSize = 64 * 1024
	stat, err := file.Stat()
	if err != nil {
		return 0, nil, err
	}
	fileSize := stat.Size()

	var (
		buf       []byte
		remainder []byte
		offset    = fileSize
		lines     []string
	)

	pageStart := int64((req.Page - 1) * req.PageSize)
	pageEnd := pageStart + int64(req.PageSize)

	for offset > 0 {
		readSize := chunkSize
		if offset < int64(readSize) {
			readSize = int(offset)
		}
		offset -= int64(readSize)

		tmp := make([]byte, readSize)
		if _, err := file.ReadAt(tmp, offset); err != nil && err != io.EOF {
			return 0, nil, err
		}

		buf = append(tmp, remainder...)
		linesSplit := strings.Split(string(buf), "\n")

		if offset > 0 {
			remainder = []byte(linesSplit[0])
			linesSplit = linesSplit[1:]
		}

		for i := len(linesSplit) - 1; i >= 0; i-- {
			line := strings.TrimSpace(linesSplit[i])
			if line == "" {
				continue
			}
			total++
			if total > pageStart && total <= pageEnd {
				lines = append(lines, line)
			}
			if total >= pageEnd {
				break
			}
		}
		if total >= pageEnd {
			break
		}
	}

	for _, line := range lines {
		var entry response.FileConvertLog
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			data = append(data, entry)
		}
	}

	return total, data, nil
}

func (f *FileService) AISearch(req request.FileAISearch) (*response.FileAISearchResult, error) {
	root := filepath.Clean(strings.TrimSpace(req.Path))
	if root == "" {
		return nil, buserr.WithDetail("ErrInvalidParams", "path is required", nil)
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return nil, buserr.WithDetail("ErrInvalidParams", "query is required", nil)
	}
	st, err := os.Stat(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, buserr.New("ErrPathNotFound")
		}
		return nil, err
	}
	if !st.IsDir() {
		return nil, buserr.New("ErrPathNotFound")
	}

	maxItems := req.MaxItems
	if maxItems <= 0 {
		maxItems = files.DefaultFileAIMaxItems
	}
	if maxItems > 2000 {
		maxItems = 2000
	}

	containSub := true
	if req.ContainSub != nil {
		containSub = *req.ContainSub
	}

	searchOpts, err := files.MergeContentSearchOptions(
		req.MatchCase, req.WholeWord, req.UseRegex,
		req.Extensions,
		req.MinSize, req.MaxSize,
		req.ModifiedAfter, req.ModifiedBefore,
		req.MaxScanFiles,
		req.MaxFileBytes,
		req.MaxHitsPerFile, req.MaxTotalHits,
		req.ContentHitsPromptMaxBytes,
		req.LlmMaxOutputTokens,
	)
	if err != nil {
		return nil, buserr.WithDetail("ErrInvalidParams", err.Error(), nil)
	}

	matchFn, err := files.NewContentLineMatcher(query, searchOpts)
	if err != nil {
		return nil, buserr.WithDetail("ErrFileAISearchBadPattern", err.Error(), nil)
	}

	cfg, timeout, err := terminalai.LoadFileAIRuntimeConfig()
	aiEnabled := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	items, truncated, err := files.CollectDirInventory(root, containSub, maxItems)
	if err != nil {
		return nil, err
	}

	preFiltered := false
	llmItems := items
	qLower := strings.ToLower(query)
	if len(llmItems) > 0 && query != "" {
		filtered := make([]files.AISearchInventoryItem, 0, len(llmItems))
		for _, it := range llmItems {
			rel := strings.TrimSpace(it.RelPath)
			if rel == "" {
				continue
			}
			if !req.UseRegex && !req.MatchCase && !req.WholeWord && strings.Contains(strings.ToLower(rel), qLower) {
				filtered = append(filtered, it)
			}
		}
		if len(filtered) >= 8 {
			llmItems = filtered
			preFiltered = true
		}
	}

	start := time.Now()
	contentHits, scannedFiles, hitsTrunc := files.SearchFileAIContentHits(root, llmItems, searchOpts, matchFn)
	hitsDTO := make([]response.FileAIContentHit, 0, len(contentHits))
	for _, h := range contentHits {
		hitsDTO = append(hitsDTO, response.FileAIContentHit{Path: h.Path, Line: h.Line, Text: h.Text})
	}

	matchDesc := searchOpts.ContentMatchDescription()
	result := &response.FileAISearchResult{
		Hits:                 hitsDTO,
		ContentScannedFiles:  scannedFiles,
		ContentHitsTruncated: hitsTrunc,
		Truncated:            truncated,
		PreFiltered:          preFiltered,
		ItemCount:            len(llmItems),
	}

	if len(llmItems) == 0 {
		if aiEnabled {
			result.Mode = "ai"
			result.Summary = i18n.GetMsgByKey("FileAISearchEmptyDir")
			if result.Summary == "" || result.Summary == "FileAISearchEmptyDir" {
				result.Summary = "No files or directories found under this path (or all entries were filtered)."
			}
		} else {
			result.Mode = "grep"
			result.Summary = ""
		}
		result.Duration = time.Since(start).Round(time.Millisecond).String()
		return result, nil
	}

	if !aiEnabled {
		result.Mode = "grep"
		result.Summary = ""
		result.Duration = time.Since(start).Round(time.Millisecond).String()
		return result, nil
	}

	result.Mode = "ai"

	clientTimeout := timeout
	if clientTimeout < 30*time.Second {
		clientTimeout = 90 * time.Second
	}
	if clientTimeout > 5*time.Minute {
		clientTimeout = 5 * time.Minute
	}

	runCtx, cancel := context.WithTimeout(context.Background(), timeout+time.Minute)
	defer cancel()

	llmMaxOut := searchOpts.LlmMaxOutputTokens
	summary, usage, err := files.RunFileAISearchLLM(runCtx, cfg, clientTimeout, root, query, req.ResponseLanguage, llmItems, truncated, preFiltered, contentHits, scannedFiles, hitsTrunc, matchDesc, searchOpts.ContentHitsPromptMaxBytes, llmMaxOut)
	if err != nil {
		result.Mode = "grep"
		result.Summary = ""
		result.Duration = time.Since(start).Round(time.Millisecond).String()
		if errors.Is(err, context.DeadlineExceeded) || strings.Contains(strings.ToLower(err.Error()), "timeout") {
			return result, nil
		}
		return result, nil
	}
	result.Summary = summary
	result.PromptTokens = usage.PromptTokens
	result.CompletionTokens = usage.CompletionTokens
	result.TotalTokens = usage.TotalTokens
	if result.TotalTokens == 0 {
		result.TotalTokens = usage.PromptTokens + usage.CompletionTokens
	}
	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}
