package service

import (
	"errors"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"gorm.io/gorm"
)

func appInstallOperationFailureStatus(status string) string {
	switch status {
	case constant.StatusInstalling:
		return constant.StatusInstallErr
	case constant.StatusUpgrading:
		return constant.StatusUpgradeErr
	case constant.StatusUninstalling:
		return constant.StatusError
	default:
		return constant.StatusUpErr
	}
}

func appInstallOperationPending(status string) bool {
	switch status {
	case constant.StatusInstalling, constant.StatusRebuilding, constant.StatusUpgrading, constant.StatusUninstalling,
		constant.StatusStarting, constant.StatusRestarting, constant.StatusWaiting:
		return true
	default:
		return false
	}
}

func appInstallOperationInterruptedOnRestart(status string) bool {
	// ReStarting can also describe a container's Docker restart policy.
	return appInstallOperationPending(status) && status != constant.StatusRestarting
}

func reconcileAppInstallTaskFailure(install *model.AppInstall, operation string, appDB, taskDB *gorm.DB) error {
	if !appInstallOperationPending(install.Status) {
		return nil
	}
	var operationTask model.Task
	err := taskDB.Where("type = ? AND resource_id = ? AND operate = ?", "App", install.ID, operation).
		Order("created_at DESC").First(&operationTask).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if operationTask.Status != constant.StatusFailed && operationTask.Status != constant.StatusCanceled {
		return nil
	}
	// A previous operation must not turn a newly queued retry into a failure.
	if operationTask.CreatedAt.Before(install.UpdatedAt) &&
		(operationTask.EndAt.IsZero() || operationTask.EndAt.Before(install.UpdatedAt)) {
		return nil
	}
	message := operationTask.ErrorMsg
	if message == "" {
		message = "the application operation task failed"
	}
	status := appInstallOperationFailureStatus(install.Status)
	result := appDB.Model(&model.AppInstall{}).
		Where("id = ? AND status = ? AND updated_at = ?", install.ID, install.Status, install.UpdatedAt).
		Updates(map[string]interface{}{"status": status, "message": message})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		install.Status = status
		install.Message = message
	}
	return nil
}
