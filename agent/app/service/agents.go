package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"gorm.io/gorm"
)

type IAgentService interface {
	Create(req dto.AgentCreateReq) (*dto.AgentItem, error)
	Page(req dto.SearchWithPage) (int64, []dto.AgentItem, error)
	Delete(req dto.AgentDeleteReq) error
	ResetToken(req dto.AgentTokenResetReq) error
	UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error
	GetOverview(req dto.AgentOverviewReq) (*dto.AgentOverview, error)
	GetProviders() ([]dto.ProviderInfo, error)
	GetSecurityConfig(req dto.AgentSecurityConfigReq) (*dto.AgentSecurityConfig, error)
	UpdateSecurityConfig(req dto.AgentSecurityConfigUpdateReq) error
	GetOtherConfig(req dto.AgentOtherConfigReq) (*dto.AgentOtherConfig, error)
	UpdateOtherConfig(req dto.AgentOtherConfigUpdateReq) error
	GetConfigFile(req dto.AgentConfigFileReq) (*dto.AgentConfigFile, error)
	UpdateConfigFile(req dto.AgentConfigFileUpdateReq) error
	ListSkills(req dto.AgentSkillsReq) ([]dto.AgentSkillItem, error)
	SearchSkills(req dto.AgentSkillSearchReq) ([]dto.AgentSkillSearchItem, error)
	UpdateSkill(req dto.AgentSkillUpdateReq) error
	InstallSkill(req dto.AgentSkillInstallReq) error

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
	LoginWeixinChannel(req dto.AgentWeixinLoginReq) error
	GetQQBotConfig(req dto.AgentQQBotConfigReq) (*dto.AgentQQBotConfig, error)
	UpdateQQBotConfig(req dto.AgentQQBotConfigUpdateReq) error
	InstallPlugin(req dto.AgentPluginInstallReq) error
	CheckPlugin(req dto.AgentPluginCheckReq) (*dto.AgentPluginStatus, error)
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
	openclawPluginPackageTmpDir   = "/tmp/openclaw-plugin"
	openclawManagedSkillsDir      = "/home/node/.openclaw/skills"
	openclawGatewayPort           = 18789
	openclawAllowedOriginHost     = "127.0.0.1"
	openclawHTTPSVersion          = "2026.3.13"
	openclawHTTPVersion           = "2026.3.23"
	openclawTrustedProxyLoopback  = "127.0.0.1/32"
	defaultOpenclawNPMRegistry    = "https://registry.npmjs.org/"
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
		if isOpenclawHTTPSWindowVersion(detail.Version) {
			params["PANEL_APP_PORT_HTTPS"] = req.WebUIPort
		} else {
			params["PANEL_APP_PORT_HTTP"] = req.WebUIPort
		}
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
	agent, err := loadOpenclawAgentByID(req.ID)
	if err != nil {
		return err
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
	agent, err := loadOpenclawAgentByID(req.AgentID)
	if err != nil {
		return err
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
	definitions := providercatalog.All()
	providers := make([]dto.ProviderInfo, 0, len(definitions))
	for key, def := range definitions {
		models := make([]dto.ProviderModelInfo, 0, len(def.Models))
		for _, item := range def.Models {
			models = append(models, dto.ProviderModelInfo{
				ID:            item.ID,
				Name:          item.Name,
				ContextWindow: item.ContextWindow,
				MaxTokens:     item.MaxTokens,
				Reasoning:     item.Reasoning,
				Input:         append([]string(nil), item.Input...),
			})
		}
		providers = append(providers, dto.ProviderInfo{
			Sort:        def.Sort,
			Provider:    key,
			DisplayName: def.DisplayName,
			BaseURL:     def.DefaultBaseURL,
			Models:      models,
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
	resolvedInput, err := resolveAgentAccountInput(provider, req.APIKey, req.BaseURL)
	if err != nil {
		return err
	}
	account := &model.AgentAccount{
		Provider:       resolvedInput.Provider,
		Name:           req.Name,
		APIKey:         resolvedInput.APIKey,
		RememberAPIKey: req.RememberAPIKey,
		BaseURL:        resolvedInput.BaseURL,
		APIType:        req.APIType,
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
	resolvedInput, err := resolveAgentAccountInput(provider, req.APIKey, req.BaseURL)
	if err != nil {
		return err
	}
	account.Name = req.Name
	account.APIKey = resolvedInput.APIKey
	account.RememberAPIKey = req.RememberAPIKey
	account.BaseURL = resolvedInput.BaseURL
	account.APIType = req.APIType
	account.Remark = req.Remark
	account.Verified = true

	if err := global.DB.Save(account).Error; err != nil {
		return err
	}
	terminalai.InvalidateTerminalRuntimeCache()
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
			ProviderName:   providercatalog.DisplayName(item.Provider),
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
	nextModel := cloneAgentAccountModel(req.Model)
	if _, ok := findAgentAccountModelForProvider(account.Provider, models, nextModel.ID); ok {
		return buserr.New("ErrRecordExist")
	}
	inputPayload, err := json.Marshal(nextModel.Input)
	if err != nil {
		return err
	}
	sortOrder := len(models) + 1
	record := &model.AgentAccountModel{
		AccountID:     account.ID,
		Model:         nextModel.ID,
		Name:          nextModel.Name,
		ContextWindow: nextModel.ContextWindow,
		MaxTokens:     nextModel.MaxTokens,
		Reasoning:     nextModel.Reasoning,
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
	nextModel := cloneAgentAccountModel(req.Model)
	for _, item := range models {
		if item.RecordID == req.Model.RecordID {
			continue
		}
		if sameProviderModelID(account.Provider, item.ID, nextModel.ID) {
			return buserr.New("ErrRecordExist")
		}
	}
	nextModels := make([]dto.AgentAccountModel, 0, len(models))
	for _, item := range models {
		if item.RecordID == req.Model.RecordID {
			nextModels = append(nextModels, nextModel)
			continue
		}
		nextModels = append(nextModels, item)
	}
	if err := ensureAccountModelsNotBound(account, nextModels); err != nil {
		return err
	}
	inputPayload, err := json.Marshal(nextModel.Input)
	if err != nil {
		return err
	}
	record.Model = nextModel.ID
	record.Name = nextModel.Name
	record.ContextWindow = nextModel.ContextWindow
	record.MaxTokens = nextModel.MaxTokens
	record.Reasoning = nextModel.Reasoning
	record.Input = string(inputPayload)
	if err := agentAccountModelRepo.Save(record); err != nil {
		return err
	}
	terminalai.InvalidateTerminalRuntimeCache()
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
	terminalai.InvalidateTerminalRuntimeCache()
	return a.syncAgentsByAccount(account)
}

func (a AgentService) SyncAgentsByAccount(account *model.AgentAccount) error {
	if account == nil || account.ID == 0 {
		return nil
	}
	return a.syncAgentsByAccount(account)
}

func (a AgentService) VerifyAccount(req dto.AgentAccountVerifyReq) error {
	_, err := resolveAgentAccountInput(req.Provider, req.APIKey, req.BaseURL)
	return err
}

func (a AgentService) DeleteAccount(req dto.AgentAccountDeleteReq) error {
	if exists, _ := agentRepo.GetFirst(repo.WithByAccountID(req.ID)); exists != nil && exists.ID > 0 {
		return buserr.New("ErrAgentAccountBound")
	}
	if err := agentAccountModelRepo.Delete(repo.WithByAccountID(req.ID)); err != nil {
		return err
	}
	terminalai.InvalidateTerminalRuntimeCache()
	return agentAccountRepo.DeleteByID(req.ID)
}

func (a AgentService) GetSecurityConfig(req dto.AgentSecurityConfigReq) (*dto.AgentSecurityConfig, error) {
	agent, _, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractSecurityConfig(conf)
	return &result, nil
}

func (a AgentService) UpdateSecurityConfig(req dto.AgentSecurityConfigUpdateReq) error {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
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
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	conf, err := readOpenclawConfig(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	result := extractOtherConfig(conf)
	npmRegistry, err := getOpenclawNPMRegistry(install.ContainerName)
	if err == nil {
		result.NPMRegistry = npmRegistry
	}
	return &result, nil
}

func (a AgentService) UpdateOtherConfig(req dto.AgentOtherConfigUpdateReq) error {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := ensureContainerRunning(install.ContainerName); err != nil {
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
	return setOpenclawNPMRegistry(install.ContainerName, req.NPMRegistry)
}

func (a AgentService) GetConfigFile(req dto.AgentConfigFileReq) (*dto.AgentConfigFile, error) {
	agent, _, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(agent.ConfigPath)
	if err != nil {
		return nil, err
	}
	return &dto.AgentConfigFile{Content: string(content)}, nil
}

func (a AgentService) UpdateConfigFile(req dto.AgentConfigFileUpdateReq) error {
	agent, install, err := a.loadOpenclawAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(req.Content), &payload); err != nil {
		return err
	}
	info, err := os.Stat(agent.ConfigPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(agent.ConfigPath, []byte(req.Content), info.Mode()); err != nil {
		return err
	}
	return NewIAppInstalledService().Operate(request.AppInstalledOperate{
		InstallId: install.ID,
		Operate:   constant.Restart,
	})
}

func getOpenclawNPMRegistry(containerName string) (string, error) {
	registry, err := cmd.RunDefaultWithStdoutBashCfAndTimeOut("docker exec %s npm get registry", 20*time.Second, containerName)
	if err != nil {
		return "", err
	}
	registry = strings.TrimSpace(registry)
	if registry == "" {
		return defaultOpenclawNPMRegistry, nil
	}
	return registry, nil
}

func setOpenclawNPMRegistry(containerName, registry string) error {
	return cmd.RunDefaultBashCf("docker exec %s npm set registry %q", containerName, registry)
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

func (a AgentService) loadOpenclawAgentAndInstall(agentID uint) (*model.Agent, *model.AppInstall, error) {
	agent, install, err := a.loadAgentAndInstall(agentID)
	if err != nil {
		return nil, nil, err
	}
	if agent.AgentType == constant.AppCopaw {
		return nil, nil, fmt.Errorf("copaw does not support")
	}
	return agent, install, nil
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

func (a AgentService) loadOpenclawAgentConfig(agentID uint) (*model.Agent, *model.AppInstall, map[string]interface{}, error) {
	agent, install, conf, err := a.loadAgentConfig(agentID)
	if err != nil {
		return nil, nil, nil, err
	}
	if agent.AgentType == constant.AppCopaw {
		return nil, nil, nil, fmt.Errorf("copaw does not support")
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
