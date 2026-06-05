package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
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
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/docker/docker/api/types/container"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type IAgentService interface {
	Create(req dto.AgentCreateReq) (*dto.AgentItem, error)
	BatchInstall(req dto.AgentBatchInstallReq) (*dto.AgentItem, error)
	BatchUpgrade(req dto.AgentBatchUpgradeReq) ([]dto.AgentBatchUpgradeResult, error)
	BatchInstallSkill(req dto.AgentBatchSkillInstallReq) ([]dto.AgentBatchSkillInstallResult, error)
	Page(req dto.SearchWithPage) (int64, []dto.AgentItem, error)
	DeleteCheck(req dto.AgentIDReq) ([]dto.AppResource, error)
	Delete(req dto.AgentDeleteReq) error
	ResetToken(req dto.AgentTokenResetReq) error
	UpdateRemark(req dto.AgentRemarkUpdateReq) error
	BindWebsite(req dto.AgentWebsiteBindReq) error
	UnbindWebsite(req dto.AgentIDReq) error
	GetModelConfig(req dto.AgentIDReq) (*dto.AgentModelConfig, error)
	UpdateModelConfig(req dto.AgentModelConfigUpdateReq) error
	GetHermesChatSessions(req dto.AgentIDReq) ([]dto.AgentHermesChatSessionItem, error)
	RenameHermesChatSession(req dto.AgentHermesChatSessionRenameReq) error
	DeleteHermesChatSession(req dto.AgentHermesChatSessionDeleteReq) error
	GetOverview(req dto.AgentOverviewReq) (*dto.AgentOverview, error)
	GetProviders() ([]dto.ProviderInfo, error)
	GetSecurityConfig(req dto.AgentIDReq) (*dto.AgentSecurityConfig, error)
	UpdateSecurityConfig(req dto.AgentSecurityConfigUpdateReq) error
	GetOtherConfig(req dto.AgentIDReq) (*dto.AgentOtherConfig, error)
	UpdateOtherConfig(req dto.AgentOtherConfigUpdateReq) error
	GetConfigFile(req dto.AgentConfigFileReq) (*dto.AgentConfigFile, error)
	UpdateConfigFile(req dto.AgentConfigFileUpdateReq) error
	ListSkills(req dto.AgentIDReq) ([]dto.AgentSkillItem, error)
	SearchSkills(req dto.AgentSkillSearchReq) ([]dto.AgentSkillSearchItem, error)
	UpdateSkill(req dto.AgentSkillUpdateReq) error
	InstallSkill(req dto.AgentSkillInstallReq) error
	UninstallSkill(req dto.AgentSkillUninstallReq) error

	CreateRole(req dto.AgentRoleCreateReq) (*dto.AgentRoleCreateResp, error)
	DeleteRole(req dto.AgentRoleDeleteReq) error
	BindRole(req dto.AgentRoleBindReq) error
	UnbindRole(req dto.AgentRoleBindReq) error
	GetConfiguredAgents(req dto.AgentConfiguredAgentsReq) ([]dto.AgentConfiguredAgentItem, error)
	GetRoleChannels(req dto.AgentRoleChannelsReq) ([]dto.AgentRoleChannelItem, error)
	GetRoleMarkdownFiles(req dto.AgentRoleMarkdownFilesReq) ([]dto.AgentRoleMarkdownFileItem, error)
	UpdateRoleMarkdownFiles(req dto.AgentRoleMarkdownFilesUpdateReq) error

	CreateAccount(req dto.AgentAccountCreateReq) error
	UpdateAccount(req dto.AgentAccountUpdateReq) error
	SyncAgentsByAccount(account *model.AgentAccount) error
	PageAccounts(req dto.AgentAccountSearch) (int64, []dto.AgentAccountInfo, error)
	CountAccountsByProviders(req dto.AgentAccountProviderCountReq) (map[string]int64, error)
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
	GetDiscordConfig(req dto.AgentIDReq) (*dto.AgentDiscordConfig, error)
	UpdateDiscordConfig(req dto.AgentDiscordConfigUpdateReq) error
	GetWecomConfig(req dto.AgentIDReq) (*dto.AgentWecomConfig, error)
	UpdateWecomConfig(req dto.AgentWecomConfigUpdateReq) error
	GetDingTalkConfig(req dto.AgentIDReq) (*dto.AgentDingTalkConfig, error)
	UpdateDingTalkConfig(req dto.AgentDingTalkConfigUpdateReq) error
	GetWeixinConfig(req dto.AgentIDReq) (*dto.AgentWeixinConfig, error)
	LoginWeixinChannel(req dto.AgentWeixinLoginReq) error
	GetQQBotConfig(req dto.AgentIDReq) (*dto.AgentQQBotConfig, error)
	UpdateQQBotConfig(req dto.AgentQQBotConfigUpdateReq) error
	DeleteChannelConfig(req dto.AgentChannelDeleteReq) error
	InstallPlugin(req dto.AgentPluginInstallReq) error
	UpgradePlugin(req dto.AgentPluginUpgradeReq) error
	UninstallPlugin(req dto.AgentPluginUninstallReq) error
	CheckPlugin(req dto.AgentPluginCheckReq) (*dto.AgentPluginStatus, error)
	ApproveChannelPairing(req dto.AgentChannelPairingApproveReq) error
}

