package terminal

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
)

const lineClearControl = 21

type aiInputInterceptor struct {
	config  terminalai.GeneratorConfig
	timeout time.Duration
	shell   string
	lang    string
	prefix  string
	version uint64

	mu             sync.Mutex
	currentLine    string
	recentCommands []string
	riskCommands   []string
}

func newAIInputInterceptor(shell string, lang string) *aiInputInterceptor {
	settings, config, timeout, err := terminalai.LoadTerminalRuntimeSettings()
	if err != nil {
		if !os.IsNotExist(err) {
			global.LOG.Warnf("load terminal ai config failed: %v", err)
		}
		return nil
	}
	if strings.TrimSpace(config.APIKey) == "" && !strings.EqualFold(strings.TrimSpace(config.Provider), "ollama") {
		return nil
	}
	return &aiInputInterceptor{
		config:       config,
		timeout:      timeout,
		shell:        strings.TrimSpace(shell),
		lang:         strings.TrimSpace(lang),
		prefix:       settings.Prefix,
		riskCommands: append([]string(nil), settings.RiskCommands...),
		version:      terminalai.CurrentTerminalRuntimeVersion(),
	}
}

func (i *aiInputInterceptor) refreshSettings() error {
	if !i.needsRefresh() {
		return nil
	}
	settings, config, timeout, err := terminalai.LoadTerminalRuntimeSettings()
	if err != nil {
		return err
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.config = config
	i.timeout = timeout
	i.prefix = settings.Prefix
	i.riskCommands = append([]string(nil), settings.RiskCommands...)
	i.version = terminalai.CurrentTerminalRuntimeVersion()
	return nil
}

func (i *aiInputInterceptor) SetCurrentLine(line string) {
	if i == nil {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.currentLine = normalizeCurrentLine(line, i.prefix)
}

func (i *aiInputInterceptor) needsRefresh() bool {
	if i == nil {
		return false
	}
	currentVersion := terminalai.CurrentTerminalRuntimeVersion()
	i.mu.Lock()
	defer i.mu.Unlock()
	return i.version != currentVersion
}

func (i *aiInputInterceptor) HandleEnter(onStart func(), onDone func(string), onError func(string)) (string, bool) {
	if i == nil {
		return "", false
	}
	if err := i.refreshSettings(); err != nil {
		if !os.IsNotExist(err) {
			global.LOG.Warnf("refresh terminal ai config failed: %v", err)
		}
		return "", false
	}
	i.mu.Lock()
	currentLine := i.currentLine
	i.currentLine = ""
	recentCommands := append([]string(nil), i.recentCommands...)
	i.mu.Unlock()

	if !matchesAIPrefix(currentLine, i.prefix) {
		if currentLine != "" {
			i.pushRecentCommand(currentLine)
		}
		return "", false
	}
	prompt := strings.TrimSpace(strings.TrimPrefix(currentLine, i.prefix))
	if prompt == "" {
		return "", true
	}
	if onStart != nil {
		onStart()
	}

	start := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), i.timeout)
	defer cancel()
	generator, err := terminalai.NewCommandGeneratorFromConfig(i.config)
	if err != nil {
		global.LOG.Errorf("create terminal ai generator failed: %v", err)
		if onError != nil {
			onError(i18n.GetMsgWithMapAndLang(i.lang, "TerminalAIRequestFailed", map[string]interface{}{
				"err": err.Error(),
			}))
		}
		return "", true
	}
	resp, err := generator.Generate(ctx, terminalai.CommandGenerateRequest{
		Input:          prompt,
		Shell:          firstNonEmpty(i.shell, filepath.Base(strings.TrimSpace(os.Getenv("SHELL")))),
		OS:             runtime.GOOS,
		RecentCommands: recentCommands,
	})
	if err != nil {
		global.LOG.Errorf("generate terminal ai command failed: %v", err)
		if onError != nil {
			onError(i18n.GetMsgWithMapAndLang(i.lang, "TerminalAIRequestFailed", map[string]interface{}{
				"err": err.Error(),
			}))
		}
		return "", true
	}
	if onDone != nil {
		onDone(i18n.GetMsgWithMapAndLang(i.lang, "TerminalAIReadyToExecute", map[string]interface{}{
			"duration": formatAIDuration(time.Since(start)),
			"tokens":   resolveTotalTokens(resp.Usage),
		}))
	}
	if i.isRiskCommand(resp.Command) {
		return "# " + i18n.GetMsgWithMapAndLang(i.lang, "TerminalAIBlockedRiskyCommand", map[string]interface{}{
			"command": resp.Command,
		}), true
	}
	return resp.Command, strings.TrimSpace(resp.Command) != ""
}

func resolveTotalTokens(usage terminalai.ResponseUsage) int {
	if usage.TotalTokens > 0 {
		return usage.TotalTokens
	}
	return usage.PromptTokens + usage.CompletionTokens
}

func formatAIDuration(duration time.Duration) string {
	if duration < time.Millisecond {
		return duration.String()
	}
	return duration.Round(time.Millisecond).String()
}

func (i *aiInputInterceptor) pushRecentCommand(command string) {
	if i == nil {
		return
	}
	command = strings.TrimSpace(command)
	if command == "" || strings.HasPrefix(command, i.prefix) {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	i.recentCommands = append([]string{command}, i.recentCommands...)
	if len(i.recentCommands) > 8 {
		i.recentCommands = i.recentCommands[:8]
	}
}

func isEnterInput(data []byte) bool {
	if len(data) == 1 && (data[0] == '\r' || data[0] == '\n') {
		return true
	}
	return len(data) == 2 && data[0] == '\r' && data[1] == '\n'
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func sanitizeInputLine(raw string) string {
	return strings.TrimSpace(raw)
}

func matchesAIPrefix(line, prefix string) bool {
	line = strings.TrimSpace(line)
	prefix = strings.TrimSpace(prefix)
	if line == "" || prefix == "" {
		return false
	}
	return line == prefix || strings.HasPrefix(line, prefix+" ")
}

func normalizeCurrentLine(raw, prefix string) string {
	line := sanitizeInputLine(raw)
	prefix = strings.TrimSpace(prefix)
	if line == "" || prefix == "" {
		return line
	}
	if matchesAIPrefix(line, prefix) {
		return line
	}
	prefixIdx := strings.Index(line, prefix)
	if prefixIdx < 0 {
		return line
	}
	for _, marker := range []string{"# ", "$ ", "% ", "> "} {
		if idx := strings.LastIndex(line[:prefixIdx], marker); idx >= 0 {
			promptPart := strings.TrimSpace(line[:idx+len(marker)-1])
			if looksLikePromptPrefix(promptPart) {
				return strings.TrimSpace(line[idx+len(marker):])
			}
		}
	}
	return line
}

func looksLikePromptPrefix(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	if strings.ContainsAny(value, "'\"`") {
		return false
	}
	return true
}

func (i *aiInputInterceptor) isRiskCommand(command string) bool {
	command = strings.ToLower(strings.TrimSpace(command))
	if command == "" {
		return false
	}
	for _, riskCommand := range i.riskCommands {
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
