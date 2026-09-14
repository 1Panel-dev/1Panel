package service

import (
	"context"
	"fmt"
	"io"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
)

const (
	firewallTaskHost       = "FirewallTaskHost"
	firewallTaskForwarding = "FirewallTaskForwarding"
	firewallTaskDocker     = "FirewallTaskDocker"
)

func firewallTaskName(operation, subsystem, backend string) string {
	name := i18n.GetMsgByKey(subsystem)
	if backend != "" {
		name += " · " + backend
	}
	key := "FirewallRule" + operation
	if operation == task.TaskExec {
		key = "FirewallTaskInitialize"
	}
	return i18n.GetMsgWithMap(key, map[string]interface{}{"name": name})
}

func queueFirewallRuleTask(subsystem, operation string, labels []string, apply func(context.Context) error) (dto.FilterChainOperationResponse, error) {
	taskItem, err := task.NewTask(firewallTaskName(operation, subsystem, ""), operation, task.TaskScopeFirewall, "", 0)
	if err != nil {
		return dto.FilterChainOperationResponse{}, err
	}
	taskItem.AddSubTaskWithOps(taskItem.Name, func(t *task.Task) error {
		t.Logf("rules=%d", len(labels))
		err := t.TaskCtx.Err()
		if err == nil {
			err = apply(t.TaskCtx)
		}
		succeeded, failed := 0, 0
		for _, label := range labels {
			if err != nil {
				failed++
				t.LogFailedWithErr(label, err)
			} else {
				succeeded++
				t.LogSuccess(label)
			}
		}
		t.Log(i18n.GetMsgWithMap("FirewallRuleOperationResult", map[string]interface{}{
			"succeeded": succeeded, "failed": failed,
		}))
		return err
	}, nil, 0, 0)
	if err := repo.NewITaskRepo().Save(context.Background(), taskItem.Task); err != nil {
		taskItem.LogFailedWithErr(taskItem.Name, err)
		closeUnstartedFirewallTask(taskItem)
		return dto.FilterChainOperationResponse{}, fmt.Errorf("save firewall rule task: %w", err)
	}
	go func() { _ = taskItem.Execute() }()
	return dto.FilterChainOperationResponse{TaskID: taskItem.TaskID, Queued: true}, nil
}

func closeUnstartedFirewallTask(t *task.Task) {
	if cancel, ok := global.LoadTaskCancel(t.TaskID); ok {
		cancel()
	}
	global.RemoveTaskCancel(t.TaskID)
	if closer, ok := t.Logger.Out.(io.Closer); ok {
		_ = closer.Close()
	}
}
