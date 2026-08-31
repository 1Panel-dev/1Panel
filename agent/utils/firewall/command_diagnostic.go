package firewall

import (
	"fmt"
	"strings"
)

const maxCommandDiagnosticInput = 16 * 1024

// WrapBatchCommandError keeps successful task logs compact while making failed
// generated firewall batches directly diagnosable from the task log.
func WrapBatchCommandError(command, input string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("command=%s failed: %w\nbatch input:\n%s", command, err, numberedCommandInput(input))
}

func numberedCommandInput(input string) string {
	input = strings.TrimSpace(input)
	truncated := false
	if len(input) > maxCommandDiagnosticInput {
		input = input[:maxCommandDiagnosticInput]
		truncated = true
	}
	lines := strings.Split(input, "\n")
	var result strings.Builder
	for index, line := range lines {
		fmt.Fprintf(&result, "%04d  %s\n", index+1, line)
	}
	if truncated {
		result.WriteString("... batch input truncated ...\n")
	}
	return strings.TrimRight(result.String(), "\n")
}
