package service

import (
	"errors"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
)

var networkCleanupSlot = make(chan struct{}, 1)

func executeNetworkCleanup(t *task.Task, cli docker.NetworkCleanupClient, until time.Time) error {
	select {
	case networkCleanupSlot <- struct{}{}:
		defer func() { <-networkCleanupSlot }()
	case <-t.TaskCtx.Done():
		return t.TaskCtx.Err()
	}
	if err := t.TaskCtx.Err(); err != nil {
		return err
	}
	t.Log(i18n.GetMsgByKey("PruneStart"))
	report, err := docker.CleanUnusedNetworks(t.TaskCtx, cli, until, func(item docker.NetworkCleanupItem) {
		key := "NetworkCleanupDeleted"
		switch item.Reason {
		case "recent":
			key = "NetworkCleanupRecent"
		case "protected":
			key = "NetworkCleanupProtected"
		case "container_connected":
			key = "NetworkCleanupConnected"
		case "unsupported_network":
			key = "NetworkCleanupUnsupported"
		case "already_removed":
			key = "NetworkCleanupGone"
		case "inspect_failed":
			key = "NetworkCleanupInspectFailed"
		case "remove_failed":
			key = "NetworkCleanupRemoveFailed"
		}
		t.Log(i18n.GetMsgWithMap(key, map[string]interface{}{"name": item.Name, "id": item.ID}))
	})
	if report != nil {
		t.Log(i18n.GetMsgWithMap("NetworkCleanupSummary", map[string]interface{}{
			"deleted": report.Deleted, "skipped": report.Skipped, "failed": report.Failed,
		}))
	}
	if err != nil {
		return err
	}
	if report.Failed > 0 {
		return errors.New(i18n.GetMsgByKey("NetworkCleanupPartialFailure"))
	}
	return nil
}
