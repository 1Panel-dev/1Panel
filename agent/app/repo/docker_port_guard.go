package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IDockerPortGuardRepo interface {
	List(context.Context) ([]model.DockerPortGuardPolicy, error)
	DeleteBatch(context.Context, []string) error
	UpsertBatch(context.Context, []model.DockerPortGuardPolicy) error
}

type DockerPortGuardRepo struct{}

func NewIDockerPortGuardRepo() IDockerPortGuardRepo { return &DockerPortGuardRepo{} }

func (r *DockerPortGuardRepo) List(ctx context.Context) ([]model.DockerPortGuardPolicy, error) {
	var policies []model.DockerPortGuardPolicy
	err := global.DB.WithContext(ctx).Order("family, host_ip, host_port, protocol").Find(&policies).Error
	return policies, err
}

func (r *DockerPortGuardRepo) DeleteBatch(ctx context.Context, uuids []string) error {
	return global.DB.WithContext(ctx).Where("uuid IN ?", uuids).Delete(&model.DockerPortGuardPolicy{}).Error
}

func (r *DockerPortGuardRepo) UpsertBatch(ctx context.Context, policies []model.DockerPortGuardPolicy) error {
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range policies {
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "family"}, {Name: "host_ip"}, {Name: "host_port"}, {Name: "protocol"}},
				DoUpdates: clause.AssignmentColumns([]string{"mode", "sources", "description", "updated_at"}),
			}).Create(&policies[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
