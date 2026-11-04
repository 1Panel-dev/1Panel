package service

import (
	"errors"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"gorm.io/gorm"
)

var phpExtensionsMu sync.Mutex

type PHPExtensionsService struct {
}

type IPHPExtensionsService interface {
	Page(req request.PHPExtensionsSearch) (int64, []response.PHPExtensionsDTO, error)
	List() ([]response.PHPExtensionsDTO, error)
	Create(req request.PHPExtensionsCreate) error
	Update(req request.PHPExtensionsUpdate) error
	Delete(req request.PHPExtensionsDelete) error
	Get(id uint) (model.PHPExtensions, error)
	Import(req request.PHPExtensionsCreate) (string, error)
	Push(req dto.PHPExtensionsPush) error
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
	total, extensions, err = phpExtensionsRepo.Page(req.Page, req.PageSize, repo.WithByLikeName(req.Name))
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
	phpExtensionsMu.Lock()
	defer phpExtensionsMu.Unlock()
	exist, _ := phpExtensionsRepo.GetFirst(repo.WithByName(req.Name))
	if exist.ID > 0 {
		return buserr.New("ErrNameIsExist")
	}
	extension := model.PHPExtensions{
		Name:       req.Name,
		Extensions: req.Extensions,
	}
	return phpExtensionsRepo.Create(&extension)
}

func (p PHPExtensionsService) Update(req request.PHPExtensionsUpdate) error {
	phpExtensionsMu.Lock()
	defer phpExtensionsMu.Unlock()
	exist, err := phpExtensionsRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	exist.Extensions = req.Extensions
	return phpExtensionsRepo.Save(&exist)
}

func (p PHPExtensionsService) Delete(req request.PHPExtensionsDelete) error {
	phpExtensionsMu.Lock()
	defer phpExtensionsMu.Unlock()
	return phpExtensionsRepo.DeleteBy(repo.WithByID(req.ID))
}

func (p PHPExtensionsService) Get(id uint) (model.PHPExtensions, error) {
	return phpExtensionsRepo.GetFirst(repo.WithByID(id))
}

func (p PHPExtensionsService) Import(req request.PHPExtensionsCreate) (string, error) {
	phpExtensionsMu.Lock()
	defer phpExtensionsMu.Unlock()
	existing, err := phpExtensionsRepo.GetFirst(repo.WithByName(req.Name))
	if err == nil {
		if existing.Extensions == req.Extensions {
			return "skipped", nil
		}
		return "", buserr.New("ErrPHPExtensionsConflict")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	item := model.PHPExtensions{Name: req.Name, Extensions: req.Extensions}
	if err := phpExtensionsRepo.Create(&item); err != nil {
		return "", err
	}
	return "created", nil
}

func (p PHPExtensionsService) Push(req dto.PHPExtensionsPush) error {
	if !global.IsMaster {
		return errors.New("PHP extension template sync must be submitted to the master node")
	}
	if !xpack.MultiNodeProvider.IsXpack() {
		return buserr.New("ErrXpackNotFound")
	}
	return xpack.MultiNodeProvider.PushPHPExtensionsToNode(req)
}
