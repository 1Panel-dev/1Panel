package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

type AgentService struct{}

type IAgentService interface {
	Create(req dto.AgentCreateReq) (*dto.AgentItem, error)
	Page(req dto.SearchWithPage) (int64, []dto.AgentItem, error)
	Delete(req dto.AgentDeleteReq) error
	UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error
	GetProviders() ([]dto.ProviderInfo, error)
	CreateAccount(req dto.AgentAccountCreateReq) error
	UpdateAccount(req dto.AgentAccountUpdateReq) error
	PageAccounts(req dto.AgentAccountSearch) (int64, []dto.AgentAccountInfo, error)
	VerifyAccount(req dto.AgentAccountVerifyReq) error
	DeleteAccount(req dto.AgentAccountDeleteReq) error
	GetFeishuConfig(req dto.AgentFeishuConfigReq) (*dto.AgentFeishuConfig, error)
	UpdateFeishuConfig(req dto.AgentFeishuConfigUpdateReq) error
	ApproveFeishuPairing(req dto.AgentFeishuPairingApproveReq) error
}

func NewIAgentService() IAgentService {
	return &AgentService{}
}

func (a AgentService) Create(req dto.AgentCreateReq) (*dto.AgentItem, error) {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if !isSupportedAgentProvider(provider) {
		return nil, buserr.New("ErrAgentProviderNotSupported")
	}
	if req.AccountID == 0 {
		return nil, buserr.New("ErrAgentAccountRequired")
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return nil, err
	}
	if !account.Verified {
		return nil, buserr.New("ErrAgentAccountNotVerified")
	}
	if account.Provider != "" && provider != "" && account.Provider != provider {
		return nil, buserr.New("ErrAgentProviderMismatch")
	}
	provider = strings.ToLower(strings.TrimSpace(account.Provider))
	baseURL := strings.TrimSpace(account.BaseURL)
	if baseURL == "" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return nil, buserr.New("ErrAgentBaseURLRequired")
	}
	if provider != "ollama" && strings.TrimSpace(account.APIKey) == "" {
		return nil, buserr.New("ErrAgentApiKeyRequired")
	}
	if err := checkPortExist(req.WebUIPort); err != nil {
		return nil, err
	}
	if err := checkPortExist(req.BridgePort); err != nil {
		return nil, err
	}
	if exist, _ := agentRepo.GetFirst(repo.WithByLowerName(req.Name)); exist != nil && exist.ID > 0 {
		return nil, buserr.New("ErrNameIsExist")
	}
	if installs, _ := appInstallRepo.ListBy(context.Background(), repo.WithByLowerName(req.Name)); len(installs) > 0 {
		return nil, buserr.New("ErrNameIsExist")
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(constant.AppOpenclaw))
	if err != nil || app.ID == 0 {
		return nil, buserr.New("ErrRecordNotFound")
	}
	detail, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(app.ID), appDetailRepo.WithVersion(req.AppVersion))
	if err != nil || detail.ID == 0 {
		return nil, buserr.New("ErrRecordNotFound")
	}

	token := strings.TrimSpace(req.Token)
	if token == "" {
		token = randomToken()
	}
	params := map[string]interface{}{
		"PROVIDER":               provider,
		"MODEL":                  req.Model,
		"BASE_URL":               baseURL,
		"API_KEY":                account.APIKey,
		"OPENCLAW_GATEWAY_TOKEN": token,
		"PANEL_APP_PORT_HTTP":    req.WebUIPort,
		"PANEL_APP_PORT_BRIDGE":  req.BridgePort,
		constant.CPUS:            "0",
		constant.MemoryLimit:     "0",
		constant.HostIP:          "",
	}

	if req.EditCompose && strings.TrimSpace(req.DockerCompose) == "" {
		return nil, buserr.New("ErrAgentComposeRequired")
	}
	installReq := request.AppInstallCreate{
		AppDetailId: detail.ID,
		Name:        req.Name,
		Params:      params,
		TaskID:      req.TaskID,
		AppContainerConfig: request.AppContainerConfig{
			Advanced:      req.Advanced,
			ContainerName: req.ContainerName,
			AllowPort:     req.AllowPort,
			SpecifyIP:     req.SpecifyIP,
			RestartPolicy: req.RestartPolicy,
			CpuQuota:      req.CpuQuota,
			MemoryLimit:   req.MemoryLimit,
			MemoryUnit:    req.MemoryUnit,
			PullImage:     req.PullImage,
			EditCompose:   req.EditCompose,
			DockerCompose: req.DockerCompose,
		},
	}
	appInstall, err := NewIAppService().Install(installReq, false)
	if err != nil {
		return nil, err
	}
	configPath := path.Join(appInstall.GetPath(), "data", "conf", "openclaw.json")
	agent := &model.Agent{
		Name:         req.Name,
		Provider:     provider,
		Model:        req.Model,
		BaseURL:      baseURL,
		APIKey:       account.APIKey,
		Token:        token,
		Status:       appInstall.Status,
		Message:      appInstall.Message,
		AppInstallID: appInstall.ID,
		AccountID:    account.ID,
		ConfigPath:   configPath,
	}
	if err := agentRepo.Create(agent); err != nil {
		return nil, err
	}
	go a.writeConfigWithRetry(appInstall, provider, req.Model, baseURL, req.APIKey, token, agent.ID)

	item := buildAgentItem(agent, appInstall, nil)
	return &item, nil
}

