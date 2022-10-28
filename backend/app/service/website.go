package service

import (
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"reflect"
	"time"
)

type WebsiteService struct {
}

func (w WebsiteService) CreateWebsite(create dto.WebSiteCreate) error {

	defaultDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)

	website := &model.WebSite{
		PrimaryDomain:  create.PrimaryDomain,
		Type:           create.Type,
		Alias:          create.Alias,
		Remark:         create.Remark,
		Status:         constant.WebRunning,
		ExpireDate:     defaultDate,
		AppInstallID:   create.AppInstallID,
		WebSiteGroupID: create.WebSiteGroupID,
	}

	tx, ctx := getTxAndContext()
	if err := websiteRepo.Create(ctx, website); err != nil {
		return err
	}
	var domains []model.WebSiteDomain
	domains = append(domains, model.WebSiteDomain{Domain: website.PrimaryDomain, WebSiteID: website.ID, Port: 80})
	for _, domain := range create.Domains {
		domainModel, err := getDomain(domain, website.ID)
		if err != nil {
			tx.Rollback()
			return err
		}
		if reflect.DeepEqual(domainModel, model.WebSiteDomain{}) {
			continue
		}
		domains = append(domains, domainModel)
	}
	if len(domains) > 0 {
		if err := websiteDomainRepo.BatchCreate(ctx, domains); err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := configDefaultNginx(*website, domains); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}
