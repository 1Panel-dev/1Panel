package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
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
	"github.com/1Panel-dev/1Panel/agent/utils/re"
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
	stdout, err := cmd.RunDockerExecWithStdout(20*time.Second, containerName, "ollama", "show", name)
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
			return runOllamaPullWithProcess(t, containerName, info.Name)
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
	if err := cmd.NewCommandMgr().Run("docker", "exec", containerName, "ollama", "stop", name); err != nil {
		return fmt.Errorf("handle ollama stop %s failed, %v", name, err)
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
			return runOllamaPullWithProcess(t, containerName, req.Name)
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
			if err := cmd.NewCommandMgr().Run("docker", "exec", containerName, "ollama", "rm", item.Name); err != nil && !req.ForceDelete {
				return fmt.Errorf("handle ollama rm %s failed, %v", item.Name, err)
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
	stdout, err := cmd.RunDockerExecWithStdout(20*time.Second, containerName, "ollama", "list")
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
	stdout, err := cmd.RunDockerExecWithStdout(20*time.Second, containerName, "ollama", "list")
	if err != nil {
		return "", err
	}
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if !strings.Contains(line, name) {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 5 {
			continue
		}
		return parts[2] + " " + parts[3], nil
	}
	return "", fmt.Errorf("no such model %s in ollama list, std: %s", name, stdout)
}

func runOllamaPullWithProcess(taskItem *task.Task, containerName, modelName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()

	cmdItem := exec.CommandContext(ctx, "docker", "exec", containerName, "ollama", "pull", modelName)
	cmdItem.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	writer := &ollamaPullLogWriter{taskItem: taskItem}
	cmdItem.Stdout = writer
	cmdItem.Stderr = writer

	if err := cmdItem.Start(); err != nil {
		return fmt.Errorf("failed to start ollama pull %s: %w", modelName, err)
	}
	waitErr := cmdItem.Wait()
	writer.Flush()
	if ctx.Err() == context.DeadlineExceeded {
		if cmdItem.Process != nil && cmdItem.Process.Pid > 0 {
			_ = syscall.Kill(-cmdItem.Process.Pid, syscall.SIGKILL)
		}
		return buserr.New("ErrCmdTimeout")
	}
	if waitErr == nil {
		return nil
	}
	if len(strings.TrimSpace(writer.errBuf.String())) > 0 {
		return fmt.Errorf("%s", strings.TrimSpace(writer.errBuf.String()))
	}
	return waitErr
}

type ollamaPullLogWriter struct {
	taskItem *task.Task
	errBuf   bytes.Buffer
}

func (w *ollamaPullLogWriter) Write(p []byte) (n int, err error) {
	w.errBuf.Write(p)
	for _, segment := range splitOllamaPullSegments(string(p)) {
		w.logLine(segment)
	}
	return len(p), nil
}

func (w *ollamaPullLogWriter) Flush() {}

func (w *ollamaPullLogWriter) logLine(line string) {
	if w == nil || w.taskItem == nil {
		return
	}
	line = sanitizeOllamaPullLine(line)
	if line == "" {
		return
	}
	if prefix := resolveOllamaPullLogPrefix(line); prefix != "" {
		_ = replaceOllamaPullLogLine(prefix, line, w.taskItem)
		return
	}
	w.taskItem.Log(line)
}

func sanitizeOllamaPullLine(line string) string {
	line = re.StripAnsiControlSeq(line)
	line = strings.TrimSpace(line)
	if strings.HasPrefix(line, "pulling manifes") {
		return ""
	}
	return line
}

func splitOllamaPullSegments(chunk string) []string {
	if chunk == "" {
		return nil
	}
	chunk = strings.ReplaceAll(chunk, "\r", "\n")
	parts := strings.Split(chunk, "\x1b[1G")
	segments := make([]string, 0, len(parts))
	for _, part := range parts {
		for _, line := range strings.Split(part, "\n") {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			segments = append(segments, line)
		}
	}
	return segments
}

func resolveOllamaPullLogPrefix(line string) string {
	if strings.HasPrefix(line, "pulling ") {
		if idx := strings.Index(line, ":"); idx > 0 {
			return strings.TrimSpace(line[:idx])
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			return strings.TrimSpace(fields[0] + " " + fields[1])
		}
	}
	return ""
}

func replaceOllamaPullLogLine(prefix, newLine string, taskItem *task.Task) error {
	if taskItem == nil || taskItem.Task == nil {
		return nil
	}
	data, err := os.ReadFile(taskItem.Task.LogFile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	for idx, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.Contains(trimmed, prefix) {
			lines[idx] = time.Now().Format("2006/01/02 15:04:05") + " " + newLine
			return os.WriteFile(taskItem.Task.LogFile, []byte(strings.Join(lines, "\n")), os.ModePerm)
		}
	}
	taskItem.Log(newLine)
	return nil
}
