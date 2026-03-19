package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	openclawutil "github.com/1Panel-dev/1Panel/agent/utils/openclaw"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"gorm.io/gorm"
)

type AgentService struct{}

func NewIAgentService() IAgentService {
	return &AgentService{}
}

type resolvedAgentAccountVerification struct {
	Provider string
	APIKey   string
	BaseURL  string
}

type resolvedAgentAccountInput struct {
	Provider string
	APIKey   string
	BaseURL  string
	APIType  string
}

func resolveAgentAccountVerification(provider, apiKey, baseURL string) (resolvedAgentAccountVerification, error) {
	resolvedAPIKey := strings.TrimSpace(apiKey)
	if resolvedAPIKey == "" {
		return resolvedAgentAccountVerification{}, buserr.New("ErrAgentApiKeyRequired")
	}
	resolvedBaseURL := strings.TrimSpace(baseURL)
	if fixedURL, ok := fixedProviderBaseURL(provider); ok {
		resolvedBaseURL = fixedURL
	}
	if (provider == "custom" || provider == "vllm") && resolvedBaseURL == "" {
		return resolvedAgentAccountVerification{}, buserr.New("ErrAgentBaseURLRequired")
	}
	if provider != "custom" && provider != "vllm" && resolvedBaseURL == "" {
		if defaultURL, ok := providerDefaultBaseURL(provider); ok {
			resolvedBaseURL = defaultURL
		}
	}
	if provider == "ollama" && resolvedBaseURL == "" {
		return resolvedAgentAccountVerification{}, buserr.New("ErrAgentBaseURLRequired")
	}
	return resolvedAgentAccountVerification{
		Provider: provider,
		APIKey:   resolvedAPIKey,
		BaseURL:  resolvedBaseURL,
	}, nil
}

func verifyResolvedAgentAccount(input resolvedAgentAccountVerification) error {
	if providercatalog.SkipVerification(input.Provider) {
		return nil
	}
	return providercatalog.VerifyAccount(input.Provider, input.BaseURL, input.APIKey)
}

func resolveAgentAccountAPIType(provider, apiType, fallbackAPIType string) (string, error) {
	resolvedAPIType := normalizeAPIType(apiType)
	if provider == "custom" || provider == "vllm" {
		if !isSupportedAPIType(resolvedAPIType) {
			return "", fmt.Errorf("apiType is invalid")
		}
		return resolvedAPIType, nil
	}
	if provider == "ollama" {
		if apiType == "" && fallbackAPIType != "" {
			resolvedAPIType = normalizeAPIType(fallbackAPIType)
			if !isSupportedOllamaAPIType(resolvedAPIType) {
				resolvedAPIType = "openai-responses"
			}
			return resolvedAPIType, nil
		}
		if !isSupportedOllamaAPIType(resolvedAPIType) {
			return "", fmt.Errorf("apiType is invalid")
		}
		return resolvedAPIType, nil
	}
	if fallbackAPIType != "" {
		return normalizeAPIType(fallbackAPIType), nil
	}
	return resolvedAPIType, nil
}

