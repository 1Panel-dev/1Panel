package service

import (
	"bufio"
	"fmt"
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
	"os"
	"path"
	"strings"
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
	location.ChangePath("~*", fmt.Sprintf("^%s", req.Path))
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
		return nil
	}
	return nil
}
