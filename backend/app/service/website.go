package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/jinzhu/copier"

	"github.com/1Panel-dev/1Panel/backend/i18n"
	"github.com/spf13/afero"

	"github.com/1Panel-dev/1Panel/backend/utils/compose"
	"github.com/1Panel-dev/1Panel/backend/utils/env"

	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/ini.v1"
	"gorm.io/gorm"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/global"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
)

type WebsiteService struct {
}

type IWebsiteService interface {
	PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteRes, error)
	GetWebsites() ([]response.WebsiteDTO, error)
	CreateWebsite(create request.WebsiteCreate) error
	OpWebsite(req request.WebsiteOp) error
	GetWebsiteOptions() ([]response.WebsiteOption, error)
	UpdateWebsite(req request.WebsiteUpdate) error
	DeleteWebsite(req request.WebsiteDelete) error
	GetWebsite(id uint) (response.WebsiteDTO, error)

	CreateWebsiteDomain(create request.WebsiteDomainCreate) ([]model.WebsiteDomain, error)
	GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error)
	DeleteWebsiteDomain(domainId uint) error

	GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error)
	UpdateNginxConfigByScope(req request.NginxConfigUpdate) error
	GetWebsiteNginxConfig(websiteId uint, configType string) (response.FileInfo, error)
	UpdateNginxConfigFile(req request.WebsiteNginxUpdate) error
	GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error)
	OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (*response.WebsiteHTTPS, error)
	OpWebsiteLog(req request.WebsiteLogReq) (*response.WebsiteLog, error)
	ChangeDefaultServer(id uint) error
	PreInstallCheck(req request.WebsiteInstallCheckReq) ([]response.WebsitePreInstallCheck, error)

	GetPHPConfig(id uint) (*response.PHPConfig, error)
	UpdatePHPConfig(req request.WebsitePHPConfigUpdate) error
	UpdatePHPConfigFile(req request.WebsitePHPFileUpdate) error
	ChangePHPVersion(req request.WebsitePHPVersionReq) error

	GetRewriteConfig(req request.NginxRewriteReq) (*response.NginxRewriteRes, error)
	UpdateRewriteConfig(req request.NginxRewriteUpdate) error
	LoadWebsiteDirConfig(req request.WebsiteCommonReq) (*response.WebsiteDirConfig, error)
	UpdateSiteDir(req request.WebsiteUpdateDir) error
	UpdateSitePermission(req request.WebsiteUpdateDirPermission) error

	OperateProxy(req request.WebsiteProxyConfig) (err error)
	GetProxies(id uint) (res []request.WebsiteProxyConfig, err error)
	UpdateProxyFile(req request.NginxProxyUpdate) (err error)
	DeleteProxy(req request.WebsiteProxyDel) (err error)

	GetAuthBasics(req request.NginxAuthReq) (res response.NginxAuthRes, err error)
	UpdateAuthBasic(req request.NginxAuthUpdate) (err error)
	GetAntiLeech(id uint) (*response.NginxAntiLeechRes, error)
	UpdateAntiLeech(req request.NginxAntiLeechUpdate) (err error)
	OperateRedirect(req request.NginxRedirectReq) (err error)
	GetRedirect(id uint) (res []response.NginxRedirectConfig, err error)
	UpdateRedirectFile(req request.NginxRedirectUpdate) (err error)

	UpdateDefaultHtml(req request.WebsiteHtmlUpdate) error
	GetDefaultHtml(resourceType string) (*response.WebsiteHtmlRes, error)
}

func NewIWebsiteService() IWebsiteService {
	return &WebsiteService{}
}

func (w WebsiteService) PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteRes, error) {
	var (
		websiteDTOs []response.WebsiteRes
		opts        []repo.DBOption
	)
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, nil
		}
		return 0, nil, err
	}
	opts = append(opts, commonRepo.WithOrderRuleBy(req.OrderBy, req.Order))
	if req.Name != "" {
		domains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithDomainLike(req.Name))
		if len(domains) > 0 {
			var websiteIds []uint
			for _, domain := range domains {
				websiteIds = append(websiteIds, domain.WebsiteID)
			}
			opts = append(opts, websiteRepo.WithIDs(websiteIds))
		} else {
			opts = append(opts, websiteRepo.WithDomainLike(req.Name))
		}
	}
	if req.WebsiteGroupID != 0 {
		opts = append(opts, websiteRepo.WithGroupID(req.WebsiteGroupID))
	}
	total, websites, err := websiteRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	for _, web := range websites {
		var (
			appName      string
			runtimeName  string
			runtimeType  string
			appInstallID uint
		)
		switch web.Type {
		case constant.Deployment:
			appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(web.AppInstallID))
			if err != nil {
				return 0, nil, err
			}
			appName = appInstall.Name
			appInstallID = appInstall.ID
		case constant.Runtime:
			runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(web.RuntimeID))
			if err != nil {
				return 0, nil, err
			}
			runtimeName = runtime.Name
			runtimeType = runtime.Type
			appInstallID = runtime.ID
		}
		sitePath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", web.Alias)

		websiteDTOs = append(websiteDTOs, response.WebsiteRes{
			ID:            web.ID,
			CreatedAt:     web.CreatedAt,
			Protocol:      web.Protocol,
			PrimaryDomain: web.PrimaryDomain,
			Type:          web.Type,
			Remark:        web.Remark,
			Status:        web.Status,
			Alias:         web.Alias,
			AppName:       appName,
			ExpireDate:    web.ExpireDate,
			SSLExpireDate: web.WebsiteSSL.ExpireDate,
			SSLStatus:     checkSSLStatus(web.WebsiteSSL.ExpireDate),
			RuntimeName:   runtimeName,
			SitePath:      sitePath,
			AppInstallID:  appInstallID,
			RuntimeType:   runtimeType,
		})
	}
	return total, websiteDTOs, nil
}

func (w WebsiteService) GetWebsites() ([]response.WebsiteDTO, error) {
	var websiteDTOs []response.WebsiteDTO
	websites, err := websiteRepo.List(commonRepo.WithOrderRuleBy("primary_domain", "ascending"))
	if err != nil {
		return nil, err
	}
	for _, web := range websites {
		websiteDTOs = append(websiteDTOs, response.WebsiteDTO{
			Website: web,
		})
	}
	return websiteDTOs, nil
}

