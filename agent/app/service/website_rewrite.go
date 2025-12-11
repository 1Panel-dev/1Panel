package service

import (
	"bytes"
	"context"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"os"
	"path"
	"strings"
)

func (w WebsiteService) UpdateRewriteConfig(req request.NginxRewriteUpdate) error {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	includePath := fmt.Sprintf("/www/sites/%s/rewrite/%s.conf", website.Alias, website.Alias)
	absolutePath := GetSitePath(website, SiteReWritePath)
	fileOp := files.NewFileOp()
	var oldRewriteContent []byte
	if !fileOp.Stat(path.Dir(absolutePath)) {
		if err := fileOp.CreateDir(path.Dir(absolutePath), constant.DirPerm); err != nil {
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
	if err := fileOp.WriteFile(absolutePath, strings.NewReader(req.Content), constant.DirPerm); err != nil {
		return err
	}

	if err := updateNginxConfig(constant.NginxScopeServer, []dto.NginxParam{{Name: "include", Params: []string{includePath}}}, &website); err != nil {
		_ = fileOp.WriteFile(absolutePath, bytes.NewReader(oldRewriteContent), constant.DirPerm)
		return err
	}
	website.Rewrite = req.Name
	return websiteRepo.Save(context.Background(), &website)
}

func (w WebsiteService) GetRewriteConfig(req request.NginxRewriteReq) (*response.NginxRewriteRes, error) {
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
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
		contentByte, _ = nginx_conf.Rewrites.ReadFile(rewriteFile)
		if contentByte == nil {
			customRewriteDir := GetOpenrestyDir(DefaultRewriteDir)
			customRewriteFile := path.Join(customRewriteDir, fmt.Sprintf("%s.conf", strings.ToLower(req.Name)))
			contentByte, err = files.NewFileOp().GetContent(customRewriteFile)
		}
	}
	return &response.NginxRewriteRes{
		Content: string(contentByte),
	}, err
}

func (w WebsiteService) OperateCustomRewrite(req request.CustomRewriteOperate) error {
	rewriteDir := GetOpenrestyDir(DefaultRewriteDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(rewriteDir) {
		if err := fileOp.CreateDir(rewriteDir, constant.DirPerm); err != nil {
			return err
		}
	}
	rewriteFile := path.Join(rewriteDir, fmt.Sprintf("%s.conf", req.Name))
	switch req.Operate {
	case "create":
		if fileOp.Stat(rewriteFile) {
			return buserr.New("ErrNameIsExist")
		}
		return fileOp.WriteFile(rewriteFile, strings.NewReader(req.Content), constant.DirPerm)
	case "delete":
		return fileOp.DeleteFile(rewriteFile)
	}
	return nil
}

func (w WebsiteService) ListCustomRewrite() ([]string, error) {
	rewriteDir := GetOpenrestyDir(DefaultRewriteDir)
	fileOp := files.NewFileOp()
	if !fileOp.Stat(rewriteDir) {
		return nil, nil
	}
	entries, err := os.ReadDir(rewriteDir)
	if err != nil {
		return nil, err
	}
	var res []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		res = append(res, strings.TrimSuffix(entry.Name(), ".conf"))
	}
	return res, nil
}
