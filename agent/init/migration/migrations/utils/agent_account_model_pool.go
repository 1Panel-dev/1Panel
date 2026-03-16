package utils

import (
	"encoding/json"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/service"

	"gorm.io/gorm"
)

func MigrateAgentAccountModelPool(tx *gorm.DB) error {
	var accounts []model.AgentAccount
	if err := tx.Find(&accounts).Error; err != nil {
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
			inputPayload, err := json.Marshal(item.Input)
			if err != nil {
				return err
			}
			record := &model.AgentAccountModel{
				AccountID:     account.ID,
				Model:         item.ID,
				Name:          item.Name,
				ContextWindow: item.ContextWindow,
				MaxTokens:     item.MaxTokens,
				Reasoning:     item.Reasoning,
				Input:         string(inputPayload),
				SortOrder:     index + 1,
			}
			if err := tx.Create(record).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func buildMigratedAgentAccountModels(tx *gorm.DB, account *model.AgentAccount) ([]dto.AgentAccountModel, error) {
	if account == nil {
		return nil, nil
	}
	requested := make([]dto.AgentAccountModel, 0)
	if strings.TrimSpace(account.Models) != "" {
		if err := json.Unmarshal([]byte(account.Models), &requested); err != nil {
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
	appendModel(account.Model)
	if account.ID > 0 {
		var agents []model.Agent
		if err := tx.Where("account_id = ?", account.ID).Find(&agents).Error; err != nil {
			return nil, err
		}
		for _, agent := range agents {
			appendModel(agent.Model)
		}
	}
	models, err := service.MergeCatalogAgentAccountModelsForMigration(account, requested)
	if err != nil {
		if strings.TrimSpace(err.Error()) == "model is required" {
			return nil, nil
		}
		return nil, err
	}
	return models, nil
}
