package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
)

type WebSiteAcmeAccountService struct {
}

func (w WebSiteAcmeAccountService) Page(search dto.PageInfo) (int64, []dto.WebsiteAcmeAccountDTO, error) {
	total, accounts, err := websiteAcmeRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var accountDTOs []dto.WebsiteAcmeAccountDTO
	for _, account := range accounts {
		accountDTOs = append(accountDTOs, dto.WebsiteAcmeAccountDTO{
			WebsiteAcmeAccount: account,
		})
	}
	return total, accountDTOs, err
}

func (w WebSiteAcmeAccountService) Create(create dto.WebsiteAcmeAccountCreate) (dto.WebsiteAcmeAccountDTO, error) {
	client, err := ssl.NewAcmeClient(create.Email, "")
	if err != nil {
		return dto.WebsiteAcmeAccountDTO{}, err
	}
	acmeAccount := model.WebsiteAcmeAccount{
		Email:      create.Email,
		URL:        client.User.Registration.URI,
		PrivateKey: string(ssl.GetPrivateKey(client.User.GetPrivateKey())),
	}
	if err := websiteAcmeRepo.Create(acmeAccount); err != nil {
		return dto.WebsiteAcmeAccountDTO{}, err
	}
	return dto.WebsiteAcmeAccountDTO{WebsiteAcmeAccount: acmeAccount}, nil
}

func (w WebSiteAcmeAccountService) Delete(id uint) error {
	return websiteAcmeRepo.DeleteBy(commonRepo.WithByID(id))
}
