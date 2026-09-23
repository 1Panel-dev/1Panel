package service

import (
	"fmt"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
)

var networkCleanupMu sync.Mutex

func (u *ContainerService) CleanNetworks() (*dto.NetworkCleanupTask, error) {
	taskItem, err := task.NewTaskWithOps(i18n.GetMsgByKey("Network"), task.TaskClean, task.TaskScopeContainer, "", 0)
	if err != nil {
		return nil, err
	}
	taskItem.AddSubTask(i18n.GetMsgByKey("TaskClean"), func(t *task.Task) error {
		networkCleanupMu.Lock()
		defer networkCleanupMu.Unlock()
		if err := t.TaskCtx.Err(); err != nil {
			return err
		}
		cli, err := docker.NewDockerClient()
		if err != nil {
			return err
		}
		defer cli.Close()
		return executeNetworkCleanup(t, cli)
	}, nil)
	go func() {
		if err := taskItem.Execute(); err != nil {
			global.LOG.Errorf("network cleanup task %s failed: %v", taskItem.TaskID, err)
		}
	}()
	return &dto.NetworkCleanupTask{TaskID: taskItem.TaskID}, nil
}

func executeNetworkCleanup(t *task.Task, cli docker.NetworkCleanupClient) error {
	t.Log(i18n.GetMsgByKey("PruneStart"))
	report, err := docker.CleanUnusedNetworks(t.TaskCtx, cli, func(status string, item dto.NetworkCleanupItem) {
		key := "NetworkCleanupDeleted"
		if status == "skipped" || status == "failed" {
			key = map[string]string{
				"protected":           "NetworkCleanupProtected",
				"container_connected": "NetworkCleanupConnected",
				"network_in_use":      "NetworkCleanupConnected",
				"unsupported_network": "NetworkCleanupUnsupported",
				"already_removed":     "NetworkCleanupGone",
				"inspect_failed":      "NetworkCleanupInspectFailed",
				"remove_failed":       "NetworkCleanupRemoveFailed",
			}[item.Reason]
		}
		t.Log(i18n.GetMsgWithMap(key, map[string]interface{}{"name": item.Name, "id": item.ID}))
	})
	if report != nil {
		t.Log(i18n.GetMsgWithMap("NetworkCleanupSummary", map[string]interface{}{
			"deleted": len(report.Deleted), "skipped": len(report.Skipped), "failed": len(report.Failed),
		}))
	}
	if err != nil {
		return err
	}
	if len(report.Failed) > 0 {
		return fmt.Errorf("%s", i18n.GetMsgByKey("NetworkCleanupPartialFailure"))
	}
	return nil
}