func (a AgentService) Page(req dto.SearchWithPage) (int64, []dto.AgentItem, error) {
	var opts []repo.DBOption
	if strings.TrimSpace(req.Info) != "" {
		opts = append(opts, repo.WithByLikeName(req.Info))
	}
	count, list, err := agentRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	items := make([]dto.AgentItem, 0, len(list))
	for _, item := range list {
		appInstall, _ := appInstallRepo.GetFirst(repo.WithByID(item.AppInstallID))
		envMap := readInstallEnv(appInstall.Env)
		agentItem := buildAgentItem(&item, &appInstall, envMap)
		agentItem.Upgradable = checkAgentUpgradable(appInstall)
		items = append(items, agentItem)
	}
	return count, items, nil
}

func (a AgentService) Delete(req dto.AgentDeleteReq) error {
	if req.ID == 0 {
		return buserr.New("ErrAgentIDRequired")
	}
	agent, err := agentRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if agent.AppInstallID == 0 {
		return agentRepo.DeleteByID(agent.ID)
	}
	operate := request.AppInstalledOperate{
		InstallId:   agent.AppInstallID,
		Operate:     constant.Delete,
		TaskID:      req.TaskID,
		ForceDelete: req.ForceDelete,
	}
	if err := NewIAppInstalledService().Operate(operate); err != nil {
		return err
	}
	go a.waitAndDeleteAgent(agent.ID, agent.AppInstallID)
	return nil
}

func (a AgentService) UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return err
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return err
	}
	if !account.Verified {
		return buserr.New("ErrAgentAccountNotVerified")
	}
	provider := strings.ToLower(strings.TrimSpace(account.Provider))
	if !isSupportedAgentProvider(provider) {
		return buserr.New("ErrAgentProviderNotSupported")
	}
	modelName := strings.TrimSpace(req.Model)
	if modelName == "" {
		return buserr.New("ErrAgentProviderMismatch")
	}
	if !strings.HasPrefix(modelName, provider+"/") {
		return buserr.New("ErrAgentProviderMismatch")
	}
	baseURL := strings.TrimSpace(account.BaseURL)
	if provider != "ollama" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if provider != "ollama" && strings.TrimSpace(account.APIKey) == "" {
		return buserr.New("ErrAgentApiKeyRequired")
	}
	confDir := ""
	if agent.ConfigPath != "" {
		confDir = path.Dir(agent.ConfigPath)
	} else if agent.AppInstallID > 0 {
		install, errGet := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
		if errGet == nil {
			confDir = path.Join(install.GetPath(), "data", "conf")
		}
	}
	if confDir == "" {
		return buserr.New("ErrRecordNotFound")
	}
	if err := writeOpenclawConfig(confDir, provider, modelName, baseURL, account.APIKey, agent.Token); err != nil {
		return err
	}
	agent.Provider = provider
	agent.Model = modelName
	agent.BaseURL = baseURL
	agent.APIKey = account.APIKey
	agent.AccountID = account.ID
	return agentRepo.Save(agent)
}

func (a AgentService) GetProviders() ([]dto.ProviderInfo, error) {
	definitions := providerDefinitions()
	providers := make([]dto.ProviderInfo, 0, len(definitions))
	for key, def := range definitions {
		providers = append(providers, dto.ProviderInfo{
			Sort:     def.Sort,
			Provider: key,
			BaseURL:  def.BaseURL,
			Models:   def.Models,
		})
	}
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Sort < providers[j].Sort
	})
	return providers, nil
}

