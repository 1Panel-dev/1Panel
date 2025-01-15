package dto

type CreateOrUpdateAlert struct {
	AlertTitle string `json:"alertTitle"`
	AlertType  string `json:"alertType"`
	AlertCount uint   `json:"alertCount"`
	EntryID    uint   `json:"entryID"`
}

type AlertBase struct {
	AlertType string `json:"alertType"`
	EntryID   uint   `json:"entryID"`
}

type PushAlert struct {
	TaskName  string `json:"taskName"`
	AlertType string `json:"alertType"`
	EntryID   uint   `json:"entryID"`
	Param     string `json:"param"`
}
