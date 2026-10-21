package service

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
)

func TestMiniMaxOpenclawModelMetadata(t *testing.T) {
	for _, provider := range []string{"minimax", "custom", "bailian-coding-plan"} {
		for _, protocol := range []string{"anthropic-messages", "openai-completions"} {
			for _, id := range []string{"MiniMax-M3.1-Flash", "MiniMax-M3", "MiniMax-M2.7", "MiniMax-M2.5", "saved-custom"} {
				t.Run(provider+"/"+protocol+"/"+id, func(t *testing.T) {
					account := &model.AgentAccount{Provider: provider, APIType: protocol, BaseURL: "https://example.test/v1", APIKey: "test-key"}
					_, entry, _, _, err := buildOpenclawAccountModelConfig(account, dto.AgentAccountModel{ID: id, Name: "Saved name"})
					if err != nil {
						t.Fatal(err)
					}
					payload, err := json.Marshal(entry)
					if err != nil {
						t.Fatal(err)
					}
					var actual modelEntry
					if err := json.Unmarshal(payload, &actual); err != nil {
						t.Fatal(err)
					}
					if actual.ID != id || actual.Name != "Saved name" {
						t.Fatalf("changed saved identity: %s", payload)
					}
					if provider == "minimax" && id == "MiniMax-M3.1-Flash" {
						if actual.ContextWindow != 1000000 || actual.MaxTokens != 524288 || !reflect.DeepEqual(actual.Input, []string{"text", "image"}) {
							t.Fatalf("incorrect direct Flash metadata: %s", payload)
						}
					} else if actual.ContextWindow != 0 || actual.MaxTokens != 0 {
						t.Fatalf("changed existing/gateway limits: %s", payload)
					}
				})
			}
		}
	}
}

func TestMiniMaxPreservesSavedOpenclawMetadata(t *testing.T) {
	current := &modelsConfig{Mode: "merge", Providers: map[string]modelProvider{
		"minimax": {Models: []modelEntry{{ID: "MiniMax-M3.1-Flash", Input: []string{"text"}, ContextWindow: 32000, MaxTokens: 4096}}},
	}}
	conf := map[string]interface{}{}
	if err := applyOpenclawModelsConfig(conf, current); err != nil {
		t.Fatal(err)
	}
	next := &modelsConfig{Mode: "merge", Providers: map[string]modelProvider{
		"minimax": {Models: []modelEntry{{ID: "MiniMax-M3.1-Flash", Input: []string{"text", "image"}, ContextWindow: 1000000, MaxTokens: 524288}}},
	}}
	preserveOpenclawModelMetadata(conf, next)
	if !reflect.DeepEqual(next.Providers["minimax"].Models, current.Providers["minimax"].Models) {
		t.Fatalf("saved metadata changed: %+v", next)
	}
}

func TestMiniMaxDefaultAccountRequest(t *testing.T) {
	for _, protocol := range []string{"anthropic-messages", "openai-completions"} {
		account := &model.AgentAccount{Provider: "minimax", APIType: protocol}
		models, err := buildInitialAgentAccountModels(account, nil)
		if err != nil {
			t.Fatal(err)
		}
		selected, err := resolveAgentAccountVerifyModel("minimax", "", models)
		if err != nil || selected != "MiniMax-M3.1-Flash" || len(models) != 2 || models[1].ID != "MiniMax-M3" {
			t.Fatalf("unexpected default: %s, %+v, %v", selected, models, err)
		}
		request := providercatalog.BuildVerifyRequest("minimax", protocol, "", "https://example.test", "test-key", selected)
		var body map[string]interface{}
		if err := json.Unmarshal(request.Body, &body); err != nil || body["model"] != selected {
			t.Fatalf("default request: %s, %v", request.Body, err)
		}
	}
}

func TestMiniMaxHermesConfiguration(t *testing.T) {
	for _, protocol := range []string{"anthropic-messages", "openai-completions"} {
		for _, id := range []string{"MiniMax-M3.1-Flash", "MiniMax-M3", "MiniMax-M2.7", "my-custom-model"} {
			dir := t.TempDir()
			account := &model.AgentAccount{Provider: "minimax", APIType: protocol, BaseURL: "https://api.minimaxi.com/anthropic", APIKey: "test-key"}
			if err := writeHermesConfig(dir, account, id, ""); err != nil {
				t.Fatal(err)
			}
			conf, err := readHermesConfig(filepath.Join(dir, "config.yaml"))
			if err != nil {
				t.Fatal(err)
			}
			if conf.Model.Default != id || conf.Model.Provider != "minimax-cn" || conf.Model.BaseURL != account.BaseURL {
				t.Fatalf("changed Hermes selection: %+v", conf.Model)
			}
		}
	}
}

func TestMiniMaxExplicitAccountModels(t *testing.T) {
	for _, provider := range []string{"minimax", "bailian-coding-plan", "custom"} {
		account := &model.AgentAccount{Provider: provider, APIType: "openai-completions"}
		requested := []dto.AgentAccountModel{{ID: "MiniMax-M2.5", Name: "Saved legacy"}, {ID: "my-custom-model", Name: "Saved custom"}}
		actual, err := buildInitialAgentAccountModels(account, requested)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(actual, requested) {
			t.Fatalf("changed explicit %s models: %+v", provider, actual)
		}
	}
}
