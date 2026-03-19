package service

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
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
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"gorm.io/gorm"
)

type IAgentService interface {
	Create(req dto.AgentCreateReq) (*dto.AgentItem, error)
	Page(req dto.SearchWithPage) (int64, []dto.AgentItem, error)
	Delete(req dto.AgentDeleteReq) error
	ResetToken(req dto.AgentTokenResetReq) error
	UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error
	GetProviders() ([]dto.ProviderInfo, error)
	CreateAccount(req dto.AgentAccountCreateReq) error
	UpdateAccount(req dto.AgentAccountUpdateReq) error
	SyncAgentsByAccount(account *model.AgentAccount) error
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
	GetDingTalkConfig(req dto.AgentDingTalkConfigReq) (*dto.AgentDingTalkConfig, error)
	UpdateDingTalkConfig(req dto.AgentDingTalkConfigUpdateReq) error
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
	agentType := req.AgentType
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
	app, err := appRepo.GetFirst(appRepo.WithKey(agentType))
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
	var account *model.AgentAccount
	var installHooks *appInstallHooks

	if agentType == constant.AppOpenclaw {
		var err error
		allowedOrigins, err = normalizeAllowedOrigins(req.AllowedOrigins)
		if err != nil {
			return nil, err
		}
		if len(allowedOrigins) == 0 {
			return nil, fmt.Errorf("allowed origins is required")
		}
		if req.AccountID == 0 {
			return nil, buserr.New("ErrAgentAccountRequired")
		}
		account, err = agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
		if err != nil {
			return nil, err
		}
		if !account.Verified {
			return nil, buserr.New("ErrAgentAccountNotVerified")
		}
		provider = account.Provider
		baseURL = strings.TrimSpace(account.BaseURL)
		resolvedRuntime, err := resolveOpenclawAccountModelRuntimeByID(account, req.Model)
		if err != nil {
			return nil, err
		}
		storedModel = resolvedRuntime.StoredModel
		apiType = resolvedRuntime.APIType
		maxTokens = resolvedRuntime.MaxTokens
		contextWindow = resolvedRuntime.ContextWindow
		runtimeModel = resolvedRuntime.PrimaryModel
		apiKey = account.APIKey
		accountID = account.ID
		token = strings.TrimSpace(req.Token)
		if token == "" {
			token = generateToken()
		}
		installHooks = &appInstallHooks{
			AfterCopyData: func(appInstall *model.AppInstall) error {
				return prepareOpenclawInstallFiles(appInstall, account, storedModel, token, allowedOrigins)
			},
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
	appInstall, err := AppService{}.installWithHooks(installReq, false, installHooks)
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
	return nil
}

func (a AgentService) ResetToken(req dto.AgentTokenResetReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppCopaw {
		return fmt.Errorf("copaw does not support token")
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
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
	if err := writeOpenclawConfigRaw(agent.ConfigPath, conf); err != nil {
		return err
	}
	agent.Token = newToken
	return agentRepo.Save(agent)
}

func (a AgentService) UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return err
	}
	if agent.AgentType == constant.AppCopaw {
		return fmt.Errorf("copaw does not support model config")
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.AccountID))
	if err != nil {
		return err
	}
	resolvedRuntime, err := resolveOpenclawAccountModelRuntimeByID(account, req.Model)
	if err != nil {
		return err
	}
	modelName := resolvedRuntime.StoredModel
	apiType, maxTokens, contextWindow := resolvedRuntime.APIType, resolvedRuntime.MaxTokens, resolvedRuntime.ContextWindow
	confDir := path.Dir(agent.ConfigPath)
	if err := writeOpenclawConfig(confDir, account, modelName, agent.Token, nil); err != nil {
		return err
	}
	agent.Provider = account.Provider
	agent.Model = modelName
	agent.APIType = apiType
	agent.MaxTokens = maxTokens
	agent.ContextWindow = contextWindow
	agent.BaseURL = account.BaseURL
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
	provider := req.Provider
	if exist, _ := agentAccountRepo.GetFirst(repo.WithByProvider(provider), repo.WithByName(req.Name)); exist != nil && exist.ID > 0 {
		return buserr.New("ErrRecordExist")
	}
	resolvedInput, err := resolveAgentAccountInput(provider, req.APIKey, req.BaseURL, req.APIType, "")
	if err != nil {
		return err
	}
	account := &model.AgentAccount{
		Provider:       resolvedInput.Provider,
		Name:           req.Name,
		APIKey:         resolvedInput.APIKey,
		RememberAPIKey: req.RememberAPIKey,
		BaseURL:        resolvedInput.BaseURL,
		APIType:        resolvedInput.APIType,
		Verified:       true,
		Remark:         req.Remark,
	}
	initialModels, err := buildInitialAgentAccountModels(account, req.Models)
	if err != nil {
		return err
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(account).Error; err != nil {
			return err
		}
		if len(initialModels) == 0 {
			return nil
		}
		return replacePersistedAgentAccountModelsWithTx(tx, account.ID, initialModels)
	}); err != nil {
		return err
	}
	asyncReportAIProviderInstall(provider)
	return nil
}

