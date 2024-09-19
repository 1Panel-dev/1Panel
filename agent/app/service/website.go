package service

import (
	"bufio"
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"

	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/jinzhu/copier"

	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/spf13/afero"

	"github.com/1Panel-dev/1Panel/agent/app/api/v2/helper"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"golang.org/x/crypto/bcrypt"
)

type WebsiteService struct {
}

type IWebsiteService interface {
	PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteRes, error)
	GetWebsites() ([]response.WebsiteDTO, error)
	CreateWebsite(create request.WebsiteCreate) error
	OpWebsite(req request.WebsiteOp) error
	GetWebsiteOptions(req request.WebsiteOptionReq) ([]response.WebsiteOption, error)
	UpdateWebsite(req request.WebsiteUpdate) error
	DeleteWebsite(req request.WebsiteDelete) error
	GetWebsite(id uint) (response.WebsiteDTO, error)

	CreateWebsiteDomain(create request.WebsiteDomainCreate) ([]model.WebsiteDomain, error)
	GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error)
	DeleteWebsiteDomain(domainId uint) error
	UpdateWebsiteDomain(req request.WebsiteDomainUpdate) error

	GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error)
	UpdateNginxConfigByScope(req request.NginxConfigUpdate) error
	GetWebsiteNginxConfig(websiteId uint, configType string) (*response.FileInfo, error)
	UpdateNginxConfigFile(req request.WebsiteNginxUpdate) error
	GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error)
	OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (*response.WebsiteHTTPS, error)
	OpWebsiteLog(req request.WebsiteLogReq) (*response.WebsiteLog, error)
	ChangeDefaultServer(id uint) error
	PreInstallCheck(req request.WebsiteInstallCheckReq) ([]response.WebsitePreInstallCheck, error)

	ChangePHPVersion(req request.WebsitePHPVersionReq) error

	GetRewriteConfig(req request.NginxRewriteReq) (*response.NginxRewriteRes, error)
	UpdateRewriteConfig(req request.NginxRewriteUpdate) error
	LoadWebsiteDirConfig(req request.WebsiteCommonReq) (*response.WebsiteDirConfig, error)
	UpdateSiteDir(req request.WebsiteUpdateDir) error
	UpdateSitePermission(req request.WebsiteUpdateDirPermission) error

	OperateProxy(req request.WebsiteProxyConfig) (err error)
	GetProxies(id uint) (res []request.WebsiteProxyConfig, err error)
	UpdateProxyFile(req request.NginxProxyUpdate) (err error)
	UpdateProxyCache(req request.NginxProxyCacheUpdate) (err error)
	GetProxyCache(id uint) (res response.NginxProxyCache, err error)

	GetAntiLeech(id uint) (*response.NginxAntiLeechRes, error)
	UpdateAntiLeech(req request.NginxAntiLeechUpdate) (err error)

	OperateRedirect(req request.NginxRedirectReq) (err error)
	GetRedirect(id uint) (res []response.NginxRedirectConfig, err error)
	UpdateRedirectFile(req request.NginxRedirectUpdate) (err error)

	GetAuthBasics(req request.NginxAuthReq) (res response.NginxAuthRes, err error)
	UpdateAuthBasic(req request.NginxAuthUpdate) (err error)
	GetPathAuthBasics(req request.NginxAuthReq) (res []response.NginxPathAuthRes, err error)
	UpdatePathAuthBasic(req request.NginxPathAuthUpdate) error

	UpdateDefaultHtml(req request.WebsiteHtmlUpdate) error
	GetDefaultHtml(resourceType string) (*response.WebsiteHtmlRes, error)

	GetLoadBalances(id uint) ([]dto.NginxUpstream, error)
	CreateLoadBalance(req request.WebsiteLBCreate) error
	DeleteLoadBalance(req request.WebsiteLBDelete) error
	UpdateLoadBalance(req request.WebsiteLBUpdate) error
	UpdateLoadBalanceFile(req request.WebsiteLBUpdateFile) error

	SetRealIPConfig(req request.WebsiteRealIP) error
	GetRealIPConfig(websiteID uint) (*response.WebsiteRealIP, error)

	ChangeGroup(group, newGroup uint) error
}

func NewIWebsiteService() IWebsiteService {
	return &WebsiteService{}
}

func (w WebsiteService) PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteRes, error) {
	var (
		websiteDTOs []response.WebsiteRes
		opts        []repo.DBOption
	)
	opts = append(opts, commonRepo.WithOrderRuleBy(req.OrderBy, req.Order))
	if req.Name != "" {
		domains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithDomainLike(req.Name))
		if len(domains) > 0 {
			var websiteIds []uint
			for _, domain := range domains {
				websiteIds = append(websiteIds, domain.WebsiteID)
			}
			opts = append(opts, commonRepo.WithByIDs(websiteIds))
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
			runtime, _ := runtimeRepo.GetFirst(commonRepo.WithByID(web.RuntimeID))
			if runtime != nil {
				runtimeName = runtime.Name
				runtimeType = runtime.Type
			}
		}
		sitePath := GetSitePath(web, SiteDir)

		siteDTO := response.WebsiteRes{
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
		}

		sites, _ := websiteRepo.List(websiteRepo.WithParentID(web.ID))
		if len(sites) > 0 {
			for _, site := range sites {
				siteDTO.ChildSites = append(siteDTO.ChildSites, site.PrimaryDomain)
			}
		}
		websiteDTOs = append(websiteDTOs, siteDTO)
	}
	return total, websiteDTOs, nil
}