func (a AgentService) CreateAccount(req dto.AgentAccountCreateReq) error {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if !isSupportedAgentProvider(provider) {
		return buserr.New("ErrAgentProviderNotSupported")
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		return buserr.New("ErrAgentApiKeyRequired")
	}
	baseURL := strings.TrimSpace(req.BaseURL)
	if provider != "ollama" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if exist, _ := agentAccountRepo.GetFirst(repo.WithByProvider(provider), repo.WithByName(req.Name)); exist != nil && exist.ID > 0 {
		return buserr.New("ErrRecordExist")
	}
	if err := a.VerifyAccount(dto.AgentAccountVerifyReq{Provider: provider, BaseURL: baseURL, APIKey: apiKey}); err != nil {
		return err
	}
	account := &model.AgentAccount{
		Provider: provider,
		Name:     req.Name,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Verified: true,
		Remark:   req.Remark,
	}
	return agentAccountRepo.Create(account)
}

func (a AgentService) UpdateAccount(req dto.AgentAccountUpdateReq) error {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	provider := strings.ToLower(strings.TrimSpace(account.Provider))
	baseURL := strings.TrimSpace(req.BaseURL)
	if provider != "ollama" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if err := a.VerifyAccount(dto.AgentAccountVerifyReq{Provider: provider, BaseURL: baseURL, APIKey: req.APIKey}); err != nil {
		return err
	}
	account.Name = req.Name
	account.APIKey = req.APIKey
	account.BaseURL = baseURL
	account.Remark = req.Remark
	account.Verified = true
	if err := agentAccountRepo.Save(account); err != nil {
		return err
	}
	if req.SyncAgents {
		if err := a.syncAgentsByAccount(account); err != nil {
			return err
		}
	}
	return nil
}

func (a AgentService) PageAccounts(req dto.AgentAccountSearch) (int64, []dto.AgentAccountInfo, error) {
	var opts []repo.DBOption
	if strings.TrimSpace(req.Provider) != "" {
		opts = append(opts, repo.WithByProvider(req.Provider))
	}
	if strings.TrimSpace(req.Name) != "" {
		opts = append(opts, repo.WithByLikeName(req.Name))
	}
	count, list, err := agentAccountRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	items := make([]dto.AgentAccountInfo, 0, len(list))
	for _, item := range list {
		items = append(items, dto.AgentAccountInfo{
			ID:        item.ID,
			Provider:  item.Provider,
			Name:      item.Name,
			APIKey:    item.APIKey,
			BaseURL:   item.BaseURL,
			Verified:  item.Verified,
			Remark:    item.Remark,
			CreatedAt: item.CreatedAt,
		})
	}
	return count, items, nil
}

func (a AgentService) VerifyAccount(req dto.AgentAccountVerifyReq) error {
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if !isSupportedAgentProvider(provider) {
		return buserr.New("ErrAgentProviderNotSupported")
	}
	apiKey := strings.TrimSpace(req.APIKey)
	if apiKey == "" {
		return buserr.New("ErrAgentApiKeyRequired")
	}
	baseURL := strings.TrimSpace(req.BaseURL)
	if baseURL == "" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if provider == "ollama" {
		return nil
	}
	return verifyProvider(provider, baseURL, apiKey)
}

func (a AgentService) DeleteAccount(req dto.AgentAccountDeleteReq) error {
	if req.ID == 0 {
		return buserr.New("ErrAgentAccountIDRequired")
	}
	if exists, _ := agentRepo.GetFirst(repo.WithByAccountID(req.ID)); exists != nil && exists.ID > 0 {
		return buserr.New("ErrAgentAccountBound")
	}
	return agentAccountRepo.DeleteByID(req.ID)
}

