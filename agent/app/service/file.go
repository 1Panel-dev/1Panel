package service

import (
	"bufio"
	"context"
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
	DeCompress(c request.FileDeCompress) error
	GetContent(op request.FileContentReq) (response.FileInfo, error)
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
}

var filteredPaths = []string{
	"/.1panel_clash",
}

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
	fileInfo.FileInfo = *info
	return fileInfo, nil
}

func (f *FileService) SearchUploadWithPage(req request.SearchUploadWithPage) (int64, interface{}, error) {
	var (
		files    []response.UploadInfo
		backData []response.UploadInfo
	)
	_ = filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			files = append(files, response.UploadInfo{
				CreatedAt: info.ModTime().Format(constant.DateTimeLayout),
				Size:      int(info.Size()),
				Name:      info.Name(),
			})
		}
		return nil
	})
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
	cleanedPath := filepath.Clean(path)
	for _, filteredPath := range filteredPaths {
		cleanedFilteredPath := filepath.Clean(filteredPath)
		if cleanedFilteredPath == cleanedPath || strings.HasPrefix(cleanedPath, cleanedFilteredPath+"/") {
			return true
		}
	}
	return false
}

// 递归构建文件树(只取当前目录以及当前目录下的第一层子节点)
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

func (f *FileService) Create(op request.FileCreate) error {
	if files.IsInvalidChar(op.Path) {
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
		if filepath.Base(op.Path) == ".1panel_clash" || op.Path == excludeDir {
			return buserr.New("ErrPathNotDelete")
		}
	}
	fo := files.NewFileOp()
	recycleBinStatus, _ := settingRepo.Get(settingRepo.WithByKey("FileRecycleBin"))
	if recycleBinStatus.Value == "Disable" {
		op.ForceDelete = true
	}
	if op.ForceDelete {
		if op.IsDir {
			return fo.DeleteDir(op.Path)
		} else {
			return fo.DeleteFile(op.Path)
		}
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
			if err := fo.DeleteDir(file); err != nil {
				return err
			}
		}
	} else {
		for _, file := range op.Paths {
			if err := fo.DeleteFile(file); err != nil {
				return err
			}
		}
	}
	return nil
}

func (f *FileService) ChangeMode(op request.FileCreate) error {
	fo := files.NewFileOp()
	return fo.ChmodR(op.Path, op.Mode, op.Sub)
}

func (f *FileService) BatchChangeModeAndOwner(op request.FileRoleReq) error {
	fo := files.NewFileOp()
	for _, path := range op.Paths {
		if !fo.Stat(path) {
			return buserr.New("ErrPathNotFound")
		}
		if err := fo.ChownR(path, op.User, op.Group, op.Sub); err != nil {
			return err
		}
		if err := fo.ChmodR(path, op.Mode, op.Sub); err != nil {
			return err
		}
	}
	return nil

}

func (f *FileService) ChangeOwner(req request.FileRoleUpdate) error {
	fo := files.NewFileOp()
	return fo.ChownR(req.Path, req.User, req.Group, req.Sub)
}

func (f *FileService) Compress(c request.FileCompress) error {
	fo := files.NewFileOp()
	if !c.Replace && fo.Stat(filepath.Join(c.Dst, c.Name)) {
		return buserr.New("ErrFileIsExist")
	}
	return fo.Compress(c.Files, c.Dst, c.Name, files.CompressType(c.Type), c.Secret)
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

func (f *FileService) SaveContent(edit request.FileEdit) error {
	info, err := files.NewFileInfo(files.FileOption{
		Path:   edit.Path,
		Expand: false,
	})
	if err != nil {
		return err
	}

	fo := files.NewFileOp()
	return fo.WriteFile(edit.Path, strings.NewReader(edit.Content), info.FileMode)
}

func (f *FileService) ChangeName(req request.FileRename) error {
	if files.IsInvalidChar(req.NewName) {
		return buserr.New("ErrInvalidChar")
	}
	fo := files.NewFileOp()
	return fo.Rename(req.OldName, req.NewName)
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
		return fo.Cut(m.OldPaths, m.NewPath, m.Name, m.Cover)
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

func (f *FileService) FileDownload(d request.FileDownload) (string, error) {
	filePath := d.Paths[0]
	if d.Compress {
		tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().UnixNano()))
		if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
			return "", err
		}
		fo := files.NewFileOp()
		if err := fo.Compress(d.Paths, tempPath, d.Name, files.CompressType(d.Type), ""); err != nil {
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
		logFilePath = path.Join(global.Dir.DataDir, fmt.Sprintf("apps/mysql/%s/data/1Panel-slow.log", req.Name))
	case "mariadb-slow-logs":
		logFilePath = path.Join(global.Dir.DataDir, fmt.Sprintf("apps/mariadb/%s/db/data/1Panel-slow.log", req.Name))
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
		total       int
		scope       string
	)

	if stat.Size() > 500*1024*1024 {
		lines, err = files.TailFromEnd(logFilePath, req.PageSize)
		isEndOfFile = true
		scope = "tail"
	} else {
		lines, isEndOfFile, total, err = files.ReadFileByLine(logFilePath, req.Page, req.PageSize, req.Latest)
		if err != nil {
			return nil, err
		}
		if req.Latest && req.Page == 1 && len(lines) < 1000 && total > 1 {
			preLines, _, _, err := files.ReadFileByLine(logFilePath, total-1, req.PageSize, false)
			if err != nil {
				return nil, err
			}
			lines = append(preLines, lines...)
		}
		scope = "page"
	}

	res := &response.FileLineContent{
		End:        isEndOfFile,
		Path:       logFilePath,
		Total:      total,
		TaskStatus: taskStatus,
		Lines:      lines,
		Scope:      scope,
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
