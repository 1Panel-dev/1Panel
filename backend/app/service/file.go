package service

import (
	"crypto/rand"
	"fmt"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/global"
	"github.com/1Panel-dev/1Panel/utils/files"
	"github.com/pkg/errors"
	"io"
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
		ID:   getUuid(),
		Name: info.Name,
		Path: info.Path,
	}
	for _, v := range info.Items {
		if v.IsDir {
			node.Children = append(node.Children, dto.FileTree{
				ID:   getUuid(),
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

func (f FileService) Wget(w dto.FileWget) error {
	fo := files.NewFileOp()
	return fo.DownloadFile(w.Url, filepath.Join(w.Path, w.Name))
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

func getUuid() string {
	b := make([]byte, 16)
	_, _ = io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
