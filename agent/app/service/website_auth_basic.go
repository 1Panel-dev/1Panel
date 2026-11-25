package service

import (
	"bufio"
	"fmt"
	"os"
	"path"
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
	"golang.org/x/crypto/bcrypt"
)

func (w WebsiteService) GetAuthBasics(req request.NginxAuthReq) (res response.NginxAuthRes, err error) {
	var (
		website     model.Website
		authContent []byte
		nginxParams []response.NginxParam
	)
	website, err = websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	absoluteAuthPath := GetSitePath(website, SiteRootAuthBasicPath)
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
		website     model.Website
		params      []dto.NginxParam
		authContent []byte
		authArray   []string
	)
	website, err = websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	authPath := fmt.Sprintf("/www/sites/%s/auth_basic/auth.pass", website.Alias)
	absoluteAuthPath := GetSitePath(website, SiteRootAuthBasicPath)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(path.Dir(absoluteAuthPath)) {
		_ = fileOp.CreateDir(path.Dir(absoluteAuthPath), constant.DirPerm)
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
				err = buserr.New("ErrUsernameIsExist")
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
			err = buserr.New("ErrUsernameIsNotExist")
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
		website     model.Website
		authContent []byte
	)
	website, err = websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	absoluteAuthDir := GetSitePath(website, SitePathAuthBasicDir)
	passDir := path.Join(absoluteAuthDir, "pass")
	if !fileOp.Stat(absoluteAuthDir) || !fileOp.Stat(passDir) {
		return
	}

	entries, err := os.ReadDir(absoluteAuthDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".conf") {
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
			pathAuth.Path = strings.TrimPrefix(location.Match, "^")
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
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	authDir := GetSitePath(website, SitePathAuthBasicDir)
	if !fileOp.Stat(authDir) {
		_ = fileOp.CreateDir(authDir, constant.DirPerm)
	}
	passDir := path.Join(authDir, "pass")
	if !fileOp.Stat(passDir) {
		_ = fileOp.CreateDir(passDir, constant.DirPerm)
	}
	safeName := path.Base(req.Name)
	if safeName != req.Name || strings.Contains(safeName, "..") {
		return buserr.New("ErrInvalidParams")
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
			return buserr.New("ErrNameIsExist")
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
	applyPathAuthLocationMatch(location, req.Path)
	siteServer := loadSiteServer(fileOp, website)
	if err = syncPathAuthSensitiveDeny(location, siteServer); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	serverPHPLoc := findServerPHPLocation(siteServer)
	if err = syncPathAuthLocationPHP(location, serverPHPLoc != nil, website.Proxy, serverPHPLoc); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	var passwdHash []byte
	passwdHash, err = bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	line := fmt.Sprintf("%s:%s\n", req.Username, passwdHash)
	if req.Remark != "" {
		line = fmt.Sprintf("%s:%s:%s\n", req.Username, passwdHash, req.Remark)
	}
	_ = fileOp.SaveFile(passPath, line, constant.DirPerm)
	if err = nginx.WriteConfig(config, nginx.IndentedStyle); err != nil {
		return buserr.WithErr("ErrUpdateBuWebsite", err)
	}
	nginxInclude := fmt.Sprintf("/www/sites/%s/path_auth/*.conf", website.Alias)
	if err = updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{nginxInclude}}}, &website); err != nil {
		return err
	}
	return nil
}

const (
	phpLocationMatch              = "~ [^/]\\.php(/|$)"
	defaultSensitiveDenyMatch     = `^/(\.user.ini|\.htaccess|\.git|\.env|\.svn|\.project|LICENSE|README.md)`
	defaultSensitiveDenyDirective = `~ ^/(\.user.ini|\.htaccess|\.git|\.env|\.svn|\.project|LICENSE|README.md)`
)

func hasRegexMeta(pathVal string) bool {
	return strings.ContainsAny(pathVal, "*?()[]|$\\+{}")
}

