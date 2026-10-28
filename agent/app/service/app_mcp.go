package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
)

func (a AppService) InstallForMCP(req request.AppInstallCreate, executeScript bool) (*model.AppInstall, error) {
	return a.installWithHooks(req, executeScript, &appInstallHooks{PrepareSiteDir: prepareMCPInstallSiteDirectory})
}

func prepareMCPInstallSiteDirectory(app model.App, detail model.AppDetail, req request.AppInstallCreate, _ *task.Task) error {
	if !managesOpenrestySiteDirectory(app, detail) {
		return nil
	}
	directory, _ := req.Params["WEBSITE_DIR"].(string)
	return ensureCurrentWebsiteDirectory(directory)
}
