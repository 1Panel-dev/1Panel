package service

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/jinzhu/copier"
)

type AIToolService struct{}

type IAIToolService interface {
	Search(search dto.SearchWithPage) (int64, []dto.OllamaModelInfo, error)
	Create(req dto.OllamaModelName) error
	Close(name string) error
	Recreate(req dto.OllamaModelName) error
	Delete(req dto.ForceDelete) error
	Sync() ([]dto.OllamaModelDropList, error)
	LoadDetail(name string) (string, error)
	BindDomain(req dto.OllamaBindDomain) error
	GetBindDomain(req dto.OllamaBindDomainReq) (*dto.OllamaBindDomainRes, error)
	UpdateBindDomain(req dto.OllamaBindDomain) error
}

func NewIAIToolService() IAIToolService {
	return &AIToolService{}
}

func (u *AIToolService) Search(req dto.SearchWithPage) (int64, []dto.OllamaModelInfo, error) {
	var options []repo.DBOption
	if len(req.Info) != 0 {
		options = append(options, repo.WithByLikeName(req.Info))
	}
	total, list, err := aiRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	var dtoLists []dto.OllamaModelInfo
	for _, itemModel := range list {
		var item dto.OllamaModelInfo
		if err := copier.Copy(&item, &itemModel); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		taskModel, _ := taskRepo.GetFirst(taskRepo.WithResourceID(item.ID), repo.WithByType(task.TaskScopeAI))
		if len(taskModel.ID) != 0 {
			item.LogFileExist = true
		}
		dtoLists = append(dtoLists, item)
	}
	return total, dtoLists, err
}

func (u *AIToolService) LoadDetail(name string) (string, error) {
	if cmd.CheckIllegal(name) {
		return "", buserr.New("ErrCmdIllegal")
	}
	containerName, err := LoadContainerName()
	if err != nil {
		return "", err
	}
	stdout, err := cmd.Execf("docker exec %s ollama show %s", containerName, name)
	if err != nil {
		return "", err
	}
	return stdout, err
}

