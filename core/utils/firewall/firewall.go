package firewall

import (
	"fmt"

	"github.com/1Panel-dev/1Panel/core/utils/cmd"
)

func UpdatePort(oldPort, newPort string) error {
	firewalld := cmd.Which("firewalld")
	if firewalld {
		status, _ := cmd.Exec("LANGUAGE=en_US:en firewall-cmd --state")
		isRunning := status == "running\n"
		if isRunning {
			return firewallUpdatePort(oldPort, newPort)
		}
	}

	ufw := cmd.Which("ufw")
	if !ufw {
		return nil
	}
	status, _ := cmd.Exec("LANGUAGE=en_US:en ufw status | grep Status")
	isRuning := status == "Status: active\n"
	if isRuning {
		return ufwUpdatePort(oldPort, newPort)
	}
	return nil
}

func firewallUpdatePort(oldPort, newPort string) error {
	stdout, err := cmd.Execf("firewall-cmd --zone=public --add-port=%s/tcp --permanent", newPort)
	if err != nil {
		return fmt.Errorf("add (port: %s/tcp) failed, err: %s", newPort, stdout)
	}

	_, _ = cmd.Execf("firewall-cmd --zone=public --remove-port=%s/tcp --permanent", oldPort)
	_, _ = cmd.Exec("firewall-cmd --reload")
	return nil
}

func ufwUpdatePort(oldPort, newPort string) error {
	stdout, err := cmd.Execf("ufw allow %s", newPort)
	if err != nil {
		return fmt.Errorf("add (port: %s/tcp) failed, err: %s", newPort, stdout)
	}

	_, _ = cmd.Execf("ufw delete allow %s", oldPort)
	return nil
}
