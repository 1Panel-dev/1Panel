package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/backend/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/cmd/server/nginx_conf"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteDTO, error)
	GetWebsites() ([]response.WebsiteDTO, error)
	CreateWebsite(create request.WebsiteCreate) error
	OpWebsite(req request.WebsiteOp) error
	GetWebsiteOptions() ([]string, error)
	UpdateWebsite(req request.WebsiteUpdate) error
	DeleteWebsite(req request.WebsiteDelete) error
	GetWebsite(id uint) (response.WebsiteDTO, error)
	CreateWebsiteDomain(create request.WebsiteDomainCreate) (model.WebsiteDomain, error)
	GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error)
	DeleteWebsiteDomain(domainId uint) error
	GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error)
	UpdateNginxConfigByScope(req request.NginxConfigUpdate) error
	GetWebsiteNginxConfig(websiteId uint, configType string) (response.FileInfo, error)
	GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error)
	OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (response.WebsiteHTTPS, error)
	PreInstallCheck(req request.WebsiteInstallCheckReq) ([]response.WebsitePreInstallCheck, error)
	GetWafConfig(req request.WebsiteWafReq) (response.WebsiteWafConfig, error)
	UpdateWafConfig(req request.WebsiteWafUpdate) error
	UpdateNginxConfigFile(req request.WebsiteNginxUpdate) error
	OpWebsiteLog(req request.WebsiteLogReq) (*response.WebsiteLog, error)
	ChangeDefaultServer(id uint) error
	GetPHPConfig(id uint) (*response.PHPConfig, error)
	UpdatePHPConfig(req request.WebsitePHPConfigUpdate) error
	UpdatePHPConfigFile(req request.WebsitePHPFileUpdate) error
	GetRewriteConfig(req request.NginxRewriteReq) (*response.NginxRewriteRes, error)
	UpdateRewriteConfig(req request.NginxRewriteUpdate) error
	UpdateSiteDir(req request.WebsiteUpdateDir) error
	UpdateSitePermission(req request.WebsiteUpdateDirPermission) error
	OperateProxy(req request.WebsiteProxyConfig) (err error)
	GetProxies(id uint) (res []request.WebsiteProxyConfig, err error)
	UpdateProxyFile(req request.NginxProxyUpdate) (err error)
	GetAuthBasics(req request.NginxAuthReq) (res response.NginxAuthRes, err error)
	UpdateAuthBasic(req request.NginxAuthUpdate) (err error)
}

func NewIWebsiteService() IWebsiteService {
	return &WebsiteService{}
}

func (w WebsiteService) PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteDTO, error) {
	var (
		websiteDTOs []response.WebsiteDTO
		opts        []repo.DBOption
	)
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, nil
		}
		return 0, nil, err
	}
	opts = append(opts, commonRepo.WithOrderBy("created_at desc"))
	if req.Name != "" {
		opts = append(opts, websiteRepo.WithDomainLike(req.Name))
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
			appName     string
			runtimeName string
		)
		switch web.Type {
		case constant.Deployment:
			appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(web.AppInstallID))
			if err != nil {
				return 0, nil, err
			}
			appName = appInstall.Name
		case constant.Runtime:
			runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(web.RuntimeID))
			if err != nil {
				return 0, nil, err
			}
			runtimeName = runtime.Name
		}
		sitePath := path.Join(constant.AppInstallDir, constant.AppOpenresty, nginxInstall.Name, "www", "sites", web.Alias)
		websiteDTOs = append(websiteDTOs, response.WebsiteDTO{
			Website:     web,
			AppName:     appName,
			RuntimeName: runtimeName,
			SitePath:    sitePath,
		})
	}
	return total, websiteDTOs, nil
}