func (u *AIToolService) Create(req dto.OllamaModelName) error {
	if cmd.CheckIllegal(req.Name) {
		return buserr.New("ErrCmdIllegal")
	}
	modelInfo, _ := aiRepo.Get(repo.WithByName(req.Name))
	if modelInfo.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	containerName, err := LoadContainerName()
	if err != nil {
		return err
	}
	info := model.OllamaModel{
		Name:   req.Name,
		From:   "local",
		Status: constant.StatusWaiting,
	}
	if err := aiRepo.Create(&info); err != nil {
		return err
	}
	taskItem, err := task.NewTaskWithOps(fmt.Sprintf("ollama-model-%s", req.Name), task.TaskPull, task.TaskScopeAI, req.TaskID, info.ID)
	if err != nil {
		global.LOG.Errorf("new task for exec shell failed, err: %v", err)
		return err
	}
	go func() {
		taskItem.AddSubTask(i18n.GetWithName("OllamaModelPull", req.Name), func(t *task.Task) error {
			return cmd.ExecShellWithTask(taskItem, time.Hour, "docker", "exec", containerName, "ollama", "pull", info.Name)
		}, nil)
		taskItem.AddSubTask(i18n.GetWithName("OllamaModelSize", req.Name), func(t *task.Task) error {
			itemSize, err := loadModelSize(info.Name, containerName)
			if len(itemSize) != 0 {
				_ = aiRepo.Update(info.ID, map[string]interface{}{"status": constant.StatusSuccess, "size": itemSize})
			} else {
				_ = aiRepo.Update(info.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			}
			return nil
		}, nil)
		if err := taskItem.Execute(); err != nil {
			_ = aiRepo.Update(info.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
		}
	}()
	return nil
}

func (u *AIToolService) Close(name string) error {
	if cmd.CheckIllegal(name) {
		return buserr.New("ErrCmdIllegal")
	}
	containerName, err := LoadContainerName()
	if err != nil {
		return err
	}
	stdout, err := cmd.Execf("docker exec %s ollama stop %s", containerName, name)
	if err != nil {
		return fmt.Errorf("handle ollama stop %s failed, stdout: %s, err: %v", name, stdout, err)
	}
	return nil
}

func (u *AIToolService) Recreate(req dto.OllamaModelName) error {
	if cmd.CheckIllegal(req.Name) {
		return buserr.New("ErrCmdIllegal")
	}
	modelInfo, _ := aiRepo.Get(repo.WithByName(req.Name))
	if modelInfo.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	containerName, err := LoadContainerName()
	if err != nil {
		return err
	}
	if err := aiRepo.Update(modelInfo.ID, map[string]interface{}{"status": constant.StatusWaiting, "from": "local"}); err != nil {
		return err
	}
	taskItem, err := task.NewTaskWithOps(fmt.Sprintf("ollama-model-%s", req.Name), task.TaskPull, task.TaskScopeAI, req.TaskID, modelInfo.ID)
	if err != nil {
		global.LOG.Errorf("new task for exec shell failed, err: %v", err)
		return err
	}
	go func() {
		taskItem.AddSubTask(i18n.GetWithName("OllamaModelPull", req.Name), func(t *task.Task) error {
			return cmd.ExecShellWithTask(taskItem, time.Hour, "docker", "exec", containerName, "ollama", "pull", req.Name)
		}, nil)
		taskItem.AddSubTask(i18n.GetWithName("OllamaModelSize", req.Name), func(t *task.Task) error {
			itemSize, err := loadModelSize(modelInfo.Name, containerName)
			if len(itemSize) != 0 {
				_ = aiRepo.Update(modelInfo.ID, map[string]interface{}{"status": constant.StatusSuccess, "size": itemSize})
			} else {
				_ = aiRepo.Update(modelInfo.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			}
			return nil
		}, nil)
		if err := taskItem.Execute(); err != nil {
			_ = aiRepo.Update(modelInfo.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
		}
	}()
	return nil
}

func (u *AIToolService) Delete(req dto.ForceDelete) error {
	ollamaList, _ := aiRepo.List(repo.WithByIDs(req.IDs))
	if len(ollamaList) == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	containerName, err := LoadContainerName()
	if err != nil && !req.ForceDelete {
		return err
	}
	for _, item := range ollamaList {
		if item.Status != constant.StatusDeleted {
			stdout, err := cmd.Execf("docker exec %s ollama rm %s", containerName, item.Name)
			if err != nil && !req.ForceDelete {
				return fmt.Errorf("handle ollama rm %s failed, stdout: %s, err: %v", item.Name, stdout, err)
			}
		}
		_ = aiRepo.Delete(repo.WithByID(item.ID))
		logItem := path.Join(global.Dir.DataDir, "log", "AITools", item.Name)
		_ = os.Remove(logItem)
	}
	return nil
}

func (u *AIToolService) Sync() ([]dto.OllamaModelDropList, error) {
	containerName, err := LoadContainerName()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.Execf("docker exec %s ollama list", containerName)
	if err != nil {
		return nil, err
	}
	var list []model.OllamaModel
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		if parts[0] == "NAME" {
			continue
		}
		list = append(list, model.OllamaModel{Name: parts[0], Size: parts[2] + " " + parts[3]})
	}
	listInDB, _ := aiRepo.List()
	var dropList []dto.OllamaModelDropList
	for _, itemModel := range listInDB {
		isExit := false
		for i := 0; i < len(list); i++ {
			if list[i].Name == itemModel.Name {
				_ = aiRepo.Update(itemModel.ID, map[string]interface{}{"status": constant.StatusSuccess, "message": "", "size": list[i].Size})
				list = append(list[:i], list[(i+1):]...)
				isExit = true
				break
			}
		}
		if !isExit && itemModel.Status != constant.StatusWaiting {
			_ = aiRepo.Update(itemModel.ID, map[string]interface{}{"status": constant.StatusDeleted, "message": "not exist", "size": ""})
			dropList = append(dropList, dto.OllamaModelDropList{ID: itemModel.ID, Name: itemModel.Name})
			continue
		}
	}
	for _, item := range list {
		item.Status = constant.StatusSuccess
		item.From = "remote"
		_ = aiRepo.Create(&item)
	}

	return dropList, nil
}

func (u *AIToolService) BindDomain(req dto.OllamaBindDomain) error {
	nginxInstall, _ := getAppInstallByKey(constant.AppOpenresty)
	if nginxInstall.ID == 0 {
		return buserr.New("ErrOpenrestyInstall")
	}
	var (
		ipList []string
		err    error
	)
	if len(req.IPList) > 0 {
		ipList, err = common.HandleIPList(req.IPList)
		if err != nil {
			return err
		}
	}
	createWebsiteReq := request.WebsiteCreate{
		Domains:      []request.WebsiteDomain{{Domain: req.Domain, Port: 80}},
		Alias:        strings.ToLower(req.Domain),
		Type:         constant.Deployment,
		AppType:      constant.InstalledApp,
		AppInstallID: req.AppInstallID,
	}
	if req.SSLID > 0 {
		createWebsiteReq.WebsiteSSLID = req.SSLID
		createWebsiteReq.EnableSSL = true
	}
	res, _ := NewIGroupService().GetDefault()
	createWebsiteReq.WebsiteGroupID = res.ID
	websiteService := NewIWebsiteService()
	if err = websiteService.CreateWebsite(createWebsiteReq); err != nil {
		return err
	}
	website, err := websiteRepo.GetFirst(websiteRepo.WithAlias(strings.ToLower(req.Domain)))
	if err != nil {
		return err
	}
	if len(ipList) > 0 {
		if err = ConfigAllowIPs(ipList, website); err != nil {
			return err
		}
	}
	if err = ConfigAIProxy(website); err != nil {
		return err
	}
	return nil
}

func (u *AIToolService) GetBindDomain(req dto.OllamaBindDomainReq) (*dto.OllamaBindDomainRes, error) {
	install, err := appInstallRepo.GetFirst(repo.WithByID(req.AppInstallID))
	if err != nil {
		return nil, err
	}
	res := &dto.OllamaBindDomainRes{}
	website, _ := websiteRepo.GetFirst(websiteRepo.WithAppInstallId(install.ID))
	if website.ID == 0 {
		return res, nil
	}
	res.WebsiteID = website.ID
	res.Domain = website.PrimaryDomain
	if website.WebsiteSSLID > 0 {
		res.SSLID = website.WebsiteSSLID
		ssl, _ := websiteSSLRepo.GetFirst(repo.WithByID(website.WebsiteSSLID))
		res.AcmeAccountID = ssl.AcmeAccountID
	}
	res.ConnUrl = fmt.Sprintf("%s://%s", strings.ToLower(website.Protocol), website.PrimaryDomain)
	res.AllowIPs = GetAllowIps(website)
	return res, nil
}

func (u *AIToolService) UpdateBindDomain(req dto.OllamaBindDomain) error {
	nginxInstall, _ := getAppInstallByKey(constant.AppOpenresty)
	if nginxInstall.ID == 0 {
		return buserr.New("ErrOpenrestyInstall")
	}
	var (
		ipList []string
		err    error
	)
	if len(req.IPList) > 0 {
		ipList, err = common.HandleIPList(req.IPList)
		if err != nil {
			return err
		}
	}
	websiteService := NewIWebsiteService()
	website, err := websiteRepo.GetFirst(repo.WithByID(req.WebsiteID))
	if err != nil {
		return err
	}
	if err = ConfigAllowIPs(ipList, website); err != nil {
		return err
	}
	if req.SSLID > 0 {
		sslReq := request.WebsiteHTTPSOp{
			WebsiteID:    website.ID,
			Enable:       true,
			Type:         constant.SSLExisted,
			WebsiteSSLID: req.SSLID,
			HttpConfig:   constant.HTTPToHTTPS,
		}
		if _, err = websiteService.OpWebsiteHTTPS(context.Background(), sslReq); err != nil {
			return err
		}
		return nil
	}
	if website.WebsiteSSLID > 0 && req.SSLID == 0 {
		sslReq := request.WebsiteHTTPSOp{
			WebsiteID: website.ID,
			Enable:    false,
		}
		if _, err = websiteService.OpWebsiteHTTPS(context.Background(), sslReq); err != nil {
			return err
		}
	}
	return nil
}

func LoadContainerName() (string, error) {
	ollamaBaseInfo, err := appInstallRepo.LoadBaseInfo("ollama", "")
	if err != nil {
		return "", fmt.Errorf("ollama service is not found, err: %v", err)
	}
	if ollamaBaseInfo.Status != constant.StatusRunning {
		return "", fmt.Errorf("container %s of ollama is not running, please check and retry!", ollamaBaseInfo.ContainerName)
	}
	return ollamaBaseInfo.ContainerName, nil
}

func loadModelSize(name string, containerName string) (string, error) {
	stdout, err := cmd.Execf("docker exec %s ollama list | grep %s", containerName, name)
	if err != nil {
		return "", err
	}
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		return parts[2] + " " + parts[3], nil
	}
	return "", fmt.Errorf("no such model %s in ollama list, std: %s", name, stdout)
}
