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