func (w WebsiteService) CreateWebsite(create request.WebsiteCreate) (err error) {
	alias := create.Alias
	if alias == "default" {
		return buserr.New("ErrDefaultAlias")
	}
	if common.ContainsChinese(alias) {
		alias, err = common.PunycodeEncode(alias)
		if err != nil {
			return
		}
	}
	if exist, _ := websiteRepo.GetBy(websiteRepo.WithAlias(alias)); len(exist) > 0 {
		return buserr.New(constant.ErrAliasIsExist)
	}

	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	defaultHttpPort := nginxInstall.HttpPort

	var (
		otherDomains []model.WebsiteDomain
		domains      []model.WebsiteDomain
	)
	domains, _, _, err = getWebsiteDomains(create.PrimaryDomain, defaultHttpPort, 0)
	if err != nil {
		return err
	}
	otherDomains, _, _, err = getWebsiteDomains(create.OtherDomains, defaultHttpPort, 0)
	if err != nil {
		return err
	}
	domains = append(domains, otherDomains...)

	defaultDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)
	website := &model.Website{
		PrimaryDomain:  create.PrimaryDomain,
		Type:           create.Type,
		Alias:          alias,
		Remark:         create.Remark,
		Status:         constant.WebRunning,
		ExpireDate:     defaultDate,
		WebsiteGroupID: create.WebsiteGroupID,
		Protocol:       constant.ProtocolHTTP,
		Proxy:          create.Proxy,
		SiteDir:        "/",
		AccessLog:      true,
		ErrorLog:       true,
		IPV6:           create.IPV6,
	}

	var (
		appInstall *model.AppInstall
		runtime    *model.Runtime
	)

	defer func() {
		if err != nil {
			if website.AppInstallID > 0 {
				req := request.AppInstalledOperate{
					InstallId:   website.AppInstallID,
					Operate:     constant.Delete,
					ForceDelete: true,
				}
				if err := NewIAppInstalledService().Operate(req); err != nil {
					global.LOG.Errorf(err.Error())
				}
			}
		}
	}()
	var proxy string

	switch create.Type {
	case constant.Deployment:
		if create.AppType == constant.NewApp {
			var (
				req     request.AppInstallCreate
				install *model.AppInstall
			)
			req.Name = create.AppInstall.Name
			req.AppDetailId = create.AppInstall.AppDetailId
			req.Params = create.AppInstall.Params
			req.AppContainerConfig = create.AppInstall.AppContainerConfig
			tx, installCtx := getTxAndContext()
			install, err = NewIAppService().Install(installCtx, req)
			if err != nil {
				tx.Rollback()
				return err
			}
			tx.Commit()
			appInstall = install
			website.AppInstallID = install.ID
			website.Proxy = fmt.Sprintf("127.0.0.1:%d", appInstall.HttpPort)
		} else {
			var install model.AppInstall
			install, err = appInstallRepo.GetFirst(commonRepo.WithByID(create.AppInstallID))
			if err != nil {
				return err
			}
			appInstall = &install
			website.AppInstallID = appInstall.ID
			website.Proxy = fmt.Sprintf("127.0.0.1:%d", appInstall.HttpPort)
		}
	case constant.Runtime:
		runtime, err = runtimeRepo.GetFirst(commonRepo.WithByID(create.RuntimeID))
		if err != nil {
			return err
		}
		website.RuntimeID = runtime.ID
		switch runtime.Type {
		case constant.RuntimePHP:
			if runtime.Resource == constant.ResourceAppstore {
				if !checkImageLike(runtime.Image) {
					return buserr.WithName("ErrImageNotExist", runtime.Name)
				}
				var (
					req     request.AppInstallCreate
					install *model.AppInstall
				)
				reg, _ := regexp.Compile(`[^a-z0-9_-]+`)
				req.Name = reg.ReplaceAllString(strings.ToLower(alias), "")
				req.AppDetailId = create.AppInstall.AppDetailId
				req.Params = create.AppInstall.Params
				req.Params["IMAGE_NAME"] = runtime.Image
				req.AppContainerConfig = create.AppInstall.AppContainerConfig
				req.Params["PANEL_WEBSITE_DIR"] = path.Join(nginxInstall.GetPath(), "/www")
				tx, installCtx := getTxAndContext()
				install, err = NewIAppService().Install(installCtx, req)
				if err != nil {
					tx.Rollback()
					return err
				}
				tx.Commit()
				website.AppInstallID = install.ID
				appInstall = install
				website.Proxy = fmt.Sprintf("127.0.0.1:%d", appInstall.HttpPort)
			} else {
				website.ProxyType = create.ProxyType
				if website.ProxyType == constant.RuntimeProxyUnix {
					proxy = fmt.Sprintf("unix:%s", path.Join("/www/sites", website.Alias, "php-pool", "php-fpm.sock"))
				}
				if website.ProxyType == constant.RuntimeProxyTcp {
					proxy = fmt.Sprintf("127.0.0.1:%d", create.Port)
				}
				website.Proxy = proxy
			}
		case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo, constant.RuntimePython, constant.RuntimeDotNet:
			website.Proxy = fmt.Sprintf("127.0.0.1:%d", runtime.Port)
		}
	}

	if err = configDefaultNginx(website, domains, appInstall, runtime); err != nil {
		return err
	}

	if len(create.FtpUser) != 0 && len(create.FtpPassword) != 0 {
		indexDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "index")
		itemID, err := NewIFtpService().Create(dto.FtpCreate{User: create.FtpUser, Password: create.FtpPassword, Path: indexDir})
		if err != nil {
			global.LOG.Errorf("create ftp for website failed, err: %v", err)
		}
		website.FtpID = itemID
	}

	if err = createWafConfig(website, domains); err != nil {
		return err
	}

	tx, ctx := helper.GetTxAndContext()
	defer tx.Rollback()
	if err = websiteRepo.Create(ctx, website); err != nil {
		return err
	}
	for i := range domains {
		domains[i].WebsiteID = website.ID
	}
	if err = websiteDomainRepo.BatchCreate(ctx, domains); err != nil {
		return err
	}
	tx.Commit()
	return nil
}

func (w WebsiteService) OpWebsite(req request.WebsiteOp) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if err := opWebsite(&website, req.Operate); err != nil {
		return err
	}
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) GetWebsiteOptions() ([]response.WebsiteOption, error) {
	webs, err := websiteRepo.List()
	if err != nil {
		return nil, err
	}
	var datas []response.WebsiteOption
	for _, web := range webs {
		var item response.WebsiteOption
		if err := copier.Copy(&item, &web); err != nil {
			return nil, err
		}
		datas = append(datas, item)
	}
	return datas, nil
}

func (w WebsiteService) UpdateWebsite(req request.WebsiteUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if website.IPV6 != req.IPV6 {
		if err := changeIPV6(website, req.IPV6); err != nil {
			return err
		}
	}
	website.PrimaryDomain = req.PrimaryDomain
	website.WebsiteGroupID = req.WebsiteGroupID
	website.Remark = req.Remark
	website.IPV6 = req.IPV6

	if req.ExpireDate != "" {
		expireDate, err := time.Parse(constant.DateLayout, req.ExpireDate)
		if err != nil {
			return err
		}
		website.ExpireDate = expireDate
	}

	return websiteRepo.Save(context.TODO(), &website)
}

func (w WebsiteService) GetWebsite(id uint) (response.WebsiteDTO, error) {
	var res response.WebsiteDTO
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return res, err
	}
	res.Website = website

	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return res, err
	}
	sitePath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", website.Alias)
	res.ErrorLogPath = path.Join(sitePath, "log", "error.log")
	res.AccessLogPath = path.Join(sitePath, "log", "access.log")
	res.SitePath = sitePath
	res.SiteDir = website.SiteDir
	return res, nil
}

func (w WebsiteService) DeleteWebsite(req request.WebsiteDelete) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if err = delNginxConfig(website, req.ForceDelete); err != nil {
		return err
	}

	if err = delWafConfig(website, req.ForceDelete); err != nil {
		return err
	}

	if checkIsLinkApp(website) && req.DeleteApp {
		appInstall, _ := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if appInstall.ID > 0 {
			if err = deleteAppInstall(appInstall, true, req.ForceDelete, true); err != nil && !req.ForceDelete {
				return err
			}
		}
	}

	tx, ctx := helper.GetTxAndContext()
	defer tx.Rollback()

	go func() {
		_ = NewIBackupService().DeleteRecordByName("website", website.PrimaryDomain, website.Alias, req.DeleteBackup)
	}()
	if err := websiteRepo.DeleteBy(ctx, commonRepo.WithByID(req.ID)); err != nil {
		return err
	}
	if err := websiteDomainRepo.DeleteBy(ctx, websiteDomainRepo.WithWebsiteId(req.ID)); err != nil {
		return err
	}
	tx.Commit()

	uploadDir := path.Join(global.CONF.System.BaseDir, fmt.Sprintf("1panel/uploads/website/%s", website.Alias))
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	return nil
}

func (w WebsiteService) CreateWebsiteDomain(create request.WebsiteDomainCreate) ([]model.WebsiteDomain, error) {
	var (
		domainModels []model.WebsiteDomain
		addPorts     []int
		addDomains   []string
	)
	httpPort, _, err := getAppInstallPort(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(create.WebsiteID))
	if err != nil {
		return nil, err
	}

	domainModels, addPorts, addDomains, err = getWebsiteDomains(create.Domains, httpPort, create.WebsiteID)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = OperateFirewallPort(nil, addPorts)
	}()

	if err := addListenAndServerName(website, addPorts, addDomains); err != nil {
		return nil, err
	}

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
		if err := fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, 0644); err != nil {
			return nil, err
		}
	}

	return domainModels, websiteDomainRepo.BatchCreate(context.TODO(), domainModels)
}

func (w WebsiteService) GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error) {
	return websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(websiteId))
}

func (w WebsiteService) DeleteWebsiteDomain(domainId uint) error {
	webSiteDomain, err := websiteDomainRepo.GetFirst(commonRepo.WithByID(domainId))
	if err != nil {
		return err
	}

	if websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(webSiteDomain.WebsiteID)); len(websiteDomains) == 1 {
		return fmt.Errorf("can not delete last domain")
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(webSiteDomain.WebsiteID))
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
				removed := false
				for _, domain := range oldDomains {
					if domain == webSiteDomain.Domain && !removed {
						removed = true
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
		if err = fileOp.SaveFileWithByte(websitesConfigPath, websitesContent, 0644); err != nil {
			return err
		}
	}

	return websiteDomainRepo.DeleteBy(context.TODO(), commonRepo.WithByID(domainId))
}

func (w WebsiteService) GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error) {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil, nil
	}

	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return nil, err
	}
	var config response.WebsiteNginxConfig
	params, err := getNginxParamsByKeys(constant.NginxScopeServer, keys, &website)
	if err != nil {
		return nil, err
	}
	config.Params = params
	config.Enable = len(params[0].Params) > 0

	return &config, nil
}

func (w WebsiteService) UpdateNginxConfigByScope(req request.NginxConfigUpdate) error {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if req.Operate == constant.ConfigDel {
		var nginxParams []dto.NginxParam
		for _, key := range keys {
			nginxParams = append(nginxParams, dto.NginxParam{
				Name: key,
			})
		}
		return deleteNginxConfig(constant.NginxScopeServer, nginxParams, &website)
	}
	params := getNginxParams(req.Params, keys)
	if req.Operate == constant.ConfigNew {
		if _, ok := dto.StaticFileKeyMap[req.Scope]; ok {
			params = getNginxParamsFromStaticFile(req.Scope, params)
		}
	}
	return updateNginxConfig(constant.NginxScopeServer, params, &website)
}