func (w WebsiteService) GetWebsites() ([]response.WebsiteDTO, error) {
	var websiteDTOs []response.WebsiteDTO
	websites, err := websiteRepo.List()
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
	if exist, _ := websiteRepo.GetBy(websiteRepo.WithDomain(create.PrimaryDomain)); len(exist) > 0 {
		return buserr.New(constant.ErrDomainIsExist)
	}
	if exist, _ := websiteRepo.GetBy(websiteRepo.WithAlias(create.Alias)); len(exist) > 0 {
		return buserr.New(constant.ErrAliasIsExist)
	}
	if exist, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithDomain(create.PrimaryDomain)); len(exist) > 0 {
		return buserr.New(constant.ErrDomainIsExist)
	}

	defaultDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)
	website := &model.Website{
		PrimaryDomain:  create.PrimaryDomain,
		Type:           create.Type,
		Alias:          create.Alias,
		Remark:         create.Remark,
		Status:         constant.WebRunning,
		ExpireDate:     defaultDate,
		WebsiteGroupID: create.WebsiteGroupID,
		Protocol:       constant.ProtocolHTTP,
		Proxy:          create.Proxy,
		SiteDir:        "/",
		AccessLog:      true,
		ErrorLog:       true,
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
		if runtime.Resource == constant.ResourceAppstore {
			var (
				req          request.AppInstallCreate
				nginxInstall model.AppInstall
				install      *model.AppInstall
			)
			reg, _ := regexp.Compile(`[^a-z0-9_-]+`)
			req.Name = reg.ReplaceAllString(strings.ToLower(create.PrimaryDomain), "")
			req.AppDetailId = create.AppInstall.AppDetailId
			req.Params = create.AppInstall.Params
			req.Params["IMAGE_NAME"] = runtime.Image
			nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
			if err != nil {
				return err
			}
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
	}

	var domains []model.WebsiteDomain
	domains = append(domains, model.WebsiteDomain{Domain: website.PrimaryDomain, Port: 80})
	otherDomainArray := strings.Split(create.OtherDomains, "\n")
	for _, domain := range otherDomainArray {
		if domain == "" {
			continue
		}
		domainModel, err := getDomain(domain)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(domainModel, model.WebsiteDomain{}) {
			continue
		}
		domains = append(domains, domainModel)
	}
	if err = configDefaultNginx(website, domains, appInstall, runtime); err != nil {
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

func (w WebsiteService) GetWebsiteOptions() ([]string, error) {
	webs, err := websiteRepo.GetBy()
	if err != nil {
		return nil, err
	}
	var datas []string
	for _, web := range webs {
		datas = append(datas, web.PrimaryDomain)
	}
	return datas, nil
}

func (w WebsiteService) UpdateWebsite(req request.WebsiteUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	website.PrimaryDomain = req.PrimaryDomain
	website.WebsiteGroupID = req.WebsiteGroupID
	website.Remark = req.Remark
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
	return res, nil
}

func (w WebsiteService) DeleteWebsite(req request.WebsiteDelete) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if err := delNginxConfig(website, req.ForceDelete); err != nil {
		return err
	}

	if checkIsLinkApp(website) && req.DeleteApp {
		appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !reflect.DeepEqual(model.AppInstall{}, appInstall) {
			if err := deleteAppInstall(appInstall, true, req.ForceDelete, true); err != nil && !req.ForceDelete {
				return err
			}
		}
	}

	tx, ctx := helper.GetTxAndContext()
	defer tx.Rollback()
	_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType("website"), commonRepo.WithByName(website.Alias))
	if err := websiteRepo.DeleteBy(ctx, commonRepo.WithByID(req.ID)); err != nil {
		return err
	}
	if err := websiteDomainRepo.DeleteBy(ctx, websiteDomainRepo.WithWebsiteId(req.ID)); err != nil {
		return err
	}
	tx.Commit()

	if req.DeleteBackup {
		localDir, _ := loadLocalDir()
		backupDir := fmt.Sprintf("%s/website/%s", localDir, website.Alias)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		global.LOG.Infof("delete website %s backups successful", website.Alias)
	}
	uploadDir := fmt.Sprintf("%s/1panel/uploads/website/%s", global.CONF.System.BaseDir, website.Alias)
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	return nil
}

func (w WebsiteService) CreateWebsiteDomain(create request.WebsiteDomainCreate) (model.WebsiteDomain, error) {
	var (
		domainModel model.WebsiteDomain
		ports       []int
		domains     []string
	)
	if create.Port != 80 {
		if common.ScanPort(create.Port) {
			return domainModel, buserr.WithDetail(constant.ErrPortInUsed, create.Port, nil)
		}
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(create.WebsiteID))
	if err != nil {
		return domainModel, err
	}
	if oldDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(create.WebsiteID), websiteDomainRepo.WithPort(create.Port)); len(oldDomains) == 0 {
		ports = append(ports, create.Port)
	}
	domains = append(domains, create.Domain)
	if err := addListenAndServerName(website, ports, domains); err != nil {
		return domainModel, err
	}
	domainModel = model.WebsiteDomain{
		Domain:    create.Domain,
		Port:      create.Port,
		WebsiteID: create.WebsiteID,
	}
	if create.Port != 80 {
		go func() {
			_ = OperateFirewallPort(nil, []int{create.Port})
		}()
	}
	return domainModel, websiteDomainRepo.Create(context.TODO(), &domainModel)
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
	res.SSL = websiteSSL
	res.Enable = true
	if website.HttpConfig != "" {
		res.HttpConfig = website.HttpConfig
	} else {
		res.HttpConfig = constant.HTTPToHTTPS
	}
	params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"ssl_protocols", "ssl_ciphers"}, &website)
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
	}
	return res, nil
}

