package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"gorm.io/gorm"
)

func (a AgentService) BindWebsite(req dto.AgentWebsiteBindReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return err
	}
	if agent.WebsiteID != 0 {
		return buserr.New("ErrAgentWebsiteBound")
	}

	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if !isBindableAgentWebsiteType(website.Type) {
		return buserr.New("ErrAgentWebsiteTypeUnsupported")
	}

	boundAgent, err := agentRepo.GetFirst(repo.WithByWebsiteID(req.WebsiteID))
	if err == nil && boundAgent.ID > 0 {
		return buserr.New("ErrAgentWebsiteInUse")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	agent.WebsiteID = req.WebsiteID
	if err := agentRepo.Save(agent); err != nil {
		return err
	}
	return ensureOpenclawWebsiteAllowedOrigin(agent, &website)
}

func (a AgentService) UnbindWebsite(req dto.AgentIDReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return err
	}
	if agent.WebsiteID == 0 {
		return nil
	}

	website, err := websiteRepo.GetFirst(repo.WithByID(agent.WebsiteID))
	if err != nil {
		return err
	}
	if website.Type == constant.Deployment {
		return buserr.New("ErrAgentWebsiteUnbindUnsupported")
	}

	agent.WebsiteID = 0
	return agentRepo.Save(agent)
}

func hydrateAgentWebsiteItems(items []dto.AgentItem) error {
	explicitWebsiteMap, err := loadAgentWebsiteMapByID(items)
	if err != nil {
		return err
	}
	websiteDomainMap, err := loadAgentWebsiteDomainMapByWebsiteID(items)
	if err != nil {
		return err
	}
	fillAgentWebsiteItems(items, explicitWebsiteMap, websiteDomainMap)
	return nil
}

func loadAgentWebsiteMapByID(items []dto.AgentItem) (map[uint]model.Website, error) {
	websiteIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.WebsiteID > 0 {
			websiteIDs = append(websiteIDs, item.WebsiteID)
		}
	}
	if len(websiteIDs) == 0 {
		return map[uint]model.Website{}, nil
	}
	websites, err := websiteRepo.GetBy(repo.WithByIDs(uniqueUintList(websiteIDs)))
	if err != nil {
		return nil, err
	}
	websiteMap := make(map[uint]model.Website, len(websites))
	for _, website := range websites {
		websiteMap[website.ID] = website
	}
	return websiteMap, nil
}

func loadAgentWebsiteDomainMapByWebsiteID(items []dto.AgentItem) (map[uint][]model.WebsiteDomain, error) {
	websiteIDs := make([]uint, 0, len(items))
	for _, item := range items {
		if item.WebsiteID > 0 {
			websiteIDs = append(websiteIDs, item.WebsiteID)
		}
	}
	if len(websiteIDs) == 0 {
		return map[uint][]model.WebsiteDomain{}, nil
	}
	websiteDomains, err := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteIds(uniqueUintList(websiteIDs)), repo.WithOrderAsc("id"))
	if err != nil {
		return nil, err
	}
	websiteDomainMap := make(map[uint][]model.WebsiteDomain, len(websiteDomains))
	for _, websiteDomain := range websiteDomains {
		websiteDomainMap[websiteDomain.WebsiteID] = append(websiteDomainMap[websiteDomain.WebsiteID], websiteDomain)
	}
	return websiteDomainMap, nil
}

func loadAgentWebsiteResourceName(website model.Website) (string, error) {
	websiteDomains, err := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID), repo.WithOrderAsc("id"))
	if err != nil {
		return "", err
	}
	if len(websiteDomains) > 0 {
		return websiteDomains[0].Domain, nil
	}
	if website.PrimaryDomain != "" {
		return website.PrimaryDomain, nil
	}
	return fmt.Sprintf("%d", website.ID), nil
}

func fillAgentWebsiteItems(items []dto.AgentItem, explicitWebsiteMap map[uint]model.Website, websiteDomainMap map[uint][]model.WebsiteDomain) {
	for index := range items {
		if items[index].WebsiteID == 0 {
			continue
		}
		website, ok := explicitWebsiteMap[items[index].WebsiteID]
		if !ok {
			items[index].WebsiteID = 0
			items[index].WebsitePrimaryDomain = ""
			items[index].WebsiteType = ""
			items[index].WebsiteProtocol = ""
			continue
		}
		items[index].WebsiteType = website.Type
		items[index].WebsiteProtocol = website.Protocol
		websiteDomains := websiteDomainMap[items[index].WebsiteID]
		if len(websiteDomains) == 0 {
			items[index].WebsitePrimaryDomain = ""
			continue
		}
		items[index].WebsitePrimaryDomain = websiteDomains[0].Domain
	}
}

func uniqueDeploymentWebsiteMapByAppInstall(websites []model.Website) map[uint]model.Website {
	websiteMap := make(map[uint]model.Website)
	duplicateAppInstallIDs := make(map[uint]struct{})
	for _, website := range websites {
		if website.AppInstallID == 0 {
			continue
		}
		if _, duplicated := duplicateAppInstallIDs[website.AppInstallID]; duplicated {
			continue
		}
		if _, exists := websiteMap[website.AppInstallID]; exists {
			delete(websiteMap, website.AppInstallID)
			duplicateAppInstallIDs[website.AppInstallID] = struct{}{}
			continue
		}
		websiteMap[website.AppInstallID] = website
	}
	return websiteMap
}

