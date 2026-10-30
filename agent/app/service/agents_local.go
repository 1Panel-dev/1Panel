package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
	"gorm.io/gorm"
)

func SyncLocalModelAccount(install model.AppInstall, accountID uint, baseURL string, models []dto.AgentAccountModel, updateBaseURL, createIfMissing bool) (*model.AgentAccount, error) {
	if install.ID == 0 || (install.App.Key != "ollama" && install.App.Key != "vllm") {
		return nil, fmt.Errorf("invalid local model installation")
	}
	agentAccountMutationMu.Lock()
	defer agentAccountMutationMu.Unlock()
	account, err := agentAccountRepo.GetFirst(repo.WithByAppInstallID(install.ID))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) && accountID != 0 {
		account, err = agentAccountRepo.GetFirst(repo.WithByID(accountID), repo.WithByProvider(install.App.Key))
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if err == nil && account.AppInstallID != 0 && account.AppInstallID != install.ID {
			return nil, fmt.Errorf("model account belongs to another installation")
		}
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !createIfMissing {
			return nil, nil
		}
		name := install.Name
		for suffix := 1; ; suffix++ {
			existing, err := agentAccountRepo.List(repo.WithByProvider(install.App.Key), repo.WithByName(name))
			if err != nil {
				return nil, err
			}
			if len(existing) == 0 {
				break
			}
			name = fmt.Sprintf("%s-%d", install.Name, suffix)
		}
		apiType, apiKey := "openai-completions", "vllm-local"
		if install.App.Key == "ollama" {
			apiType, apiKey = "openai-responses", "ollama-local"
		}
		account = &model.AgentAccount{Name: name, Provider: install.App.Key, BaseURL: baseURL, APIType: apiType, APIKey: apiKey, RememberAPIKey: true, Verified: true}
	}
	accountChanged := account.ID == 0 || account.AppInstallID != install.ID || (updateBaseURL && account.BaseURL != baseURL)
	account.AppInstallID = install.ID
	if updateBaseURL {
		account.BaseURL = baseURL
	}
	rows := make([]model.AgentAccountModel, 0, len(models))
	verifyModelExists := account.VerifyModel == ""
	for _, item := range models {
		rows = append(rows, model.AgentAccountModel{Model: item.ID, Name: item.Name})
		if item.ID == account.VerifyModel {
			verifyModelExists = true
		}
	}
	if !verifyModelExists {
		account.VerifyModel = ""
		accountChanged = true
	}
	var previous []model.AgentAccountModel
	if account.ID != 0 {
		previous, err = agentAccountModelRepo.List(repo.WithByAccountID(account.ID), repo.WithOrderAsc("sort_order"), repo.WithOrderAsc("id"))
		if err != nil {
			return nil, err
		}
	}
	modelsChanged := len(previous) != len(rows)
	if !modelsChanged {
		for i := range rows {
			if previous[i].Model != rows[i].Model || previous[i].Name != rows[i].Name {
				modelsChanged = true
				break
			}
		}
	}
	if modelsChanged {
		err = agentAccountRepo.SaveWithModels(account, rows)
	} else if accountChanged {
		err = agentAccountRepo.Save(account)
	} else {
		return account, nil
	}
	if err != nil {
		return nil, err
	}
	terminalai.InvalidateTerminalRuntimeCache()
	terminalai.InvalidateFileAIRuntimeCache()
	return account, nil
}

