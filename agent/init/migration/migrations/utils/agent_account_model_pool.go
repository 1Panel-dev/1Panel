package utils

import (
	"encoding/json"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	providercatalog "github.com/1Panel-dev/1Panel/agent/app/provider"
	"github.com/1Panel-dev/1Panel/agent/app/service"

	"gorm.io/gorm"
)

type legacyAgentAccountModelPoolSource struct {
	model.AgentAccount
	LegacyModel  string `gorm:"column:model"`
	LegacyModels string `gorm:"column:models"`
}

func MigrateAgentAccountModelPool(tx *gorm.DB) error {
	var accounts []legacyAgentAccountModelPoolSource
	if err := tx.Table(model.AgentAccount{}.TableName()).Find(&accounts).Error; err != nil {
		return err
	}
	for _, account := range accounts {
		count := int64(0)
		if err := tx.Model(&model.AgentAccountModel{}).Where("account_id = ?", account.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		models, err := buildMigratedAgentAccountModels(tx, &account)
		if err != nil {
			return err
		}
		if len(models) == 0 {
			continue
		}
		for index, item := range models {
			record := &model.AgentAccountModel{
				AccountID: account.ID,
				Model:     item.ID,
				Name:      item.Name,
				SortOrder: index + 1,
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func buildMigratedAgentAccountModels(tx *gorm.DB, account *legacyAgentAccountModelPoolSource) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, nil
	}
	baseAccount := account.AgentAccount
	requested := make([]dto.AgentAccountModel, 0)
	if strings.TrimSpace(account.LegacyModels) != "" {
		if err := json.Unmarshal([]byte(account.LegacyModels), &requested); err != nil {
			return nil, err
		}
	}
	seen := make(map[string]struct{}, len(requested))
	for _, item := range requested {
		target := strings.TrimSpace(item.ID)
		if target == "" {
			continue
		}
		seen[target] = struct{}{}
	}
	appendModel := func(modelID string) {
		target := strings.TrimSpace(modelID)
		if target == "" {
			return
		}
		if _, ok := seen[target]; ok {
			return
		}
		seen[target] = struct{}{}
		requested = append(requested, dto.AgentAccountModel{ID: target})
	}
	appendModel(account.LegacyModel)
	if account.ID > 0 {
		var agents []model.Agent
		if err := tx.Where("account_id = ?", account.ID).Find(&agents).Error; err != nil {
			return nil, err
		}
		for _, agent := range agents {
			appendModel(agent.Model)
		}
	}
	models, err := service.MergeCatalogAgentAccountModelsForMigration(&baseAccount, requested)
	if err != nil {
		if strings.TrimSpace(err.Error()) == "model is required" {
			return nil, nil
		}
		return nil, err
	}
	return models, nil
}

func NormalizeAgentAccountModelIDs(tx *gorm.DB) error {
	var accounts []model.AgentAccount
	if err := tx.Find(&accounts).Error; err != nil {
		return err
	}
	for _, account := range accounts {
		var accountModels []model.AgentAccountModel
		if err := tx.Where("account_id = ?", account.ID).Find(&accountModels).Error; err != nil {
			return err
		}
		for _, item := range accountModels {
			normalized := providercatalog.NormalizeModelID(account.Provider, item.Model)
			if normalized != item.Model {
				if err := tx.Model(&model.AgentAccountModel{}).Where("id = ?", item.ID).Update("model", normalized).Error; err != nil {
					return err
				}
			}
		}

		var agents []model.Agent
		if err := tx.Where("account_id = ?", account.ID).Find(&agents).Error; err != nil {
			return err
		}
		for _, agent := range agents {
			normalized := providercatalog.NormalizeModelID(account.Provider, agent.Model)
			if normalized != agent.Model {
				if err := tx.Model(&model.Agent{}).Where("id = ?", agent.ID).Update("model", normalized).Error; err != nil {
					return err
				}
			}
		}
	}
	if err := tx.Model(&model.AgentAccount{}).
		Where("provider = ? AND api_type = ? AND (base_url = '' OR base_url = ? OR base_url = ?)", "deepseek", "anthropic-messages", "https://api.deepseek.com", "https://api.deepseek.com/v1").
		Update("base_url", "https://api.deepseek.com/anthropic").Error; err != nil {
		return err
	}
	return tx.Model(&model.AgentAccount{}).
		Where("provider = ?", "gemini").
		Update("api_type", "gemini-generate-content").Error
}
