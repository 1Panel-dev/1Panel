package service

import (
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"gopkg.in/yaml.v3"
)

const hermesWorkspaceDir = "/opt/data/workspace"
const hermesExecutablePath = "/opt/hermes/.venv/bin/hermes"
const hermesDashboardUsernameEnvKey = "HERMES_DASHBOARD_USERNAME"
const hermesDashboardPasswordEnvKey = "HERMES_DASHBOARD_PASSWORD"
const defaultHermesDashboardUsername = "admin"
const hermesDashboardPasswordLength = 8

type hermesDashboardAuth struct {
	Username string
	Password string
}

type hermesConfig struct {
	Model    hermesModelConfig    `yaml:"model"`
	Terminal hermesTerminalConfig `yaml:"terminal"`
	Timezone string               `yaml:"timezone,omitempty"`
}

type hermesModelConfig struct {
	Default  string `yaml:"default"`
	Provider string `yaml:"provider"`
	BaseURL  string `yaml:"base_url,omitempty"`
	APIKey   string `yaml:"api_key,omitempty"`
}

type hermesTerminalConfig struct {
	Backend string `yaml:"backend"`
	Cwd     string `yaml:"cwd"`
}

type hermesEnvEntry struct {
	Key   string
	Value string
}

func buildHermesDockerExecCommandArgs(containerName, command string, commandArgs ...string) []string {
	args := []string{"exec", "-u", "hermes", containerName, command}
	return append(args, commandArgs...)
}

func buildHermesDockerExecArgs(containerName string, hermesArgs ...string) []string {
	return buildHermesDockerExecCommandArgs(containerName, "hermes", hermesArgs...)
}

func writeHermesConfig(confDir string, account *model.AgentAccount, modelName string, timezone string) error {
	if strings.TrimSpace(confDir) == "" {
		return fmt.Errorf("config dir is required")
	}
	if account == nil {
		return fmt.Errorf("account is required")
	}
	if strings.TrimSpace(modelName) == "" {
		return fmt.Errorf("model is required")
	}
	fileOp := files.NewFileOp()
	if !fileOp.Stat(confDir) {
		if err := fileOp.CreateDir(confDir, constant.DirPerm); err != nil {
			return err
		}
	}

	provider := resolveHermesProvider(account.Provider)
	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	model := map[string]interface{}{
		"default":  resolveHermesModel(account.Provider, provider, modelName),
		"provider": provider,
		"base_url": account.BaseURL,
	}
	if provider == "custom" && account.APIKey != "" {
		model["api_key"] = "${CUSTOM_API_KEY}"
	}
	cfg["model"] = model
	cfg["terminal"] = map[string]interface{}{
		"backend": "local",
		"cwd":     hermesWorkspaceDir,
	}
	if timezone != "" {
		cfg["timezone"] = timezone
	} else {
		delete(cfg, "timezone")
	}
	if err := writeHermesConfigMap(configPath, cfg); err != nil {
		return err
	}
	return writeHermesModelEnv(path.Join(confDir, ".env"), account)
}

func prepareHermesInstallFiles(appInstall *model.AppInstall, account *model.AgentAccount, modelName string) error {
	if appInstall == nil {
		return fmt.Errorf("app install is required")
	}
	dataDir := path.Join(appInstall.GetPath(), "data")
	if err := writeHermesConfig(dataDir, account, modelName, normalizeHermesTimezone(common.LoadTimeZoneByCmd())); err != nil {
		return err
	}
	return files.NewFileOp().ChownR(dataDir, "1000", "1000", true)
}

func normalizeHermesDashboardAuth(username, password string) hermesDashboardAuth {
	auth := hermesDashboardAuth{
		Username: strings.TrimSpace(username),
		Password: strings.TrimSpace(password),
	}
	if auth.Username == "" {
		auth.Username = defaultHermesDashboardUsername
	}
	if auth.Password == "" {
		auth.Password = generateHermesDashboardPassword()
	}
	return auth
}

