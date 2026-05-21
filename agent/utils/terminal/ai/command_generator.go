package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"unicode"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

type CommandGenerator struct {
	client Client
}

var builtInRiskCommands = mustLoadBuiltInRiskCommands()

type CommandGenerateRequest struct {
	Input          string
	Shell          string
	WorkingDir     string
	OS             string
	RecentCommands []string
	DirectoryHints []string
}

type CommandGenerateResponse struct {
	Command  string
	Model    string
	Provider string
	RawText  string
	Usage    ResponseUsage
}

func NewCommandGeneratorFromConfig(cfg GeneratorConfig) (*CommandGenerator, error) {
	client, err := NewClient(ClientConfig{
		Provider:  cfg.Provider,
		BaseURL:   cfg.BaseURL,
		APIKey:    cfg.APIKey,
		Model:     cfg.Model,
		APIType:   cfg.APIType,
		MaxTokens: cfg.MaxTokens,
	})
	if err != nil {
		return nil, err
	}
	return NewCommandGenerator(client)
}

func NewCommandGenerator(client Client) (*CommandGenerator, error) {
	if client == nil {
		return nil, fmt.Errorf("client is required")
	}
	return &CommandGenerator{client: client}, nil
}

func (g *CommandGenerator) Generate(ctx context.Context, req CommandGenerateRequest) (*CommandGenerateResponse, error) {
	if strings.TrimSpace(req.Input) == "" {
		return nil, fmt.Errorf("input is required")
	}

	resp, err := g.client.ChatCompletion(ctx, ChatCompletionRequest{
		Messages: []ChatMessage{
			{Role: "system", Content: buildCommandSystemPrompt()},
			{Role: "user", Content: buildCommandUserPrompt(req)},
		},
	})
	if err != nil {
		return nil, err
	}

	command := sanitizeCommand(resp.Content)
	if command == "" {
		return nil, fmt.Errorf("model returned empty command")
	}
	if err := validateGeneratedCommand(command); err != nil {
		return nil, err
	}

	return &CommandGenerateResponse{
		Command:  command,
		Model:    resp.Model,
		Provider: providerNameFromModel(resp.Model),
		RawText:  resp.RawText,
		Usage:    resp.Usage,
	}, nil
}

func buildCommandSystemPrompt() string {
	return strings.Join([]string{
		"You are a shell command generator.",
		"Return exactly one command suitable for direct execution in the user's shell.",
		"Do not include markdown, code fences, explanations, numbering, comments, or backticks.",
		"If multiple commands are required, join them with shell operators in a single line.",
		"Prefer safe, non-destructive commands unless the user explicitly asks for destructive behavior.",
		"Preserve the user's language when filenames or arguments are ambiguous, but output only the command.",
	}, "\n")
}

func buildCommandUserPrompt(req CommandGenerateRequest) string {
	var sections []string
	sections = append(sections, "Task:\n"+strings.TrimSpace(req.Input))

	var env []string
	if shell := strings.TrimSpace(req.Shell); shell != "" {
		env = append(env, "Shell: "+shell)
	}
	if wd := strings.TrimSpace(req.WorkingDir); wd != "" {
		env = append(env, "Working directory: "+wd)
	}
	if osName := strings.TrimSpace(req.OS); osName != "" {
		env = append(env, "Operating system: "+osName)
	}
	if len(env) > 0 {
		sections = append(sections, "Environment:\n"+strings.Join(env, "\n"))
	}

	if block := formatBulletBlock(req.DirectoryHints); block != "" {
		sections = append(sections, "Directory hints:\n"+block)
	}
	if block := formatBulletBlock(req.RecentCommands); block != "" {
		sections = append(sections, "Recent commands:\n"+block)
	}

	sections = append(sections, "Output requirement:\nReturn one shell command only.")
	return strings.Join(sections, "\n\n")
}

func formatBulletBlock(values []string) string {
	var lines []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		lines = append(lines, "- "+value)
	}
	return strings.Join(lines, "\n")
}

func sanitizeCommand(raw string) string {
	command := strings.TrimSpace(raw)
	if command == "" {
		return ""
	}
	command = strings.TrimPrefix(command, "```sh")
	command = strings.TrimPrefix(command, "```bash")
	command = strings.TrimPrefix(command, "```zsh")
	command = strings.TrimPrefix(command, "```shell")
	command = strings.TrimPrefix(command, "```")
	command = strings.TrimSuffix(command, "```")
	command = strings.TrimSpace(command)

	lines := strings.Split(command, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "command:") {
			line = strings.TrimSpace(line[len("command:"):])
		}
		return strings.Trim(line, "` ")
	}
	return ""
}

func validateGeneratedCommand(command string) error {
	command = strings.TrimSpace(command)
	if command == "" {
		return fmt.Errorf("model returned empty command")
	}
	if containsControlCharacters(command) {
		return fmt.Errorf("model returned unsafe command")
	}
	if strings.ContainsAny(command, "\x00\r\n") {
		return fmt.Errorf("model returned unsafe command")
	}
	if containsShellControlSyntax(command) {
		return fmt.Errorf("model returned unsafe command")
	}
	if isRiskCommand(command, builtInRiskCommands) {
		return fmt.Errorf("model returned risky command")
	}
	if isDangerousExecutionCommand(command) {
		return fmt.Errorf("model returned unsafe command")
	}
	return nil
}

func containsShellControlSyntax(command string) bool {
	if strings.ContainsAny(command, ";|&`") {
		return true
	}
	if strings.Contains(command, "$(") || strings.Contains(command, "${") {
		return true
	}
	return false
}

func containsControlCharacters(command string) bool {
	for _, r := range command {
		if unicode.IsControl(r) {
			return true
		}
	}
	return false
}

func isRiskCommand(command string, riskCommands []string) bool {
	command = strings.ToLower(strings.TrimSpace(command))
	if command == "" {
		return false
	}
	for _, riskCommand := range riskCommands {
		riskCommand = strings.ToLower(strings.TrimSpace(riskCommand))
		if riskCommand == "" {
			continue
		}
		if strings.Contains(command, riskCommand) {
			return true
		}
	}
	return false
}

func isDangerousExecutionCommand(command string) bool {
	fields := strings.Fields(strings.ToLower(strings.TrimSpace(command)))
	if len(fields) == 0 {
		return false
	}

	switch fields[0] {
	case "bash", "sh", "zsh", "dash", "ksh", "fish", "tcsh", "csh",
		"python", "python3", "perl", "ruby", "php", "node", "deno", "lua",
		"pwsh", "powershell", "cmd", "osascript":
		return true
	default:
		return false
	}
}

func mustLoadBuiltInRiskCommands() []string {
	commands := make([]string, 0)
	if err := json.Unmarshal([]byte(constant.DefaultTerminalAIRiskCommands), &commands); err != nil {
		return nil
	}
	return commands
}

func providerNameFromModel(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	if parts := strings.SplitN(model, "/", 2); len(parts) == 2 {
		return parts[0]
	}
	return ""
}