type batchUpgradePlan struct {
	agent model.Agent
	req   request.AppInstallUpgrade
}

const (
	defaultBrowserExecutablePath  = "/home/node/.cache/ms-playwright/openclaw-browser"
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
	if !xpack.MultiNodeProvider.IsXpack() {
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

	if agentType == constant.AppOpenclaw || agentType == constant.AppHermesAgent {
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
		baseURL = account.BaseURL
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
	}

	if agentType == constant.AppOpenclaw {
		var err error
		allowedOrigins, err = normalizeAllowedOrigins(req.AllowedOrigins)
		if err != nil {
			return nil, err
		}
		if len(allowedOrigins) == 0 {
			return nil, fmt.Errorf("allowed origins is required")
		}
		token = strings.TrimSpace(req.Token)
		if token == "" {
			token = generateToken()
		}
		installHooks = &appInstallHooks{
			AfterCopyData: func(appInstall *model.AppInstall) error {
				return prepareOpenclawInstallFiles(appInstall, account, storedModel, token, allowedOrigins)
			},
		}
	} else if agentType == constant.AppHermesAgent {
		installHooks = &appInstallHooks{
			AfterCopyData: func(appInstall *model.AppInstall) error {
				return prepareHermesInstallFiles(appInstall, account, storedModel)
			},
		}
	}

	params := map[string]interface{}{
		constant.CPUS:        "0",
		constant.MemoryLimit: "0",
		constant.HostIP:      "",
	}
	setAgentWebUIParams(params, agentType, detail.Version, req.WebUIPort)
	if agentType == constant.AppOpenclaw {
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
	if agentType == constant.AppHermesAgent {
		configPath = path.Join(appInstall.GetPath(), "data", "config.yaml")
	}
	agent := &model.Agent{
		Name:          req.Name,
		Remark:        req.Remark,
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

func (a AgentService) BatchInstall(req dto.AgentBatchInstallReq) (*dto.AgentItem, error) {
	accountID, err := a.ensureBatchInstallAccount(req)
	if err != nil {
		return nil, err
	}
	createReq := buildCreateReqFromBatchInstallReq(req)
	createReq.AccountID = accountID
	if strings.TrimSpace(createReq.Name) == "" {
		name, err := buildBatchInstallAgentName(createReq.AgentType)
		if err != nil {
			return nil, err
		}
		createReq.Name = name
	}
	if createReq.AgentType == constant.AppOpenclaw && len(createReq.AllowedOrigins) == 0 {
		allowedOrigin, err := buildBatchInstallOpenclawAllowedOrigin(req)
		if err != nil {
			return nil, err
		}
		createReq.AllowedOrigins = []string{allowedOrigin}
	}
	return a.Create(createReq)
}

func (a AgentService) BatchUpgrade(req dto.AgentBatchUpgradeReq) ([]dto.AgentBatchUpgradeResult, error) {
	plans, results, err := buildBatchUpgradePlans(req)
	if err != nil {
		return nil, err
	}
	for _, plan := range plans {
		result := dto.AgentBatchUpgradeResult{
			AgentID:      plan.agent.ID,
			AgentName:    plan.agent.Name,
			AppInstallID: plan.agent.AppInstallID,
		}
		if err := upgradeInstall(plan.req); err != nil {
			result.Message = err.Error()
		} else {
			result.Success = true
		}
		results = append(results, result)
	}
	return results, nil
}

func (a AgentService) BatchInstallSkill(req dto.AgentBatchSkillInstallReq) ([]dto.AgentBatchSkillInstallResult, error) {
	if req.AgentType != constant.AppOpenclaw && req.AgentType != constant.AppHermesAgent {
		return nil, fmt.Errorf("%s does not support skill batch install", req.AgentType)
	}
	packagePath, err := validateLocalSkillPackagePath(req.PackagePath)
	if err != nil {
		return nil, err
	}
	if format, err := localSkillPackageFormat(packagePath); err != nil {
		return nil, err
	} else if format != "zip" {
		return nil, fmt.Errorf("only .zip skill packages can be installed")
	}
	skillName := sanitizeLocalSkillDirName(req.SkillName, packagePath, req.SkillName)
	agents, err := agentRepo.List(func(db *gorm.DB) *gorm.DB {
		return db.Where("agent_type = ?", req.AgentType).Order("id ASC")
	})
	if err != nil {
		return nil, err
	}
	results := make([]dto.AgentBatchSkillInstallResult, 0, len(agents))
	for _, agent := range agents {
		result := dto.AgentBatchSkillInstallResult{
			AgentID:      agent.ID,
			AgentName:    agent.Name,
			AppInstallID: agent.AppInstallID,
		}
		if agent.AppInstallID == 0 {
			result.Message = "agent app install id is empty"
			results = append(results, result)
			continue
		}
		install, err := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
		if err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}
		result.AppInstallID = install.ID
		if install.App.Key != req.AgentType {
			result.Message = fmt.Sprintf("app key %s does not match agent type %s", install.App.Key, req.AgentType)
			results = append(results, result)
			continue
		}
		if install.Status == constant.StatusInstalling || install.Status == constant.StatusUpgrading {
			result.Message = fmt.Sprintf("agent status is %s", install.Status)
			results = append(results, result)
			continue
		}
		if err := ensureContainerRunning(install.ContainerName); err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}
		installTask, err := task.NewTaskWithOps(skillName, task.TaskInstall, task.TaskScopeAI, buildBatchSkillInstallTaskID(req.TaskID, agent.ID), agent.ID)
		if err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}
		currentAgent := agent
		currentInstall := install
		installTask.AddSubTask("Install local skill", func(t *task.Task) error {
			mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(20*time.Minute))
			return installLocalSkillPackage(mgr, currentInstall.ContainerName, currentAgent.AgentType, currentAgent.ConfigPath, packagePath, skillName)
		}, nil)
		go func() {
			if err := installTask.Execute(); err != nil {
				global.LOG.Errorf("batch install local skill failed: %v", err)
			}
		}()
		result.Success = true
		results = append(results, result)
	}
	return results, nil
}

