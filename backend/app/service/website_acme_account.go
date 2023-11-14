package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/ssl"
)

type WebsiteAcmeAccountService struct {
}

type IWebsiteAcmeAccountService interface {
	Page(search dto.PageInfo) (int64, []response.WebsiteAcmeAccountDTO, error)
	Create(create request.WebsiteAcmeAccountCreate) (response.WebsiteAcmeAccountDTO, error)
	Delete(id uint) error
}

func NewIWebsiteAcmeAccountService() IWebsiteAcmeAccountService {
	return &WebsiteAcmeAccountService{}
}

func (w WebsiteAcmeAccountService) Page(search dto.PageInfo) (int64, []response.WebsiteAcmeAccountDTO, error) {
	total, accounts, err := websiteAcmeRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
	var accountDTOs []response.WebsiteAcmeAccountDTO
	for _, account := range accounts {
		accountDTOs = append(accountDTOs, response.WebsiteAcmeAccountDTO{
			WebsiteAcmeAccount: account,
		})
	}
	return total, accountDTOs, err
}

func (w WebsiteAcmeAccountService) Create(create request.WebsiteAcmeAccountCreate) (response.WebsiteAcmeAccountDTO, error) {
	exist, _ := websiteAcmeRepo.GetFirst(websiteAcmeRepo.WithEmail(create.Email))
	if exist != nil {
		return response.WebsiteAcmeAccountDTO{}, buserr.New(constant.ErrEmailIsExist)
	}

	if create.Type == "google" && (create.EabKid == "" || create.EabHmacKey == "") {
		return response.WebsiteAcmeAccountDTO{}, buserr.New(constant.ErrEabKidOrEabHmacKeyCannotBlank)
	} else {
		create.EabKid = ""
		create.EabHmacKey = ""
	}

	acmeAccount := &model.WebsiteAcmeAccount{
		Email: create.Email,
		Type:  create.Type,
	}
	client, err := ssl.NewAcmeClient(acmeAccount)
	if err != nil {
		return response.WebsiteAcmeAccountDTO{}, err
	}
	acmeAccount.PrivateKey = string(ssl.GetPrivateKey(client.User.GetPrivateKey()))
	acmeAccount.URL = client.User.Registration.URI

	if err := websiteAcmeRepo.Create(*acmeAccount); err != nil {
		return response.WebsiteAcmeAccountDTO{}, err
	}
	return response.WebsiteAcmeAccountDTO{WebsiteAcmeAccount: *acmeAccount}, nil
}

func (w WebsiteAcmeAccountService) Delete(id uint) error {
	if ssls, _ := websiteSSLRepo.List(websiteSSLRepo.WithByAcmeAccountId(id)); len(ssls) > 0 {
		return buserr.New(constant.ErrAccountCannotDelete)
	}
	return websiteAcmeRepo.DeleteBy(commonRepo.WithByID(id))
}