func (w WebsiteService) GetWebsiteNginxConfig(websiteId uint, configType string) (response.FileInfo, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteId))
	if err != nil {
		return response.FileInfo{}, err
	}
	configPath := ""
	switch configType {
	case constant.AppOpenresty:
		nginxApp, err := appRepo.GetFirst(appRepo.WithKey(constant.AppOpenresty))
		if err != nil {
			return response.FileInfo{}, err
		}
		nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
		if err != nil {
			return response.FileInfo{}, err
		}
		configPath = path.Join(nginxInstall.GetPath(), "conf", "conf.d", website.Alias+".conf")
	case constant.ConfigFPM:
		runtimeInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return response.FileInfo{}, err
		}
		configPath = path.Join(runtimeInstall.GetPath(), "conf", "php-fpm.conf")
	case constant.ConfigPHP:
		runtimeInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return response.FileInfo{}, err
		}
		configPath = path.Join(runtimeInstall.GetPath(), "conf", "php.ini")
	}
	info, err := files.NewFileInfo(files.FileOption{
		Path:   configPath,
		Expand: true,
	})
	if err != nil {
		return response.FileInfo{}, err
	}
	return response.FileInfo{FileInfo: *info}, nil
}

func (w WebsiteService) GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteId))
	if err != nil {
		return response.WebsiteHTTPS{}, err
	}
	var res response.WebsiteHTTPS
	if website.WebsiteSSLID == 0 {
		res.Enable = false
		return res, nil
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(website.WebsiteSSLID))
	if err != nil {
		return response.WebsiteHTTPS{}, err
	}
	res.SSL = *websiteSSL
	res.Enable = true
	if website.HttpConfig != "" {
		res.HttpConfig = website.HttpConfig
	} else {
		res.HttpConfig = constant.HTTPToHTTPS
	}
	params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"ssl_protocols", "ssl_ciphers", "add_header"}, &website)
	if err != nil {
		return res, err
	}
	for _, p := range params {
		if p.Name == "ssl_protocols" {
			res.SSLProtocol = p.Params
		}
		if p.Name == "ssl_ciphers" {
			res.Algorithm = p.Params[0]
		}
		if p.Name == "add_header" && len(p.Params) > 0 && p.Params[0] == "Strict-Transport-Security" {
			res.Hsts = true
		}
	}
	return res, nil
}

func (w WebsiteService) OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (*response.WebsiteHTTPS, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return nil, err
	}
	var (
		res        response.WebsiteHTTPS
		websiteSSL model.WebsiteSSL
	)
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	if err = ChangeHSTSConfig(req.Hsts, nginxInstall, website); err != nil {
		return nil, err
	}
	res.Enable = req.Enable
	res.SSLProtocol = req.SSLProtocol
	res.Algorithm = req.Algorithm
	if !req.Enable {
		website.Protocol = constant.ProtocolHTTP
		website.WebsiteSSLID = 0
		_, httpsPort, err := getAppInstallPort(constant.AppOpenresty)
		if err != nil {
			return nil, err
		}
		httpsPortStr := strconv.Itoa(httpsPort)
		if err := deleteListenAndServerName(website, []string{httpsPortStr, "[::]:" + httpsPortStr}, []string{}); err != nil {
			return nil, err
		}
		nginxParams := getNginxParamsFromStaticFile(dto.SSL, nil)
		nginxParams = append(nginxParams,
			dto.NginxParam{
				Name:   "if",
				Params: []string{"($scheme", "=", "http)"},
			},
			dto.NginxParam{
				Name: "ssl_certificate",
			},
			dto.NginxParam{
				Name: "ssl_certificate_key",
			},
			dto.NginxParam{
				Name: "ssl_protocols",
			},
			dto.NginxParam{
				Name: "ssl_ciphers",
			},
		)
		if err = deleteNginxConfig(constant.NginxScopeServer, nginxParams, &website); err != nil {
			return nil, err
		}
		if err = websiteRepo.Save(ctx, &website); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if req.Type == constant.SSLExisted {
		websiteModel, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(req.WebsiteSSLID))
		if err != nil {
			return nil, err
		}
		website.WebsiteSSLID = websiteModel.ID
		res.SSL = *websiteModel
		websiteSSL = *websiteModel
	}
	if req.Type == constant.SSLManual {
		var (
			certificate string
			privateKey  string
		)
		switch req.ImportType {
		case "paste":
			certificate = req.Certificate
			privateKey = req.PrivateKey
		case "local":
			fileOp := files.NewFileOp()
			if !fileOp.Stat(req.PrivateKeyPath) {
				return nil, buserr.New("ErrSSLKeyNotFound")
			}
			if !fileOp.Stat(req.CertificatePath) {
				return nil, buserr.New("ErrSSLCertificateNotFound")
			}
			if content, err := fileOp.GetContent(req.PrivateKeyPath); err != nil {
				return nil, err
			} else {
				privateKey = string(content)
			}
			if content, err := fileOp.GetContent(req.CertificatePath); err != nil {
				return nil, err
			} else {
				certificate = string(content)
			}
		}

		privateKeyCertBlock, _ := pem.Decode([]byte(privateKey))
		if privateKeyCertBlock == nil {
			return nil, buserr.New("ErrSSLKeyFormat")
		}

		certBlock, _ := pem.Decode([]byte(certificate))
		if certBlock == nil {
			return nil, buserr.New("ErrSSLCertificateFormat")
		}
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return nil, err
		}
		websiteSSL.ExpireDate = cert.NotAfter
		websiteSSL.StartDate = cert.NotBefore
		websiteSSL.Type = cert.Issuer.CommonName
		if len(cert.Issuer.Organization) > 0 {
			websiteSSL.Organization = cert.Issuer.Organization[0]
		} else {
			websiteSSL.Organization = cert.Issuer.CommonName
		}
		if len(cert.DNSNames) > 0 {
			websiteSSL.PrimaryDomain = cert.DNSNames[0]
			websiteSSL.Domains = strings.Join(cert.DNSNames, ",")
		}
		websiteSSL.Provider = constant.Manual
		websiteSSL.PrivateKey = privateKey
		websiteSSL.Pem = certificate
		websiteSSL.Status = constant.SSLReady

		res.SSL = websiteSSL
	}

	website.Protocol = constant.ProtocolHTTPS
	if err := applySSL(website, websiteSSL, req); err != nil {
		return nil, err
	}
	website.HttpConfig = req.HttpConfig

	if websiteSSL.ID == 0 {
		if err := websiteSSLRepo.Create(ctx, &websiteSSL); err != nil {
			return nil, err
		}
		website.WebsiteSSLID = websiteSSL.ID
	}
	if err := websiteRepo.Save(ctx, &website); err != nil {
		return nil, err
	}
	return &res, nil
}

func (w WebsiteService) PreInstallCheck(req request.WebsiteInstallCheckReq) ([]response.WebsitePreInstallCheck, error) {
	var (
		res      []response.WebsitePreInstallCheck
		checkIds []uint
		showErr  = false
	)

	app, err := appRepo.GetFirst(appRepo.WithKey(constant.AppOpenresty))
	if err != nil {
		return nil, err
	}
	appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID))
	if reflect.DeepEqual(appInstall, model.AppInstall{}) {
		res = append(res, response.WebsitePreInstallCheck{
			Name:    appInstall.Name,
			AppName: app.Name,
			Status:  buserr.WithDetail(constant.ErrNotInstall, app.Name, nil).Error(),
			Version: appInstall.Version,
		})
		showErr = true
	} else {
		checkIds = append(req.InstallIds, appInstall.ID)
	}
	if len(checkIds) > 0 {
		installList, _ := appInstallRepo.ListBy(commonRepo.WithIdsIn(checkIds))
		for _, install := range installList {
			if err = syncAppInstallStatus(&install, false); err != nil {
				return nil, err
			}
			res = append(res, response.WebsitePreInstallCheck{
				Name:    install.Name,
				Status:  install.Status,
				Version: install.Version,
				AppName: install.App.Name,
			})
			if install.Status != constant.Running {
				showErr = true
			}
		}
	}
	if showErr {
		return res, nil
	}
	return nil, nil
}

func (w WebsiteService) UpdateNginxConfigFile(req request.WebsiteNginxUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return err
	}
	filePath := nginxFull.SiteConfig.FilePath
	if err := files.NewFileOp().WriteFile(filePath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxFull.SiteConfig.OldContent, filePath, nginxFull.Install.ContainerName)
}

