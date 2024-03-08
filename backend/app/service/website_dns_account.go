package service

import (
	"encoding/json"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
)

type WebsiteDnsAccountService struct {
}

type IWebsiteDnsAccountService interface {
	Page(search dto.PageInfo) (int64, []response.WebsiteDnsAccountDTO, error)
	Create(create request.WebsiteDnsAccountCreate) (request.WebsiteDnsAccountCreate, error)
	Update(update request.WebsiteDnsAccountUpdate) (request.WebsiteDnsAccountUpdate, error)
	Delete(id uint) error
}

func NewIWebsiteDnsAccountService() IWebsiteDnsAccountService {
	return &WebsiteDnsAccountService{}
}

func (w WebsiteDnsAccountService) Page(search dto.PageInfo) (int64, []response.WebsiteDnsAccountDTO, error) {
	total, accounts, err := websiteDnsRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var accountDTOs []response.WebsiteDnsAccountDTO
	for _, account := range accounts {
		auth := make(map[string]string)
		_ = json.Unmarshal([]byte(account.Authorization), &auth)
		accountDTOs = append(accountDTOs, response.WebsiteDnsAccountDTO{
			WebsiteDnsAccount: account,
			Authorization:     auth,
		})
	}
	return total, accountDTOs, err
}

func (w WebsiteDnsAccountService) Create(create request.WebsiteDnsAccountCreate) (request.WebsiteDnsAccountCreate, error) {
	exist, _ := websiteDnsRepo.GetFirst(commonRepo.WithByName(create.Name))
	if exist != nil {
		return request.WebsiteDnsAccountCreate{}, buserr.New(constant.ErrNameIsExist)
	}

	authorization, err := json.Marshal(create.Authorization)
	if err != nil {
		return request.WebsiteDnsAccountCreate{}, err
	}

	if err := websiteDnsRepo.Create(model.WebsiteDnsAccount{
		Name:          create.Name,
		Type:          create.Type,
		Authorization: string(authorization),
	}); err != nil {
		return request.WebsiteDnsAccountCreate{}, err
	}

	return create, nil
}

func (w WebsiteDnsAccountService) Update(update request.WebsiteDnsAccountUpdate) (request.WebsiteDnsAccountUpdate, error) {
	authorization, err := json.Marshal(update.Authorization)
	if err != nil {
		return request.WebsiteDnsAccountUpdate{}, err
	}
	exists, _ := websiteDnsRepo.List(commonRepo.WithByName(update.Name))
	for _, exist := range exists {
		if exist.ID != update.ID {
			return request.WebsiteDnsAccountUpdate{}, buserr.New(constant.ErrNameIsExist)
		}
	}
	if err := websiteDnsRepo.Save(model.WebsiteDnsAccount{
		BaseModel: model.BaseModel{
			ID: update.ID,
		},
		Name:          update.Name,
		Type:          update.Type,
		Authorization: string(authorization),
	}); err != nil {
		return request.WebsiteDnsAccountUpdate{}, err
	}

	return update, nil
}

func (w WebsiteDnsAccountService) Delete(id uint) error {
	if ssls, _ := websiteSSLRepo.List(websiteSSLRepo.WithByDnsAccountId(id)); len(ssls) > 0 {
		return buserr.New(constant.ErrAccountCannotDelete)
	}
	return websiteDnsRepo.DeleteBy(commonRepo.WithByID(id))
}