func applyPathAuthLocationMatch(location *components.Location, rawPath string) {
	if location == nil {
		return
	}
	trimmed := strings.TrimPrefix(strings.TrimSpace(rawPath), "^")
	if trimmed == "" || hasRegexMeta(trimmed) {
		location.ChangePath("~*", fmt.Sprintf("^%s", trimmed))
		return
	}
	if !strings.HasPrefix(trimmed, "/") {
		trimmed = "/" + trimmed
	}
	location.ChangePath("^~", trimmed)
}

func upgradeLegacyPathAuthLocation(location *components.Location) {
	if location == nil || location.Modifier != "~*" {
		return
	}
	rawPath := strings.TrimPrefix(location.Match, "^")
	if rawPath == "" || !strings.HasPrefix(rawPath, "/") || hasRegexMeta(rawPath) {
		return
	}
	location.ChangePath("^~", rawPath)
}

func findServerPHPLocation(server *components.Server) components.IDirective {
	if server == nil {
		return nil
	}
	for _, dir := range server.Directives {
		if dir.GetName() == "location" && strings.Join(dir.GetParameters(), " ") == phpLocationMatch && dir.GetBlock() != nil {
			return dir
		}
	}
	return nil
}

func isDirectDenyLocation(dir components.IDirective) bool {
	if dir == nil || dir.GetName() != "location" || dir.GetBlock() == nil {
		return false
	}
	params := dir.GetParameters()
	if len(params) < 2 || (params[0] != "~" && params[0] != "~*") {
		return false
	}
	if strings.Join(params, " ") == phpLocationMatch {
		return false
	}
	blockDirs := dir.GetBlock().GetDirectives()
	if len(blockDirs) == 0 {
		return false
	}
	for _, child := range blockDirs {
		name := child.GetName()
		childParams := child.GetParameters()
		switch name {
		case "return":
			if len(childParams) > 0 && (childParams[0] == "403" || childParams[0] == "404") {
				return true
			}
		case "deny":
			if len(childParams) > 0 && childParams[0] == "all" {
				return true
			}
		}
	}
	return false
}

func findServerSensitiveDenyLocations(server *components.Server) []components.IDirective {
	if server == nil {
		return []components.IDirective{newDefaultSensitiveDenyLocation()}
	}
	var denyLocations []components.IDirective
	for _, dir := range server.Directives {
		if isDirectDenyLocation(dir) {
			denyLocations = append(denyLocations, dir)
		}
	}
	return denyLocations
}

func newDefaultSensitiveDenyLocation() components.IDirective {
	return &components.Directive{
		Name:       "location",
		Parameters: []string{"~", defaultSensitiveDenyMatch},
		Block: &components.Block{
			Directives: []components.IDirective{
				&components.Directive{
					Name:       "return",
					Parameters: []string{"404"},
				},
			},
		},
	}
}

func syncPathAuthSensitiveDeny(location *components.Location, server *components.Server) error {
	if location == nil || location.Modifier != "^~" {
		return nil
	}
	denyLocations := findServerSensitiveDenyLocations(server)
	if len(denyLocations) == 0 {
		return nil
	}
	denyKeys := make(map[string]struct{}, len(denyLocations)+1)
	denyKeys[defaultSensitiveDenyDirective] = struct{}{}
	seenClone := make(map[string]struct{}, len(denyLocations))
	clonedDenies := make([]components.IDirective, 0, len(denyLocations))
	for _, dir := range denyLocations {
		key := strings.Join(dir.GetParameters(), " ")
		if _, seen := seenClone[key]; seen {
			continue
		}
		seenClone[key] = struct{}{}
		denyKeys[key] = struct{}{}
		cloned, err := cloneNginxDirective(dir)
		if err != nil {
			return err
		}
		clonedDenies = append(clonedDenies, cloned)
	}

	var (
		beforePHP   []components.IDirective
		phpAndAfter []components.IDirective
		seenPHP     bool
	)
	for _, dir := range location.Directives {
		if dir.GetName() == "location" {
			key := strings.Join(dir.GetParameters(), " ")
			if _, isDeny := denyKeys[key]; isDeny {
				continue
			}
			if key == phpLocationMatch {
				seenPHP = true
			}
		}
		if seenPHP {
			phpAndAfter = append(phpAndAfter, dir)
		} else {
			beforePHP = append(beforePHP, dir)
		}
	}
	next := make([]components.IDirective, 0, len(beforePHP)+len(clonedDenies)+len(phpAndAfter))
	next = append(next, beforePHP...)
	next = append(next, clonedDenies...)
	next = append(next, phpAndAfter...)
	location.Directives = next
	return nil
}

