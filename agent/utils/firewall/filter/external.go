package filter

import (
	"context"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type CommentRuleReader interface {
	ReadRulesByComment(context.Context, Scope, string) (string, error)
}

func ReadRulesByComment(ctx context.Context, executable string, args []string, comment string) (string, error) {
	name, args := cmd.WrapWithOptionalSudo(executable, args...)
	return cmd.NewCommandMgr(cmd.WithContext(ctx), cmd.WithTimeout(60*time.Second), cmd.WithEnv("LC_ALL=C", "LANGUAGE=en_US:en")).RunPipe(
		cmd.PipeCommand{Name: name, Args: args},
		cmd.PipeCommand{Name: "sh", Args: []string{"-c", `grep -F -- "$1"; result=$?; if [ "$result" -eq 1 ]; then exit 0; fi; exit "$result"`, "sh", comment}},
	)
}
