package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	openclawutil "github.com/1Panel-dev/1Panel/agent/utils/openclaw"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"gorm.io/gorm"
)

type AgentService struct{}

type IAgentService interface {
	Create(req dto.AgentCreateReq) (*dto.AgentItem, error)
	Page(req dto.SearchWithPage) (int64, []dto.AgentItem, error)
	Delete(req dto.AgentDeleteReq) error
	ResetToken(req dto.AgentTokenResetReq) error
	UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error
	GetProviders() ([]dto.ProviderInfo, error)
	CreateAccount(req dto.AgentAccountCreateReq) error
	UpdateAccount(req dto.AgentAccountUpdateReq) error
	SyncAgentsByAccountID(accountID uint) error
	PageAccounts(req dto.AgentAccountSearch) (int64, []dto.AgentAccountInfo, error)
	GetAccountModels(req dto.AgentAccountModelReq) ([]dto.AgentAccountModel, error)
	CreateAccountModel(req dto.AgentAccountModelCreateReq) error
	UpdateAccountModel(req dto.AgentAccountModelUpdateReq) error
	DeleteAccountModel(req dto.AgentAccountModelDeleteReq) error
	VerifyAccount(req dto.AgentAccountVerifyReq) error
	DeleteAccount(req dto.AgentAccountDeleteReq) error
	GetFeishuConfig(req dto.AgentFeishuConfigReq) (*dto.AgentFeishuConfig, error)
	UpdateFeishuConfig(req dto.AgentFeishuConfigUpdateReq) error
	GetTelegramConfig(req dto.AgentTelegramConfigReq) (*dto.AgentTelegramConfig, error)
	UpdateTelegramConfig(req dto.AgentTelegramConfigUpdateReq) error
	GetDiscordConfig(req dto.AgentDiscordConfigReq) (*dto.AgentDiscordConfig, error)
	UpdateDiscordConfig(req dto.AgentDiscordConfigUpdateReq) error
	GetWecomConfig(req dto.AgentWecomConfigReq) (*dto.AgentWecomConfig, error)
	UpdateWecomConfig(req dto.AgentWecomConfigUpdateReq) error
	GetQQBotConfig(req dto.AgentQQBotConfigReq) (*dto.AgentQQBotConfig, error)
	UpdateQQBotConfig(req dto.AgentQQBotConfigUpdateReq) error
	InstallPlugin(req dto.AgentPluginInstallReq) error
	CheckPlugin(req dto.AgentPluginCheckReq) (*dto.AgentPluginStatus, error)
	GetSecurityConfig(req dto.AgentSecurityConfigReq) (*dto.AgentSecurityConfig, error)
	UpdateSecurityConfig(req dto.AgentSecurityConfigUpdateReq) error
	GetOtherConfig(req dto.AgentOtherConfigReq) (*dto.AgentOtherConfig, error)
	UpdateOtherConfig(req dto.AgentOtherConfigUpdateReq) error
	ApproveChannelPairing(req dto.AgentChannelPairingApproveReq) error
}

func NewIAgentService() IAgentService {
	return &AgentService{}
}

const (
	defaultBrowserExecutablePath  = "/home/node/.cache/ms-playwright/chromium-1208/chrome-linux64/chrome"
	defaultBrowserProfile         = "openclaw"
	defaultUserTimezone           = "Asia/Shanghai"
	defaultToolsProfile           = "full"
	defaultToolsSessionVisibility = "all"
	maxCommunityAIAgents          = int64(5)
	openclawPluginBaseDir         = "/home/node/.openclaw/extensions"
	openclawGatewayPort           = 18789
	openclawAllowedOriginHost     = "127.0.0.1"
	openclawHTTPSVersion          = "2026.3.13"
	openclawTrustedProxyLoopback  = "127.0.0.1/32"
)

