package ai

import (
	"context"
	"fmt"
	"runtime"
	"time"
)

const validateTimeout = 15 * time.Second

func ValidateTerminalAccount(accountID uint) error {
	cfg, _, err := ResolveGeneratorConfig(accountID)
	if err != nil {
		return err
	}

	generator, err := NewCommandGeneratorFromConfig(cfg)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), validateTimeout)
	defer cancel()

	resp, err := generator.Generate(ctx, CommandGenerateRequest{
		Input: "Print current directory.",
		Shell: "bash",
		OS:    runtime.GOOS,
	})
	if err != nil {
		return err
	}
	if resp == nil || resp.Command == "" {
		return fmt.Errorf("terminal ai account returned empty response")
	}
	return nil
}
