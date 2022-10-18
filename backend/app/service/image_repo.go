package service

import (
	"encoding/json"
	"io/ioutil"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type ImageRepoService struct{}

type IImageRepoService interface {
	Page(search dto.PageInfo) (int64, interface{}, error)
	List() ([]dto.ImageRepoOption, error)
	Create(imageRepoDto dto.ImageRepoCreate) error
	Update(id uint, upMap map[string]interface{}) error
	BatchDelete(ids []uint) error
}

func NewIImageRepoService() IImageRepoService {
	return &ImageRepoService{}
}

func (u *ImageRepoService) Page(search dto.PageInfo) (int64, interface{}, error) {
	total, ops, err := imageRepoRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.ImageRepoInfo
	for _, op := range ops {
		var item dto.ImageRepoInfo
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoOps = append(dtoOps, item)
	}
	return total, dtoOps, err
}

func (u *ImageRepoService) List() ([]dto.ImageRepoOption, error) {
	ops, err := imageRepoRepo.List(commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.ImageRepoOption
	for _, op := range ops {
		var item dto.ImageRepoOption
		if err := copier.Copy(&item, &op); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoOps = append(dtoOps, item)
	}
	return dtoOps, err
}

func (u *ImageRepoService) Create(imageRepoDto dto.ImageRepoCreate) error {
	imageRepo, _ := imageRepoRepo.Get(commonRepo.WithByName(imageRepoDto.Name))
	if imageRepo.ID != 0 {
		return constant.ErrRecordExist
	}
	if imageRepo.Protocol == "http" {
		file, err := ioutil.ReadFile(constant.DaemonJsonDir)
		if err != nil {
			return err
		}

		deamonMap := make(map[string]interface{})
		if err := json.Unmarshal(file, &deamonMap); err != nil {
			return err
		}
		if _, ok := deamonMap["insecure-registries"]; ok {
			if k, v := deamonMap["insecure-registries"].([]interface{}); v {
				k = append(k, imageRepoDto.DownloadUrl)
				deamonMap["insecure-registries"] = k
			}
		}
		newJson, err := json.Marshal(deamonMap)
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(constant.DaemonJsonDir, newJson, 0777); err != nil {
			return err
		}
	}
	if err := copier.Copy(&imageRepo, &imageRepoDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := imageRepoRepo.Create(&imageRepo); err != nil {
		return err
	}
	return nil
}

type DeamonJson struct {
	InsecureRegistries []string `json:"insecure-registries"`
}

func (u *ImageRepoService) BatchDelete(ids []uint) error {
	for _, id := range ids {
		if id == 1 {
			return errors.New("The default value cannot be edit !")
		}
	}
	return imageRepoRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *ImageRepoService) Update(id uint, upMap map[string]interface{}) error {
	if id == 1 {
		return errors.New("The default value cannot be deleted !")
	}
	return imageRepoRepo.Update(id, upMap)
}
