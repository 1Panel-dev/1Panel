package repo

import (
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"google.golang.org/genproto/googleapis/type/date"
	"gorm.io/gorm"
	"time"
)

type AlertRepo struct{}

type IAlertRepo interface {
	WithByType(alertType string) DBOption
	WithByStatusIn(status []string) DBOption
	WithByProject(project string) DBOption
	WithByCount(count uint) DBOption
	WithByAlertId(alertId uint) DBOption
	WithByCreateAt(date *date.Date) DBOption
	WithByLicenseId(licenseId string) DBOption
	WithByRecordId(recordId uint) DBOption
	WithByMethod(method string) DBOption

	Create(alert *model.Alert) error
	Get(opts ...DBOption) (model.Alert, error)
	Page(page, size int, opts ...DBOption) (int64, []model.Alert, error)
	List(opts ...DBOption) ([]model.Alert, error)
	Delete(opts ...DBOption) error
	Save(alert *model.Alert) error
	Update(maps map[string]interface{}, opts ...DBOption) error

	GetLog(opts ...DBOption) (model.AlertLog, error)
	CreateLog(alertLog *model.AlertLog) error
	PageLog(limit, offset int, opts ...DBOption) (int64, []model.AlertLog, error)
	ListLog(opts ...DBOption) ([]model.AlertLog, error)
	UpdateLog(id uint, maps map[string]interface{}) error
	BatchUpdateLogBy(maps map[string]interface{}, opts ...DBOption) error
	DeleteLog(opts ...DBOption) error
	CleanAlertLogs() error

	CreateAlertTask(alertTaskBase *model.AlertTask) error
	DeleteAlertTask(opts ...DBOption) error
	GetAlertTask(opts ...DBOption) (model.AlertTask, error)
	LoadTaskCount(alertType string, project string, method string) (uint, uint, error)
	GetTaskLog(alertType string, alertId uint) (time.Time, error)
	GetLicensePushCount(method string) (uint, error)

	GetConfig(opts ...DBOption) (model.AlertConfig, error)
	AlertConfigList(opts ...DBOption) ([]model.AlertConfig, error)
	UpdateAlertConfig(maps map[string]interface{}, opts ...DBOption) error
	CreateAlertConfig(config *model.AlertConfig) error
	DeleteAlertConfig(opts ...DBOption) error

	SyncAll(data []model.AlertConfig) error
}

func NewIAlertRepo() IAlertRepo {
	return &AlertRepo{}
}

func (a *AlertRepo) WithByType(alertType string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`type` = ?", alertType)
	}
}

func (a *AlertRepo) WithByStatusIn(status []string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("status in (?)", status)
	}
}

func (a *AlertRepo) WithByCount(count uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("count = ?", count)
	}
}

func (a *AlertRepo) WithByProject(project string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("project = ?", project)
	}
}

func (a *AlertRepo) WithByAlertId(alertId uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("alert_id = ?", alertId)
	}
}

func (a *AlertRepo) WithByLicenseId(licenseId string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("license_id = ?", licenseId)
	}
}

func (a *AlertRepo) WithByRecordId(recordId uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("record_id = ?", recordId)
	}
}

func (a *AlertRepo) WithByMethod(method string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("method = ?", method)
	}
}

func (a *AlertRepo) WithByCreateAt(createAt *date.Date) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("DATE(created_at) = DATE(?)", createAt)
	}
}

func (a *AlertRepo) Create(alert *model.Alert) error {
	return global.AlertDB.Model(&model.Alert{}).Create(alert).Error
}

func (a *AlertRepo) Save(alert *model.Alert) error {
	return global.AlertDB.Save(alert).Error
}

func (a *AlertRepo) Get(opts ...DBOption) (model.Alert, error) {
	var alert model.Alert
	db, _ := getAlertDB(opts...)
	err := db.First(&alert).Error
	return alert, err
}

func (a *AlertRepo) Page(page, size int, opts ...DBOption) (int64, []model.Alert, error) {
	var alerts []model.Alert
	alertDb, _ := getAlertDB(opts...)
	db := alertDb.Model(&model.Alert{})
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&alerts).Error
	return count, alerts, err
}

func (a *AlertRepo) List(opts ...DBOption) ([]model.Alert, error) {
	var alert []model.Alert
	db, _ := getAlertDB(opts...)
	err := db.Find(&alert).Error
	return alert, err
}

func (a *AlertRepo) Update(maps map[string]interface{}, opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	return db.Model(&model.Alert{}).Updates(maps).Error
}

func (a *AlertRepo) Delete(opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	return db.Delete(&model.Alert{}).Error
}

func (a *AlertRepo) GetLog(opts ...DBOption) (model.AlertLog, error) {
	var alertLog model.AlertLog
	db, _ := getAlertDB(opts...)
	err := db.First(&alertLog).Error
	return alertLog, err
}

func (a *AlertRepo) CreateLog(log *model.AlertLog) error {
	return global.AlertDB.Model(&model.AlertLog{}).Create(&log).Error
}

func (a *AlertRepo) UpdateLog(id uint, maps map[string]interface{}) error {
	return global.AlertDB.Model(&model.AlertLog{}).Where("id = ?", id).Updates(maps).Error
}

func (a *AlertRepo) BatchUpdateLogBy(maps map[string]interface{}, opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	if len(opts) == 0 {
		db = db.Where("1=1")
	}
	return db.Model(&model.AlertLog{}).Updates(&maps).Error
}

