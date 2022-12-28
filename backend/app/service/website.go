package service

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"os"
	"path"
	"reflect"
	"strings"
	"time"

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
	CreateWebsite(create request.WebsiteCreate) error
	OpWebsite(req request.WebsiteOp) error
	GetWebsiteOptions() ([]string, error)
	Backup(id uint) error
	Recover(req request.WebsiteRecover) error
	RecoverByUpload(req request.WebsiteRecoverByFile) error
	UpdateWebsite(req request.WebsiteUpdate) error
	DeleteWebsite(req request.WebsiteDelete) error
	GetWebsite(id uint) (response.WebsiteDTO, error)
	CreateWebsiteDomain(create request.WebsiteDomainCreate) (model.WebsiteDomain, error)
	GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error)
	DeleteWebsiteDomain(domainId uint) error
	GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error)
	UpdateNginxConfigByScope(req request.NginxConfigUpdate) error
	GetWebsiteNginxConfig(websiteId uint) (response.FileInfo, error)
	GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error)
	OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (response.WebsiteHTTPS, error)
	PreInstallCheck(req request.WebsiteInstallCheckReq) ([]response.WebsitePreInstallCheck, error)
	GetWafConfig(req request.WebsiteWafReq) (response.WebsiteWafConfig, error)
	UpdateWafConfig(req request.WebsiteWafUpdate) error
}

func NewWebsiteService() IWebsiteService {
	return &WebsiteService{}
}

func (w WebsiteService) PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteDTO, error) {
	var websiteDTOs []response.WebsiteDTO
	total, websites, err := websiteRepo.Page(req.Page, req.PageSize)
	if err != nil {
		return 0, nil, err
	}
	for _, web := range websites {
		websiteDTOs = append(websiteDTOs, response.WebsiteDTO{
			Website: web,
		})
	}
	return total, websiteDTOs, nil
}