func UniqueDeploymentWebsiteMapForMigration(websites []model.Website) map[uint]model.Website {
	return uniqueDeploymentWebsiteMapByAppInstall(websites)
}

func uniqueUintList(items []uint) []uint {
	itemMap := make(map[uint]struct{}, len(items))
	uniq := make([]uint, 0, len(items))
	for _, item := range items {
		if item == 0 {
			continue
		}
		if _, exists := itemMap[item]; exists {
			continue
		}
		itemMap[item] = struct{}{}
		uniq = append(uniq, item)
	}
	return uniq
}

func isBindableAgentWebsiteType(websiteType string) bool {
	return websiteType == constant.Proxy || websiteType == constant.Static
}

func bindDeploymentWebsiteToAgentByAppInstall(website *model.Website) error {
	if website.ID == 0 || website.Type != constant.Deployment || website.AppInstallID == 0 {
		return nil
	}

	appInstall, err := appInstallRepo.GetFirst(repo.WithByID(website.AppInstallID))
	if err != nil {
		return err
	}
	if appInstall.App.Key != constant.AppOpenclaw && appInstall.App.Key != constant.AppCopaw && appInstall.App.Key != constant.AppHermesAgent {
		return nil
	}

	agent, err := agentRepo.GetFirst(repo.WithByAppInstallID(website.AppInstallID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if agent.WebsiteID != 0 {
		return nil
	}

	agent.WebsiteID = website.ID
	if err := agentRepo.Save(agent); err != nil {
		return err
	}
	return ensureOpenclawWebsiteAllowedOrigin(agent, website)
}

func ensureOpenclawWebsiteAllowedOrigin(agent *model.Agent, website *model.Website) error {
	if agent == nil || website == nil || agent.AgentType != constant.AppOpenclaw {
		return nil
	}

	origins, err := buildWebsiteAllowedOrigins(website)
	if err != nil {
		return err
	}
	if len(origins) == 0 {
		return nil
	}

	install, err := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
	if err != nil {
		return err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}

	allowedOrigins := extractSecurityConfig(conf).AllowedOrigins
	allowedOrigins, err = normalizeAllowedOrigins(append(allowedOrigins, origins...))
	if err != nil {
		return err
	}
	setSecurityConfig(conf, dto.AgentSecurityConfig{AllowedOrigins: allowedOrigins})
	if err := writeOpenclawConfigRaw(agent.ConfigPath, conf); err != nil {
		return err
	}
	if err := syncOpenclawAllowedOriginEnv(&install, allowedOrigins); err != nil {
		return err
	}
	return appInstallRepo.Save(context.Background(), &install)
}

func buildWebsiteAllowedOrigins(website *model.Website) ([]string, error) {
	websiteDomains, err := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID), repo.WithOrderAsc("id"))
	if err != nil {
		return nil, err
	}
	if len(websiteDomains) == 0 {
		return nil, nil
	}
	defaultHTTPSPort := 443
	if strings.EqualFold(website.Protocol, "https") {
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err == nil && nginxInstall.HttpsPort > 0 {
			defaultHTTPSPort = nginxInstall.HttpsPort
		}
	}
	return buildWebsiteAllowedOriginsFromDomains(website, websiteDomains, defaultHTTPSPort)
}

func buildWebsiteAllowedOriginsFromDomains(website *model.Website, websiteDomains []model.WebsiteDomain, defaultHTTPSPort int) ([]string, error) {
	if len(websiteDomains) == 0 {
		return nil, nil
	}
	sort.Slice(websiteDomains, func(i, j int) bool {
		return websiteDomains[i].ID < websiteDomains[j].ID
	})
	origins := make([]string, 0, len(websiteDomains))
	for _, websiteDomain := range websiteDomains {
		origin, err := buildWebsiteDomainOrigin(website.Protocol, websiteDomain, defaultHTTPSPort)
		if err != nil {
			return nil, err
		}
		origins = append(origins, origin)
	}
	return origins, nil
}

func buildWebsiteDomainOrigin(protocol string, websiteDomain model.WebsiteDomain, defaultHTTPSPort int) (string, error) {
	origin := strings.ToLower(strings.TrimSpace(protocol)) + "://" + formatWebsiteDomainHost(websiteDomain.Domain)
	switch strings.ToLower(strings.TrimSpace(protocol)) {
	case "http":
		if websiteDomain.Port > 0 && websiteDomain.Port != 80 {
			origin = fmt.Sprintf("%s:%d", origin, websiteDomain.Port)
		}
	case "https":
		port := websiteDomain.Port
		if !websiteDomain.SSL {
			port = defaultHTTPSPort
		}
		if port > 0 && port != 443 {
			origin = fmt.Sprintf("%s:%d", origin, port)
		}
	}
	return normalizeAllowedOrigin(origin)
}

func formatWebsiteDomainHost(domain string) string {
	host := strings.Trim(strings.TrimSpace(domain), "[]")
	if ip := net.ParseIP(host); ip != nil && strings.Contains(host, ":") {
		return "[" + host + "]"
	}
	return host
}