func (a AgentService) GetFeishuConfig(req dto.AgentFeishuConfigReq) (*dto.AgentFeishuConfig, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	_ = install
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractFeishuConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateFeishuConfig(req dto.AgentFeishuConfigUpdateReq) error {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	if req.DmPolicy == "" {
		req.DmPolicy = "pairing"
	}
	setFeishuConfig(conf, dto.AgentFeishuConfig{
		Enabled:   req.Enabled,
		DmPolicy:  req.DmPolicy,
		BotName:   req.BotName,
		AppID:     req.AppID,
		AppSecret: req.AppSecret,
	})
	if err := writeOpenclawConfigRaw(agent.ConfigPath, conf); err != nil {
		return err
	}
	return nil
}

func (a AgentService) ApproveFeishuPairing(req dto.AgentFeishuPairingApproveReq) error {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := cmd.RunDefaultBashCf(
		"docker exec %s openclaw pairing approve feishu %q",
		install.ContainerName,
		strings.TrimSpace(req.PairingCode),
	); err != nil {
		return err
	}
	return nil
}

func (a AgentService) loadAgentAndInstall(agentID uint) (*model.Agent, *model.AppInstall, error) {
	agent, err := agentRepo.GetFirst(repo.WithByID(agentID))
	if err != nil {
		return nil, nil, err
	}
	if agent.AppInstallID == 0 {
		return nil, nil, buserr.New("ErrRecordNotFound")
	}
	install, err := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
	if err != nil {
		return nil, nil, err
	}
	return agent, &install, nil
}

func readOpenclawConfig(configPath string) (map[string]interface{}, error) {
	if strings.TrimSpace(configPath) == "" {
		return nil, buserr.New("ErrRecordNotFound")
	}
	fileOp := files.NewFileOp()
	content, err := fileOp.GetContent(configPath)
	if err != nil {
		return nil, err
	}
	conf := map[string]interface{}{}
	if err := json.Unmarshal(content, &conf); err != nil {
		return nil, err
	}
	return conf, nil
}

func writeOpenclawConfigRaw(configPath string, conf map[string]interface{}) error {
	payload, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	return fileOp.SaveFile(configPath, string(payload), 0600)
}

func extractFeishuConfig(conf map[string]interface{}) dto.AgentFeishuConfig {
	result := dto.AgentFeishuConfig{Enabled: true, DmPolicy: "pairing"}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	feishu, ok := channels["feishu"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := feishu["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy, ok := feishu["dmPolicy"].(string); ok && strings.TrimSpace(dmPolicy) != "" {
		result.DmPolicy = dmPolicy
	}
	accounts, ok := feishu["accounts"].(map[string]interface{})
	if !ok {
		return result
	}
	main, ok := accounts["main"].(map[string]interface{})
	if !ok {
		return result
	}
	if appID, ok := main["appId"].(string); ok {
		result.AppID = appID
	}
	if appSecret, ok := main["appSecret"].(string); ok {
		result.AppSecret = appSecret
	}
	if botName, ok := main["botName"].(string); ok {
		result.BotName = botName
	}
	return result
}

func setFeishuConfig(conf map[string]interface{}, config dto.AgentFeishuConfig) {
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		channels = map[string]interface{}{}
		conf["channels"] = channels
	}
	feishu := map[string]interface{}{
		"enabled":  config.Enabled,
		"dmPolicy": config.DmPolicy,
		"accounts": map[string]interface{}{
			"main": map[string]interface{}{
				"appId":     config.AppID,
				"appSecret": config.AppSecret,
				"botName":   config.BotName,
			},
		},
	}
	channels["feishu"] = feishu
}

func (a AgentService) syncAgentsByAccount(account *model.AgentAccount) error {
	agents, err := agentRepo.List(repo.WithByAccountID(account.ID))
	if err != nil {
		return err
	}
	baseURL := account.BaseURL
	if account.Provider != "ollama" {
		if defaultURL, ok := providerDefaultBaseURL(account.Provider); ok {
			baseURL = defaultURL
		}
	}
	for _, agent := range agents {
		confDir := ""
		if agent.ConfigPath != "" {
			confDir = path.Dir(agent.ConfigPath)
		} else if agent.AppInstallID > 0 {
			install, err := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
			if err == nil {
				confDir = path.Join(install.GetPath(), "data", "conf")
			}
		}
		if confDir == "" {
			continue
		}
		if err := writeOpenclawConfig(confDir, account.Provider, agent.Model, baseURL, account.APIKey, agent.Token); err != nil {
			return err
		}
		agent.BaseURL = baseURL
		agent.APIKey = account.APIKey
		agent.Provider = account.Provider
		_ = agentRepo.Save(&agent)
	}
	return nil
}

func verifyProvider(provider, baseURL, apiKey string) error {
	if provider == "minimax" {
		return verifyMinimax(baseURL, apiKey)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	reqURL, headers := buildVerifyRequest(provider, baseURL, apiKey)
	request, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	resp, err := client.Do(request)
	if err != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return buserr.WithErr("ErrAgentAccountUnavailable", fmt.Errorf("verify failed: %s", resp.Status))
	}
	return nil
}

func verifyMinimax(baseURL, apiKey string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	base := strings.TrimRight(baseURL, "/")
	if !strings.Contains(base, "/v1") {
		base = base + "/v1"
	}
	reqURL := base + "/chat/completions"
	body := map[string]interface{}{
		"model": "MiniMax-M2.1",
		"messages": []map[string]string{
			{"role": "user", "content": "test"},
		},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	request, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	request.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(request)
	if err != nil {
		return buserr.WithErr("ErrAgentAccountUnavailable", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return buserr.WithErr("ErrAgentAccountUnavailable", fmt.Errorf("verify failed: %s", resp.Status))
	}
	return nil
}

func buildAgentItem(agent *model.Agent, appInstall *model.AppInstall, envMap map[string]interface{}) dto.AgentItem {
	item := dto.AgentItem{
		ID:           agent.ID,
		Name:         agent.Name,
		Provider:     agent.Provider,
		Model:        agent.Model,
		BaseURL:      agent.BaseURL,
		APIKey:       maskKey(agent.APIKey),
		Token:        agent.Token,
		Status:       agent.Status,
		Message:      agent.Message,
		AppInstallID: agent.AppInstallID,
		AccountID:    agent.AccountID,
		ConfigPath:   agent.ConfigPath,
		CreatedAt:    agent.CreatedAt,
	}
	if appInstall != nil && appInstall.ID > 0 {
		item.Container = appInstall.ContainerName
		item.AppVersion = appInstall.Version
		item.WebUIPort = appInstall.HttpPort
		item.Path = appInstall.GetPath()
		item.Status = appInstall.Status
		item.Message = appInstall.Message
		if envMap != nil {
			if bridge, ok := envMap["PANEL_APP_PORT_BRIDGE"]; ok {
				item.BridgePort = toInt(bridge)
			}
		}
	}
	return item
}

func checkAgentUpgradable(install model.AppInstall) bool {
	if install.ID == 0 || install.Version == "" || install.Version == "latest" {
		return false
	}
	if install.App.ID == 0 {
		return false
	}
	details, err := appDetailRepo.GetBy(appDetailRepo.WithAppId(install.App.ID))
	if err != nil || len(details) == 0 {
		return false
	}
	versions := make([]string, 0, len(details))
	for _, item := range details {
		ignores, _ := appIgnoreUpgradeRepo.List(runtimeRepo.WithDetailId(item.ID), appIgnoreUpgradeRepo.WithScope("version"))
		if len(ignores) > 0 {
			continue
		}
		if common.IsCrossVersion(install.Version, item.Version) && !install.App.CrossVersionUpdate {
			continue
		}
		versions = append(versions, item.Version)
	}
	if len(versions) == 0 {
		return false
	}
	versions = common.GetSortedVersions(versions)
	lastVersion := versions[0]
	if common.IsCrossVersion(install.Version, lastVersion) {
		return install.App.CrossVersionUpdate
	}
	return common.CompareVersion(lastVersion, install.Version)
}

func (a AgentService) waitAndDeleteAgent(agentID uint, appInstallID uint) {
	if appInstallID == 0 {
		_ = agentRepo.DeleteByID(agentID)
		return
	}
	for i := 0; i < 180; i++ {
		_, err := appInstallRepo.GetFirst(repo.WithByID(appInstallID))
		if err != nil {
			_ = agentRepo.DeleteByID(agentID)
			return
		}
		time.Sleep(2 * time.Second)
	}
}

func (a AgentService) writeConfigWithRetry(appInstall *model.AppInstall, provider, modelName, baseURL, apiKey, token string, agentID uint) {
	if appInstall == nil {
		return
	}
	fileOp := files.NewFileOp()
	composePath := appInstall.GetComposePath()
	for i := 0; i < 60; i++ {
		if fileOp.Stat(composePath) {
			break
		}
		time.Sleep(time.Second)
	}
	confDir := path.Join(appInstall.GetPath(), "data", "conf")
	if err := writeOpenclawConfig(confDir, provider, modelName, baseURL, apiKey, token); err != nil {
		global.LOG.Errorf("write openclaw config failed: %v", err)
		agent, errGet := agentRepo.GetFirst(repo.WithByID(agentID))
		if errGet == nil && agent != nil {
			agent.Message = err.Error()
			agent.Status = constant.StatusError
			_ = agentRepo.Save(agent)
		}
		return
	}
	dataDir := path.Join(appInstall.GetPath(), "data")
	for i := 0; i < 60; i++ {
		if fileOp.Stat(dataDir) {
			if err := fileOp.ChownR(dataDir, "1000", "1000", true); err != nil {
				global.LOG.Errorf("chown data dir failed: %v", err)
				agent, errGet := agentRepo.GetFirst(repo.WithByID(agentID))
				if errGet == nil && agent != nil {
					agent.Message = err.Error()
					agent.Status = constant.StatusError
					_ = agentRepo.Save(agent)
				}
			}
			break
		}
		time.Sleep(time.Second)
	}
}

type openclawConfig struct {
	Gateway gatewayConfig `json:"gateway"`
	Agents  agentsConfig  `json:"agents"`
	Models  *modelsConfig `json:"models,omitempty"`
}

type gatewayConfig struct {
	Mode      string           `json:"mode"`
	Bind      string           `json:"bind"`
	Port      int              `json:"port"`
	Auth      gatewayAuth      `json:"auth"`
	ControlUi gatewayControlUi `json:"controlUi"`
}

type gatewayControlUi struct {
	AllowInsecureAuth bool `json:"allowInsecureAuth"`
}

type gatewayAuth struct {
	Mode  string `json:"mode"`
	Token string `json:"token"`
}

type agentsConfig struct {
	Defaults agentDefaults `json:"defaults"`
}

type agentDefaults struct {
	Model modelRef `json:"model"`
}

type modelRef struct {
	Primary string `json:"primary"`
}

type modelsConfig struct {
	Mode      string                   `json:"mode,omitempty"`
	Providers map[string]modelProvider `json:"providers,omitempty"`
}

type modelProvider struct {
	ApiKey  string       `json:"apiKey,omitempty"`
	BaseUrl string       `json:"baseUrl,omitempty"`
	Api     string       `json:"api,omitempty"`
	Models  []modelEntry `json:"models,omitempty"`
}

type modelEntry struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Reasoning     bool      `json:"reasoning"`
	Input         []string  `json:"input"`
	ContextWindow int       `json:"contextWindow"`
	MaxTokens     int       `json:"maxTokens"`
	Cost          modelCost `json:"cost"`
}

type modelCost struct {
	Input      float64 `json:"input"`
	Output     float64 `json:"output"`
	CacheRead  float64 `json:"cacheRead"`
	CacheWrite float64 `json:"cacheWrite"`
}

func writeOpenclawConfig(confDir, provider, modelName, baseURL, apiKey, token string) error {
	if strings.TrimSpace(confDir) == "" {
		return fmt.Errorf("config dir is required")
	}
	if strings.TrimSpace(modelName) == "" {
		return fmt.Errorf("model is required")
	}
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("gateway token is required")
	}
	fileOp := files.NewFileOp()
	if !fileOp.Stat(confDir) {
		if err := fileOp.CreateDir(confDir, constant.DirPerm); err != nil {
			return err
		}
	}

	cfg := openclawConfig{
		Gateway: gatewayConfig{
			Mode: "local",
			Bind: "lan",
			Port: 18789,
			Auth: gatewayAuth{
				Mode:  "token",
				Token: token,
			},
			ControlUi: gatewayControlUi{
				AllowInsecureAuth: true,
			},
		},
		Agents: agentsConfig{
			Defaults: agentDefaults{
				Model: modelRef{Primary: modelName},
			},
		},
	}

	provider = strings.ToLower(strings.TrimSpace(provider))
	modelID := modelName
	if parts := strings.SplitN(modelName, "/", 2); len(parts) == 2 {
		modelID = parts[1]
	}
	configProvider := provider
	primaryModel := modelName
	if provider == "kimi" {
		configProvider = "moonshot"
		primaryModel = "moonshot/" + modelID
	}
	if provider == "deepseek" {
		cfg.Agents.Defaults.Model.Primary = modelName
		base := baseURL
		if base == "" {
			base = "https://api.deepseek.com/v1"
		}
		plainKey := strings.TrimSpace(apiKey)
		cfg.Models = &modelsConfig{
			Mode: "merge",
			Providers: map[string]modelProvider{
				"deepseek": {
					ApiKey:  plainKey,
					BaseUrl: base,
					Api:     "openai-completions",
					Models: []modelEntry{
						{
							ID:            "deepseek-chat",
							Name:          "DeepSeek Chat",
							Reasoning:     false,
							Input:         []string{"text"},
							ContextWindow: 128000,
							MaxTokens:     8192,
							Cost:          modelCost{},
						},
					},
				},
			},
		}
	} else if provider == "moonshot" || provider == "kimi" {
		cfg.Agents.Defaults.Model.Primary = primaryModel
		base := baseURL
		if base == "" {
			if defaultURL, ok := providerDefaultBaseURL(provider); ok {
				base = defaultURL
			}
		}
		plainKey := strings.TrimSpace(apiKey)
		cfg.Models = &modelsConfig{
			Mode: "merge",
			Providers: map[string]modelProvider{
				configProvider: {
					ApiKey:  plainKey,
					BaseUrl: base,
					Api:     "openai-completions",
					Models: []modelEntry{
						{
							ID:            modelID,
							Name:          modelID,
							Reasoning:     strings.Contains(modelID, "thinking"),
							Input:         []string{"text"},
							ContextWindow: 256000,
							MaxTokens:     8192,
							Cost:          modelCost{},
						},
					},
				},
			},
		}
	} else if provider == "ollama" {
		cfg.Agents.Defaults.Model.Primary = modelName
		cfg.Models = &modelsConfig{
			Mode: "merge",
			Providers: map[string]modelProvider{
				"ollama": {
					ApiKey:  "ollama",
					BaseUrl: baseURL,
					Api:     "openai-responses",
					Models: []modelEntry{
						{
							ID:            modelID,
							Name:          modelID,
							Reasoning:     true,
							Input:         []string{"text"},
							ContextWindow: 160000,
							MaxTokens:     8192,
							Cost:          modelCost{},
						},
					},
				},
			},
		}
	} else if provider == "kimi-coding" {
		cfg.Agents.Defaults.Model.Primary = modelName
		base := baseURL
		if base == "" {
			if defaultURL, ok := providerDefaultBaseURL(provider); ok {
				base = defaultURL
			}
		}
		plainKey := strings.TrimSpace(apiKey)
		cfg.Models = &modelsConfig{
			Mode: "merge",
			Providers: map[string]modelProvider{
				"kimi-coding": {
					ApiKey:  plainKey,
					BaseUrl: base,
					Api:     "anthropic-messages",
					Models: []modelEntry{
						{
							ID:            modelID,
							Name:          modelID,
							Reasoning:     true,
							Input:         []string{"text"},
							ContextWindow: 200000,
							MaxTokens:     8192,
							Cost:          modelCost{},
						},
					},
				},
			},
		}
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	configPath := path.Join(confDir, "openclaw.json")
	if err := fileOp.SaveFile(configPath, string(payload), 0600); err != nil {
		return err
	}

	envPath := path.Join(confDir, ".env")
	lines := []string{fmt.Sprintf("OPENCLAW_GATEWAY_TOKEN=%s", token)}
	if envKey := providerEnvKey(provider); envKey != "" && strings.TrimSpace(apiKey) != "" {
		lines = append(lines, fmt.Sprintf("%s=%s", envKey, apiKey))
	}
	content := strings.Join(lines, "\n") + "\n"
	return fileOp.SaveFile(envPath, content, 0600)
}

func providerEnvKey(provider string) string {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY"
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "minimax":
		return "MINIMAX_API_KEY"
	case "deepseek":
		return "DEEPSEEK_API_KEY"
	case "moonshot":
		return "MOONSHOT_API_KEY"
	case "kimi":
		return "KIMI_API_KEY"
	case "kimi-coding":
		return "KIMI_API_KEY"
	case "qwen":
		return "QWEN_API_KEY"
	case "ollama":
		return ""
	default:
		return ""
	}
}

type providerDefinition struct {
	Sort    uint
	BaseURL string
	Models  []dto.ProviderModelInfo
}

func providerDefinitions() map[string]providerDefinition {
	return map[string]providerDefinition{
		"deepseek": {
			Sort:    2,
			BaseURL: "https://api.deepseek.com/v1",
			Models: []dto.ProviderModelInfo{
				{ID: "deepseek/deepseek-chat", Name: "DeepSeek Chat"},
				{ID: "deepseek/deepseek-reasoner", Name: "DeepSeek Reasoner"},
				{ID: "deepseek/deepseek-r1:1.5b", Name: "DeepSeek R1 1.5B"},
			},
		},
		"openai": {
			Sort:    3,
			BaseURL: "https://api.openai.com/v1",
			Models: []dto.ProviderModelInfo{
				{ID: "openai/codex-mini-latest", Name: "Codex Mini"},
				{ID: "openai/gpt-4.1", Name: "GPT-4.1"},
				{ID: "openai/gpt-4o", Name: "GPT-4o"},
				{ID: "openai/gpt-4o-mini", Name: "GPT-4o Mini"},
				{ID: "openai/gpt-5", Name: "GPT-5"},
				{ID: "openai/gpt-5-mini", Name: "GPT-5 Mini"},
			},
		},
		"anthropic": {
			Sort:    4,
			BaseURL: "https://api.anthropic.com",
			Models: []dto.ProviderModelInfo{
				{ID: "anthropic/claude-3-haiku-20240307", Name: "Claude 3 Haiku"},
				{ID: "anthropic/claude-3-5-haiku-latest", Name: "Claude 3.5 Haiku"},
				{ID: "anthropic/claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet"},
				{ID: "anthropic/claude-3-7-sonnet-20250219", Name: "Claude 3.7 Sonnet"},
				{ID: "anthropic/claude-opus-4-1", Name: "Claude Opus 4.1"},
			},
		},
		"gemini": {
			Sort:    5,
			BaseURL: "https://generativelanguage.googleapis.com",
			Models: []dto.ProviderModelInfo{
				{ID: "google/gemini-1.5-flash", Name: "Gemini 1.5 Flash"},
				{ID: "google/gemini-1.5-pro", Name: "Gemini 1.5 Pro"},
				{ID: "google/gemini-2.0-flash", Name: "Gemini 2.0 Flash"},
				{ID: "google/gemini-2.5-flash", Name: "Gemini 2.5 Flash"},
				{ID: "google/gemini-2.5-pro", Name: "Gemini 2.5 Pro"},
				{ID: "google/gemini-3-flash-preview", Name: "Gemini 3 Flash Preview"},
			},
		},
		"ollama": {
			Sort:    1,
			BaseURL: "",
			Models:  []dto.ProviderModelInfo{},
		},
		"minimax": {
			Sort:    6,
			BaseURL: "https://api.minimaxi.com/v1",
			Models: []dto.ProviderModelInfo{
				{ID: "minimax/Minimax-M2.1", Name: "Minimax M2.1"},
			},
		},
		"moonshot": {
			Sort:    7,
			BaseURL: "https://api.moonshot.ai/v1",
			Models: []dto.ProviderModelInfo{
				{ID: "moonshot/kimi-k2.5", Name: "Kimi K2.5"},
				{ID: "moonshot/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
				{ID: "moonshot/kimi-k2-thinking", Name: "Kimi K2 Thinking"},
			},
		},
		"kimi": {
			Sort:    8,
			BaseURL: "https://api.moonshot.cn/v1",
			Models: []dto.ProviderModelInfo{
				{ID: "kimi/kimi-k2.5", Name: "Kimi K2.5"},
				{ID: "kimi/kimi-k2-0905-preview", Name: "Kimi K2 0905 Preview"},
				{ID: "kimi/kimi-k2-thinking", Name: "Kimi K2 Thinking"},
			},
		},
		"kimi-coding": {
			Sort:    9,
			BaseURL: "https://api.moonshot.cn/anthropic/v1",
			Models: []dto.ProviderModelInfo{
				{ID: "kimi-coding/k2p5", Name: "Kimi K2.5"},
			},
		},
	}
}

func providerDefaultBaseURL(provider string) (string, bool) {
	defs := providerDefinitions()
	if def, ok := defs[provider]; ok {
		if def.BaseURL == "" {
			return "", false
		}
		return def.BaseURL, true
	}
	return "", false
}

func isSupportedAgentProvider(provider string) bool {
	_, ok := providerDefinitions()[provider]
	return ok
}

func buildVerifyRequest(provider, baseURL, apiKey string) (string, map[string]string) {
	headers := map[string]string{}
	base := strings.TrimRight(baseURL, "/")
	switch provider {
	case "anthropic":
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
		if strings.Contains(base, "/v1") {
			return base + "/models", headers
		}
		return base + "/v1/models", headers
	case "kimi-coding":
		headers["x-api-key"] = apiKey
		headers["anthropic-version"] = "2023-06-01"
		if strings.Contains(base, "/v1") {
			return base + "/models", headers
		}
		return base + "/v1/models", headers
	case "gemini":
		if strings.Contains(base, "/v1beta") {
			return fmt.Sprintf("%s/models?key=%s", base, apiKey), headers
		}
		return fmt.Sprintf("%s/v1beta/models?key=%s", base, apiKey), headers
	default:
		headers["Authorization"] = fmt.Sprintf("Bearer %s", apiKey)
		if strings.Contains(base, "/v1") {
			return base + "/models", headers
		}
		return base + "/v1/models", headers
	}
}

func readInstallEnv(envStr string) map[string]interface{} {
	if strings.TrimSpace(envStr) == "" {
		return nil
	}
	data := map[string]interface{}{}
	if err := json.Unmarshal([]byte(envStr), &data); err != nil {
		return nil
	}
	return data
}

func maskKey(value string) string {
	trim := strings.TrimSpace(value)
	if len(trim) <= 6 {
		return trim
	}
	return fmt.Sprintf("%s****%s", trim[:3], trim[len(trim)-3:])
}

func toInt(value interface{}) int {
	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		if v == "" {
			return 0
		}
		parsed, _ := strconv.Atoi(v)
		return parsed
	default:
		return 0
	}
}

func randomToken() string {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