func resolveAgentAccountInput(provider, apiKey, baseURL, apiType, fallbackAPIType string) (resolvedAgentAccountInput, error) {
	resolvedVerification, err := resolveAgentAccountVerification(provider, apiKey, baseURL)
	if err != nil {
		return resolvedAgentAccountInput{}, err
	}
	if err := verifyResolvedAgentAccount(resolvedVerification); err != nil {
		return resolvedAgentAccountInput{}, err
	}
	resolvedAPIType, err := resolveAgentAccountAPIType(provider, apiType, fallbackAPIType)
	if err != nil {
		return resolvedAgentAccountInput{}, err
	}
	return resolvedAgentAccountInput{
		Provider: provider,
		APIKey:   resolvedVerification.APIKey,
		BaseURL:  resolvedVerification.BaseURL,
		APIType:  resolvedAPIType,
	}, nil
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
	switch pluginType {
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

func prepareOpenclawInstallFiles(appInstall *model.AppInstall, account *model.AgentAccount, modelName, token string, allowedOrigins []string) error {
	if appInstall == nil {
		return fmt.Errorf("app install is required")
	}
	confDir := path.Join(appInstall.GetPath(), "data", "conf")
	if err := writeOpenclawConfig(confDir, account, modelName, token, allowedOrigins); err != nil {
		return err
	}
	dataDir := path.Join(appInstall.GetPath(), "data")
	return files.NewFileOp().ChownR(dataDir, "1000", "1000", true)
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
		selectedModel = strings.TrimSpace(accountModels[0].ID)
	}
	if selectedModel == "" {
		return "", nil, nil, fmt.Errorf("model is required")
	}
	selectedAccountModel, err := requireAgentAccountModelForProvider(account.Provider, accountModels, selectedModel)
	if err != nil {
		return "", nil, nil, err
	}
	selectedModel = selectedAccountModel.ID

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

type openclawAccountModelRuntime struct {
	StoredModel   string
	PrimaryModel  string
	APIType       string
	MaxTokens     int
	ContextWindow int
}

func buildOpenclawAccountModelRuntime(account *model.AgentAccount, model dto.AgentAccountModel) (openclawAccountModelRuntime, error) {
	apiType, maxTokens, contextWindow := resolveRuntimeParams(
		account.Provider,
		account.APIType,
		model.MaxTokens,
		model.ContextWindow,
	)
	primaryModel, _, _, _, err := buildOpenclawCatalogModel(account, model)
	if err != nil {
		return openclawAccountModelRuntime{}, err
	}
	return openclawAccountModelRuntime{
		StoredModel:   model.ID,
		PrimaryModel:  primaryModel,
		APIType:       apiType,
		MaxTokens:     maxTokens,
		ContextWindow: contextWindow,
	}, nil
}

func resolveOpenclawAccountModelRuntimeByID(account *model.AgentAccount, modelID string) (openclawAccountModelRuntime, error) {
	accountModels, err := loadAgentAccountModels(account)
	if err != nil {
		return openclawAccountModelRuntime{}, err
	}
	selectedAccountModel, err := requireAgentAccountModelForProvider(account.Provider, accountModels, modelID)
	if err != nil {
		return openclawAccountModelRuntime{}, err
	}
	return buildOpenclawAccountModelRuntime(account, selectedAccountModel)
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

func buildInitialAgentAccountModels(account *model.AgentAccount, requested []dto.AgentAccountModel) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	if requiresInitialAgentAccountModels(account.Provider) && len(requested) > 1 {
		return nil, buserr.New("ErrAgentAccountSingleInitialModel")
	}
	if len(requested) > 0 {
		models, _, err := normalizeAgentAccountModels(account, requested, "", true)
		if err != nil {
			return nil, err
		}
		return models, nil
	}
	meta, ok := providercatalog.Get(account.Provider)
	if !ok || len(meta.Models) == 0 {
		if requiresInitialAgentAccountModels(account.Provider) {
			return nil, buserr.New("ErrAgentAccountModelsRequired")
		}
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
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	return listPersistedAgentAccountModels(account.ID)
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

func replacePersistedAgentAccountModelsWithTx(tx *gorm.DB, accountID uint, models []dto.AgentAccountModel) error {
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
}

func replacePersistedAgentAccountModels(accountID uint, models []dto.AgentAccountModel) error {
	return global.DB.Transaction(func(tx *gorm.DB) error {
		return replacePersistedAgentAccountModelsWithTx(tx, accountID, models)
	})
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
	switch provider {
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
			targetPrefix := parts[0]
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
	if account.ID > 0 {
		if agents, err := agentRepo.List(repo.WithByAccountID(account.ID)); err == nil {
			for _, agent := range agents {
				appendModel(agent.Model)
			}
		}
	}
	if definitions, ok := providerDefinitions()[account.Provider]; ok && len(definitions.Models) > 0 {
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

func sanitizeAgentAccountModelInputs(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := value
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

func requiresInitialAgentAccountModels(provider string) bool {
	switch provider {
	case "custom", "vllm", "ollama":
		return true
	default:
		return false
	}
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
	prefix := parts[0]
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

func requireAgentAccountModelForProvider(provider string, models []dto.AgentAccountModel, modelID string) (dto.AgentAccountModel, error) {
	selectedAccountModel, ok := findAgentAccountModelForProvider(provider, models, modelID)
	if !ok {
		return dto.AgentAccountModel{}, buserr.New("ErrAgentModelNotInAccount")
	}
	return selectedAccountModel, nil
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
	switch provider {
	case "bailian-coding-plan":
		return providerDefaultBaseURL(provider)
	case "ark-coding-plan":
		return providerDefaultBaseURL(provider)
	default:
		return "", false
	}
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

func normalizeAgentType(agentType string) string {
	trim := agentType
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
	switch provider {
	case "gemini":
		return "google"
	case "minimax":
		return "minimax-portal"
	case "kimi":
		return "moonshot"
	default:
		return provider
	}
}

func poolModelPrefix(provider string) string {
	target := provider
	if definitions, ok := providerDefinitions()[target]; ok && len(definitions.Models) > 0 {
		parts := strings.SplitN(strings.TrimSpace(definitions.Models[0].ID), "/", 2)
		if len(parts) == 2 && strings.TrimSpace(parts[0]) != "" {
			return parts[0]
		}
	}
	return target
}

func supportedProviderModelPrefixes(provider string) []string {
	values := []string{poolModelPrefix(provider), runtimeProviderModelPrefix(provider)}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		target := value
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

func normalizeAPIType(apiType string) string {
	trim := apiType
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
	provider = provider
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
