package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"gopkg.in/square/go-jose.v2/json"
)

type WebSiteDnsAccountService struct {
}

func (w WebSiteDnsAccountService) Page(search dto.PageInfo) (int64, []dto.WebsiteDnsAccountDTO, error) {
	total, accounts, err := websiteDnsRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var accountDTOs []dto.WebsiteDnsAccountDTO
	for _, account := range accounts {
		auth := make(map[string]string)
		_ = json.Unmarshal([]byte(account.Authorization), &auth)
		accountDTOs = append(accountDTOs, dto.WebsiteDnsAccountDTO{
			WebsiteDnsAccount: account,
			Authorization:     auth,
		})
	}
	return total, accountDTOs, err
}

func (w WebSiteDnsAccountService) Create(create dto.WebsiteDnsAccountCreate) (dto.WebsiteDnsAccountCreate, error) {
	authorization, err := json.Marshal(create.Authorization)
	if err != nil {
		return dto.WebsiteDnsAccountCreate{}, err
	}

	if err := websiteDnsRepo.Create(model.WebsiteDnsAccount{
		Name:          create.Name,
		Type:          create.Type,
		Authorization: string(authorization),
	}); err != nil {
		return dto.WebsiteDnsAccountCreate{}, err
	}

	return create, nil
}

func (w WebSiteDnsAccountService) Update(update dto.WebsiteDnsAccountUpdate) (dto.WebsiteDnsAccountUpdate, error) {
	authorization, err := json.Marshal(update.Authorization)
	if err != nil {
		return dto.WebsiteDnsAccountUpdate{}, err
	}

	if err := websiteDnsRepo.Save(model.WebsiteDnsAccount{
		BaseModel: model.BaseModel{
			ID: update.ID,
		},
		Name:          update.Name,
		Type:          update.Type,
		Authorization: string(authorization),
	}); err != nil {
		return dto.WebsiteDnsAccountUpdate{}, err
	}

	return update, nil
}

func (w WebSiteDnsAccountService) Delete(id uint) error {
	if ssls, _ := websiteSSLRepo.List(websiteSSLRepo.WithByDnsAccountId(id)); len(ssls) > 0 {
		return buserr.New(constant.ErrAccountCannotDelete)
	}
	return websiteDnsRepo.DeleteBy(commonRepo.WithByID(id))
}