func (w WebsiteService) OpWebsiteLog(req request.WebsiteLogReq) (*response.WebsiteLog, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return nil, err
	}
	nginx, err := getNginxFull(&website)
	if err != nil {
		return nil, err
	}
	sitePath := path.Join(nginx.SiteDir, "sites", website.Alias)
	res := &response.WebsiteLog{
		Content: "",
	}
	switch req.Operate {
	case constant.GetLog:
		switch req.LogType {
		case constant.AccessLog:
			res.Enable = website.AccessLog
			if !website.AccessLog {
				return res, nil
			}
		case constant.ErrorLog:
			res.Enable = website.ErrorLog
			if !website.ErrorLog {
				return res, nil
			}
		}
		filePath := path.Join(sitePath, "log", req.LogType)
		lines, end, _, err := files.ReadFileByLine(filePath, req.Page, req.PageSize, false)
		if err != nil {
			return nil, err
		}
		res.End = end
		res.Path = filePath
		res.Content = strings.Join(lines, "\n")
		return res, nil
	case constant.DisableLog:
		params := dto.NginxParam{}
		switch req.LogType {
		case constant.AccessLog:
			params.Name = "access_log"
			params.Params = []string{"off"}
			website.AccessLog = false
		case constant.ErrorLog:
			params.Name = "error_log"
			params.Params = []string{"/dev/null", "crit"}
			website.ErrorLog = false
		}
		var nginxParams []dto.NginxParam
		nginxParams = append(nginxParams, params)

		if err := updateNginxConfig(constant.NginxScopeServer, nginxParams, &website); err != nil {
			return nil, err
		}
		if err := websiteRepo.Save(context.Background(), &website); err != nil {
			return nil, err
		}
	case constant.EnableLog:
		key := "access_log"
		logPath := path.Join("/www", "sites", website.Alias, "log", req.LogType)
		params := []string{logPath}
		switch req.LogType {
		case constant.AccessLog:
			params = append(params, "main")
			website.AccessLog = true
		case constant.ErrorLog:
			key = "error_log"
			website.ErrorLog = true
		}
		if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: key, Params: params}}, &website); err != nil {
			return nil, err
		}
		if err := websiteRepo.Save(context.Background(), &website); err != nil {
			return nil, err
		}
	case constant.DeleteLog:
		logPath := path.Join(nginx.Install.GetPath(), "www", "sites", website.Alias, "log", req.LogType)
		if err := files.NewFileOp().WriteFile(logPath, strings.NewReader(""), 0755); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (w WebsiteService) ChangeDefaultServer(id uint) error {
	defaultWebsite, _ := websiteRepo.GetFirst(websiteRepo.WithDefaultServer())
	if defaultWebsite.ID > 0 {
		params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"listen"}, &defaultWebsite)
		if err != nil {
			return err
		}
		var changeParams []dto.NginxParam
		for _, param := range params {
			paramLen := len(param.Params)
			var newParam []string
			if paramLen > 1 && param.Params[paramLen-1] == components.DefaultServer {
				newParam = param.Params[:paramLen-1]
			}
			changeParams = append(changeParams, dto.NginxParam{
				Name:   param.Name,
				Params: newParam,
			})
		}
		if err := updateNginxConfig(constant.NginxScopeServer, changeParams, &defaultWebsite); err != nil {
			return err
		}
		defaultWebsite.DefaultServer = false
		if err := websiteRepo.Save(context.Background(), &defaultWebsite); err != nil {
			return err
		}
	}
	if id > 0 {
		website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
		if err != nil {
			return err
		}
		params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"listen"}, &website)
		if err != nil {
			return err
		}
		httpPort, httpsPort, err := getAppInstallPort(constant.AppOpenresty)
		if err != nil {
			return err
		}

		var changeParams []dto.NginxParam
		for _, param := range params {
			paramLen := len(param.Params)
			bind := param.Params[0]
			var newParam []string
			if bind == strconv.Itoa(httpPort) || bind == strconv.Itoa(httpsPort) || bind == "[::]:"+strconv.Itoa(httpPort) || bind == "[::]:"+strconv.Itoa(httpsPort) {
				if param.Params[paramLen-1] == components.DefaultServer {
					newParam = param.Params
				} else {
					newParam = append(param.Params, components.DefaultServer)
				}
			}
			changeParams = append(changeParams, dto.NginxParam{
				Name:   param.Name,
				Params: newParam,
			})
		}
		if err := updateNginxConfig(constant.NginxScopeServer, changeParams, &website); err != nil {
			return err
		}
		website.DefaultServer = true
		return websiteRepo.Save(context.Background(), &website)
	}
	return nil
}

func (w WebsiteService) GetPHPConfig(id uint) (*response.PHPConfig, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
	if err != nil {
		return nil, err
	}
	phpConfigPath := path.Join(appInstall.GetPath(), "conf", "php.ini")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(phpConfigPath) {
		return nil, buserr.WithMap("ErrFileNotFound", map[string]interface{}{"name": "php.ini"}, nil)
	}
	params := make(map[string]string)
	configFile, err := fileOp.OpenFile(phpConfigPath)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()
	scanner := bufio.NewScanner(configFile)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, ";") {
			continue
		}
		matches := regexp.MustCompile(`^\s*([a-z_]+)\s*=\s*(.*)$`).FindStringSubmatch(line)
		if len(matches) == 3 {
			params[matches[1]] = matches[2]
		}
	}
	cfg, err := ini.Load(phpConfigPath)
	if err != nil {
		return nil, err
	}
	phpConfig, err := cfg.GetSection("PHP")
	if err != nil {
		return nil, err
	}
	disableFunctionStr := phpConfig.Key("disable_functions").Value()
	res := &response.PHPConfig{Params: params}
	if disableFunctionStr != "" {
		disableFunctions := strings.Split(disableFunctionStr, ",")
		if len(disableFunctions) > 0 {
			res.DisableFunctions = disableFunctions
		}
	}
	uploadMaxSize := phpConfig.Key("upload_max_filesize").Value()
	if uploadMaxSize != "" {
		res.UploadMaxSize = uploadMaxSize
	}
	return res, nil
}

func (w WebsiteService) UpdatePHPConfig(req request.WebsitePHPConfigUpdate) (err error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
	if err != nil {
		return err
	}
	phpConfigPath := path.Join(appInstall.GetPath(), "conf", "php.ini")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(phpConfigPath) {
		return buserr.WithMap("ErrFileNotFound", map[string]interface{}{"name": "php.ini"}, nil)
	}
	configFile, err := fileOp.OpenFile(phpConfigPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	contentBytes, err := fileOp.GetContent(phpConfigPath)
	if err != nil {
		return err
	}

	content := string(contentBytes)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, ";") {
			continue
		}
		switch req.Scope {
		case "params":
			for key, value := range req.Params {
				pattern := "^" + regexp.QuoteMeta(key) + "\\s*=\\s*.*$"
				if matched, _ := regexp.MatchString(pattern, line); matched {
					lines[i] = key + " = " + value
				}
			}
		case "disable_functions":
			pattern := "^" + regexp.QuoteMeta("disable_functions") + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(pattern, line); matched {
				lines[i] = "disable_functions" + " = " + strings.Join(req.DisableFunctions, ",")
				break
			}
		case "upload_max_filesize":
			pattern := "^" + regexp.QuoteMeta("post_max_size") + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(pattern, line); matched {
				lines[i] = "post_max_size" + " = " + req.UploadMaxSize
			}
			patternUpload := "^" + regexp.QuoteMeta("upload_max_filesize") + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(patternUpload, line); matched {
				lines[i] = "upload_max_filesize" + " = " + req.UploadMaxSize
			}
		}
	}
	updatedContent := strings.Join(lines, "\n")
	if err := fileOp.WriteFile(phpConfigPath, strings.NewReader(updatedContent), 0755); err != nil {
		return err
	}

	appInstallReq := request.AppInstalledOperate{
		InstallId: appInstall.ID,
		Operate:   constant.Restart,
	}
	if err = NewIAppInstalledService().Operate(appInstallReq); err != nil {
		_ = fileOp.WriteFile(phpConfigPath, strings.NewReader(string(contentBytes)), 0755)
		return err
	}

	return nil
}

