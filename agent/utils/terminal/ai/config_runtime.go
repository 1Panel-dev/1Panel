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
		Provider: provider,
		BaseURL:  baseURL,
		APIKey:   strings.TrimSpace(apiKey),
		Model:    model,
	}, 30 * time.Second, nil
}

func lookupProviderAPIKey(provider string) string {
	envKey := providercatalog.EnvKey(provider)
	if envKey == "" {
		return strings.TrimSpace(os.Getenv("LLM_API_KEY"))
	}
	return strings.TrimSpace(os.Getenv(envKey))
}

func defaultModelForProvider(provider string) string {
	meta, ok := providercatalog.Get(provider)
	if !ok || len(meta.Models) == 0 {
		if strings.EqualFold(strings.TrimSpace(provider), "deepseek") {
			return "deepseek-chat"
		}
		return ""
	}
	return meta.Models[0].ID
}

func ResolveGeneratorConfigFromCoreSettings() (GeneratorConfig, uint, time.Duration, error) {
	status, err := loadCoreSettingValue("AIStatus")
	if err != nil && !os.IsNotExist(err) {
		return GeneratorConfig{}, 0, 0, err
	}
	if !strings.EqualFold(strings.TrimSpace(status), "Enable") {
		return GeneratorConfig{}, 0, 0, os.ErrNotExist
	}
	accountValue, err := loadCoreSettingValue("AIAccountID")
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
	config, accountID, timeout, err := ResolveGeneratorConfigFromCoreSettings()
	if err != nil {
		return TerminalRuntimeSettings{}, GeneratorConfig{}, 0, err
	}
	prefix, err := loadCoreSettingValue("AIPrefix")
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
	if accountID > 0 {
		account, err := agentAccountRepo.GetFirst(repo.WithByID(accountID))
		if err != nil {
			return nil, err
		}
		if !account.Verified && !providercatalog.SkipVerification(account.Provider) {
			return nil, fmt.Errorf("agent account %d is not verified", account.ID)
		}
		return account, nil
	}

	accounts, err := agentAccountRepo.List(repo.WithOrderDesc("created_at"))
	if err != nil {
		return nil, err
	}
	account := selectAgentAccount(accounts)
	if account == nil {
		return nil, os.ErrNotExist
	}
	return account, nil
}

func selectAgentAccount(accounts []model.AgentAccount) *model.AgentAccount {
	for idx := range accounts {
		account := accounts[idx]
		if strings.TrimSpace(account.Provider) == "" {
			continue
		}
		if !account.Verified && !providercatalog.SkipVerification(account.Provider) {
			continue
		}
		if strings.TrimSpace(account.APIKey) == "" && !strings.EqualFold(strings.TrimSpace(account.Provider), "ollama") {
			continue
		}
		return &accounts[idx]
	}
	return nil
}

func loadCoreSettingValue(key string) (string, error) {
	var setting model.Setting
	if err := global.CoreDB.Where("key = ?", key).First(&setting).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", os.ErrNotExist
		}
		return "", err
	}
	return setting.Value, nil
}

func loadRiskCommands() ([]string, error) {
	value, err := loadCoreSettingValue("AIRiskCommands")
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
