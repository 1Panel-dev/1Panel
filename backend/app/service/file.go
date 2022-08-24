package service

import (
	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/utils/files"
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
		Name: info.Name,
		Path: info.Path,
	}
	for _, v := range info.Items {
		if v.IsDir {
			node.Children = append(node.Children, dto.FileTree{
				Name: v.Name,
				Path: v.Path,
			})
		}
	}
	return append(treeArray, node), nil
}
