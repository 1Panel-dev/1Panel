package migrations

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var MigrateModelDownloader = &gormigrate.Migration{
	ID: "20261102-migrate-model-downloader",
	Migrate: func(db *gorm.DB) error {
		return migrateModelDownloader(db, filepath.Join(global.Dir.DbDir, "xpack.db"))
	},
}

var AddFeaturedModelWeightProvider = &gormigrate.Migration{
	ID: "20261103-add-featured-model-weight-provider",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.FeaturedModel{}); err != nil {
			return err
		}
		if err := tx.Model(&model.FeaturedModel{}).
			Where("(weight_provider IS NULL OR weight_provider = '') AND instr(repo_id, '/') > 0").
			Update("weight_provider", gorm.Expr("substr(repo_id, 1, instr(repo_id, '/') - 1)")).Error; err != nil {
			return err
		}
		return tx.Model(&model.Setting{}).Where("key = ?", "FeaturedModelsVersion").Update("value", "").Error
	},
}

func migrateModelDownloader(db *gorm.DB, legacyPath string) error {
	var legacy *gorm.DB
	if _, err := os.Stat(legacyPath); err == nil {
		source := url.URL{Scheme: "file", Path: legacyPath, RawQuery: "mode=ro"}
		legacy, err = gorm.Open(sqlite.Open(source.String()), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
		if err != nil {
			return err
		}
		sqlDB, err := legacy.DB()
		if err != nil {
			return err
		}
		defer sqlDB.Close()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&model.ModelDownloadTask{}, &model.FeaturedModel{}); err != nil {
			return err
		}
		resetVersion := false
		if legacy != nil {
			if legacy.Migrator().HasTable(&model.ModelDownloadTask{}) {
				var tasks []model.ModelDownloadTask
				if err := legacy.FindInBatches(&tasks, 100, func(_ *gorm.DB, _ int) error {
					for i := range tasks {
						if tasks[i].DownloadType == "" {
							tasks[i].DownloadType = "all"
						}
					}
					return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&tasks).Error
				}).Error; err != nil {
					return err
				}
			}
			if legacy.Migrator().HasTable(&model.FeaturedModel{}) {
				resetVersion = !legacy.Migrator().HasColumn(&model.FeaturedModel{}, "Sources") || !legacy.Migrator().HasColumn(&model.FeaturedModel{}, "SourceModelIDs")
				var models []model.FeaturedModel
				if err := legacy.FindInBatches(&models, 100, func(_ *gorm.DB, _ int) error {
					for i := range models {
						if len(models[i].Sources) == 0 {
							models[i].Sources = []string{"modelscope"}
						}
					}
					return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&models).Error
				}).Error; err != nil {
					return err
				}
			} else {
				resetVersion = true
			}
			if legacy.Migrator().HasTable("agent_settings") {
				var settings []model.Setting
				keys := []string{"ModelDownloaderDir", "ModelDownloaderHFEndpoint", "ModelDownloaderHFToken", "ModelDownloaderModelScopeEndpoint", "ModelDownloaderModelScopeToken", "FeaturedModelsSync", "FeaturedModelsVersion"}
				if err := legacy.Table("agent_settings").Where("key IN ?", keys).Find(&settings).Error; err != nil {
					return err
				}
				for _, setting := range settings {
					if setting.Key == "FeaturedModelsVersion" && resetVersion {
						setting.Value = ""
					}
					if err := tx.Where("key = ?", setting.Key).FirstOrCreate(&model.Setting{Key: setting.Key, Value: setting.Value, About: setting.About}).Error; err != nil {
						return err
					}
				}
			}
		}
		var count int64
		if err := tx.Model(&model.FeaturedModel{}).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		if err := tx.Model(&model.Setting{}).Where("key = ?", "FeaturedModelsVersion").Update("value", "").Error; err != nil {
			return err
		}
		models := []model.FeaturedModel{
			{Family: "Qwen", RepoID: "Qwen/Qwen3.6-35B-A3B", Architecture: "MoE", BackboneParams: "35B", ActiveParams: "3B", Engram: "N/A", WeightGB: 67},
			{Family: "Qwen", RepoID: "Qwen/Qwen3.6-35B-A3B-FP8", Architecture: "MoE", BackboneParams: "35B", ActiveParams: "3B", Engram: "N/A", WeightGB: 35},
			{Family: "Qwen", RepoID: "nvidia/Qwen3.6-35B-A3B-NVFP4", Architecture: "MoE", BackboneParams: "35B", ActiveParams: "3B", Engram: "N/A", WeightGB: 22},
			{Family: "Qwen", RepoID: "Qwen/Qwen3.8-27B", Architecture: "Dense", BackboneParams: "N/A", ActiveParams: "N/A", Engram: "N/A", WeightGB: 55},
			{Family: "Qwen", RepoID: "Qwen/Qwen3.8-27B-FP8", Architecture: "Dense", BackboneParams: "N/A", ActiveParams: "N/A", Engram: "N/A", WeightGB: 28},
			{Family: "Qwen", RepoID: "nvidia/Qwen3.8-27B-NVFP4", Architecture: "Dense", BackboneParams: "N/A", ActiveParams: "N/A", Engram: "N/A", WeightGB: 22},
			{Family: "Qwen", RepoID: "Qwen/Qwen3.8-Flash-Next", Architecture: "MoE", BackboneParams: "125B", ActiveParams: "6B", Engram: "51B", WeightGB: 335},
			{Family: "Qwen", RepoID: "Qwen/Qwen3.8-Flash-Next-FP8", Architecture: "MoE", BackboneParams: "125B", ActiveParams: "6B", Engram: "51B", WeightGB: 172},
			{Family: "Qwen", RepoID: "nvidia/Qwen3.8-Flash-Next-FP4-Local", Architecture: "MoE", BackboneParams: "125B", ActiveParams: "6B", Engram: "51B", WeightGB: 130},
			{Family: "DeepSeek", RepoID: "deepseek-ai/DeepSeek-V4-Flash-0731", Architecture: "MoE", BackboneParams: "284B", ActiveParams: "13B", Engram: "N/A", WeightGB: 167},
			{Family: "DeepSeek", RepoID: "deepseek-ai/DeepSeek-V4-Flash-Vision-Exp", Architecture: "MoE", BackboneParams: "305B", ActiveParams: "undisclosed", Engram: "N/A", WeightGB: 167},
			{Family: "DeepSeek", RepoID: "deepseek-ai/DeepSeek-V4.1-Flash-1017", Architecture: "MoE", BackboneParams: "552B", ActiveParams: "8B/16B", Engram: "196B", WeightGB: 510},
			{Family: "GLM", RepoID: "zai-org/GLM-5.3-Flash", Architecture: "MoE", BackboneParams: "320B", ActiveParams: "18B", Engram: "N/A", WeightGB: 306},
			{Family: "GLM", RepoID: "nvidia/GLM-5.3-Flash-FP4-Local", Architecture: "MoE", BackboneParams: "320B", ActiveParams: "18B", Engram: "N/A", WeightGB: 185},
		}
		for i := range models {
			models[i].Key = models[i].RepoID
			models[i].Sort = i + 1
			models[i].Description = "{}"
			models[i].Sources = []string{"modelscope"}
		}
		return tx.Create(&models).Error
	})
}