func (w WebsiteService) GetWebsites() ([]response.WebsiteDTO, error) {
	var websiteDTOs []response.WebsiteDTO
	websites, _ := websiteRepo.List(commonRepo.WithOrderRuleBy("primary_domain", "ascending"))
	for _, web := range websites {
		res := response.WebsiteDTO{
			Website: web,
		}
		websiteDTOs = append(websiteDTOs, res)
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
	if len(create.FtpPassword) != 0 {
		pass, err := base64.StdEncoding.DecodeString(create.FtpPassword)
		if err != nil {
			return err
		}
		create.FtpPassword = string(pass)
	}

	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	defaultHttpPort := nginxInstall.HttpPort
	var (
		domains []model.WebsiteDomain
	)
	domains, _, _, err = getWebsiteDomains(create.Domains, defaultHttpPort, 0)
	if err != nil {
		return err
	}
	primaryDomain := domains[0].Domain
	if domains[0].Port != defaultHttpPort {
		primaryDomain = fmt.Sprintf("%s:%v", domains[0].Domain, domains[0].Port)
	}

	defaultDate, _ := time.Parse(constant.DateLayout, constant.DefaultDate)
	website := &model.Website{
		PrimaryDomain:  primaryDomain,
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

	createTask, err := task.NewTaskWithOps(primaryDomain, task.TaskCreate, task.TaskScopeWebsite, create.TaskID, 0)
	if err != nil {
		return err
	}

	if create.CreateDb {
		createDataBase := func(t *task.Task) error {
			database, _ := databaseRepo.Get(commonRepo.WithByName(create.DbHost))
			if database.ID == 0 {
				return nil
			}
			dbConfig := create.DataBaseConfig
			switch database.Type {
			case constant.AppPostgresql, constant.AppPostgres:
				oldPostgresqlDb, _ := postgresqlRepo.Get(commonRepo.WithByName(create.DbName), commonRepo.WithByFrom(constant.ResourceLocal))
				if oldPostgresqlDb.ID > 0 {
					return buserr.New(constant.ErrDbUserNotValid)
				}
				var createPostgresql dto.PostgresqlDBCreate
				createPostgresql.Name = dbConfig.DbName
				createPostgresql.Username = dbConfig.DbUser
				createPostgresql.Database = database.Name
				createPostgresql.Format = dbConfig.DBFormat
				createPostgresql.Password = dbConfig.DbPassword
				createPostgresql.From = database.From
				createPostgresql.SuperUser = true
				pgDB, err := NewIPostgresqlService().Create(context.Background(), createPostgresql)
				if err != nil {
					return err
				}
				website.DbID = pgDB.ID
				website.DbType = database.Type
			case constant.AppMysql, constant.AppMariaDB:
				oldMysqlDb, _ := mysqlRepo.Get(commonRepo.WithByName(dbConfig.DbName), commonRepo.WithByFrom(constant.ResourceLocal))
				if oldMysqlDb.ID > 0 {
					return buserr.New(constant.ErrDbUserNotValid)
				}
				var createMysql dto.MysqlDBCreate
				createMysql.Name = dbConfig.DbName
				createMysql.Username = dbConfig.DbUser
				createMysql.Database = database.Name
				createMysql.Format = dbConfig.DBFormat
				createMysql.Permission = "%"
				createMysql.Password = dbConfig.DbPassword
				createMysql.From = database.From
				mysqlDB, err := NewIMysqlService().Create(context.Background(), createMysql)
				if err != nil {
					return err
				}
				website.DbID = mysqlDB.ID
				website.DbType = database.Type
			}
			return nil
		}
		createTask.AddSubTask(task.GetTaskName(create.DbName, task.TaskCreate, task.TaskScopeDatabase), createDataBase, nil)
	}

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
			install, err = NewIAppService().Install(req)
			if err != nil {
				return err
			}
			appInstall = install
			website.AppInstallID = install.ID
			website.Proxy = fmt.Sprintf("127.0.0.1:%d", appInstall.HttpPort)
		} else {
			var install model.AppInstall
			install, err = appInstallRepo.GetFirst(commonRepo.WithByID(create.AppInstallID))
			if err != nil {
				return err
			}
			configApp := func(t *task.Task) error {
				appInstall = &install
				website.AppInstallID = appInstall.ID
				website.Proxy = fmt.Sprintf("127.0.0.1:%d", appInstall.HttpPort)
				return nil
			}
			createTask.AddSubTask(i18n.GetMsgByKey("ConfigApp"), configApp, nil)
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
				website.Proxy = fmt.Sprintf("127.0.0.1:%d", runtime.Port)
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
		case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimeGo:
			website.Proxy = fmt.Sprintf("127.0.0.1:%d", runtime.Port)
		}
	case constant.Subsite:
		parentWebsite, err := websiteRepo.GetFirst(commonRepo.WithByID(create.ParentWebsiteID))
		if err != nil {
			return err
		}
		website.ParentWebsiteID = parentWebsite.ID
	}

	if len(create.FtpUser) != 0 && len(create.FtpPassword) != 0 {
		createFtpUser := func(t *task.Task) error {
			indexDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "index")
			itemID, err := NewIFtpService().Create(dto.FtpCreate{User: create.FtpUser, Password: create.FtpPassword, Path: indexDir})
			if err != nil {
				createTask.Log(fmt.Sprintf("create ftp for website failed, err: %v", err))
			}
			website.FtpID = itemID
			return nil
		}
		deleteFtpUser := func(t *task.Task) {
			if website.FtpID > 0 {
				if err = NewIFtpService().Delete(dto.BatchDeleteReq{Ids: []uint{website.FtpID}}); err != nil {
					createTask.Log(err.Error())
				}
			}
		}
		createTask.AddSubTask(i18n.GetWithName("ConfigFTP", create.FtpUser), createFtpUser, deleteFtpUser)
	}

	configNginx := func(t *task.Task) error {
		if err = configDefaultNginx(website, domains, appInstall, runtime); err != nil {
			return err
		}
		if err = createWafConfig(website, domains); err != nil {
			return err
		}
		tx, ctx := helper.GetTxAndContext()
		defer tx.Rollback()
		if err = websiteRepo.Create(ctx, website); err != nil {
			return err
		}
		t.Task.ResourceID = website.ID
		for i := range domains {
			domains[i].WebsiteID = website.ID
		}
		if err = websiteDomainRepo.BatchCreate(ctx, domains); err != nil {
			return err
		}
		tx.Commit()
		return nil
	}

	deleteWebsite := func(t *task.Task) {
		_ = deleteWebsiteFolder(nginxInstall, website)
	}

	createTask.AddSubTask(i18n.GetMsgByKey("ConfigOpenresty"), configNginx, deleteWebsite)

	if create.EnableSSL {
		enableSSL := func(t *task.Task) error {
			websiteModel, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(create.WebsiteSSLID))
			if err != nil {
				return err
			}
			website.Protocol = constant.ProtocolHTTPS
			website.WebsiteSSLID = create.WebsiteSSLID
			appSSLReq := request.WebsiteHTTPSOp{
				WebsiteID:    website.ID,
				Enable:       true,
				WebsiteSSLID: websiteModel.ID,
				Type:         "existed",
				HttpConfig:   "HTTPToHTTPS",
				SSLProtocol:  []string{"TLSv1.3", "TLSv1.2", "TLSv1.1", "TLSv1"},
				Algorithm:    "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:!aNULL:!eNULL:!EXPORT:!DSS:!DES:!RC4:!3DES:!MD5:!PSK:!KRB5:!SRP:!CAMELLIA:!SEED",
				Hsts:         true,
			}
			if err = applySSL(website, *websiteModel, appSSLReq); err != nil {
				return err
			}
			if err = websiteRepo.Save(context.Background(), website); err != nil {
				return err
			}
			return nil
		}
		createTask.AddSubTaskWithIgnoreErr(i18n.GetMsgByKey("EnableSSL"), enableSSL)
	}

	return createTask.Execute()
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

