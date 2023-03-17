package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type ComposeTemplateService struct{}

type IComposeTemplateService interface {
	List() ([]dto.ComposeTemplateInfo, error)
	SearchWithPage(search dto.SearchWithPage) (int64, interface{}, error)
	Create(composeDto dto.ComposeTemplateCreate) error
	Update(id uint, upMap map[string]interface{}) error
	Delete(ids []uint) error
}

func NewIComposeTemplateService() IComposeTemplateService {
	return &ComposeTemplateService{}
}

func (u *ComposeTemplateService) List() ([]dto.ComposeTemplateInfo, error) {
	composes, err := composeRepo.List()
	if err != nil {
		return nil, constant.ErrRecordNotFound
	}
	var dtoLists []dto.ComposeTemplateInfo
	for _, compose := range composes {
		var item dto.ComposeTemplateInfo
		if err := copier.Copy(&item, &compose); err != nil {
			return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoLists = append(dtoLists, item)
	}
	return dtoLists, err
}

func (u *ComposeTemplateService) SearchWithPage(req dto.SearchWithPage) (int64, interface{}, error) {
	total, composes, err := composeRepo.Page(req.Page, req.PageSize, commonRepo.WithLikeName(req.Info))
	var dtoComposeTemplates []dto.ComposeTemplateInfo
	for _, compose := range composes {
		var item dto.ComposeTemplateInfo
		if err := copier.Copy(&item, &compose); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoComposeTemplates = append(dtoComposeTemplates, item)
	}
	return total, dtoComposeTemplates, err
}

func (u *ComposeTemplateService) Create(composeDto dto.ComposeTemplateCreate) error {
	compose, _ := composeRepo.Get(commonRepo.WithByName(composeDto.Name))
	if compose.ID != 0 {
		return constant.ErrRecordExist
	}
	if err := copier.Copy(&compose, &composeDto); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	if err := composeRepo.Create(&compose); err != nil {
		return err
	}
	return nil
}

func (u *ComposeTemplateService) Delete(ids []uint) error {
	if len(ids) == 1 {
		compose, _ := composeRepo.Get(commonRepo.WithByID(ids[0]))
		if compose.ID == 0 {
			return constant.ErrRecordNotFound
		}
		return composeRepo.Delete(commonRepo.WithByID(ids[0]))
	}
	return composeRepo.Delete(commonRepo.WithIdsIn(ids))
}

func (u *ComposeTemplateService) Update(id uint, upMap map[string]interface{}) error {
	return composeRepo.Update(id, upMap)
}
