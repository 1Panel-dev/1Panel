package service

import (
	"bufio"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/api/v1/helper"
	"os"
	"path"
	"reflect"
	"regexp"
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
	"github.com/pkg/errors"
	"gorm.io/gorm"
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
	DeleteWebsite(ctx context.Context, req request.WebsiteDelete) error
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
}

func NewIWebsiteService() IWebsiteService {
	return &WebsiteService{}
}

func (w WebsiteService) PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteDTO, error) {
	var (
		websiteDTOs []response.WebsiteDTO
		opts        []repo.DBOption
	)
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
		websiteDTOs = append(websiteDTOs, response.WebsiteDTO{
			Website:     web,
			AppName:     appName,
			RuntimeName: runtimeName,
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
				if err := NewIAppInstalledService().Operate(context.Background(), req); err != nil {
					global.LOG.Errorf(err.Error())
				}
			}
		}
	}()

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
			website.AppInstallID = install.ID
			appInstall = install
		} else {
			var install model.AppInstall
			install, err = appInstallRepo.GetFirst(commonRepo.WithByID(create.AppInstallID))
			if err != nil {
				return err
			}
			appInstall = &install
			website.AppInstallID = appInstall.ID
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
		}
	}

	tx, ctx := helper.GetTxAndContext()
	defer tx.Rollback()
	if err = websiteRepo.Create(ctx, website); err != nil {
		return err
	}
	var domains []model.WebsiteDomain
	domains = append(domains, model.WebsiteDomain{Domain: website.PrimaryDomain, WebsiteID: website.ID, Port: 80})

	otherDomainArray := strings.Split(create.OtherDomains, "\n")
	for _, domain := range otherDomainArray {
		if domain == "" {
			continue
		}
		domainModel, err := getDomain(domain, website.ID)
		if err != nil {
			return err
		}
		if reflect.DeepEqual(domainModel, model.WebsiteDomain{}) {
			continue
		}
		domains = append(domains, domainModel)
	}
	if len(domains) > 0 {
		if err = websiteDomainRepo.BatchCreate(ctx, domains); err != nil {
			return err
		}
	}
	if err != configDefaultNginx(website, domains, appInstall, runtime, create.RuntimeConfig) {
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

func (w WebsiteService) DeleteWebsite(ctx context.Context, req request.WebsiteDelete) error {
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
			if err := deleteAppInstall(ctx, appInstall, true, req.ForceDelete, true); err != nil && !req.ForceDelete {
				return err
			}
		}
	}

	uploadDir := fmt.Sprintf("%s/1panel/uploads/website/%s", global.CONF.System.BaseDir, website.Alias)
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	if req.DeleteBackup {
		localDir, err := loadLocalDir()
		if err != nil && !req.ForceDelete {
			return err
		}
		backupDir := fmt.Sprintf("%s/website/%s", localDir, website.Alias)
		if _, err := os.Stat(backupDir); err == nil {
			_ = os.RemoveAll(backupDir)
		}
		global.LOG.Infof("delete website %s backups successful", website.Alias)
	}
	_ = backupRepo.DeleteRecord(ctx, commonRepo.WithByType("website"), commonRepo.WithByName(website.Alias))

	if err := websiteRepo.DeleteBy(ctx, commonRepo.WithByID(req.ID)); err != nil {
		return err
	}
	if err := websiteDomainRepo.DeleteBy(ctx, websiteDomainRepo.WithWebsiteId(req.ID)); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) CreateWebsiteDomain(create request.WebsiteDomainCreate) (model.WebsiteDomain, error) {
	var domainModel model.WebsiteDomain
	var ports []int
	var domains []string

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
		if err := deleteListenAndServerName(website, ports, domains); err != nil {
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
		runtimeInstall.GetPath()
		configPath = path.Join(runtimeInstall.GetPath(), "conf", "php-fpm.conf")
	case constant.ConfigPHP:
		runtimeInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil {
			return response.FileInfo{}, err
		}
		runtimeInstall.GetPath()
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
		if err := deleteListenAndServerName(website, []int{443}, []string{}); err != nil {
			return response.WebsiteHTTPS{}, err
		}
		nginxParams := getNginxParamsFromStaticFile(dto.SSL, nil)
		nginxParams = append(nginxParams, dto.NginxParam{
			Name:   "if",
			Params: []string{"($scheme", "=", "http)"},
		})
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
		content, err := os.ReadFile(path.Join(sitePath, "log", req.LogType))
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
	if err = NewIAppInstalledService().Operate(context.Background(), appInstallReq); err != nil {
		_ = fileOp.WriteFile(phpConfigPath, strings.NewReader(string(contentBytes)), 0755)
		return err
	}
	return nil
}