func (w WebsiteService) GetWebsiteOptions(req request.WebsiteOptionReq) ([]response.WebsiteOption, error) {
	var options []repo.DBOption
	if len(req.Types) > 0 {
		options = append(options, commonRepo.WithTypes(req.Types))
	}
	webs, _ := websiteRepo.List(options...)
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
	res.ErrorLogPath = GetSitePath(website, SiteErrorLog)
	res.AccessLogPath = GetSitePath(website, SiteAccessLog)
	res.SitePath = GetSitePath(website, SiteDir)
	res.SiteDir = website.SiteDir
	if website.Type == constant.Runtime {
		runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
		if err != nil {
			return res, err
		}
		res.RuntimeType = runtime.Type
		res.RuntimeName = runtime.Name
	}
	return res, nil
}

func (w WebsiteService) DeleteWebsite(req request.WebsiteDelete) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if website.Type != constant.Subsite {
		parentWebsites, _ := websiteRepo.List(websiteRepo.WithParentID(website.ID))
		if len(parentWebsites) > 0 {
			var names []string
			for _, site := range parentWebsites {
				names = append(names, site.PrimaryDomain)
			}
			return buserr.WithName("ErrParentWebsite", strings.Join(names, ","))
		}
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
			deleteReq := request.AppInstallDelete{
				Install:      appInstall,
				ForceDelete:  req.ForceDelete,
				DeleteBackup: true,
				DeleteDB:     true,
			}
			if err = deleteAppInstall(deleteReq); err != nil && !req.ForceDelete {
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

func (w WebsiteService) UpdateWebsiteDomain(req request.WebsiteDomainUpdate) error {
	domain, err := websiteDomainRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	domain.SSL = req.SSL
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(domain.WebsiteID))
	if err != nil {
		return err
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	nginxConfig := nginxFull.SiteConfig
	config := nginxFull.SiteConfig.Config
	server := config.FindServers()[0]
	server.DeleteListen(strconv.Itoa(domain.Port))
	if website.IPV6 {
		server.DeleteListen("[::]:" + strconv.Itoa(domain.Port))
	}
	http3 := isHttp3(server)
	setListen(server, strconv.Itoa(domain.Port), website.IPV6, http3, website.DefaultServer, domain.SSL && website.Protocol == constant.ProtocolHTTPS)
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err = nginxCheckAndReload(nginxConfig.OldContent, nginxConfig.FilePath, nginxFull.Install.ContainerName); err != nil {
		return err
	}
	return websiteDomainRepo.Save(context.TODO(), &domain)
}

func (w WebsiteService) CreateWebsiteDomain(create request.WebsiteDomainCreate) ([]model.WebsiteDomain, error) {
	var (
		domainModels []model.WebsiteDomain
		addPorts     []int
	)
	httpPort, _, err := getAppInstallPort(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(create.WebsiteID))
	if err != nil {
		return nil, err
	}

	domainModels, addPorts, _, err = getWebsiteDomains(create.Domains, httpPort, create.WebsiteID)
	if err != nil {
		return nil, err
	}
	go func() {
		_ = OperateFirewallPort(nil, addPorts)
	}()

	if err = addListenAndServerName(website, domainModels); err != nil {
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

func (w WebsiteService) GetWebsiteNginxConfig(websiteID uint, configType string) (*response.FileInfo, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteID))
	if err != nil {
		return nil, err
	}
	configPath := ""
	switch configType {
	case constant.AppOpenresty:
		configPath = GetSitePath(website, SiteConf)
	}
	info, err := files.NewFileInfo(files.FileOption{
		Path:   configPath,
		Expand: true,
	})
	if err != nil {
		return nil, err
	}
	return &response.FileInfo{FileInfo: *info}, nil
}

func (w WebsiteService) GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteId))
	if err != nil {
		return response.WebsiteHTTPS{}, err
	}
	var (
		res        response.WebsiteHTTPS
		httpsPorts []string
	)
	websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(websiteId))
	for _, domain := range websiteDomains {
		if domain.SSL {
			httpsPorts = append(httpsPorts, strconv.Itoa(domain.Port))
		}
	}
	if len(httpsPorts) == 0 {
		nginxInstall, _ := getAppInstallByKey(constant.AppOpenresty)
		res.HttpsPort = strconv.Itoa(nginxInstall.HttpsPort)
	} else {
		res.HttpsPort = strings.Join(httpsPorts, ",")
	}
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
	params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"ssl_protocols", "ssl_ciphers", "add_header", "listen"}, &website)
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
		if p.Name == "add_header" && len(p.Params) > 0 {
			if p.Params[0] == "Strict-Transport-Security" {
				res.Hsts = true
			}
			if p.Params[0] == "Alt-Svc" {
				res.Http3 = true
			}
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

		httpsPorts, err := getHttpsPort(&website)
		if err != nil {
			return nil, err
		}
		if len(httpsPorts) == 1 && httpsPorts[0] == nginxInstall.HttpsPort {
			httpsPortStr := strconv.Itoa(httpsPorts[0])
			if err = deleteListenAndServerName(website, []string{httpsPortStr, "[::]:" + httpsPortStr}, []string{}); err != nil {
				return nil, err
			}
		} else {
			for _, port := range httpsPorts {
				httpsPortStr := strconv.Itoa(port)
				if err = removeSSLListen(website, []string{httpsPortStr}); err != nil {
					return nil, err
				}
			}
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
			dto.NginxParam{
				Name: "http2",
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

		res.SSL = websiteSSL
	}

	website.Protocol = constant.ProtocolHTTPS
	if err = applySSL(&website, websiteSSL, req); err != nil {
		return nil, err
	}
	website.HttpConfig = req.HttpConfig

	if websiteSSL.ID == 0 {
		if err = websiteSSLRepo.Create(ctx, &websiteSSL); err != nil {
			return nil, err
		}
		website.WebsiteSSLID = websiteSSL.ID
	}
	if err = websiteRepo.Save(ctx, &website); err != nil {
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
		installList, _ := appInstallRepo.ListBy(commonRepo.WithByIDs(checkIds))
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
	if err = files.NewFileOp().WriteFile(filePath, strings.NewReader(req.Content), 0755); err != nil {
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
			Name:   key,
			Params: []string{"off"},
		})

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

func (w WebsiteService) ChangePHPVersion(req request.WebsitePHPVersionReq) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if website.Type == constant.Runtime {
		oldRuntime, err := runtimeRepo.GetFirst(commonRepo.WithByID(website.RuntimeID))
		if err != nil {
			return err
		}
		if oldRuntime.Resource == constant.ResourceLocal {
			return buserr.New("ErrPHPResource")
		}
	}
	configPath := GetSitePath(website, SiteConf)
	nginxContent, err := files.NewFileOp().GetContent(configPath)
	if err != nil {
		return err
	}
	config, err := parser.NewStringParser(string(nginxContent)).Parse()
	if err != nil {
		return err
	}
	servers := config.FindServers()
	if len(servers) == 0 {
		return errors.New("nginx config is not valid")
	}
	server := servers[0]

	if req.RuntimeID > 0 {
		runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.RuntimeID))
		if err != nil {
			return err
		}
		if runtime.Resource == constant.ResourceLocal {
			return buserr.New("ErrPHPResource")
		}
		website.RuntimeID = req.RuntimeID
		phpProxy := fmt.Sprintf("127.0.0.1:%d", runtime.Port)
		website.Proxy = phpProxy
		server.UpdatePHPProxy([]string{website.Proxy}, "")
		website.Type = constant.Runtime
	} else {
		website.RuntimeID = 0
		website.Type = constant.Static
		website.Proxy = ""
		server.RemoveDirective("location", []string{"~", "[^/]\\.php(/|$)"})
	}

	config.FilePath = configPath
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if err = nginxCheckAndReload(string(nginxContent), configPath, nginxInstall.ContainerName); err != nil {
		return err
	}
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) UpdateRewriteConfig(req request.NginxRewriteUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	includePath := fmt.Sprintf("/www/sites/%s/rewrite/%s.conf", website.Alias, website.Alias)
	absolutePath := GetSitePath(website, SiteReWritePath)
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
		rewriteConfPath := GetSitePath(website, SiteReWritePath)
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

func (w WebsiteService) OperateProxy(req request.WebsiteProxyConfig) (err error) {
	var (
		website    model.Website
		par        *parser.Parser
		oldContent []byte
	)

	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	includeDir := GetSitePath(website, SiteProxyDir)
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
		if err = openProxyCache(website); err != nil {
			return
		}
		location.AddCache(req.CacheTime, req.CacheUnit, fmt.Sprintf("proxy_cache_zone_of_%s", website.Alias))
	} else {
		location.RemoveCache(fmt.Sprintf("proxy_cache_zone_of_%s", website.Alias))
	}
	if len(req.Replaces) > 0 {
		location.AddSubFilter(req.Replaces)
	} else {
		location.RemoveSubFilter()
	}
	if req.SNI {
		location.UpdateDirective("proxy_ssl_server_name", []string{"on"})
	} else {
		location.UpdateDirective("proxy_ssl_server_name", []string{"off"})
	}
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
	return updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website)
}

