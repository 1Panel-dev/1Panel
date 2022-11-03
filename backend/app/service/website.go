package service

import (
	"context"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"reflect"
	"strings"
	"time"
)

type WebsiteService struct {
}

func (w WebsiteService) PageWebSite(req dto.WebSiteReq) (int64, []dto.WebSiteDTO, error) {
	var websiteDTOs []dto.WebSiteDTO
	total, websites, err := websiteRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	for _, web := range websites {
		websiteDTOs = append(websiteDTOs, dto.WebSiteDTO{
			WebSite: web,
		})
	}
	return total, websiteDTOs, nil
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

	if create.AppType == dto.NewApp {
		install, err := ServiceGroupApp.Install(create.AppInstall.Name, create.AppInstall.AppDetailId, create.AppInstall.Params)
		if err != nil {
			return err
		}
		website.AppInstallID = install.ID
	}

	tx, ctx := getTxAndContext()
	if err := websiteRepo.Create(ctx, website); err != nil {
		return err
	}
	var domains []model.WebSiteDomain
	domains = append(domains, model.WebSiteDomain{Domain: website.PrimaryDomain, WebSiteID: website.ID, Port: 80})

	otherDomainArray := strings.Split(create.OtherDomains, "\n")
	for _, domain := range otherDomainArray {
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

	if err := configDefaultNginx(website, domains); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

func (w WebsiteService) DeleteWebSite(req dto.WebSiteDel) error {

	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if err := delNginxConfig(website); err != nil {
		return err
	}
	tx, ctx := getTxAndContext()

	if req.DeleteApp {
		websites, _ := websiteRepo.GetBy(websiteRepo.WithAppInstallId(website.AppInstallID))
		if len(websites) > 1 {
			return errors.New("other website use this app")
		}
		appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !reflect.DeepEqual(model.AppInstall{}, appInstall) {
			if err := deleteAppInstall(ctx, appInstall); err != nil {
				return err
			}
		}
	}
	//TODO 删除备份
	if err := websiteRepo.DeleteBy(ctx, commonRepo.WithByID(req.ID)); err != nil {
		tx.Rollback()
		return err
	}
	if err := websiteDomainRepo.DeleteBy(ctx, websiteDomainRepo.WithWebSiteId(req.ID)); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (w WebsiteService) CreateWebsiteDomain() {

}

func (w WebsiteService) GetWebsiteDomain(websiteId uint) ([]model.WebSiteDomain, error) {
	return websiteDomainRepo.GetBy(websiteDomainRepo.WithWebSiteId(websiteId))
}

func (w WebsiteService) DeleteWebsiteDomain(domainId uint) error {

	webSiteDomain, err := websiteDomainRepo.GetFirst(commonRepo.WithByID(domainId))
	if err != nil {
		return err
	}

	if websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebSiteId(webSiteDomain.WebSiteID)); len(websiteDomains) == 1 {
		return fmt.Errorf("can not delete last domain")
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(webSiteDomain.WebSiteID))
	if err != nil {
		return err
	}
	var ports []int
	if oldDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebSiteId(webSiteDomain.WebSiteID), websiteDomainRepo.WithPort(webSiteDomain.Port)); len(oldDomains) == 1 {
		ports = append(ports, webSiteDomain.Port)
	}

	var domains []string
	if oldDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebSiteId(webSiteDomain.WebSiteID), websiteDomainRepo.WithDomain(webSiteDomain.Domain)); len(oldDomains) == 1 {
		domains = append(domains, webSiteDomain.Domain)
	}
	if len(ports) > 0 || len(domains) > 0 {
		if err := deleteListenAndServerName(website, ports, domains); err != nil {
			return err
		}
	}

	return websiteDomainRepo.DeleteBy(context.TODO(), commonRepo.WithByID(domainId))
}
