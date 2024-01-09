package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
)

type PHPExtensionsService struct {
}

type IPHPExtensionsService interface {
	Page(req request.PHPExtensionsSearch) (int64, []response.PHPExtensionsDTO, error)
	List() ([]response.PHPExtensionsDTO, error)
	Create(req request.PHPExtensionsCreate) error
	Update(req request.PHPExtensionsUpdate) error
	Delete(req request.PHPExtensionsDelete) error
}

func NewIPHPExtensionsService() IPHPExtensionsService {
	return &PHPExtensionsService{}
}

func (p PHPExtensionsService) Page(req request.PHPExtensionsSearch) (int64, []response.PHPExtensionsDTO, error) {
	var (
		total      int64
		extensions []model.PHPExtensions
		err        error
		result     []response.PHPExtensionsDTO
	)
	total, extensions, err = phpExtensionsRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	for _, extension := range extensions {
		result = append(result, response.PHPExtensionsDTO{
			PHPExtensions: extension,
		})
	}
	return total, result, nil
}

func (p PHPExtensionsService) List() ([]response.PHPExtensionsDTO, error) {
	var (
		extensions []model.PHPExtensions
		err        error
		result     []response.PHPExtensionsDTO
	)
	extensions, err = phpExtensionsRepo.List()
	if err != nil {
		return nil, err
	}
	for _, extension := range extensions {
		result = append(result, response.PHPExtensionsDTO{
			PHPExtensions: extension,
		})
	}
	return result, nil
}

func (p PHPExtensionsService) Create(req request.PHPExtensionsCreate) error {
	exist, _ := phpExtensionsRepo.GetFirst(commonRepo.WithByName(req.Name))
	if exist.ID > 0 {
		return buserr.New(constant.ErrNameIsExist)
	}
	extension := model.PHPExtensions{
		Name:       req.Name,
		Extensions: req.Extensions,
	}
	return phpExtensionsRepo.Create(&extension)
}

func (p PHPExtensionsService) Update(req request.PHPExtensionsUpdate) error {
	exist, err := phpExtensionsRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	exist.Extensions = req.Extensions
	return phpExtensionsRepo.Save(&exist)
}

func (p PHPExtensionsService) Delete(req request.PHPExtensionsDelete) error {
	return phpExtensionsRepo.DeleteBy(commonRepo.WithByID(req.ID))
}