func (w WebsiteService) UpdatePHPConfigFile(req request.WebsitePHPFileUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if website.Type != constant.Runtime {
		return nil
	}
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
	if err != nil {
		return err
	}
	if runtime.Resource != constant.ResourceAppstore {
		return nil
	}
	runtimeInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
	if err != nil {
		return err
	}
	configPath := ""
	if req.Type == constant.ConfigFPM {
		configPath = path.Join(runtimeInstall.GetPath(), "conf", "php-fpm.conf")
	} else {
		configPath = path.Join(runtimeInstall.GetPath(), "conf", "php.ini")
	}
	if err := files.NewFileOp().WriteFile(configPath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}
	if _, err := compose.Restart(runtimeInstall.GetComposePath()); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) ChangePHPVersion(req request.WebsitePHPVersionReq) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.RuntimeID))
	if err != nil {
		return err
	}
	oldRuntime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
	if err != nil {
		return err
	}
	if runtime.Resource == constant.ResourceLocal || oldRuntime.Resource == constant.ResourceLocal {
		return buserr.New("ErrPHPResource")
	}
	client, err := docker.NewDockerClient()
	if err != nil {
		return err
	}
	defer client.Close()
	if !checkImageExist(client, runtime.Image) {
		return buserr.WithName("ErrImageNotExist", runtime.Name)
	}
	appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
	if err != nil {
		return err
	}
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(runtime.AppDetailID))
	if err != nil {
		return err
	}

	envs := make(map[string]interface{})
	if err = json.Unmarshal([]byte(appInstall.Env), &envs); err != nil {
		return err
	}
	if out, err := compose.Down(appInstall.GetComposePath()); err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}

	var (
		busErr          error
		fileOp          = files.NewFileOp()
		envPath         = appInstall.GetEnvPath()
		composePath     = appInstall.GetComposePath()
		confDir         = path.Join(appInstall.GetPath(), "conf")
		backupConfDir   = path.Join(appInstall.GetPath(), "conf_bak")
		fpmConfDir      = path.Join(confDir, "php-fpm.conf")
		phpDir          = path.Join(constant.RuntimeDir, runtime.Type, runtime.Name, "php")
		oldFmContent, _ = fileOp.GetContent(fpmConfDir)
		newComposeByte  []byte
	)
	envParams := make(map[string]string, len(envs))
	handleMap(envs, envParams)
	envParams["IMAGE_NAME"] = runtime.Image
	defer func() {
		if busErr != nil {
			envParams["IMAGE_NAME"] = oldRuntime.Image
			_ = env.Write(envParams, envPath)
			_ = fileOp.WriteFile(composePath, strings.NewReader(appInstall.DockerCompose), 0775)
			if fileOp.Stat(backupConfDir) {
				_ = fileOp.DeleteDir(confDir)
				_ = fileOp.Rename(backupConfDir, confDir)
			}
		}
	}()

	if busErr = env.Write(envParams, envPath); busErr != nil {
		return busErr
	}

	newComposeByte, busErr = changeServiceName(appDetail.DockerCompose, appInstall.ServiceName)
	if busErr != nil {
		return err
	}

	if busErr = fileOp.WriteFile(composePath, bytes.NewReader(newComposeByte), 0775); busErr != nil {
		return busErr
	}
	if !req.RetainConfig {
		if busErr = fileOp.Rename(confDir, backupConfDir); busErr != nil {
			return busErr
		}
		_ = fileOp.CreateDir(confDir, 0755)
		if busErr = fileOp.CopyFile(path.Join(phpDir, "php-fpm.conf"), confDir); busErr != nil {
			return busErr
		}
		if busErr = fileOp.CopyFile(path.Join(phpDir, "php.ini"), confDir); busErr != nil {
			_ = fileOp.WriteFile(fpmConfDir, bytes.NewReader(oldFmContent), 0775)
			return busErr
		}
	}

	if out, err := compose.Up(appInstall.GetComposePath()); err != nil {
		if out != "" {
			busErr = errors.New(out)
			return busErr
		}
		busErr = err
		return busErr
	}

	_ = fileOp.DeleteDir(backupConfDir)

	appInstall.AppDetailId = runtime.AppDetailID
	appInstall.AppId = appDetail.AppId
	appInstall.Version = appDetail.Version
	appInstall.DockerCompose = string(newComposeByte)

	_ = appInstallRepo.Save(context.Background(), &appInstall)
	website.RuntimeID = req.RuntimeID
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) UpdateRewriteConfig(req request.NginxRewriteUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return err
	}
	includePath := fmt.Sprintf("/www/sites/%s/rewrite/%s.conf", website.Alias, website.PrimaryDomain)
	absolutePath := path.Join(nginxFull.Install.GetPath(), includePath)
	fileOp := files.NewFileOp()
	var oldRewriteContent []byte
	if !fileOp.Stat(path.Dir(absolutePath)) {
		if err := fileOp.CreateDir(path.Dir(absolutePath), 0755); err != nil {
			return err
		}
	}
	if !fileOp.Stat(absolutePath) {
		if err := fileOp.CreateFile(absolutePath); err != nil {
			return err
		}
	} else {
		oldRewriteContent, err = fileOp.GetContent(absolutePath)
		if err != nil {
			return err
		}
	}
	if err := fileOp.WriteFile(absolutePath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}

	if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{includePath}}}, &website); err != nil {
		_ = fileOp.WriteFile(absolutePath, bytes.NewReader(oldRewriteContent), 0755)
		return err
	}
	website.Rewrite = req.Name
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) GetRewriteConfig(req request.NginxRewriteReq) (*response.NginxRewriteRes, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return nil, err
	}
	var contentByte []byte
	if req.Name == "current" {
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return nil, err
		}
		rewriteConfPath := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "rewrite", fmt.Sprintf("%s.conf", website.PrimaryDomain))
		fileOp := files.NewFileOp()
		if fileOp.Stat(rewriteConfPath) {
			contentByte, err = fileOp.GetContent(rewriteConfPath)
			if err != nil {
				return nil, err
			}
		}
	} else {
		rewriteFile := fmt.Sprintf("rewrite/%s.conf", strings.ToLower(req.Name))
		contentByte, err = nginx_conf.Rewrites.ReadFile(rewriteFile)
		if err != nil {
			return nil, err
		}
	}
	return &response.NginxRewriteRes{
		Content: string(contentByte),
	}, err
}

func (w WebsiteService) UpdateSiteDir(req request.WebsiteUpdateDir) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	runDir := req.SiteDir
	siteDir := path.Join("/www/sites", website.Alias, "index")
	if req.SiteDir != "/" {
		siteDir = fmt.Sprintf("%s%s", siteDir, req.SiteDir)
	}
	if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "root", Params: []string{siteDir}}}, &website); err != nil {
		return err
	}
	website.SiteDir = runDir
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) UpdateSitePermission(req request.WebsiteUpdateDirPermission) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	absoluteIndexPath := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "index")
	chownCmd := fmt.Sprintf("chown -R %s:%s %s", req.User, req.Group, absoluteIndexPath)
	if cmd.HasNoPasswordSudo() {
		chownCmd = fmt.Sprintf("sudo %s", chownCmd)
	}
	if out, err := cmd.ExecWithTimeOut(chownCmd, 10*time.Second); err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	website.User = req.User
	website.Group = req.Group
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) DeleteProxy(req request.WebsiteProxyDel) (err error) {
	fileOp := files.NewFileOp()
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "proxy")
	if !fileOp.Stat(includeDir) {
		_ = fileOp.CreateDir(includeDir, 0755)
	}
	fileName := fmt.Sprintf("%s.conf", req.Name)
	includePath := path.Join(includeDir, fileName)
	backName := fmt.Sprintf("%s.bak", req.Name)
	backPath := path.Join(includeDir, backName)
	_ = fileOp.DeleteFile(includePath)
	_ = fileOp.DeleteFile(backPath)
	return updateNginxConfig(constant.NginxScopeServer, nil, &website)
}

