package service

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

const hermesWeixinLoginScript = `import asyncio
from gateway.platforms.weixin import qr_login
from hermes_cli.config import save_env_value

creds = asyncio.run(qr_login('/opt/data'))
print('RESULT =', creds)

if not creds:
    raise SystemExit(1)

save_env_value('WEIXIN_ACCOUNT_ID', creds.get('account_id', ''))
save_env_value('WEIXIN_TOKEN', creds.get('token', ''))
if creds.get('base_url'):
    save_env_value('WEIXIN_BASE_URL', creds['base_url'])
save_env_value('WEIXIN_CDN_BASE_URL', 'https://novac2c.cdn.weixin.qq.com/c2c')
save_env_value('WEIXIN_DM_POLICY', 'open')
save_env_value('WEIXIN_ALLOW_ALL_USERS', 'true')
save_env_value('WEIXIN_ALLOWED_USERS', '')
save_env_value('WEIXIN_GROUP_POLICY', 'disabled')
save_env_value('WEIXIN_GROUP_ALLOWED_USERS', '')
save_env_value('WEIXIN_HOME_CHANNEL', creds.get('user_id', ''))

print('WEIXIN ENV WRITTEN')
print('Restart the Hermes-Agent container to apply the new Weixin settings.')
`

func readHermesQQBotChannelConfig(confDir string) (*dto.AgentQQBotConfig, error) {
	envMap, err := readHermesEnvMap(path.Join(confDir, ".env"))
	if err != nil {
		return nil, err
	}
	cfg, err := readHermesConfigMap(path.Join(confDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	platform := childMap(childMap(cfg, "platforms"), "qq")
	extra := childMap(platform, "extra")
	appID := envMap["QQ_APP_ID"]
	clientSecret := envMap["QQ_CLIENT_SECRET"]
	dmPolicy := extractStringValue(extra["dm_policy"])
	if dmPolicy == "" {
		dmPolicy = "open"
	}
	allowFrom := extractStringList(extra["allow_from"])
	groupPolicy := extractStringValue(extra["group_policy"])
	if groupPolicy == "" {
		groupPolicy = "open"
	}
	groupAllowFrom := extractStringList(extra["group_allow_from"])

	result := &dto.AgentQQBotConfig{
		Enabled:        extractBoolValue(platform["enabled"], false) && appID != "" && clientSecret != "",
		DmPolicy:       dmPolicy,
		AllowFrom:      allowFrom,
		GroupPolicy:    groupPolicy,
		GroupAllowFrom: groupAllowFrom,
	}
	if appID != "" || clientSecret != "" {
		result.Bots = []dto.AgentQQBotBot{
			{
				AgentChannelBotBase: dto.AgentChannelBotBase{
					AccountID: "default",
					Name:      "Default",
					Enabled:   true,
					IsDefault: true,
				},
				AppID:        appID,
				ClientSecret: clientSecret,
			},
		}
	}
	return result, nil
}

func writeHermesQQBotChannelConfig(confDir string, config dto.AgentQQBotConfig) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readHermesEnvMap(envPath)
	if err != nil {
		return err
	}

	defaultBot := getDefaultQQBot(config.Bots)
	envMap["QQ_APP_ID"] = defaultBot.AppID
	envMap["QQ_CLIENT_SECRET"] = defaultBot.ClientSecret
	delete(envMap, "QQ_ALLOWED_USERS")
	delete(envMap, "QQ_ALLOW_ALL_USERS")
	delete(envMap, "QQ_MARKDOWN_SUPPORT")
	if config.DmPolicy == "open" {
		envMap["QQ_ALLOW_ALL_USERS"] = "true"
	}
	if err := writeHermesEnvMap(envPath, envMap, []string{
		"QQ_APP_ID",
		"QQ_CLIENT_SECRET",
		"QQ_ALLOW_ALL_USERS",
		"QQ_HOME_CHANNEL",
		"QQ_HOME_CHANNEL_NAME",
		"QQ_STT_API_KEY",
		"QQ_STT_BASE_URL",
		"QQ_STT_MODEL",
	}); err != nil {
		return err
	}

	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}

	platform := ensureChildMap(ensureChildMap(cfg, "platforms"), "qq")
	platform["enabled"] = config.Enabled && defaultBot.AppID != "" && defaultBot.ClientSecret != ""
	extra := ensureChildMap(platform, "extra")
	extra["markdown_support"] = true
	extra["dm_policy"] = config.DmPolicy
	extra["group_policy"] = config.GroupPolicy
	delete(extra, "app_id")
	delete(extra, "client_secret")
	if config.DmPolicy == "allowlist" {
		extra["allow_from"] = append([]string(nil), config.AllowFrom...)
	} else {
		delete(extra, "allow_from")
	}
	if config.GroupPolicy == "allowlist" {
		extra["group_allow_from"] = append([]string(nil), config.GroupAllowFrom...)
	} else {
		delete(extra, "group_allow_from")
	}
	return writeHermesConfigMap(configPath, cfg)
}

