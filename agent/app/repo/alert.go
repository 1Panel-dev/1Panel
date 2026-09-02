package repo

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/google/uuid"
	"google.golang.org/genproto/googleapis/type/date"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AlertRepo struct{}

var (
	ErrAlertConfigRevisionConflict = errors.New("alert config revision conflict")
	ErrAlertConfigRevisionRequired = errors.New("alert config revision is required")
)

type IAlertRepo interface {
	WithByType(alertType string) DBOption
	WithByStatusIn(status []string) DBOption
	WithByProject(project string) DBOption
	WithByCount(count uint) DBOption
	WithByAlertId(alertId uint) DBOption
	WithByCreateAt(date *date.Date) DBOption
	WithByLicenseId(licenseId string) DBOption
	WithByRecordId(recordId uint) DBOption
	WithByDeliveryLogID(logID uint) DBOption
	WithByAlertMethodContainsConfigID(id uint) DBOption
	WithByMethodConfigIDs(ids []uint) DBOption

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
	CreatePendingAlertTask(logID, alertID uint, alertTask *model.AlertTask) (bool, error)
	FinalizePendingAlertTask(logID uint, succeeded bool, message string, fallback *model.AlertTask) (bool, error)
	DeleteAlertTask(opts ...DBOption) error
	GetAlertTask(opts ...DBOption) (model.AlertTask, error)
	LoadTaskCount(alertType string, project string, method string) (uint, uint, error)
	GetTaskLog(alertType string, alertId uint) (time.Time, error)
	GetLicensePushCount(method string) (uint, error)

	GetConfig(opts ...DBOption) (model.AlertConfig, error)
	GetConfigById(id uint) (model.AlertConfig, error)
	AlertConfigList(opts ...DBOption) ([]model.AlertConfig, error)
	UpdateAlertConfig(maps map[string]interface{}, opts ...DBOption) error
	UpdateAlertConfigWithRevision(maps map[string]interface{}, revision *time.Time, opts ...DBOption) error
	CreateAlertConfig(config *model.AlertConfig) error
	DeleteAlertConfig(opts ...DBOption) error

	WithByTypeNotIn(types []string) DBOption
	PageAlertConfig(page, size int, opts ...DBOption) (int64, []model.AlertConfig, error)

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

func (a *AlertRepo) WithByAlertMethodContainsConfigID(id uint) DBOption {
	method := strconv.Itoa(int(id))
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("(method = ? OR method LIKE ? OR method LIKE ? OR method LIKE ?)", method, method+",%", "%,"+method, "%,"+method+",%")
	}
}

func (a *AlertRepo) WithByMethodConfigIDs(ids []uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		methods := make([]string, 0, len(ids))
		for _, id := range ids {
			methods = append(methods, strconv.Itoa(int(id)))
		}
		return g.Where("method IN ?", methods)
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
	return global.AlertDB.Where("status <> ?", constant.AlertPushing).Delete(&model.AlertLog{}).Error
}

func (a *AlertRepo) CreateAlertTask(alertTaskBase *model.AlertTask) error {
	return global.AlertDB.Model(&model.AlertTask{}).Create(&alertTaskBase).Error
}

func (a *AlertRepo) CreatePendingAlertTask(logID, alertID uint, alertTask *model.AlertTask) (bool, error) {
	if alertTask == nil {
		return false, fmt.Errorf("pending alert task is required")
	}
	created := false
	err := global.AlertDB.Transaction(func(tx *gorm.DB) error {
		var log model.AlertLog
		if err := tx.Where("id = ? AND status = ?", logID, constant.AlertPushing).First(&log).Error; err != nil {
			return err
		}
		if log.AlertId != alertID || log.Type != alertTask.Type || log.Method != alertTask.Method {
			return fmt.Errorf("pending alert task does not match delivery log %d", logID)
		}
		alertTask.DeliveryLogID = &logID
		result := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "delivery_log_id"}},
			DoNothing: true,
		}).Create(alertTask)
		if result.Error != nil {
			return result.Error
		}
		created = result.RowsAffected > 0
		return nil
	})
	return created, err
}

