package provider

import "testing"

// OpenClaw splits a model ref on the first "/": the provider key, then the
// model id inside models.providers.<key>.
func TestBuildOpenClawProviderPatchPrimaryModel(t *testing.T) {
	cases := []struct {
		provider, apiType, model, primary string
	}{
		{"openrouter", "openai-completions", "anthropic/claude-sonnet-4.5", "openrouter/anthropic/claude-sonnet-4.5"},
		{"openrouter", "openai-completions", "openrouter/auto", "openrouter/auto"},
		{"openai", "openai-responses", "gpt-5", "openai/gpt-5"},
		{"openai", "openai-responses", "openai/gpt-5", "openai/gpt-5"},
		{"anthropic", "anthropic-messages", "claude-sonnet-4-5", "anthropic/claude-sonnet-4-5"},
	}
	for _, tc := range cases {
		patch, err := BuildOpenClawProviderPatch(tc.provider, tc.model, tc.apiType, "", "https://example.test/v1", "key")
		if err != nil {
			t.Fatalf("%s %s: %v", tc.provider, tc.model, err)
		}
		if patch.PrimaryModel != tc.primary {
			t.Errorf("%s %s: PrimaryModel = %q, want %q", tc.provider, tc.model, patch.PrimaryModel, tc.primary)
		}
	}
}