func (a AgentService) Create(req dto.AgentCreateReq) (*dto.AgentItem, error) {
	agentType := normalizeAgentType(req.AgentType)
	if !isSupportedAgentType(agentType) {
		return nil, fmt.Errorf("agent type is invalid")
	}
	if err := checkPortExist(req.WebUIPort); err != nil {
		return nil, err
	}
	if exist, _ := agentRepo.GetFirst(repo.WithByLowerName(req.Name)); exist != nil && exist.ID > 0 {
		return nil, buserr.New("ErrNameIsExist")
	}
	if installs, _ := appInstallRepo.ListBy(context.Background(), repo.WithByLowerName(req.Name)); len(installs) > 0 {
		return nil, buserr.New("ErrNameIsExist")
	}
	if !xpack.IsXpack() {
		count, _, err := agentRepo.Page(1, 1)
		if err != nil {
			return nil, err
		}
		if count >= maxCommunityAIAgents {
			return nil, buserr.WithMap("ErrAgentLimitReached", map[string]interface{}{"max": maxCommunityAIAgents}, nil)
		}
	}
	appKey := constant.AppOpenclaw
	if agentType == constant.AppCopaw {
		appKey = constant.AppCopaw
	}
	app, err := appRepo.GetFirst(appRepo.WithKey(appKey))
	if err != nil || app.ID == 0 {
		return nil, buserr.New("ErrRecordNotFound")
	}
	detail, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(app.ID), appDetailRepo.WithVersion(req.AppVersion))
	if err != nil || detail.ID == 0 {
		return nil, buserr.New("ErrRecordNotFound")
	}

	provider := ""
	baseURL := ""
	apiType := ""
	maxTokens := 0
	contextWindow := 0
	apiKey := ""
	runtimeModel := ""
	accountID := uint(0)
	token := ""
	configPath := ""
	storedModel := ""
	var allowedOrigins []string

	if agentType == constant.AppOpenclaw {
		var err error
		allowedOrigins, err = normalizeAllowedOrigins(req.AllowedOrigins)
		if err != nil {
			return nil, err
		}
		if len(allowedOrigins) == 0 {
			return nil, fmt.Errorf("allowed origins is required")
		}
		provider = strings.ToLower(strings.TrimSpace(req.Provider))
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
		if !account.Verified && !providercatalog.SkipVerification(account.Provider) {
			return nil, buserr.New("ErrAgentAccountNotVerified")
		}
		if account.Provider != "" && provider != "" && account.Provider != provider {
			return nil, buserr.New("ErrAgentProviderMismatch")
		}
		provider = strings.ToLower(strings.TrimSpace(account.Provider))
		baseURL = strings.TrimSpace(account.BaseURL)
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
		accountModels, err := loadAgentAccountModels(account)
		if err != nil {
			return nil, err
		}
		storedModel = strings.TrimSpace(req.Model)
		if storedModel == "" {
			storedModel = strings.TrimSpace(account.Model)
		}
		if storedModel == "" && len(accountModels) > 0 {
			storedModel = strings.TrimSpace(accountModels[0].ID)
		}
		if storedModel == "" {
			return nil, buserr.New("ErrAgentModelNotInAccount")
		}
		selectedAccountModel, ok := findAgentAccountModelForProvider(provider, accountModels, storedModel)
		if !ok {
			return nil, buserr.New("ErrAgentModelNotInAccount")
		}
		storedModel = strings.TrimSpace(selectedAccountModel.ID)
		apiType, maxTokens, contextWindow = resolveRuntimeParams(
			provider,
			account.APIType,
			selectedAccountModel.MaxTokens,
			selectedAccountModel.ContextWindow,
		)
		runtimeModel, err = buildOpenclawPrimaryModel(account, storedModel)
		if err != nil {
			return nil, err
		}
		apiKey = account.APIKey
		accountID = account.ID
		token = strings.TrimSpace(req.Token)
		if token == "" {
			token = generateToken()
		}
	}

	params := map[string]interface{}{
		constant.CPUS:        "0",
		constant.MemoryLimit: "0",
		constant.HostIP:      "",
	}
	if agentType == constant.AppOpenclaw {
		params["PANEL_APP_PORT_HTTPS"] = req.WebUIPort
		if allowedOrigin := firstAllowedOrigin(allowedOrigins); allowedOrigin != "" {
			params["ALLOWED_ORIGIN"] = allowedOrigin
		}
		params["PROVIDER"] = provider
		params["MODEL"] = runtimeModel
		params["API_TYPE"] = apiType
		params["MAX_TOKENS"] = maxTokens
		params["CONTEXT_WINDOW"] = contextWindow
		params["BASE_URL"] = baseURL
		params["API_KEY"] = apiKey
		params["OPENCLAW_GATEWAY_TOKEN"] = token
	} else {
		params["PANEL_APP_PORT_HTTP"] = req.WebUIPort
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
	if agentType == constant.AppOpenclaw {
		configPath = path.Join(appInstall.GetPath(), "data", "conf", "openclaw.json")
	}
	agent := &model.Agent{
		Name:          req.Name,
		AgentType:     agentType,
		Provider:      provider,
		Model:         storedModel,
		APIType:       apiType,
		MaxTokens:     maxTokens,
		ContextWindow: contextWindow,
		BaseURL:       baseURL,
		APIKey:        apiKey,
		Token:         token,
		Status:        appInstall.Status,
		Message:       appInstall.Message,
		AppInstallID:  appInstall.ID,
		AccountID:     accountID,
		ConfigPath:    configPath,
	}
	if err := agentRepo.Create(agent); err != nil {
		return nil, err
	}
	if agentType == constant.AppOpenclaw {
		go a.writeConfigWithRetry(
			appInstall,
			accountID,
			storedModel,
			token,
			agent.ID,
			allowedOrigins,
		)
	}

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

func (a AgentService) ResetToken(req dto.AgentTokenResetReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if normalizeAgentType(agent.AgentType) == constant.AppCopaw {
		return fmt.Errorf("copaw does not support token")
	}
	configPath := strings.TrimSpace(agent.ConfigPath)
	if configPath == "" && agent.AppInstallID > 0 {
		install, err := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
		if err != nil {
			return err
		}
		configPath = path.Join(install.GetPath(), "data", "conf", "openclaw.json")
	}
	if configPath == "" {
		return buserr.New("ErrRecordNotFound")
	}
	conf, err := readOpenclawConfig(configPath)
	if err != nil {
		return err
	}
	newToken := generateToken()
	if newToken == "" {
		return fmt.Errorf("generate token failed")
	}
	gatewayMap := ensureChildMap(conf, "gateway")
	authMap := ensureChildMap(gatewayMap, "auth")
	if _, ok := authMap["mode"]; !ok {
		authMap["mode"] = "token"
	}
	authMap["token"] = newToken
	if err := writeOpenclawConfigRaw(configPath, conf); err != nil {
		return err
	}
	agent.Token = newToken
	if agent.ConfigPath == "" {
		agent.ConfigPath = configPath
	}
	return agentRepo.Save(agent)
}

func (a AgentService) UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return err
	}
	if normalizeAgentType(agent.AgentType) == constant.AppCopaw {
		return fmt.Errorf("copaw does not support model config")
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return err
	}
	if !account.Verified && !providercatalog.SkipVerification(account.Provider) {
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
	if provider != "custom" && provider != "vllm" && !modelMatchesProvider(provider, modelName) {
		return buserr.New("ErrAgentProviderMismatch")
	}
	baseURL := strings.TrimSpace(account.BaseURL)
	if baseURL == "" {
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
	accountModels, err := loadAgentAccountModels(account)
	if err != nil {
		return err
	}
	selectedAccountModel, ok := findAgentAccountModelForProvider(provider, accountModels, modelName)
	if !ok {
		return buserr.New("ErrAgentModelNotInAccount")
	}
	modelName = strings.TrimSpace(selectedAccountModel.ID)
	apiType, maxTokens, contextWindow := resolveRuntimeParams(
		provider,
		account.APIType,
		selectedAccountModel.MaxTokens,
		selectedAccountModel.ContextWindow,
	)
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
	if err := writeOpenclawConfig(confDir, account, modelName, agent.Token, nil); err != nil {
		return err
	}
	agent.Provider = provider
	agent.Model = modelName
	agent.APIType = apiType
	agent.MaxTokens = maxTokens
	agent.ContextWindow = contextWindow
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
			Sort:        def.Sort,
			Provider:    key,
			DisplayName: def.DisplayName,
			BaseURL:     def.BaseURL,
			Models:      def.Models,
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
	if fixedURL, ok := fixedProviderBaseURL(provider); ok {
		baseURL = fixedURL
	}
	if (provider == "custom" || provider == "vllm") && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if provider != "custom" && provider != "vllm" && baseURL == "" {
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
	modelName := strings.TrimSpace(req.Model)
	apiType := normalizeAPIType(req.APIType)
	if provider == "custom" || provider == "vllm" {
		if !isSupportedAPIType(apiType) {
			return fmt.Errorf("apiType is invalid")
		}
	}
	if provider == "ollama" {
		if !isSupportedOllamaAPIType(apiType) {
			return fmt.Errorf("apiType is invalid")
		}
	}
	if err := a.VerifyAccount(dto.AgentAccountVerifyReq{Provider: provider, BaseURL: baseURL, APIKey: apiKey}); err != nil {
		return err
	}
	verified := !providercatalog.SkipVerification(provider)
	_, maxTokens, contextWindow := resolveRuntimeParams(provider, apiType, req.MaxTokens, req.ContextWindow)
	account := &model.AgentAccount{
		Provider:       provider,
		Name:           req.Name,
		APIKey:         apiKey,
		RememberAPIKey: req.RememberAPIKey,
		BaseURL:        baseURL,
		Model:          "",
		Models:         "",
		APIType:        apiType,
		MaxTokens:      0,
		ContextWindow:  0,
		Verified:       verified,
		Remark:         req.Remark,
	}
	if provider == "custom" || provider == "vllm" || provider == "ollama" {
		account.MaxTokens = maxTokens
		account.ContextWindow = contextWindow
	}
	if err := agentAccountRepo.Create(account); err != nil {
		return err
	}
	initialModels, err := buildInitialAgentAccountModels(account, req.Models, modelName)
	if err != nil {
		_ = agentAccountRepo.DeleteByID(account.ID)
		return err
	}
	if len(initialModels) > 0 {
		if err := replacePersistedAgentAccountModels(account.ID, initialModels); err != nil {
			_ = agentAccountRepo.DeleteByID(account.ID)
			return err
		}
	}
	asyncReportAIProviderInstall(provider)
	return nil
}

func (a AgentService) UpdateAccount(req dto.AgentAccountUpdateReq) error {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	provider := strings.ToLower(strings.TrimSpace(account.Provider))
	baseURL := strings.TrimSpace(req.BaseURL)
	if fixedURL, ok := fixedProviderBaseURL(provider); ok {
		baseURL = fixedURL
	}
	if (provider == "custom" || provider == "vllm") && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if provider != "custom" && provider != "vllm" && baseURL == "" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	apiType := normalizeAPIType(req.APIType)
	rawAPIType := strings.TrimSpace(req.APIType)
	if (provider == "custom" || provider == "vllm") && !isSupportedAPIType(apiType) {
		return fmt.Errorf("apiType is invalid")
	}
	if provider == "ollama" {
		if rawAPIType == "" {
			apiType = normalizeAPIType(account.APIType)
			if !isSupportedOllamaAPIType(apiType) {
				apiType = "openai-responses"
			}
		} else if !isSupportedOllamaAPIType(apiType) {
			return fmt.Errorf("apiType is invalid")
		}
	}
	if provider != "custom" && provider != "vllm" && provider != "ollama" {
		apiType = normalizeAPIType(account.APIType)
	}
	_, maxTokens, contextWindow := resolveRuntimeParams(provider, apiType, req.MaxTokens, req.ContextWindow)
	if err := a.VerifyAccount(dto.AgentAccountVerifyReq{Provider: provider, BaseURL: baseURL, APIKey: req.APIKey}); err != nil {
		return err
	}
	verified := !providercatalog.SkipVerification(provider)
	account.Provider = provider
	account.APIKey = req.APIKey
	account.BaseURL = baseURL
	account.APIType = apiType
	account.MaxTokens = 0
	account.ContextWindow = 0
	if provider == "custom" || provider == "vllm" || provider == "ollama" {
		account.MaxTokens = maxTokens
		account.ContextWindow = contextWindow
	}
	account.Name = req.Name
	account.APIKey = req.APIKey
	account.RememberAPIKey = req.RememberAPIKey
	account.BaseURL = baseURL
	account.APIType = apiType
	account.Remark = req.Remark
	account.Verified = verified

	var nextAccountModels []dto.AgentAccountModel
	if len(req.Models) > 0 || strings.TrimSpace(req.Model) != "" {
		accountModels, _, err := normalizeAgentAccountModels(account, req.Models, req.Model, true)
		if err != nil {
			return err
		}
		nextAccountModels = accountModels
		if err := ensureAccountModelsNotBound(account, nextAccountModels); err != nil {
			return err
		}
	}
	if err := agentAccountRepo.Save(account); err != nil {
		return err
	}
	if len(nextAccountModels) > 0 {
		if err := replacePersistedAgentAccountModels(account.ID, nextAccountModels); err != nil {
			return err
		}
	} else if shouldRefreshAccountModelRuntimeLimits(provider) {
		accountModels, err := loadAgentAccountModels(account)
		if err != nil {
			return err
		}
		if len(accountModels) > 0 {
			accountModels = refreshAccountModelRuntimeLimits(account, accountModels)
			if err := replacePersistedAgentAccountModels(account.ID, accountModels); err != nil {
				return err
			}
		}
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
		apiKey := ""
		if item.RememberAPIKey {
			apiKey = item.APIKey
		}
		items = append(items, dto.AgentAccountInfo{
			ID:             item.ID,
			Provider:       item.Provider,
			ProviderName:   providerDisplayName(item.Provider),
			Name:           item.Name,
			APIKey:         apiKey,
			RememberAPIKey: item.RememberAPIKey,
			BaseURL:        item.BaseURL,
			Model:          "",
			Models:         nil,
			APIType:        item.APIType,
			MaxTokens:      item.MaxTokens,
			ContextWindow:  item.ContextWindow,
			Verified:       item.Verified,
			Remark:         item.Remark,
			CreatedAt:      item.CreatedAt,
		})
	}
	for i := range items {
		models, err := loadAgentAccountModels(&list[i])
		if err != nil {
			return 0, nil, err
		}
		items[i].Models = models
	}
	return count, items, nil
}

func (a AgentService) GetAccountModels(req dto.AgentAccountModelReq) ([]dto.AgentAccountModel, error) {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return nil, err
	}
	return loadAgentAccountModels(account)
}

func (a AgentService) CreateAccountModel(req dto.AgentAccountModelCreateReq) error {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return err
	}
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return err
	}
	normalized, err := normalizeAgentAccountModel(account, req.Model)
	if err != nil {
		return err
	}
	if _, ok := findAgentAccountModelForProvider(account.Provider, models, normalized.ID); ok {
		return buserr.New("ErrRecordExist")
	}
	inputPayload, err := json.Marshal(sanitizeAgentAccountModelInputs(normalized.Input))
	if err != nil {
		return err
	}
	sortOrder := len(models) + 1
	record := &model.AgentAccountModel{
		AccountID:     account.ID,
		Model:         normalized.ID,
		Name:          normalized.Name,
		ContextWindow: normalized.ContextWindow,
		MaxTokens:     normalized.MaxTokens,
		Reasoning:     normalized.Reasoning,
		Input:         string(inputPayload),
		SortOrder:     sortOrder,
	}
	if err := agentAccountModelRepo.Create(record); err != nil {
		return err
	}
	return a.syncAgentsByAccount(account)
}

func (a AgentService) UpdateAccountModel(req dto.AgentAccountModelUpdateReq) error {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return err
	}
	record, err := agentAccountModelRepo.GetFirst(repo.WithByID(req.Model.RecordID), repo.WithByAccountID(req.AccountID))
	if err != nil {
		return err
	}
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return err
	}
	normalized, err := normalizeAgentAccountModel(account, req.Model)
	if err != nil {
		return err
	}
	for _, item := range models {
		if item.RecordID == req.Model.RecordID {
			continue
		}
		if sameProviderModelID(account.Provider, item.ID, normalized.ID) {
			return buserr.New("ErrRecordExist")
		}
	}
	nextModels := make([]dto.AgentAccountModel, 0, len(models))
	for _, item := range models {
		if item.RecordID == req.Model.RecordID {
			nextModels = append(nextModels, normalized)
			continue
		}
		nextModels = append(nextModels, item)
	}
	if err := ensureAccountModelsNotBound(account, nextModels); err != nil {
		return err
	}
	inputPayload, err := json.Marshal(sanitizeAgentAccountModelInputs(normalized.Input))
	if err != nil {
		return err
	}
	record.Model = normalized.ID
	record.Name = normalized.Name
	record.ContextWindow = normalized.ContextWindow
	record.MaxTokens = normalized.MaxTokens
	record.Reasoning = normalized.Reasoning
	record.Input = string(inputPayload)
	if err := agentAccountModelRepo.Save(record); err != nil {
		return err
	}
	return a.syncAgentsByAccount(account)
}

func (a AgentService) DeleteAccountModel(req dto.AgentAccountModelDeleteReq) error {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return err
	}
	if _, err := agentAccountModelRepo.GetFirst(repo.WithByID(req.RecordID), repo.WithByAccountID(req.AccountID)); err != nil {
		return err
	}
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return err
	}
	nextModels := make([]dto.AgentAccountModel, 0, len(models))
	for _, item := range models {
		if item.RecordID == req.RecordID {
			continue
		}
		nextModels = append(nextModels, item)
	}
	if err := ensureAccountModelsNotBound(account, nextModels); err != nil {
		return err
	}
	if err := agentAccountModelRepo.DeleteByID(req.RecordID); err != nil {
		return err
	}
	if err := compactPersistedAgentAccountModelSortOrder(req.AccountID); err != nil {
		return err
	}
	return a.syncAgentsByAccount(account)
}

func (a AgentService) SyncAgentsByAccountID(accountID uint) error {
	if accountID == 0 {
		return nil
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(accountID))
	if err != nil {
		return err
	}
	return a.syncAgentsByAccount(account)
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
	if fixedURL, ok := fixedProviderBaseURL(provider); ok {
		baseURL = fixedURL
	}
	if baseURL == "" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	if provider == "ollama" && baseURL == "" {
		return buserr.New("ErrAgentBaseURLRequired")
	}
	if providercatalog.SkipVerification(provider) {
		return nil
	}
	return providercatalog.VerifyAccount(provider, baseURL, apiKey)
}