func (w WebsiteService) CreateWebsite(create request.WebsiteCreate) error {
	if exist, _ := websiteRepo.GetBy(websiteRepo.WithDomain(create.PrimaryDomain)); len(exist) > 0 {
		return buserr.New(constant.ErrNameIsExist)
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
		AppInstallID:   create.AppInstallID,
		WebsiteGroupID: create.WebsiteGroupID,
		Protocol:       constant.ProtocolHTTP,
		Proxy:          create.Proxy,
	}

	tx, ctx := getTxAndContext()
	var appInstall *model.AppInstall
	switch create.Type {
	case constant.Deployment:
		if create.AppType == constant.NewApp {
			var req request.AppInstallCreate
			req.Name = create.AppInstall.Name
			req.AppDetailId = create.AppInstall.AppDetailId
			req.Params = create.AppInstall.Params
			install, err := ServiceGroupApp.Install(ctx, req)
			if err != nil {
				return err
			}
			website.AppInstallID = install.ID
			appInstall = install
		} else {
			install, err := appInstallRepo.GetFirst(commonRepo.WithByID(create.AppInstallID))
			if err != nil {
				return err
			}
			appInstall = &install
			website.AppInstallID = 0
		}
	}

	if err := websiteRepo.Create(ctx, website); err != nil {
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
			tx.Rollback()
			return err
		}
		if reflect.DeepEqual(domainModel, model.WebsiteDomain{}) {
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

	if err := configDefaultNginx(website, domains, appInstall); err != nil {
		tx.Rollback()
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

func (w WebsiteService) Backup(id uint) error {
	localDir, err := loadLocalDir()
	if err != nil {
		return err
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return err
	}

	fileName := fmt.Sprintf("%s_%s", website.PrimaryDomain, time.Now().Format("20060102150405"))
	backupDir := fmt.Sprintf("website/%s", website.PrimaryDomain)
	if err := handleWebsiteBackup("LOCAL", localDir, backupDir, website.PrimaryDomain, fileName); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) RecoverByUpload(req request.WebsiteRecoverByFile) error {
	if err := handleUnTar(fmt.Sprintf("%s/%s", req.FileDir, req.FileName), req.FileDir); err != nil {
		return err
	}
	tmpDir := fmt.Sprintf("%s/%s", req.FileDir, strings.ReplaceAll(req.FileName, ".tar.gz", ""))
	webJson, err := os.ReadFile(fmt.Sprintf("%s/website.json", tmpDir))
	if err != nil {
		return err
	}
	var websiteInfo WebsiteInfo
	if err := json.Unmarshal(webJson, &websiteInfo); err != nil {
		return err
	}
	if websiteInfo.WebsiteName != req.WebsiteName || websiteInfo.WebsiteType != req.Type {
		return errors.New("上传文件与选中网站不匹配，无法恢复")
	}

	website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(req.WebsiteName))
	if err != nil {
		return err
	}
	if err := handleWebsiteRecover(&website, tmpDir); err != nil {
		return err
	}

	return nil
}

func (w WebsiteService) Recover(req request.WebsiteRecover) error {
	website, err := websiteRepo.GetFirst(websiteRepo.WithDomain(req.WebsiteName))
	if err != nil {
		return err
	}

	if !strings.Contains(req.BackupName, "/") {
		return errors.New("error path of request")
	}
	fileDir := path.Dir(req.BackupName)
	pathName := strings.ReplaceAll(path.Base(req.BackupName), ".tar.gz", "")
	if err := handleUnTar(req.BackupName, fileDir); err != nil {
		return err
	}
	fileDir = fileDir + "/" + pathName

	if err := handleWebsiteRecover(&website, fileDir); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) UpdateWebsite(req request.WebsiteUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	website.PrimaryDomain = req.PrimaryDomain
	website.WebsiteGroupID = req.WebsiteGroupID
	website.Remark = req.Remark

	return websiteRepo.Save(context.TODO(), &website)
}

func (w WebsiteService) GetWebsite(id uint) (response.WebsiteDTO, error) {
	var res response.WebsiteDTO
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return res, err
	}
	res.Website = website

	nginxInstall, err := getAppInstallByKey(constant.AppNginx)
	if err != nil {
		return res, err
	}
	sitePath := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "www", "sites", website.Alias)
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
	tx, ctx := getTxAndContext()

	if req.DeleteApp {
		websites, _ := websiteRepo.GetBy(websiteRepo.WithAppInstallId(website.AppInstallID))
		if len(websites) > 1 {
			return buserr.New(constant.ErrAppDelete)
		}
		appInstall, err := appInstallRepo.GetFirst(commonRepo.WithByID(website.AppInstallID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if !reflect.DeepEqual(model.AppInstall{}, appInstall) {
			if err := deleteAppInstall(ctx, appInstall, true); err != nil && !req.ForceDelete {
				return err
			}
		}
	}
	if req.DeleteBackup {
		backups, _ := backupRepo.ListRecord(backupRepo.WithByType("website-"+website.Type), commonRepo.WithByName(website.PrimaryDomain))
		if len(backups) > 0 {
			fileOp := files.NewFileOp()
			for _, b := range backups {
				_ = fileOp.DeleteDir(b.FileDir)
			}
			if err := backupRepo.DeleteRecord(ctx, backupRepo.WithByType("website-"+website.Type), commonRepo.WithByName(website.PrimaryDomain)); err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	if err := websiteRepo.DeleteBy(ctx, commonRepo.WithByID(req.ID)); err != nil {
		tx.Rollback()
		return err
	}
	if err := websiteDomainRepo.DeleteBy(ctx, websiteDomainRepo.WithWebsiteId(req.ID)); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
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
		return deleteNginxConfig(constant.NginxScopeServer, keys, &website)
	}
	params := getNginxParams(req.Params, keys)
	if req.Operate == constant.ConfigNew {
		if _, ok := dto.StaticFileKeyMap[req.Scope]; ok {
			params = getNginxParamsFromStaticFile(req.Scope, params)
		}
	}
	return updateNginxConfig(constant.NginxScopeServer, params, &website)
}

func (w WebsiteService) GetWebsiteNginxConfig(websiteId uint) (response.FileInfo, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteId))
	if err != nil {
		return response.FileInfo{}, err
	}

	nginxApp, err := appRepo.GetFirst(appRepo.WithKey(constant.AppNginx))
	if err != nil {
		return response.FileInfo{}, err
	}
	nginxInstall, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(nginxApp.ID))
	if err != nil {
		return response.FileInfo{}, err
	}

	configPath := path.Join(constant.AppInstallDir, constant.AppNginx, nginxInstall.Name, "conf", "conf.d", website.Alias+".conf")

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

	if !req.Enable {
		website.Protocol = constant.ProtocolHTTP
		website.WebsiteSSLID = 0
		if err := deleteListenAndServerName(website, []int{443}, []string{}); err != nil {
			return response.WebsiteHTTPS{}, err
		}
		if err := deleteNginxConfig(constant.NginxScopeServer, getKeysFromStaticFile(dto.SSL), &website); err != nil {
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
	if err := applySSL(website, websiteSSL, req.HttpConfig); err != nil {
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

	app, err := appRepo.GetFirst(appRepo.WithKey(constant.AppNginx))
	if err != nil {
		return nil, err
	}
	appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID))
	if reflect.DeepEqual(appInstall, model.AppInstall{}) {
		res = append(res, response.WebsitePreInstallCheck{
			Name:    appInstall.Name,
			AppName: app.Name,
			Status:  buserr.WithMessage(constant.ErrNotInstall, app.Name, nil).Error(),
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
		installList, _ := appInstallRepo.GetBy(commonRepo.WithIdsIn(checkIds))
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

	filePath := path.Join(nginxFull.SiteDir, "sites", website.Alias, "waf", "rules", req.Rule)
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