func loadSiteServer(fileOp files.FileOp, website model.Website) *components.Server {
	configPath := GetSitePath(website, SiteConf)
	content, err := fileOp.GetContent(configPath)
	if err != nil {
		return nil
	}
	config, err := parser.NewStringParser(string(content)).Parse()
	if err != nil {
		return nil
	}
	servers := config.FindServers()
	if len(servers) == 0 {
		return nil
	}
	return servers[0]
}

func cloneNginxDirective(dir components.IDirective) (components.IDirective, error) {
	dumped := nginx.DumpDirective(dir, nginx.IndentedStyle)
	parsed, err := parser.NewStringParser(dumped).Parse()
	if err != nil {
		return nil, err
	}
	if len(parsed.Directives) == 0 {
		return nil, fmt.Errorf("failed to clone nginx directive")
	}
	return parsed.Directives[0], nil
}

func syncPathAuthLocationPHP(location *components.Location, enablePHP bool, proxy string, serverPHPLoc components.IDirective) error {
	if location == nil {
		return nil
	}
	if !enablePHP {
		filtered := make([]components.IDirective, 0, len(location.Directives))
		for _, dir := range location.Directives {
			if dir.GetName() == "location" && strings.Join(dir.GetParameters(), " ") == phpLocationMatch {
				continue
			}
			filtered = append(filtered, dir)
		}
		location.Directives = filtered
		return nil
	}
	for _, dir := range location.Directives {
		if dir.GetName() == "location" && strings.Join(dir.GetParameters(), " ") == phpLocationMatch && dir.GetBlock() != nil {
			if proxy != "" {
				dir.GetBlock().UpdateDirective("fastcgi_pass", []string{proxy})
			}
			return nil
		}
	}
	var phpDir components.IDirective
	if serverPHPLoc != nil {
		cloned, err := cloneNginxDirective(serverPHPLoc)
		if err != nil {
			return err
		}
		if proxy != "" && cloned.GetBlock() != nil {
			cloned.GetBlock().UpdateDirective("fastcgi_pass", []string{proxy})
		}
		phpDir = cloned
	} else if proxy != "" {
		phpDir = components.NewPHPProxyDirective([]string{proxy}, "")
	}
	if phpDir != nil {
		location.Directives = append(location.Directives, phpDir)
	}
	return nil
}

func collectSyncedPathAuthConfigs(
	fileOp files.FileOp,
	authDir string,
	enablePHP bool,
	proxy string,
	server *components.Server,
) ([]*components.Config, map[string][]byte, error) {
	if !fileOp.Stat(authDir) {
		return nil, nil, nil
	}
	entries, err := os.ReadDir(authDir)
	if err != nil {
		return nil, nil, err
	}
	serverPHPLoc := findServerPHPLocation(server)
	var syncedConfigs []*components.Config
	oldContents := make(map[string][]byte)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".conf") {
			continue
		}
		confPath := path.Join(authDir, entry.Name())
		content, err := fileOp.GetContent(confPath)
		if err != nil {
			return nil, nil, err
		}
		config, err := parser.NewStringParser(string(content)).Parse()
		if err != nil {
			return nil, nil, err
		}
		if len(config.Directives) == 0 {
			continue
		}
		location, ok := config.Directives[0].(*components.Location)
		if !ok {
			continue
		}
		upgradeLegacyPathAuthLocation(location)
		if err = syncPathAuthSensitiveDeny(location, server); err != nil {
			return nil, nil, err
		}
		if err = syncPathAuthLocationPHP(location, enablePHP, proxy, serverPHPLoc); err != nil {
			return nil, nil, err
		}
		config.FilePath = confPath
		syncedConfigs = append(syncedConfigs, config)
		oldContents[confPath] = content
	}
	return syncedConfigs, oldContents, nil
}