func generateHermesDashboardPassword() string {
	for i := 0; i < 100; i++ {
		value := common.RandStr(hermesDashboardPasswordLength)
		if hermesDashboardPasswordValid(value) {
			return value
		}
	}
	return "Aa1" + common.RandStr(hermesDashboardPasswordLength-3)
}

func hermesDashboardPasswordValid(value string) bool {
	if len(value) != hermesDashboardPasswordLength {
		return false
	}
	var hasUpper, hasLower, hasDigit bool
	for _, ch := range value {
		switch {
		case ch >= 'A' && ch <= 'Z':
			hasUpper = true
		case ch >= 'a' && ch <= 'z':
			hasLower = true
		case ch >= '0' && ch <= '9':
			hasDigit = true
		default:
			return false
		}
	}
	return hasUpper && hasLower && hasDigit
}

func writeHermesDashboardAuthEnv(envPath string, auth hermesDashboardAuth, overwrite bool) error {
	return upsertAgentEnv(envPath, map[string]string{
		hermesDashboardUsernameEnvKey: auth.Username,
		hermesDashboardPasswordEnvKey: auth.Password,
	}, []string{
		hermesDashboardUsernameEnvKey,
		hermesDashboardPasswordEnvKey,
	}, overwrite)
}

func readHermesDashboardAuthEnv(envPath string) (hermesDashboardAuth, error) {
	envMap, err := readAgentEnvMap(envPath)
	if err != nil {
		return hermesDashboardAuth{}, err
	}
	return hermesDashboardAuth{
		Username: strings.TrimSpace(envMap[hermesDashboardUsernameEnvKey]),
		Password: strings.TrimSpace(envMap[hermesDashboardPasswordEnvKey]),
	}, nil
}

func readHermesDashboardAuthFromInstall(appInstall *model.AppInstall) hermesDashboardAuth {
	if appInstall == nil || appInstall.ID == 0 {
		return hermesDashboardAuth{}
	}
	auth, err := readHermesDashboardAuthEnv(path.Join(appInstall.GetPath(), ".env"))
	if err != nil {
		return hermesDashboardAuth{}
	}
	return auth
}