func (w WebsiteService) OperateProxy(req request.WebsiteProxyConfig) (err error) {
	var (
		website      model.Website
		params       []response.NginxParam
		nginxInstall model.AppInstall
		par          *parser.Parser
		oldContent   []byte
	)

	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return
	}
	params, err = getNginxParamsByKeys(constant.NginxScopeHttp, []string{"proxy_cache"}, &website)
	if err != nil {
		return
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	if len(params) == 0 || len(params[0].Params) == 0 {
		commonDir := path.Join(nginxInstall.GetPath(), "www", "common", "proxy")
		proxyTempPath := path.Join(commonDir, "proxy_temp_dir")
		if !fileOp.Stat(proxyTempPath) {
			_ = fileOp.CreateDir(proxyTempPath, 0755)
		}
		proxyCacheDir := path.Join(commonDir, "proxy_temp_dir")
		if !fileOp.Stat(proxyCacheDir) {
			_ = fileOp.CreateDir(proxyCacheDir, 0755)
		}
		nginxParams := getNginxParamsFromStaticFile(dto.CACHE, nil)
		if err = updateNginxConfig(constant.NginxScopeHttp, nginxParams, &website); err != nil {
			return
		}
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "proxy")
	if !fileOp.Stat(includeDir) {
		_ = fileOp.CreateDir(includeDir, 0755)
	}
	fileName := fmt.Sprintf("%s.conf", req.Name)
	includePath := path.Join(includeDir, fileName)
	backName := fmt.Sprintf("%s.bak", req.Name)
	backPath := path.Join(includeDir, backName)

	if req.Operate == "create" && (fileOp.Stat(includePath) || fileOp.Stat(backPath)) {
		err = buserr.New(constant.ErrNameIsExist)
		return
	}

	defer func() {
		if err != nil {
			switch req.Operate {
			case "create":
				_ = fileOp.DeleteFile(includePath)
			case "edit":
				_ = fileOp.WriteFile(includePath, bytes.NewReader(oldContent), 0755)
			}
		}
	}()

	var config *components.Config

	switch req.Operate {
	case "create":
		config, err = parser.NewStringParser(string(nginx_conf.Proxy)).Parse()
		if err != nil {
			return
		}
	case "edit":
		par, err = parser.NewParser(includePath)
		if err != nil {
			return
		}
		config, err = par.Parse()
		if err != nil {
			return
		}
		oldContent, err = fileOp.GetContent(includePath)
		if err != nil {
			return
		}
	case "delete":
		_ = fileOp.DeleteFile(includePath)
		_ = fileOp.DeleteFile(backPath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	case "disable":
		_ = fileOp.Rename(includePath, backPath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	case "enable":
		_ = fileOp.Rename(backPath, includePath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	}

	config.FilePath = includePath
	directives := config.Directives
	location, ok := directives[0].(*components.Location)
	if !ok {
		err = errors.New("error")
		return
	}
	location.UpdateDirective("proxy_pass", []string{req.ProxyPass})
	location.UpdateDirective("proxy_set_header", []string{"Host", req.ProxyHost})
	location.ChangePath(req.Modifier, req.Match)
	if req.Cache {
		location.AddCache(req.CacheTime, req.CacheUnit)
	} else {
		location.RemoveCache()
	}
	if len(req.Replaces) > 0 {
		location.AddSubFilter(req.Replaces)
	} else {
		location.RemoveSubFilter()
	}
	if req.SNI {
		location.UpdateDirective("proxy_ssl_server_name", []string{"on"})
		if req.ProxySSLName != "" {
			location.UpdateDirective("proxy_ssl_name", []string{req.ProxySSLName})
		}
	} else {
		location.UpdateDirective("proxy_ssl_server_name", []string{"off"})
	}
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return
	}
	return
}

func (w WebsiteService) GetProxies(id uint) (res []request.WebsiteProxyConfig, err error) {
	var (
		website      model.Website
		nginxInstall model.AppInstall
		fileList     response.FileInfo
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "proxy")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		return
	}
	fileList, err = NewIFileService().GetFileList(request.FileOption{FileOption: files.FileOption{Path: includeDir, Expand: true, Page: 1, PageSize: 100}})
	if len(fileList.Items) == 0 {
		return
	}
	var (
		content []byte
		config  *components.Config
	)
	for _, configFile := range fileList.Items {
		proxyConfig := request.WebsiteProxyConfig{
			ID: website.ID,
		}
		parts := strings.Split(configFile.Name, ".")
		proxyConfig.Name = parts[0]
		if parts[1] == "conf" {
			proxyConfig.Enable = true
		} else {
			proxyConfig.Enable = false
		}
		proxyConfig.FilePath = configFile.Path
		content, err = fileOp.GetContent(configFile.Path)
		if err != nil {
			return
		}
		proxyConfig.Content = string(content)
		config, err = parser.NewStringParser(string(content)).Parse()
		if err != nil {
			return nil, err
		}
		directives := config.GetDirectives()

		location, ok := directives[0].(*components.Location)
		if !ok {
			err = errors.New("error")
			return
		}
		proxyConfig.ProxyPass = location.ProxyPass
		proxyConfig.Cache = location.Cache
		if location.CacheTime > 0 {
			proxyConfig.CacheTime = location.CacheTime
			proxyConfig.CacheUnit = location.CacheUint
		}
		proxyConfig.Match = location.Match
		proxyConfig.Modifier = location.Modifier
		proxyConfig.ProxyHost = location.Host
		proxyConfig.Replaces = location.Replaces
		for _, directive := range location.Directives {
			if directive.GetName() == "proxy_ssl_server_name" {
				proxyConfig.SNI = directive.GetParameters()[0] == "on"
			}
			if directive.GetName() == "proxy_ssl_name" {
				proxyConfig.ProxySSLName = directive.GetParameters()[0]
			}
		}
		res = append(res, proxyConfig)
	}
	return
}

func (w WebsiteService) UpdateProxyFile(req request.NginxProxyUpdate) (err error) {
	var (
		website           model.Website
		nginxFull         dto.NginxFull
		oldRewriteContent []byte
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxFull, err = getNginxFull(&website)
	if err != nil {
		return err
	}
	includePath := fmt.Sprintf("/www/sites/%s/proxy/%s.conf", website.Alias, req.Name)
	absolutePath := path.Join(nginxFull.Install.GetPath(), includePath)
	fileOp := files.NewFileOp()
	oldRewriteContent, err = fileOp.GetContent(absolutePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(absolutePath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = fileOp.WriteFile(absolutePath, bytes.NewReader(oldRewriteContent), 0755)
		}
	}()
	return updateNginxConfig(constant.NginxScopeServer, nil, &website)
}

func (w WebsiteService) UpdateAuthBasic(req request.NginxAuthUpdate) (err error) {
	var (
		website      model.Website
		nginxInstall model.AppInstall
		params       []dto.NginxParam
		authContent  []byte
		authArray    []string
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	authPath := fmt.Sprintf("/www/sites/%s/auth_basic/auth.pass", website.Alias)
	absoluteAuthPath := path.Join(nginxInstall.GetPath(), authPath)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(path.Dir(absoluteAuthPath)) {
		_ = fileOp.CreateDir(path.Dir(absoluteAuthPath), 0755)
	}
	if !fileOp.Stat(absoluteAuthPath) {
		_ = fileOp.CreateFile(absoluteAuthPath)
	}

	params = append(params, dto.NginxParam{Name: "auth_basic", Params: []string{`"Authentication"`}})
	params = append(params, dto.NginxParam{Name: "auth_basic_user_file", Params: []string{authPath}})
	authContent, err = fileOp.GetContent(absoluteAuthPath)
	if err != nil {
		return
	}
	if len(authContent) > 0 {
		authArray = strings.Split(string(authContent), "\n")
	}
	switch req.Operate {
	case "disable":
		return deleteNginxConfig(constant.NginxScopeServer, params, &website)
	case "enable":
		return updateNginxConfig(constant.NginxScopeServer, params, &website)
	case "create":
		for _, line := range authArray {
			authParams := strings.Split(line, ":")
			username := authParams[0]
			if username == req.Username {
				err = buserr.New(constant.ErrUsernameIsExist)
				return
			}
		}
		var passwdHash []byte
		passwdHash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return
		}
		line := fmt.Sprintf("%s:%s\n", req.Username, passwdHash)
		if req.Remark != "" {
			line = fmt.Sprintf("%s:%s:%s\n", req.Username, passwdHash, req.Remark)
		}
		authArray = append(authArray, line)
	case "edit":
		userExist := false
		for index, line := range authArray {
			authParams := strings.Split(line, ":")
			username := authParams[0]
			if username == req.Username {
				userExist = true
				var passwdHash []byte
				passwdHash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
				if err != nil {
					return
				}
				userPasswd := fmt.Sprintf("%s:%s\n", req.Username, passwdHash)
				if req.Remark != "" {
					userPasswd = fmt.Sprintf("%s:%s:%s\n", req.Username, passwdHash, req.Remark)
				}
				authArray[index] = userPasswd
			}
		}
		if !userExist {
			err = buserr.New(constant.ErrUsernameIsNotExist)
			return
		}
	case "delete":
		deleteIndex := -1
		for index, line := range authArray {
			authParams := strings.Split(line, ":")
			username := authParams[0]
			if username == req.Username {
				deleteIndex = index
			}
		}
		if deleteIndex < 0 {
			return
		}
		authArray = append(authArray[:deleteIndex], authArray[deleteIndex+1:]...)
	}

	var passFile *os.File
	passFile, err = os.Create(absoluteAuthPath)
	if err != nil {
		return
	}
	defer passFile.Close()
	writer := bufio.NewWriter(passFile)
	for _, line := range authArray {
		if line == "" {
			continue
		}
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		return
	}
	authContent, err = fileOp.GetContent(absoluteAuthPath)
	if err != nil {
		return
	}
	if len(authContent) == 0 {
		if err = deleteNginxConfig(constant.NginxScopeServer, params, &website); err != nil {
			return
		}
	}
	return
}

func (w WebsiteService) GetAuthBasics(req request.NginxAuthReq) (res response.NginxAuthRes, err error) {
	var (
		website      model.Website
		nginxInstall model.AppInstall
		authContent  []byte
		nginxParams  []response.NginxParam
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	authPath := fmt.Sprintf("/www/sites/%s/auth_basic/auth.pass", website.Alias)
	absoluteAuthPath := path.Join(nginxInstall.GetPath(), authPath)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(absoluteAuthPath) {
		return
	}
	nginxParams, err = getNginxParamsByKeys(constant.NginxScopeServer, []string{"auth_basic"}, &website)
	if err != nil {
		return
	}
	res.Enable = len(nginxParams[0].Params) > 0
	authContent, err = fileOp.GetContent(absoluteAuthPath)
	authArray := strings.Split(string(authContent), "\n")
	for _, line := range authArray {
		if line == "" {
			continue
		}
		params := strings.Split(line, ":")
		auth := dto.NginxAuth{
			Username: params[0],
		}
		if len(params) == 3 {
			auth.Remark = params[2]
		}
		res.Items = append(res.Items, auth)
	}
	return
}

func (w WebsiteService) UpdateAntiLeech(req request.NginxAntiLeechUpdate) (err error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	backupContent, err := fileOp.GetContent(nginxFull.SiteConfig.Config.FilePath)
	if err != nil {
		return
	}
	block := nginxFull.SiteConfig.Config.FindServers()[0]
	locations := block.FindDirectives("location")
	for _, location := range locations {
		loParams := location.GetParameters()
		if len(loParams) > 1 || loParams[0] == "~" {
			extendStr := loParams[1]
			if strings.HasPrefix(extendStr, `.*\.(`) && strings.HasSuffix(extendStr, `)$`) {
				block.RemoveDirective("location", loParams)
			}
		}
	}
	if req.Enable {
		exts := strings.Split(req.Extends, ",")
		newDirective := components.Directive{
			Name:       "location",
			Parameters: []string{"~", fmt.Sprintf(`.*\.(%s)$`, strings.Join(exts, "|"))},
		}

		newBlock := &components.Block{}
		newBlock.Directives = make([]components.IDirective, 0)
		if req.Cache {
			newBlock.Directives = append(newBlock.Directives, &components.Directive{
				Name:       "expires",
				Parameters: []string{strconv.Itoa(req.CacheTime) + req.CacheUint},
			})
		}
		newBlock.Directives = append(newBlock.Directives, &components.Directive{
			Name:       "log_not_found",
			Parameters: []string{"off"},
		})
		validDir := &components.Directive{
			Name:       "valid_referers",
			Parameters: []string{},
		}
		if req.NoneRef {
			validDir.Parameters = append(validDir.Parameters, "none")
		}
		if len(req.ServerNames) > 0 {
			validDir.Parameters = append(validDir.Parameters, strings.Join(req.ServerNames, " "))
		}
		newBlock.Directives = append(newBlock.Directives, validDir)

		ifDir := &components.Directive{
			Name:       "if",
			Parameters: []string{"($invalid_referer)"},
		}
		ifDir.Block = &components.Block{
			Directives: []components.IDirective{
				&components.Directive{
					Name:       "return",
					Parameters: []string{req.Return},
				},
				&components.Directive{
					Name:       "access_log",
					Parameters: []string{"off"},
				},
			},
		}
		newBlock.Directives = append(newBlock.Directives, ifDir)
		newDirective.Block = newBlock
		block.Directives = append(block.Directives, &newDirective)
	}

	if err = nginx.WriteConfig(nginxFull.SiteConfig.Config, nginx.IndentedStyle); err != nil {
		return
	}
	if err = updateNginxConfig(constant.NginxScopeServer, nil, &website); err != nil {
		_ = fileOp.WriteFile(nginxFull.SiteConfig.Config.FilePath, bytes.NewReader(backupContent), 0755)
		return
	}
	return
}

func (w WebsiteService) GetAntiLeech(id uint) (*response.NginxAntiLeechRes, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil, err
	}
	res := &response.NginxAntiLeechRes{
		LogEnable:   true,
		ServerNames: []string{},
	}
	block := nginxFull.SiteConfig.Config.FindServers()[0]
	locations := block.FindDirectives("location")
	for _, location := range locations {
		loParams := location.GetParameters()
		if len(loParams) > 1 || loParams[0] == "~" {
			extendStr := loParams[1]
			if strings.HasPrefix(extendStr, `.*\.(`) && strings.HasSuffix(extendStr, `)$`) {
				str1 := strings.TrimPrefix(extendStr, `.*\.(`)
				str2 := strings.TrimSuffix(str1, ")$")
				res.Extends = strings.Join(strings.Split(str2, "|"), ",")
			}
		}
		lDirectives := location.GetBlock().GetDirectives()
		for _, lDir := range lDirectives {
			if lDir.GetName() == "valid_referers" {
				res.Enable = true
				params := lDir.GetParameters()
				for _, param := range params {
					if param == "none" {
						res.NoneRef = true
						continue
					}
					if param == "blocked" {
						res.Blocked = true
						continue
					}
					if param == "server_names" {
						continue
					}
					res.ServerNames = append(res.ServerNames, param)
				}
			}
			if lDir.GetName() == "if" && lDir.GetParameters()[0] == "($invalid_referer)" {
				directives := lDir.GetBlock().GetDirectives()
				for _, dir := range directives {
					if dir.GetName() == "return" {
						res.Return = strings.Join(dir.GetParameters(), " ")
					}
					if dir.GetName() == "access_log" {
						if strings.Join(dir.GetParameters(), "") == "off" {
							res.LogEnable = false
						}
					}
				}
			}
			if lDir.GetName() == "expires" {
				res.Cache = true
				re := regexp.MustCompile(`^(\d+)(\w+)$`)
				matches := re.FindStringSubmatch(lDir.GetParameters()[0])
				if matches == nil {
					continue
				}
				cacheTime, err := strconv.Atoi(matches[1])
				if err != nil {
					continue
				}
				unit := matches[2]
				res.CacheUint = unit
				res.CacheTime = cacheTime
			}
		}
	}
	return res, nil
}

func (w WebsiteService) OperateRedirect(req request.NginxRedirectReq) (err error) {
	var (
		website      model.Website
		nginxInstall model.AppInstall
		oldContent   []byte
	)

	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "redirect")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		_ = fileOp.CreateDir(includeDir, 0755)
	}
	fileName := fmt.Sprintf("%s.conf", req.Name)
	includePath := path.Join(includeDir, fileName)
	backName := fmt.Sprintf("%s.bak", req.Name)
	backPath := path.Join(includeDir, backName)

	if req.Operate == "create" && (fileOp.Stat(includePath) || fileOp.Stat(backPath)) {
		err = buserr.New(constant.ErrNameIsExist)
		return
	}

	defer func() {
		if err != nil {
			switch req.Operate {
			case "create":
				_ = fileOp.DeleteFile(includePath)
			case "edit":
				_ = fileOp.WriteFile(includePath, bytes.NewReader(oldContent), 0755)
			}
		}
	}()

	var (
		config *components.Config
		oldPar *parser.Parser
	)

	switch req.Operate {
	case "create":
		config = &components.Config{}
	case "edit":
		oldPar, err = parser.NewParser(includePath)
		if err != nil {
			return
		}
		config, err = oldPar.Parse()
		if err != nil {
			return
		}
		oldContent, err = fileOp.GetContent(includePath)
		if err != nil {
			return
		}
	case "delete":
		_ = fileOp.DeleteFile(includePath)
		_ = fileOp.DeleteFile(backPath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	case "disable":
		_ = fileOp.Rename(includePath, backPath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	case "enable":
		_ = fileOp.Rename(backPath, includePath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	}

	target := req.Target
	block := &components.Block{}

	switch req.Type {
	case "path":
		if req.KeepPath {
			target = req.Target + "$1"
		} else {
			target = req.Target + "?"
		}
		redirectKey := "permanent"
		if req.Redirect == "302" {
			redirectKey = "redirect"
		}
		block = &components.Block{
			Directives: []components.IDirective{
				&components.Directive{
					Name:       "rewrite",
					Parameters: []string{fmt.Sprintf("^%s(.*)", req.Path), target, redirectKey},
				},
			},
		}
	case "domain":
		if req.KeepPath {
			target = req.Target + "$request_uri"
		}
		returnBlock := &components.Block{
			Directives: []components.IDirective{
				&components.Directive{
					Name:       "return",
					Parameters: []string{req.Redirect, target},
				},
			},
		}
		for _, domain := range req.Domains {
			block.Directives = append(block.Directives, &components.Directive{
				Name:       "if",
				Parameters: []string{"($host", "~", fmt.Sprintf("'^%s')", domain)},
				Block:      returnBlock,
			})
		}
	case "404":
		if req.RedirectRoot {
			target = "/"
		}
		block = &components.Block{
			Directives: []components.IDirective{
				&components.Directive{
					Name:       "error_page",
					Parameters: []string{"404", "=", "@notfound"},
				},
				&components.Directive{
					Name:       "location",
					Parameters: []string{"@notfound"},
					Block: &components.Block{
						Directives: []components.IDirective{
							&components.Directive{
								Name:       "return",
								Parameters: []string{req.Redirect, target},
							},
						},
					},
				},
			},
		}
	}
	config.FilePath = includePath
	config.Block = block

	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}

	nginxInclude := fmt.Sprintf("/www/sites/%s/redirect/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return
	}
	return
}

func (w WebsiteService) GetRedirect(id uint) (res []response.NginxRedirectConfig, err error) {
	var (
		website      model.Website
		nginxInstall model.AppInstall
		fileList     response.FileInfo
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "redirect")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		return
	}
	fileList, err = NewIFileService().GetFileList(request.FileOption{FileOption: files.FileOption{Path: includeDir, Expand: true, Page: 1, PageSize: 100}})
	if len(fileList.Items) == 0 {
		return
	}
	var (
		content []byte
		config  *components.Config
	)
	for _, configFile := range fileList.Items {
		redirectConfig := response.NginxRedirectConfig{
			WebsiteID: website.ID,
		}
		parts := strings.Split(configFile.Name, ".")
		redirectConfig.Name = parts[0]
		if parts[1] == "conf" {
			redirectConfig.Enable = true
		} else {
			redirectConfig.Enable = false
		}
		redirectConfig.FilePath = configFile.Path
		content, err = fileOp.GetContent(configFile.Path)
		if err != nil {
			return
		}
		redirectConfig.Content = string(content)
		config, err = parser.NewStringParser(string(content)).Parse()
		if err != nil {
			return
		}

		dirs := config.GetDirectives()
		if len(dirs) > 0 {
			firstName := dirs[0].GetName()
			switch firstName {
			case "if":
				for _, ifDir := range dirs {
					params := ifDir.GetParameters()
					if len(params) > 2 && params[0] == "($host" {
						domain := strings.Trim(strings.Trim(params[2], "'"), "^")
						redirectConfig.Domains = append(redirectConfig.Domains, domain)
						if len(redirectConfig.Domains) > 1 {
							continue
						}
						redirectConfig.Type = "domain"
					}
					childDirs := ifDir.GetBlock().GetDirectives()
					for _, dir := range childDirs {
						if dir.GetName() == "return" {
							dirParams := dir.GetParameters()
							if len(dirParams) > 1 {
								redirectConfig.Redirect = dirParams[0]
								if strings.HasSuffix(dirParams[1], "$request_uri") {
									redirectConfig.KeepPath = true
									redirectConfig.Target = strings.TrimSuffix(dirParams[1], "$request_uri")
								} else {
									redirectConfig.KeepPath = false
									redirectConfig.Target = dirParams[1]
								}
							}
						}
					}
				}
			case "rewrite":
				redirectConfig.Type = "path"
				for _, pathDir := range dirs {
					if pathDir.GetName() == "rewrite" {
						params := pathDir.GetParameters()
						if len(params) > 2 {
							redirectConfig.Path = strings.Trim(strings.Trim(params[0], "^"), "(.*)")
							if strings.HasSuffix(params[1], "$1") {
								redirectConfig.KeepPath = true
								redirectConfig.Target = strings.TrimSuffix(params[1], "$1")
							} else {
								redirectConfig.KeepPath = false
								redirectConfig.Target = strings.TrimSuffix(params[1], "?")
							}
							if params[2] == "permanent" {
								redirectConfig.Redirect = "301"
							} else {
								redirectConfig.Redirect = "302"
							}
						}
					}
				}
			case "error_page":
				redirectConfig.Type = "404"
				for _, errDir := range dirs {
					if errDir.GetName() == "location" {
						childDirs := errDir.GetBlock().GetDirectives()
						for _, dir := range childDirs {
							if dir.GetName() == "return" {
								dirParams := dir.GetParameters()
								if len(dirParams) > 1 {
									redirectConfig.Redirect = dirParams[0]
									if strings.HasSuffix(dirParams[1], "$request_uri") {
										redirectConfig.KeepPath = true
										redirectConfig.Target = strings.TrimSuffix(dirParams[1], "$request_uri")
										redirectConfig.RedirectRoot = false
									} else {
										redirectConfig.KeepPath = false
										redirectConfig.Target = dirParams[1]
										redirectConfig.RedirectRoot = redirectConfig.Target == "/"
									}
								}
							}
						}
					}
				}
			}
		}
		res = append(res, redirectConfig)
	}
	return
}

func (w WebsiteService) UpdateRedirectFile(req request.NginxRedirectUpdate) (err error) {
	var (
		website           model.Website
		nginxFull         dto.NginxFull
		oldRewriteContent []byte
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxFull, err = getNginxFull(&website)
	if err != nil {
		return err
	}
	includePath := fmt.Sprintf("/www/sites/%s/redirect/%s.conf", website.Alias, req.Name)
	absolutePath := path.Join(nginxFull.Install.GetPath(), includePath)
	fileOp := files.NewFileOp()
	oldRewriteContent, err = fileOp.GetContent(absolutePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(absolutePath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = fileOp.WriteFile(absolutePath, bytes.NewReader(oldRewriteContent), 0755)
		}
	}()
	return updateNginxConfig(constant.NginxScopeServer, nil, &website)
}

func (w WebsiteService) LoadWebsiteDirConfig(req request.WebsiteCommonReq) (*response.WebsiteDirConfig, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return nil, err
	}
	res := &response.WebsiteDirConfig{}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	absoluteIndexPath := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "index")
	var appFs = afero.NewOsFs()
	info, err := appFs.Stat(absoluteIndexPath)
	if err != nil {
		return nil, err
	}
	res.User = strconv.FormatUint(uint64(info.Sys().(*syscall.Stat_t).Uid), 10)
	res.UserGroup = strconv.FormatUint(uint64(info.Sys().(*syscall.Stat_t).Gid), 10)

	indexFiles, err := os.ReadDir(absoluteIndexPath)
	if err != nil {
		return nil, err
	}
	res.Dirs = []string{"/"}
	for _, file := range indexFiles {
		if !file.IsDir() {
			continue
		}
		res.Dirs = append(res.Dirs, fmt.Sprintf("/%s", file.Name()))
		fileInfo, _ := file.Info()
		if fileInfo.Sys().(*syscall.Stat_t).Uid != 1000 || fileInfo.Sys().(*syscall.Stat_t).Gid != 1000 {
			res.Msg = i18n.GetMsgByKey("ErrPathPermission")
		}
		childFiles, _ := os.ReadDir(absoluteIndexPath + "/" + file.Name())
		for _, childFile := range childFiles {
			if !childFile.IsDir() {
				continue
			}
			childInfo, _ := childFile.Info()
			if childInfo.Sys().(*syscall.Stat_t).Uid != 1000 || childInfo.Sys().(*syscall.Stat_t).Gid != 1000 {
				res.Msg = i18n.GetMsgByKey("ErrPathPermission")
			}
			res.Dirs = append(res.Dirs, fmt.Sprintf("/%s/%s", file.Name(), childFile.Name()))
		}
	}

	return res, nil
}

