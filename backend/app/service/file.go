package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/common"
	"github.com/1Panel-dev/1Panel/utils/files"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileService struct {
}

func (f FileService) GetFileList(op dto.FileOption) (dto.FileInfo, error) {
	var fileInfo dto.FileInfo
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return fileInfo, err
	}
	fileInfo.FileInfo = *info
	return fileInfo, nil
}

func (f FileService) GetFileTree(op dto.FileOption) ([]dto.FileTree, error) {
	var treeArray []dto.FileTree
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return nil, err
	}
	node := dto.FileTree{
		ID:   common.GetUuid(),
		Name: info.Name,
		Path: info.Path,
	}
	for _, v := range info.Items {
		if v.IsDir {
			node.Children = append(node.Children, dto.FileTree{
				ID:   common.GetUuid(),
				Name: v.Name,
				Path: v.Path,
			})
		}
	}
	return append(treeArray, node), nil
}

func (f FileService) Create(op dto.FileCreate) error {

	fo := files.NewFileOp()
	if fo.Stat(op.Path) {
		return errors.New("file is exist")
	}
	if op.IsDir {
		return fo.CreateDir(op.Path, fs.FileMode(op.Mode))
	} else {
		if op.IsLink {
			return fo.LinkFile(op.LinkPath, op.Path, op.IsSymlink)
		} else {
			return fo.CreateFile(op.Path)
		}
	}
}

func (f FileService) Delete(op dto.FileDelete) error {
	fo := files.NewFileOp()
	if op.IsDir {
		return fo.DeleteDir(op.Path)
	} else {
		return fo.DeleteFile(op.Path)
	}
}

func (f FileService) ChangeMode(op dto.FileCreate) error {
	fo := files.NewFileOp()
	return fo.Chmod(op.Path, fs.FileMode(op.Mode))
}

func (f FileService) Compress(c dto.FileCompress) error {
	fo := files.NewFileOp()
	if !c.Replace && fo.Stat(filepath.Join(c.Dst, c.Name)) {
		return errors.New("file is exist")
	}

	return fo.Compress(c.Files, c.Dst, c.Name, files.CompressType(c.Type))
}

func (f FileService) DeCompress(c dto.FileDeCompress) error {
	fo := files.NewFileOp()
	return fo.Decompress(c.Path, c.Dst, files.CompressType(c.Type))
}

func (f FileService) GetContent(op dto.FileOption) (dto.FileInfo, error) {
	info, err := files.NewFileInfo(op.FileOption)
	if err != nil {
		return dto.FileInfo{}, err
	}
	return dto.FileInfo{FileInfo: *info}, nil
}

func (f FileService) SaveContent(edit dto.FileEdit) error {

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

func (f FileService) ChangeName(re dto.FileRename) error {
	fo := files.NewFileOp()
	return fo.Rename(re.OldName, re.NewName)
}

func (f FileService) Wget(w dto.FileWget) (string, error) {
	fo := files.NewFileOp()
	key := "file-wget-" + uuid.NewV4().String()
	return key, fo.DownloadFile(w.Url, filepath.Join(w.Path, w.Name), key)
}

func (f FileService) MvFile(m dto.FileMove) error {
	fo := files.NewFileOp()
	if m.Type == "cut" {
		return fo.Cut(m.OldPaths, m.NewPath)
	}
	var errs []error
	if m.Type == "copy" {
		for _, src := range m.OldPaths {
			if err := fo.Copy(src, m.NewPath); err != nil {
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

func (f FileService) FileDownload(d dto.FileDownload) (string, error) {
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("%d", time.Now().UnixNano()))
	if err := os.MkdirAll(tempPath, os.ModePerm); err != nil {
		return "", err
	}
	fo := files.NewFileOp()
	if err := fo.Compress(d.Paths, tempPath, d.Name, files.CompressType(d.Type)); err != nil {
		return "", err
	}
	filePath := filepath.Join(tempPath, d.Name)
	return filePath, nil
}

func (f FileService) DirSize(req dto.DirSizeReq) (dto.DirSizeRes, error) {
	fo := files.NewFileOp()
	size, err := fo.GetDirSize(req.Path)
	if err != nil {
		return dto.DirSizeRes{}, err
	}
	return dto.DirSizeRes{Size: size}, nil
}
