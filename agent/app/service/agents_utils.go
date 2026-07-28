package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
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
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	agentenv "github.com/1Panel-dev/1Panel/agent/utils/env"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"github.com/joho/godotenv"
	"gorm.io/gorm"
)

type AgentService struct{}

func NewIAgentService() IAgentService {
	return &AgentService{}
}

type resolvedAgentAccountInput struct {
	Provider string
	APIKey   string
	BaseURL  string
	APIType  string
	AuthMode string
}

func ensureAgentAccountNameAvailable(provider, name string, excludeID uint) error {
	opts := []repo.DBOption{repo.WithByProvider(provider), repo.WithByName(name)}
	if excludeID > 0 {
		opts = append(opts, repo.WithByNOTID(excludeID))
	}
	account, err := agentAccountRepo.GetFirst(opts...)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if account != nil && account.ID > 0 {
		return buserr.New("ErrRecordExist")
	}
	return nil
}

func agentAccountUsedBySetting(tx *gorm.DB, accountID uint, statusKey, accountIDKey string) (bool, error) {
	var settings []model.Setting
	if err := tx.Where("key IN ?", []string{statusKey, accountIDKey}).Find(&settings).Error; err != nil {
		return false, err
	}
	values := make(map[string]string, len(settings))
	for _, setting := range settings {
		values[setting.Key] = setting.Value
	}
	return strings.EqualFold(strings.TrimSpace(values[statusKey]), constant.StatusEnable) &&
		strings.TrimSpace(values[accountIDKey]) == strconv.FormatUint(uint64(accountID), 10), nil
}

func loadOpenclawAgentByID(agentID uint) (*model.Agent, error) {
	agent, err := agentRepo.GetFirst(repo.WithByID(agentID))
	if err != nil {
		return nil, err
	}
	if agent.AgentType != constant.AppOpenclaw {
		return nil, fmt.Errorf("%s does not support", agent.AgentType)
	}
	return agent, nil
}

func ensureContainerRunning(containerName string) error {
	status, err := checkContainerStatus(containerName)
	if err != nil {
		return err
	}
	if status != "running" {
		return fmt.Errorf("container %s is not running, please check and retry", containerName)
	}
	return nil
}