func buildBatchUpgradePlans(req dto.AgentBatchUpgradeReq) ([]batchUpgradePlan, []dto.AgentBatchUpgradeResult, error) {
	agents, err := agentRepo.List(func(db *gorm.DB) *gorm.DB {
		return db.Where("agent_type = ?", req.AgentType).Order("id ASC")
	})
	if err != nil {
		return nil, nil, err
	}
	plans := make([]batchUpgradePlan, 0, len(agents))
	results := make([]dto.AgentBatchUpgradeResult, 0)
	for _, agent := range agents {
		result := dto.AgentBatchUpgradeResult{
			AgentID:      agent.ID,
			AgentName:    agent.Name,
			AppInstallID: agent.AppInstallID,
		}
		if agent.AppInstallID == 0 {
			result.Message = "agent app install id is empty"
			results = append(results, result)
			continue
		}
		install, err := appInstallRepo.GetFirst(repo.WithByID(agent.AppInstallID))
		if err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}
		result.AppInstallID = install.ID
		if install.App.Key != req.AgentType {
			result.Message = fmt.Sprintf("app key %s does not match agent type %s", install.App.Key, req.AgentType)
			results = append(results, result)
			continue
		}
		if install.Status == constant.StatusInstalling || install.Status == constant.StatusUpgrading {
			result.Message = fmt.Sprintf("agent status is %s", install.Status)
			results = append(results, result)
			continue
		}
		if install.Version == req.TargetVersion {
			result.Success = true
			result.Skipped = true
			result.Message = "already target version"
			results = append(results, result)
			continue
		}
		detail, err := appDetailRepo.GetFirst(appDetailRepo.WithAppId(install.AppId), appDetailRepo.WithVersion(req.TargetVersion))
		if err != nil {
			result.Message = err.Error()
			results = append(results, result)
			continue
		}
		if detail.ID == 0 {
			result.Message = fmt.Sprintf("target version %s not found", req.TargetVersion)
			results = append(results, result)
			continue
		}
		plans = append(plans, batchUpgradePlan{
			agent: agent,
			req: request.AppInstallUpgrade{
				InstallID: install.ID,
				DetailID:  detail.ID,
				Backup:    req.Backup,
				PullImage: req.PullImage,
				TaskID:    buildBatchUpgradeTaskID(req.TaskID, install.ID),
			},
		})
	}
	return plans, results, nil
}

