package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/google/uuid"
)

var networkCleanupMu sync.Mutex

func (u *ContainerService) CleanNetworks() (*dto.NetworkCleanupTask, error) {
	taskID := uuid.NewString()
	if err := u.Prune(dto.ContainerPrune{TaskID: taskID, PruneType: "network"}); err != nil {
		return nil, err
	}
	return &dto.NetworkCleanupTask{TaskID: taskID}, nil
}

func executeNetworkCleanup(t *task.Task, cli docker.NetworkCleanupClient, cutoff ...time.Time) error {
	networkCleanupMu.Lock()
	defer networkCleanupMu.Unlock()
	if err := t.TaskCtx.Err(); err != nil {
		return err
	}
	var until time.Time
	if len(cutoff) > 0 {
		until = cutoff[0]
	}
	t.Log(i18n.GetMsgByKey("PruneStart"))
	report, err := docker.CleanUnusedNetworksBefore(t.TaskCtx, cli, until, func(status string, item dto.NetworkCleanupItem) {
		key := "NetworkCleanupDeleted"
		if status == "skipped" || status == "failed" {
			key = map[string]string{
				"recent":              "NetworkCleanupRecent",
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
