package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

var agentAccountRepo = repo.NewIAgentAccountRepo()
var agentAccountModelRepo = repo.NewIAgentAccountModelRepo()
var terminalRuntimeVersion atomic.Uint64
var fileAIRuntimeVersion atomic.Uint64

type TerminalRuntimeSettings struct {
	AccountID    uint
	Prefix       string
	RiskCommands []string
}

func CurrentTerminalRuntimeVersion() uint64 {
	return terminalRuntimeVersion.Load()
}

func InvalidateTerminalRuntimeCache() {
	terminalRuntimeVersion.Add(1)
}

func InvalidateFileAIRuntimeCache() {
	fileAIRuntimeVersion.Add(1)
}

func ResolveGeneratorConfig(accountID uint) (GeneratorConfig, time.Duration, error) {
	account, err := loadAgentAccount(accountID)
	if err != nil {
		return GeneratorConfig{}, 0, err
	}

	provider := strings.ToLower(strings.TrimSpace(account.Provider))
	if provider == "" {
		return GeneratorConfig{}, 0, fmt.Errorf("agent account provider is required")
	}
	model, maxTokens, err := resolveAccountModelConfig(account.ID, provider)
	if err != nil {
		return GeneratorConfig{}, 0, err
	}
	baseURL := strings.TrimSpace(account.BaseURL)
	if baseURL == "" {
		if defaultURL, ok := providercatalog.DefaultBaseURL(provider); ok {
			baseURL = defaultURL
		}
	}
	apiKey := strings.TrimSpace(account.APIKey)
	if apiKey == "" {
		apiKey = lookupProviderAPIKey(provider)
	}
	if apiKey == "" && provider != "ollama" {
		return GeneratorConfig{}, 0, fmt.Errorf("agent account api key is required")
	}
	return GeneratorConfig{
		Provider:  provider,
		BaseURL:   baseURL,
		APIKey:    strings.TrimSpace(apiKey),
		Model:     model,
		APIType:   strings.TrimSpace(account.APIType),
		MaxTokens: maxTokens,
	}, 30 * time.Second, nil
}

func lookupProviderAPIKey(provider string) string {
	envKey := providercatalog.EnvKey(provider)
	if envKey == "" {
		return ""
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

func defaultModelForProvider(provider string) (string, int) {
	meta, ok := providercatalog.Get(provider)
	if !ok || len(meta.Models) == 0 {
		return "", 0
	}
	return meta.Models[0].ID, meta.Models[0].MaxTokens
}

func resolveAccountModelConfig(accountID uint, provider string) (string, int, error) {
	if accountID > 0 {
		rows, err := agentAccountModelRepo.List(repo.WithByAccountID(accountID), repo.WithOrderAsc("sort_order"), repo.WithOrderAsc("id"))
		if err != nil {
			return "", 0, err
		}
		if len(rows) > 0 {
			return strings.TrimSpace(rows[0].Model), rows[0].MaxTokens, nil
		}
	}
	model, maxTokens := defaultModelForProvider(provider)
	return model, maxTokens, nil
}

func ResolveGeneratorConfigFromAgentSettings() (GeneratorConfig, uint, time.Duration, error) {
	status, err := loadAgentSettingValue("AIStatus")
	if err != nil && !os.IsNotExist(err) {
		return GeneratorConfig{}, 0, 0, err
	}
	if !strings.EqualFold(strings.TrimSpace(status), "Enable") {
		return GeneratorConfig{}, 0, 0, os.ErrNotExist
	}
	accountValue, err := loadAgentSettingValue("AIAccountID")
	if err != nil {
		return GeneratorConfig{}, 0, 0, err
	}
	accountID, err := strconv.ParseUint(strings.TrimSpace(accountValue), 10, 64)
	if err != nil || accountID == 0 {
		return GeneratorConfig{}, 0, 0, os.ErrNotExist
	}
	config, timeout, err := ResolveGeneratorConfig(uint(accountID))
	return config, uint(accountID), timeout, err
}

// ResolveGeneratorConfigFromFileSettings loads file-management AI search settings (independent from terminal AI).
func ResolveGeneratorConfigFromFileSettings() (GeneratorConfig, uint, time.Duration, error) {
	status, err := loadAgentSettingValue("FileAIStatus")
	if err != nil {
		if os.IsNotExist(err) {
			return GeneratorConfig{}, 0, 0, os.ErrNotExist
		}
		return GeneratorConfig{}, 0, 0, err
	}
	if !strings.EqualFold(strings.TrimSpace(status), "Enable") {
		return GeneratorConfig{}, 0, 0, os.ErrNotExist
	}
	accountValue, err := loadAgentSettingValue("FileAIAccountID")
	if err != nil {
		return GeneratorConfig{}, 0, 0, err
	}
	accountID, err := strconv.ParseUint(strings.TrimSpace(accountValue), 10, 64)
	if err != nil || accountID == 0 {
		return GeneratorConfig{}, 0, 0, os.ErrNotExist
	}
	config, timeout, err := ResolveGeneratorConfig(uint(accountID))
	return config, uint(accountID), timeout, err
}

// LoadFileAIRuntimeConfig returns LLM client config for file-management AI search.
func LoadFileAIRuntimeConfig() (GeneratorConfig, time.Duration, error) {
	cfg, _, timeout, err := ResolveGeneratorConfigFromFileSettings()
	return cfg, timeout, err
}

func LoadTerminalRuntimeSettings() (TerminalRuntimeSettings, GeneratorConfig, time.Duration, error) {
	config, accountID, timeout, err := ResolveGeneratorConfigFromAgentSettings()
	if err != nil {
		return TerminalRuntimeSettings{}, GeneratorConfig{}, 0, err
	}
	prefix, err := loadAgentSettingValue("AIPrefix")
	if err != nil && !os.IsNotExist(err) {
		return TerminalRuntimeSettings{}, GeneratorConfig{}, 0, err
	}
	riskCommands, err := loadRiskCommands()
	if err != nil {
		return TerminalRuntimeSettings{}, GeneratorConfig{}, 0, err
	}
	return TerminalRuntimeSettings{
		AccountID:    accountID,
		Prefix:       strings.TrimSpace(prefix),
		RiskCommands: riskCommands,
	}, config, timeout, nil
}

func loadAgentAccount(accountID uint) (*model.AgentAccount, error) {
	if accountID == 0 {
		return nil, os.ErrNotExist
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByID(accountID))
	if err != nil {
		return nil, err
	}
	return account, nil
}

func loadAgentSettingValue(key string) (string, error) {
	var setting model.Setting
	if err := global.DB.Where("key = ?", key).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", os.ErrNotExist
		}
		return "", err
	}
	return setting.Value, nil
}

func loadRiskCommands() ([]string, error) {
	value, err := loadAgentSettingValue("AIRiskCommands")
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return []string{}, nil
	}
	var commands []string
	if err := json.Unmarshal([]byte(value), &commands); err != nil {
		return nil, err
	}
	return normalizeRiskCommands(commands), nil
}

func normalizeRiskCommands(commands []string) []string {
	seen := make(map[string]struct{}, len(commands))
	result := make([]string, 0, len(commands))
	for _, command := range commands {
		command = strings.TrimSpace(command)
		if command == "" {
			continue
		}
		if _, ok := seen[command]; ok {
			continue
		}
		seen[command] = struct{}{}
		result = append(result, command)
	}
	return result
}
