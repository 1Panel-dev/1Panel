package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/docker"

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
	"github.com/1Panel-dev/1Panel/agent/utils/re"
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

	ChangePHPVersion(req request.WebsitePHPVersionReq) error
	OperateCrossSiteAccess(req request.CrossSiteAccessOp) error
	ExecComposer(req request.ExecComposerReq) error
	ChangeGroup(group, newGroup uint) error
	ChangeDefaultServer(id uint) error
	PreInstallCheck(req request.WebsiteInstallCheckReq) ([]response.WebsitePreInstallCheck, error)
	OpWebsiteLog(req request.WebsiteLogReq) (*response.WebsiteLog, error)
	UpdateStream(req request.StreamUpdate) error

	GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error)
	UpdateNginxConfigByScope(req request.NginxConfigUpdate) error
	GetWebsiteNginxConfig(websiteId uint, configType string) (*response.FileInfo, error)
	UpdateNginxConfigFile(req request.WebsiteNginxUpdate) error

	GetWebsiteHTTPS(websiteId uint) (response.WebsiteHTTPS, error)
	OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (*response.WebsiteHTTPS, error)

	LoadWebsiteDirConfig(req request.WebsiteCommonReq) (*response.WebsiteDirConfig, error)
	UpdateSiteDir(req request.WebsiteUpdateDir) error
	UpdateSitePermission(req request.WebsiteUpdateDirPermission) error

	UpdateCors(req request.CorsConfigReq) error
	GetCors(websiteID uint) (*request.CorsConfig, error)

	GetAntiLeech(id uint) (*response.NginxAntiLeechRes, error)
	UpdateAntiLeech(req request.NginxAntiLeechUpdate) (err error)

	OperateRedirect(req request.NginxRedirectReq) (err error)
	GetRedirect(id uint) (res []response.NginxRedirectConfig, err error)
	UpdateRedirectFile(req request.NginxRedirectUpdate) (err error)

	UpdateDefaultHtml(req request.WebsiteHtmlUpdate) error
	GetDefaultHtml(resourceType string) (*response.WebsiteHtmlRes, error)

	SetRealIPConfig(req request.WebsiteRealIP) error
	GetRealIPConfig(websiteID uint) (*response.WebsiteRealIP, error)

	GetWebsiteResource(websiteID uint) ([]response.Resource, error)
	ListDatabases() ([]response.Database, error)
	ChangeDatabase(req request.ChangeDatabase) error

	GetLoadBalances(id uint) ([]dto.NginxUpstream, error)
	CreateLoadBalance(req request.WebsiteLBCreate) error
	DeleteLoadBalance(req request.WebsiteLBDelete) error
	UpdateLoadBalance(req request.WebsiteLBUpdate) error
	UpdateLoadBalanceFile(req request.WebsiteLBUpdateFile) error

	OperateProxy(req request.WebsiteProxyConfig) (err error)
	GetProxies(id uint) (res []request.WebsiteProxyConfig, err error)
	UpdateProxyFile(req request.NginxProxyUpdate) (err error)
	UpdateProxyCache(req request.NginxProxyCacheUpdate) (err error)
	GetProxyCache(id uint) (res response.NginxProxyCache, err error)
	ClearProxyCache(req request.NginxCommonReq) error
	DeleteProxy(req request.WebsiteProxyDel) (err error)

	CreateWebsiteDomain(create request.WebsiteDomainCreate) ([]model.WebsiteDomain, error)
	GetWebsiteDomain(websiteId uint) ([]model.WebsiteDomain, error)
	DeleteWebsiteDomain(domainId uint) error
	UpdateWebsiteDomain(req request.WebsiteDomainUpdate) error

	GetRewriteConfig(req request.NginxRewriteReq) (*response.NginxRewriteRes, error)
	UpdateRewriteConfig(req request.NginxRewriteUpdate) error
	OperateCustomRewrite(req request.CustomRewriteOperate) error
	ListCustomRewrite() ([]string, error)

	GetAuthBasics(req request.NginxAuthReq) (res response.NginxAuthRes, err error)
	UpdateAuthBasic(req request.NginxAuthUpdate) (err error)
	GetPathAuthBasics(req request.NginxAuthReq) (res []response.NginxPathAuthRes, err error)
	UpdatePathAuthBasic(req request.NginxPathAuthUpdate) error

	BatchOpWebsite(req request.BatchWebsiteOp) error
	BatchSetGroup(req request.BatchWebsiteGroup) error
	BatchSetHttps(ctx context.Context, req request.BatchWebsiteHttps) error
}

func NewIWebsiteService() IWebsiteService {
	return &WebsiteService{}
}