func readHermesConfig(configPath string) (*hermesConfig, error) {
	content, err := files.NewFileOp().GetContent(configPath)
	if err != nil {
		return nil, err
	}
	var cfg hermesConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func readHermesConfigMap(configPath string) (map[string]interface{}, error) {
	fileOp := files.NewFileOp()
	if !fileOp.Stat(configPath) {
		return map[string]interface{}{}, nil
	}
	content, err := fileOp.GetContent(configPath)
	if err != nil {
		return nil, err
	}
	cfg := map[string]interface{}{}
	if strings.TrimSpace(string(content)) == "" {
		return cfg, nil
	}
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

func writeHermesConfigMap(configPath string, cfg map[string]interface{}) error {
	payload, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return files.NewFileOp().SaveFile(configPath, string(payload), 0600)
}

func readHermesTelegramChannelConfig(confDir string) (*dto.AgentTelegramConfig, error) {
	envMap, err := readAgentEnvMap(path.Join(confDir, ".env"))
	if err != nil {
		return nil, err
	}
	cfg, err := readHermesConfigMap(path.Join(confDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	allowFrom := splitHermesEnvList(envMap["TELEGRAM_ALLOWED_USERS"])
	requireMention := extractBoolValue(childMap(cfg, "telegram")["require_mention"], false)
	dmPolicy := ""
	if extractHermesEnvBool(envMap, "TELEGRAM_ALLOW_ALL_USERS", false) {
		dmPolicy = "open"
	} else if len(allowFrom) > 0 {
		dmPolicy = "allowlist"
	} else if envMap["TELEGRAM_BOT_TOKEN"] != "" {
		dmPolicy = "pairing"
	}
	groupPolicy := ""
	if requireMention {
		groupPolicy = "allowlist"
	} else if envMap["TELEGRAM_BOT_TOKEN"] != "" {
		groupPolicy = "open"
	}
	token := envMap["TELEGRAM_BOT_TOKEN"]
	result := &dto.AgentTelegramConfig{
		Enabled:        token != "",
		DmPolicy:       dmPolicy,
		AllowFrom:      allowFrom,
		RequireMention: requireMention,
		GroupPolicy:    groupPolicy,
		GroupAllowFrom: []string{},
		Streaming:      "",
		DefaultAccount: "",
	}
	if token != "" {
		result.Bots = []dto.AgentTelegramBot{
			{
				AgentChannelBotBase: dto.AgentChannelBotBase{
					AccountID: "default",
					Name:      "Default",
					Enabled:   true,
					IsDefault: true,
				},
				BotToken:    token,
				DmPolicy:    dmPolicy,
				GroupPolicy: groupPolicy,
				Streaming:   "",
			},
		}
	}
	return result, nil
}

func writeHermesTelegramChannelConfig(confDir string, config dto.AgentTelegramConfig) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readAgentEnvMap(envPath)
	if err != nil {
		return err
	}
	token := firstHermesTelegramBotToken(config.Bots, config.DefaultAccount)
	if token != "" {
		envMap["TELEGRAM_BOT_TOKEN"] = token
	} else {
		delete(envMap, "TELEGRAM_BOT_TOKEN")
	}
	delete(envMap, "TELEGRAM_ALLOWED_USERS")
	delete(envMap, "TELEGRAM_ALLOW_ALL_USERS")
	switch config.DmPolicy {
	case "open":
		envMap["TELEGRAM_ALLOW_ALL_USERS"] = "true"
	case "allowlist":
		if allow := joinHermesEnvList(config.AllowFrom); allow != "" {
			envMap["TELEGRAM_ALLOWED_USERS"] = allow
		}
	}
	if err := writeAgentEnvMap(envPath, envMap, []string{
		"TELEGRAM_BOT_TOKEN",
		"TELEGRAM_ALLOWED_USERS",
		"TELEGRAM_ALLOW_ALL_USERS",
	}); err != nil {
		return err
	}

	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	telegram := ensureChildMap(cfg, "telegram")
	telegram["require_mention"] = config.RequireMention
	return writeHermesConfigMap(configPath, cfg)
}

func readHermesDiscordChannelConfig(confDir string) (*dto.AgentDiscordConfig, error) {
	envMap, err := readAgentEnvMap(path.Join(confDir, ".env"))
	if err != nil {
		return nil, err
	}
	cfg, err := readHermesConfigMap(path.Join(confDir, "config.yaml"))
	if err != nil {
		return nil, err
	}

	allowFrom := splitHermesEnvList(envMap["DISCORD_ALLOWED_USERS"])
	requireMention := extractBoolValue(childMap(cfg, "discord")["require_mention"], false)
	dmPolicy := ""
	if extractHermesEnvBool(envMap, "DISCORD_ALLOW_ALL_USERS", false) {
		dmPolicy = "open"
	} else if len(allowFrom) > 0 {
		dmPolicy = "allowlist"
	} else if envMap["DISCORD_BOT_TOKEN"] != "" {
		dmPolicy = "pairing"
	}
	groupPolicy := ""
	if requireMention {
		groupPolicy = "allowlist"
	} else if envMap["DISCORD_BOT_TOKEN"] != "" {
		groupPolicy = "open"
	}
	token := envMap["DISCORD_BOT_TOKEN"]
	result := &dto.AgentDiscordConfig{
		Enabled:        token != "",
		DmPolicy:       dmPolicy,
		AllowFrom:      allowFrom,
		RequireMention: requireMention,
		GroupPolicy:    groupPolicy,
		DefaultAccount: "",
	}
	if token != "" {
		result.Bots = []dto.AgentDiscordBot{
			{
				AgentChannelBotBase: dto.AgentChannelBotBase{
					AccountID: "default",
					Name:      "Default",
					Enabled:   true,
					IsDefault: true,
				},
				Token: token,
			},
		}
	}
	return result, nil
}

func writeHermesDiscordChannelConfig(confDir string, config dto.AgentDiscordConfig) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readAgentEnvMap(envPath)
	if err != nil {
		return err
	}
	token := firstHermesDiscordBotToken(config.Bots, config.DefaultAccount)
	if token != "" {
		envMap["DISCORD_BOT_TOKEN"] = token
	} else {
		delete(envMap, "DISCORD_BOT_TOKEN")
	}
	delete(envMap, "DISCORD_ALLOWED_USERS")
	delete(envMap, "DISCORD_ALLOW_ALL_USERS")
	switch config.DmPolicy {
	case "open":
		envMap["DISCORD_ALLOW_ALL_USERS"] = "true"
	case "allowlist":
		if allow := joinHermesEnvList(config.AllowFrom); allow != "" {
			envMap["DISCORD_ALLOWED_USERS"] = allow
		}
	}
	if err := writeAgentEnvMap(envPath, envMap, []string{
		"DISCORD_BOT_TOKEN",
		"DISCORD_ALLOWED_USERS",
		"DISCORD_ALLOW_ALL_USERS",
	}); err != nil {
		return err
	}

	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	discord := ensureChildMap(cfg, "discord")
	discord["require_mention"] = config.RequireMention
	return writeHermesConfigMap(configPath, cfg)
}

func deleteHermesEnvKeys(confDir string, keys ...string) error {
	envPath := path.Join(confDir, ".env")
	envMap, err := readAgentEnvMap(envPath)
	if err != nil {
		return err
	}
	for _, key := range keys {
		delete(envMap, key)
	}
	return writeAgentEnvMap(envPath, envMap, keys)
}

func deleteHermesConfigSections(confDir string, topLevelKeys []string, platformKeys []string) error {
	configPath := path.Join(confDir, "config.yaml")
	cfg, err := readHermesConfigMap(configPath)
	if err != nil {
		return err
	}
	for _, key := range topLevelKeys {
		delete(cfg, key)
	}
	if len(platformKeys) > 0 {
		if platforms, ok := cfg["platforms"].(map[string]interface{}); ok {
			for _, key := range platformKeys {
				delete(platforms, key)
			}
			if len(platforms) == 0 {
				delete(cfg, "platforms")
			}
		}
	}
	return writeHermesConfigMap(configPath, cfg)
}

func deleteHermesTelegramChannelConfig(confDir string) error {
	if err := deleteHermesEnvKeys(confDir,
		"TELEGRAM_BOT_TOKEN",
		"TELEGRAM_ALLOWED_USERS",
		"TELEGRAM_ALLOW_ALL_USERS",
		"TELEGRAM_HOME_CHANNEL",
	); err != nil {
		return err
	}
	return deleteHermesConfigSections(confDir, []string{"telegram"}, []string{"telegram"})
}

func deleteHermesDiscordChannelConfig(confDir string) error {
	if err := deleteHermesEnvKeys(confDir,
		"DISCORD_BOT_TOKEN",
		"DISCORD_ALLOWED_USERS",
		"DISCORD_ALLOW_ALL_USERS",
		"DISCORD_HOME_CHANNEL",
	); err != nil {
		return err
	}
	return deleteHermesConfigSections(confDir, []string{"discord"}, []string{"discord"})
}

func normalizeHermesTimezone(timezone string) string {
	timezone = strings.TrimSpace(timezone)
	if timezone == "" {
		return ""
	}
	if _, err := time.LoadLocation(timezone); err != nil {
		return ""
	}
	return timezone
}

func resolveHermesProvider(provider string) string {
	switch provider {
	case "":
		return "custom"
	case "openrouter", "anthropic", "gemini", "zai", "kimi-coding", "xiaomi":
		return provider
	case "minimax":
		return "minimax-cn"
	default:
		return "custom"
	}
}

func resolveHermesModel(sourceProvider, targetProvider, modelName string) string {
	target := strings.TrimSpace(modelName)
	if target == "" {
		return ""
	}
	if targetProvider != "custom" {
		return target
	}
	if sourceProvider == "custom" || sourceProvider == "vllm" {
		return normalizeCustomModel(target)
	}
	if strings.Contains(target, "/") {
		parts := strings.SplitN(target, "/", 2)
		model := strings.TrimSpace(parts[1])
		if model != "" {
			return model
		}
	}
	return target
}

func resolveHermesConfiguredModelID(account *model.AgentAccount, accountModels []dto.AgentAccountModel, configuredModel string) (string, error) {
	if account == nil {
		return "", buserr.New("ErrAgentModelNotInAccount")
	}
	configuredModel = strings.TrimSpace(configuredModel)
	if configuredModel == "" {
		return "", buserr.New("ErrAgentModelNotInAccount")
	}
	provider := resolveHermesProvider(account.Provider)
	for _, item := range accountModels {
		if resolveHermesModel(account.Provider, provider, item.ID) == configuredModel {
			return item.ID, nil
		}
	}
	return "", buserr.New("ErrAgentModelNotInAccount")
}

func resolveHermesEnvEntries(account *model.AgentAccount) []hermesEnvEntry {
	if account == nil {
		return nil
	}
	if resolveHermesProvider(account.Provider) == "custom" {
		if account.APIKey == "" {
			return nil
		}
		return []hermesEnvEntry{
			{Key: "CUSTOM_API_KEY", Value: account.APIKey},
		}
	}
	apiKey := account.APIKey
	baseURL := account.BaseURL
	entries := make([]hermesEnvEntry, 0, 4)

	appendEntry := func(key, value string) {
		if key == "" || value == "" {
			return
		}
		entries = append(entries, hermesEnvEntry{Key: key, Value: value})
	}

	switch account.Provider {
	case "openrouter":
		appendEntry("OPENROUTER_API_KEY", apiKey)
		appendEntry("OPENROUTER_BASE_URL", baseURL)
	case "anthropic":
		appendEntry("ANTHROPIC_API_KEY", apiKey)
	case "gemini":
		appendEntry("GOOGLE_API_KEY", apiKey)
		appendEntry("GEMINI_API_KEY", apiKey)
		appendEntry("GEMINI_BASE_URL", baseURL)
	case "zai":
		appendEntry("GLM_API_KEY", apiKey)
		appendEntry("ZAI_API_KEY", apiKey)
		appendEntry("Z_AI_API_KEY", apiKey)
		appendEntry("GLM_BASE_URL", baseURL)
	case "kimi-coding", "moonshot":
		appendEntry("KIMI_API_KEY", apiKey)
		appendEntry("KIMI_BASE_URL", baseURL)
	case "kimi":
		appendEntry("KIMI_CN_API_KEY", apiKey)
		appendEntry("KIMI_BASE_URL", baseURL)
	case "minimax":
		appendEntry("MINIMAX_CN_API_KEY", apiKey)
		appendEntry("MINIMAX_CN_BASE_URL", baseURL)
	case "xiaomi":
		appendEntry("XIAOMI_API_KEY", apiKey)
		appendEntry("XIAOMI_BASE_URL", baseURL)
	case "deepseek":
		appendEntry("DEEPSEEK_API_KEY", apiKey)
		appendEntry("DEEPSEEK_BASE_URL", baseURL)
	case "bailian-coding-plan":
		appendEntry("DASHSCOPE_API_KEY", apiKey)
		appendEntry("DASHSCOPE_BASE_URL", baseURL)
	case "openai":
		appendEntry("OPENAI_API_KEY", apiKey)
		appendEntry("OPENAI_BASE_URL", baseURL)
	case "custom", "vllm":
		if account.APIType == "openai-completions" || account.APIType == "openai-responses" {
			appendEntry("OPENAI_API_KEY", apiKey)
			appendEntry("OPENAI_BASE_URL", baseURL)
		}
	default:
		if envKey := providercatalog.EnvKey(account.Provider); envKey != "" {
			appendEntry(envKey, apiKey)
		}
	}

	return entries
}

func writeHermesModelEnv(envPath string, account *model.AgentAccount) error {
	envMap, err := readAgentEnvMap(envPath)
	if err != nil {
		return err
	}
	for _, key := range hermesManagedModelEnvKeys() {
		delete(envMap, key)
	}
	entries := resolveHermesEnvEntries(account)
	order := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Key == "" || entry.Value == "" {
			continue
		}
		envMap[entry.Key] = entry.Value
		order = append(order, entry.Key)
	}
	return writeAgentEnvMap(envPath, envMap, order)
}

func hermesManagedModelEnvKeys() []string {
	keys := []string{
		"OPENROUTER_API_KEY",
		"OPENROUTER_BASE_URL",
		"ANTHROPIC_API_KEY",
		"GOOGLE_API_KEY",
		"GEMINI_API_KEY",
		"GEMINI_BASE_URL",
		"GLM_API_KEY",
		"ZAI_API_KEY",
		"Z_AI_API_KEY",
		"GLM_BASE_URL",
		"KIMI_API_KEY",
		"KIMI_CN_API_KEY",
		"KIMI_BASE_URL",
		"MINIMAX_CN_API_KEY",
		"MINIMAX_CN_BASE_URL",
		"XIAOMI_API_KEY",
		"XIAOMI_BASE_URL",
		"DEEPSEEK_API_KEY",
		"DEEPSEEK_BASE_URL",
		"DASHSCOPE_API_KEY",
		"DASHSCOPE_BASE_URL",
		"OPENAI_API_KEY",
		"OPENAI_BASE_URL",
	}
	seen := make(map[string]struct{}, len(keys))
	result := make([]string, 0, len(keys))
	appendKey := func(key string) {
		if key == "" {
			return
		}
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		result = append(result, key)
	}
	for _, key := range keys {
		appendKey(key)
	}
	for _, meta := range providercatalog.All() {
		appendKey(meta.EnvKey)
	}
	return result
}

func writeHermesEnv(envPath string, entries []hermesEnvEntry) error {
	values := map[string]string{}
	for _, entry := range entries {
		if entry.Key == "" || entry.Value == "" {
			continue
		}
		values[entry.Key] = entry.Value
	}
	order := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.Key == "" {
			continue
		}
		order = append(order, entry.Key)
	}
	return upsertAgentEnv(envPath, values, order, true)
}