func (a *AlertRepo) FinalizePendingAlertTask(logID uint, succeeded bool, message string, fallback *model.AlertTask) (bool, error) {
	finalized := false
	err := global.AlertDB.Transaction(func(tx *gorm.DB) error {
		status := constant.AlertError
		if succeeded {
			status = constant.AlertSuccess
			message = ""
		}
		result := tx.Model(&model.AlertLog{}).
			Where("id = ? AND status = ?", logID, constant.AlertPushing).
			Updates(map[string]interface{}{"status": status, "message": message})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		finalized = true
		if !succeeded {
			return tx.Where("delivery_log_id = ?", logID).Delete(&model.AlertTask{}).Error
		}

		var count int64
		if err := tx.Model(&model.AlertTask{}).Where("delivery_log_id = ?", logID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		if fallback == nil {
			return fmt.Errorf("pending alert task metadata is unavailable for delivery log %d", logID)
		}
		fallback.DeliveryLogID = &logID
		return tx.Create(fallback).Error
	})
	return finalized, err
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

func (a *AlertRepo) UpdateAlertConfigWithRevision(maps map[string]interface{}, revision *time.Time, opts ...DBOption) error {
	if revision == nil {
		return a.UpdateAlertConfig(maps, opts...)
	}
	db, _ := getAlertDB(opts...)
	result := db.Model(&model.AlertConfig{}).Where("updated_at = ?", *revision).Updates(maps)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAlertConfigRevisionConflict
	}
	return nil
}

func (a *AlertRepo) CreateAlertConfig(config *model.AlertConfig) error {
	ensureAlertConfigUID(config)
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

func (a *AlertRepo) GetConfigById(id uint) (model.AlertConfig, error) {
	var config model.AlertConfig
	err := global.AlertDB.First(&config, id).Error
	return config, err
}

func (a *AlertRepo) WithByTypeNotIn(types []string) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("`type` NOT IN (?)", types)
	}
}

func (a *AlertRepo) WithByDeliveryLogID(logID uint) DBOption {
	return func(g *gorm.DB) *gorm.DB {
		return g.Where("delivery_log_id = ?", logID)
	}
}

func (a *AlertRepo) PageAlertConfig(page, size int, opts ...DBOption) (int64, []model.AlertConfig, error) {
	var configs []model.AlertConfig
	db := global.AlertDB.Model(&model.AlertConfig{})
	for _, opt := range opts {
		db = opt(db)
	}
	count := int64(0)
	db = db.Count(&count)
	err := db.Limit(size).Offset(size * (page - 1)).Find(&configs).Error
	return count, configs, err
}

var singletonTypes = map[string]bool{
	constant.CommonConfig: true,
}