func (w WebsiteService) PageWebsite(req request.WebsiteSearch) (int64, []response.WebsiteRes, error) {
	var (
		websiteDTOs []response.WebsiteRes
		opts        []repo.DBOption
	)
	opts = append(opts, repo.WithOrderRuleBy(req.OrderBy, req.Order), repo.WithOrderRuleBy("updated_at", "descending"))
	if req.Name != "" {
		domains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithDomainLike(req.Name))
		if len(domains) > 0 {
			var websiteIds []uint
			for _, domain := range domains {
				websiteIds = append(websiteIds, domain.WebsiteID)
			}
			opts = append(opts, repo.WithByIDs(websiteIds))
		} else {
			opts = append(opts, websiteRepo.WithDomainLike(req.Name))
		}
	}
	if req.WebsiteGroupID != 0 {
		opts = append(opts, websiteRepo.WithGroupID(req.WebsiteGroupID))
	}
	if req.Type != "" {
		opts = append(opts, websiteRepo.WithType(req.Type))
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
			appInstall, err := appInstallRepo.GetFirst(repo.WithByID(web.AppInstallID))
			if err != nil {
				return 0, nil, err
			}
			appName = appInstall.Name
			appInstallID = appInstall.ID
		case constant.Runtime:
			runtime, _ := runtimeRepo.GetFirst(context.Background(), repo.WithByID(web.RuntimeID))
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
			Favorite:      web.Favorite,
			IPV6:          web.IPV6,
		}

		if siteDTO.Type == constant.Subsite {
			parentWeb, _ := websiteRepo.GetFirst(repo.WithByID(web.ParentWebsiteID))
			if parentWeb.ID != 0 {
				siteDTO.ParentSite = parentWeb.PrimaryDomain
			}
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
	websites, _ := websiteRepo.List(repo.WithOrderRuleBy("primary_domain", "ascending"))
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
		return buserr.New("ErrAliasIsExist")
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
	defaultDate, _ := time.Parse(constant.DateLayout, constant.WebsiteDefaultExpireDate)

	website := &model.Website{
		Type:           create.Type,
		Alias:          alias,
		Remark:         create.Remark,
		Status:         constant.WebRunning,
		ExpireDate:     defaultDate,
		WebsiteGroupID: create.WebsiteGroupID,
		Proxy:          create.Proxy,
		SiteDir:        "/",
		AccessLog:      true,
		ErrorLog:       true,
		IPV6:           create.IPV6,
	}

	var (
		domains       []model.WebsiteDomain
		appInstall    *model.AppInstall
		runtime       *model.Runtime
		primaryDomain string
	)
	if website.Type == constant.Stream {
		if create.StreamConfig.StreamPorts == "" {
			return buserr.New("ErrTypePortRange")
		}
		website.PrimaryDomain = create.Name
		website.Protocol = constant.ProtocolStream
		website.StreamPorts = create.StreamConfig.StreamPorts
		ports := strings.Split(create.StreamConfig.StreamPorts, ",")
		for _, port := range ports {
			portNum, _ := strconv.Atoi(port)
			if err = checkWebsitePort(nginxInstall.HttpsPort, portNum, website.Type); err != nil {
				return err
			}
		}
	} else {
		domains, _, _, err = getWebsiteDomains(create.Domains, defaultHttpPort, nginxInstall.HttpsPort, 0)
		if err != nil {
			return err
		}
		primaryDomain = domains[0].Domain
		if domains[0].Port != defaultHttpPort {
			primaryDomain = fmt.Sprintf("%s:%v", domains[0].Domain, domains[0].Port)
		}
		website.PrimaryDomain = primaryDomain
		website.Protocol = constant.ProtocolHTTP
	}

	createTask, err := task.NewTaskWithOps(website.PrimaryDomain, task.TaskCreate, task.TaskScopeWebsite, create.TaskID, 0)
	if err != nil {
		return err
	}

	if create.CreateDb {
		createDataBase := func(t *task.Task) error {
			database, _ := databaseRepo.Get(repo.WithByName(create.DbHost))
			if database.ID == 0 {
				return nil
			}
			dbConfig := create.DataBaseConfig
			switch database.Type {
			case constant.AppPostgresql, constant.AppPostgres:
				oldPostgresqlDb, _ := postgresqlRepo.Get(repo.WithByName(create.DbName), repo.WithByFrom(constant.ResourceLocal))
				if oldPostgresqlDb.ID > 0 {
					return buserr.New("ErrDbUserNotValid")
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
				oldMysqlDb, _ := mysqlRepo.Get(repo.WithByName(dbConfig.DbName), repo.WithByFrom(constant.ResourceLocal))
				if oldMysqlDb.ID > 0 {
					return buserr.New("ErrDbUserNotValid")
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
			install, err = NewIAppService().Install(req, true)
			if err != nil {
				return err
			}
			appInstall = install
			website.AppInstallID = install.ID
			website.Proxy = fmt.Sprintf("127.0.0.1:%d", appInstall.HttpPort)
		} else {
			var install model.AppInstall
			install, err = appInstallRepo.GetFirst(repo.WithByID(create.AppInstallID))
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
		runtime, err = runtimeRepo.GetFirst(context.Background(), repo.WithByID(create.RuntimeID))
		if err != nil {
			return err
		}
		website.RuntimeID = runtime.ID

		switch runtime.Type {
		case constant.RuntimePHP:
			if runtime.Resource == constant.ResourceAppstore {
				if !checkImageLike(nil, runtime.Image) {
					return buserr.WithName("ErrImageNotExist", runtime.Name)
				}
				website.Proxy = fmt.Sprintf("127.0.0.1:%s", runtime.Port)
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
			proxyPort := runtime.Port
			if proxyPort == "" {
				return buserr.New("ErrRuntimeNoPort")
			}
			proxyPort = strconv.Itoa(create.Port)
			website.Proxy = fmt.Sprintf("127.0.0.1:%s", proxyPort)
		}
	case constant.Subsite:
		parentWebsite, err := websiteRepo.GetFirst(repo.WithByID(create.ParentWebsiteID))
		if err != nil {
			return err
		}
		website.ParentWebsiteID = parentWebsite.ID
		website.SiteDir = create.SiteDir
	}

	configNginx := func(t *task.Task) error {
		if err = configDefaultNginx(website, domains, appInstall, runtime, create.StreamConfig); err != nil {
			return err
		}
		if website.Type != constant.Stream {
			if err = createWafConfig(website, domains); err != nil {
				return err
			}
			if create.Type == constant.Runtime {
				runtime, err = runtimeRepo.GetFirst(context.Background(), repo.WithByID(create.RuntimeID))
				if err != nil {
					return err
				}
				if runtime.Type == constant.RuntimePHP && runtime.Resource == constant.ResourceAppstore {
					createOpenBasedirConfig(website)
				}
			}
		}

		tx, ctx := helper.GetTxAndContext()
		defer tx.Rollback()
		if err = websiteRepo.Create(ctx, website); err != nil {
			return err
		}
		t.Task.ResourceID = website.ID
		if len(domains) > 0 {
			for i := range domains {
				domains[i].WebsiteID = website.ID
			}
			if err = websiteDomainRepo.BatchCreate(ctx, domains); err != nil {
				return err
			}
		}

		tx.Commit()
		return nil
	}

	deleteWebsite := func(t *task.Task) {
		_ = deleteWebsiteFolder(website)
	}

	createTask.AddSubTask(i18n.GetMsgByKey("ConfigOpenresty"), configNginx, deleteWebsite)

	if create.EnableSSL {
		enableSSL := func(t *task.Task) error {
			websiteModel, err := websiteSSLRepo.GetFirst(repo.WithByID(create.WebsiteSSLID))
			if err != nil {
				return err
			}
			website.Protocol = constant.ProtocolHTTPS
			website.WebsiteSSLID = create.WebsiteSSLID
			appSSLReq := request.WebsiteHTTPSOp{
				WebsiteID:             website.ID,
				Enable:                true,
				WebsiteSSLID:          websiteModel.ID,
				Type:                  "existed",
				HttpConfig:            "HTTPToHTTPS",
				SSLProtocol:           []string{"TLSv1.3", "TLSv1.2"},
				Algorithm:             "ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA256:!aNULL:!eNULL:!EXPORT:!DSS:!DES:!RC4:!3DES:!MD5:!PSK:!KRB5:!SRP:!CAMELLIA:!SEED",
				Hsts:                  true,
				HstsIncludeSubDomains: true,
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

	if len(create.FtpUser) != 0 && len(create.FtpPassword) != 0 {
		createFtpUser := func(t *task.Task) error {
			indexDir := GetSitePath(*website, SiteIndexDir)
			itemID, err := NewIFtpService().Create(dto.FtpCreate{User: create.FtpUser, Password: create.FtpPassword, Path: indexDir})
			if err != nil {
				return err
			}
			website.FtpID = itemID
			return nil
		}
		createTask.AddSubTaskWithIgnoreErr(i18n.GetWithName("ConfigFTP", create.FtpUser), createFtpUser)
	}

	return createTask.Execute()
}

func (w WebsiteService) OpWebsite(req request.WebsiteOp) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
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
		options = append(options, repo.WithTypes(req.Types))
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
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if website.IPV6 != req.IPV6 {
		if err := changeIPV6(website, req.IPV6); err != nil {
			return err
		}
	}
	website.PrimaryDomain = req.PrimaryDomain
	if req.WebsiteGroupID != 0 {
		website.WebsiteGroupID = req.WebsiteGroupID
	}
	website.Remark = req.Remark
	website.IPV6 = req.IPV6
	website.Favorite = req.Favorite

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
	website, err := websiteRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return res, err
	}
	res.Website = website
	res.ErrorLogPath = GetSitePath(website, SiteErrorLog)
	res.AccessLogPath = GetSitePath(website, SiteAccessLog)
	res.SitePath = GetSitePath(website, SiteDir)
	res.SiteDir = website.SiteDir
	fileOp := files.NewFileOp()
	switch website.Type {
	case constant.Runtime:
		runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
		if err != nil {
			return res, err
		}
		res.RuntimeType = runtime.Type
		res.RuntimeName = runtime.Name
		if runtime.Type == constant.RuntimePHP {
			res.OpenBaseDir = fileOp.Stat(path.Join(GetSitePath(website, SiteIndexDir), ".user.ini"))
		}
	case constant.Stream:
		nginxParser, err := parser.NewParser(GetSitePath(website, StreamConf))
		if err != nil {
			return res, err
		}
		config, err := nginxParser.Parse()
		if err != nil {
			return res, err
		}
		listens := config.FindDirectives("listen")
		for _, listen := range listens {
			params := listen.GetParameters()
			if len(params) > 1 && params[1] == "udp" {
				res.UDP = true
			}
		}
		upstreams := config.FindUpstreams()
		for _, up := range upstreams {
			directives := up.GetDirectives()
			for _, d := range directives {
				dName := d.GetName()
				if _, ok := dto.LBAlgorithms[dName]; ok {
					res.Algorithm = dName
					break
				}
			}
			res.Servers = getNginxUpstreamServers(up.UpstreamServers)
		}
	}
	return res, nil
}

func (w WebsiteService) DeleteWebsite(req request.WebsiteDelete) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
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

	if website.Type == constant.Runtime && req.DeleteDB && website.DbID != 0 {
		switch website.DbType {
		case constant.AppMysql, constant.AppMariaDB:
			mysqlDB, _ := mysqlRepo.Get(repo.WithByID(website.DbID))
			if mysqlDB.ID > 0 {
				deleteReq := dto.MysqlDBDelete{
					ID:          mysqlDB.ID,
					Database:    mysqlDB.MysqlName,
					ForceDelete: req.ForceDelete,
				}
				if err = NewIMysqlService().Delete(context.TODO(), deleteReq); err != nil && !req.ForceDelete {
					return err
				}
			}
		case constant.AppPostgresql, constant.AppPostgres:
			pgDB, _ := postgresqlRepo.Get(repo.WithByID(website.DbID))
			if pgDB.ID > 0 {
				deleteReq := dto.PostgresqlDBDelete{
					ID:          pgDB.ID,
					ForceDelete: req.ForceDelete,
					Database:    pgDB.PostgresqlName,
				}
				if err = NewIPostgresqlService().Delete(context.TODO(), deleteReq); err != nil && !req.ForceDelete {
					return err
				}
			}
		}
	}

	if err = delNginxConfig(website, req.ForceDelete); err != nil {
		return err
	}

	if website.Type != constant.Stream {
		if err = delWafConfig(website, req.ForceDelete); err != nil {
			return err
		}
	}

	if checkIsLinkApp(website) && req.DeleteApp {
		appInstall, _ := appInstallRepo.GetFirst(repo.WithByID(website.AppInstallID))
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
		_ = NewIBackupRecordService().DeleteRecordByName("website", website.PrimaryDomain, website.Alias, req.DeleteBackup)
	}()

	if err := websiteRepo.DeleteBy(ctx, repo.WithByID(req.ID)); err != nil {
		return err
	}
	if err := websiteDomainRepo.DeleteBy(ctx, websiteDomainRepo.WithWebsiteId(req.ID)); err != nil {
		return err
	}
	tx.Commit()

	uploadDir := path.Join(global.Dir.DataDir, "uploads/website", website.Alias)
	if _, err := os.Stat(uploadDir); err == nil {
		_ = os.RemoveAll(uploadDir)
	}
	return nil
}

func (w WebsiteService) UpdateWebsiteDomain(req request.WebsiteDomainUpdate) error {
	domain, err := websiteDomainRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	domain.SSL = req.SSL
	website, err := websiteRepo.GetFirst(repo.WithByID(domain.WebsiteID))
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

func (w WebsiteService) GetNginxConfigByScope(req request.NginxScopeReq) (*response.WebsiteNginxConfig, error) {
	keys, ok := dto.ScopeKeyMap[req.Scope]
	if !ok || len(keys) == 0 {
		return nil, nil
	}

	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
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
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
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
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
	if err != nil {
		return nil, err
	}
	configPath := ""
	switch configType {
	case constant.AppOpenresty:
		configPath = GetWebsiteConfigPath(website)
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
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteId))
	if err != nil {
		return response.WebsiteHTTPS{}, err
	}
	var (
		res        response.WebsiteHTTPS
		httpsPorts []string
	)

	httpsPortsMap := getHttpsPort(websiteId)
	for port := range httpsPortsMap {
		httpsPorts = append(httpsPorts, strconv.Itoa(port))
	}
	res.HttpsPort = strings.Join(httpsPorts, ",")
	if website.WebsiteSSLID == 0 {
		res.Enable = false
		return res, nil
	}
	websiteSSL, err := websiteSSLRepo.GetFirst(repo.WithByID(website.WebsiteSSLID))
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
				if len(p.Params) > 1 {
					hstsValue := p.Params[1]
					if strings.Contains(hstsValue, "includeSubDomains") {
						res.HstsIncludeSubDomains = true
					}
				}
			}
			if p.Params[0] == "Alt-Svc" {
				res.Http3 = true
			}
		}
	}
	return res, nil
}

func (w WebsiteService) OpWebsiteHTTPS(ctx context.Context, req request.WebsiteHTTPSOp) (*response.WebsiteHTTPS, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return nil, err
	}
	var (
		res        response.WebsiteHTTPS
		websiteSSL model.WebsiteSSL
	)
	if err = ChangeHSTSConfig(req.Hsts, req.HstsIncludeSubDomains, req.Http3, website); err != nil {
		return nil, err
	}
	res.Enable = req.Enable
	res.SSLProtocol = req.SSLProtocol
	res.Algorithm = req.Algorithm
	res.HstsIncludeSubDomains = req.HstsIncludeSubDomains
	if !req.Enable {
		website.Protocol = constant.ProtocolHTTP
		website.WebsiteSSLID = 0

		websiteDomains, _ := websiteDomainRepo.GetBy(websiteDomainRepo.WithWebsiteId(website.ID))

		ports := make(map[int]struct{})
		for _, domain := range websiteDomains {
			ports[domain.Port] = struct{}{}
		}
		for port := range ports {
			if err = removeSSLListen(website, []string{strconv.Itoa(port)}); err != nil {
				return nil, err
			}
		}
		nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
		if err != nil {
			return nil, err
		}
		if _, ok := ports[nginxInstall.HttpsPort]; !ok {
			httpsPortStr := strconv.Itoa(nginxInstall.HttpsPort)
			if err = deleteListenAndServerName(website, []string{httpsPortStr, "[::]:" + httpsPortStr}, []string{}); err != nil {
				return nil, err
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
			dto.NginxParam{
				Name:   "add_header",
				Params: []string{"Strict-Transport-Security"},
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
		websiteModel, err := websiteSSLRepo.GetFirst(repo.WithByID(req.WebsiteSSLID))
		if err != nil {
			return nil, err
		}
		if websiteModel.Pem == "" {
			return nil, buserr.New("ErrSSLValid")
		}
		website.WebsiteSSLID = websiteModel.ID
		res.SSL = *websiteModel
		websiteSSL = *websiteModel
	}
	if req.Type == constant.SSLManual {
		websiteSSL, err = getManualWebsiteSSL(req)
		if err != nil {
			return nil, err
		}
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
			Status:  buserr.WithDetail("ErrNotInstall", app.Name, nil).Error(),
			Version: appInstall.Version,
		})
		showErr = true
	} else {
		checkIds = append(req.InstallIds, appInstall.ID)
	}
	if len(checkIds) > 0 {
		installList, _ := appInstallRepo.ListBy(context.Background(), repo.WithByIDs(checkIds))
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
			if install.Status != constant.StatusRunning {
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
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return err
	}

	filePath := nginxFull.SiteConfig.FilePath
	if err = files.NewFileOp().WriteFile(filePath, strings.NewReader(req.Content), constant.DirPerm); err != nil {
		return err
	}
	return nginxCheckAndReload(nginxFull.SiteConfig.OldContent, filePath, nginxFull.Install.ContainerName)
}

func (w WebsiteService) OpWebsiteLog(req request.WebsiteLogReq) (*response.WebsiteLog, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return nil, err
	}
	sitePath := GetSitePath(website, SiteDir)
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
		logFileRes, err := files.ReadFileByLine(filePath, req.Page, req.PageSize, false)
		if err != nil {
			return nil, err
		}
		res.End = logFileRes.IsEndOfFile
		res.Path = filePath
		res.Content = strings.Join(logFileRes.Lines, "\n")
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
			if website.Type != constant.Stream {
				params = append(params, "main")
			} else {
				params = append(params, "streamlog")
			}
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
		logPath := path.Join(sitePath, "log", req.LogType)
		if err := files.NewFileOp().WriteFile(logPath, strings.NewReader(""), constant.DirPerm); err != nil {
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
	if err := updateDefaultServerConfig(!(id > 0)); err != nil {
		return err
	}
	if id == 0 {
		return nil
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return err
	}
	params, err := getNginxParamsByKeys(constant.NginxScopeServer, []string{"listen"}, &website)
	if err != nil {
		return err
	}
	var changeParams []dto.NginxParam
	for _, param := range params {
		if hasHttp3(param.Params) || hasDefaultServer(param.Params) {
			continue
		}
		newParam := append(param.Params, components.DefaultServer)
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

func (w WebsiteService) ChangePHPVersion(req request.WebsitePHPVersionReq) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if website.Type == constant.Runtime {
		oldRuntime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
		if err != nil {
			return err
		}
		if oldRuntime.Resource == constant.ResourceLocal {
			return buserr.New("ErrPHPResource")
		}
		client, err := docker.NewDockerClient()
		if err != nil {
			return err
		}
		defer client.Close()
		if !checkImageLike(client, oldRuntime.Image) {
			return buserr.WithName("ErrImageNotExist", oldRuntime.Name)
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
	fileOp := files.NewFileOp()
	indexPHPPath := path.Join(GetSitePath(website, SiteIndexDir), "index.php")
	indexHtmlPath := path.Join(GetSitePath(website, SiteIndexDir), "index.html")
	if req.RuntimeID > 0 {
		server.UpdateDirective("index", []string{"index.php index.html index.htm default.php default.htm default.html"})
		server.RemoveDirective("location", []string{"~", "[^/]\\.php(/|$)"})
		runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(req.RuntimeID))
		if err != nil {
			return err
		}
		if runtime.Resource == constant.ResourceLocal {
			return buserr.New("ErrPHPResource")
		}
		website.RuntimeID = req.RuntimeID
		website.AppInstallID = 0
		phpProxy := fmt.Sprintf("127.0.0.1:%s", runtime.Port)
		website.Proxy = phpProxy
		server.UpdatePHPProxy([]string{website.Proxy}, "")
		website.Type = constant.Runtime
		if !fileOp.Stat(indexPHPPath) {
			_ = fileOp.WriteFile(indexPHPPath, strings.NewReader(string(nginx_conf.IndexPHP)), constant.FilePerm)
		}
	} else {
		server.UpdateDirective("index", []string{"index.html index.php index.htm default.php default.htm default.html"})
		website.RuntimeID = 0
		website.Type = constant.Static
		website.Proxy = ""
		server.RemoveDirective("location", []string{"~", "[^/]\\.php(/|$)"})
		if !fileOp.Stat(indexHtmlPath) {
			_ = fileOp.WriteFile(indexHtmlPath, strings.NewReader(string(nginx_conf.Index)), constant.FilePerm)
		}
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

func (w WebsiteService) UpdateSiteDir(req request.WebsiteUpdateDir) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
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
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	absoluteIndexPath := GetSitePath(website, SiteIndexDir)
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(10 * time.Second))
	if err := cmdMgr.RunBashCf("%s chown -R %s:%s %s", cmd.SudoHandleCmd(), req.User, req.Group, absoluteIndexPath); err != nil {
		return err
	}
	website.User = req.User
	website.Group = req.Group
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) UpdateCors(req request.CorsConfigReq) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	params := []dto.NginxParam{
		{Name: "add_header", Params: []string{"Access-Control-Allow-Origin"}},
		{Name: "add_header", Params: []string{"Access-Control-Allow-Methods"}},
		{Name: "add_header", Params: []string{"Access-Control-Allow-Headers"}},
		{Name: "add_header", Params: []string{"Access-Control-Allow-Credentials"}},
		{Name: "if", Params: []string{"(", "$request_method", "=", "'OPTIONS'", ")"}},
	}
	if err := deleteNginxConfig(constant.NginxScopeServer, params, &website); err != nil {
		return err
	}
	if req.Cors {
		return updateWebsiteConfig(website, func(server *components.Server) error {
			server.UpdateDirective("add_header", []string{"Access-Control-Allow-Origin", req.AllowOrigins, "always"})
			if req.AllowMethods != "" {
				server.UpdateDirective("add_header", []string{"Access-Control-Allow-Methods", req.AllowMethods, "always"})
			}
			if req.AllowHeaders != "" {
				server.UpdateDirective("add_header", []string{"Access-Control-Allow-Headers", req.AllowHeaders, "always"})
			}
			if req.AllowCredentials {
				server.UpdateDirective("add_header", []string{"Access-Control-Allow-Credentials", "true", "always"})
			}
			if req.Preflight {
				server.AddCorsOption()
			}
			return nil
		})
	}
	return nil
}

func (w WebsiteService) GetCors(websiteID uint) (*request.CorsConfig, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
	if err != nil {
		return nil, err
	}
	server, err := getServer(website)
	if err != nil {
		return nil, err
	}
	if server == nil {
		return nil, nil
	}
	cors := &request.CorsConfig{
		Cors:             server.Cors,
		AllowOrigins:     server.AllowOrigins,
		AllowMethods:     server.AllowMethods,
		AllowHeaders:     server.AllowHeaders,
		AllowCredentials: server.AllowCredentials,
		Preflight:        server.Preflight,
	}
	return cors, nil
}

func (w WebsiteService) UpdateAntiLeech(req request.NginxAntiLeechUpdate) (err error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
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
	if req.Enable || req.Cache {
		exts := strings.Split(req.Extends, ",")
		newDirective := components.Directive{
			Name:       "location",
			Parameters: []string{"~", fmt.Sprintf(`.*\.(%s)$`, strings.Join(exts, "|"))},
		}

		newBlock := &components.Block{}
		newBlock.Directives = make([]components.IDirective, 0)
		if req.Cache {
			newBlock.AppendDirectives(&components.Directive{
				Name:       "expires",
				Parameters: []string{strconv.Itoa(req.CacheTime) + req.CacheUint},
			})
		}
		if !req.LogEnable {
			newBlock.AppendDirectives(&components.Directive{
				Name:       "access_log",
				Parameters: []string{"off"},
			})
		}
		newBlock.AppendDirectives(&components.Directive{
			Name:       "log_not_found",
			Parameters: []string{"off"},
		})
		if req.Enable {
			validDir := &components.Directive{
				Name:       "valid_referers",
				Parameters: []string{},
			}
			if req.NoneRef {
				validDir.Parameters = append(validDir.Parameters, "none")
			}
			if req.Blocked {
				validDir.Parameters = append(validDir.Parameters, "blocked")
			}
			if len(req.ServerNames) > 0 {
				validDir.Parameters = append(validDir.Parameters, strings.Join(req.ServerNames, " "))
			}
			newBlock.AppendDirectives(validDir)

			ifDir := &components.Directive{
				Name:       "if",
				Parameters: []string{"($invalid_referer)"},
			}
			if !req.LogEnable {
				ifDir.Block = &components.Block{
					Directives: []components.IDirective{
						&components.Directive{
							Name:       "access_log",
							Parameters: []string{"off"},
						},
						&components.Directive{
							Name:       "return",
							Parameters: []string{req.Return},
						},
					},
				}
			} else {
				ifDir.Block = &components.Block{
					Directives: []components.IDirective{
						&components.Directive{
							Name:       "return",
							Parameters: []string{req.Return},
						},
					},
				}
			}
			newBlock.AppendDirectives(ifDir)
		}
		if website.Type == constant.Deployment {
			newBlock.AppendDirectives(
				&components.Directive{
					Name:       "proxy_set_header",
					Parameters: []string{"Host", "$host"},
				},
				&components.Directive{
					Name:       "proxy_set_header",
					Parameters: []string{"X-Real-IP", "$remote_addr"},
				},
				&components.Directive{
					Name:       "proxy_set_header",
					Parameters: []string{"X-Forwarded-For", "$proxy_add_x_forwarded_for"},
				},
				&components.Directive{
					Name:       "proxy_pass",
					Parameters: []string{fmt.Sprintf("http://%s", website.Proxy)},
				})
		}
		newDirective.Block = newBlock
		index := -1
		for i, directive := range block.Directives {
			if directive.GetName() == "include" {
				index = i
				break
			}
		}
		if index != -1 {
			block.Directives = append(block.Directives[:index], append([]components.IDirective{&newDirective}, block.Directives[index:]...)...)
		} else {
			block.Directives = append(block.Directives, &newDirective)
		}
	}

	if err = nginx.WriteConfig(nginxFull.SiteConfig.Config, nginx.IndentedStyle); err != nil {
		return
	}
	if err = updateNginxConfig(constant.NginxScopeServer, nil, &website); err != nil {
		_ = fileOp.WriteFile(nginxFull.SiteConfig.Config.FilePath, bytes.NewReader(backupContent), constant.DirPerm)
		return
	}
	return
}

func (w WebsiteService) GetAntiLeech(id uint) (*response.NginxAntiLeechRes, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(id))
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
			if lDir.GetName() == "access_log" {
				if strings.Join(lDir.GetParameters(), "") == "off" {
					res.LogEnable = false
				}
			}
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
				}
			}
			if lDir.GetName() == "expires" {
				res.Cache = true
				matches := re.GetRegex(re.NumberWordPattern).FindStringSubmatch(lDir.GetParameters()[0])
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

	website, err = websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	includeDir := GetSitePath(website, SiteRedirectDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		_ = fileOp.CreateDir(includeDir, constant.DirPerm)
	}
	fileName := fmt.Sprintf("%s.conf", req.Name)
	includePath := path.Join(includeDir, fileName)
	backName := fmt.Sprintf("%s.bak", req.Name)
	backPath := path.Join(includeDir, backName)

	if req.Operate == "create" && (fileOp.Stat(includePath) || fileOp.Stat(backPath)) {
		err = buserr.New("ErrNameIsExist")
		return
	}

	defer func() {
		if err != nil {
			switch req.Operate {
			case "create":
				_ = fileOp.DeleteFile(includePath)
			case "edit":
				_ = fileOp.WriteFile(includePath, bytes.NewReader(oldContent), constant.DirPerm)
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
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}

	nginxInclude := fmt.Sprintf("/www/sites/%s/redirect/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return
	}
	return
}

func (w WebsiteService) GetRedirect(id uint) (res []response.NginxRedirectConfig, err error) {
	var (
		website model.Website
	)
	website, err = websiteRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return
	}
	includeDir := GetSitePath(website, SiteRedirectDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		return
	}
	entries, err := os.ReadDir(includeDir)
	if err != nil {
		return
	}
	if len(entries) == 0 {
		return
	}

	var (
		content []byte
		config  *components.Config
	)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		fileName := entry.Name()
		if !strings.HasSuffix(fileName, ".conf") && !strings.HasSuffix(fileName, ".bak") {
			continue
		}
		redirectConfig := response.NginxRedirectConfig{
			WebsiteID: website.ID,
		}
		parts := strings.Split(fileName, ".")
		redirectConfig.Name = parts[0]
		if parts[1] == "conf" {
			redirectConfig.Enable = true
		} else {
			redirectConfig.Enable = false
		}
		filePath := path.Join(includeDir, fileName)
		redirectConfig.FilePath = filePath
		content, err = fileOp.GetContent(filePath)
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
		oldRewriteContent []byte
	)
	website, err = websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	absolutePath := path.Join(GetSitePath(website, SiteRedirectDir), req.Name+".conf")
	fileOp := files.NewFileOp()
	oldRewriteContent, err = fileOp.GetContent(absolutePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(absolutePath, strings.NewReader(req.Content), constant.DirPerm); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = fileOp.WriteFile(absolutePath, bytes.NewReader(oldRewriteContent), constant.DirPerm)
		}
	}()
	return updateNginxConfig(constant.NginxScopeServer, nil, &website)
}

func (w WebsiteService) LoadWebsiteDirConfig(req request.WebsiteCommonReq) (*response.WebsiteDirConfig, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
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
	checkAndAppendDirs := func(relPath string, entries []os.DirEntry) {
		for _, entry := range entries {
			if !entry.IsDir() || entry.Name() == "node_modules" || entry.Name() == "vendor" {
				continue
			}
			nextRelPath := path.Join(relPath, entry.Name())
			res.Dirs = append(res.Dirs, "/"+strings.TrimPrefix(nextRelPath, "/"))
			entryInfo, _ := entry.Info()
			if entryInfo != nil {
				if stat, ok := entryInfo.Sys().(*syscall.Stat_t); ok {
					if stat.Uid != 1000 || stat.Gid != 1000 {
						res.Msg = i18n.GetMsgByKey("ErrPathPermission")
					}
				}
			}
		}
	}

	checkAndAppendDirs("", indexFiles)
	for _, firstDir := range indexFiles {
		if !firstDir.IsDir() || firstDir.Name() == "node_modules" || firstDir.Name() == "vendor" {
			continue
		}
		secondLevelPath := path.Join(absoluteIndexPath, firstDir.Name())
		secondLevelDirs, _ := os.ReadDir(secondLevelPath)
		checkAndAppendDirs(firstDir.Name(), secondLevelDirs)

		for _, secondDir := range secondLevelDirs {
			if !secondDir.IsDir() || secondDir.Name() == "node_modules" || secondDir.Name() == "vendor" {
				continue
			}
			thirdLevelPath := path.Join(secondLevelPath, secondDir.Name())
			thirdLevelDirs, _ := os.ReadDir(thirdLevelPath)
			checkAndAppendDirs(path.Join(firstDir.Name(), secondDir.Name()), thirdLevelDirs)
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
		_ = fileOp.CreateDir(defaultPath, constant.DirPerm)
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
		_ = fileOp.CreateDir(defaultPath, constant.DirPerm)
	}
	var resourcePath string
	switch req.Type {
	case "404":
		resourcePath = path.Join(defaultPath, "404.html")
		if req.Sync {
			websites, _ := websiteRepo.GetBy(repo.WithTypes([]string{constant.Static, constant.Runtime}))
			for _, website := range websites {
				filePath := path.Join(GetSitePath(website, SiteIndexDir), "404.html")
				if fileOp.Stat(filePath) {
					_ = fileOp.SaveFile(filePath, req.Content, constant.DirPerm)
				}
			}
		}
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
	return fileOp.SaveFile(resourcePath, req.Content, constant.DirPerm)
}

func (w WebsiteService) ChangeGroup(group, newGroup uint) error {
	return websiteRepo.UpdateGroup(group, newGroup)
}

func (w WebsiteService) SetRealIPConfig(req request.WebsiteRealIP) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
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
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
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

func (w WebsiteService) GetWebsiteResource(websiteID uint) ([]response.Resource, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(websiteID))
	if err != nil {
		return nil, err
	}
	var (
		res          []response.Resource
		databaseID   uint
		databaseType string
	)
	if website.Type == constant.Runtime {
		runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
		if err != nil {
			return nil, err
		}
		res = append(res, response.Resource{
			Name:       runtime.Name,
			Type:       "runtime",
			ResourceID: runtime.ID,
			Detail:     runtime,
		})
	}
	if website.Type == constant.Deployment {
		install, err := appInstallRepo.GetFirst(repo.WithByID(website.AppInstallID))
		if err != nil {
			return nil, err
		}
		res = append(res, response.Resource{
			Name:       install.Name,
			Type:       "app",
			ResourceID: install.ID,
			Detail:     install,
		})
		installResources, _ := appInstallResourceRepo.GetBy(appInstallResourceRepo.WithAppInstallId(install.ID))
		for _, resource := range installResources {
			if resource.Key == constant.AppMysql || resource.Key == constant.AppMariaDB || resource.Key == constant.AppPostgres || resource.Key == constant.AppPostgresql {
				databaseType = resource.Key
				databaseID = resource.ResourceId
			}
		}
	}
	if website.DbID > 0 {
		databaseType = website.DbType
		databaseID = website.DbID
	}
	if databaseID > 0 {
		switch databaseType {
		case constant.AppMysql, constant.AppMariaDB:
			db, _ := mysqlRepo.Get(repo.WithByID(databaseID))
			if db.ID > 0 {
				res = append(res, response.Resource{
					Name:       db.Name,
					Type:       "database",
					ResourceID: db.ID,
					Detail:     db,
				})
			}
		case constant.AppPostgresql, constant.AppPostgres:
			db, _ := postgresqlRepo.Get(repo.WithByID(databaseID))
			if db.ID > 0 {
				res = append(res, response.Resource{
					Name:       db.Name,
					Type:       "database",
					ResourceID: db.ID,
					Detail:     db,
				})
			}
		}
	}

	return res, nil
}

func (w WebsiteService) ListDatabases() ([]response.Database, error) {
	var res []response.Database
	mysqlDBs, _ := mysqlRepo.List()
	for _, db := range mysqlDBs {
		database, _ := databaseRepo.Get(repo.WithByName(db.MysqlName))
		if database.ID > 0 {
			res = append(res, response.Database{
				ID:           db.ID,
				Name:         db.Name,
				Type:         database.Type,
				From:         database.From,
				DatabaseName: database.Name,
			})
		}
	}
	pgSqls, _ := postgresqlRepo.List()
	for _, db := range pgSqls {
		database, _ := databaseRepo.Get(repo.WithByName(db.PostgresqlName))
		if database.ID > 0 {
			res = append(res, response.Database{
				ID:           db.ID,
				Name:         db.Name,
				Type:         database.Type,
				From:         database.From,
				DatabaseName: database.Name,
			})
		}
	}
	return res, nil
}

func (w WebsiteService) ChangeDatabase(req request.ChangeDatabase) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if website.DbID == req.DatabaseID {
		return nil
	}
	website.DbID = req.DatabaseID
	website.DbType = req.DatabaseType
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) OperateCrossSiteAccess(req request.CrossSiteAccessOp) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if req.Operation == constant.StatusEnable {
		createOpenBasedirConfig(&website)
	}
	if req.Operation == constant.StatusDisable {
		fileOp := files.NewFileOp()
		return fileOp.DeleteFile(path.Join(GetSitePath(website, SiteIndexDir), ".user.ini"))
	}
	return nil
}

