package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"path"
	"strconv"
)

func (w WebsiteService) CreateWebsiteDomain(create request.WebsiteDomainCreate) ([]model.WebsiteDomain, error) {
	var (
		domainModels []model.WebsiteDomain
		addPorts     []int
	)
	httpPort, httpsPort, err := getAppInstallPort(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(create.WebsiteID))
	if err != nil {
		return nil, err
	}

	domainModels, addPorts, _, err = getWebsiteDomains(create.Domains, httpPort, httpsPort, create.WebsiteID)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = OperateFirewallPort(nil, addPorts)
	}()

	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	wafDataPath := path.Join(nginxInstall.GetPath(), "1pwaf", "data")
	fileOp := files.NewFileOp()
	if fileOp.Stat(wafDataPath) {
		websitesConfigPath := path.Join(wafDataPath, "conf", "sites.json")
		content, err := fileOp.GetContent(websitesConfigPath)
		if err != nil {
			return nil, err
		}
		var websitesArray []request.WafWebsite
		if content != nil {
			if err := json.Unmarshal(content, &websitesArray); err != nil {
				return nil, err
			}
		}
		for index, wafWebsite := range websitesArray {
			if wafWebsite.Key == website.Alias {
				wafSite := request.WafWebsite{
					Key:     website.Alias,
					Domains: wafWebsite.Domains,
					Host:    wafWebsite.Host,
				}
				for _, domain := range domainModels {
					wafSite.Domains = append(wafSite.Domains, domain.Domain)
					wafSite.Host = append(wafSite.Host, domain.Domain+":"+strconv.Itoa(domain.Port))
				}
				if len(wafSite.Host) == 0 {
					wafSite.Host = []string{}
				}
				websitesArray[index] = wafSite
				break
			}
		}
		websitesContent, err := json.Marshal(websitesArray)
		if err != nil {
			return nil, err
		}
		if err := fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, constant.DirPerm); err != nil {
			return nil, err
		}
	}

	if err = addListenAndServerName(website, domainModels); err != nil {
		return nil, err
	}

	return domainModels, websiteDomainRepo.BatchCreate(context.TODO(), domainModels)
}

func (w WebsiteService) GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error) {
	return websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(websiteId))
}

func (w WebsiteService) DeleteWebsiteDomain(domainId uint) error {
	webSiteDomain, err := websiteDomainRepo.GetFirst(repo.WithByID(domainId))
	if err != nil {
		return err
	}

	if websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(webSiteDomain.WebsiteID)); len(websiteDomains) == 1 {
		return fmt.Errorf("can not delete last domain")
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(webSiteDomain.WebsiteID))
	if err != nil {
		return err
	}
	var ports []int
	if oldDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(webSiteDomain.WebsiteID), websiteDomainRepo.WithPort(webSiteDomain.Port)); len(oldDomains) == 1 {
		ports = append(ports, webSiteDomain.Port)
	}

	var domains []string
	if oldDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(webSiteDomain.WebsiteID), websiteDomainRepo.WithDomain(webSiteDomain.Domain)); len(oldDomains) == 1 {
		domains = append(domains, webSiteDomain.Domain)
	}

	if len(ports) > 0 || len(domains) > 0 {
		stringBinds := make([]string, len(ports))
		for i := 0; i < len(ports); i++ {
			stringBinds[i] = strconv.Itoa(ports[i])
		}
		if err := deleteListenAndServerName(website, stringBinds, domains); err != nil {
			return err
		}
	}

	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	wafDataPath := path.Join(nginxInstall.GetPath(), "1pwaf", "data")
	fileOp := files.NewFileOp()
	if fileOp.Stat(wafDataPath) {
		websitesConfigPath := path.Join(wafDataPath, "conf", "sites.json")
		content, err := fileOp.GetContent(websitesConfigPath)
		if err != nil {
			return err
		}
		var websitesArray []request.WafWebsite
		var newWebsitesArray []request.WafWebsite
		if content != nil {
			if err := json.Unmarshal(content, &websitesArray); err != nil {
				return err
			}
		}
		for _, wafWebsite := range websitesArray {
			if wafWebsite.Key == website.Alias {
				wafSite := wafWebsite
				oldDomains := wafSite.Domains
				var newDomains []string
				for _, domain := range oldDomains {
					if domain == webSiteDomain.Domain {
						continue
					}
					newDomains = append(newDomains, domain)
				}
				wafSite.Domains = newDomains
				oldHostArray := wafSite.Host
				var newHostArray []string
				for _, host := range oldHostArray {
					if host == webSiteDomain.Domain+":"+strconv.Itoa(webSiteDomain.Port) {
						continue
					}
					newHostArray = append(newHostArray, host)
				}
				wafSite.Host = newHostArray
				if len(wafSite.Host) == 0 {
					wafSite.Host = []string{}
				}
				newWebsitesArray = append(newWebsitesArray, wafSite)
			} else {
				newWebsitesArray = append(newWebsitesArray, wafWebsite)
			}
		}
		websitesContent, err := json.Marshal(newWebsitesArray)
		if err != nil {
			return err
		}
		if err = fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, constant.DirPerm); err != nil {
			return err
		}
	}

	return websiteDomainRepo.DeleteBy(context.TODO(), repo.WithByID(domainId))
}
