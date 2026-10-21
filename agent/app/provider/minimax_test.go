package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestMiniMaxCatalog(t *testing.T) {
	meta, ok := Get("minimax")
	if !ok || !reflect.DeepEqual(meta.Models, []Model{{ID: "MiniMax-M3.1-Flash", Name: "MiniMax M3.1 Flash"}, {ID: "MiniMax-M3", Name: "MiniMax M3"}}) {
		t.Fatalf("unexpected MiniMax models: %+v", meta.Models)
	}
	if meta.DefaultAPIType != "anthropic-messages" {
		t.Fatalf("changed default protocol: %s", meta.DefaultAPIType)
	}
	for _, provider := range All() {
		for _, model := range provider.Models {
			if strings.Contains(strings.ToLower(model.ID), "minimax") && model.ID != "MiniMax-M3.1-Flash" && model.ID != "MiniMax-M3" {
				t.Errorf("legacy suggestion in %s: %s", provider.Key, model.ID)
			}
		}
	}
}

func TestMiniMaxVerificationRequests(t *testing.T) {
	for _, region := range []string{"https://api.minimaxi.com", "https://api.minimax.io"} {
		for _, protocol := range []string{"anthropic-messages", "openai-completions"} {
			for _, auth := range []string{AuthModeXAPIKey, AuthModeBearer} {
				for _, model := range []string{"MiniMax-M3.1-Flash", "MiniMax-M3", "MiniMax-M2.7", "MiniMax-M2.7-highspeed", "custom-model"} {
					t.Run(region+"/"+protocol+"/"+auth+"/"+model, func(t *testing.T) {
						base, endpoint := region+"/v1", "/chat/completions"
						if protocol == "anthropic-messages" {
							base, endpoint = region+"/anthropic", "/v1/messages"
						}
						built := BuildVerifyRequest("minimax", protocol, auth, base, "test-key", model)
						if built.URL != base+endpoint {
							t.Fatalf("URL = %s", built.URL)
						}
						var captured map[string]interface{}
						server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
							if r.Method != http.MethodPost || r.URL.Path != strings.TrimPrefix(base, region)+endpoint {
								t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
							}
							if protocol == "anthropic-messages" && auth == AuthModeXAPIKey {
								if r.Header.Get("x-api-key") != "test-key" || r.Header.Get("Authorization") != "" {
									t.Error("incorrect API key authentication")
								}
							} else if r.Header.Get("Authorization") != "Bearer test-key" {
								t.Error("incorrect bearer authentication")
							}
							if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
								t.Error(err)
							}
							w.WriteHeader(http.StatusOK)
						}))
						defer server.Close()
						if err := VerifyAccount("minimax", protocol, auth, server.URL+strings.TrimPrefix(base, region), "test-key", model); err != nil {
							t.Fatal(err)
						}
						if captured["model"] != model || captured["stream"] != false || captured["max_tokens"] != float64(defaultVerifyMaxTokens) {
							t.Fatalf("unexpected serialized body: %#v", captured)
						}
						if model == "MiniMax-M3.1-Flash" {
							if protocol == "anthropic-messages" {
								if !reflect.DeepEqual(captured["thinking"], map[string]interface{}{"type": "adaptive"}) || !reflect.DeepEqual(captured["output_config"], map[string]interface{}{"effort": "max"}) || captured["reasoning_effort"] != nil {
									t.Fatalf("invalid Anthropic Flash reasoning: %#v", captured)
								}
							} else if captured["reasoning_effort"] != "max" || captured["thinking"] != nil || captured["output_config"] != nil {
								t.Fatalf("invalid OpenAI Flash reasoning: %#v", captured)
							}
						} else if captured["reasoning_effort"] != nil || captured["thinking"] != nil || captured["output_config"] != nil {
							t.Fatalf("changed legacy/custom reasoning: %#v", captured)
						}
					})
				}
			}
		}
	}
}

func TestMiniMaxExplicitGatewayAndCustomModels(t *testing.T) {
	for _, provider := range []string{"bailian-coding-plan", "custom", "openrouter"} {
		for _, protocol := range []string{"openai-completions", "anthropic-messages"} {
			for _, model := range []string{"MiniMax-M2.5", "MiniMax-M3.1-Flash", "my-custom-model"} {
				request := BuildVerifyRequest(provider, protocol, AuthModeBearer, "https://example.test/v1", "test-key", model)
				var body map[string]interface{}
				if err := json.Unmarshal(request.Body, &body); err != nil {
					t.Fatal(err)
				}
				if body["model"] != model || body["reasoning_effort"] != nil || body["output_config"] != nil || body["thinking"] != nil {
					t.Fatalf("changed explicit %s request: %#v", provider, body)
				}
			}
		}
	}
}
