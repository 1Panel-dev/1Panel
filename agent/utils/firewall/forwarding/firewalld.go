package forwarding

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type firewalldCommandRunner interface {
	Run(name string, args ...string) error
	RunWithStdout(name string, args ...string) (string, error)
}

type defaultFirewalldCommandRunner struct{}

func (defaultFirewalldCommandRunner) Run(name string, args ...string) error {
	return cmd.NewCommandMgr().Run(name, args...)
}

func (defaultFirewalldCommandRunner) RunWithStdout(name string, args ...string) (string, error) {
	return cmd.NewCommandMgr().RunWithStdout(name, args...)
}

type firewalldAdapter struct {
	runner firewalldCommandRunner
}

func newFirewalldAdapter() *firewalldAdapter {
	return &firewalldAdapter{runner: defaultFirewalldCommandRunner{}}
}

func (f *firewalldAdapter) Name() string {
	return "firewalld"
}

func (f *firewalldAdapter) List() ([]Rule, error) {
	if err := f.Enable(); err != nil {
		global.LOG.Errorf("init port forward failed, err: %v", err)
	}
	stdout, err := f.runner.RunWithStdout("firewall-cmd", "--zone=public", "--list-forward-ports")
	if err != nil {
		return nil, err
	}
	return parseFirewalldRules(stdout), nil
}

func parseFirewalldRules(stdout string) []Rule {
	var rules []Rule
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		parts := strings.Split(line, ":")
		if len(parts) < 4 {
			continue
		}
		if parts[3] == "toaddr=" {
			parts[3] = "127.0.0.1"
		}
		rules = append(rules, Rule{
			Port:       strings.TrimPrefix(parts[0], "port="),
			Protocol:   strings.TrimPrefix(parts[1], "proto="),
			TargetIP:   strings.TrimPrefix(parts[3], "toaddr="),
			TargetPort: strings.TrimPrefix(parts[2], "toport="),
		})
	}
	return rules
}

func (f *firewalldAdapter) Operate(rule Rule, operation string) error {
	if cmd.CheckIllegal(operation, rule.Port, rule.Protocol, rule.TargetIP, rule.TargetPort) {
		return buserr.New("ErrCmdIllegal")
	}
	args := buildFirewalldForwardArgs(rule, operation)
	if err := f.runner.Run("firewall-cmd", args...); err != nil {
		return fmt.Errorf("%s port forward failed, %s", operation, err)
	}
	return f.reload()
}

func buildFirewalldForwardArgs(rule Rule, operation string) []string {
	forwardRule := fmt.Sprintf("--%s-forward-port=port=%s:proto=%s:toport=%s", operation, rule.Port, rule.Protocol, rule.TargetPort)
	if rule.TargetIP != "" && rule.TargetIP != "127.0.0.1" && rule.TargetIP != "localhost" {
		forwardRule = fmt.Sprintf("--%s-forward-port=port=%s:proto=%s:toaddr=%s:toport=%s", operation, rule.Port, rule.Protocol, rule.TargetIP, rule.TargetPort)
	}
	return []string{"--zone=public", forwardRule, "--permanent"}
}

func (f *firewalldAdapter) Enable() error {
	stdout, err := f.runner.RunWithStdout("firewall-cmd", "--zone=public", "--query-masquerade")
	if err != nil {
		if strings.HasSuffix(strings.TrimSpace(stdout), "no") {
			if err := f.runner.Run("firewall-cmd", "--zone=public", "--add-masquerade", "--permanent"); err != nil {
				return err
			}
			return f.reload()
		}
		return err
	}
	return nil
}

func (f *firewalldAdapter) reload() error {
	if err := f.runner.Run("firewall-cmd", "--reload"); err != nil {
		return fmt.Errorf("reload firewall failed, err: %v", err)
	}
	return nil
}

func (f *firewalldAdapter) InitStatus() (bool, bool) {
	return true, true
}

func (f *firewalldAdapter) Replay() error {
	return nil
}

// Native firewalld forward-port entries have no historical 1Panel ownership
// marker. Conservatively leave them untouched during provider cleanup.
func (f *firewalldAdapter) CleanupOwned() error {
	return nil
}