func (a *AlertRepo) PageLog(page, size int, opts ...DBOption) (int64, []model.AlertLog, error) {
	var alerts []model.AlertLog
	db := global.AlertDB.Model(&model.AlertLog{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Order("created_at desc").Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&alerts).Error
	return count, alerts, err
}

func (a *AlertRepo) ListLog(opts ...DBOption) ([]model.AlertLog, error) {
	var alertLog []model.AlertLog
	db, _ := getAlertDB(opts...)
	err := db.Find(&alertLog).Error
	return alertLog, err
}

func (a *AlertRepo) DeleteLog(opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	return db.Delete(&model.AlertLog{}).Error
}

func (a *AlertRepo) CleanAlertLogs() error {
	return global.AlertDB.Where("1 = 1").Delete(&model.AlertLog{}).Error
}

func (a *AlertRepo) CreateAlertTask(alertTaskBase *model.AlertTask) error {
	return global.AlertDB.Model(&model.AlertTask{}).Create(&alertTaskBase).Error
}

func (a *AlertRepo) DeleteAlertTask(opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	return db.Delete(&model.AlertTask{}).Error
}

func (a *AlertRepo) GetAlertTask(opts ...DBOption) (model.AlertTask, error) {
	var data model.AlertTask
	db, _ := getAlertDB(opts...)
	err := db.First(&data).Error
	return data, err
}

func (a *AlertRepo) LoadTaskCount(alertType string, project string, method string) (uint, uint, error) {
	var (
		todayCount int64
		totalCount int64
	)
	_ = global.AlertDB.Model(&model.AlertTask{}).Where("type = ? AND quota_type = ? AND method = ?", alertType, project, method).Count(&totalCount).Error

	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowMidnight := todayMidnight.Add(24 * time.Hour)
	err := global.AlertDB.Model(&model.AlertTask{}).Where("type =  ? AND quota_type = ?  AND method = ? AND created_at > ? AND created_at < ?", alertType, project, method, todayMidnight, tomorrowMidnight).Count(&todayCount).Error
	return uint(todayCount), uint(totalCount), err
}

func (a *AlertRepo) GetTaskLog(alertType string, alertId uint) (time.Time, error) {
	var newDate time.Time
	status := []string{constant.AlertSuccess, constant.AlertPushSuccess, constant.AlertSyncError, constant.AlertPushing}
	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowMidnight := todayMidnight.Add(24 * time.Hour)
	err := global.AlertDB.Model(&model.AlertLog{}).
		Where("type = ? AND alert_id = ? AND status in ? AND created_at > ? AND created_at < ?", alertType, alertId, status, todayMidnight, tomorrowMidnight).
		Order("created_at DESC").
		Limit(1).
		Pluck("created_at", &newDate).Error
	if err != nil {
		return time.Time{}, err
	}

	if newDate.IsZero() {
		return time.Time{}, nil
	}

	return newDate, nil
}

func getAlertDB(opts ...DBOption) (*gorm.DB, error) {
	var db *gorm.DB
	db = global.AlertDB
	for _, opt := range opts {
		db = opt(db)
	}
	return db, nil
}

func (a *AlertRepo) GetLicensePushCount(method string) (uint, error) {
	var (
		todayCount int64
	)
	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowMidnight := todayMidnight.Add(24 * time.Hour)
	err := global.AlertDB.Model(&model.AlertTask{}).Where("created_at > ? AND created_at < ? AND method = ?", todayMidnight, tomorrowMidnight, method).Count(&todayCount).Error
	return uint(todayCount), err
}

func (a *AlertRepo) AlertConfigList(opts ...DBOption) ([]model.AlertConfig, error) {
	var config []model.AlertConfig
	db, _ := getAlertDB(opts...)
	err := db.Find(&config).Error
	return config, err
}

func (a *AlertRepo) UpdateAlertConfig(maps map[string]interface{}, opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	return db.Model(&model.AlertConfig{}).Updates(maps).Error
}

func (a *AlertRepo) CreateAlertConfig(config *model.AlertConfig) error {
	return global.AlertDB.Model(&model.AlertConfig{}).Create(config).Error
}

func (a *AlertRepo) DeleteAlertConfig(opts ...DBOption) error {
	db, _ := getAlertDB(opts...)
	return db.Delete(&model.AlertConfig{}).Error
}

func (a *AlertRepo) GetConfig(opts ...DBOption) (model.AlertConfig, error) {
	var alertConfig model.AlertConfig
	db, _ := getAlertDB(opts...)
	err := db.First(&alertConfig).Error
	return alertConfig, err
}

func (a *AlertRepo) SyncAll(data []model.AlertConfig) error {
	tx := global.AlertDB.Begin()
	var oldConfigs []model.AlertConfig
	_ = tx.Find(&oldConfigs).Error
	oldConfigMap := make(map[string]uint)
	for _, item := range oldConfigs {
		oldConfigMap[item.Type] = item.ID
	}
	for _, item := range data {
		if val, ok := oldConfigMap[item.Type]; ok {
			item.ID = val
			delete(oldConfigMap, item.Type)
		} else {
			item.ID = 0
		}
		if err := tx.Model(model.AlertConfig{}).Where("id = ?", item.ID).Save(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, val := range oldConfigMap {
		if err := tx.Where("id = ?", val).Delete(&model.AlertConfig{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	tx.Commit()
	return nil
}