func splitHermesEnvList(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func joinHermesEnvList(values []string) string {
	if len(values) == 0 {
		return ""
	}
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		item := strings.TrimSpace(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return strings.Join(result, ",")
}

func firstHermesTelegramBotToken(bots []dto.AgentTelegramBot, defaultAccount string) string {
	for _, bot := range bots {
		if bot.AccountID == defaultAccount || bot.IsDefault {
			return bot.BotToken
		}
	}
	if len(bots) == 0 {
		return ""
	}
	return bots[0].BotToken
}

func firstHermesDiscordBotToken(bots []dto.AgentDiscordBot, defaultAccount string) string {
	for _, bot := range bots {
		if bot.AccountID == defaultAccount || bot.IsDefault {
			return bot.Token
		}
	}
	if len(bots) == 0 {
		return ""
	}
	return bots[0].Token
}

func extractHermesEnvBool(envMap map[string]string, key string, defaultValue bool) bool {
	value, ok := envMap[key]
	if !ok || value == "" {
		return defaultValue
	}
	return strings.EqualFold(value, "true")
}

func validateHermesPairingApproveResult(output string, err error) error {
	if strings.Contains(output, "Approved!") {
		return nil
	}
	if strings.Contains(output, "not found or expired for platform") {
		return buserr.New("ErrHermesPairingCodeUnavailable")
	}
	if err == nil {
		if strings.TrimSpace(output) == "" {
			return fmt.Errorf("unexpected hermes pairing approve result")
		}
		return errors.New(strings.TrimSpace(output))
	}
	if strings.Contains(err.Error(), "not found or expired for platform") {
		return buserr.New("ErrHermesPairingCodeUnavailable")
	}
	return err
}