func (w WebsiteService) GetDefaultHtml(resourceType string) (*response.WebsiteHtmlRes, error) {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	rootPath := path.Join(nginxInstall.GetPath(), "root")
	fileOp := files.NewFileOp()
	defaultPath := path.Join(rootPath, "default")
	if !fileOp.Stat(defaultPath) {
		_ = fileOp.CreateDir(defaultPath, 0755)
	}

	res := &response.WebsiteHtmlRes{}

	switch resourceType {
	case "404":
		resourcePath := path.Join(defaultPath, "404.html")
		if content, _ := getResourceContent(fileOp, resourcePath); content != "" {
			res.Content = content
			return res, nil
		}
		res.Content = string(nginx_conf.NotFoundHTML)
		return res, nil
	case "php":
		resourcePath := path.Join(defaultPath, "index.php")
		if content, _ := getResourceContent(fileOp, resourcePath); content != "" {
			res.Content = content
			return res, nil
		}
		res.Content = string(nginx_conf.IndexPHP)
		return res, nil
	case "index":
		resourcePath := path.Join(defaultPath, "index.html")
		if content, _ := getResourceContent(fileOp, resourcePath); content != "" {
			res.Content = content
			return res, nil
		}
		res.Content = string(nginx_conf.Index)
		return res, nil
	case "domain404":
		resourcePath := path.Join(rootPath, "404.html")
		if content, _ := getResourceContent(fileOp, resourcePath); content != "" {
			res.Content = content
			return res, nil
		}
		res.Content = string(nginx_conf.DomainNotFoundHTML)
		return res, nil
	case "stop":
		resourcePath := path.Join(rootPath, "stop", "index.html")
		if content, _ := getResourceContent(fileOp, resourcePath); content != "" {
			res.Content = content
			return res, nil
		}
		res.Content = string(nginx_conf.StopHTML)
		return res, nil
	}
	return res, nil
}

func (w WebsiteService) UpdateDefaultHtml(req request.WebsiteHtmlUpdate) error {
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	rootPath := path.Join(nginxInstall.GetPath(), "root")
	fileOp := files.NewFileOp()
	defaultPath := path.Join(rootPath, "default")
	if !fileOp.Stat(defaultPath) {
		_ = fileOp.CreateDir(defaultPath, 0755)
	}
	var resourcePath string
	switch req.Type {
	case "404":
		resourcePath = path.Join(defaultPath, "404.html")
	case "php":
		resourcePath = path.Join(defaultPath, "index.php")
	case "index":
		resourcePath = path.Join(defaultPath, "index.html")
	case "domain404":
		resourcePath = path.Join(rootPath, "404.html")
	case "stop":
		resourcePath = path.Join(rootPath, "stop", "index.html")
	default:
		return nil
	}
	return fileOp.SaveFile(resourcePath, req.Content, 0644)
}
