package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const validateTimeout = 15 * time.Second

func ValidateTerminalAccount(accountID uint) error {
	if accountID == 0 {
		return fmt.Errorf("ai account is required")
	}

	cfg, _, err := ResolveGeneratorConfig(accountID)
	if err != nil {
		return err
	}

	client, err := NewClient(ClientConfig{
		Provider:  cfg.Provider,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		APIType:   cfg.APIType,
		MaxTokens: cfg.MaxTokens,
		Timeout:   validateTimeout,
	})
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), validateTimeout)
	defer cancel()

	resp, err := client.ChatCompletion(ctx, ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: "You are a shell command generator. Return exactly one shell command."},
			{Role: "user", Content: "Print current directory."},
		},
		MaxTokens: 16,
	})
	if err != nil {
		return err
	}
	if resp == nil || strings.TrimSpace(resp.Content) == "" {
		return fmt.Errorf("terminal ai account returned empty response")
	}
	return nil
}
