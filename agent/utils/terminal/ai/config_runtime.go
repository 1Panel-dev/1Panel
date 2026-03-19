package ai

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

var agentAccountRepo = repo.NewIAgentAccountRepo()

type TerminalRuntimeSettings struct {
	AccountID    uint
	Prefix       string
	RiskCommands []string
}

var defaultRiskCommands = []string{
	"rm -rf",
	"mkfs",
	"dd if=",
	"curl | sh",
	"wget | sh",
	"chmod -R 777 /",
	"shutdown",
	"reboot",
	"poweroff",
	"init 0",
	":(){ :|:& };:",
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
	model := strings.TrimSpace(account.Model)
	if model == "" {
		model = defaultModelForProvider(provider)
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
		MaxTokens: account.MaxTokens,
	}, 30 * time.Second, nil
}

func lookupProviderAPIKey(provider string) string {
	envKey := providercatalog.EnvKey(provider)
	if envKey == "" {
		return ""
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

func defaultModelForProvider(provider string) string {
	meta, ok := providercatalog.Get(provider)
	if !ok || len(meta.Models) == 0 {
		return ""
	}
	return meta.Models[0].ID
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

func LoadTerminalRuntimeSettings() (TerminalRuntimeSettings, GeneratorConfig, time.Duration, error) {
	config, accountID, timeout, err := ResolveGeneratorConfigFromAgentSettings()
	if err != nil {
		return TerminalRuntimeSettings{}, GeneratorConfig{}, 0, err
	}
	prefix, err := loadAgentSettingValue("AIPrefix")
	if err != nil && !os.IsNotExist(err) {
		return TerminalRuntimeSettings{}, GeneratorConfig{}, 0, err
	}
	if strings.TrimSpace(prefix) == "" {
		prefix = "#"
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
			return append([]string(nil), defaultRiskCommands...), nil
		}
		return nil, err
	}
	if strings.TrimSpace(value) == "" {
		return append([]string(nil), defaultRiskCommands...), nil
	}
	var commands []string
	if err := json.Unmarshal([]byte(value), &commands); err != nil {
		return nil, err
	}
	return normalizeRiskCommands(commands), nil
}

func normalizeRiskCommands(commands []string) []string {
	if len(commands) == 0 {
		return append([]string(nil), defaultRiskCommands...)
	}
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
	if len(result) == 0 {
		return append([]string(nil), defaultRiskCommands...)
	}
	return result
}
