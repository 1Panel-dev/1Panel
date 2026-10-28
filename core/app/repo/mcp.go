package repo

import (
	"time"

	"github.com/1Panel-dev/1Panel/core/app/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const mcpInvocationSecond = "CAST(strftime('%s', substr(started_at, 1, 19) || substr(started_at, -6)) AS INTEGER)"
const mcpInvocationNanosecond = "CASE WHEN substr(started_at, 20, 1) = '.' THEN CAST(substr(substr(started_at, 21, length(started_at) - 26) || '000000000', 1, 9) AS INTEGER) ELSE 0 END"

type MCPRepo struct{ DB *gorm.DB }

func mcpOwnerQuery(db *gorm.DB, ownerType, ownerID string) *gorm.DB {
	return db.Where("owner_type = ? AND owner_id = ?", ownerType, ownerID)
}

func (r MCPRepo) Credential(ownerType, ownerID, id string) (model.MCPCredential, error) {
	var key model.MCPCredential
	err := mcpOwnerQuery(r.DB, ownerType, ownerID).Where("id = ?", id).First(&key).Error
	return key, err
}

func (r MCPRepo) CredentialByRequest(ownerType, ownerID, requestID string) (model.MCPCredential, error) {
	var key model.MCPCredential
	err := mcpOwnerQuery(r.DB, ownerType, ownerID).Where("request_id = ?", requestID).First(&key).Error
	return key, err
}

func (r MCPRepo) SearchCredentials(ownerType, ownerID string, page, pageSize int, excludeRevoked bool) ([]model.MCPCredential, int64, error) {
	q := mcpOwnerQuery(r.DB.Model(&model.MCPCredential{}), ownerType, ownerID)
	if excludeRevoked {
		q = q.Where("status <> ?", "Revoked")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	keys := []model.MCPCredential{}
	err := q.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&keys).Error
	return keys, total, err
}

func (r MCPRepo) CountActiveName(ownerType, ownerID, name, excludeID string) (int64, error) {
	q := mcpOwnerQuery(r.DB.Model(&model.MCPCredential{}), ownerType, ownerID)
	if excludeID == "" {
		q = q.Where("active_name = ?", name)
	} else {
		q = q.Where("active_name = ? AND id <> ?", name, excludeID)
	}
	var count int64
	err := q.Count(&count).Error
	return count, err
}

func (r MCPRepo) CreateCredential(key *model.MCPCredential) error {
	return r.DB.Create(key).Error
}

func (r MCPRepo) UpdateCredential(key *model.MCPCredential, revision uint64) (int64, error) {
	result := r.DB.Model(&model.MCPCredential{}).Where("id = ? AND revision = ?", key.ID, revision).
		Select("name", "active_name", "description", "node_ids", "status", "expires_at", "revision", "revoked_at", "secret_ciphertext", "secret_version", "token_hash", "key_hint").Updates(key)
	return result.RowsAffected, result.Error
}

func (r MCPRepo) DeleteCredential(ownerType, ownerID, id string, revision uint64) (int64, error) {
	result := mcpOwnerQuery(r.DB, ownerType, ownerID).Where("id = ? AND revision = ?", id, revision).Delete(&model.MCPCredential{})
	return result.RowsAffected, result.Error
}

func (r MCPRepo) AdmissionCredential(id string, revision uint64) (model.MCPCredential, error) {
	var key model.MCPCredential
	err := r.DB.Where("id = ? AND revision = ? AND status = ? AND revoked_at IS NULL", id, revision, "Enable").Take(&key).Error
	return key, err
}

func (r MCPRepo) InvocationCredential(id string, revision uint64) (model.MCPCredential, error) {
	var key model.MCPCredential
	err := r.DB.Select("expires_at").
		Where("id = ? AND revision = ? AND status = ? AND revoked_at IS NULL", id, revision, "Enable").
		Where("EXISTS (SELECT 1 FROM mcp_settings WHERE id = ? AND enabled = ?)", 1, true).Take(&key).Error
	return key, err
}

func (r MCPRepo) Token(hash string) (model.MCPCredential, error) {
	var key model.MCPCredential
	err := r.DB.Where("token_hash = ?", hash).First(&key).Error
	return key, err
}
func (r MCPRepo) Grants(id string) ([]model.MCPGrant, error) {
	grants := []model.MCPGrant{}
	err := r.DB.Where("credential_id = ?", id).Order("tool_id").Find(&grants).Error
	return grants, err
}

func (r MCPRepo) DeleteGrants(id string) error {
	return r.DB.Where("credential_id = ?", id).Delete(&model.MCPGrant{}).Error
}

func (r MCPRepo) ReplaceTools(id string, grants []model.MCPGrant) error {
	if err := r.DeleteGrants(id); err != nil {
		return err
	}
	for _, grant := range grants {
		grant.CredentialID = id
		if err := r.DB.Create(&grant).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r MCPRepo) Setting() (model.MCPSetting, error) {
	var setting model.MCPSetting
	err := r.DB.First(&setting, 1).Error
	return setting, err
}

func (r MCPRepo) Enabled() (bool, error) {
	var setting struct{ Enabled bool }
	err := r.DB.Table("mcp_settings").Where("id = ?", 1).Take(&setting).Error
	return setting.Enabled, err
}

func (r MCPRepo) AdmissionEnabled() (bool, error) {
	var setting model.MCPSetting
	err := r.DB.Select("enabled").Take(&setting, 1).Error
	return setting.Enabled, err
}

func (r MCPRepo) PanelSetting(key string) (model.Setting, error) {
	var setting model.Setting
	err := r.DB.Where("key = ?", key).First(&setting).Error
	return setting, err
}

func (r MCPRepo) CreateSettingIfAbsent(setting *model.MCPSetting) error {
	return r.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(setting).Error
}

func (r MCPRepo) UpdateRuntimeSetting(setting model.MCPSetting, revision uint64) (int64, error) {
	result := r.DB.Model(&model.MCPSetting{}).Where("id = ? AND revision = ?", 1, revision).Updates(map[string]interface{}{
		"runtime_type": setting.RuntimeType, "docker_image": setting.DockerImage,
		"enabled": setting.Enabled, "listen_port": setting.ListenPort, "external_url": setting.ExternalURL, "revision": setting.Revision,
		"access_policy": setting.AccessPolicy, "allowed_ips": setting.AllowedIPs,
	})
	return result.RowsAffected, result.Error
}

func (r MCPRepo) UpdateWebsiteSetting(setting model.MCPSetting, revision uint64) (int64, error) {
	result := r.DB.Model(&model.MCPSetting{}).Where("id = 1 AND revision = ?", revision).Updates(map[string]any{
		"website_id": setting.WebsiteID, "website_binding_id": setting.WebsiteBindingID, "website_operation": setting.WebsiteOperation, "revision": setting.Revision,
	})
	return result.RowsAffected, result.Error
}

func (r MCPRepo) CreateEnrollmentAudit(record *model.OperationLog) error {
	return r.DB.Create(record).Error
}

func (r MCPRepo) CreateInvocation(record *model.MCPInvocation) error {
	return r.DB.Create(record).Error
}

func (r MCPRepo) SearchInvocations(ownerType, ownerID string, page, pageSize int, credentialID, toolID, status string) ([]model.MCPInvocation, int64, error) {
	q := mcpOwnerQuery(r.DB.Model(&model.MCPInvocation{}), ownerType, ownerID)
	if credentialID != "" {
		q = q.Where("credential_id = ?", credentialID)
	}
	if toolID != "" {
		q = q.Where("tool_id = ?", toolID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := []model.MCPInvocation{}
	if err := q.Order(mcpInvocationSecond + " DESC, " + mcpInvocationNanosecond + " DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r MCPRepo) UpdateRunningInvocation(record model.MCPInvocation) error {
	return r.DB.Model(&model.MCPInvocation{}).Where("id = ? AND status = ?", record.ID, "running").
		Updates(map[string]any{"status": record.Status, "code": record.Code, "finished_at": record.FinishedAt, "duration_ms": record.DurationMs}).Error
}

func (r MCPRepo) UpdateInterruptedInvocations(processStartedAt time.Time, status, code string, finishedAt time.Time) error {
	return r.DB.Model(&model.MCPInvocation{}).
		Where("status = ? AND finished_at IS NULL", "running").
		Where("("+mcpInvocationSecond+" < ? OR ("+mcpInvocationSecond+" = ? AND "+mcpInvocationNanosecond+" < ?))", processStartedAt.Unix(), processStartedAt.Unix(), processStartedAt.Nanosecond()).
		Updates(map[string]any{"status": status, "code": code, "finished_at": finishedAt}).Error
}
