package lifecycle

import (
	"errors"
	"fmt"
	"os"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
)

const fail2BanRestoreWithFirewallMarker = "/run/1panel_fail2ban_restore_with_firewall"

type Operation string

const (
	OperationStart   Operation = "start"
	OperationStop    Operation = "stop"
	OperationRestart Operation = "restart"
)

type Operator struct {
	client Client
}

func NewOperator(client Client) *Operator {
	return &Operator{client: client}
}

func (o *Operator) Operate(operation Operation, withDockerRestart bool, prepareStart func(Client) error) error {
	if o.client.Name() == ProviderFirewalld && operation == OperationStop {
		if err := rememberFail2BanBeforeFirewallStop(); err != nil {
			return err
		}
	}

	switch operation {
	case OperationStart:
		if err := o.client.Start(); err != nil {
			return err
		}
		if prepareStart != nil {
			if err := prepareStart(o.client); err != nil {
				return errors.Join(err, o.client.Stop())
			}
		}
	case OperationStop:
		if err := o.client.Stop(); err != nil {
			return err
		}
	case OperationRestart:
		if err := o.client.Restart(); err != nil {
			return err
		}
		if prepareStart != nil {
			if err := prepareStart(o.client); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("not supported operation: %s", operation)
	}

	if withDockerRestart {
		if err := controller.HandleRestart("docker"); err != nil {
			return fmt.Errorf("failed to restart Docker: %v", err)
		}
	}
	if o.client.Name() == ProviderFirewalld && operation == OperationStart {
		return restoreFail2BanAfterFirewallStart()
	}
	return nil
}

func rememberFail2BanBeforeFirewallStop() error {
	exists, err := controller.CheckExist("fail2ban.service")
	if err != nil {
		global.LOG.Warnf("check fail2ban.service installation before stopping the firewall failed: %v", err)
	}
	if !exists {
		return nil
	}
	active, err := controller.CheckActive("fail2ban.service")
	if err != nil {
		global.LOG.Warnf("check fail2ban.service status before stopping the firewall failed: %v", err)
	}
	if !active {
		return nil
	}
	if err := os.WriteFile(fail2BanRestoreWithFirewallMarker, nil, 0600); err != nil {
		return fmt.Errorf("mark Fail2Ban for restoration with the firewall: %w", err)
	}
	return nil
}

func restoreFail2BanAfterFirewallStart() error {
	if _, err := os.Stat(fail2BanRestoreWithFirewallMarker); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("load Fail2Ban restore marker after starting the firewall: %w", err)
	}
	if err := controller.HandleStart("fail2ban.service"); err != nil {
		return fmt.Errorf("restore Fail2Ban after starting the firewall: %w", err)
	}
	if err := os.Remove(fail2BanRestoreWithFirewallMarker); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("clear Fail2Ban firewall restore status: %w", err)
	}
	return nil
}