func (w WebsiteService) UpdateProxyCache(req request.NginxProxyCacheUpdate) (err error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	cacheDir := GetSitePath(website, SiteCacheDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(cacheDir) {
		_ = fileOp.CreateDir(cacheDir, 0755)
	}
	if req.Open {
		proxyCachePath := fmt.Sprintf("/www/sites/%s/cache levels=1:2 keys_zone=proxy_cache_zone_of_%s:%d%s max_size=%d%s inactive=%d%s", website.Alias, website.Alias, req.ShareCache, req.ShareCacheUnit, req.CacheLimit, req.CacheLimitUnit, req.CacheExpire, req.CacheExpireUnit)
		return updateNginxConfig("", []dto.NginxParam{{Name: "proxy_cache_path", Params: []string{proxyCachePath}}}, &website)
	}
	return deleteNginxConfig("", []dto.NginxParam{{Name: "proxy_cache_path"}}, &website)
}

func (w WebsiteService) GetProxyCache(id uint) (res response.NginxProxyCache, err error) {
	var (
		website model.Website
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return
	}

	parser, err := parser.NewParser(GetSitePath(website, SiteConf))
	if err != nil {
		return
	}
	config, err := parser.Parse()
	if err != nil {
		return
	}
	var params []string
	for _, d := range config.GetDirectives() {
		if d.GetName() == "proxy_cache_path" {
			params = d.GetParameters()
		}
	}
	if len(params) == 0 {
		return
	}
	for _, param := range params {
		if match, _ := regexp.MatchString(`keys_zone=proxy_cache_zone_of_[\w.]+:\d+[kmgt]?`, param); match {
			r := regexp.MustCompile(`keys_zone=proxy_cache_zone_of_[\w.]+:(\d+)([kmgt]?)`)
			matches := r.FindStringSubmatch(param)
			if len(matches) > 0 {
				res.ShareCache, _ = strconv.Atoi(matches[1])
				res.ShareCacheUnit = matches[2]
			}
		}

		if match, _ := regexp.MatchString(`max_size=\d+(\.\d+)?[kmgt]?`, param); match {
			r := regexp.MustCompile(`max_size=([0-9.]+)([kmgt]?)`)
			matches := r.FindStringSubmatch(param)
			if len(matches) > 0 {
				res.CacheLimit, _ = strconv.ParseFloat(matches[1], 64)
				res.CacheLimitUnit = matches[2]
			}
		}
		if match, _ := regexp.MatchString(`inactive=\d+[smhd]`, param); match {
			r := regexp.MustCompile(`inactive=(\d+)([smhd])`)
			matches := r.FindStringSubmatch(param)
			if len(matches) > 0 {
				res.CacheExpire, _ = strconv.Atoi(matches[1])
				res.CacheExpireUnit = matches[2]
			}
		}
	}
	res.Open = true
	return
}

func (w WebsiteService) GetProxies(id uint) (res []request.WebsiteProxyConfig, err error) {
	var (
		website  model.Website
		fileList response.FileInfo
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return
	}
	includeDir := GetSitePath(website, SiteProxyDir)
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

func (w WebsiteService) GetPathAuthBasics(req request.NginxAuthReq) (res []response.NginxPathAuthRes, err error) {
	var (
		website      model.Website
		nginxInstall model.AppInstall
		authContent  []byte
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	nginxInstall, err = getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	authDir := fmt.Sprintf("/www/sites/%s/path_auth", website.Alias)
	absoluteAuthDir := path.Join(nginxInstall.GetPath(), authDir)
	passDir := path.Join(absoluteAuthDir, "pass")
	if !fileOp.Stat(absoluteAuthDir) || !fileOp.Stat(passDir) {
		return
	}

	entries, err := os.ReadDir(absoluteAuthDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := strings.TrimSuffix(entry.Name(), ".conf")
			pathAuth := dto.NginxPathAuth{
				Name: name,
			}
			configPath := path.Join(absoluteAuthDir, entry.Name())
			content, err := fileOp.GetContent(configPath)
			if err != nil {
				return nil, err
			}
			config, err := parser.NewStringParser(string(content)).Parse()
			if err != nil {
				return nil, err
			}
			directives := config.Directives
			location, _ := directives[0].(*components.Location)
			pathAuth.Path = location.Match
			passPath := path.Join(passDir, fmt.Sprintf("%s.pass", name))
			authContent, err = fileOp.GetContent(passPath)
			if err != nil {
				return nil, err
			}
			authArray := strings.Split(string(authContent), "\n")
			for _, line := range authArray {
				if line == "" {
					continue
				}
				params := strings.Split(line, ":")
				pathAuth.Username = params[0]
				if len(params) == 3 {
					pathAuth.Remark = params[2]
				}
			}
			res = append(res, response.NginxPathAuthRes{
				NginxPathAuth: pathAuth,
			})
		}
	}
	return
}

func (w WebsiteService) UpdatePathAuthBasic(req request.NginxPathAuthUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	authDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "path_auth")
	if !fileOp.Stat(authDir) {
		_ = fileOp.CreateDir(authDir, 0755)
	}
	passDir := path.Join(authDir, "pass")
	if !fileOp.Stat(passDir) {
		_ = fileOp.CreateDir(passDir, 0755)
	}
	confPath := path.Join(authDir, fmt.Sprintf("%s.conf", req.Name))
	passPath := path.Join(passDir, fmt.Sprintf("%s.pass", req.Name))
	var config *components.Config
	switch req.Operate {
	case "delete":
		_ = fileOp.DeleteFile(confPath)
		_ = fileOp.DeleteFile(passPath)
		return updateNginxConfig(constant.NginxScopeServer, nil, &website)
	case "create":
		config, err = parser.NewStringParser(string(nginx_conf.PathAuth)).Parse()
		if err != nil {
			return err
		}
		if fileOp.Stat(confPath) || fileOp.Stat(passPath) {
			return buserr.New(constant.ErrNameIsExist)
		}
	case "edit":
		par, err := parser.NewParser(confPath)
		if err != nil {
			return err
		}
		config, err = par.Parse()
		if err != nil {
			return err
		}
	}
	config.FilePath = confPath
	directives := config.Directives
	location, _ := directives[0].(*components.Location)
	location.UpdateDirective("auth_basic_user_file", []string{fmt.Sprintf("/www/sites/%s/path_auth/pass/%s", website.Alias, fmt.Sprintf("%s.pass", req.Name))})
	location.ChangePath("~*", req.Path)
	var passwdHash []byte
	passwdHash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("%s:%s\n", req.Username, passwdHash)
	if req.Remark != "" {
		line = fmt.Sprintf("%s:%s:%s\n", req.Username, passwdHash, req.Remark)
	}
	_ = fileOp.SaveFile(passPath, line, 0644)
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/path_auth/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return nil
	}
	return nil
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
		website    model.Website
		oldContent []byte
	)

	website, err = websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	includeDir := GetSitePath(website, SiteRedirectDir)
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
		website  model.Website
		fileList response.FileInfo
	)
	website, err = websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return
	}
	includeDir := GetSitePath(website, SiteRedirectDir)
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
	absoluteIndexPath := GetSitePath(website, SiteIndexDir)
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
		if !file.IsDir() || file.Name() == "node_modules" || file.Name() == "vendor" {
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

func (w WebsiteService) GetLoadBalances(id uint) ([]dto.NginxUpstream, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "upstream")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		return nil, nil
	}
	entries, err := os.ReadDir(includeDir)
	if err != nil {
		return nil, err
	}
	var res []dto.NginxUpstream
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".conf") {
			continue
		}
		upstreamName := strings.TrimSuffix(name, ".conf")
		upstream := dto.NginxUpstream{
			Name: upstreamName,
		}
		upstreamPath := path.Join(includeDir, name)
		content, err := fileOp.GetContent(upstreamPath)
		if err != nil {
			return nil, err
		}
		upstream.Content = string(content)
		nginxParser, err := parser.NewParser(upstreamPath)
		if err != nil {
			return nil, err
		}
		config, err := nginxParser.Parse()
		if err != nil {
			return nil, err
		}
		upstreams := config.FindUpstreams()
		for _, up := range upstreams {
			if up.UpstreamName == upstreamName {
				directives := up.GetDirectives()
				for _, d := range directives {
					dName := d.GetName()
					if _, ok := dto.LBAlgorithms[dName]; ok {
						upstream.Algorithm = dName
					}
				}
				var servers []dto.NginxUpstreamServer
				for _, ups := range up.UpstreamServers {
					server := dto.NginxUpstreamServer{
						Server: ups.Address,
					}
					parameters := ups.Parameters
					if weight, ok := parameters["weight"]; ok {
						num, err := strconv.Atoi(weight)
						if err == nil {
							server.Weight = num
						}
					}
					if maxFails, ok := parameters["max_fails"]; ok {
						num, err := strconv.Atoi(maxFails)
						if err == nil {
							server.MaxFails = num
						}
					}
					if failTimeout, ok := parameters["fail_timeout"]; ok {
						server.FailTimeout = failTimeout
					}
					if maxConns, ok := parameters["max_conns"]; ok {
						num, err := strconv.Atoi(maxConns)
						if err == nil {
							server.MaxConns = num
						}
					}
					for _, flag := range ups.Flags {
						server.Flag = flag
					}
					servers = append(servers, server)
				}
				upstream.Servers = servers
			}
		}
		res = append(res, upstream)
	}
	return res, nil
}

