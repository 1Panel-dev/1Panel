package repo

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IDockerPortGuardRepo interface {
	ListManaged(context.Context) ([]model.DockerPortGuardPolicy, error)
	ListRuntimeReadOnly(context.Context) ([]model.DockerPortGuardPolicy, error)
	DeleteBatch(context.Context, []string) error
	UpsertBatch(context.Context, []model.DockerPortGuardPolicy) error
	ReplaceRuntimeReadOnly(context.Context, []model.DockerPortGuardPolicy) error
}

type DockerPortGuardRepo struct{}

func NewIDockerPortGuardRepo() IDockerPortGuardRepo { return &DockerPortGuardRepo{} }

func (r *DockerPortGuardRepo) ListManaged(ctx context.Context) ([]model.DockerPortGuardPolicy, error) {
	var policies []model.DockerPortGuardPolicy
	err := global.DB.WithContext(ctx).
		Where("read_only = ?", false).
		Order("family, host_ip, host_port, protocol").
		Find(&policies).Error
	return policies, err
}

func (r *DockerPortGuardRepo) ListRuntimeReadOnly(ctx context.Context) ([]model.DockerPortGuardPolicy, error) {
	var policies []model.DockerPortGuardPolicy
	err := global.DB.WithContext(ctx).
		Where("read_only = ?", true).
		Order("family, sequence, host_ip, host_port, protocol").
		Find(&policies).Error
	return policies, err
}

func (r *DockerPortGuardRepo) DeleteBatch(ctx context.Context, uuids []string) error {
	return global.DB.WithContext(ctx).
		Where("read_only = ? AND uuid IN ?", false, uuids).
		Delete(&model.DockerPortGuardPolicy{}).Error
}

func (r *DockerPortGuardRepo) UpsertBatch(ctx context.Context, policies []model.DockerPortGuardPolicy) error {
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range policies {
			policies[i].ReadOnly = false
			if err := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "read_only"}, {Name: "family"}, {Name: "host_ip"}, {Name: "host_port"}, {Name: "protocol"}},
				DoUpdates: clause.AssignmentColumns([]string{"mode", "sources", "description", "updated_at"}),
			}).Create(&policies[i]).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *DockerPortGuardRepo) ReplaceRuntimeReadOnly(ctx context.Context, policies []model.DockerPortGuardPolicy) error {
	return global.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("read_only = ?", true).
			Delete(&model.DockerPortGuardPolicy{}).Error; err != nil {
			return err
		}
		if len(policies) == 0 {
			return nil
		}
		for i := range policies {
			policies[i].ReadOnly = true
		}
		return tx.Create(&policies).Error
	})
}
