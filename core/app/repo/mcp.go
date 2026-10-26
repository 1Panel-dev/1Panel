package repo

import (
	"github.com/1Panel-dev/1Panel/core/app/model"
	"gorm.io/gorm"
)

type MCPRepo struct{ DB *gorm.DB }

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
func (r MCPRepo) ReplaceTools(id string, grants []model.MCPGrant) error {
	if err := r.DB.Where("credential_id = ?", id).Delete(&model.MCPGrant{}).Error; err != nil {
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