func buildBatchUpgradeTaskID(taskID string, appInstallID uint) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		taskID = fmt.Sprintf("batch-upgrade-%d-%d", appInstallID, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%d", taskID, appInstallID)
}

func buildBatchSkillInstallTaskID(taskID string, agentID uint) string {
	taskID = strings.TrimSpace(taskID)
	if taskID == "" {
		taskID = fmt.Sprintf("batch-skill-install-%d-%d", agentID, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%d", taskID, agentID)
}

func buildCreateReqFromBatchInstallReq(req dto.AgentBatchInstallReq) dto.AgentCreateReq {
	return dto.AgentCreateReq{
		Name:           req.Name,
		Remark:         req.Remark,
		AppVersion:     req.AppVersion,
		WebUIPort:      req.WebUIPort,
		BridgePort:     req.BridgePort,
		AllowedOrigins: req.AllowedOrigins,
		AgentType:      req.AgentType,
		Model:          req.Model,
		AccountID:      req.AccountID,
		Token:          req.Token,
		TaskID:         req.TaskID,
		Advanced:       req.Advanced,
		ContainerName:  req.ContainerName,
		AllowPort:      req.AllowPort,
		SpecifyIP:      req.SpecifyIP,
		RestartPolicy:  req.RestartPolicy,
		CpuQuota:       req.CpuQuota,
		MemoryLimit:    req.MemoryLimit,
		MemoryUnit:     req.MemoryUnit,
		PullImage:      req.PullImage,
		EditCompose:    req.EditCompose,
		DockerCompose:  req.DockerCompose,
	}
}

func buildBatchInstallAgentName(agentType string) (string, error) {
	prefix := strings.ToLower(strings.TrimSpace(agentType))
	if prefix == "" {
		prefix = "agent"
	}
	for i := 1; i <= 10000; i++ {
		name := fmt.Sprintf("%s-%d", prefix, i)
		exists, err := agentNameExists(name)
		if err != nil {
			return "", err
		}
		if !exists {
			return name, nil
		}
	}
	return "", fmt.Errorf("no available agent name")
}

func agentNameExists(name string) (bool, error) {
	if exist, err := agentRepo.GetFirst(repo.WithByLowerName(name)); err == nil && exist != nil && exist.ID > 0 {
		return true, nil
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}
	installs, err := appInstallRepo.ListBy(context.Background(), repo.WithByLowerName(name))
	if err != nil {
		return false, err
	}
	return len(installs) > 0, nil
}

func buildBatchInstallOpenclawAllowedOrigin(req dto.AgentBatchInstallReq) (string, error) {
	host := ""
	if value, err := settingRepo.GetValueByKey("SystemIP"); err == nil {
		host = strings.TrimSpace(value)
	}
	if host == "" {
		host = strings.TrimSpace(req.FallbackAccessHost)
	}
	if host == "" {
		host = "127.0.0.1"
	}
	return buildOpenclawAllowedOrigin(openclawAllowedOriginScheme(req.AppVersion), host, req.WebUIPort)
}

func (a AgentService) ensureBatchInstallAccount(req dto.AgentBatchInstallReq) (uint, error) {
	if !requiresAgentAccount(req.AgentType) {
		return 0, nil
	}
	if req.MasterAccountID == 0 {
		return 0, buserr.New("ErrInvalidParams")
	}
	snapshot := req.AccountSnapshot
	account, err := agentAccountRepo.GetFirst(agentAccountRepo.WithByMasterAccountID(req.MasterAccountID))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if account == nil && global.IsMaster {
		account, err = agentAccountRepo.GetFirst(repo.WithByID(req.MasterAccountID))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, err
		}
	}
	if account == nil {
		account = &model.AgentAccount{
			MasterAccountID: req.MasterAccountID,
		}
	}
	account.MasterAccountID = req.MasterAccountID
	account.Provider = snapshot.Provider
	account.Name = snapshot.Name
	account.APIKey = snapshot.APIKey
	account.RememberAPIKey = snapshot.RememberAPIKey
	account.BaseURL = snapshot.BaseURL
	account.APIType = snapshot.APIType
	account.Remark = snapshot.Remark
	account.Verified = true

	initialModels, err := buildInitialAgentAccountModels(account, req.AccountModels)
	if err != nil {
		return 0, err
	}
	if err := global.DB.Transaction(func(tx *gorm.DB) error {
		if account.ID == 0 {
			if err := tx.Create(account).Error; err != nil {
				return err
			}
		} else {
			if err := tx.Save(account).Error; err != nil {
				return err
			}
		}
		return replacePersistedAgentAccountModelsWithTx(tx, account.ID, initialModels)
	}); err != nil {
		return 0, err
	}
	return account.ID, nil
}

func requiresAgentAccount(agentType string) bool {
	return agentType != constant.AppCopaw
}

func setAgentWebUIParams(params map[string]interface{}, agentType, appVersion string, webUIPort int) {
	if agentType == constant.AppOpenclaw && isOpenclawHTTPSWindowVersion(appVersion) {
		params["PANEL_APP_PORT_HTTPS"] = webUIPort
		return
	}
	params["PANEL_APP_PORT_HTTP"] = webUIPort
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
	appInstalls := make([]model.AppInstall, 0, len(list))
	for _, item := range list {
		appInstall, _ := appInstallRepo.GetFirst(repo.WithByID(item.AppInstallID))
		appInstalls = append(appInstalls, appInstall)
	}
	syncAgentAppInstalls(appInstalls)
	for index, item := range list {
		appInstall := appInstalls[index]
		envMap := readInstallEnv(appInstall.Env)
		agentItem := buildAgentItem(&item, &appInstall, envMap)
		agentItem.Upgradable = checkAgentUpgradable(appInstall)
		items = append(items, agentItem)
	}
	if err := hydrateAgentWebsiteItems(items); err != nil {
		return 0, nil, err
	}
	return count, items, nil
}

func (a AgentService) Delete(req dto.AgentDeleteReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	resources, err := a.deleteCheckByAgent(agent)
	if err != nil {
		return err
	}
	if len(resources) > 0 {
		return buserr.New("ErrAgentWebsiteBound")
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

func (a AgentService) DeleteCheck(req dto.AgentIDReq) ([]dto.AppResource, error) {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return nil, err
	}
	return a.deleteCheckByAgent(agent)
}

func (a AgentService) deleteCheckByAgent(agent *model.Agent) ([]dto.AppResource, error) {
	if agent == nil || agent.WebsiteID == 0 {
		return nil, nil
	}
	website, err := websiteRepo.GetFirst(repo.WithByID(agent.WebsiteID))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	websiteName, err := loadAgentWebsiteResourceName(website)
	if err != nil {
		return nil, err
	}
	return []dto.AppResource{{Type: "website", Name: websiteName}}, nil
}

func syncAgentAppInstalls(appInstalls []model.AppInstall) {
	if len(appInstalls) == 0 {
		return
	}

	var containersMap map[string]container.Summary
	cli, err := docker.NewClient()
	if err == nil {
		defer cli.Close()
		containers, err := cli.ListAllContainers()
		if err == nil {
			containersMap = make(map[string]container.Summary, len(containers))
			for _, contain := range containers {
				containersMap[contain.Names[0]] = contain
			}
		}
	}

	for index := range appInstalls {
		if appInstalls[index].ID == 0 || doNotNeedSync(appInstalls[index]) {
			continue
		}
		synAppInstall(containersMap, &appInstalls[index], false)
	}
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

func (a AgentService) UpdateRemark(req dto.AgentRemarkUpdateReq) error {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	agent.Remark = req.Remark
	return agentRepo.Save(agent)
}

func (a AgentService) GetModelConfig(req dto.AgentIDReq) (*dto.AgentModelConfig, error) {
	agent, err := agentRepo.GetFirst(repo.WithByID(req.AgentID))
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		cfg, err := readHermesConfig(agent.ConfigPath)
		if err != nil {
			return nil, err
		}
		account, err := agentAccountRepo.GetFirst(repo.WithByID(agent.AccountID))
		if err != nil {
			return nil, err
		}
		accountModels, err := loadAgentAccountModels(account)
		if err != nil {
			return nil, err
		}
		model, err := resolveHermesConfiguredModelID(account, accountModels, cfg.Model.Default)
		if err != nil {
			return nil, err
		}
		return &dto.AgentModelConfig{
			AccountID: agent.AccountID,
			Model:     model,
			Fallbacks: []string{},
		}, nil
	}
	agent, _, conf, err := a.loadOpenclawAgentConfig(req.AgentID)
	if err != nil {
		return nil, err
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(agent.AccountID))
	if err != nil {
		return nil, err
	}
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return nil, err
	}
	model := extractOpenclawPrimaryModelID(conf, account, models)
	if model == "" {
		model = agent.Model
	}
	return &dto.AgentModelConfig{
		AccountID: agent.AccountID,
		Model:     model,
		Fallbacks: extractOpenclawFallbackModelIDs(conf, account, models, model),
	}, nil
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
	resolvedRuntime, err := resolveOpenclawAccountModelRuntimeByID(account, req.Model)
	if err != nil {
		return err
	}
	modelName := resolvedRuntime.StoredModel
	apiType, maxTokens, contextWindow := resolvedRuntime.APIType, resolvedRuntime.MaxTokens, resolvedRuntime.ContextWindow
	confDir := path.Dir(agent.ConfigPath)
	if agent.AgentType == constant.AppHermesAgent {
		cfg, err := readHermesConfig(agent.ConfigPath)
		if err != nil {
			return err
		}
		if err := writeHermesConfig(confDir, account, modelName, cfg.Timezone); err != nil {
			return err
		}
	} else {
		if agent.AgentType != constant.AppOpenclaw {
			return fmt.Errorf("%s does not support", agent.AgentType)
		}
		if err := writeOpenclawConfig(confDir, account, modelName, agent.Token, nil, req.Fallbacks); err != nil {
			return err
		}
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
	terminalai.InvalidateFileAIRuntimeCache()
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
			ID:              item.ID,
			MasterAccountID: item.MasterAccountID,
			Provider:        item.Provider,
			ProviderName:    providercatalog.DisplayName(item.Provider),
			Name:            item.Name,
			APIKey:          apiKey,
			RememberAPIKey:  item.RememberAPIKey,
			BaseURL:         item.BaseURL,
			Models:          nil,
			APIType:         item.APIType,
			Verified:        item.Verified,
			Remark:          item.Remark,
			CreatedAt:       item.CreatedAt,
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

func (a AgentService) CountAccountsByProviders(req dto.AgentAccountProviderCountReq) (map[string]int64, error) {
	return agentAccountRepo.CountByProviders(req.Providers)
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
	terminalai.InvalidateFileAIRuntimeCache()
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
	terminalai.InvalidateFileAIRuntimeCache()
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
	if aiStatus, _ := settingRepo.GetValueByKey("AIStatus"); strings.EqualFold(strings.TrimSpace(aiStatus), constant.StatusEnable) {
		if aiAccountID, _ := settingRepo.GetValueByKey("AIAccountID"); strings.TrimSpace(aiAccountID) == strconv.FormatUint(uint64(req.ID), 10) {
			return buserr.New("ErrTerminalAIAccountInUse")
		}
	}
	if err := agentAccountModelRepo.Delete(repo.WithByAccountID(req.ID)); err != nil {
		return err
	}
	terminalai.InvalidateTerminalRuntimeCache()
	terminalai.InvalidateFileAIRuntimeCache()
	return agentAccountRepo.DeleteByID(req.ID)
}

func (a AgentService) GetSecurityConfig(req dto.AgentIDReq) (*dto.AgentSecurityConfig, error) {
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

func (a AgentService) GetOtherConfig(req dto.AgentIDReq) (*dto.AgentOtherConfig, error) {
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return nil, err
	}
	if agent.AgentType == constant.AppHermesAgent {
		cfg, err := readHermesConfig(agent.ConfigPath)
		if err != nil {
			return nil, err
		}
		return &dto.AgentOtherConfig{
			UserTimezone:   cfg.Timezone,
			BrowserEnabled: true,
			NPMRegistry:    "https://registry.npmjs.org/",
		}, nil
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
	if agent.AgentType == constant.AppHermesAgent {
		account, err := agentAccountRepo.GetFirst(repo.WithByID(agent.AccountID))
		if err != nil {
			return err
		}
		if err := writeHermesConfig(path.Dir(agent.ConfigPath), account, agent.Model, strings.TrimSpace(req.UserTimezone)); err != nil {
			return err
		}
		return NewIAppInstalledService().Operate(request.AppInstalledOperate{
			InstallId: install.ID,
			Operate:   constant.Restart,
		})
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
	if err := setOpenclawNPMRegistry(install.ContainerName, req.NPMRegistry); err != nil {
		return err
	}
	return nil
}

func (a AgentService) GetConfigFile(req dto.AgentConfigFileReq) (*dto.AgentConfigFile, error) {
	agent, _, err := a.loadAgentAndInstall(req.AgentID)
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
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := validateAgentConfigFileContent(agent.AgentType, req.Content); err != nil {
		return err
	}
	info, err := os.Stat(agent.ConfigPath)
	if err != nil {
		return err
	}
	if err := os.WriteFile(agent.ConfigPath, []byte(req.Content), info.Mode()); err != nil {
		return err
	}
	if agent.AgentType == constant.AppHermesAgent {
		cfg, err := readHermesConfig(agent.ConfigPath)
		if err != nil {
			return err
		}
		if cfg.Model.Default != "" {
			agent.Model = cfg.Model.Default
			if err := agentRepo.Save(agent); err != nil {
				return err
			}
		}
	}
	return NewIAppInstalledService().Operate(request.AppInstalledOperate{
		InstallId: install.ID,
		Operate:   constant.Restart,
	})
}

func validateAgentConfigFileContent(agentType, content string) error {
	var payload interface{}
	if agentType == constant.AppHermesAgent {
		return yaml.Unmarshal([]byte(content), &payload)
	}
	return json.Unmarshal([]byte(content), &payload)
}

func getOpenclawNPMRegistry(containerName string) (string, error) {
	registry, err := cmd.RunDockerExecWithStdout(20*time.Second, containerName, "npm", "get", "registry")
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
	return cmd.NewCommandMgr().Run("docker", "exec", containerName, "npm", "set", "registry", registry)
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
	if agent.AgentType != constant.AppOpenclaw {
		return nil, nil, fmt.Errorf("%s does not support", agent.AgentType)
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
	if agent.AgentType != constant.AppOpenclaw {
		return nil, nil, nil, fmt.Errorf("%s does not support", agent.AgentType)
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
	for _, agent := range agents {
		selectedAccountModel, err := requireAgentAccountModelForProvider(account.Provider, accountModels, agent.Model)
		if err != nil {
			return buserr.WithName("ErrAgentModelInUse", agent.Name)
		}
		resolvedRuntime, err := buildOpenclawAccountModelRuntime(account, selectedAccountModel)
		if err != nil {
			return err
		}
		modelName := resolvedRuntime.StoredModel
		apiType, maxTokens, contextWindow := resolvedRuntime.APIType, resolvedRuntime.MaxTokens, resolvedRuntime.ContextWindow
		confDir := path.Dir(agent.ConfigPath)
		switch agent.AgentType {
		case constant.AppOpenclaw:
			conf, err := readOpenclawConfig(agent.ConfigPath)
			if err != nil {
				return err
			}
			fallbacks := extractOpenclawFallbackModelIDs(conf, account, accountModels, selectedAccountModel.ID)
			if err := writeOpenclawConfig(confDir, account, modelName, agent.Token, nil, fallbacks); err != nil {
				return err
			}
		case constant.AppHermesAgent:
			cfg, err := readHermesConfig(agent.ConfigPath)
			if err != nil {
				return err
			}
			if err := writeHermesConfig(confDir, account, modelName, cfg.Timezone); err != nil {
				return err
			}
		default:
			continue
		}
		agent.BaseURL = account.BaseURL
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
