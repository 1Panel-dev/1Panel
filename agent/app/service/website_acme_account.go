package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/ssl"
)

type WebsiteAcmeAccountService struct {
}

type IWebsiteAcmeAccountService interface {
	Page(search dto.PageInfo) (int64, []response.WebsiteAcmeAccountDTO, error)
	Create(create request.WebsiteAcmeAccountCreate) (*response.WebsiteAcmeAccountDTO, error)
	Delete(id uint) error
}

func NewIWebsiteAcmeAccountService() IWebsiteAcmeAccountService {
	return &WebsiteAcmeAccountService{}
}

func (w WebsiteAcmeAccountService) Page(search dto.PageInfo) (int64, []response.WebsiteAcmeAccountDTO, error) {
	total, accounts, err := websiteAcmeRepo.Page(search.Page, search.PageSize, repo.WithOrderBy("created_at desc"))
	var accountDTOs []response.WebsiteAcmeAccountDTO
	for _, account := range accounts {
		accountDTOs = append(accountDTOs, response.WebsiteAcmeAccountDTO{
			WebsiteAcmeAccount: account,
		})
	}
	return total, accountDTOs, err
}

func (w WebsiteAcmeAccountService) Create(create request.WebsiteAcmeAccountCreate) (*response.WebsiteAcmeAccountDTO, error) {
	exist, _ := websiteAcmeRepo.GetFirst(websiteAcmeRepo.WithEmail(create.Email), websiteAcmeRepo.WithType(create.Type))
	if exist != nil {
		return nil, buserr.New("ErrEmailIsExist")
	}
	acmeAccount := &model.WebsiteAcmeAccount{
		Email:   create.Email,
		Type:    create.Type,
		KeyType: create.KeyType,
	}

	if create.Type == "google" {
		if create.EabKid == "" || create.EabHmacKey == "" {
			return nil, buserr.New("ErrEabKidOrEabHmacKeyCannotBlank")
		}
		acmeAccount.EabKid = create.EabKid
		acmeAccount.EabHmacKey = create.EabHmacKey
	}

	client, err := ssl.NewAcmeClient(acmeAccount)
	if err != nil {
		return nil, err
	}
	privateKey, err := ssl.GetPrivateKey(client.User.GetPrivateKey(), ssl.KeyType(create.KeyType))
	if err != nil {
		return nil, err
	}
	acmeAccount.PrivateKey = string(privateKey)
	acmeAccount.URL = client.User.Registration.URI

	if err := websiteAcmeRepo.Create(*acmeAccount); err != nil {
		return nil, err
	}
	return &response.WebsiteAcmeAccountDTO{WebsiteAcmeAccount: *acmeAccount}, nil
}

func (w WebsiteAcmeAccountService) Delete(id uint) error {
	if ssls, _ := websiteSSLRepo.List(websiteSSLRepo.WithByAcmeAccountId(id)); len(ssls) > 0 {
		return buserr.New("ErrAccountCannotDelete")
	}
	return websiteAcmeRepo.DeleteBy(repo.WithByID(id))
}
