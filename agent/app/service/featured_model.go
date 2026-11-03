package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

var featuredModelsSyncMu sync.Mutex

func (m *ModelDownloaderService) PageFeaturedModels(req dto.FeaturedModelSearch) (*dto.FeaturedModelPage, error) {
	total, models, err := featuredModelRepo.Page(req.Page, req.PageSize, strings.TrimSpace(req.Family), strings.TrimSpace(req.WeightProvider), strings.TrimSpace(req.Info))
	if err != nil {
		return nil, err
	}
	items := make([]dto.FeaturedModel, 0, len(models))
	for _, item := range models {
		var description map[string]string
		if err := json.Unmarshal([]byte(item.Description), &description); err != nil {
			return nil, fmt.Errorf("parse featured model %s description failed: %w", item.Key, err)
		}
		items = append(items, dto.FeaturedModel{
			Key: item.Key, Sort: item.Sort, Family: item.Family, RepoID: item.RepoID, Sources: item.Sources,
			SourceModelIDs: item.SourceModelIDs, WeightProvider: item.WeightProvider,
			Architecture: item.Architecture, BackboneParams: item.BackboneParams,
			ActiveParams: item.ActiveParams, Engram: item.Engram, WeightGB: item.WeightGB,
			Description: description,
		})
	}
	families, err := featuredModelRepo.Families()
	if err != nil {
		return nil, err
	}
	weightProviders, err := featuredModelRepo.WeightProviders()
	if err != nil {
		return nil, err
	}
	return &dto.FeaturedModelPage{PageResult: dto.PageResult{Total: total, Items: items}, Families: families, WeightProviders: weightProviders}, nil
}

func (m *ModelDownloaderService) SyncFeaturedModels() error {
	if global.CONF.Base.IsOffline {
		return nil
	}
	featuredModelsSyncMu.Lock()
	defer featuredModelsSyncMu.Unlock()
	localVersion, err := settingRepo.GetValueByKey("FeaturedModelsVersion")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	data, version, err := req_helper.LoadAIResource("featured-models", localVersion)
	if err != nil || data == nil {
		return err
	}
	var remote struct {
		Models []dto.FeaturedModel `yaml:"models"`
	}
	if err := yaml.Unmarshal(data, &remote); err != nil {
		return fmt.Errorf("parse featured models failed: %w", err)
	}
	if len(remote.Models) == 0 {
		return errors.New("featured models data is empty")
	}
	keys := make(map[string]bool, len(remote.Models))
	models := make([]model.FeaturedModel, 0, len(remote.Models))
	for i := range remote.Models {
		item := &remote.Models[i]
		item.Key = strings.TrimSpace(item.Key)
		item.Family = strings.TrimSpace(item.Family)
		item.RepoID = strings.TrimSpace(item.RepoID)
		item.WeightProvider = strings.TrimSpace(item.WeightProvider)
		if item.WeightProvider == "" {
			if provider, _, ok := strings.Cut(item.RepoID, "/"); ok {
				item.WeightProvider = provider
			}
		}
		if item.Key == "" || item.Family == "" || item.RepoID == "" || (item.Architecture != "MoE" && item.Architecture != "Dense") || item.WeightGB < 0 {
			return fmt.Errorf("invalid featured model: %s", item.Key)
		}
		if keys[item.Key] {
			return fmt.Errorf("duplicate featured model key: %s", item.Key)
		}
		keys[item.Key] = true
		if len(item.Sources) == 0 {
			item.Sources = []string{"modelscope"}
		}
		for _, source := range item.Sources {
			if source != "modelscope" && source != "huggingface" {
				return fmt.Errorf("invalid featured model %s source: %s", item.Key, source)
			}
		}
		description, err := json.Marshal(item.Description)
		if err != nil {
			return err
		}
		models = append(models, model.FeaturedModel{
			Key: item.Key, Sort: item.Sort, Family: item.Family, RepoID: item.RepoID, Sources: item.Sources,
			SourceModelIDs: item.SourceModelIDs, WeightProvider: item.WeightProvider,
			Architecture: item.Architecture, BackboneParams: item.BackboneParams,
			ActiveParams: item.ActiveParams, Engram: item.Engram, WeightGB: item.WeightGB,
			Description: string(description),
		})
	}
	return featuredModelRepo.SyncAll(models, version)
}

func SyncFeaturedModels() {
	if global.CONF.Base.IsOffline {
		return
	}
	status, err := settingRepo.GetValueByKey("FeaturedModelsSync")
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		global.LOG.Errorf("load featured model sync setting failed: %v", err)
		return
	}
	if err == nil && status != constant.StatusEnable {
		return
	}
	if err := NewIModelDownloaderService().SyncFeaturedModels(); err != nil {
		global.LOG.Errorf("sync featured models failed: %v", err)
	}
}
