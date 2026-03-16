package service

import (
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

func TestNormalizeAgentAccountModels(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "openai",
		APIKey:   "sk-test",
		BaseURL:  "https://api.openai.com/v1",
		APIType:  "openai-completions",
	}

	models, defaultModel, err := normalizeAgentAccountModels(account, []dto.AgentAccountModel{
		{ID: "openai/gpt-4o"},
	}, "openai/gpt-4o", false)
	if err != nil {
		t.Fatalf("normalizeAgentAccountModels failed: %v", err)
	}
	if defaultModel != "openai/gpt-4o" {
		t.Fatalf("unexpected default model: %s", defaultModel)
	}
	if len(models) != 1 {
		t.Fatalf("unexpected models length: %d", len(models))
	}
	if models[0].Name == "" {
		t.Fatal("expected model name to be inferred")
	}
	if models[0].ContextWindow <= 0 || models[0].MaxTokens <= 0 {
		t.Fatalf("expected runtime params to be filled: %+v", models[0])
	}
	if len(models[0].Input) == 0 {
		t.Fatal("expected model input types to be filled")
	}
}

func TestBuildOpenclawModelsFromAccount(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "openai",
		APIKey:   "sk-test",
		BaseURL:  "https://api.openai.com/v1",
		APIType:  "openai-completions",
	}
	models, _, err := normalizeAgentAccountModels(account, []dto.AgentAccountModel{
		{ID: "openai/gpt-4o"},
		{ID: "openai/gpt-4.1"},
	}, "openai/gpt-4o", false)
	if err != nil {
		t.Fatalf("normalizeAgentAccountModels failed: %v", err)
	}
	payload, err := marshalAgentAccountModels(models)
	if err != nil {
		t.Fatalf("marshalAgentAccountModels failed: %v", err)
	}
	account.Model = "openai/gpt-4o"
	account.Models = payload

	primaryModel, defaultsModels, cfg, err := buildOpenclawModelsFromAccount(account, "openai/gpt-4.1")
	if err != nil {
		t.Fatalf("buildOpenclawModelsFromAccount failed: %v", err)
	}
	if primaryModel != "openai/gpt-4.1" {
		t.Fatalf("unexpected primary model: %s", primaryModel)
	}
	if cfg == nil || cfg.Mode != "merge" {
		t.Fatalf("unexpected models config: %+v", cfg)
	}
	if len(cfg.Providers) != 1 {
		t.Fatalf("unexpected providers length: %d", len(cfg.Providers))
	}
	if len(defaultsModels) != 2 {
		t.Fatalf("unexpected defaults models length: %d", len(defaultsModels))
	}
	for _, provider := range cfg.Providers {
		if len(provider.Models) != 2 {
			t.Fatalf("unexpected model entries length: %d", len(provider.Models))
		}
	}
}

func TestBuildOpenclawModelsFromAccountGemini(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "gemini",
		APIKey:   "gemini-key",
		BaseURL:  "https://generativelanguage.googleapis.com",
		APIType:  "openai-completions",
	}
	models, _, err := normalizeAgentAccountModels(account, []dto.AgentAccountModel{
		{ID: "google/gemini-flash-latest"},
	}, "google/gemini-flash-latest", false)
	if err != nil {
		t.Fatalf("normalizeAgentAccountModels failed: %v", err)
	}
	payload, err := marshalAgentAccountModels(models)
	if err != nil {
		t.Fatalf("marshalAgentAccountModels failed: %v", err)
	}
	account.Model = "google/gemini-flash-latest"
	account.Models = payload

	primaryModel, defaultsModels, cfg, err := buildOpenclawModelsFromAccount(account, account.Model)
	if err != nil {
		t.Fatalf("buildOpenclawModelsFromAccount failed: %v", err)
	}
	if primaryModel != "google/gemini-flash-latest" {
		t.Fatalf("unexpected primary model: %s", primaryModel)
	}
	if cfg == nil || len(cfg.Providers) != 1 {
		t.Fatalf("unexpected models config: %+v", cfg)
	}
	if len(defaultsModels) != 1 {
		t.Fatalf("unexpected defaults models length: %d", len(defaultsModels))
	}
}

func TestLoadLegacyAgentAccountModelsForMigrationSkipsEmptyAccount(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "custom",
	}
	models, err := LoadLegacyAgentAccountModelsForMigration(account)
	if err != nil {
		t.Fatalf("LoadLegacyAgentAccountModelsForMigration failed: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected no models, got %d", len(models))
	}
}

func TestEnsurePersistedAgentAccountModelsSkipsEmptyLegacyAccount(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "ollama",
	}
	models, err := ensurePersistedAgentAccountModels(account)
	if err != nil {
		t.Fatalf("ensurePersistedAgentAccountModels failed: %v", err)
	}
	if len(models) != 0 {
		t.Fatalf("expected no persisted models, got %d", len(models))
	}
}

func TestLoadLegacyAgentAccountModelsForMigrationUsesCatalogModels(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "openai",
	}
	models, err := LoadLegacyAgentAccountModelsForMigration(account)
	if err != nil {
		t.Fatalf("LoadLegacyAgentAccountModelsForMigration failed: %v", err)
	}
	if len(models) < 2 {
		t.Fatalf("expected catalog models to be migrated, got %d", len(models))
	}
	if models[0].ID == "" {
		t.Fatal("expected first catalog model id to be populated")
	}
}

func TestMergeCatalogAgentAccountModelsForMigration(t *testing.T) {
	account := &model.AgentAccount{
		Provider: "deepseek",
		APIKey:   "sk-test",
		BaseURL:  "https://api.deepseek.com/v1",
		APIType:  "openai-completions",
	}
	merged, err := MergeCatalogAgentAccountModelsForMigration(account, []dto.AgentAccountModel{
		{ID: "deepseek/deepseek-chat"},
	})
	if err != nil {
		t.Fatalf("MergeCatalogAgentAccountModelsForMigration failed: %v", err)
	}
	if len(merged) != 3 {
		t.Fatalf("expected 3 deepseek catalog models, got %d", len(merged))
	}
}

func TestModelMatchesProviderSupportsAliasedProviderPrefixes(t *testing.T) {
	if !modelMatchesProvider("minimax", "minimax-portal/MiniMax-M2.5") {
		t.Fatal("expected minimax provider to accept minimax-portal model prefix")
	}
	if !modelMatchesProvider("kimi", "moonshot/kimi-k2.5") {
		t.Fatal("expected kimi provider to accept moonshot model prefix")
	}
}
