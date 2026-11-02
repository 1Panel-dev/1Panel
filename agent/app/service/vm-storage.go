package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/types"
	"gorm.io/gorm"
)

func (v *VMService) ListStorages() ([]dto.VMStorageItem, error) {
	records, err := vmRepo.ListStorages()
	if err != nil {
		return nil, err
	}
	items := v.mergeStorageItems(records)
	return items, nil
}

func (v *VMService) CreateStorage(req dto.VMStorageCreate) error {
	ctx, cancel := vmResourceOperationContext(false)
	defer cancel()
	unlock, err := vmResourceLocks.lock(ctx, "storagePool", req.Name)
	if err != nil {
		return err
	}
	defer unlock()
	req.Path = strings.TrimSpace(req.Path)
	unlockPath, err := vmResourceLocks.lock(ctx, "storagePath", normalizeStoragePath(req.Path))
	if err != nil {
		return err
	}
	defer unlockPath()
	itemStorage, err := vmRepo.GetStorage(repo.WithByName(req.Name))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if itemStorage.ID != 0 {
		return buserr.New("ErrRecordExist")
	}
	records, err := vmRepo.ListStorages()
	if err != nil {
		return err
	}
	for _, record := range records {
		if sameStoragePath(record.Path, req.Path) {
			return buserr.WithName("ErrVMStoragePathInUse", record.Name)
		}
	}
	if !v.helperExists() {
		return buserr.New("ErrVMHelperNotFound")
	}
	liveItems, err := v.provider.ListResources(ctx, "storagePool")
	if err != nil {
		return err
	}
	for _, live := range liveItems {
		if live.Name == req.Name {
			return buserr.New("ErrRecordExist")
		}
		if sameStoragePath(live.Path, req.Path) {
			return buserr.WithName("ErrVMStoragePathInUse", live.Name)
		}
	}
	record := model.VMStorage{
		Name: req.Name, Type: req.Type, Path: req.Path,
		AutoStart: req.AutoStart, Status: vmResourceCreating, Description: req.Description,
	}
	if err := vmRepo.CreateStorage(&record); err != nil {
		return err
	}
	if err := v.provider.CreateResource(ctx, types.VMResourceCreate{
		Type: "storagePool", Name: req.Name, SubType: req.Type, Path: req.Path, AutoStart: req.AutoStart,
	}); err != nil {
		if updateErr := vmRepo.UpdateStorage(record.ID, map[string]interface{}{
			"status": constant.StatusFailed, "error": err.Error(),
		}); updateErr != nil {
			return errors.Join(err, fmt.Errorf("save storage pool creation failure: %w", updateErr))
		}
		return err
	}
	return vmRepo.UpdateStorage(record.ID, map[string]interface{}{
		"status": constant.StatusRunning, "error": "",
	})
}

func sameStoragePath(left, right string) bool {
	left = normalizeStoragePath(left)
	right = normalizeStoragePath(right)
	return left != "" && left == right
}

func normalizeStoragePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.Clean(path)
}

