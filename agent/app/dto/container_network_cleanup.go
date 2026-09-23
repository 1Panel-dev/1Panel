package dto

// NetworkCleanupReport records partial success; failed entries were not deleted.
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

// NetworkCleanupTask identifies an asynchronous task and its log.
type NetworkCleanupTask struct {
	TaskID string `json:"taskID"`
}