func (w WebsiteService) ExecComposer(req request.ExecComposerReq) error {
	if cmd.CheckIllegal(req.User, req.Mirror, req.Command, req.ExtCommand) {
		return buserr.New("ErrCmdIllegal")
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	sitePath := GetSitePath(website, SiteDir)
	if !strings.Contains(req.Dir, sitePath) {
		return buserr.New("ErrWebsiteDir")
	}
	if !files.NewFileOp().Stat(path.Join(req.Dir, "composer.json")) {
		return buserr.New("ErrComposerFileNotFound")
	}
	if task.CheckResourceTaskIsExecuting(task.TaskExec, req.Command, website.ID) {
		return buserr.New("ErrInstallExtension")
	}
	runtime, err := runtimeRepo.GetFirst(context.Background(), repo.WithByID(website.RuntimeID))
	if err != nil {
		return err
	}
	var command string
	if req.Command != "custom" {
		command = fmt.Sprintf("%s %s", req.Command, req.ExtCommand)
	} else {
		command = req.ExtCommand
	}
	resourceName := fmt.Sprintf("composer %s", command)
	composerTask, err := task.NewTaskWithOps(resourceName, task.TaskExec, req.Command, req.TaskID, website.ID)
	if err != nil {
		return err
	}
	cmdMgr := cmd.NewCommandMgr(cmd.WithTask(*composerTask), cmd.WithTimeout(20*time.Minute))
	siteDir, _ := settingRepo.Get(settingRepo.WithByKey("WEBSITE_DIR"))
	execDir := strings.ReplaceAll(req.Dir, siteDir.Value, "/www")
	composerTask.AddSubTask("", func(t *task.Task) error {
		cmdStr := fmt.Sprintf("docker exec -u %s %s sh -c 'composer config -g repo.packagist composer %s && composer %s --working-dir=%s'", req.User, runtime.ContainerName, req.Mirror, command, execDir)
		err = cmdMgr.RunBashC(cmdStr)
		if err != nil {
			return err
		}
		return nil
	}, nil)
	go func() {
		_ = composerTask.Execute()
	}()
	return nil
}

func (w WebsiteService) UpdateStream(req request.StreamUpdate) error {
	if req.StreamConfig.StreamPorts == "" {
		return buserr.New("ErrTypePortRange")
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxFull, err := getNginxFull(&website)
	if err != nil {
		return nil
	}
	website.StreamPorts = req.StreamConfig.StreamPorts
	ports := strings.Split(req.StreamConfig.StreamPorts, ",")
	for _, port := range ports {
		portNum, _ := strconv.Atoi(port)
		if err = checkWebsitePort(nginxFull.Install.HttpsPort, portNum, website.Type); err != nil {
			return err
		}
	}

	config := nginxFull.SiteConfig.Config
	servers := config.FindServers()
	if len(servers) == 0 {
		return errors.New("nginx config is not valid")
	}
	server := servers[0]
	server.Listens = []*components.ServerListen{}
	var params []string
	if req.UDP {
		params = []string{"udp"}
	}
	for _, port := range ports {
		server.UpdateListen(port, false, params...)
		if website.IPV6 {
			server.UpdateListen("[::]:"+port, false, params...)
		}
	}
	upstream := components.Upstream{
		UpstreamName: website.Alias,
	}
	if req.Algorithm != "default" {
		upstream.UpdateDirective(req.Algorithm, []string{})
	}
	upstream.UpstreamServers = parseUpstreamServers(req.Servers)
	for i, dir := range config.Block.Directives {
		if dir.GetName() == "upstream" && dir.GetParameters()[0] == website.Alias {
			config.Block.Directives[i] = &upstream
		}
	}

	if err := nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return err
	}
	if err = nginxCheckAndReload(nginxFull.SiteConfig.OldContent, config.FilePath, nginxFull.Install.ContainerName); err != nil {
		return err
	}
	website.StreamPorts = req.StreamConfig.StreamPorts
	return websiteRepo.Save(context.Background(), &website)
}
