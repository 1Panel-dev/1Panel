package service

import (
	"bytes"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
)

func (w WebsiteService) OperateProxy(req request.WebsiteProxyConfig) (err error) {
	var (
		website    model.Website
		par        *parser.Parser
		oldContent []byte
	)

	website, err = websiteRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	includeDir := GetSitePath(website, SiteProxyDir)
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

	var config *components.Config

	switch req.Operate {
	case "create":
		config, err = parser.NewStringParser(string(nginx_conf.GetWebsiteFile("proxy.conf"))).Parse()
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

	var location *components.Location
	for _, directive := range directives {
		if loc, ok := directive.(*components.Location); ok {
			location = loc
			break
		}
	}
	if location == nil {
		err = errors.New("invalid proxy config, no location found")
		return
	}
	location.UpdateDirective("proxy_pass", []string{req.ProxyPass})
	location.UpdateDirective("proxy_set_header", []string{"Host", req.ProxyHost})
	location.ChangePath(req.Modifier, req.Match)
	// Server Cache Settings
	if req.Cache {
		if err = openProxyCache(website); err != nil {
			return
		}
		location.AddServerCache(fmt.Sprintf("proxy_cache_zone_of_%s", website.Alias), req.ServerCacheTime, req.ServerCacheUnit)
	} else {
		location.RemoveServerCache(fmt.Sprintf("proxy_cache_zone_of_%s", website.Alias))
	}
	// Browser Cache Settings
	// cacheTime > 0: enable expires;
	// cacheTime == 0: use upstream cache-control, remove self-set;
	// cacheTime < 0: force no-cache
	if req.CacheTime > 0 {
		location.AddBrowserCache(req.CacheTime, req.CacheUnit)
	} else if req.CacheTime < 0 {
		location.AddBroswerNoCache()
	} else {
		location.RemoveBrowserCache()
	}
	// Content Replace Settings
	if len(req.Replaces) > 0 {
		location.AddSubFilter(req.Replaces)
	} else {
		location.RemoveSubFilter()
	}
	// SSL Settings
	if req.SNI {
		location.UpdateDirective("proxy_ssl_server_name", []string{"on"})
		if req.ProxySSLName != "" {
			location.UpdateDirective("proxy_ssl_name", []string{req.ProxySSLName})
		}
	} else {
		location.UpdateDirective("proxy_ssl_server_name", []string{"off"})
	}
	// CORS Settings
	if req.Cors {
		location.UpdateDirective("add_header", []string{"Access-Control-Allow-Origin", req.AllowOrigins, "always"})
		if req.AllowMethods != "" {
			location.UpdateDirective("add_header", []string{"Access-Control-Allow-Methods", req.AllowMethods, "always"})
		} else {
			location.RemoveDirective("add_header", []string{"Access-Control-Allow-Methods"})
		}
		if req.AllowHeaders != "" {
			location.UpdateDirective("add_header", []string{"Access-Control-Allow-Headers", req.AllowHeaders, "always"})
		} else {
			location.RemoveDirective("add_header", []string{"Access-Control-Allow-Headers"})
		}
		if req.AllowCredentials {
			location.UpdateDirective("add_header", []string{"Access-Control-Allow-Credentials", "true", "always"})
		} else {
			location.RemoveDirective("add_header", []string{"Access-Control-Allow-Credentials"})
		}
		if req.Preflight {
			location.AddCorsOption()
		} else {
			location.RemoveCorsOption()
		}
	} else {
		location.RemoveDirective("add_header", []string{"Access-Control-Allow-Origin"})
		location.RemoveDirective("add_header", []string{"Access-Control-Allow-Methods"})
		location.RemoveDirective("add_header", []string{"Access-Control-Allow-Headers"})
		location.RemoveDirective("add_header", []string{"Access-Control-Allow-Credentials"})
		location.RemoveDirectiveByFullParams("if", []string{"(", "$request_method", "=", "'OPTIONS'", ")"})
	}
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/proxy/*.conf", website.Alias)
	return updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website)
}

func (w WebsiteService) UpdateProxyCache(req request.NginxProxyCacheUpdate) (err error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	cacheDir := GetSitePath(website, SiteCacheDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(cacheDir) {
		_ = fileOp.CreateDir(cacheDir, constant.DirPerm)
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
	website, err = websiteRepo.GetFirst(repo.WithByID(id))
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
		if re.GetRegex(re.ProxyCacheZonePattern).MatchString(param) {
			matches := re.GetRegex(re.ProxyCacheZonePattern).FindStringSubmatch(param)
			if len(matches) > 0 {
				res.ShareCache, _ = strconv.Atoi(matches[1])
				res.ShareCacheUnit = matches[2]
			}
		}

		if re.GetRegex(re.ProxyCacheMaxSizeValidationPattern).MatchString(param) {
			matches := re.GetRegex(re.ProxyCacheMaxSizePattern).FindStringSubmatch(param)
			if len(matches) > 0 {
				res.CacheLimit, _ = strconv.ParseFloat(matches[1], 64)
				res.CacheLimitUnit = matches[2]
			}
		}
		if re.GetRegex(re.ProxyCacheInactivePattern).MatchString(param) {
			matches := re.GetRegex(re.ProxyCacheInactivePattern).FindStringSubmatch(param)
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
	website, err = websiteRepo.GetFirst(repo.WithByID(id))
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

		var location *components.Location
		for _, directive := range directives {
			if loc, ok := directive.(*components.Location); ok {
				location = loc
				break
			}
		}
		if location == nil {
			err = errors.New("invalid proxy config, no location found")
			return
		}
		proxyConfig.ProxyPass = location.ProxyPass
		proxyConfig.Cache = location.Cache
		if location.CacheTime > 0 {
			proxyConfig.CacheTime = location.CacheTime
			proxyConfig.CacheUnit = location.CacheUint
		}
		if location.ServerCacheTime > 0 {
			proxyConfig.ServerCacheTime = location.ServerCacheTime
			proxyConfig.ServerCacheUnit = location.ServerCacheUint
		}
		proxyConfig.Match = location.Match
		proxyConfig.Modifier = location.Modifier
		proxyConfig.ProxyHost = location.Host
		proxyConfig.Replaces = location.Replaces
		for _, directive := range location.Directives {
			if directive.GetName() == "proxy_ssl_server_name" {
				proxyConfig.SNI = directive.GetParameters()[0] == "on"
			}
			if directive.GetName() == "proxy_ssl_name" && len(directive.GetParameters()) > 0 {
				proxyConfig.ProxySSLName = directive.GetParameters()[0]
			}
		}
		proxyConfig.Cors = location.Cors
		proxyConfig.AllowCredentials = location.AllowCredentials
		proxyConfig.AllowHeaders = location.AllowHeaders
		proxyConfig.AllowOrigins = location.AllowOrigins
		proxyConfig.AllowMethods = location.AllowMethods
		proxyConfig.Preflight = location.Preflight
		res = append(res, proxyConfig)
	}
	return
}

func (w WebsiteService) UpdateProxyFile(req request.NginxProxyUpdate) (err error) {
	var (
		website           model.Website
		oldRewriteContent []byte
	)
	website, err = websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	absolutePath := fmt.Sprintf("%s/%s.conf", GetSitePath(website, SiteProxyDir), req.Name)
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

func (w WebsiteService) ClearProxyCache(req request.NginxCommonReq) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	cacheDir := GetSitePath(website, SiteCacheDir)
	fileOp := files.NewFileOp()
	if fileOp.Stat(cacheDir) {
		if err = fileOp.CleanDir(cacheDir); err != nil {
			return err
		}
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if err = opNginx(nginxInstall.ContainerName, constant.NginxReload); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) DeleteProxy(req request.WebsiteProxyDel) (err error) {
	fileOp := files.NewFileOp()
	website, err := websiteRepo.GetFirst(repo.WithByID(req.ID))
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