func (a AgentService) UpdateAccount(req dto.AgentAccountUpdateReq) error {
	account, err := agentAccountRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	provider := account.Provider
	resolvedInput, err := resolveAgentAccountInput(provider, req.APIKey, req.BaseURL, req.APIType, account.APIType)
	if err != nil {
		return err
	}
	account.Name = req.Name
	account.APIKey = resolvedInput.APIKey
	account.RememberAPIKey = req.RememberAPIKey
	account.BaseURL = resolvedInput.BaseURL
	account.APIType = resolvedInput.APIType
	account.Remark = req.Remark
	account.Verified = true

	if err := global.DB.Save(account).Error; err != nil {
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
			Models:         nil,
			APIType:        item.APIType,
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

func (a AgentService) SyncAgentsByAccount(account *model.AgentAccount) error {
	if account == nil || account.ID == 0 {
		return nil
	}
	return a.syncAgentsByAccount(account)
}

func (a AgentService) VerifyAccount(req dto.AgentAccountVerifyReq) error {
	resolvedVerification, err := resolveAgentAccountVerification(req.Provider, req.APIKey, req.BaseURL)
	if err != nil {
		return err
	}
	return verifyResolvedAgentAccount(resolvedVerification)
}

func (a AgentService) DeleteAccount(req dto.AgentAccountDeleteReq) error {
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
		setFeishuConfig(conf, dto.AgentFeishuConfig{
			Enabled:   req.Enabled,
			DmPolicy:  req.DmPolicy,
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
		setTelegramConfig(conf, dto.AgentTelegramConfig{
			Enabled:  req.Enabled,
			DmPolicy: req.DmPolicy,
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
		setDiscordConfig(conf, dto.AgentDiscordConfig{
			Enabled:     req.Enabled,
			DmPolicy:    req.DmPolicy,
			GroupPolicy: req.GroupPolicy,
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

func (a AgentService) GetDingTalkConfig(req dto.AgentDingTalkConfigReq) (*dto.AgentDingTalkConfig, error) {
	_, install, conf, err := a.loadAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	result := extractDingTalkConfig(conf)
	installed, _ := checkPluginInstalled(install.ContainerName, "dingtalk")
	result.Installed = installed
	return &result, nil
}

func (a AgentService) UpdateDingTalkConfig(req dto.AgentDingTalkConfigUpdateReq) error {
	return a.mutateAgentConfig(req.AgentID, func(_ *model.Agent, _ *model.AppInstall, conf map[string]interface{}) error {
		setDingTalkConfig(conf, dto.AgentDingTalkConfig{
			Enabled:        req.Enabled,
			ClientID:       req.ClientID,
			ClientSecret:   req.ClientSecret,
			DmPolicy:       req.DmPolicy,
			AllowFrom:      req.AllowFrom,
			GroupPolicy:    req.GroupPolicy,
			GroupAllowFrom: req.GroupAllowFrom,
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
	if agent.AgentType == constant.AppCopaw {
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
	if agent.AgentType == constant.AppCopaw {
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
	channelType := req.Type
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
		confDir := path.Dir(agent.ConfigPath)
		modelName := strings.TrimSpace(agent.Model)
		var selectedAccountModel dto.AgentAccountModel
		if modelName != "" {
			selectedAccountModel, err = requireAgentAccountModelForProvider(account.Provider, accountModels, modelName)
			if err != nil {
				return buserr.WithName("ErrAgentModelInUse", agent.Name)
			}
		} else {
			selectedAccountModel = accountModels[0]
		}
		resolvedRuntime, err := buildOpenclawAccountModelRuntime(account, selectedAccountModel)
		if err != nil {
			return err
		}
		modelName = resolvedRuntime.StoredModel
		apiType, maxTokens, contextWindow := resolvedRuntime.APIType, resolvedRuntime.MaxTokens, resolvedRuntime.ContextWindow
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
