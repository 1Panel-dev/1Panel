package service

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
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
	ChangeName(req request.FileRename) error
	Wget(w request.FileWget) (string, error)
	MvFile(m request.FileMove) error
	ChangeOwner(req request.FileRoleUpdate) error
	ChangeMode(op request.FileCreate) error
	BatchChangeModeAndOwner(op request.FileRoleReq) error
	ReadLogByLine(req request.FileReadByLineReq) (*response.FileLineContent, error)
}

func NewIFileService() IFileService {
	return &FileService{}
}

func (f *FileService) GetFileList(op request.FileOption) (response.FileInfo, error) {
	var fileInfo response.FileInfo
	if _, err := os.Stat(op.Path); err != nil && os.IsNotExist(err) {
		return fileInfo, nil
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
				CreatedAt: info.ModTime().Format("2006-01-02 15:04:05"),
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
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return nil, err
	}
	node := response.FileTree{
		ID:   common.GetUuid(),
		Name: info.Name,
		Path: info.Path,
	}
	for _, v := range info.Items {
		if v.IsDir {
			node.Children = append(node.Children, response.FileTree{
				ID:   common.GetUuid(),
				Name: v.Name,
				Path: v.Path,
			})
		}
	}
	return append(treeArray, node), nil
}

func (f *FileService) Create(op request.FileCreate) error {
	fo := files.NewFileOp()
	if fo.Stat(op.Path) {
		return buserr.New(constant.ErrFileIsExit)
	}
	if op.IsDir {
		return fo.CreateDir(op.Path, fs.FileMode(op.Mode))
	} else {
		if op.IsLink {
			if !fo.Stat(op.LinkPath) {
				return buserr.New(constant.ErrLinkPathNotFound)
			}
			return fo.LinkFile(op.LinkPath, op.Path, op.IsSymlink)
		} else {
			return fo.CreateFile(op.Path)
		}
	}
}

func (f *FileService) Delete(op request.FileDelete) error {
	fo := files.NewFileOp()
	recycleBinStatus, _ := settingRepo.Get(settingRepo.WithByKey("FileRecycleBin"))
	if recycleBinStatus.Value == "disable" {
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
			return buserr.New(constant.ErrPathNotFound)
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
		return buserr.New(constant.ErrFileIsExit)
	}
	return fo.Compress(c.Files, c.Dst, c.Name, files.CompressType(c.Type))
}

func (f *FileService) DeCompress(c request.FileDeCompress) error {
	fo := files.NewFileOp()
	return fo.Decompress(c.Path, c.Dst, files.CompressType(c.Type))
}

func (f *FileService) GetContent(op request.FileContentReq) (response.FileInfo, error) {
	info, err := files.NewFileInfo(files.FileOption{
		Path:   op.Path,
		Expand: true,
	})
	if err != nil {
		return response.FileInfo{}, err
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
		return buserr.New(constant.ErrPathNotFound)
	}
	for _, oldPath := range m.OldPaths {
		if !fo.Stat(oldPath) {
			return buserr.WithName(constant.ErrFileNotFound, oldPath)
		}
		if oldPath == m.NewPath || strings.Contains(m.NewPath, filepath.Clean(oldPath)+"/") {
			return buserr.New(constant.ErrMovePathFailed)
		}
	}
	if m.Type == "cut" {
		return fo.Cut(m.OldPaths, m.NewPath, m.Name, m.Cover)
	}
	var errs []error
	if m.Type == "copy" {
		for _, src := range m.OldPaths {
			if err := fo.CopyAndReName(src, m.NewPath, m.Name, m.Cover); err != nil {
				errs = append(errs, err)
				global.LOG.Errorf("copy file [%s] to [%s] failed, err: %s", src, m.NewPath, err.Error())
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
		if err := fo.Compress(d.Paths, tempPath, d.Name, files.CompressType(d.Type)); err != nil {
			return "", err
		}
		filePath = filepath.Join(tempPath, d.Name)
	}
	return filePath, nil
}

func (f *FileService) DirSize(req request.DirSizeReq) (response.DirSizeRes, error) {
	fo := files.NewFileOp()
	size, err := fo.GetDirSize(req.Path)
	if err != nil {
		return response.DirSizeRes{}, err
	}
	return response.DirSizeRes{Size: size}, nil
}

func (f *FileService) ReadLogByLine(req request.FileReadByLineReq) (*response.FileLineContent, error) {
	logFilePath := ""
	switch req.Type {
	case constant.TypeWebsite:
		website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		nginx, err := getNginxFull(&website)
		if err != nil {
			return nil, err
		}
		sitePath := path.Join(nginx.SiteDir, "sites", website.Alias)
		logFilePath = path.Join(sitePath, "log", req.Name)
	case constant.TypePhp:
		php, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		logFilePath = php.GetLogPath()
	case constant.TypeSSL:
		ssl, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(req.ID))
		if err != nil {
			return nil, err
		}
		logFilePath = ssl.GetLogPath()
	case constant.TypeSystem:
		fileName := ""
		if len(req.Name) == 0 || req.Name == time.Now().Format("2006-01-02") {
			fileName = "1Panel.log"
		} else {
			fileName = "1Panel-" + req.Name + ".log"
		}
		logFilePath = path.Join(global.CONF.System.DataDir, "log", fileName)
		if _, err := os.Stat(logFilePath); err != nil {
			fileGzPath := path.Join(global.CONF.System.DataDir, "log", fileName+".gz")
			if _, err := os.Stat(fileGzPath); err != nil {
				return nil, buserr.New("ErrHttpReqNotFound")
			}
			if err := handleGunzip(fileGzPath); err != nil {
				return nil, fmt.Errorf("handle ungzip file %s failed, err: %v", fileGzPath, err)
			}
		}
	case "image-pull", "image-push", "image-build", "compose-create":
		logFilePath = path.Join(global.CONF.System.TmpDir, fmt.Sprintf("docker_logs/%s", req.Name))
	}

	lines, isEndOfFile, err := files.ReadFileByLine(logFilePath, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	res := &response.FileLineContent{
		Content: strings.Join(lines, "\n"),
		End:     isEndOfFile,
		Path:    logFilePath,
	}
	return res, nil
}
