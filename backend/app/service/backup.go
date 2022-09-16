package service

import (
	"encoding/json"

	"github.com/1Panel-dev/1Panel/app/dto"
	"github.com/1Panel-dev/1Panel/constant"
	"github.com/1Panel-dev/1Panel/utils/cloud_storage"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type BackupService struct{}

type IBackupService interface {
	Page(search dto.PageInfo) (int64, interface{}, error)
	Create(backupDto dto.BackupOperate) error
	GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error)
	Update(id uint, upMap map[string]interface{}) error
	BatchDelete(ids []uint) error
}

func NewIBackupService() IBackupService {
	return &BackupService{}
}

func (u *BackupService) Page(search dto.PageInfo) (int64, interface{}, error) {
	total, ops, err := backupRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var dtobas []dto.BackupInfo
	for _, group := range ops {
		var item dto.BackupInfo
		if err := copier.Copy(&item, &group); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtobas = append(dtobas, item)
	}
	return total, dtobas, err
}

func (u *BackupService) Create(backupDto dto.BackupOperate) error {
	backup, _ := backupRepo.Get(commonRepo.WithByName(backupDto.Name))
	if backup.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&backup, &backupDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := backupRepo.Create(&backup); err != nil {
		return err
	}
	var backupinfo dto.BackupInfo
	if err := copier.Copy(&backupinfo, &backup); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	return nil
}

func (u *BackupService) GetBuckets(backupDto dto.ForBuckets) ([]interface{}, error) {
	varMap := make(map[string]interface{})
	if err := json.Unmarshal([]byte(backupDto.Vars), &varMap); err != nil {
		return nil, err
	}
	varMap["type"] = backupDto.Type
	switch backupDto.Type {
	case constant.Sftp:
		varMap["password"] = backupDto.Credential
	case constant.OSS, constant.S3, constant.MinIo:
		varMap["secretKey"] = backupDto.Credential
	}
	client, err := cloud_storage.NewCloudStorageClient(varMap)
	if err != nil {
		return nil, err
	}
	return client.ListBuckets()
}

func (u *BackupService) BatchDelete(ids []uint) error {
	return backupRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *BackupService) Update(id uint, upMap map[string]interface{}) error {
	return backupRepo.Update(id, upMap)
}
