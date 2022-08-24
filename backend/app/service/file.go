package service

import (
	"crypto/rand"
	"fmt"
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/utils/files"
	"io"
)

type FileService struct {
}

type IFileService interface {
	GetFileList(op dto.FileOption) (dto.FileInfo, error)
}

func NewFileService() IFileService {
	return FileService{}
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

func getUuid() string {
	b := make([]byte, 16)
	io.ReadFull(rand.Reader, b)
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}
