package terminal

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

type TabCompleter struct {
	shell string
}

func NewTabCompleter() *TabCompleter {
	return &TabCompleter{
		shell: os.Getenv("SHELL"),
	}
}

func (tc *TabCompleter) GetCompletions(input string) []string {
	input = strings.TrimLeft(input, " \t")
	trailingSpace := strings.HasSuffix(input, " ") || strings.HasSuffix(input, "\t")
	if tc.shell == "/bin/bash" || tc.shell == "/usr/bin/bash" {
		return tc.bashComplete(input, trailingSpace)
	}
	if tc.shell == "/bin/zsh" || tc.shell == "/usr/bin/zsh" {
		return tc.zshComplete(input)
	}
	return tc.bashComplete(input, trailingSpace)
}

func (tc *TabCompleter) bashComplete(input string, trailingSpace bool) []string {
	var completions []string
	parts := strings.Fields(input)
	if trailingSpace {
		parts = append(parts, "")
	}

	if len(parts) <= 1 {
		cmdStr := fmt.Sprintf("compgen -c -- '%s'", input)
		cmd := exec.Command("bash", "-c", cmdStr)
		out, err := cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				if line := strings.TrimSpace(scanner.Text()); line != "" {
					completions = append(completions, line)
				}
			}
		}
	} else {
		mainCmd := parts[0]
		curWord := parts[len(parts)-1]
		if trailingSpace {
			curWord = ""
		}

		cmdStr := fmt.Sprintf(`
			[ -f /etc/bash_completion ] && source /etc/bash_completion
			[ -f /usr/share/bash-completion/bash_completion ] && 
				source /usr/share/bash-completion/bash_completion
			
			_completion_loader %s 2>/dev/null || true
			[ -f /usr/share/bash-completion/completions/%s ] && 
				source /usr/share/bash-completion/completions/%s
			
			COMP_WORDS=(%s)
			COMP_CWORD=%d
			COMP_LINE='%s'
			COMP_POINT=%d
			
			_%s 2>/dev/null || complete -p %s &>/dev/null || compgen -f -- '%s'
			
			printf '%%s\n' "${COMPREPLY[@]}"
		`, mainCmd, mainCmd, mainCmd,
			strings.Join(parts, " "), len(parts)-1,
			input, len(input),
			mainCmd, mainCmd, curWord)

		cmd := exec.Command("bash", "-c", cmdStr)
		out, err := cmd.Output()
		if err == nil {
			scanner := bufio.NewScanner(strings.NewReader(string(out)))
			for scanner.Scan() {
				if line := strings.TrimSpace(scanner.Text()); line != "" {
					completions = append(completions, line)
				}
			}
		}
	}

	return tc.filterCompletions(input, completions)
}

func (tc *TabCompleter) filterCompletions(input string, completions []string) []string {
	var filtered []string
	seen := make(map[string]bool)
	inputLower := strings.ToLower(input)
	lastToken := inputLower
	if idx := strings.LastIndex(inputLower, " "); idx != -1 {
		lastToken = strings.TrimSpace(inputLower[idx+1:])
	}
	if lastToken == "" {
		lastToken = inputLower
	}

	for _, comp := range completions {
		if seen[comp] {
			continue
		}
		seen[comp] = true

		if strings.HasPrefix(strings.ToLower(comp), lastToken) {
			filtered = append(filtered, comp)
		}
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		if len(filtered[i]) != len(filtered[j]) {
			return len(filtered[i]) < len(filtered[j])
		}
		return filtered[i] < filtered[j]
	})

	if len(filtered) > 20 {
		filtered = filtered[:20]
	}

	return filtered
}

func (tc *TabCompleter) zshComplete(input string) []string {
	var completions []string

	cmdStr := fmt.Sprintf(`
		autoload -Uz compinit
		compinit
		compset -P '%s'
		compadd -x '%s' 2>/dev/null
	`, input, input)

	cmd := exec.Command("zsh", "-c", cmdStr)
	out, _ := cmd.Output()

	scanner := bufio.NewScanner(strings.NewReader(string(out)))
	for scanner.Scan() {
		if line := strings.TrimSpace(scanner.Text()); line != "" {
			completions = append(completions, line)
		}
	}

	return completions
}
