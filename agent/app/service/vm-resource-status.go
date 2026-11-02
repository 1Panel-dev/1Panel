package service

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/constant"
)

const vmResourceCreating = "creating"

func vmResourceCreationIncomplete(status string) bool {
	return vmResourceStatusEquals(status, vmResourceCreating) || vmResourceStatusEquals(status, constant.StatusFailed)
}

func normalizeVMResourceStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return constant.StatusUnknown
	}
	switch strings.ToLower(status) {
	case vmResourceCreating:
		return vmResourceCreating
	case strings.ToLower(constant.StatusFailed):
		return constant.StatusFailed
	case strings.ToLower(constant.StatusRunning):
		return constant.StatusRunning
	case strings.ToLower(constant.StatusInactive):
		return constant.StatusInactive
	case strings.ToLower(constant.StatusDeleted):
		return constant.StatusDeleted
	case strings.ToLower(constant.StatusUnknown):
		return constant.StatusUnknown
	case strings.ToLower(constant.StatusBuilding):
		return constant.StatusBuilding
	case strings.ToLower(constant.StatusDegraded):
		return constant.StatusDegraded
	case strings.ToLower(constant.StatusInaccessible):
		return constant.StatusInaccessible
	default:
		return status
	}
}

func vmResourceStatusEquals(status string, target string) bool {
	return strings.EqualFold(strings.TrimSpace(status), target)
}

func vmResourceExists(status string) bool {
	status = normalizeVMResourceStatus(status)
	return !vmResourceStatusEquals(status, constant.StatusDeleted)
}

func validVMResourceOperateStatus(status string) bool {
	return vmResourceStatusEquals(status, constant.StatusRunning) || vmResourceStatusEquals(status, constant.StatusInactive)
}
