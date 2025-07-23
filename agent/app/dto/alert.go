package dto

import (
	"encoding/json"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"time"
)

type CreateOrUpdateAlert struct {
	AlertTitle  string `json:"alertTitle"`
	AlertType   string `json:"alertType"`
	AlertMethod string `json:"alertMethod"`
	AlertCount  uint   `json:"alertCount"`
	EntryID     uint   `json:"entryID"`
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

type AlertSearch struct {
	PageInfo
	OrderBy string `json:"orderBy" validate:"required,oneof=created_at"`
	Order   string `json:"order" validate:"required,oneof=null ascending descending"`
	Type    string `json:"type"`
	Status  string `json:"status"`
	Method  string `json:"method"`
}

type AlertDTO struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Cycle     uint      `json:"cycle"`
	Count     uint      `json:"count"`
	Method    string    `json:"method"`
	Title     string    `json:"title"`
	Project   string    `json:"project"`
	Status    string    `json:"status"`
	SendCount uint      `json:"sendCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AlertCreate struct {
	Type      string `json:"type" validate:"required"`
	Cycle     uint   `json:"cycle"`
	Count     uint   `json:"count"`
	Method    string `json:"method" validate:"required"`
	Title     string `json:"title"`
	Project   string `json:"project"`
	Status    string `json:"status"`
	SendCount uint   `json:"sendCount"`
}

type AlertUpdate struct {
	ID        uint   `json:"id" validate:"required"`
	Type      string `json:"type"`
	Cycle     uint   `json:"cycle"`
	Count     uint   `json:"count"`
	Method    string `json:"method"`
	Title     string `json:"title"`
	Project   string `json:"project"`
	Status    string `json:"status"`
	SendCount uint   `json:"sendCount"`
}

type DeleteRequest struct {
	ID uint `json:"id" validate:"required"`
}

type AlertUpdateStatus struct {
	ID     uint   `json:"id" validate:"required"`
	Status string `json:"status" validate:"required"`
}

type DiskDTO struct {
	Path        string  `json:"path"`
	Type        string  `json:"type"`
	Device      string  `json:"device"`
	Total       uint64  `json:"total"`
	Free        uint64  `json:"free"`
	Used        uint64  `json:"used"`
	UsedPercent float64 `json:"usedPercent"`

	InodesTotal       uint64  `json:"inodesTotal"`
	InodesUsed        uint64  `json:"inodesUsed"`
	InodesFree        uint64  `json:"inodesFree"`
	InodesUsedPercent float64 `json:"inodesUsedPercent"`
}

type AlertLogSearch struct {
	PageInfo
	Count  uint   `json:"count"`
	Status string `json:"status"`
}

type AlertLogDTO struct {
	ID          uint        `json:"id"`
	Type        string      `json:"type"`
	Count       uint        `json:"count"`
	AlertId     uint        `json:"alertId"`
	AlertDetail AlertDetail `json:"alertDetail"`
	AlertRule   AlertRule   `json:"alertRule"`
	Status      string      `json:"status"`
	Method      string      `json:"method"`
	Message     string      `json:"message"`
	CreatedAt   time.Time   `json:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt"`
}

type AlertLogCreate struct {
	Type        string `json:"type" validate:"required"`
	Count       uint   `json:"count" validate:"required"`
	AlertId     uint   `json:"alertId" validate:"required"`
	AlertDetail string `json:"alertDetail" validate:"required"`
	AlertRule   string `json:"alertRule" validate:"required"`
	Status      string `json:"status" validate:"required"`
	Method      string `json:"method" validate:"required"`
	Message     string `json:"message"`
	RecordId    uint   `json:"recordId"`
	LicenseId   string `json:"licenseId" validate:"required"`
}

type AlertLog struct {
	ID uint `json:"id" validate:"required"`
}

type AlertDetail struct {
	LicenseId   string  `json:"licenseId"`
	Type        string  `json:"type"`
	SubType     string  `json:"subType"`
	Title       string  `json:"title"`
	Method      string  `json:"method"`
	LicenseCode string  `json:"licenseCode"`
	DeviceId    string  `json:"deviceId"`
	Project     string  `json:"project"`
	Params      []Param `json:"params"`
	Phone       string  `json:"phone"`
}