func (w WebsiteService) OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (response.WebsiteHTTPS, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return response.WebsiteHTTPS{}, err
	}
	var (
		res        response.WebsiteHTTPS
		websiteSSL model.WebsiteSSL
	)
	res.Enable = req.Enable
	res.SSLProtocol = req.SSLProtocol
	res.Algorithm = req.Algorithm
	if !req.Enable {
		website.Protocol = constant.ProtocolHTTP
		website.WebsiteSSLID = 0
		if err := deleteListenAndServerName(website, []string{"443", "[::]:443"}, []string{}); err != nil {
			return response.WebsiteHTTPS{}, err
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
		if err := deleteNginxConfig(constant.NginxScopeServer, nginxParams, &website); err != nil {
			return response.WebsiteHTTPS{}, err
		}
		if err := websiteRepo.Save(ctx, &website); err != nil {
			return response.WebsiteHTTPS{}, err
		}
		return res, nil
	}

	if req.Type == constant.SSLExisted {
		websiteSSL, err = websiteSSLRepo.GetFirst(commonRepo.WithByID(req.WebsiteSSLID))
		if err != nil {
			return response.WebsiteHTTPS{}, err
		}
		website.WebsiteSSLID = websiteSSL.ID
		res.SSL = websiteSSL
	}
	if req.Type == constant.SSLManual {
		certBlock, _ := pem.Decode([]byte(req.Certificate))
		cert, err := x509.ParseCertificate(certBlock.Bytes)
		if err != nil {
			return response.WebsiteHTTPS{}, err
		}
		websiteSSL.ExpireDate = cert.NotAfter
		websiteSSL.StartDate = cert.NotBefore
		websiteSSL.Type = cert.Issuer.CommonName
		websiteSSL.Organization = cert.Issuer.Organization[0]
		websiteSSL.PrimaryDomain = cert.Subject.CommonName
		if len(cert.Subject.Names) > 0 {
			var domains []string
			for _, name := range cert.Subject.Names {
				if v, ok := name.Value.(string); ok {
					if v != cert.Subject.CommonName {
						domains = append(domains, v)
					}
				}
			}
			if len(domains) > 0 {
				websiteSSL.Domains = strings.Join(domains, "")
			}
		}
		websiteSSL.Provider = constant.Manual
		websiteSSL.PrivateKey = req.PrivateKey
		websiteSSL.Pem = req.Certificate
		res.SSL = websiteSSL
	}
	website.Protocol = constant.ProtocolHTTPS
	if err := applySSL(website, websiteSSL, req); err != nil {
		return response.WebsiteHTTPS{}, err
	}
	website.HttpConfig = req.HttpConfig

	if websiteSSL.ID == 0 {
		if err := websiteSSLRepo.Create(ctx, &websiteSSL); err != nil {
			return response.WebsiteHTTPS{}, err
		}
		website.WebsiteSSLID = websiteSSL.ID
	}
	if err := websiteRepo.Save(ctx, &website); err != nil {
		return response.WebsiteHTTPS{}, err
	}
	return res, nil
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
	for _, id := range checkIds {
		if err := syncById(id); err != nil {
			return nil, err
		}
	}
	if len(checkIds) > 0 {
		installList, _ := appInstallRepo.ListBy(commonRepo.WithIdsIn(checkIds))
		for _, install := range installList {
			res = append(res, response.WebsitePreInstallCheck{
				Name:    install.Name,
				Status:  install.Status,
				Version: install.Version,
				AppName: install.App.Name,
			})
			if install.Status != constant.StatusRunning {
				showErr = true
			}
		}
	}
	if showErr {
		return res, nil
	} else {
		return nil, nil
	}
}

func (w WebsiteService) GetWafConfig(req request.WebsiteWafReq) (response.WebsiteWafConfig, error) {
	var res response.WebsiteWafConfig
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return res, nil
	}

	res.Enable = true
	if req.Key != "" {
		params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"set"}, &website)
		if err != nil {
			return res, nil
		}
		for _, param := range params {
			if param.Params[0] == req.Key {
				res.Enable = len(param.Params) > 1 && param.Params[1] == "on"
				break
			}
		}
	}

	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return res, nil
	}

	filePath := path.Join(nginxFull.SiteDir, "sites", website.Alias, "waf", "rules", req.Rule+".json")
	content, err := os.ReadFile(filePath)
	if err != nil {
		return res, nil
	}
	res.FilePath = filePath
	res.Content = string(content)

	return res, nil
}

