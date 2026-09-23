package dto

type NetworkCleanupReport struct {
	Deleted []NetworkCleanupItem `json:"deleted"`
	Skipped []NetworkCleanupItem `json:"skipped"`
	Failed  []NetworkCleanupItem `json:"failed"`
}

type NetworkCleanupItem struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Reason string `json:"reason,omitempty"`
}

type NetworkCleanupTask struct {
	TaskID string `json:"taskID"`
}
