package service

import (
	"bytes"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
	"os"
	"path"
	"strings"
)

func (w WebsiteService) GetLoadBalances(id uint) ([]dto.NginxUpstream, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(id))
	if err != nil {
		return nil, err
	}
	includeDir := GetSitePath(website, SiteUpstreamDir)
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
				upstream.Servers = getNginxUpstreamServers(up.UpstreamServers)
			}
		}
		res = append(res, upstream)
	}
	return res, nil
}

func (w WebsiteService) CreateLoadBalance(req request.WebsiteLBCreate) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	includeDir := GetSitePath(website, SiteUpstreamDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(includeDir) {
		_ = fileOp.CreateDir(includeDir, constant.DirPerm)
	}
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	if fileOp.Stat(filePath) {
		return buserr.New("ErrNameIsExist")
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
	upstream.UpstreamServers = parseUpstreamServers(req.Servers)
	config.Block.Directives = append(config.Block.Directives, &upstream)

	defer func() {
		if err != nil {
			_ = fileOp.DeleteFile(filePath)
		}
	}()

	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/upstream/*.conf", website.Alias)
	if err = updateNginxConfig("", []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return err
	}
	return nil
}

func (w WebsiteService) UpdateLoadBalance(req request.WebsiteLBUpdate) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	includeDir := GetSitePath(website, SiteUpstreamDir)
	fileOp := files.NewFileOp()
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	if !fileOp.Stat(filePath) {
		return nil
	}
	oldContent, err := fileOp.GetContent(filePath)
	if err != nil {
		return err
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
			up.UpstreamServers = parseUpstreamServers(req.Servers)
		}
	}
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	return nginxCheckAndReload(string(oldContent), filePath, nginxInstall.ContainerName)
}

func (w WebsiteService) DeleteLoadBalance(req request.WebsiteLBDelete) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	proxies, _ := w.GetProxies(website.ID)
	if len(proxies) > 0 {
		for _, proxy := range proxies {
			if strings.HasSuffix(proxy.ProxyPass, fmt.Sprintf("://%s", req.Name)) {
				return buserr.New("ErrProxyIsUsed")
			}
		}
	}

	includeDir := GetSitePath(website, SiteUpstreamDir)
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
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	nginxInstall, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	includeDir := GetSitePath(website, SiteUpstreamDir)
	filePath := path.Join(includeDir, fmt.Sprintf("%s.conf", req.Name))
	fileOp := files.NewFileOp()
	oldContent, err := fileOp.GetContent(filePath)
	if err != nil {
		return err
	}
	if err = fileOp.WriteFile(filePath, strings.NewReader(req.Content), constant.DirPerm); err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = fileOp.WriteFile(filePath, bytes.NewReader(oldContent), constant.DirPerm)
		}
	}()
	return opNginx(nginxInstall.ContainerName, constant.NginxReload)
}