func (v *VMService) SyncStorages() error {
	revision, idle := vmResourceLocks.snapshot("storagePool")
	if !idle {
		return nil
	}
	if !v.helperExists() {
		return buserr.New("ErrVMHelperNotFound")
	}
	liveItems, err := v.provider.ListResources(context.Background(), "storagePool")
	if err != nil {
		return err
	}
	return vmResourceLocks.applySnapshot("storagePool", revision, func() error {
		records, err := vmRepo.ListStorages()
		if err != nil {
			return err
		}
		liveByName := map[string]dto.VMResourceInfo{}
		for _, live := range liveItems {
			if live.Name == "" {
				continue
			}
			liveByName[live.Name] = live
		}
		recordsByName := map[string]model.VMStorage{}
		for _, record := range records {
			if vmResourceStatusEquals(record.Status, vmResourceCreating) {
				record.Status = constant.StatusFailed
				record.Error = "Storage pool creation was interrupted before its result was saved"
				if err := vmRepo.UpdateStorage(record.ID, map[string]interface{}{
					"status": record.Status, "error": record.Error,
				}); err != nil {
					return err
				}
			}
			recordsByName[record.Name] = record
		}
		for name, live := range liveByName {
			vars := map[string]interface{}{
				"type":       live.SubType,
				"path":       live.Path,
				"auto_start": live.AutoStart,
				"status":     normalizeVMResourceStatus(live.Status),
			}
			if record, ok := recordsByName[name]; ok {
				if vmResourceStatusEquals(record.Status, constant.StatusFailed) {
					continue
				}
				if err := vmRepo.UpdateStorage(record.ID, vars); err != nil {
					return err
				}
				continue
			}
			if err := vmRepo.CreateStorage(&model.VMStorage{
				Name: live.Name, Type: live.SubType, Path: live.Path,
				AutoStart: live.AutoStart, Status: normalizeVMResourceStatus(live.Status),
			}); err != nil {
				return err
			}
		}
		for _, record := range recordsByName {
			if vmResourceStatusEquals(record.Status, constant.StatusFailed) {
				continue
			}
			if _, ok := liveByName[record.Name]; !ok {
				if err := vmRepo.UpdateStorage(record.ID, map[string]interface{}{
					"status": constant.StatusDeleted,
				}); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func (v *VMService) UpdateStorage(req dto.VMStorageUpdate) error {
	ctx, cancel := vmResourceOperationContext(false)
	defer cancel()
	itemStorage, err := vmRepo.GetStorage(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	unlock, err := vmResourceLocks.lock(ctx, "storagePool", itemStorage.Name)
	if err != nil {
		return err
	}
	defer unlock()
	itemStorage, err = vmRepo.GetStorage(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if itemStorage.ID == 0 {
		return buserr.New("ErrRecordNotFound")
	}
	vars := map[string]interface{}{}
	if req.Name != "" && req.Name != itemStorage.Name {
		return buserr.New("ErrVMStorageNameCannotChange")
	}
	if req.Path != "" && req.Path != itemStorage.Path {
		return buserr.New("ErrVMStoragePathCannotChange")
	}
	if req.Description != nil {
		vars["description"] = *req.Description
	}
	if req.AutoStart != nil || req.Status != nil {
		if vmResourceCreationIncomplete(itemStorage.Status) {
			return buserr.WithDetail("ErrInvalidParams", "storage pool creation has not succeeded", nil)
		}
		if !v.helperExists() {
			return buserr.New("ErrVMHelperNotFound")
		}
		live, ok, err := v.findVMResourceWithContext(ctx, "storagePool", itemStorage.Name)
		if err != nil {
			return err
		}
		if !ok {
			_ = vmRepo.UpdateStorage(req.ID, map[string]interface{}{"status": constant.StatusDeleted})
			return buserr.WithDetail("ErrResourceNotExists", "Storage pool does not exist", nil)
		}
		if live.Status != "" {
			vars["status"] = normalizeVMResourceStatus(live.Status)
		}
	}
	if req.AutoStart != nil {
		if err := v.provider.UpdateResource(ctx, types.VMResourceUpdate{
			Type: "storagePool", Name: itemStorage.Name, AutoStart: req.AutoStart,
		}); err != nil {
			return err
		}
		vars["auto_start"] = *req.AutoStart
	}
	if req.Status != nil {
		status := normalizeVMResourceStatus(*req.Status)
		if !validVMResourceOperateStatus(status) {
			return buserr.WithDetail("ErrInvalidParams", status, nil)
		}
		if err := v.provider.UpdateResource(ctx, types.VMResourceUpdate{
			Type: "storagePool", Name: itemStorage.Name, Status: &status,
		}); err != nil {
			return err
		}
		vars["status"] = status
	}
	if len(vars) == 0 {
		return nil
	}
	return vmRepo.UpdateStorage(req.ID, vars)
}

func (v *VMService) DeleteStorage(req dto.VMResourceDelete) error {
	ctx, cancel := vmResourceOperationContext(req.Force)
	defer cancel()
	itemStorage, err := vmRepo.GetStorage(repo.WithByID(req.ID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return buserr.New("ErrRecordNotFound")
		}
		return err
	}
	unlock, cleanup, err := vmResourceLocks.lockForDelete(ctx, "storagePool", itemStorage.Name, req.Force)
	if err != nil {
		return err
	}
	defer unlock()
	itemStorage, err = vmRepo.GetStorage(repo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if len(loadVMResourceRefs("storagePool")[itemStorage.Name]) > 0 {
		return buserr.New("ErrVMStorageInUse")
	}
	if cleanup {
		if err := v.deleteVMResource(ctx, "storagePool", itemStorage.Name, req.Force); err != nil {
			return err
		}
	}
	return vmRepo.DeleteStorage(repo.WithByID(req.ID))
}

func (v *VMService) mergeStorageItems(records []model.VMStorage) []dto.VMStorageItem {
	itemsByValue := map[string]dto.VMStorageItem{}
	for _, record := range records {
		itemsByValue[record.Name] = dto.VMStorageItem{
			ID: record.ID, Name: record.Name, Type: record.Type,
			Path: record.Path, AutoStart: record.AutoStart, Description: record.Description,
			CreatedAt: record.CreatedAt.Format("2006-01-02 15:04:05"),
			Status:    normalizeVMResourceStatus(record.Status),
			Error:     record.Error,
		}
	}
	if v.helperExists() {
		if liveItems, err := v.provider.ListResources(context.Background(), "storagePool"); err == nil {
			for _, live := range liveItems {
				item, ok := itemsByValue[live.Name]
				if !ok || vmResourceCreationIncomplete(item.Status) {
					continue
				}
				item.Type = live.SubType
				item.Status = normalizeVMResourceStatus(live.Status)
				item.AutoStart = live.AutoStart
				item.Capacity = live.Capacity
				item.Available = live.Available
				itemsByValue[live.Name] = item
			}
		}
	}
	vmRefs := loadVMResourceRefs("storagePool")
	items := make([]dto.VMStorageItem, 0, len(records))
	for _, record := range records {
		item := itemsByValue[record.Name]
		if item.Status == "" {
			item.Status = constant.StatusDeleted
		}
		item.VMs = vmRefs[record.Name]
		item.VMCount = len(item.VMs)
		items = append(items, item)
	}
	return items
}

func loadVMResourceRefs(resourceType string) map[string][]dto.VMResourceRef {
	records, err := vmRepo.List()
	if err != nil {
		return map[string][]dto.VMResourceRef{}
	}
	refs := map[string][]dto.VMResourceRef{}
	for _, record := range records {
		ref := dto.VMResourceRef{ID: record.ID, Name: record.Name}
		switch resourceType {
		case "network":
			if record.Network != "" {
				refs[record.Network] = append(refs[record.Network], ref)
			}
		case "storagePool":
			if record.StoragePool != "" {
				refs[record.StoragePool] = append(refs[record.StoragePool], ref)
			}
		}
	}
	for name := range refs {
		sort.Slice(refs[name], func(i, j int) bool {
			return strings.ToLower(refs[name][i].Name) < strings.ToLower(refs[name][j].Name)
		})
	}
	return refs
}

func (v *VMService) fillVMStoragePools(items []dto.VMInfo) {
	needsPools := false
	for _, item := range items {
		if item.StoragePool == "" && item.DiskPath != "" {
			needsPools = true
			break
		}
	}
	if !needsPools {
		return
	}
	pools, err := v.provider.ListResources(context.Background(), "storagePool")
	if err != nil {
		global.LOG.Warnf("[xpack] resolve VM storage pools failed, retaining metadata: %v", err)
		return
	}
	for i := range items {
		if items[i].StoragePool == "" {
			items[i].StoragePool = storagePoolForDisk(items[i].DiskPath, pools)
		}
	}
}

func storagePoolForDisk(diskPath string, pools []dto.VMResourceInfo) string {
	if !filepath.IsAbs(diskPath) {
		return ""
	}
	poolName, bestLength := "", 0
	for _, pool := range pools {
		poolPath := normalizeStoragePath(pool.Path)
		if pool.Name == "" || !filepath.IsAbs(poolPath) {
			continue
		}
		rel, err := filepath.Rel(poolPath, diskPath)
		if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		if len(poolPath) > bestLength {
			poolName, bestLength = pool.Name, len(poolPath)
		} else if len(poolPath) == bestLength && poolName != pool.Name {
			poolName = ""
		}
	}
	return poolName
}