func (a AgentService) DeleteAccount(req dto.AgentAccountDeleteReq) error {
	if req.ID == 0 {
		return buserr.New("ErrAgentAccountIDRequired")
	}
	if exists, _ := agentRepo.GetFirst(repo.WithByAccountID(req.ID)); exists != nil && exists.ID > 0 {
		return buserr.New("ErrAgentAccountBound")
	}
	if err := agentAccountModelRepo.Delete(repo.WithByAccountID(req.ID)); err != nil {
		return err
	}
	return agentAccountRepo.DeleteByID(req.ID)
}

func (a AgentService) GetFeishuConfig(req dto.AgentFeishuConfigReq) (*dto.AgentFeishuConfig, error) {
	_, _, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	result := extractFeishuConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateFeishuConfig(req dto.AgentFeishuConfigUpdateReq) error {
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		dmPolicy := req.DmPolicy
		if dmPolicy == "" {
			dmPolicy = "pairing"
		}
		setFeishuConfig(conf, dto.AgentFeishuConfig{
			Enabled:   req.Enabled,
			DmPolicy:  dmPolicy,
			BotName:   req.BotName,
			AppID:     req.AppID,
			AppSecret: req.AppSecret,
		})
		setFeishuPluginEnabled(conf, req.Enabled)
		return nil
	})
}

func (a AgentService) GetTelegramConfig(req dto.AgentTelegramConfigReq) (*dto.AgentTelegramConfig, error) {
	_, _, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	result := extractTelegramConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateTelegramConfig(req dto.AgentTelegramConfigUpdateReq) error {
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		dmPolicy := req.DmPolicy
		if dmPolicy == "" {
			dmPolicy = "pairing"
		}
		setTelegramConfig(conf, dto.AgentTelegramConfig{
			Enabled:  req.Enabled,
			DmPolicy: dmPolicy,
			BotToken: req.BotToken,
			Proxy:    req.Proxy,
		})
		return nil
	})
}

func (a AgentService) GetDiscordConfig(req dto.AgentDiscordConfigReq) (*dto.AgentDiscordConfig, error) {
	_, _, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	result := extractDiscordConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateDiscordConfig(req dto.AgentDiscordConfigUpdateReq) error {
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		dmPolicy := req.DmPolicy
		if dmPolicy == "" {
			dmPolicy = "pairing"
		}
		groupPolicy := req.GroupPolicy
		if groupPolicy == "" {
			groupPolicy = "open"
		}
		setDiscordConfig(conf, dto.AgentDiscordConfig{
			Enabled:     req.Enabled,
			DmPolicy:    dmPolicy,
			GroupPolicy: groupPolicy,
			Token:       req.Token,
			Proxy:       req.Proxy,
		})
		return nil
	})
}

func (a AgentService) GetQQBotConfig(req dto.AgentQQBotConfigReq) (*dto.AgentQQBotConfig, error) {
	_, install, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	result := extractQQBotConfig(conf)
	installed, _ := checkPluginInstalled(install.ContainerName, "qqbot")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateQQBotConfig(req dto.AgentQQBotConfigUpdateReq) error {
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setQQBotConfig(conf, dto.AgentQQBotConfig{
			Enabled:      req.Enabled,
			AppID:        req.AppID,
			ClientSecret: req.ClientSecret,
		})
		return nil
	})
}

func (a AgentService) GetWecomConfig(req dto.AgentWecomConfigReq) (*dto.AgentWecomConfig, error) {
	_, install, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	result := extractWecomConfig(conf)
	installed, _ := checkPluginInstalled(install.ContainerName, "wecom")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateWecomConfig(req dto.AgentWecomConfigUpdateReq) error {
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setWecomConfig(conf, dto.AgentWecomConfig{
			Enabled:  req.Enabled,
			DmPolicy: req.DmPolicy,
			BotID:    req.BotID,
			Secret:   req.Secret,
		})
		return nil
	})
}