type AlertRule struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	Cycle     uint      `json:"cycle"`
	Count     uint      `json:"count"`
	Method    string    `json:"method"`
	Title     string    `json:"title"`
	Project   string    `json:"project"`
	Status    string    `json:"status"`
	SendCount uint      `json:"sendCount"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type Param struct {
	Index string `json:"index"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ClamDTO struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Path      string    `json:"path"`
}

type CronJobDTO struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	Type      string    `json:"type"`
}

type CronJobReq struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

type UpgradeInfo struct {
	TestVersion   string `json:"testVersion"`
	NewVersion    string `json:"newVersion"`
	LatestVersion string `json:"latestVersion"`
	ReleaseNote   string `json:"releaseNote"`
}

type AlertDiskInfo struct {
	Type   string
	Mount  string
	Device string
}

type SyncResult struct {
	ID          uint   `json:"id"`
	LicenseID   string `json:"licenseId"`
	SendStatus  string `json:"sendStatus"`
	CreateTime  string `json:"createTime"`
	UpdateTime  string `json:"updateTime"`
	Remarks     string `json:"remarks"`
	MsgCount    int    `json:"msgCount"`
	MsgCountMax int    `json:"msgCountMax"`
}

type QueryRequest struct {
	QueryIds  []uint `json:"queryIds"`
	LicenseId string `json:"licenseId"`
}

type AlertResponse struct {
	Result  bool            `json:"result"`
	Data    json.RawMessage `json:"data"`
	Message string          `json:"message"`
}

type PushResult struct {
	RecordId  uint   `json:"recordId"`
	LicenseId string `json:"licenseId"`
}

type UpdateOfflineAlertLog struct {
	ID        uint   `json:"id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	RecordId  uint   `json:"recordId"`
	LicenseId string `json:"licenseId"`
}

type SyncOfflineAlertLogDTO struct {
	QueryRequest string           `json:"queryRequest"`
	Ids          []uint           `json:"ids"`
	AlertLogs    []model.AlertLog `json:"alertLogs"`
	LicenseId    string           `json:"licenseId"`
}

type OfflineAlertResponse struct {
	ID             uint             `json:"id"`
	Type           string           `json:"type"`
	RemoteErr      error            `json:"remoteErr"`
	ResponseStruct AlertResponse    `json:"responseStruct"`
	AlertLogs      []model.AlertLog `json:"alertLogs"`
	Ids            []uint           `json:"ids"`
	LicenseId      string           `json:"licenseId"`
}

type OfflineAlertLogDTO struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Count       uint      `json:"count"`
	AlertId     uint      `json:"alertId"`
	AlertDetail string    `json:"alertDetail"`
	AlertRule   string    `json:"alertRule"`
	Status      string    `json:"status"`
	Method      string    `json:"method"`
	Message     string    `json:"message"`
	RecordId    uint      `json:"recordId"`
	LicenseId   string    `json:"licenseId"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type OfflineQueryRequest struct {
	Ids       []uint `json:"ids"`
	QueryIds  []uint `json:"queryIds"`
	LicenseId string `json:"licenseId"`
}

type AlertConfigUpdate struct {
	ID     uint   `json:"id"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Config string `json:"config"`
}

type AlertConfigTest struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Sender      string `json:"sender"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Encryption  string `json:"encryption"` // "ssl" / "tls" / "none"
	Recipient   string `json:"recipient"`
}

type AlertSendTimeRange struct {
	NoticeAlert struct {
		SendTimeRange string   `json:"sendTimeRange"`
		Type          []string `json:"type"`
	} `json:"noticeAlert"`
	ResourceAlert struct {
		SendTimeRange string   `json:"sendTimeRange"`
		Type          []string `json:"type"`
	} `json:"resourceAlert"`
}

type AlertCommonConfig struct {
	IsOffline          string             `json:"isOffline"`
	AlertSendTimeRange AlertSendTimeRange `json:"alertSendTimeRange"`
}

type AlertSmsConfig struct {
	Phone         string `json:"phone"`
	AlertDailyNum uint   `json:"alertDailyNum"`
}

type AlertEmailConfig struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Sender      string `json:"sender"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	Encryption  string `json:"encryption"` // "ssl" / "tls" / "none"
	Recipient   string `json:"recipient"`
}