func (w WebsiteService) UpdateWafConfig(req request.WebsiteWafUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	updateValue := "on"
	if !req.Enable {
		updateValue = "off"
	}
	return updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{
		{Name: "set", Params: []string{req.Key, updateValue}},
	}, &website)
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
		fileInfo, err := os.Stat(filePath)
		if err != nil {
			return nil, err
		}
		if fileInfo.Size() > 10*1024*1024 {
			return nil, buserr.New(constant.ErrFileTooLarge)
		}
		fileInfo.Size()
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, err
		}
		res.Content = string(content)
		return res, nil
	case constant.DisableLog:
		key := "access_log"
		switch req.LogType {
		case constant.AccessLog:
			website.AccessLog = false
		case constant.ErrorLog:
			key = "error_log"
			website.ErrorLog = false
		}
		var nginxParams []dto.NginxParam
		nginxParams = append(nginxParams, dto.NginxParam{
			Name: key,
		})

		if err := deleteNginxConfig(constant.NginxScopeServer, nginxParams, &website); err != nil {
			return nil, err
		}
		if err := websiteRepo.Save(context.Background(), &website); err != nil {
			return nil, err
		}
	case constant.EnableLog:
		key := "access_log"
		logPath := path.Join("/www", "sites", website.Alias, "log", req.LogType)
		switch req.LogType {
		case constant.AccessLog:
			website.AccessLog = true
		case constant.ErrorLog:
			key = "error_log"
			website.ErrorLog = true
		}
		if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: key, Params: []string{logPath}}}, &website); err != nil {
			return nil, err
		}
		if err := websiteRepo.Save(context.Background(), &website); err != nil {
			return nil, err
		}
	}
	return res, nil
}

func (w WebsiteService) ChangeDefaultServer(id uint) error {
	defaultWebsite, _ := websiteRepo.GetFirst(websiteRepo.WithDefaultServer())
	if defaultWebsite.ID > 0 {
		if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "listen", Params: []string{"80"}}}, &defaultWebsite); err != nil {
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
		if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "listen", Params: []string{"80", "default_server"}}}, &website); err != nil {
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
		return nil, buserr.WithDetail(constant.ErrFileCanNotRead, "php.ini", nil)
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
	return &response.PHPConfig{Params: params}, nil
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
		return buserr.WithDetail(constant.ErrFileCanNotRead, "php.ini", nil)
	}
	configFile, err := fileOp.OpenFile(phpConfigPath)
	if err != nil {
		return err
	}
	defer configFile.Close()

	contentBytes, err := fileOp.GetContent(phpConfigPath)
	content := string(contentBytes)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, ";") {
			continue
		}
		for key, value := range req.Params {
			pattern := "^" + regexp.QuoteMeta(key) + "\\s*=\\s*.*$"
			if matched, _ := regexp.MatchString(pattern, line); matched {
				lines[i] = key + " = " + value
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
	return nil
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
		siteDir = fmt.Sprintf("%s/%s", siteDir, req.SiteDir)
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
	if website.SiteDir != "/" {
		absoluteIndexPath = path.Join(absoluteIndexPath, website.SiteDir)
	}
	chownCmd := fmt.Sprintf("chown -R %s:%s %s", req.User, req.Group, absoluteIndexPath)
	if cmd.HasNoPasswordSudo() {
		chownCmd = fmt.Sprintf("sudo %s", chownCmd)
	}
	if out, err := cmd.ExecWithTimeOut(chownCmd, 1*time.Second); err != nil {
		if out != "" {
			return errors.New(out)
		}
		return err
	}
	website.User = req.User
	website.Group = req.Group
	return websiteRepo.Save(context.Background(), &website)
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
		config = parser.NewStringParser(string(nginx_conf.Proxy)).Parse()
	case "edit":
		par, err = parser.NewParser(includePath)
		if err != nil {
			return
		}
		config = par.Parse()
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
		config = parser.NewStringParser(string(content)).Parse()
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
	defer func() {
		if err != nil {
			switch req.Operate {
			case "create":

			}
		}
	}()

	params = append(params, dto.NginxParam{Name: "auth_basic", Params: []string{`"Authentication"`}})
	params = append(params, dto.NginxParam{Name: "auth_basic_user_file", Params: []string{authPath}})
	authContent, err = fileOp.GetContent(absoluteAuthPath)
	if err != nil {
		return
	}
	authArray = strings.Split(string(authContent), "\n")
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
		_, err = writer.WriteString(line + "\n")
		if err != nil {
			return
		}
	}
	err = writer.Flush()
	if err != nil {
		return
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
