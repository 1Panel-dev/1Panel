package service

import (
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

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

func (a AgentService) GetDiscordConfig(req dto.AgentIDReq) (*dto.AgentDiscordConfig, error) {
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

func (a AgentService) GetQQBotConfig(req dto.AgentIDReq) (*dto.AgentQQBotConfig, error) {
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

func (a AgentService) GetWecomConfig(req dto.AgentIDReq) (*dto.AgentWecomConfig, error) {
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

func (a AgentService) GetDingTalkConfig(req dto.AgentIDReq) (*dto.AgentDingTalkConfig, error) {
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
	agent, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	spec, pluginID, err := resolvePluginMeta(req.Type)
	if err != nil {
		return err
	}
	installTask, err := task.NewTaskWithOps(req.Type, task.TaskInstall, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	installTask.AddSubTask("Install OpenClaw plugin", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(10*time.Minute))
		if req.Type == "qqbot" {
			legacyPluginPath := path.Join(openclawPluginBaseDir, "qqbot")
			if err := mgr.RunBashCf("docker exec %s test -d %s", install.ContainerName, legacyPluginPath); err == nil {
				if err := mgr.Run("docker", "exec", "-i", install.ContainerName, "sh", "-c", "printf 'yes\\n' | openclaw plugins uninstall qqbot"); err != nil {
					return err
				}
			}
		}
		if err := mgr.Run("docker", "exec", install.ContainerName, "sh", "-c", buildOpenclawPluginInstallScript(spec, pluginID)); err != nil {
			return err
		}
		conf, err := readOpenclawConfig(agent.ConfigPath)
		if err != nil {
			return err
		}
		appendPluginAllow(conf, pluginID)
		return writeOpenclawConfigRaw(agent.ConfigPath, conf)
	}, nil)
	go func() {
		if err := installTask.Execute(); err != nil {
			global.LOG.Errorf("install openclaw plugin failed: %v", err)
		}
	}()
	return nil
}

func (a AgentService) LoginWeixinChannel(req dto.AgentWeixinLoginReq) error {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	loginTask, err := task.NewTaskWithOps("weixin", task.TaskExec, task.TaskScopeAI, req.TaskID, req.AgentID)
	if err != nil {
		return err
	}
	loginTask.AddSubTask("Login OpenClaw Weixin channel", func(t *task.Task) error {
		mgr := cmd.NewCommandMgr(cmd.WithTask(*t), cmd.WithContext(t.TaskCtx), cmd.WithTimeout(30*time.Minute))
		return mgr.RunBashCf("docker exec %s openclaw channels login --channel openclaw-weixin", install.ContainerName)
	}, nil)
	go func() {
		if err := loginTask.Execute(); err != nil {
			global.LOG.Errorf("login openclaw weixin channel failed: %v", err)
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

func (a AgentService) ApproveChannelPairing(req dto.AgentChannelPairingApproveReq) error {
	_, install, err := a.loadAgentAndInstall(req.AgentID)
	if err != nil {
		return err
	}
	if err := cmd.RunDefaultBashCf(
		"docker exec %s openclaw pairing approve %s %q",
		install.ContainerName,
		req.Type,
		strings.TrimSpace(req.PairingCode),
	); err != nil {
		return err
	}
	return nil
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

func extractDingTalkConfig(conf map[string]interface{}) dto.AgentDingTalkConfig {
	result := dto.AgentDingTalkConfig{
		Enabled:        true,
		DmPolicy:       "pairing",
		GroupPolicy:    "disabled",
		AllowFrom:      []string{},
		GroupAllowFrom: []string{},
	}
	channels, ok := conf["channels"].(map[string]interface{})
	if !ok {
		return result
	}
	dingtalk, ok := channels["dingtalk-connector"].(map[string]interface{})
	if !ok {
		return result
	}
	if enabled, ok := dingtalk["enabled"].(bool); ok {
		result.Enabled = enabled
	}
	if clientID, ok := dingtalk["clientId"].(string); ok {
		result.ClientID = clientID
	}
	if clientSecret, ok := dingtalk["clientSecret"].(string); ok {
		result.ClientSecret = clientSecret
	}
	if dmPolicy, ok := dingtalk["dmPolicy"].(string); ok && strings.TrimSpace(dmPolicy) != "" {
		result.DmPolicy = dmPolicy
	}
	if groupPolicy, ok := dingtalk["groupPolicy"].(string); ok && strings.TrimSpace(groupPolicy) != "" {
		result.GroupPolicy = groupPolicy
	}
	result.AllowFrom = extractStringList(dingtalk["allowFrom"])
	result.GroupAllowFrom = extractStringList(dingtalk["groupAllowFrom"])
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

func setDingTalkConfig(conf map[string]interface{}, config dto.AgentDingTalkConfig) {
	channels := ensureChildMap(conf, "channels")
	dingtalk := ensureChildMap(channels, "dingtalk-connector")
	dingtalk["enabled"] = config.Enabled
	dingtalk["clientId"] = strings.TrimSpace(config.ClientID)
	dingtalk["clientSecret"] = strings.TrimSpace(config.ClientSecret)
	dingtalk["dmPolicy"] = config.DmPolicy
	dingtalk["groupPolicy"] = config.GroupPolicy
	dingtalk["gatewayToken"] = extractGatewayToken(conf)
	switch config.DmPolicy {
	case "open":
		dingtalk["allowFrom"] = []string{"*"}
	case "allowlist":
		dingtalk["allowFrom"] = append([]string(nil), config.AllowFrom...)
	default:
		delete(dingtalk, "allowFrom")
	}
	switch config.GroupPolicy {
	case "open":
		dingtalk["groupAllowFrom"] = []string{"*"}
	case "allowlist":
		dingtalk["groupAllowFrom"] = append([]string(nil), config.GroupAllowFrom...)
	default:
		delete(dingtalk, "groupAllowFrom")
	}

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	dingtalkEntry := ensureChildMap(entries, "dingtalk-connector")
	dingtalkEntry["enabled"] = config.Enabled

	gateway := ensureChildMap(conf, "gateway")
	httpMap := ensureChildMap(gateway, "http")
	endpoints := ensureChildMap(httpMap, "endpoints")
	chatCompletions := ensureChildMap(endpoints, "chatCompletions")
	chatCompletions["enabled"] = true
}

func setQQBotConfig(conf map[string]interface{}, config dto.AgentQQBotConfig) {
	channels := ensureChildMap(conf, "channels")
	qqbot := ensureChildMap(channels, "qqbot")
	delete(qqbot, "dmPolicy")
	qqbot["enabled"] = config.Enabled
	qqbot["allowFrom"] = []string{"*"}
	qqbot["appId"] = strings.TrimSpace(config.AppID)
	qqbot["clientSecret"] = strings.TrimSpace(config.ClientSecret)

	plugins := ensureChildMap(conf, "plugins")
	entries := ensureChildMap(plugins, "entries")
	delete(entries, "qqbot")
	qqbotEntry := ensureChildMap(entries, "openclaw-qqbot")
	qqbotEntry["enabled"] = config.Enabled
}

func appendPluginAllow(conf map[string]interface{}, pluginID string) {
	plugins := ensureChildMap(conf, "plugins")
	allow := make([]string, 0, 4)
	seen := map[string]struct{}{}
	switch values := plugins["allow"].(type) {
	case []interface{}:
		for _, value := range values {
			text, ok := value.(string)
			if !ok || text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			seen[text] = struct{}{}
			allow = append(allow, text)
		}
	}
	if _, ok := seen[pluginID]; ok {
		plugins["allow"] = allow
		return
	}
	plugins["allow"] = append(allow, pluginID)
}

func buildOpenclawPluginInstallScript(spec, pluginID string) string {
	return fmt.Sprintf(
		"set -e; workdir=%s/%s; rm -rf \"$workdir\"; mkdir -p \"$workdir\"; cd \"$workdir\"; npm pack --silent %q >/dev/null 2>&1; pkg=$(find \"$workdir\" -maxdepth 1 -type f -name '*.tgz' | head -n 1); printf '%%s\\n' \"$pkg\"; openclaw plugins install \"$pkg\"; rm -rf \"$workdir\"",
		openclawPluginPackageTmpDir,
		pluginID,
		spec,
	)
}

func resolvePluginMeta(pluginType string) (string, string, error) {
	switch pluginType {
	case "qqbot":
		return "@tencent-connect/openclaw-qqbot@latest", "openclaw-qqbot", nil
	case "wecom":
		return "@wecom/wecom-openclaw-plugin", "wecom-openclaw-plugin", nil
	case "dingtalk":
		return "@dingtalk-real-ai/dingtalk-connector", "dingtalk-connector", nil
	case "weixin":
		return "@tencent-weixin/openclaw-weixin", "openclaw-weixin", nil
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
