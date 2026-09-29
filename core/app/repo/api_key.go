package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"github.com/1Panel-dev/1Panel/core/global"
	"gorm.io/gorm"
)

type APIKeyRepo struct{}

func NewAPIKeyRepo() *APIKeyRepo { return &APIKeyRepo{} }
func APIKeyOwnerQuery(db *gorm.DB, ownerType, ownerID string) *gorm.DB {
	return db.Where("owner_type = ? AND owner_id = ?", ownerType, ownerID)
}
func (r *APIKeyRepo) Get(id string) (model.APIKey, error) {
	var key model.APIKey
	err := global.DB.Where("id = ?", id).First(&key).Error
	return key, err
}
func (r *APIKeyRepo) List(ownerType, ownerID string, excludeRevoked bool) ([]model.APIKey, error) {
	items := make([]model.APIKey, 0)
	q := APIKeyOwnerQuery(global.DB, ownerType, ownerID)
	if excludeRevoked {
		q = q.Where("status <> ?", "Revoked")
	}
	err := q.Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}