func readHermesWecomChannelConfig(confDir string) (*dto.AgentWecomConfig, error) {
	envMap, err := readHermesEnvMap(path.Join(confDir, ".env"))
	if err != nil {
		return nil, err
	}
	cfg, err := readHermesConfigMap(path.Join(confDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	platform := childMap(childMap(cfg, "platforms"), "wecom")
	extra := childMap(platform, "extra")
	allowFrom := extractStringList(extra["allow_from"])
	if len(allowFrom) == 0 {
		allowFrom = splitHermesEnvList(envMap["WECOM_ALLOWED_USERS"])
	}
	groupAllowFrom := extractStringList(extra["group_allow_from"])
	dmPolicy := "pairing"
	if extractHermesEnvBool(envMap, "WECOM_ALLOW_ALL_USERS", false) {
		dmPolicy = "open"
	} else if len(allowFrom) > 0 {
		dmPolicy = "allowlist"
	} else if policy := extractStringValue(extra["dm_policy"]); policy != "" {
		dmPolicy = policy
	}
	groupPolicy := extractStringValue(extra["group_policy"])
	if groupPolicy == "" {
		groupPolicy = "open"
	}

	botID := envMap["WECOM_BOT_ID"]
	secret := envMap["WECOM_SECRET"]

	return &dto.AgentWecomConfig{
		Enabled:        extractBoolValue(platform["enabled"], false) && botID != "" && secret != "",
		DmPolicy:       dmPolicy,
		AllowFrom:      allowFrom,
		GroupPolicy:    groupPolicy,
		GroupAllowFrom: groupAllowFrom,
		BotID:          botID,
		Secret:         secret,
	}, nil
}

func writeHermesWecomChannelConfig(confDir string, config dto.AgentWecomConfig) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readHermesEnvMap(envPath)
	if err != nil {
		return err
	}
	if config.BotID != "" {
		envMap["WECOM_BOT_ID"] = config.BotID
	} else {
		delete(envMap, "WECOM_BOT_ID")
	}
	if config.Secret != "" {
		envMap["WECOM_SECRET"] = config.Secret
	} else {
		delete(envMap, "WECOM_SECRET")
	}
	delete(envMap, "WECOM_ALLOWED_USERS")
	delete(envMap, "WECOM_ALLOW_ALL_USERS")
	delete(envMap, "WECOM_DM_POLICY")
	delete(envMap, "WECOM_GROUP_POLICY")
	switch config.DmPolicy {
	case "open":
		envMap["WECOM_ALLOW_ALL_USERS"] = "true"
	case "allowlist":
		if allow := joinHermesEnvList(config.AllowFrom); allow != "" {
			envMap["WECOM_ALLOWED_USERS"] = allow
		}
	}
	if err := writeHermesEnvMap(envPath, envMap, []string{
		"WECOM_BOT_ID",
		"WECOM_SECRET",
		"WECOM_ALLOW_ALL_USERS",
		"WECOM_ALLOWED_USERS",
	}); err != nil {
		return err
	}

	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	platform := ensureChildMap(ensureChildMap(cfg, "platforms"), "wecom")
	platform["enabled"] = config.Enabled && config.BotID != "" && config.Secret != ""
	extra := ensureChildMap(platform, "extra")
	extra["dm_policy"] = config.DmPolicy
	extra["group_policy"] = config.GroupPolicy
	delete(extra, "bot_id")
	delete(extra, "secret")
	if config.DmPolicy == "allowlist" {
		extra["allow_from"] = append([]string(nil), config.AllowFrom...)
	} else {
		delete(extra, "allow_from")
	}
	if config.GroupPolicy == "allowlist" {
		extra["group_allow_from"] = append([]string(nil), config.GroupAllowFrom...)
	} else {
		delete(extra, "group_allow_from")
	}
	return writeHermesConfigMap(configPath, cfg)
}

func readHermesDingTalkChannelConfig(confDir string) (*dto.AgentDingTalkConfig, error) {
	envMap, err := readHermesEnvMap(path.Join(confDir, ".env"))
	if err != nil {
		return nil, err
	}
	cfg, err := readHermesConfigMap(path.Join(confDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	platform := childMap(childMap(cfg, "platforms"), "dingtalk")
	allowFrom := splitHermesEnvList(envMap["DINGTALK_ALLOWED_USERS"])
	dmPolicy := "open"
	if len(allowFrom) > 0 {
		dmPolicy = "allowlist"
	}
	clientID := envMap["DINGTALK_CLIENT_ID"]
	clientSecret := envMap["DINGTALK_CLIENT_SECRET"]
	result := &dto.AgentDingTalkConfig{
		Enabled:        extractBoolValue(platform["enabled"], clientID != "" && clientSecret != ""),
		DmPolicy:       dmPolicy,
		AllowFrom:      allowFrom,
		GroupPolicy:    "open",
		GroupAllowFrom: []string{},
	}
	if clientID != "" || clientSecret != "" {
		result.Bots = []dto.AgentDingTalkBot{
			{
				AgentChannelBotBase: dto.AgentChannelBotBase{
					AccountID: "default",
					Name:      "Default",
					Enabled:   true,
					IsDefault: true,
				},
				ClientID:     clientID,
				ClientSecret: clientSecret,
			},
		}
	}
	return result, nil
}

func writeHermesDingTalkChannelConfig(confDir string, config dto.AgentDingTalkConfig) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readHermesEnvMap(envPath)
	if err != nil {
		return err
	}
	clientID, clientSecret := firstHermesDingTalkBotCredentials(config.Bots)
	if clientID != "" {
		envMap["DINGTALK_CLIENT_ID"] = clientID
	} else {
		delete(envMap, "DINGTALK_CLIENT_ID")
	}
	if clientSecret != "" {
		envMap["DINGTALK_CLIENT_SECRET"] = clientSecret
	} else {
		delete(envMap, "DINGTALK_CLIENT_SECRET")
	}
	delete(envMap, "DINGTALK_ALLOWED_USERS")
	if config.DmPolicy == "allowlist" {
		if allow := joinHermesEnvList(config.AllowFrom); allow != "" {
			envMap["DINGTALK_ALLOWED_USERS"] = allow
		}
	}
	if err := writeHermesEnvMap(envPath, envMap, []string{
		"DINGTALK_CLIENT_ID",
		"DINGTALK_CLIENT_SECRET",
		"DINGTALK_ALLOWED_USERS",
	}); err != nil {
		return err
	}

	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	platform := ensureChildMap(ensureChildMap(cfg, "platforms"), "dingtalk")
	platform["enabled"] = config.Enabled && clientID != "" && clientSecret != ""
	return writeHermesConfigMap(configPath, cfg)
}

func readHermesFeishuChannelConfig(confDir string) (*dto.AgentFeishuConfig, error) {
	envMap, err := readHermesEnvMap(path.Join(confDir, ".env"))
	if err != nil {
		return nil, err
	}
	cfg, err := readHermesConfigMap(path.Join(confDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	platform := childMap(childMap(cfg, "platforms"), "feishu")
	allowFrom := splitHermesEnvList(envMap["FEISHU_ALLOWED_USERS"])
	dmPolicy := "pairing"
	if extractHermesEnvBool(envMap, "FEISHU_ALLOW_ALL_USERS", false) {
		dmPolicy = "open"
	} else if len(allowFrom) > 0 {
		dmPolicy = "allowlist"
	}
	groupPolicy := envMap["FEISHU_GROUP_POLICY"]
	if groupPolicy == "" {
		groupPolicy = "allowlist"
	}
	appID := envMap["FEISHU_APP_ID"]
	appSecret := envMap["FEISHU_APP_SECRET"]
	result := &dto.AgentFeishuConfig{
		Enabled:        extractBoolValue(platform["enabled"], appID != "" && appSecret != ""),
		ThreadSession:  true,
		ReplyMode:      "auto",
		Streaming:      false,
		RequireMention: "true",
		GroupPolicy:    groupPolicy,
		GroupAllowFrom: []string{},
		Domain:         firstHermesEnvValue(envMap, "FEISHU_DOMAIN", "feishu"),
		ConnectionMode: firstHermesEnvValue(envMap, "FEISHU_CONNECTION_MODE", "websocket"),
	}
	if appID != "" || appSecret != "" {
		result.Bots = []dto.AgentFeishuBot{
			{
				AgentChannelBotBase: dto.AgentChannelBotBase{
					AccountID: "default",
					Name:      "Default",
					Enabled:   true,
					IsDefault: true,
				},
				AppID:     appID,
				AppSecret: appSecret,
				DmPolicy:  dmPolicy,
				AllowFrom: allowFrom,
			},
		}
	}
	return result, nil
}

func writeHermesFeishuChannelConfig(confDir string, config dto.AgentFeishuConfig) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readHermesEnvMap(envPath)
	if err != nil {
		return err
	}
	bot := firstHermesFeishuBot(config.Bots)
	if bot.AppID != "" {
		envMap["FEISHU_APP_ID"] = bot.AppID
	} else {
		delete(envMap, "FEISHU_APP_ID")
	}
	if bot.AppSecret != "" {
		envMap["FEISHU_APP_SECRET"] = bot.AppSecret
	} else {
		delete(envMap, "FEISHU_APP_SECRET")
	}
	envMap["FEISHU_DOMAIN"] = "feishu"
	envMap["FEISHU_CONNECTION_MODE"] = "websocket"
	envMap["FEISHU_GROUP_POLICY"] = config.GroupPolicy
	delete(envMap, "FEISHU_ALLOW_ALL_USERS")
	delete(envMap, "FEISHU_ALLOWED_USERS")
	switch bot.DmPolicy {
	case "open":
		envMap["FEISHU_ALLOW_ALL_USERS"] = "true"
	case "allowlist":
		if allow := joinHermesEnvList(bot.AllowFrom); allow != "" {
			envMap["FEISHU_ALLOWED_USERS"] = allow
		}
	}
	if err := writeHermesEnvMap(envPath, envMap, []string{
		"FEISHU_APP_ID",
		"FEISHU_APP_SECRET",
		"FEISHU_DOMAIN",
		"FEISHU_CONNECTION_MODE",
		"FEISHU_ALLOW_ALL_USERS",
		"FEISHU_ALLOWED_USERS",
		"FEISHU_GROUP_POLICY",
	}); err != nil {
		return err
	}

	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	platform := ensureChildMap(ensureChildMap(cfg, "platforms"), "feishu")
	platform["enabled"] = config.Enabled && bot.AppID != "" && bot.AppSecret != ""
	return writeHermesConfigMap(configPath, cfg)
}

func firstHermesDingTalkBotCredentials(bots []dto.AgentDingTalkBot) (string, string) {
	for _, bot := range bots {
		if bot.IsDefault || bot.AccountID == "default" {
			return bot.ClientID, bot.ClientSecret
		}
	}
	if len(bots) == 0 {
		return "", ""
	}
	return bots[0].ClientID, bots[0].ClientSecret
}

func firstHermesFeishuBot(bots []dto.AgentFeishuBot) dto.AgentFeishuBot {
	for _, bot := range bots {
		if bot.IsDefault || bot.AccountID == "default" {
			return bot
		}
	}
	if len(bots) == 0 {
		return dto.AgentFeishuBot{}
	}
	return bots[0]
}

func firstHermesEnvValue(envMap map[string]string, key string, defaultValue string) string {
	if envMap[key] != "" {
		return envMap[key]
	}
	return defaultValue
}

func buildHermesWeixinLoginArgs(containerName string) []string {
	return buildHermesDockerExecCommandArgs(
		containerName,
		"/opt/hermes/.venv/bin/python",
		"-u",
		"-c",
		hermesWeixinLoginScript,
	)
}