func resolveAgentAccountInput(provider, apiType, authMode, apiKey, baseURL, modelID string, validateAvailability bool) (resolvedAgentAccountInput, error) {
	resolvedAPIKey := strings.TrimSpace(apiKey)
	resolvedAPIType := strings.TrimSpace(apiType)
	resolvedAuthMode, err := providercatalog.ResolveAuthMode(provider, resolvedAPIType, authMode)
	if err != nil {
		return resolvedAgentAccountInput{}, err
	}
	resolvedBaseURL, err := providercatalog.ResolveBaseURL(provider, resolvedAPIType, baseURL)
	if err != nil {
		if strings.Contains(err.Error(), "base url is required") {
			return resolvedAgentAccountInput{}, buserr.New("ErrAgentBaseURLRequired")
		}
		return resolvedAgentAccountInput{}, err
	}
	modelID = strings.TrimSpace(modelID)
	if modelID == "" {
		return resolvedAgentAccountInput{}, buserr.New("ErrAgentAccountModelsRequired")
	}
	imageAPI := providercatalog.IsImageAPIType(resolvedAPIType)
	if validateAvailability && (imageAPI || !providercatalog.SkipVerification(provider)) {
		if err := providercatalog.VerifyAccount(provider, resolvedAPIType, resolvedAuthMode, resolvedBaseURL, resolvedAPIKey, modelID); err != nil {
			return resolvedAgentAccountInput{}, err
		}
	}
	return resolvedAgentAccountInput{
		Provider: provider,
		APIKey:   resolvedAPIKey,
		BaseURL:  resolvedBaseURL,
		APIType:  resolvedAPIType,
		AuthMode: resolvedAuthMode,
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

func extractOtherConfig(conf map[string]interface{}) dto.AgentOtherConfig {
	result := dto.AgentOtherConfig{
		UserTimezone:   resolveServerTimezone(),
		BrowserEnabled: true,
		NPMRegistry:    defaultOpenclawNPMRegistry,
	}
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
	agentType := agent.AgentType
	item := dto.AgentItem{
		ID:           agent.ID,
		Name:         agent.Name,
		Remark:       agent.Remark,
		AgentType:    agentType,
		Provider:     agent.Provider,
		ProviderName: localizedAgentProviderName(agent.Provider),
		Model:        agent.Model,
		APIType:      agent.APIType,
		BaseURL:      agent.BaseURL,
		APIKey:       maskKey(agent.APIKey),
		Token:        agent.Token,
		Status:       agent.Status,
		Message:      agent.Message,
		AppInstallID: agent.AppInstallID,
		WebsiteID:    agent.WebsiteID,
		AccountID:    agent.AccountID,
		ConfigPath:   agent.ConfigPath,
		CreatedAt:    agent.CreatedAt,
	}
	if appInstall != nil && appInstall.ID > 0 {
		item.Container = appInstall.ContainerName
		item.AppVersion = appInstall.Version
		if agentType == constant.AppOpenclaw {
			if isOpenclawHTTPSWindowVersion(appInstall.Version) {
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
		if agentType == constant.AppHermesAgent {
			auth := readHermesDashboardAuthFromInstall(appInstall)
			item.DashboardUsername = auth.Username
			item.DashboardPassword = auth.Password
		}
	}
	return item
}

func localizedAgentProviderName(provider string) string {
	if key := providercatalog.DisplayNameKey(provider); key != "" {
		if name := strings.TrimSpace(i18n.GetMsgByKey(key)); name != "" {
			return name
		}
	}
	return providercatalog.DisplayName(provider)
}

func isAgentAppKey(appKey string) bool {
	return appKey == constant.AppOpenclaw || appKey == constant.AppCopaw || appKey == constant.AppHermesAgent
}

func isOpenclawLegacyHTTPVersion(version string) bool {
	return !common.CompareAppVersion(version, openclawHTTPSVersion)
}

func isOpenclawHTTPSWindowVersion(version string) bool {
	return common.CompareAppVersion(version, openclawHTTPSVersion) && !common.CompareAppVersion(version, openclawHTTPVersion)
}

func isOpenclawCurrentHTTPVersion(version string) bool {
	return common.CompareAppVersion(version, openclawHTTPVersion)
}

func shouldMigrateOpenclawHTTPSUpgrade(install *model.AppInstall, fromVersion, toVersion string) bool {
	if install == nil || install.App.Key != constant.AppOpenclaw {
		return false
	}
	return isOpenclawLegacyHTTPVersion(fromVersion) && isOpenclawHTTPSWindowVersion(toVersion)
}

func shouldMigrateOpenclawHTTPUpgrade(install *model.AppInstall, fromVersion, toVersion string) bool {
	if install == nil || install.App.Key != constant.AppOpenclaw {
		return false
	}
	return !isOpenclawCurrentHTTPVersion(fromVersion) && isOpenclawCurrentHTTPVersion(toVersion)
}

func migrateOpenclawProtocolUpgrade(install *model.AppInstall, fromVersion, toVersion string) error {
	systemIP, _ := settingRepo.GetValueByKey("SystemIP")
	return migrateOpenclawProtocolUpgradeWithSystemIP(install, fromVersion, toVersion, systemIP)
}

func migrateOpenclawProtocolUpgradeWithSystemIP(install *model.AppInstall, fromVersion, toVersion, systemIP string) error {
	if shouldMigrateOpenclawHTTPSUpgrade(install, fromVersion, toVersion) {
		return applyOpenclawProtocolUpgradeWithSystemIP(install, toVersion, systemIP, true)
	}
	if shouldMigrateOpenclawHTTPUpgrade(install, fromVersion, toVersion) {
		return applyOpenclawProtocolUpgradeWithSystemIP(install, toVersion, systemIP, false)
	}
	return nil
}

func applyOpenclawProtocolUpgradeWithSystemIP(install *model.AppInstall, toVersion, systemIP string, useHTTPS bool) error {
	if useHTTPS {
		migrateOpenclawInstallPortsToHTTPS(install)
	} else {
		migrateOpenclawInstallPortsToHTTP(install)
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
	port := install.HttpPort
	if useHTTPS {
		port = install.HttpsPort
	}
	if port > 0 {
		allowedOrigin, err := buildOpenclawAllowedOrigin(openclawAllowedOriginScheme(toVersion), originHost, port)
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

func openclawAllowedOriginScheme(version string) string {
	if isOpenclawHTTPSWindowVersion(version) {
		return "https"
	}
	return "http"
}

func migrateOpenclawInstallPortsToHTTPS(install *model.AppInstall) {
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

func migrateOpenclawInstallPortsToHTTP(install *model.AppInstall) {
	if install == nil {
		return
	}
	if install.HttpPort == 0 && install.HttpsPort > 0 {
		install.HttpPort = install.HttpsPort
	}
	if install.HttpsPort > 0 {
		install.HttpsPort = 0
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
	} else {
		delete(envMap, "PANEL_APP_PORT_HTTPS")
	}
	if install.HttpPort > 0 {
		envMap["PANEL_APP_PORT_HTTP"] = install.HttpPort
	}
	if install.HttpPort == 0 {
		delete(envMap, "PANEL_APP_PORT_HTTP")
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

func buildOpenclawAllowedOrigin(scheme, host string, port int) (string, error) {
	scheme = strings.TrimSpace(strings.ToLower(scheme))
	host = strings.TrimSpace(host)
	if (scheme != "http" && scheme != "https") || host == "" || port <= 0 {
		return "", fmt.Errorf("invalid openclaw allowed origin")
	}
	if strings.Contains(host, ":") && !strings.HasPrefix(host, "[") && strings.Count(host, ":") > 1 {
		host = "[" + host + "]"
	}
	return normalizeAllowedOrigin(fmt.Sprintf("%s://%s:%d", scheme, host, port))
}

func checkAgentUpgradable(install model.AppInstall) bool {
	if install.ID == 0 || install.Version == "" {
		return false
	}
	if install.App.ID == 0 {
		return false
	}
	if ignoreUpdate(install) {
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
	Primary   string   `json:"primary"`
	Fallbacks []string `json:"fallbacks,omitempty"`
}

type modelsConfig struct {
	Mode      string                   `json:"mode,omitempty"`
	Providers map[string]modelProvider `json:"providers,omitempty"`
}

type modelProvider struct {
	ApiKey     string       `json:"apiKey,omitempty"`
	BaseUrl    string       `json:"baseUrl,omitempty"`
	Api        string       `json:"api,omitempty"`
	AuthHeader bool         `json:"authHeader,omitempty"`
	Models     []modelEntry `json:"models,omitempty"`
}

type modelEntry struct {
	ID    string   `json:"id"`
	Name  string   `json:"name"`
	Input []string `json:"input,omitempty"`
}

func requiresOpenclawProviderModels(provider string) bool {
	return provider != "gemini"
}

func applyOpenclawModelsConfig(conf map[string]interface{}, models *modelsConfig) error {
	if models == nil {
		delete(conf, "models")
		return nil
	}
	modelsMap, err := structToMap(models)
	if err != nil {
		return err
	}
	conf["models"] = modelsMap
	return nil
}

type browserConfig struct {
	Enabled        bool   `json:"enabled"`
	ExecutablePath string `json:"executablePath"`
	Headless       bool   `json:"headless"`
	NoSandbox      bool   `json:"noSandbox"`
	DefaultProfile string `json:"defaultProfile"`
}

func writeOpenclawConfig(confDir string, account *model.AgentAccount, modelName, token string, allowedOrigins []string, fallbacks []string) error {
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
	resolvedFallbacks, err := resolveOpenclawFallbackModels(account, modelName, fallbacks)
	if err != nil {
		return err
	}

	cfg := openclawConfig{
		Gateway: gatewayConfig{
			Mode: "local",
			Bind: "lan",
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
				Model:        modelRef{Primary: primaryModel, Fallbacks: resolvedFallbacks},
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
		if err := applyOpenclawModelsConfig(conf, cfg.Models); err != nil {
			return err
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
		if len(cfg.Agents.Defaults.Model.Fallbacks) > 0 {
			modelMap["fallbacks"] = cfg.Agents.Defaults.Model.Fallbacks
		} else {
			delete(modelMap, "fallbacks")
		}
		if cfg.Agents.Defaults.Models != nil {
			defaultsMap["models"] = cfg.Agents.Defaults.Models
		}

		ensureGatewaySecurityDefaults(conf)
		gatewayMap := ensureChildMap(conf, "gateway")
		if _, ok := gatewayMap["mode"]; !ok {
			gatewayMap["mode"] = "local"
		}
		if _, ok := gatewayMap["bind"]; !ok {
			gatewayMap["bind"] = "lan"
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
	envMap := map[string]string{
		"OPENCLAW_GATEWAY_TOKEN": token,
	}
	order := []string{"OPENCLAW_GATEWAY_TOKEN"}
	if envKey := providercatalog.EnvKey(account.Provider); envKey != "" && account.APIKey != "" {
		envMap[envKey] = account.APIKey
		order = append(order, envKey)
	}
	return writeAgentEnvMap(path.Join(confDir, ".env"), envMap, order)
}

func resolveOpenclawFallbackModels(account *model.AgentAccount, primaryModel string, fallbackIDs []string) ([]string, error) {
	accountModels, err := loadAgentAccountModels(account)
	if err != nil {
		return nil, err
	}
	return resolveOpenclawFallbackModelsFromPool(account, accountModels, primaryModel, fallbackIDs, true)
}

func resolveOpenclawFallbackModelsFromPool(account *model.AgentAccount, accountModels []dto.AgentAccountModel, primaryModel string, fallbackIDs []string, strict bool) ([]string, error) {
	if len(fallbackIDs) == 0 {
		return nil, nil
	}
	result := make([]string, 0, len(fallbackIDs))
	seen := make(map[string]struct{}, len(fallbackIDs))
	for _, item := range fallbackIDs {
		target := strings.TrimSpace(item)
		if target == "" {
			continue
		}
		accountModel, err := requireAgentAccountModelForProvider(account.Provider, accountModels, target)
		if err != nil {
			if strict {
				return nil, err
			}
			continue
		}
		if sameProviderModelID(account.Provider, accountModel.ID, primaryModel) {
			if strict {
				return nil, buserr.New("ErrAgentFallbackModelPrimary")
			}
			continue
		}
		resolvedPrimary, _, _, _, err := buildOpenclawAccountModelConfig(account, accountModel)
		if err != nil {
			return nil, err
		}
		if _, ok := seen[resolvedPrimary]; ok {
			if strict {
				return nil, buserr.New("ErrAgentFallbackModelDuplicate")
			}
			continue
		}
		seen[resolvedPrimary] = struct{}{}
		result = append(result, resolvedPrimary)
	}
	return result, nil
}

func extractOpenclawPrimaryModelID(conf map[string]interface{}, account *model.AgentAccount, accountModels []dto.AgentAccountModel) string {
	modelMap := openclawModelRefMap(conf)
	primary := strings.TrimSpace(fmt.Sprintf("%v", modelMap["primary"]))
	if primary == "" {
		return ""
	}
	resolved, err := buildOpenclawResolvedModelIDMap(account, accountModels)
	if err != nil {
		return ""
	}
	return resolved[primary]
}

func extractOpenclawFallbackModelIDs(conf map[string]interface{}, account *model.AgentAccount, accountModels []dto.AgentAccountModel, primaryModel string) []string {
	modelMap := openclawModelRefMap(conf)
	raw, ok := modelMap["fallbacks"].([]interface{})
	if !ok || len(raw) == 0 {
		return nil
	}
	resolved, err := buildOpenclawResolvedModelIDMap(account, accountModels)
	if err != nil {
		return nil
	}
	result := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		target := strings.TrimSpace(fmt.Sprintf("%v", item))
		modelID, ok := resolved[target]
		if !ok || sameProviderModelID(account.Provider, modelID, primaryModel) {
			continue
		}
		if _, ok := seen[modelID]; ok {
			continue
		}
		seen[modelID] = struct{}{}
		result = append(result, modelID)
	}
	return result
}

func openclawModelRefMap(conf map[string]interface{}) map[string]interface{} {
	agentsMap, ok := conf["agents"].(map[string]interface{})
	if !ok {
		return nil
	}
	defaultsMap, ok := agentsMap["defaults"].(map[string]interface{})
	if !ok {
		return nil
	}
	modelMap, ok := defaultsMap["model"].(map[string]interface{})
	if !ok {
		return nil
	}
	return modelMap
}

func buildOpenclawResolvedModelIDMap(account *model.AgentAccount, accountModels []dto.AgentAccountModel) (map[string]string, error) {
	result := make(map[string]string, len(accountModels))
	for _, item := range accountModels {
		resolvedPrimary, _, _, _, err := buildOpenclawAccountModelConfig(account, item)
		if err != nil {
			return nil, err
		}
		result[resolvedPrimary] = item.ID
	}
	return result, nil
}

func prepareOpenclawInstallFiles(appInstall *model.AppInstall, account *model.AgentAccount, modelName, token string, allowedOrigins []string) error {
	if appInstall == nil {
		return fmt.Errorf("app install is required")
	}
	confDir := path.Join(appInstall.GetPath(), "data", "conf")
	if err := writeOpenclawConfig(confDir, account, modelName, token, allowedOrigins, nil); err != nil {
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
		return "", nil, nil, buserr.New("ErrAgentModelNotInAccount")
	}
	if selectedModel == "" {
		return "", nil, nil, buserr.New("ErrAgentModelNotInAccount")
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
		resolvedPrimary, entry, key, baseCfg, err := buildOpenclawAccountModelConfig(account, item)
		if err != nil {
			return "", nil, nil, err
		}
		if providerKey == "" {
			providerKey = key
			providerCfg.ApiKey = baseCfg.ApiKey
			providerCfg.BaseUrl = baseCfg.BaseUrl
			providerCfg.Api = baseCfg.Api
			providerCfg.AuthHeader = baseCfg.AuthHeader
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
	if !requiresOpenclawProviderModels(account.Provider) {
		return primaryModel, defaultsModels, nil, nil
	}
	providerCfg.Models = entries
	return primaryModel, defaultsModels, &modelsConfig{
		Mode: "merge",
		Providers: map[string]modelProvider{
			providerKey: providerCfg,
		},
	}, nil
}

func buildOpenclawAccountModelConfig(account *model.AgentAccount, model dto.AgentAccountModel) (string, modelEntry, string, modelProvider, error) {
	providerPatch, err := providercatalog.BuildOpenClawProviderPatch(account.Provider, model.ID, account.APIType, account.AuthMode, account.BaseURL, account.APIKey)
	if err != nil {
		return "", modelEntry{}, "", modelProvider{}, err
	}
	return providerPatch.PrimaryModel, buildOpenclawModelEntry(providerPatch.ModelID, model), providerPatch.ProviderKey, modelProvider{
		ApiKey:     providerPatch.APIKey,
		BaseUrl:    providerPatch.BaseURL,
		Api:        providerPatch.APIType,
		AuthHeader: providerPatch.AuthHeader,
	}, nil
}

var openclawVisionModelPattern = regexp.MustCompile(`(?i)(\b(gpt-4o|gpt-4\.1|gpt-[5-9]|o[134])\b|\bclaude-(3|4|sonnet|opus|haiku)\b|\bgemini\b|\b(qwen[\w.-]*-?vl|qwen-vl|qwen3\.[5-9]-plus)\b|\b(kimi-k2\.(5|6)|kimi-k2\.7-code|minimax-m3)\b|\b(vision|llava|pixtral|internvl|mllama|minicpm-v|glm-4v|omni)\b|(^|[-_/])vl([-_/]|$))`)

// openclawVideoInputModelPattern matches vision models that additionally accept video input.
var openclawVideoInputModelPattern = regexp.MustCompile(`(?i)\bminimax-m3\b`)

func buildOpenclawModelEntry(modelID string, model dto.AgentAccountModel) modelEntry {
	name := strings.TrimSpace(model.Name)
	if name == "" {
		name = strings.TrimSpace(modelID)
	}
	entry := modelEntry{ID: strings.TrimSpace(modelID), Name: name}
	if openclawVisionModelPattern.MatchString(modelID) {
		entry.Input = []string{"text", "image"}
		if openclawVideoInputModelPattern.MatchString(modelID) {
			entry.Input = []string{"text", "image", "video"}
		}
	}
	return entry
}

type openclawAccountModelRuntime struct {
	StoredModel  string
	PrimaryModel string
	APIType      string
}

func buildOpenclawAccountModelRuntime(account *model.AgentAccount, model dto.AgentAccountModel) (openclawAccountModelRuntime, error) {
	primaryModel, _, _, _, err := buildOpenclawAccountModelConfig(account, model)
	if err != nil {
		return openclawAccountModelRuntime{}, err
	}
	return openclawAccountModelRuntime{
		StoredModel:  model.ID,
		PrimaryModel: primaryModel,
		APIType:      account.APIType,
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

func buildInitialAgentAccountModels(account *model.AgentAccount, requested []dto.AgentAccountModel) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, fmt.Errorf("account is required")
	}
	if account.Provider != "custom" && requiresInitialAgentAccountModels(account.Provider) && len(requested) > 1 {
		return nil, buserr.New("ErrAgentAccountSingleInitialModel")
	}
	if len(requested) > 0 {
		return normalizeAgentAccountModels(account, requested)
	}
	defaultModels := providercatalog.DefaultModels(account.Provider, account.APIType)
	if len(defaultModels) == 0 {
		if requiresInitialAgentAccountModels(account.Provider) {
			return nil, buserr.New("ErrAgentAccountModelsRequired")
		}
		return nil, nil
	}
	requested = make([]dto.AgentAccountModel, 0, len(defaultModels))
	for _, item := range defaultModels {
		requested = append(requested, dto.AgentAccountModel{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return normalizeAgentAccountModels(account, requested)
}

func buildDiscoveredAgentAccountModels(modelIDs []string) []dto.AgentAccountModel {
	models := make([]dto.AgentAccountModel, 0, len(modelIDs))
	for _, modelID := range modelIDs {
		models = append(models, dto.AgentAccountModel{
			ID:   modelID,
			Name: modelID,
		})
	}
	return models
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
		target := strings.TrimSpace(item.ID)
		if target == "" {
			continue
		}
		seen[target] = struct{}{}
	}
	for _, item := range meta.Models {
		target := strings.TrimSpace(item.ID)
		if _, ok := seen[target]; ok {
			continue
		}
		requested = append(requested, dto.AgentAccountModel{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	if len(requested) == len(existing) {
		return append([]dto.AgentAccountModel(nil), existing...), nil
	}
	return normalizeAgentAccountModels(account, requested)
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
		result = append(result, dto.AgentAccountModel{
			RecordID: row.ID,
			ID:       strings.TrimSpace(row.Model),
			Name:     strings.TrimSpace(row.Name),
		})
	}
	return result, nil
}

func replacePersistedAgentAccountModelsWithTx(tx *gorm.DB, accountID uint, models []dto.AgentAccountModel) error {
	if err := tx.Where("account_id = ?", accountID).Delete(&model.AgentAccountModel{}).Error; err != nil {
		return err
	}
	for index, item := range models {
		record := &model.AgentAccountModel{
			AccountID: accountID,
			Model:     strings.TrimSpace(item.ID),
			Name:      strings.TrimSpace(item.Name),
			SortOrder: index + 1,
		}
		if err := tx.Create(record).Error; err != nil {
			return err
		}
	}
	return nil
}

func normalizeAgentAccountModels(account *model.AgentAccount, models []dto.AgentAccountModel) ([]dto.AgentAccountModel, error) {
	requested := append([]dto.AgentAccountModel(nil), models...)
	if len(requested) == 0 {
		return nil, fmt.Errorf("model is required")
	}
	normalized := make([]dto.AgentAccountModel, 0, len(requested))
	seen := make(map[string]struct{}, len(requested))
	for _, item := range requested {
		normalizedItem, err := normalizeAgentAccountModel(account, item)
		if err != nil {
			return nil, err
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
		return nil, fmt.Errorf("model is required")
	}
	return normalized, nil
}

func normalizeAgentAccountModel(account *model.AgentAccount, model dto.AgentAccountModel) (dto.AgentAccountModel, error) {
	modelID := strings.TrimSpace(model.ID)
	if modelID == "" {
		return dto.AgentAccountModel{}, fmt.Errorf("model is required")
	}
	modelID = providercatalog.NormalizeModelID(account.Provider, modelID)
	name := strings.TrimSpace(model.Name)
	if name == "" {
		name = modelID
	}
	return dto.AgentAccountModel{
		ID:   modelID,
		Name: name,
	}, nil
}

func requiresInitialAgentAccountModels(provider string) bool {
	switch provider {
	case "custom", "vllm", "ollama":
		return true
	default:
		return false
	}
}

func sameProviderModelID(provider, left, right string) bool {
	leftTrimmed := strings.TrimSpace(left)
	rightTrimmed := strings.TrimSpace(right)
	if leftTrimmed == rightTrimmed {
		return true
	}
	leftComparable := providercatalog.NormalizeModelID(provider, leftTrimmed)
	rightComparable := providercatalog.NormalizeModelID(provider, rightTrimmed)
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

func resolveAgentAccountVerifyModel(provider, requested string, models []dto.AgentAccountModel) (string, error) {
	if len(models) == 0 {
		return "", buserr.New("ErrAgentAccountModelsRequired")
	}
	if strings.TrimSpace(requested) == "" {
		return models[0].ID, nil
	}
	selected, ok := findAgentAccountModelForProvider(provider, models, requested)
	if !ok {
		return "", buserr.New("ErrAgentModelNotInAccount")
	}
	return selected.ID, nil
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

func extractStringList(value interface{}) []string {
	switch values := value.(type) {
	case []interface{}:
		result := make([]string, 0, len(values))
		for _, value := range values {
			text := strings.TrimSpace(fmt.Sprintf("%v", value))
			if text == "" {
				continue
			}
			result = append(result, text)
		}
		return result
	default:
		return []string{}
	}
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

func readAgentEnvMap(envPath string) (map[string]string, error) {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(envPath) {
		return map[string]string{}, nil
	}
	envMap, err := godotenv.Read(envPath)
	if err != nil {
		return nil, err
	}
	return envMap, nil
}

func writeAgentEnvMap(envPath string, envMap map[string]string, order []string) error {
	if len(envMap) == 0 {
		return files.NewFileOp().SaveFile(envPath, "", 0600)
	}
	return agentenv.WriteWithOrder(envMap, envPath, order)
}

func upsertAgentEnv(envPath string, values map[string]string, order []string, overwrite bool) error {
	envMap, err := readAgentEnvMap(envPath)
	if err != nil {
		return err
	}
	for key, value := range values {
		if key == "" {
			continue
		}
		if overwrite || strings.TrimSpace(envMap[key]) == "" {
			envMap[key] = value
		}
	}
	return writeAgentEnvMap(envPath, envMap, order)
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

func generateToken() string {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func asyncReportAIProviderInstall(provider string) {
	if global.CONF.Base.Mode != "stable" || provider == "" {
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