func SyncOllamaModelAccount(createIfMissing bool) error {
	app, err := appRepo.GetFirst(appRepo.WithKey("ollama"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	install, err := appInstallRepo.GetFirst(appInstallRepo.WithAppId(app.ID))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !install.SyncModelAccount {
		return nil
	}
	return syncOllamaModelAccount(install, fmt.Sprintf("http://%s:11434/v1", install.ContainerName), false, createIfMissing)
}

func syncOllamaModelAccount(install model.AppInstall, baseURL string, updateBaseURL, createIfMissing bool) error {
	models, err := aiRepo.List(repo.WithByStatus(constant.StatusSuccess), repo.WithOrderAsc("name"))
	if err != nil {
		return err
	}
	items := make([]dto.AgentAccountModel, 0, len(models))
	for _, item := range models {
		items = append(items, dto.AgentAccountModel{ID: item.Name, Name: item.Name})
	}
	_, err = SyncLocalModelAccount(install, 0, baseURL, items, updateBaseURL, createIfMissing)
	return err
}

func updateOllamaModelPool(modelName string, remove bool) error {
	install, err := appInstallRepo.LoadInstallAppByKeyAndName("ollama", "")
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if !install.SyncModelAccount {
		return nil
	}
	agentAccountMutationMu.Lock()
	defer agentAccountMutationMu.Unlock()
	account, err := agentAccountRepo.GetFirst(repo.WithByAppInstallID(install.ID), repo.WithByProvider("ollama"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	models, err := loadAgentAccountModels(account)
	if err != nil {
		return err
	}
	if remove {
		for _, item := range models {
			if !sameProviderModelID("ollama", item.ID, modelName) {
				continue
			}
			if err := agentAccountModelRepo.DeleteByID(item.RecordID); err != nil {
				return err
			}
		}
		if sameProviderModelID("ollama", account.VerifyModel, modelName) {
			account.VerifyModel = ""
			if err := agentAccountRepo.Save(account); err != nil {
				return err
			}
		}
		if err := compactPersistedAgentAccountModelSortOrder(account.ID); err != nil {
			return err
		}
	} else {
		if _, exists := findAgentAccountModelForProvider("ollama", models, modelName); exists {
			return nil
		}
		record := &model.AgentAccountModel{AccountID: account.ID, Model: modelName, Name: modelName, SortOrder: len(models) + 1}
		if err := agentAccountModelRepo.Create(record); err != nil {
			return err
		}
	}
	terminalai.InvalidateTerminalRuntimeCache()
	terminalai.InvalidateFileAIRuntimeCache()
	return nil
}

func loadOllamaModelAccountAddress(install model.AppInstall) (string, string, error) {
	addressType := install.ModelAccountBaseURLType
	if addressType == "" {
		addressType = "container"
	}
	account, err := agentAccountRepo.GetFirst(repo.WithByAppInstallID(install.ID), repo.WithByProvider("ollama"))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return addressType, install.ModelAccountBaseURL, nil
	}
	if err != nil {
		return "", "", err
	}
	if addressType != "custom" {
		expectedURL, err := ResolveLocalModelAccountBaseURL(addressType, "", install.ContainerName, "11434", install.HttpPort)
		if err != nil {
			return "", "", err
		}
		if strings.TrimRight(account.BaseURL, "/") != expectedURL {
			addressType = "custom"
		}
	}
	return addressType, account.BaseURL, nil
}

func CleanupLocalModelAccount(installID uint) {
	if installID == 0 {
		return
	}
	accounts, err := agentAccountRepo.List(repo.WithByAppInstallID(installID))
	if err != nil {
		global.LOG.Errorf("load local model accounts: %v", err)
		return
	}
	for _, account := range accounts {
		if err := (AgentService{}).DeleteAccount(dto.AgentAccountDeleteReq{ID: account.ID}); err != nil {
			global.LOG.Warnf("keep local model account %d after uninstall: %v", account.ID, err)
		}
	}
}

func ResolveLocalModelAccountBaseURL(addressType, customURL, containerName, containerPort string, hostPort int) (string, error) {
	switch addressType {
	case "custom":
		baseURL := strings.TrimRight(strings.TrimSpace(customURL), "/")
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			return "", fmt.Errorf("model account base URL must start with http:// or https://")
		}
		return baseURL, nil
	case "localhost":
		return fmt.Sprintf("http://127.0.0.1:%d/v1", hostPort), nil
	case "systemIP":
		systemIP, err := settingRepo.GetValueByKey("SystemIP")
		if err != nil {
			return "", err
		}
		systemIP = strings.TrimSpace(systemIP)
		if systemIP == "" {
			systemIP = "127.0.0.1"
		}
		return fmt.Sprintf("http://%s:%d/v1", systemIP, hostPort), nil
	default:
		return fmt.Sprintf("http://%s:%s/v1", containerName, containerPort), nil
	}
}

func deleteRemoteOllamaModel(account *model.AgentAccount, modelName string) error {
	body, err := json.Marshal(map[string]string{"model": modelName})
	if err != nil {
		return err
	}
	baseURL := strings.TrimSuffix(strings.TrimRight(account.BaseURL, "/"), "/v1")
	req, err := http.NewRequest(http.MethodDelete, baseURL+"/api/delete", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+account.APIKey)
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("delete Ollama model failed: %s", resp.Status)
	}
	return nil
}

func resolveExistingOllamaModel(account *model.AgentAccount, modelID string) (string, error) {
	available, err := providercatalog.DiscoverModels(account.BaseURL, account.APIKey)
	if err != nil {
		return "", buserr.WithErr("ErrAgentAccountUnavailable", err)
	}
	for _, id := range available {
		if sameProviderModelID("ollama", id, modelID) {
			return id, nil
		}
	}
	return "", buserr.New("ErrRecordNotFound")
}