func (a *AlertRepo) SyncAll(data []model.AlertConfig) error {
	tx := global.AlertDB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			panic(r)
		}
	}()

	var oldConfigs []model.AlertConfig
	if err := tx.Find(&oldConfigs).Error; err != nil {
		tx.Rollback()
		return err
	}

	usedConfigIDs, err := loadUsedAlertConfigIDs(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	oldConfigMap := make(map[string]model.AlertConfig)
	oldConfigByUID := make(map[string]model.AlertConfig)
	oldConfigByType := make(map[string][]model.AlertConfig)
	oldConfigByKey := make(map[string][]model.AlertConfig)
	consumedConfigIDs := make(map[uint]struct{})
	for _, item := range oldConfigs {
		if strings.TrimSpace(item.UID) != "" {
			oldConfigByUID[item.UID] = item
		}
		if singletonTypes[item.Type] {
			oldConfigMap[item.Type] = item
			continue
		}
		oldConfigByType[item.Type] = append(oldConfigByType[item.Type], item)
		oldConfigByKey[alertConfigSyncKey(item)] = append(oldConfigByKey[alertConfigSyncKey(item)], item)
	}
	for _, item := range data {
		if uid := strings.TrimSpace(item.UID); uid != "" {
			if matched, ok := oldConfigByUID[uid]; ok && matched.Type != item.Type {
				tx.Rollback()
				return fmt.Errorf("alert config UID %q belongs to type %q, not %q", uid, matched.Type, item.Type)
			}
		}
		if singletonTypes[item.Type] {
			if matched, ok := oldConfigMap[item.Type]; ok {
				if err := inheritAlertConfigSyncState(&item, matched); err != nil {
					tx.Rollback()
					return err
				}
				delete(oldConfigMap, item.Type)
				consumedConfigIDs[item.ID] = struct{}{}
			} else {
				item.ID = 0
				ensureAlertConfigUID(&item)
				if err := validateAlertConfigSyncSecret(&item); err != nil {
					tx.Rollback()
					return err
				}
			}
			if item.ID == 0 {
				if err := tx.Create(&item).Error; err != nil {
					tx.Rollback()
					return err
				}
			} else if err := tx.Save(&item).Error; err != nil {
				tx.Rollback()
				return err
			}
			continue
		}

		if strings.TrimSpace(item.UID) != "" {
			if matched, ok := oldConfigByUID[item.UID]; ok {
				delete(oldConfigByUID, item.UID)
				if err := inheritAlertConfigSyncState(&item, matched); err != nil {
					tx.Rollback()
					return err
				}
				consumedConfigIDs[item.ID] = struct{}{}
				if err := tx.Save(&item).Error; err != nil {
					tx.Rollback()
					return err
				}
				deleteAlertConfigByID(oldConfigByType, matched.ID)
				deleteAlertConfigByID(oldConfigByKey, matched.ID)
				continue
			}
		}

		key := alertConfigSyncKey(item)
		if matched, ok := popAlertConfigByKey(oldConfigByKey, key); ok {
			delete(oldConfigByUID, matched.UID)
			if err := inheritAlertConfigSyncState(&item, matched); err != nil {
				tx.Rollback()
				return err
			}
			consumedConfigIDs[item.ID] = struct{}{}
			if err := tx.Save(&item).Error; err != nil {
				tx.Rollback()
				return err
			}
			deleteAlertConfigByID(oldConfigByType, matched.ID)
			continue
		}

		if matched, ok := popUnusedAlertConfigByType(oldConfigByType, usedConfigIDs, item.Type); ok {
			delete(oldConfigByUID, matched.UID)
			deleteAlertConfigByID(oldConfigByKey, matched.ID)
			if err := inheritAlertConfigSyncState(&item, matched); err != nil {
				tx.Rollback()
				return err
			}
			consumedConfigIDs[item.ID] = struct{}{}
			if err := tx.Save(&item).Error; err != nil {
				tx.Rollback()
				return err
			}
			continue
		}

		item.ID = 0
		ensureAlertConfigUID(&item)
		if err := validateAlertConfigSyncSecret(&item); err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	for _, item := range oldConfigs {
		if _, used := usedConfigIDs[item.ID]; used {
			continue
		}
		if _, kept := consumedConfigIDs[item.ID]; kept {
			continue
		}
		if err := tx.Where("id = ?", item.ID).Delete(&model.AlertConfig{}).Error; err != nil {
			tx.Rollback()
			return err
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func ensureAlertConfigUID(config *model.AlertConfig) {
	if config != nil && strings.TrimSpace(config.UID) == "" {
		config.UID = uuid.NewString()
	}
}

func inheritAlertConfigSyncState(incoming *model.AlertConfig, existing model.AlertConfig) error {
	if incoming.Type != existing.Type {
		return fmt.Errorf("alert config UID %q belongs to type %q, not %q", incoming.UID, existing.Type, incoming.Type)
	}
	preserveExistingCustom := incoming.Type == constant.Custom &&
		existing.Status == constant.AlertDisable &&
		incoming.Title == existing.Title &&
		incoming.Status == existing.Status &&
		incoming.Config == existing.Config &&
		(incoming.SecretConfig == "" || incoming.SecretConfig == existing.SecretConfig)
	incoming.ID = existing.ID
	if strings.TrimSpace(incoming.UID) == "" {
		incoming.UID = existing.UID
	}
	if incoming.Type == constant.Custom && incoming.SecretConfig == "" {
		incoming.SecretConfig = existing.SecretConfig
	}
	if preserveExistingCustom {
		return nil
	}
	return validateAlertConfigSyncSecret(incoming)
}

func validateAlertConfigSyncSecret(incoming *model.AlertConfig) error {
	if incoming.Type != constant.Custom {
		incoming.SecretConfig = ""
		return nil
	}
	if strings.TrimSpace(incoming.SecretConfig) == "" {
		return fmt.Errorf("custom webhook sync secret is missing")
	}
	var version struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if err := json.Unmarshal([]byte(incoming.Config), &version); err != nil || version.SchemaVersion != 1 {
		return fmt.Errorf("custom webhook sync config must use schemaVersion 1")
	}
	secret := incoming.SecretConfig
	for _, prefix := range []string{"core:v1:", "agent:v1:"} {
		if !strings.HasPrefix(secret, prefix) {
			continue
		}
		ciphertext, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(secret, prefix))
		if err != nil || len(ciphertext) < 32 || len(ciphertext)%16 != 0 {
			return fmt.Errorf("custom webhook sync secret envelope is invalid")
		}
		return nil
	}
	return fmt.Errorf("custom webhook sync secret must use a versioned envelope")
}

func loadUsedAlertConfigIDs(tx *gorm.DB) (map[uint]struct{}, error) {
	var alerts []model.Alert
	if err := tx.Select("method").Find(&alerts).Error; err != nil {
		return nil, err
	}

	usedIDs := make(map[uint]struct{})
	for _, alert := range alerts {
		for _, item := range strings.Split(alert.Method, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			id, err := strconv.ParseUint(item, 10, 64)
			if err != nil {
				continue
			}
			usedIDs[uint(id)] = struct{}{}
		}
	}
	return usedIDs, nil
}

func alertConfigSyncKey(item model.AlertConfig) string {
	return item.Type + "::" + normalizeAlertConfigJSON(item.Config)
}

func normalizeAlertConfigJSON(config string) string {
	trimmed := strings.TrimSpace(config)
	if trimmed == "" {
		return ""
	}

	var data any
	if err := json.Unmarshal([]byte(trimmed), &data); err != nil {
		return trimmed
	}
	buf, err := json.Marshal(data)
	if err != nil {
		return trimmed
	}
	return string(buf)
}

func popAlertConfigByKey(configMap map[string][]model.AlertConfig, key string) (model.AlertConfig, bool) {
	items := configMap[key]
	if len(items) == 0 {
		return model.AlertConfig{}, false
	}

	item := items[0]
	if len(items) == 1 {
		delete(configMap, key)
	} else {
		configMap[key] = items[1:]
	}
	return item, true
}

func popUnusedAlertConfigByType(configMap map[string][]model.AlertConfig, usedConfigIDs map[uint]struct{}, configType string) (model.AlertConfig, bool) {
	items := configMap[configType]
	if len(items) == 0 {
		return model.AlertConfig{}, false
	}

	for idx, item := range items {
		if _, used := usedConfigIDs[item.ID]; used {
			continue
		}
		if idx == 0 {
			if len(items) == 1 {
				delete(configMap, configType)
			} else {
				configMap[configType] = items[1:]
			}
		} else {
			configMap[configType] = append(items[:idx], items[idx+1:]...)
		}
		return item, true
	}
	return model.AlertConfig{}, false
}

func deleteAlertConfigByID(configMap map[string][]model.AlertConfig, id uint) {
	for key, items := range configMap {
		for idx, item := range items {
			if item.ID != id {
				continue
			}
			if len(items) == 1 {
				delete(configMap, key)
			} else {
				configMap[key] = append(items[:idx], items[idx+1:]...)
			}
			return
		}
	}
}
