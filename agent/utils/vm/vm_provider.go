package vm

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/helper"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/types"
)

type VMProvider interface {
	CheckEnvironment(ctx context.Context) ([]dto.VMCheckItem, error)
	HealthCheck(ctx context.Context, name string) (dto.VMHealthCheckResult, error)
	HelperStatus() bool
	List(ctx context.Context) ([]dto.VMInfo, error)
	ListResources(ctx context.Context, resourceType string) ([]dto.VMResourceInfo, error)
	CreateResource(ctx context.Context, req types.VMResourceCreate) error
	UpdateResource(ctx context.Context, req types.VMResourceUpdate) error
	DeleteResource(ctx context.Context, resourceType string, name string) error
	ScanOrphans(ctx context.Context, protectedPaths []string) ([]dto.VMOrphanItem, error)
	CleanOrphans(ctx context.Context, items []dto.VMOrphanItem, protectedPaths []string) ([]dto.VMOrphanItem, error)
	Get(ctx context.Context, name string, checkDisks bool) (dto.VMInfo, error)
	Stats(ctx context.Context, name string) (dto.VMStats, error)
	Create(ctx context.Context, cfg types.VMConfig) (dto.VMInfo, error)
	CreateFromTemplate(ctx context.Context, cfg types.VMConfig) (dto.VMInfo, error)
	CreateTemplate(ctx context.Context, req types.VMTemplateCreate) (dto.VMInfo, error)
	Update(ctx context.Context, cfg types.VMConfig) (dto.VMInfo, error)
	ResizeDisk(ctx context.Context, req dto.VMDiskResize) (dto.VMInfo, error)
	Rename(ctx context.Context, name string, newName string) (dto.VMInfo, error)
	Delete(ctx context.Context, name string, force bool, deleteDisk bool, protectedPaths []string) error
	ListSnapshots(ctx context.Context, name string) ([]dto.VMSnapshotInfo, error)
	CreateSnapshot(ctx context.Context, name string, snapshotName string, description string) error
	RecoverSnapshot(ctx context.Context, name string, snapshotName string) error
	DeleteSnapshot(ctx context.Context, name string, snapshotName string) error
	Start(ctx context.Context, name string) error
	Shutdown(ctx context.Context, name string) error
	Reboot(ctx context.Context, name string) error
	Destroy(ctx context.Context, name string) error
	Suspend(ctx context.Context, name string) error
	Resume(ctx context.Context, name string) error
	Console(ctx context.Context, name string) (*dto.VMGraphics, error)
}

func NewVMProvider() (VMProvider, error) {
	return helper.NewProvider()
}
