package terminal

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/1Panel-dev/1Panel/agent/global"
	terminalai "github.com/1Panel-dev/1Panel/agent/utils/terminal/ai"
)

const lineClearControl = 21

type aiInputInterceptor struct {
	config  terminalai.GeneratorConfig
	timeout time.Duration
	shell   string
	prefix  string

	mu             sync.Mutex
	currentLine    []byte
	recentCommands []string
	riskCommands   []string
	inEscapeSeq    bool
}

func newAIInputInterceptor(shell string) *aiInputInterceptor {
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
		prefix:       settings.Prefix,
		riskCommands: append([]string(nil), settings.RiskCommands...),
	}
}

func (i *aiInputInterceptor) refreshSettings() error {
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
	return nil
}

func (i *aiInputInterceptor) HandleEnter() (string, bool) {
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
	line := sanitizeInputLine(string(i.currentLine))
	i.currentLine = nil
	i.inEscapeSeq = false
	recentCommands := append([]string(nil), i.recentCommands...)
	i.mu.Unlock()

	if !strings.HasPrefix(line, i.prefix) {
		if line != "" {
			i.pushRecentCommand(line)
		}
		return "", false
	}
	prompt := strings.TrimSpace(strings.TrimPrefix(line, i.prefix))
	if prompt == "" {
		return "", false
	}

	ctx, cancel := context.WithTimeout(context.Background(), i.timeout)
	defer cancel()
	generator, err := terminalai.NewCommandGeneratorFromConfig(i.config)
	if err != nil {
		global.LOG.Errorf("create terminal ai generator failed: %v", err)
		return "", false
	}
	resp, err := generator.Generate(ctx, terminalai.CommandGenerateRequest{
		Input:          prompt,
		Shell:          firstNonEmpty(i.shell, filepath.Base(strings.TrimSpace(os.Getenv("SHELL")))),
		OS:             runtime.GOOS,
		RecentCommands: recentCommands,
	})
	if err != nil {
		global.LOG.Errorf("generate terminal ai command failed: %v", err)
		return "", false
	}
	if i.isRiskCommand(resp.Command) {
		return ": # blocked risky command: " + resp.Command, true
	}
	return resp.Command, strings.TrimSpace(resp.Command) != ""
}

func (i *aiInputInterceptor) TrackInput(data []byte) {
	if i == nil || len(data) == 0 {
		return
	}
	i.mu.Lock()
	defer i.mu.Unlock()
	for _, b := range data {
		if i.inEscapeSeq {
			if isEscapeSequenceTerminator(b) {
				i.inEscapeSeq = false
			}
			continue
		}
		switch b {
		case '\r', '\n':
			i.currentLine = nil
		case 0x08, 0x7f:
			i.currentLine = trimLastRuneBytes(i.currentLine)
		case lineClearControl:
			i.currentLine = nil
		case 0x1b:
			// Ignore ANSI escape sequences such as arrow keys and bracketed paste markers.
			i.inEscapeSeq = true
		default:
			if b < 0x20 && b != '\t' {
				continue
			}
			i.currentLine = append(i.currentLine, b)
		}
	}
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
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	var builder strings.Builder
	builder.Grow(len(raw))
	for _, r := range raw {
		switch {
		case unicode.IsControl(r) && r != '\t' && r != ' ':
			continue
		case unicode.In(r, unicode.Cf):
			continue
		case r == '\u00a0' || r == '\u2007' || r == '\u202f' || r == '\u3000':
			builder.WriteRune(' ')
		default:
			builder.WriteRune(r)
		}
	}
	return strings.TrimSpace(builder.String())
}

func trimLastRuneBytes(data []byte) []byte {
	if len(data) == 0 {
		return data
	}
	_, size := utf8.DecodeLastRune(data)
	if size <= 0 || size > len(data) {
		return data[:len(data)-1]
	}
	return data[:len(data)-size]
}

func isEscapeSequenceTerminator(b byte) bool {
	return b >= 0x40 && b <= 0x7e
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
