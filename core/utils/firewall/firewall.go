package firewall

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/core/utils/cmd"
)

func UpdatePort(oldPort, newPort string) error {
	firewalld := cmd.Which("firewalld")
	if firewalld {
		status, _ := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("firewall-cmd", "--state")
		isRunning := status == "running\n"
		if isRunning {
			return firewallUpdatePort(oldPort, newPort)
		}
	}

	ufw := cmd.Which("ufw")
	if !ufw {
		return nil
	}
	status, _ := cmd.NewCommandMgr(cmd.WithEnv("LANGUAGE=en_US:en")).RunWithStdout("ufw", "status")
	isRuning := strings.Contains(status, "Status: active")
	if isRuning {
		return ufwUpdatePort(oldPort, newPort)
	}
	return nil
}

func firewallUpdatePort(oldPort, newPort string) error {
	stdout, err := cmd.NewCommandMgr().RunWithStdout("firewall-cmd", "--zone=public", "--add-port="+newPort+"/tcp", "--permanent")
	if err != nil {
		return fmt.Errorf("add (port: %s/tcp) failed, err: %s", newPort, stdout)
	}

	_, _ = cmd.NewCommandMgr().RunWithStdout("firewall-cmd", "--zone=public", "--remove-port="+oldPort+"/tcp", "--permanent")
	_, _ = cmd.NewCommandMgr().RunWithStdout("firewall-cmd", "--reload")
	return nil
}

func ufwUpdatePort(oldPort, newPort string) error {
	stdout, err := cmd.NewCommandMgr().RunWithStdout("ufw", "allow", newPort)
	if err != nil {
		return fmt.Errorf("add (port: %s/tcp) failed, err: %s", newPort, stdout)
	}

	_, _ = cmd.NewCommandMgr().RunWithStdout("ufw", "delete", "allow", oldPort)
	return nil
}
