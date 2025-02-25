package service

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/utils/common"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type AIToolService struct{}

type IAIToolService interface {
	Search(search dto.SearchWithPage) (int64, []dto.OllamaModelInfo, error)
	Create(name string) error
	Close(name string) error
	Recreate(name string) error
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
		options = append(options, commonRepo.WithLikeName(req.Info))
	}
	total, list, err := aiRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	var dtoLists []dto.OllamaModelInfo
	for _, itemModel := range list {
		var item dto.OllamaModelInfo
		if err := copier.Copy(&item, &itemModel); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		logPath := path.Join(global.CONF.System.DataDir, "log", "AITools", itemModel.Name)
		if _, err := os.Stat(logPath); err == nil {
			item.LogFileExist = true
		}
		dtoLists = append(dtoLists, item)
	}
	return int64(total), dtoLists, err
}

func (u *AIToolService) LoadDetail(name string) (string, error) {
	if cmd.CheckIllegal(name) {
		return "", buserr.New(constant.ErrCmdIllegal)
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

func (u *AIToolService) Create(name string) error {
	if cmd.CheckIllegal(name) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	modelInfo, _ := aiRepo.Get(commonRepo.WithByName(name))
	if modelInfo.ID != 0 {
		return constant.ErrRecordExist
	}
	containerName, err := LoadContainerName()
	if err != nil {
		return err
	}
	logItem := path.Join(global.CONF.System.DataDir, "log", "AITools", name)
	if _, err := os.Stat(path.Dir(logItem)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(logItem), os.ModePerm); err != nil {
			return err
		}
	}
	info := model.OllamaModel{
		Name:   name,
		From:   "local",
		Status: constant.StatusWaiting,
	}
	if err := aiRepo.Create(&info); err != nil {
		return err
	}
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	go pullOllamaModel(file, containerName, info)
	return nil
}

func (u *AIToolService) Close(name string) error {
	if cmd.CheckIllegal(name) {
		return buserr.New(constant.ErrCmdIllegal)
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

func (u *AIToolService) Recreate(name string) error {
	if cmd.CheckIllegal(name) {
		return buserr.New(constant.ErrCmdIllegal)
	}
	modelInfo, _ := aiRepo.Get(commonRepo.WithByName(name))
	if modelInfo.ID == 0 {
		return constant.ErrRecordNotFound
	}
	containerName, err := LoadContainerName()
	if err != nil {
		return err
	}
	if err := aiRepo.Update(modelInfo.ID, map[string]interface{}{"status": constant.StatusWaiting, "from": "local"}); err != nil {
		return err
	}
	logItem := path.Join(global.CONF.System.DataDir, "log", "AITools", name)
	if _, err := os.Stat(path.Dir(logItem)); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(logItem), os.ModePerm); err != nil {
			return err
		}
	}
	file, err := os.OpenFile(logItem, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	go pullOllamaModel(file, containerName, modelInfo)
	return nil
}

func (u *AIToolService) Delete(req dto.ForceDelete) error {
	ollamaList, _ := aiRepo.List(commonRepo.WithIdsIn(req.IDs))
	if len(ollamaList) == 0 {
		return constant.ErrRecordNotFound
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
		_ = aiRepo.Delete(commonRepo.WithByID(item.ID))
		logItem := path.Join(global.CONF.System.DataDir, "log", "AITools", item.Name)
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
	if req.SSLID > 0 {
		ssl, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		if ssl.Pem == "" {
			return buserr.New("ErrSSL")
		}
	}
	createWebsiteReq := request.WebsiteCreate{
		PrimaryDomain: req.Domain,
		Alias:         strings.ToLower(req.Domain),
		Type:          constant.Deployment,
		AppType:       constant.InstalledApp,
		AppInstallID:  req.AppInstallID,
	}
	websiteService := NewIWebsiteService()
	if err := websiteService.CreateWebsite(createWebsiteReq); err != nil {
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
	if req.SSLID > 0 {
		sslReq := request.WebsiteHTTPSOp{
			WebsiteID:    website.ID,
			Enable:       true,
			Type:         "existed",
			WebsiteSSLID: req.SSLID,
			HttpConfig:   "HTTPSOnly",
		}
		if _, err = websiteService.OpWebsiteHTTPS(context.Background(), sslReq); err != nil {
			return err
		}
	}
	if err = ConfigAIProxy(website); err != nil {
		return err
	}
	return nil
}

func (u *AIToolService) GetBindDomain(req dto.OllamaBindDomainReq) (*dto.OllamaBindDomainRes, error) {
	install, err := appInstallRepo.GetFirst(commonRepo.WithByID(req.AppInstallID))
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
		ssl, _ := websiteSSLRepo.GetFirst(commonRepo.WithByID(website.WebsiteSSLID))
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
	if req.SSLID > 0 {
		ssl, err := websiteSSLRepo.GetFirst(commonRepo.WithByID(req.SSLID))
		if err != nil {
			return err
		}
		if ssl.Pem == "" {
			return buserr.New("ErrSSL")
		}
	}
	websiteService := NewIWebsiteService()
	website, err := websiteRepo.GetFirst(commonRepo.WithByID(req.WebsiteID))
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
			Type:         "existed",
			WebsiteSSLID: req.SSLID,
			HttpConfig:   "HTTPSOnly",
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
	if ollamaBaseInfo.Status != constant.Running {
		return "", fmt.Errorf("container %s of ollama is not running, please check and retry!", ollamaBaseInfo.ContainerName)
	}
	return ollamaBaseInfo.ContainerName, nil
}

func pullOllamaModel(file *os.File, containerName string, info model.OllamaModel) {
	defer file.Close()
	cmd := exec.Command("docker", "exec", containerName, "ollama", "pull", info.Name)
	multiWriter := io.MultiWriter(os.Stdout, file)
	cmd.Stdout = multiWriter
	cmd.Stderr = multiWriter
	_ = cmd.Run()
	itemSize, err := loadModelSize(info.Name, containerName)
	if len(itemSize) != 0 {
		_ = aiRepo.Update(info.ID, map[string]interface{}{"status": constant.StatusSuccess, "size": itemSize})
	} else {
		_ = aiRepo.Update(info.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
	}
	_, _ = file.WriteString("ollama pull completed!")
}

func loadModelSize(name string, containerName string) (string, error) {
	stdout, err := cmd.Execf("docker exec %s ollama list | grep %s", containerName, name)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(stdout), "\n")
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		return parts[2] + " " + parts[3], nil
	}
	return "", fmt.Errorf("no such model %s in ollama list, std: %s", name, string(stdout))
}