func (a AgentService) InstallPlugin(req dto.AgentPluginInstallReq) error {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	spec, _, err := resolvePluginMeta(req.Type)
	if err != nil {
		return err
	}
	installTask, err := task.NewTaskWithOps(req.Type, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	installTask.AddSubTask("Install OpenClaw plugin", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		return mgr.RunBashCf("docker exec %s openclaw plugins install %s", install.ContainerName, spec)
	}, nil)
	go func() {
		if err := installTask.Execute(); err != nil {
			global.LOG.Errorf("install openclaw plugin failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) CheckPlugin(req dto.AgentPluginCheckReq) (*dto.AgentPluginStatus, error) {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	installed, err := checkPluginInstalled(install.ContainerName, req.Type)
	if err != nil {
		return nil, err
	}
	return &dto.AgentPluginStatus{Installed: installed}, nil
}

func (a AgentService) GetSecurityConfig(req dto.AgentSecurityConfigReq) (*dto.AgentSecurityConfig, error) {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if normalizeAgentType(agent.AgentType) == constant.AppCopaw {
		return nil, fmt.Errorf("copaw does not support security config")
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractSecurityConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateSecurityConfig(req dto.AgentSecurityConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if normalizeAgentType(agent.AgentType) == constant.AppCopaw {
		return fmt.Errorf("copaw does not support security config")
	}
	allowedOrigins, err := normalizeAllowedOrigins(req.AllowedOrigins)
	if err != nil {
		return err
	}
	if len(allowedOrigins) == 0 {
		return fmt.Errorf("allowed origins is required")
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	setSecurityConfig(conf, dto.AgentSecurityConfig{AllowedOrigins: allowedOrigins})
	if err := writeOpenclawConfigRaw(agent.ConfigPath, conf); err != nil {
		return err
	}
	if err := syncOpenclawAllowedOriginEnv(install, allowedOrigins); err != nil {
		return err
	}
	return appInstallRepo.Save(context.Background(), install)
}

func (a AgentService) GetOtherConfig(req dto.AgentOtherConfigReq) (*dto.AgentOtherConfig, error) {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractOtherConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateOtherConfig(req dto.AgentOtherConfigUpdateReq) error {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return err
	}
	setOtherConfig(conf, dto.AgentOtherConfig{
		UserTimezone:   strings.TrimSpace(req.UserTimezone),
		BrowserEnabled: req.BrowserEnabled,
	})
	if err := writeOpenclawConfigRaw(agent.ConfigPath, conf); err != nil {
		return err
	}
	return nil
}

func (a AgentService) ApproveChannelPairing(req dto.AgentChannelPairingApproveReq) error {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	channelType := strings.ToLower(strings.TrimSpace(req.Type))
	if channelType == "" {
		channelType = "feishu"
	}
	if channelType != "feishu" && channelType != "telegram" && channelType != "discord" && channelType != "wecom" {
		return fmt.Errorf("unsupported channel type: %s", channelType)
	}
	if err := cmd.RunDefaultBashCf(
		"docker exec %s openclaw pairing approve %s %q",
		install.ContainerName,
		channelType,
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

func (a AgentService) loadAgentConfig(agentID uint) (*model.Agent, *model.AppInstall, map[string]interface{}, error) {
	agent, install, err := a.loadAgentAndInstall(agentID)
	if err != nil {
		return nil, nil, nil, err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, nil, nil, err
	}
	return agent, install, conf, nil
}

func (a AgentService) mutateAgentConfig(agentID uint, fn func(agent *model.Agent, install *model.AppInstall, conf map[string]interface{}) error) error {
	agent, install, conf, err := a.loadAgentConfig(agentID)
	if err != nil {
		return err
	}
	if err := fn(agent, install, conf); err != nil {
		return err
	}
	return writeOpenclawConfigRaw(agent.ConfigPath, conf)
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
	ensureGatewaySecurityDefaults(conf)
	ensureOpenclawUpdateDefaults(conf)
	payload, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	return fileOp.SaveFile(configPath, string(payload), 0600)
}

func normalizeAllowedOrigins(origins []string) ([]string, error) {
	if len(origins) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(origins))
	seen := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}
		normalized, err := normalizeAllowedOrigin(origin)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result, nil
}

func normalizeAllowedOrigin(origin string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(origin))
	if err != nil {
		return "", fmt.Errorf("invalid allowed origin: %s", origin)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("invalid allowed origin: %s", origin)
	}
	if parsed.User != nil || parsed.Host == "" || parsed.Hostname() == "" {
		return "", fmt.Errorf("invalid allowed origin: %s", origin)
	}
	if parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid allowed origin: %s", origin)
	}
	if pathValue := strings.TrimSpace(parsed.EscapedPath()); pathValue != "" && pathValue != "/" {
		return "", fmt.Errorf("invalid allowed origin: %s", origin)
	}
	host := parsed.Hostname()
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	normalized := parsed.Scheme + "://" + host
	if parsed.Port() != "" {
		normalized += ":" + parsed.Port()
	}
	return normalized, nil
}

func extractSecurityConfig(conf map[string]interface{}) dto.AgentSecurityConfig {
	result := dto.AgentSecurityConfig{AllowedOrigins: []string{}}
	gateway, ok := conf["gateway"].(map[string]interface{})
	if !ok {
		return result
	}
	controlUi, ok := gateway["controlUi"].(map[string]interface{})
	if !ok {
		return result
	}
	switch values := controlUi["allowedOrigins"].(type) {
	case []interface{}:
		for _, value := range values {
			if text, ok := value.(string); ok && strings.TrimSpace(text) != "" {
				result.AllowedOrigins = append(result.AllowedOrigins, strings.TrimSpace(text))
			}
		}
	case []string:
		for _, value := range values {
			if strings.TrimSpace(value) != "" {
				result.AllowedOrigins = append(result.AllowedOrigins, strings.TrimSpace(value))
			}
		}
	}
	return result
}

func setSecurityConfig(conf map[string]interface{}, config dto.AgentSecurityConfig) {
	ensureGatewaySecurityDefaults(conf)
	gateway := ensureChildMap(conf, "gateway")
	controlUi := ensureChildMap(gateway, "controlUi")
	allowedOrigins := append([]string(nil), config.AllowedOrigins...)
	if len(allowedOrigins) > 0 {
		controlUi["allowedOrigins"] = allowedOrigins
	} else {
		delete(controlUi, "allowedOrigins")
	}
}

func ensureGatewaySecurityDefaults(conf map[string]interface{}) {
	gateway := ensureChildMap(conf, "gateway")
	controlUi := ensureChildMap(gateway, "controlUi")
	if _, ok := controlUi["dangerouslyDisableDeviceAuth"]; !ok {
		controlUi["dangerouslyDisableDeviceAuth"] = true
	}
	delete(controlUi, "dangerouslyAllowHostHeaderOriginFallback")
	setTrustedProxies(gateway)
}

func ensureOpenclawUpdateDefaults(conf map[string]interface{}) {
	update := ensureChildMap(conf, "update")
	if _, ok := update["checkOnStart"]; !ok {
		update["checkOnStart"] = false
	}
}

func setTrustedProxies(gateway map[string]interface{}) {
	proxies := make([]string, 0, 4)
	seen := map[string]struct{}{}
	switch values := gateway["trustedProxies"].(type) {
	case []interface{}:
		for _, value := range values {
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			seen[text] = struct{}{}
			proxies = append(proxies, text)
		}
	case []string:
		for _, value := range values {
			text := strings.TrimSpace(value)
			if text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			seen[text] = struct{}{}
			proxies = append(proxies, text)
		}
	}
	if _, ok := seen[openclawTrustedProxyLoopback]; !ok {
		proxies = append(proxies, openclawTrustedProxyLoopback)
	}
	gateway["trustedProxies"] = proxies
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
	channels := ensureChildMap(conf, "channels")
	feishu := ensureChildMap(channels, "feishu")
	feishu["enabled"] = config.Enabled
	feishu["dmPolicy"] = config.DmPolicy

	accounts := ensureChildMap(feishu, "accounts")
	main := ensureChildMap(accounts, "main")
	main["appId"] = config.AppID
	main["appSecret"] = config.AppSecret
	main["botName"] = config.BotName

	if strings.EqualFold(config.DmPolicy, "open") {
		feishu["allowFrom"] = []string{"*"}
	}
}

func setFeishuPluginEnabled(conf map[string]interface{}, enabled bool) {
	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	feishu := ensureChildMap(entries, "feishu")
	feishu["enabled"] = enabled
}

func extractTelegramConfig(conf map[string]interface{}) dto.AgentTelegramConfig {
	result := dto.AgentTelegramConfig{Enabled: true, DmPolicy: "pairing"}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	telegram, ok := channels["telegram"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := telegram["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy, ok := telegram["dmPolicy"].(string); ok && strings.TrimSpace(dmPolicy) != "" {
		result.DmPolicy = dmPolicy
	}
	if botToken, ok := telegram["botToken"].(string); ok {
		result.BotToken = botToken
	}
	if proxy, ok := telegram["proxy"].(string); ok {
		result.Proxy = proxy
	}
	return result
}

func setTelegramConfig(conf map[string]interface{}, config dto.AgentTelegramConfig) {
	channels := ensureChildMap(conf, "channels")
	telegram := map[string]interface{}{
		"enabled":  config.Enabled,
		"dmPolicy": config.DmPolicy,
		"botToken": config.BotToken,
	}
	if strings.EqualFold(config.DmPolicy, "open") {
		telegram["allowFrom"] = []string{"*"}
	}
	if strings.TrimSpace(config.Proxy) != "" {
		telegram["proxy"] = strings.TrimSpace(config.Proxy)
	}
	channels["telegram"] = telegram
}

func extractDiscordConfig(conf map[string]interface{}) dto.AgentDiscordConfig {
	result := dto.AgentDiscordConfig{Enabled: true, DmPolicy: "pairing", GroupPolicy: "open"}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	discord, ok := channels["discord"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := discord["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if token, ok := discord["token"].(string); ok {
		result.Token = token
	}
	if groupPolicy, ok := discord["groupPolicy"].(string); ok && strings.TrimSpace(groupPolicy) != "" {
		result.GroupPolicy = groupPolicy
	}
	if proxy, ok := discord["proxy"].(string); ok {
		result.Proxy = proxy
	}
	if policy, ok := discord["dmPolicy"].(string); ok && strings.TrimSpace(policy) != "" {
		result.DmPolicy = policy
		return result
	}
	// backward compatibility: old nested style
	dm, ok := discord["dm"].(map[string]interface{})
	if ok {
		if policy, ok := dm["policy"].(string); ok && strings.TrimSpace(policy) != "" {
			result.DmPolicy = policy
		}
	}
	return result
}

func setDiscordConfig(conf map[string]interface{}, config dto.AgentDiscordConfig) {
	channels := ensureChildMap(conf, "channels")
	discord := ensureChildMap(channels, "discord")
	discord["enabled"] = config.Enabled
	discord["token"] = config.Token
	discord["dmPolicy"] = config.DmPolicy
	discord["groupPolicy"] = config.GroupPolicy
	if strings.EqualFold(config.DmPolicy, "open") {
		discord["allowFrom"] = []string{"*"}
	} else {
		delete(discord, "allowFrom")
	}
	if strings.TrimSpace(config.Proxy) != "" {
		discord["proxy"] = strings.TrimSpace(config.Proxy)
	} else {
		delete(discord, "proxy")
	}
	delete(discord, "dm")
}

func extractBrowserConfig(conf map[string]interface{}) browserConfig {
	result := browserConfig{
		Enabled:        true,
		ExecutablePath: defaultBrowserExecutablePath,
		Headless:       true,
		NoSandbox:      true,
		DefaultProfile: defaultBrowserProfile,
	}
	browser, ok := conf["browser"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := browser["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if executablePath, ok := browser["executablePath"].(string); ok && strings.TrimSpace(executablePath) != "" {
		result.ExecutablePath = executablePath
	}
	if headless, ok := browser["headless"].(bool); ok {
		result.Headless = headless
	}
	if noSandbox, ok := browser["noSandbox"].(bool); ok {
		result.NoSandbox = noSandbox
	}
	if defaultProfile, ok := browser["defaultProfile"].(string); ok && strings.TrimSpace(defaultProfile) != "" {
		result.DefaultProfile = defaultProfile
	}
	return result
}

func setBrowserConfig(conf map[string]interface{}, config browserConfig) {
	browser := ensureChildMap(conf, "browser")
	browser["enabled"] = config.Enabled
	browser["executablePath"] = defaultBrowserExecutablePath
	browser["headless"] = config.Headless
	browser["noSandbox"] = config.NoSandbox
	if strings.TrimSpace(config.DefaultProfile) == "" {
		browser["defaultProfile"] = defaultBrowserProfile
	} else {
		browser["defaultProfile"] = strings.TrimSpace(config.DefaultProfile)
	}
}

func extractQQBotConfig(conf map[string]interface{}) dto.AgentQQBotConfig {
	result := dto.AgentQQBotConfig{Enabled: true}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	qqbot, ok := channels["qqbot"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := qqbot["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if appID, ok := qqbot["appId"].(string); ok {
		result.AppID = appID
	}
	if clientSecret, ok := qqbot["clientSecret"].(string); ok {
		result.ClientSecret = clientSecret
	}
	return result
}

func extractWecomConfig(conf map[string]interface{}) dto.AgentWecomConfig {
	result := dto.AgentWecomConfig{Enabled: true, DmPolicy: "pairing"}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	wecom, ok := channels["wecom"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := wecom["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if dmPolicy, ok := wecom["dmPolicy"].(string); ok && strings.TrimSpace(dmPolicy) != "" {
		result.DmPolicy = strings.TrimSpace(dmPolicy)
	}
	if botID, ok := wecom["botId"].(string); ok {
		result.BotID = botID
	}
	if secret, ok := wecom["secret"].(string); ok {
		result.Secret = secret
	}
	return result
}

func setWecomConfig(conf map[string]interface{}, config dto.AgentWecomConfig) {
	channels := ensureChildMap(conf, "channels")
	wecom := ensureChildMap(channels, "wecom")
	wecom["enabled"] = config.Enabled
	wecom["botId"] = strings.TrimSpace(config.BotID)
	wecom["secret"] = strings.TrimSpace(config.Secret)
	wecom["dmPolicy"] = strings.TrimSpace(config.DmPolicy)
	if strings.EqualFold(config.DmPolicy, "open") {
		wecom["allowFrom"] = []string{"*"}
	} else {
		wecom["allowFrom"] = []string{}
	}

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	wecomEntry := ensureChildMap(entries, "wecom-openclaw-plugin")
	wecomEntry["enabled"] = config.Enabled
}

func setQQBotConfig(conf map[string]interface{}, config dto.AgentQQBotConfig) {
	channels := ensureChildMap(conf, "channels")
	qqbot := ensureChildMap(channels, "qqbot")
	qqbot["enabled"] = config.Enabled
	qqbot["allowFrom"] = []string{"*"}
	qqbot["appId"] = strings.TrimSpace(config.AppID)
	qqbot["clientSecret"] = strings.TrimSpace(config.ClientSecret)

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	qqbotEntry := ensureChildMap(entries, "qqbot")
	qqbotEntry["enabled"] = config.Enabled
}

func resolvePluginMeta(pluginType string) (string, string, error) {
	switch strings.ToLower(strings.TrimSpace(pluginType)) {
	case "qqbot":
		return "@sliverp/qqbot@latest", "qqbot", nil
	case "wecom":
		return "@wecom/wecom-openclaw-plugin", "wecom-openclaw-plugin", nil
	default:
		return "", "", fmt.Errorf("unsupported plugin type")
	}
}

func checkPluginInstalled(containerName, pluginType string) (bool, error) {
	_, pluginDir, err := resolvePluginMeta(pluginType)
	if err != nil {
		return false, err
	}
	if strings.TrimSpace(containerName) == "" {
		return false, buserr.New("ErrRecordNotFound")
	}
	pluginPath := path.Join(openclawPluginBaseDir, pluginDir)
	mgr := cmd.NewCommandMgr(cmd.WithTimeout(20 * time.Second))
	if err := mgr.RunBashCf("docker exec %s test -d %s", containerName, pluginPath); err != nil {
		return false, nil
	}
	return true, nil
}

func extractOtherConfig(conf map[string]interface{}) dto.AgentOtherConfig {
	result := dto.AgentOtherConfig{UserTimezone: resolveServerTimezone(), BrowserEnabled: true}
	agents, ok := conf["agents"].(map[string]interface{})
	if !ok {
		browser := extractBrowserConfig(conf)
		result.BrowserEnabled = browser.Enabled
		return result
	}
	defaults, ok := agents["defaults"].(map[string]interface{})
	if !ok {
		browser := extractBrowserConfig(conf)
		result.BrowserEnabled = browser.Enabled
		return result
	}
	if timezone, ok := defaults["userTimezone"].(string); ok && strings.TrimSpace(timezone) != "" {
		result.UserTimezone = strings.TrimSpace(timezone)
	}
	browser := extractBrowserConfig(conf)
	result.BrowserEnabled = browser.Enabled
	return result
}

func setOtherConfig(conf map[string]interface{}, config dto.AgentOtherConfig) {
	agents := ensureChildMap(conf, "agents")
	defaults := ensureChildMap(agents, "defaults")
	timezone := strings.TrimSpace(config.UserTimezone)
	if timezone == "" {
		timezone = resolveServerTimezone()
	}
	defaults["userTimezone"] = timezone
	setBrowserConfig(conf, browserConfig{
		Enabled:        config.BrowserEnabled,
		ExecutablePath: defaultBrowserExecutablePath,
		Headless:       true,
		NoSandbox:      true,
		DefaultProfile: defaultBrowserProfile,
	})
}

func (a AgentService) syncAgentsByAccount(account *model.AgentAccount) error {
	agents, err := agentRepo.List(repo.WithByAccountID(account.ID))
	if err != nil {
		return err
	}
	accountModels, err := loadAgentAccountModels(account)
	if err != nil {
		return err
	}
	if len(accountModels) == 0 {
		return nil
	}
	baseURL := resolveAccountBaseURL(account)
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
		modelName := strings.TrimSpace(agent.Model)
		var selectedAccountModel dto.AgentAccountModel
		var ok bool
		if modelName != "" {
			selectedAccountModel, ok = findAgentAccountModelForProvider(account.Provider, accountModels, modelName)
			if !ok {
				return buserr.WithName("ErrAgentModelInUse", agent.Name)
			}
		} else {
			modelName = strings.TrimSpace(account.Model)
			if modelName != "" {
				selectedAccountModel, ok = findAgentAccountModelForProvider(account.Provider, accountModels, modelName)
			}
			if !ok {
				selectedAccountModel = accountModels[0]
			}
		}
		modelName = strings.TrimSpace(selectedAccountModel.ID)
		apiType, maxTokens, contextWindow := resolveRuntimeParams(
			account.Provider,
			account.APIType,
			selectedAccountModel.MaxTokens,
			selectedAccountModel.ContextWindow,
		)
		if err := writeOpenclawConfig(confDir, account, modelName, agent.Token, nil); err != nil {
			return err
		}
		agent.BaseURL = baseURL
		agent.APIKey = account.APIKey
		agent.Provider = account.Provider
		agent.Model = modelName
		agent.APIType = apiType
		agent.MaxTokens = maxTokens
		agent.ContextWindow = contextWindow
		_ = agentRepo.Save(&agent)
	}
	return nil
}

func buildAgentItem(agent *model.Agent, appInstall *model.AppInstall, envMap map[string]interface{}) dto.AgentItem {
	agentType := normalizeAgentType(agent.AgentType)
	if appInstall != nil && appInstall.ID > 0 && appInstall.App.Key == constant.AppCopaw {
		agentType = constant.AppCopaw
	}
	item := dto.AgentItem{
		ID:            agent.ID,
		Name:          agent.Name,
		AgentType:     agentType,
		Provider:      agent.Provider,
		ProviderName:  providerDisplayName(agent.Provider),
		Model:         agent.Model,
		APIType:       agent.APIType,
		MaxTokens:     agent.MaxTokens,
		ContextWindow: agent.ContextWindow,
		BaseURL:       agent.BaseURL,
		APIKey:        maskKey(agent.APIKey),
		Token:         agent.Token,
		Status:        agent.Status,
		Message:       agent.Message,
		AppInstallID:  agent.AppInstallID,
		AccountID:     agent.AccountID,
		ConfigPath:    agent.ConfigPath,
		CreatedAt:     agent.CreatedAt,
	}
	if appInstall != nil && appInstall.ID > 0 {
		item.Container = appInstall.ContainerName
		item.AppVersion = appInstall.Version
		if agentType == constant.AppOpenclaw {
			if isOpenclawHTTPSVersion(appInstall.Version) {
				item.WebUIPort = appInstall.HttpsPort
			} else {
				item.WebUIPort = appInstall.HttpPort
			}
		} else {
			item.WebUIPort = appInstall.HttpPort
		}
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

func isOpenclawHTTPSVersion(version string) bool {
	target := strings.TrimSpace(strings.ToLower(version))
	if target == "" || target == "latest" {
		return true
	}
	if !strings.ContainsAny(target, "0123456789") {
		return true
	}
	return common.CompareAppVersion(target, openclawHTTPSVersion)
}

func shouldMigrateOpenclawHTTPSUpgrade(install *model.AppInstall, fromVersion, toVersion string) bool {
	if install == nil || install.App.Key != constant.AppOpenclaw {
		return false
	}
	return !isOpenclawHTTPSVersion(fromVersion) && isOpenclawHTTPSVersion(toVersion)
}

func migrateOpenclawHTTPSUpgrade(install *model.AppInstall, fromVersion, toVersion string) error {
	systemIP, _ := settingRepo.GetValueByKey("SystemIP")
	return migrateOpenclawHTTPSUpgradeWithSystemIP(install, fromVersion, toVersion, systemIP)
}

func migrateOpenclawHTTPSUpgradeWithSystemIP(install *model.AppInstall, fromVersion, toVersion, systemIP string) error {
	if !shouldMigrateOpenclawHTTPSUpgrade(install, fromVersion, toVersion) {
		return nil
	}
	migrateOpenclawInstallPorts(install)
	if err := openclawutil.WriteCatchAllCaddyfile(install.GetPath()); err != nil {
		return err
	}
	configPath := path.Join(install.GetPath(), "data", "conf", "openclaw.json")
	var allowedOrigins []string
	if conf, err := readOpenclawConfig(configPath); err == nil {
		allowedOrigins = extractSecurityConfig(conf).AllowedOrigins
	}
	originHost := strings.TrimSpace(systemIP)
	if originHost == "" {
		originHost = openclawAllowedOriginHost
	}
	if install.HttpsPort > 0 {
		allowedOrigin, err := buildOpenclawAllowedOrigin(originHost, install.HttpsPort)
		if err == nil {
			conf, err := readOpenclawConfig(configPath)
			if err != nil {
				return err
			}
			allowedOrigins = []string{allowedOrigin}
			setSecurityConfig(conf, dto.AgentSecurityConfig{AllowedOrigins: allowedOrigins})
			if err := writeOpenclawConfigRaw(configPath, conf); err != nil {
				return err
			}
		}
	}
	return migrateOpenclawInstallEnv(install, allowedOrigins)
}

func migrateOpenclawInstallPorts(install *model.AppInstall) {
	if install == nil {
		return
	}
	if install.HttpsPort == 0 && install.HttpPort > 0 {
		install.HttpsPort = install.HttpPort
	}
	if install.HttpPort > 0 {
		install.HttpPort = 0
	}
}

func migrateOpenclawInstallEnv(install *model.AppInstall, allowedOrigins []string) error {
	if install == nil {
		return nil
	}
	envMap := make(map[string]interface{})
	if strings.TrimSpace(install.Env) != "" {
		if err := json.Unmarshal([]byte(install.Env), &envMap); err != nil {
			return err
		}
	}
	if install.HttpsPort > 0 {
		envMap["PANEL_APP_PORT_HTTPS"] = install.HttpsPort
	}
	if allowedOrigin := firstAllowedOrigin(allowedOrigins); allowedOrigin != "" {
		envMap["ALLOWED_ORIGIN"] = allowedOrigin
	}
	delete(envMap, "PANEL_APP_PORT_HTTP")
	payload, err := json.Marshal(envMap)
	if err != nil {
		return err
	}
	install.Env = string(payload)
	return nil
}

func syncOpenclawAllowedOriginEnv(install *model.AppInstall, allowedOrigins []string) error {
	if install == nil {
		return nil
	}
	envMap := make(map[string]interface{})
	if strings.TrimSpace(install.Env) != "" {
		if err := json.Unmarshal([]byte(install.Env), &envMap); err != nil {
			return err
		}
	}
	if allowedOrigin := firstAllowedOrigin(allowedOrigins); allowedOrigin != "" {
		envMap["ALLOWED_ORIGIN"] = allowedOrigin
	} else {
		delete(envMap, "ALLOWED_ORIGIN")
	}
	payload, err := json.Marshal(envMap)
	if err != nil {
		return err
	}
	install.Env = string(payload)
	return nil
}

func firstAllowedOrigin(allowedOrigins []string) string {
	for _, origin := range allowedOrigins {
		trimmed := strings.TrimSpace(origin)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}

func buildOpenclawAllowedOrigin(host string, port int) (string, error) {
	host = strings.TrimSpace(host)
	if host == "" || port <= 0 {
		return "", fmt.Errorf("invalid openclaw allowed origin")
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") && strings.Count(host, ":") > 1 {
		host = "[" + host + "]"
	}
	return normalizeAllowedOrigin(fmt.Sprintf("https://%s:%d", host, port))
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

func (a AgentService) writeConfigWithRetry(appInstall *model.AppInstall, accountID uint, modelName, token string, agentID uint, allowedOrigins []string) {
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
	account, err := agentAccountRepo.GetFirst(repo.WithByID(accountID))
	if err != nil {
		global.LOG.Errorf("load agent account failed: %v", err)
		return
	}
	if err := writeOpenclawConfig(confDir, account, modelName, token, allowedOrigins); err != nil {
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
	Browser browserConfig `json:"browser"`
	Tools   toolsConfig   `json:"tools"`
	Update  updateConfig  `json:"update"`
	Models  *modelsConfig `json:"models,omitempty"`
}

type toolsConfig struct {
	Profile  string             `json:"profile,omitempty"`
	Sessions toolSessionsConfig `json:"sessions,omitempty"`
}

type toolSessionsConfig struct {
	Visibility string `json:"visibility,omitempty"`
}

type updateConfig struct {
	CheckOnStart bool `json:"checkOnStart"`
}

type gatewayConfig struct {
	Mode           string           `json:"mode"`
	Bind           string           `json:"bind"`
	Port           int              `json:"port"`
	Auth           gatewayAuth      `json:"auth"`
	ControlUi      gatewayControlUi `json:"controlUi"`
	TrustedProxies []string         `json:"trustedProxies,omitempty"`
}

type gatewayControlUi struct {
	DangerouslyDisableDeviceAuth bool     `json:"dangerouslyDisableDeviceAuth"`
	AllowedOrigins               []string `json:"allowedOrigins,omitempty"`
}

type gatewayAuth struct {
	Mode  string `json:"mode"`
	Token string `json:"token"`
}

type agentsConfig struct {
	Defaults agentDefaults `json:"defaults"`
}

type agentDefaults struct {
	UserTimezone string                            `json:"userTimezone,omitempty"`
	Model        modelRef                          `json:"model"`
	Models       map[string]map[string]interface{} `json:"models,omitempty"`
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

type browserConfig struct {
	Enabled        bool   `json:"enabled"`
	ExecutablePath string `json:"executablePath"`
	Headless       bool   `json:"headless"`
	NoSandbox      bool   `json:"noSandbox"`
	DefaultProfile string `json:"defaultProfile"`
}

func writeOpenclawConfig(confDir string, account *model.AgentAccount, modelName, token string, allowedOrigins []string) error {
	if strings.TrimSpace(confDir) == "" {
		return fmt.Errorf("config dir is required")
	}
	if account == nil {
		return fmt.Errorf("account is required")
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
	primaryModel, defaultsModels, models, err := buildOpenclawModelsFromAccount(account, modelName)
	if err != nil {
		return err
	}

	cfg := openclawConfig{
		Gateway: gatewayConfig{
			Mode: "local",
			Bind: "loopback",
			Port: openclawGatewayPort,
			Auth: gatewayAuth{
				Mode:  "token",
				Token: token,
			},
			ControlUi: gatewayControlUi{
				DangerouslyDisableDeviceAuth: true,
				AllowedOrigins:               append([]string(nil), allowedOrigins...),
			},
			TrustedProxies: []string{openclawTrustedProxyLoopback},
		},
		Agents: agentsConfig{
			Defaults: agentDefaults{
				UserTimezone: resolveServerTimezone(),
				Model:        modelRef{Primary: primaryModel},
				Models:       defaultsModels,
			},
		},
		Browser: browserConfig{
			Enabled:        true,
			ExecutablePath: defaultBrowserExecutablePath,
			Headless:       true,
			NoSandbox:      true,
			DefaultProfile: defaultBrowserProfile,
		},
		Tools: toolsConfig{
			Profile: defaultToolsProfile,
			Sessions: toolSessionsConfig{
				Visibility: defaultToolsSessionVisibility,
			},
		},
		Update: updateConfig{
			CheckOnStart: false,
		},
		Models: models,
	}

	configPath := path.Join(confDir, "openclaw.json")
	conf := map[string]interface{}{}
	if fileOp.Stat(configPath) {
		existing, err := readOpenclawConfig(configPath)
		if err != nil {
			return err
		}
		conf = existing
	}
	if len(conf) == 0 {
		initial, err := structToMap(cfg)
		if err != nil {
			return err
		}
		conf = initial
	} else {
		if cfg.Models != nil {
			modelsMap, err := structToMap(cfg.Models)
			if err != nil {
				return err
			}
			conf["models"] = modelsMap
		}
		if _, ok := conf["browser"]; !ok {
			browserMap, err := structToMap(cfg.Browser)
			if err != nil {
				return err
			}
			conf["browser"] = browserMap
		}
		toolsMap := ensureChildMap(conf, "tools")
		if profile, ok := toolsMap["profile"]; !ok || strings.TrimSpace(fmt.Sprintf("%v", profile)) == "" {
			toolsMap["profile"] = defaultToolsProfile
		}
		sessionsMap := ensureChildMap(toolsMap, "sessions")
		if visibility, ok := sessionsMap["visibility"]; !ok || strings.TrimSpace(fmt.Sprintf("%v", visibility)) == "" {
			sessionsMap["visibility"] = defaultToolsSessionVisibility
		}
		agentsMap := ensureChildMap(conf, "agents")
		defaultsMap := ensureChildMap(agentsMap, "defaults")
		if tz, ok := defaultsMap["userTimezone"]; !ok || strings.TrimSpace(fmt.Sprintf("%v", tz)) == "" {
			defaultsMap["userTimezone"] = resolveServerTimezone()
		}
		modelMap := ensureChildMap(defaultsMap, "model")
		modelMap["primary"] = cfg.Agents.Defaults.Model.Primary
		if cfg.Agents.Defaults.Models != nil {
			defaultsMap["models"] = cfg.Agents.Defaults.Models
		}

		ensureGatewaySecurityDefaults(conf)
		gatewayMap := ensureChildMap(conf, "gateway")
		if _, ok := gatewayMap["mode"]; !ok {
			gatewayMap["mode"] = "local"
		}
		if _, ok := gatewayMap["bind"]; !ok {
			gatewayMap["bind"] = "loopback"
		}
		if _, ok := gatewayMap["port"]; !ok {
			gatewayMap["port"] = openclawGatewayPort
		}
		authMap := ensureChildMap(gatewayMap, "auth")
		if _, ok := authMap["mode"]; !ok {
			authMap["mode"] = "token"
		}
		authMap["token"] = token
	}
	if allowedOrigins != nil {
		setSecurityConfig(conf, dto.AgentSecurityConfig{AllowedOrigins: allowedOrigins})
	}
	if err := writeOpenclawConfigRaw(configPath, conf); err != nil {
		return err
	}
	envPath := path.Join(confDir, ".env")
	lines := []string{fmt.Sprintf("OPENCLAW_GATEWAY_TOKEN=%s", token)}
	if envKey := providerEnvKey(account.Provider); envKey != "" && strings.TrimSpace(account.APIKey) != "" {
		lines = append(lines, fmt.Sprintf("%s=%s", envKey, account.APIKey))
	}
	content := strings.Join(lines, "\n") + "\n"
	return fileOp.SaveFile(envPath, content, 0600)
}

func buildOpenclawModelsFromAccount(account *model.AgentAccount, selectedModel string) (string, map[string]map[string]interface{}, *modelsConfig, error) {
	accountModels, err := loadAgentAccountModels(account)
	if err != nil {
		return "", nil, nil, err
	}
	if len(accountModels) == 0 {
		return "", nil, nil, fmt.Errorf("model is required")
	}
	selectedModel = strings.TrimSpace(selectedModel)
	if selectedModel == "" {
		selectedModel = strings.TrimSpace(account.Model)
	}
	if selectedModel == "" {
		selectedModel = strings.TrimSpace(accountModels[0].ID)
	}
	if selectedModel == "" {
		return "", nil, nil, fmt.Errorf("model is required")
	}
	selectedAccountModel, ok := findAgentAccountModelForProvider(account.Provider, accountModels, selectedModel)
	if !ok {
		return "", nil, nil, buserr.New("ErrAgentModelNotInAccount")
	}
	selectedModel = strings.TrimSpace(selectedAccountModel.ID)

	providerKey := ""
	providerCfg := modelProvider{}
	entries := make([]modelEntry, 0, len(accountModels))
	primaryModel := ""
	defaultsModels := make(map[string]map[string]interface{}, len(accountModels))
	for _, item := range accountModels {
		resolvedPrimary, entry, key, baseCfg, err := buildOpenclawCatalogModel(account, item)
		if err != nil {
			return "", nil, nil, err
		}
		if providerKey == "" {
			providerKey = key
			providerCfg.ApiKey = baseCfg.ApiKey
			providerCfg.BaseUrl = baseCfg.BaseUrl
			providerCfg.Api = baseCfg.Api
		}
		entries = append(entries, entry)
		defaultsModels[resolvedPrimary] = map[string]interface{}{}
		if sameProviderModelID(account.Provider, item.ID, selectedModel) {
			primaryModel = resolvedPrimary
		}
	}
	if primaryModel == "" {
		return "", nil, nil, buserr.New("ErrAgentModelNotInAccount")
	}
	providerCfg.Models = entries
	return primaryModel, defaultsModels, &modelsConfig{
		Mode: "merge",
		Providers: map[string]modelProvider{
			providerKey: providerCfg,
		},
	}, nil
}

func buildOpenclawCatalogModel(account *model.AgentAccount, model dto.AgentAccountModel) (string, modelEntry, string, modelProvider, error) {
	primaryModel, inferredEntry, providerKey, providerCfg, err := inferOpenclawCatalogModel(account, model.ID, model.MaxTokens, model.ContextWindow)
	if err != nil {
		return "", modelEntry{}, "", modelProvider{}, err
	}
	if strings.TrimSpace(model.Name) != "" {
		inferredEntry.Name = strings.TrimSpace(model.Name)
	}
	if len(model.Input) > 0 {
		inferredEntry.Input = sanitizeAgentAccountModelInputs(model.Input)
	}
	inferredEntry.Reasoning = model.Reasoning
	if model.ContextWindow > 0 {
		inferredEntry.ContextWindow = model.ContextWindow
	}
	if model.MaxTokens > 0 {
		inferredEntry.MaxTokens = model.MaxTokens
	}
	return primaryModel, inferredEntry, providerKey, providerCfg, nil
}

func buildOpenclawPrimaryModel(account *model.AgentAccount, modelID string) (string, error) {
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return "", err
	}
	item, ok := findAgentAccountModelForProvider(account.Provider, models, modelID)
	if !ok {
		return "", buserr.New("ErrAgentModelNotInAccount")
	}
	primaryModel, _, _, _, err := inferOpenclawCatalogModel(account, item.ID, item.MaxTokens, item.ContextWindow)
	if err != nil {
		return "", err
	}
	return primaryModel, nil
}

func inferOpenclawCatalogModel(account *model.AgentAccount, modelID string, maxTokens, contextWindow int) (string, modelEntry, string, modelProvider, error) {
	baseURL := resolveAccountBaseURL(account)
	resolvedAPIType, resolvedMaxTokens, resolvedContextWindow := resolveRuntimeParams(account.Provider, account.APIType, maxTokens, contextWindow)
	patch, err := providercatalog.BuildOpenClawPatch(account.Provider, modelID, resolvedAPIType, resolvedMaxTokens, resolvedContextWindow, baseURL, account.APIKey)
	if err != nil {
		return "", modelEntry{}, "", modelProvider{}, err
	}
	if patch.Models == nil {
		return "", modelEntry{}, "", modelProvider{}, fmt.Errorf("models patch is required")
	}
	modelsCfg, err := mapToModelsConfig(patch.Models)
	if err != nil {
		return "", modelEntry{}, "", modelProvider{}, err
	}
	for key, providerCfg := range modelsCfg.Providers {
		if len(providerCfg.Models) == 0 {
			continue
		}
		return patch.PrimaryModel, providerCfg.Models[0], key, modelProvider{
			ApiKey:  providerCfg.ApiKey,
			BaseUrl: providerCfg.BaseUrl,
			Api:     providerCfg.Api,
		}, nil
	}
	return "", modelEntry{}, "", modelProvider{}, fmt.Errorf("models patch is invalid")
}

func resolveAccountBaseURL(account *model.AgentAccount) string {
	baseURL := strings.TrimSpace(account.BaseURL)
	if baseURL == "" {
		if defaultURL, ok := providerDefaultBaseURL(account.Provider); ok {
			baseURL = defaultURL
		}
	}
	return baseURL
}

func buildInitialAgentAccountModels(account *model.AgentAccount, requested []dto.AgentAccountModel, legacyModel string) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	if len(requested) > 0 || strings.TrimSpace(legacyModel) != "" {
		models, _, err := normalizeAgentAccountModels(account, requested, legacyModel, true)
		if err != nil {
			return nil, err
		}
		return models, nil
	}
	meta, ok := providercatalog.Get(account.Provider)
	if !ok || len(meta.Models) == 0 {
		return nil, nil
	}
	requested = make([]dto.AgentAccountModel, 0, len(meta.Models))
	for _, item := range meta.Models {
		requested = append(requested, dto.AgentAccountModel{
			ID:            item.ID,
			Name:          item.Name,
			ContextWindow: item.ContextWindow,
			MaxTokens:     item.MaxTokens,
			Reasoning:     item.Reasoning,
			Input:         append([]string(nil), item.Input...),
		})
	}
	models, _, err := normalizeAgentAccountModels(account, requested, "", true)
	if err != nil {
		return nil, err
	}
	return models, nil
}

func compactPersistedAgentAccountModelSortOrder(accountID uint) error {
	rows, err := agentAccountModelRepo.List(repo.WithByAccountID(accountID), repo.WithOrderAsc("sort_order"), repo.WithOrderAsc("id"))
	if err != nil {
		return err
	}
	for index := range rows {
		order := index + 1
		if rows[index].SortOrder == order {
			continue
		}
		rows[index].SortOrder = order
		if err := agentAccountModelRepo.Save(&rows[index]); err != nil {
			return err
		}
	}
	return nil
}

func loadAgentAccountModels(account *model.AgentAccount) ([]dto.AgentAccountModel, error) {
	models, err := ensurePersistedAgentAccountModels(account)
	if err != nil {
		return nil, err
	}
	return models, nil
}

func LoadLegacyAgentAccountModelsForMigration(account *model.AgentAccount) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	if !hasLegacyAgentAccountModels(account) {
		return nil, nil
	}
	models, _, err := loadLegacyAgentAccountModelCatalog(account)
	if err != nil {
		if strings.TrimSpace(err.Error()) == "model is required" {
			return nil, nil
		}
		return nil, err
	}
	return models, nil
}

func MergeCatalogAgentAccountModelsForMigration(account *model.AgentAccount, existing []dto.AgentAccountModel) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	meta, ok := providercatalog.Get(account.Provider)
	if !ok || len(meta.Models) == 0 {
		return append([]dto.AgentAccountModel(nil), existing...), nil
	}
	requested := append([]dto.AgentAccountModel(nil), existing...)
	seen := make(map[string]struct{}, len(existing))
	for _, item := range existing {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		seen[strings.TrimSpace(item.ID)] = struct{}{}
	}
	for _, item := range meta.Models {
		if _, ok := seen[strings.TrimSpace(item.ID)]; ok {
			continue
		}
		requested = append(requested, dto.AgentAccountModel{
			ID:            item.ID,
			Name:          item.Name,
			ContextWindow: item.ContextWindow,
			MaxTokens:     item.MaxTokens,
			Reasoning:     item.Reasoning,
			Input:         append([]string(nil), item.Input...),
		})
	}
	if len(requested) == len(existing) {
		return append([]dto.AgentAccountModel(nil), existing...), nil
	}
	normalized, _, err := normalizeAgentAccountModels(account, requested, "", true)
	if err != nil {
		return nil, err
	}
	return normalized, nil
}

func hasLegacyAgentAccountModels(account *model.AgentAccount) bool {
	if account == nil {
		return false
	}
	if strings.TrimSpace(account.Models) != "" || strings.TrimSpace(account.Model) != "" {
		return true
	}
	if account.ID > 0 {
		if agents, err := agentRepo.List(repo.WithByAccountID(account.ID)); err == nil {
			for _, agent := range agents {
				if strings.TrimSpace(agent.Model) != "" {
					return true
				}
			}
		}
	}
	if definitions, ok := providerDefinitions()[strings.ToLower(strings.TrimSpace(account.Provider))]; ok {
		return len(definitions.Models) > 0
	}
	return false
}

func ensurePersistedAgentAccountModels(account *model.AgentAccount) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	models, err := listPersistedAgentAccountModels(account.ID)
	if err != nil {
		return nil, err
	}
	if len(models) > 0 {
		return models, nil
	}
	if !hasLegacyAgentAccountModels(account) {
		return nil, nil
	}
	legacyModels, _, err := loadLegacyAgentAccountModelCatalog(account)
	if err != nil {
		if strings.TrimSpace(err.Error()) == "model is required" {
			return nil, nil
		}
		return nil, err
	}
	if len(legacyModels) == 0 {
		return nil, nil
	}
	if account.ID == 0 {
		return legacyModels, nil
	}
	if err := replacePersistedAgentAccountModels(account.ID, legacyModels); err != nil {
		return nil, err
	}
	return listPersistedAgentAccountModels(account.ID)
}

func listPersistedAgentAccountModels(accountID uint) ([]dto.AgentAccountModel, error) {
	if accountID == 0 {
		return nil, nil
	}
	rows, err := agentAccountModelRepo.List(repo.WithByAccountID(accountID), repo.WithOrderAsc("sort_order"), repo.WithOrderAsc("id"))
	if err != nil {
		return nil, err
	}
	result := make([]dto.AgentAccountModel, 0, len(rows))
	for _, row := range rows {
		inputs := []string{}
		if strings.TrimSpace(row.Input) != "" {
			_ = json.Unmarshal([]byte(row.Input), &inputs)
		}
		result = append(result, dto.AgentAccountModel{
			RecordID:      row.ID,
			ID:            strings.TrimSpace(row.Model),
			Name:          strings.TrimSpace(row.Name),
			ContextWindow: row.ContextWindow,
			MaxTokens:     row.MaxTokens,
			Reasoning:     row.Reasoning,
			Input:         sanitizeAgentAccountModelInputs(inputs),
		})
	}
	return result, nil
}

func replacePersistedAgentAccountModels(accountID uint, models []dto.AgentAccountModel) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("account_id = ?", accountID).Delete(&model.AgentAccountModel{}).Error; err != nil {
			return err
		}
		for index, item := range models {
			inputPayload, err := json.Marshal(sanitizeAgentAccountModelInputs(item.Input))
			if err != nil {
				return err
			}
			record := &model.AgentAccountModel{
				AccountID:     accountID,
				Model:         strings.TrimSpace(item.ID),
				Name:          strings.TrimSpace(item.Name),
				ContextWindow: item.ContextWindow,
				MaxTokens:     item.MaxTokens,
				Reasoning:     item.Reasoning,
				Input:         string(inputPayload),
				SortOrder:     index + 1,
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func loadLegacyAgentAccountModelCatalog(account *model.AgentAccount) ([]dto.AgentAccountModel, string, error) {
	if account == nil {
		return nil, "", fmt.Errorf("account is required")
	}
	models, err := parseAgentAccountModels(account.Models)
	if err != nil {
		return nil, "", err
	}
	normalized, _, err := normalizeAgentAccountModels(account, models, account.Model, true)
	if err != nil {
		return nil, "", err
	}
	_, defaultModel, err := normalizeAgentAccountModels(account, normalized, account.Model, true)
	if err != nil {
		return nil, "", err
	}
	return normalized, defaultModel, nil
}

func normalizeAgentAccountModels(account *model.AgentAccount, models []dto.AgentAccountModel, defaultModel string, allowFallbackDefault bool) ([]dto.AgentAccountModel, string, error) {
	requested := append([]dto.AgentAccountModel(nil), models...)
	if len(requested) == 0 {
		if strings.TrimSpace(defaultModel) != "" {
			requested = []dto.AgentAccountModel{{ID: defaultModel}}
		} else {
			requested = buildLegacyAgentAccountModels(account)
		}
	}
	normalized := make([]dto.AgentAccountModel, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, item := range requested {
		normalizedItem, err := normalizeAgentAccountModel(account, item)
		if err != nil {
			return nil, "", err
		}
		if strings.TrimSpace(normalizedItem.ID) == "" {
			continue
		}
		if _, ok := seen[normalizedItem.ID]; ok {
			continue
		}
		seen[normalizedItem.ID] = struct{}{}
		normalized = append(normalized, normalizedItem)
	}
	if len(normalized) == 0 {
		return nil, "", fmt.Errorf("model is required")
	}

	resolvedDefault := strings.TrimSpace(defaultModel)
	if resolvedDefault != "" {
		defaultItem, err := normalizeAgentAccountModel(account, dto.AgentAccountModel{ID: resolvedDefault})
		if err == nil {
			resolvedDefault = defaultItem.ID
		}
	}
	if resolvedDefault == "" && allowFallbackDefault {
		resolvedDefault = normalized[0].ID
	}
	if _, ok := findAgentAccountModelForProvider(account.Provider, normalized, resolvedDefault); !ok {
		if allowFallbackDefault {
			resolvedDefault = normalized[0].ID
		} else {
			return nil, "", buserr.New("ErrAgentModelNotInAccount")
		}
	}
	return normalized, resolvedDefault, nil
}

func normalizeAgentAccountModel(account *model.AgentAccount, model dto.AgentAccountModel) (dto.AgentAccountModel, error) {
	modelID := strings.TrimSpace(model.ID)
	if modelID == "" {
		return dto.AgentAccountModel{}, fmt.Errorf("model is required")
	}
	primaryModel, inferredEntry, _, _, err := inferOpenclawCatalogModel(account, modelID, model.MaxTokens, model.ContextWindow)
	if err != nil {
		return dto.AgentAccountModel{}, err
	}
	name := strings.TrimSpace(model.Name)
	if name == "" {
		name = strings.TrimSpace(inferredEntry.Name)
	}
	reasoning := model.Reasoning
	if !model.Reasoning && model.Name == "" && model.MaxTokens == 0 && model.ContextWindow == 0 && len(model.Input) == 0 {
		reasoning = inferredEntry.Reasoning
	}
	inputs := sanitizeAgentAccountModelInputs(model.Input)
	if len(inputs) == 0 {
		inputs = sanitizeAgentAccountModelInputs(inferredEntry.Input)
	}
	contextWindow := model.ContextWindow
	if contextWindow <= 0 {
		contextWindow = inferredEntry.ContextWindow
	}
	maxTokens := model.MaxTokens
	if maxTokens <= 0 {
		maxTokens = inferredEntry.MaxTokens
	}
	return dto.AgentAccountModel{
		ID:            normalizeAgentAccountModelID(account.Provider, primaryModel, modelID),
		Name:          name,
		ContextWindow: contextWindow,
		MaxTokens:     maxTokens,
		Reasoning:     reasoning,
		Input:         inputs,
	}, nil
}

func normalizeAgentAccountModelID(provider, primaryModel, requestedID string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "custom", "vllm":
		target := requestedID
		if strings.TrimSpace(target) == "" {
			target = primaryModel
		}
		return normalizeCustomModel(target)
	case "ollama":
		target := strings.TrimSpace(primaryModel)
		if strings.HasPrefix(target, "ollama/") {
			return target
		}
		target = strings.TrimSpace(requestedID)
		if strings.HasPrefix(target, "ollama/") {
			return target
		}
		target = strings.TrimLeft(strings.TrimSpace(target), "/")
		if target == "" {
			target = strings.TrimLeft(strings.TrimSpace(primaryModel), "/")
		}
		if target == "" {
			return ""
		}
		return "ollama/" + target
	default:
		target := strings.TrimSpace(requestedID)
		if target == "" {
			target = strings.TrimSpace(primaryModel)
		}
		if target == "" {
			return ""
		}
		prefix := poolModelPrefix(provider)
		if strings.Contains(target, "/") {
			parts := strings.SplitN(target, "/", 2)
			targetPrefix := strings.ToLower(strings.TrimSpace(parts[0]))
			targetModel := strings.TrimSpace(parts[1])
			if targetModel == "" {
				return strings.TrimSpace(target)
			}
			for _, item := range supportedProviderModelPrefixes(provider) {
				if item == targetPrefix {
					if prefix != "" {
						return prefix + "/" + targetModel
					}
					return strings.TrimSpace(target)
				}
			}
			return strings.TrimSpace(target)
		}
		target = strings.TrimLeft(strings.TrimSpace(target), "/")
		if prefix == "" {
			return target
		}
		return prefix + "/" + target
	}
}

func buildLegacyAgentAccountModels(account *model.AgentAccount) []dto.AgentAccountModel {
	modelIDs := make([]string, 0, 4)
	seen := make(map[string]struct{}, 4)
	appendModel := func(value string) {
		target := strings.TrimSpace(value)
		if target == "" {
			return
		}
		if _, ok := seen[target]; ok {
			return
		}
		seen[target] = struct{}{}
		modelIDs = append(modelIDs, target)
	}
	appendModel(account.Model)
	if account.ID > 0 {
		if agents, err := agentRepo.List(repo.WithByAccountID(account.ID)); err == nil {
			for _, agent := range agents {
				appendModel(agent.Model)
			}
		}
	}
	if definitions, ok := providerDefinitions()[strings.ToLower(strings.TrimSpace(account.Provider))]; ok && len(definitions.Models) > 0 {
		for _, item := range definitions.Models {
			appendModel(item.ID)
		}
	}
	models := make([]dto.AgentAccountModel, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, dto.AgentAccountModel{ID: modelID})
	}
	return models
}

func parseAgentAccountModels(value string) ([]dto.AgentAccountModel, error) {
	trim := strings.TrimSpace(value)
	if trim == "" {
		return nil, nil
	}
	var models []dto.AgentAccountModel
	if err := json.Unmarshal([]byte(trim), &models); err != nil {
		return nil, err
	}
	return models, nil
}

func marshalAgentAccountModels(models []dto.AgentAccountModel) (string, error) {
	payload, err := json.Marshal(models)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func sanitizeAgentAccountModelInputs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := strings.ToLower(strings.TrimSpace(value))
		if normalized != "text" && normalized != "image" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	if len(result) == 0 {
		return []string{"text"}
	}
	return result
}

func shouldRefreshAccountModelRuntimeLimits(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "custom", "vllm", "ollama":
		return true
	default:
		return false
	}
}

func refreshAccountModelRuntimeLimits(account *model.AgentAccount, models []dto.AgentAccountModel) []dto.AgentAccountModel {
	refreshed := make([]dto.AgentAccountModel, 0, len(models))
	for _, item := range models {
		next := item
		next.MaxTokens = account.MaxTokens
		next.ContextWindow = account.ContextWindow
		refreshed = append(refreshed, next)
	}
	return refreshed
}

func normalizeComparableProviderModelID(provider, modelID string) string {
	target := strings.TrimSpace(modelID)
	if target == "" {
		return ""
	}
	if !strings.Contains(target, "/") {
		return target
	}
	parts := strings.SplitN(target, "/", 2)
	prefix := strings.ToLower(strings.TrimSpace(parts[0]))
	model := strings.TrimSpace(parts[1])
	if model == "" {
		return target
	}
	for _, item := range supportedProviderModelPrefixes(provider) {
		if item == prefix {
			return model
		}
	}
	return target
}

func sameProviderModelID(provider, left, right string) bool {
	leftTrimmed := strings.TrimSpace(left)
	rightTrimmed := strings.TrimSpace(right)
	if leftTrimmed == rightTrimmed {
		return true
	}
	leftComparable := normalizeComparableProviderModelID(provider, leftTrimmed)
	rightComparable := normalizeComparableProviderModelID(provider, rightTrimmed)
	return leftComparable != "" && leftComparable == rightComparable
}

func findAgentAccountModelForProvider(provider string, models []dto.AgentAccountModel, modelID string) (dto.AgentAccountModel, bool) {
	for _, item := range models {
		if sameProviderModelID(provider, item.ID, modelID) {
			return item, true
		}
	}
	return dto.AgentAccountModel{}, false
}

func ensureAccountModelsNotBound(account *model.AgentAccount, models []dto.AgentAccountModel) error {
	if account == nil || account.ID == 0 {
		return nil
	}
	agents, err := agentRepo.List(repo.WithByAccountID(account.ID))
	if err != nil {
		return err
	}
	for _, agent := range agents {
		if strings.TrimSpace(agent.Model) == "" {
			continue
		}
		if _, ok := findAgentAccountModelForProvider(account.Provider, models, agent.Model); !ok {
			return buserr.WithName("ErrAgentModelInUse", agent.Name)
		}
	}
	return nil
}

func resolveServerTimezone() string {
	timezone := strings.TrimSpace(common.LoadTimeZoneByCmd())
	if timezone == "" {
		return defaultUserTimezone
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return defaultUserTimezone
	}
	return timezone
}

func ensureChildMap(parent map[string]interface{}, key string) map[string]interface{} {
	if child, ok := parent[key].(map[string]interface{}); ok {
		return child
	}
	child := map[string]interface{}{}
	parent[key] = child
	return child
}

func structToMap(value interface{}) (map[string]interface{}, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	result := map[string]interface{}{}
	if err := json.Unmarshal(payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func mapToModelsConfig(value map[string]interface{}) (*modelsConfig, error) {
	payload, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	result := &modelsConfig{}
	if err := json.Unmarshal(payload, result); err != nil {
		return nil, err
	}
	return result, nil
}

func providerEnvKey(provider string) string {
	return providercatalog.EnvKey(provider)
}

type providerDefinition struct {
	Sort        uint
	DisplayName string
	BaseURL     string
	Models      []dto.ProviderModelInfo
}

func providerDefinitions() map[string]providerDefinition {
	definitions := map[string]providerDefinition{}
	for key, meta := range providercatalog.All() {
		if !meta.Enabled {
			continue
		}
		models := make([]dto.ProviderModelInfo, 0, len(meta.Models))
		for _, m := range meta.Models {
			models = append(models, dto.ProviderModelInfo{
				ID:            m.ID,
				Name:          m.Name,
				ContextWindow: m.ContextWindow,
				MaxTokens:     m.MaxTokens,
				Reasoning:     m.Reasoning,
				Input:         append([]string(nil), m.Input...),
			})
		}
		definitions[key] = providerDefinition{
			Sort:        meta.Sort,
			DisplayName: meta.DisplayName,
			BaseURL:     meta.DefaultBaseURL,
			Models:      models,
		}
	}
	return definitions
}

func providerDefaultBaseURL(provider string) (string, bool) {
	return providercatalog.DefaultBaseURL(provider)
}

func fixedProviderBaseURL(provider string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "bailian-coding-plan":
		return providerDefaultBaseURL(provider)
	case "ark-coding-plan":
		return providerDefaultBaseURL(provider)
	default:
		return "", false
	}
}

func isSupportedAgentProvider(provider string) bool {
	return providercatalog.IsEnabled(provider)
}

func providerDisplayName(provider string) string {
	return providercatalog.DisplayName(provider)
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

func normalizeCustomModel(modelName string) string {
	trim := strings.TrimSpace(modelName)
	trim = strings.TrimLeft(trim, "/")
	if parts := strings.SplitN(trim, "/", 2); len(parts) == 2 {
		if strings.EqualFold(parts[0], "custom") {
			return strings.TrimLeft(strings.TrimSpace(parts[1]), "/")
		}
	}
	return trim
}

func normalizeBailianCodingPlanModelID(modelID string) string {
	trim := strings.TrimSpace(modelID)
	switch strings.ToLower(trim) {
	case "minimax-m2.5", "minimax m2.5", "minimax/minimax-m2.5", "minimax/minimax m2.5":
		return "MiniMax/MiniMax-M2.5"
	default:
		return trim
	}
}

func normalizeArkCodingPlanModelID(modelID string) string {
	return strings.ToLower(strings.TrimSpace(modelID))
}

func zaiModelDisplayName(modelID string) string {
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "glm-5":
		return "GLM-5"
	case "glm-4.7":
		return "GLM-4.7"
	case "glm-4.7-flash":
		return "GLM-4.7-Flash"
	case "glm-4.7-flashx":
		return "GLM-4.7-FlashX"
	default:
		return strings.TrimSpace(modelID)
	}
}

func bailianPrimaryModelID(modelID string) string {
	trim := strings.TrimSpace(modelID)
	if trim == "" {
		return ""
	}
	parts := strings.Split(trim, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := strings.TrimSpace(parts[i])
		if part != "" {
			return part
		}
	}
	return trim
}

func normalizeAgentType(agentType string) string {
	trim := strings.ToLower(strings.TrimSpace(agentType))
	if trim == "" {
		return constant.AppOpenclaw
	}
	return trim
}

func modelMatchesProvider(provider, modelName string) bool {
	target := strings.TrimSpace(modelName)
	for _, prefix := range supportedProviderModelPrefixes(provider) {
		if prefix != "" && strings.HasPrefix(target, prefix+"/") {
			return true
		}
	}
	return false
}

func runtimeProviderModelPrefix(provider string) string {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "gemini":
		return "google"
	case "minimax":
		return "minimax-portal"
	case "kimi":
		return "moonshot"
	default:
		return strings.ToLower(strings.TrimSpace(provider))
	}
}

func poolModelPrefix(provider string) string {
	target := strings.ToLower(strings.TrimSpace(provider))
	if definitions, ok := providerDefinitions()[target]; ok && len(definitions.Models) > 0 {
		parts := strings.SplitN(strings.TrimSpace(definitions.Models[0].ID), "/", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			return strings.ToLower(strings.TrimSpace(parts[0]))
		}
	}
	return target
}

func supportedProviderModelPrefixes(provider string) []string {
	values := []string{poolModelPrefix(provider), runtimeProviderModelPrefix(provider)}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		target := strings.ToLower(strings.TrimSpace(value))
		if target == "" {
			continue
		}
		if _, ok := seen[target]; ok {
			continue
		}
		seen[target] = struct{}{}
		result = append(result, target)
	}
	return result
}

func isSupportedAgentType(agentType string) bool {
	switch normalizeAgentType(agentType) {
	case constant.AppOpenclaw, constant.AppCopaw:
		return true
	default:
		return false
	}
}

func normalizeAPIType(apiType string) string {
	trim := strings.ToLower(strings.TrimSpace(apiType))
	if trim == "" {
		return "openai-completions"
	}
	return trim
}

func isSupportedAPIType(apiType string) bool {
	switch normalizeAPIType(apiType) {
	case "openai-completions", "openai-responses", "anthropic-messages":
		return true
	default:
		return false
	}
}

func isSupportedOllamaAPIType(apiType string) bool {
	switch normalizeAPIType(apiType) {
	case "openai-completions", "openai-responses":
		return true
	default:
		return false
	}
}

func resolveRuntimeParams(provider, apiType string, maxTokens, contextWindow int) (string, int, int) {
	resolvedAPI := normalizeAPIType(apiType)
	if provider == "ollama" && !isSupportedOllamaAPIType(resolvedAPI) {
		resolvedAPI = "openai-responses"
	}
	resolvedMaxTokens := maxTokens
	resolvedContextWindow := contextWindow
	if resolvedMaxTokens <= 0 {
		switch provider {
		case "deepseek":
			resolvedMaxTokens = 8192
		case "zai":
			resolvedMaxTokens = 131072
		case "openrouter":
			resolvedMaxTokens = 8192
		case "minimax", "kimi-coding", "custom":
			resolvedMaxTokens = 8192
		default:
			resolvedMaxTokens = 8192
		}
	}
	if resolvedContextWindow <= 0 {
		switch provider {
		case "deepseek":
			resolvedContextWindow = 128000
		case "zai":
			resolvedContextWindow = 204800
		case "openrouter":
			resolvedContextWindow = 128000
		case "minimax", "kimi-coding":
			resolvedContextWindow = 200000
		case "custom", "vllm":
			resolvedContextWindow = 128000
		default:
			resolvedContextWindow = 256000
		}
	}
	return resolvedAPI, resolvedMaxTokens, resolvedContextWindow
}

func generateToken() string {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func asyncReportAIProviderInstall(provider string) {
	if global.CONF.Base.Mode != "stable" {
		return
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return
	}
	go func(provider string) {
		query := url.Values{}
		query.Set("product", "ai-provider")
		query.Set("type", "install")
		query.Set("version", provider)
		reqURL := "https://community.fit2cloud.com/installation-analytics?" + query.Encode()
		_, _, _ = req_helper.HandleRequest(reqURL, http.MethodGet, constant.TimeOut5s)
	}(provider)
}