func (w WebsiteService) CreateLoadBalance(req request.WebsiteLBCreate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "upstream")
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		_ = fileOp.CreateDir(includeDir, 0644)
	}
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	if fileOp.Stat(filePath) {
		return buserr.New(constant.ErrNameIsExist)
	}
	config, err := parser.NewStringParser(string(nginx_conf.Upstream)).Parse()
	if err != nil {
		return err
	}
	config.Block = &components.Block{}
	config.FilePath = filePath
	upstream := components.Upstream{
		UpstreamName: req.Name,
	}
	if req.Algorithm != "default" {
		upstream.UpdateDirective(req.Algorithm, []string{})
	}

	servers := make([]*components.UpstreamServer, 0)

	for _, server := range req.Servers {
		upstreamServer := &components.UpstreamServer{
			Address: server.Server,
		}
		parameters := make(map[string]string)
		if server.Weight > 0 {
			parameters["weight"] = strconv.Itoa(server.Weight)
		}
		if server.MaxFails > 0 {
			parameters["max_fails"] = strconv.Itoa(server.MaxFails)
		}
		if server.FailTimeout != "" {
			parameters["fail_timeout"] = server.FailTimeout
		}
		if server.MaxConns > 0 {
			parameters["max_conns"] = strconv.Itoa(server.MaxConns)
		}
		if server.Flag != "" {
			upstreamServer.Flags = []string{server.Flag}
		}
		upstreamServer.Parameters = parameters
		servers = append(servers, upstreamServer)
	}
	upstream.UpstreamServers = servers
	config.Block.Directives = append(config.Block.Directives, &upstream)

	defer func() {
		if err != nil {
			_ = fileOp.DeleteFile(filePath)
		}
	}()

	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/upstream/*.conf", website.Alias)
	if err = updateNginxConfig("", []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) UpdateLoadBalance(req request.WebsiteLBUpdate) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "upstream")
	fileOp := files.NewFileOp()
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	if !fileOp.Stat(filePath) {
		return nil
	}
	parser, err := parser.NewParser(filePath)
	if err != nil {
		return err
	}
	config, err := parser.Parse()
	if err != nil {
		return err
	}
	upstreams := config.FindUpstreams()
	for _, up := range upstreams {
		if up.UpstreamName == req.Name {
			directives := up.GetDirectives()
			for _, d := range directives {
				dName := d.GetName()
				if _, ok := dto.LBAlgorithms[dName]; ok {
					up.RemoveDirective(dName, nil)
				}
			}
			if req.Algorithm != "default" {
				up.UpdateDirective(req.Algorithm, []string{})
			}
			var servers []*components.UpstreamServer
			for _, server := range req.Servers {
				upstreamServer := &components.UpstreamServer{
					Address: server.Server,
				}
				parameters := make(map[string]string)
				if server.Weight > 0 {
					parameters["weight"] = strconv.Itoa(server.Weight)
				}
				if server.MaxFails > 0 {
					parameters["max_fails"] = strconv.Itoa(server.MaxFails)
				}
				if server.FailTimeout != "" {
					parameters["fail_timeout"] = server.FailTimeout
				}
				if server.MaxConns > 0 {
					parameters["max_conns"] = strconv.Itoa(server.MaxConns)
				}
				if server.Flag != "" {
					upstreamServer.Flags = []string{server.Flag}
				}
				upstreamServer.Parameters = parameters
				servers = append(servers, upstreamServer)
			}
			up.UpstreamServers = servers
		}
	}
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr(constant.ErrUpdateBuWebsite, err)
	}
	if err = opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) DeleteLoadBalance(req request.WebsiteLBDelete) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "upstream")
	fileOp := files.NewFileOp()
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	if !fileOp.Stat(filePath) {
		return nil
	}
	if err = fileOp.DeleteFile(filePath); err != nil {
		return err
	}
	return opNginx(nginxInstall.ContainerName, constant.NginxReload)
}

func (w WebsiteService) UpdateLoadBalanceFile(req request.WebsiteLBUpdateFile) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	includeDir := path.Join(nginxInstall.GetPath(), "www", "sites", website.Alias, "upstream")
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	fileOp := files.NewFileOp()
	oldContent, err := fileOp.GetContent(filePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(filePath, strings.NewReader(req.Content), 0755); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = fileOp.WriteFile(filePath, bytes.NewReader(oldContent), 0755)
		}
	}()
	return opNginx(nginxInstall.ContainerName, constant.NginxReload)
}

func (w WebsiteService) ChangeGroup(group, newGroup uint) error {
	return websiteRepo.UpdateGroup(group, newGroup)
}

func (w WebsiteService) SetRealIPConfig(req request.WebsiteRealIP) error {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	params := []dto.NginxParam{
		{Name: "real_ip_recursive", Params: []string{"on"}},
		{Name: "set_real_ip_from", Params: []string{}},
		{Name: "real_ip_header", Params: []string{}},
	}
	if req.Open {
		if err := deleteNginxConfig(constant.NginxScopeServer, params, &website); err != nil {
			return err
		}
		params = []dto.NginxParam{
			{Name: "real_ip_recursive", Params: []string{"on"}},
		}
		var ips []string
		ipArray := strings.Split(req.IPFrom, "\n")
		for _, ip := range ipArray {
			if ip == "" {
				continue
			}
			if parsedIP := net.ParseIP(ip); parsedIP == nil {
				if _, _, err := net.ParseCIDR(ip); err != nil {
					return buserr.New("ErrParseIP")
				}
			}
			ips = append(ips, strings.TrimSpace(ip))
		}
		for _, ip := range ips {
			params = append(params, dto.NginxParam{Name: "set_real_ip_from", Params: []string{ip}})
		}
		if req.IPHeader == "other" {
			params = append(params, dto.NginxParam{Name: "real_ip_header", Params: []string{req.IPOther}})
		} else {
			params = append(params, dto.NginxParam{Name: "real_ip_header", Params: []string{req.IPHeader}})
		}
		return updateNginxConfig(constant.NginxScopeServer, params, &website)
	}
	return deleteNginxConfig(constant.NginxScopeServer, params, &website)
}

func (w WebsiteService) GetRealIPConfig(websiteID uint) (*response.WebsiteRealIP, error) {
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(websiteID))
	if err != nil {
		return nil, err
	}
	params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"real_ip_recursive"}, &website)
	if err != nil {
		return nil, err
	}
	if len(params) == 0 || len(params[0].Params) == 0 {
		return &response.WebsiteRealIP{Open: false}, nil
	}
	params, err = getNginxParamsByKeys(constant.NginxScopeServer, []string{"set_real_ip_from", "real_ip_header"}, &website)
	if err != nil {
		return nil, err
	}
	res := &response.WebsiteRealIP{
		Open: true,
	}
	var ips []string
	for _, param := range params {
		if param.Name == "set_real_ip_from" {
			ips = append(ips, param.Params...)
		}
		if param.Name == "real_ip_header" {
			if _, ok := dto.RealIPKeys[param.Params[0]]; ok {
				res.IPHeader = param.Params[0]
			} else {
				res.IPHeader = "other"
				res.IPOther = param.Params[0]
			}
		}
	}
	res.IPFrom = strings.Join(ips, "\n")
	return res, err
}
