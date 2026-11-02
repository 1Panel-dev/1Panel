package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/controller"
	"github.com/1Panel-dev/1Panel/agent/utils/vm"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/protocol"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/types"
	"github.com/jinzhu/copier"
	"github.com/shirou/gopsutil/v4/mem"
	"gorm.io/gorm"
)

type VMService struct {
	provider vm.VMProvider
	locksMu  sync.Mutex
	locks    map[string]*vmLockEntry
}

type vmLockEntry struct {
	mu   sync.Mutex
	refs int
}

// Share the lock across service instances. The metadata written before the
// asynchronous task starts reserves a slot, including while a VM is creating.
var vmCreateMu sync.Mutex

const (
	defaultVMResourceName = "default"
	defaultVMDiskDir      = "/var/lib/libvirt/images"
	diskBusSATA           = "sata"
	diskBusVirtIO         = "virtio"
	osTypeWindows         = "windows"
	vmMemoryMinMiB        = int64(128)
	vmDiskMinGiB          = int64(1)
	vmMemoryReserveRatio  = 0.1
	vmMemoryReserveMin    = uint64(1024 * 1024 * 1024)
	bytePerMiB            = uint64(1024 * 1024)
	bytePerGiB            = uint64(1024 * 1024 * 1024)
)

var libvirtServiceNames = []string{
	"libvirtd.service",
	"virtqemud.service",
	"virtnetworkd.service",
	"libvirt-bin.service",
	"libvirtd",
	"virtqemud",
	"virtnetworkd",
	"libvirt-bin",
	"libvirt",
}

type IVMService interface {
	CheckEnvironment() (dto.VMEnvironment, error)
	HealthCheck(ctx context.Context, name string) (dto.VMHealthCheckResult, error)
	EnableEnvironment() error
	SyncResourcesOnStartup() error
	SyncVMs() error
	Search(req dto.SearchWithPage) (int64, interface{}, error)
	Get(name string, checkDisks bool) (dto.VMInfo, error)
	GetStats(name string) (dto.VMStats, error)
	Create(req dto.VMCreate, limit int) (dto.VMTaskResp, error)
	ListTemplates() ([]dto.VMTemplateInfo, error)
	CreateTemplate(req dto.VMTemplateCreate) (dto.VMTaskResp, error)
	DeleteTemplate(req dto.VMTemplateDelete) (dto.VMTaskResp, error)
	Update(req dto.VMUpdate) (dto.VMTaskResp, error)
	ResizeDisk(req dto.VMDiskResize) (dto.VMTaskResp, error)
	Rename(req dto.VMRename) (dto.VMTaskResp, error)
	Delete(req dto.VMDelete) (dto.VMTaskResp, error)
	ListSnapshots(req dto.VMSnapshotSearch) ([]dto.VMSnapshotInfo, error)
	CreateSnapshot(req dto.VMSnapshotReq) (dto.VMTaskResp, error)
	RecoverSnapshot(req dto.VMSnapshotReq) (dto.VMTaskResp, error)
	DeleteSnapshot(req dto.VMSnapshotReq) (dto.VMTaskResp, error)
	ScanOrphans() ([]dto.VMOrphanItem, error)
	CleanOrphans(req dto.VMOrphanClean) ([]dto.VMOrphanItem, error)
	Operate(req dto.VMOperate) error
	Console(name string) (*dto.VMGraphics, error)

	SearchIso(req dto.SearchPageWithType) (int64, interface{}, error)
	ListIso() ([]dto.VMIsoInfo, error)
	CreateIso(req dto.VMIsoCreate) error
	UpdateIso(req dto.VMIsoUpdate) error
	DeleteIso(req dto.VMIsoDelete) error
	StartIsoImport(req dto.VMIsoImportStart) (uint, error)
	UpdateIsoImport(req dto.VMIsoImportUpdate) error

	ListNetworks() ([]dto.VMNetworkItem, error)
	ListNetworkBridges() ([]dto.VMNetworkBridgeItem, error)
	ListNetworkInterfaces() ([]dto.VMNetworkInterfaceItem, error)
	CreateNetwork(req dto.VMNetworkCreate) error
	SyncNetworks() error
	UpdateNetwork(req dto.VMNetworkUpdate) error
	DeleteNetwork(req dto.VMResourceDelete) error

	ListStorages() ([]dto.VMStorageItem, error)
	CreateStorage(req dto.VMStorageCreate) error
	SyncStorages() error
	UpdateStorage(req dto.VMStorageUpdate) error
	DeleteStorage(req dto.VMResourceDelete) error
}

func (v *VMService) ScanOrphans() ([]dto.VMOrphanItem, error) {
	paths, err := protectedVMPaths(0)
	if err != nil {
		return nil, err
	}
	return v.provider.ScanOrphans(context.Background(), paths)
}

func (v *VMService) CleanOrphans(req dto.VMOrphanClean) ([]dto.VMOrphanItem, error) {
	paths, err := protectedVMPaths(0)
	if err != nil {
		return nil, err
	}
	return v.provider.CleanOrphans(context.Background(), req.Items, paths)
}

func protectedVMPaths(excludeVMID uint) ([]string, error) {
	vms, err := vmRepo.List()
	if err != nil {
		return nil, err
	}
	templates, err := vmRepo.ListTemplates()
	if err != nil {
		return nil, err
	}
	paths := make([]string, 0, 2*len(vms)+len(templates))
	for _, item := range vms {
		if item.ID == excludeVMID {
			continue
		}
		if item.DiskPath != "" {
			paths = append(paths, item.DiskPath)
		}
		if item.ImagePath != "" {
			paths = append(paths, item.ImagePath)
		}
	}
	for _, item := range templates {
		if item.DiskPath != "" {
			paths = append(paths, item.DiskPath)
		}
	}
	return paths, nil
}

func NewIVMService() IVMService {
	provider, _ := vm.NewVMProvider()
	return &VMService{provider: provider, locks: make(map[string]*vmLockEntry)}
}

func (v *VMService) CheckEnvironment() (dto.VMEnvironment, error) {
	arch := runtime.GOARCH
	env := dto.VMEnvironment{
		Arch:          arch,
		ArchSupported: arch == "amd64" || arch == "arm64",
	}
	if !env.ArchSupported {
		return env, nil
	}
	helperExist := v.provider.HelperStatus()
	env.HelperExist = helperExist
	if !helperExist {
		env.HelperMessage = buserr.New("ErrVMHelperNotFound").Error()
	}
	helperReady := helperExist
	if helperExist {
		checkItems, checkErr := v.provider.CheckEnvironment(context.Background())
		if checkErr != nil {
			helperReady = false
			env.HelperMessage = checkErr.Error()
		}
		for _, item := range checkItems {
			switch item.Name {
			case "kvm":
				env.KVMExist = item.IsExist
				env.KVMMessage = item.Message
			case "storage":
				env.StorageExist = item.IsExist
				env.StorageMessage = item.Message
			case "libvirt":
				env.LibvirtExist = item.IsExist
				env.LibvirtActive = item.Passed || item.IsActive
				env.LibvirtMessage = item.Message
			}
		}
	}
	if libvirtExist, _, _ := v.loadLibvirtServiceStatus(); libvirtExist {
		env.LibvirtExist = true
	}
	env.Ready = helperReady && env.KVMExist && env.StorageExist && env.LibvirtActive
	return env, nil
}

func (v *VMService) HealthCheck(ctx context.Context, name string) (dto.VMHealthCheckResult, error) {
	if !v.helperExists() {
		return dto.VMHealthCheckResult{}, buserr.New("ErrVMHelperNotFound")
	}
	return v.provider.HealthCheck(ctx, name)
}

func (v *VMService) EnableEnvironment() error {
	client, err := controller.New()
	if err != nil {
		return err
	}
	found := false
	for _, service := range libvirtServiceNames {
		exist, _ := client.IsExist(service)
		if !exist {
			continue
		}
		found = true
		if err := client.Operate("start", service); err != nil {
			return err
		}
	}
	if !found {
		return buserr.New("ErrRecordNotFound")
	}
	return nil
}

func (v *VMService) loadLibvirtServiceStatus() (bool, bool, bool) {
	client, err := controller.New()
	if err != nil {
		return false, false, false
	}
	isExist := false
	isEnable := false
	isActive := false
	for _, service := range libvirtServiceNames {
		exist, _ := client.IsExist(service)
		if !exist {
			continue
		}
		isExist = true
		if enable, _ := client.IsEnable(service); enable {
			isEnable = true
		}
		if active, _ := client.IsActive(service); active {
			isActive = true
		}
		if isEnable && isActive {
			break
		}
	}
	return isExist, isEnable, isActive
}

func (v *VMService) SyncResourcesOnStartup() error {
	env, err := v.CheckEnvironment()
	if err != nil {
		return err
	}
	if !env.HelperExist {
		global.LOG.Info("skip VM resource sync on startup: 1panel-kvm not found")
		return nil
	}
	if !env.KVMExist {
		global.LOG.Info("skip VM resource sync on startup: kvm not found")
		return nil
	}
	if !env.LibvirtActive {
		global.LOG.Info("skip VM resource sync on startup: libvirt is not active")
		return nil
	}
	if err := v.SyncNetworks(); err != nil {
		return fmt.Errorf("sync VM networks failed: %w", err)
	}
	if err := v.SyncStorages(); err != nil {
		return fmt.Errorf("sync VM storage pools failed: %w", err)
	}
	if err := v.SyncVMs(); err != nil {
		return fmt.Errorf("sync virtual machines failed: %w", err)
	}
	global.LOG.Info("VM networks, storage pools, and virtual machines synced on startup")
	return nil
}

// SyncVMs imports libvirt domains into the panel metadata and refreshes metadata
// for already registered VMs. Metadata maintained only by the panel, such as the
// description and image path, is intentionally retained.
func (v *VMService) SyncVMs() error {
	vmCreateMu.Lock()
	defer vmCreateMu.Unlock()

	if !v.helperExists() {
		return buserr.New("ErrVMHelperNotFound")
	}
	liveItems, err := v.provider.List(context.Background())
	if err != nil {
		return err
	}
	v.fillVMStoragePools(liveItems)
	records, err := vmRepo.List()
	if err != nil {
		return err
	}
	recordsByName := make(map[string]model.VirtualMachine, len(records))
	for _, record := range records {
		recordsByName[record.Name] = record
	}
	liveByName := make(map[string]struct{}, len(liveItems))
	for _, live := range liveItems {
		if live.Name == "" {
			continue
		}
		liveByName[live.Name] = struct{}{}
		if record, ok := recordsByName[live.Name]; ok {
			if err := vmRepo.Update(record.ID, vmSyncMetadataMap(live)); err != nil {
				return err
			}
			continue
		}
		if err := vmRepo.Create(&model.VirtualMachine{
			Name: live.Name, UUID: live.UUID, CPU: live.CPU, Memory: live.Memory,
			DiskSize: live.DiskSize, DiskBus: defaultDiskBus(live.DiskBus), DiskPath: live.DiskPath, StoragePool: live.StoragePool,
			ISOPaths: live.ISOPaths, Network: live.Network, BootOrder: live.BootOrder, AutoStart: live.AutoStart, Status: live.Status,
		}); err != nil {
			return err
		}
	}
	for _, record := range records {
		if _, ok := liveByName[record.Name]; ok || record.UUID == "" {
			continue
		}
		if err := vmRepo.Update(record.ID, map[string]interface{}{"status": strings.ToLower(constant.StatusDeleted)}); err != nil {
			return err
		}
	}
	return nil
}

func vmSyncMetadataMap(item dto.VMInfo) map[string]interface{} {
	vars := map[string]interface{}{
		"uuid": item.UUID, "cpu": item.CPU, "memory": item.Memory, "disk_bus": defaultDiskBus(item.DiskBus),
		"disk_path": item.DiskPath, "iso_paths": item.ISOPaths,
		"network": item.Network, "boot_order": item.BootOrder, "auto_start": item.AutoStart, "status": item.Status,
	}
	if item.StoragePool != "" {
		vars["storage_pool"] = item.StoragePool
	}
	if item.DiskSize > 0 {
		vars["disk_size"] = item.DiskSize
	}
	return vars
}

func (v *VMService) Search(req dto.SearchWithPage) (int64, interface{}, error) {
	options := []repo.DBOption{repo.WithOrderDesc("created_at")}
	if req.Info != "" {
		options = append(options, repo.WithByName(req.Info))
	}
	total, records, err := vmRepo.Page(req.Page, req.PageSize, options...)
	if err != nil {
		return 0, nil, err
	}
	var domains []dto.VMInfo
	if v.helperExists() {
		domains, err = v.provider.List(context.Background())
		if err != nil {
			global.LOG.Warnf("[xpack] list live VM domains failed, fallback to metadata: %v", err)
		}
	}
	domainByName := make(map[string]dto.VMInfo, len(domains))
	for _, domain := range domains {
		domainByName[domain.Name] = domain
	}
	list := make([]dto.VMInfo, 0, len(records))
	for _, record := range records {
		if domain, ok := domainByName[record.Name]; ok {
			mergeVMMetadata(&domain, record)
			list = append(list, domain)
			continue
		}
		var item dto.VMInfo
		if err := copier.Copy(&item, &record); err != nil {
			return 0, nil, buserr.WithDetail("ErrStructTransform", err.Error(), nil)
		}
		v.fillMetadataVMStatus(&item)
		list = append(list, item)
	}
	return total, list, nil
}

func (v *VMService) Get(name string, checkDisks bool) (dto.VMInfo, error) {
	if !v.helperExists() {
		return dto.VMInfo{}, buserr.New("ErrVMHelperNotFound")
	}
	item, err := v.provider.Get(context.Background(), name, checkDisks)
	if err != nil {
		return dto.VMInfo{}, err
	}
	if record, err := vmRepo.Get(repo.WithByName(name)); err == nil {
		mergeVMMetadata(&item, record)
	}
	return item, nil
}

func (v *VMService) GetStats(name string) (dto.VMStats, error) {
	if !v.helperExists() {
		return dto.VMStats{}, buserr.New("ErrVMHelperNotFound")
	}
	return v.provider.Stats(context.Background(), name)
}

func (v *VMService) Create(req dto.VMCreate, limit int) (dto.VMTaskResp, error) {
	vmCreateMu.Lock()
	defer vmCreateMu.Unlock()

	if limit > 0 {
		count, err := vmRepo.Count()
		if err != nil {
			return dto.VMTaskResp{}, err
		}
		if count >= int64(limit) {
			return dto.VMTaskResp{}, buserr.WithMap("ErrVMCountLimit", map[string]interface{}{"limit": limit}, nil)
		}
	}
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	if _, err := vmRepo.Get(repo.WithByName(req.Name)); err == nil {
		return dto.VMTaskResp{}, buserr.New("ErrRecordExist")
	} else if err != gorm.ErrRecordNotFound {
		return dto.VMTaskResp{}, err
	}
	if err := v.validateMemoryLimit(req.Memory, ""); err != nil {
		return dto.VMTaskResp{}, err
	}
	if req.TemplateID == 0 && len(req.ISOPaths) == 0 {
		return dto.VMTaskResp{}, buserr.WithDetail("ErrVMInvalidInput", "ISO path is required", nil)
	}
	cfg := types.VMConfig{
		Name: req.Name, CPU: req.CPU, MemoryMiB: req.Memory, DiskSizeGiB: req.DiskSize,
		DiskBus: resolveDiskBus(req.OSType, req.DiskBus), DiskPath: req.DiskPath, StoragePool: req.StoragePool, ISOPaths: req.ISOPaths,
		ImagePath: req.ImagePath, Network: req.Network, BootOrder: req.BootOrder, AutoStart: req.AutoStart,
		StartAfterCreate: req.StartAfterCreate,
	}
	cfg.BootOrder = "hd,cdrom"
	if req.TemplateID != 0 {
		template, err := vmRepo.GetTemplate(repo.WithByID(req.TemplateID))
		if err != nil {
			return dto.VMTaskResp{}, err
		}
		if _, statErr := os.Stat(template.DiskPath); statErr != nil {
			return dto.VMTaskResp{}, statErr
		}
		cfg.ImagePath = template.DiskPath
		cfg.DiskBus = defaultDiskBus(template.DiskBus)
		cfg.ISOPaths = []string{}
		cfg.BootOrder = "hd"
		if template.DiskSize > 0 && cfg.DiskSizeGiB < template.DiskSize {
			return dto.VMTaskResp{}, buserr.WithDetail("ErrVMInvalidInput", fmt.Sprintf("disk size must be at least %d GiB", template.DiskSize), nil)
		}
		if cfg.DiskSizeGiB == 0 {
			cfg.DiskSizeGiB = 1
		}
		if cfg.StoragePool == "" {
			cfg.StoragePool = template.StoragePool
		}
	}
	if err := v.validateDiskLimit(cfg.DiskSizeGiB, cfg.StoragePool); err != nil {
		return dto.VMTaskResp{}, err
	}
	createMetadataReq := v.buildCreateMetadataReq(req, cfg)
	if createMetadataReq.DiskPath != "" {
		if _, err := os.Lstat(createMetadataReq.DiskPath); err == nil {
			return dto.VMTaskResp{}, buserr.WithDetail("ErrVMImageAlreadyExists", createMetadataReq.DiskPath, nil)
		} else if !os.IsNotExist(err) {
			return dto.VMTaskResp{}, err
		}
	}
	createTaskKey := "VMCreateTask"
	if req.TemplateID != 0 {
		createTaskKey = "VMCreateFromTemplateTask"
	}
	taskName := vmTaskName(createTaskKey, req.Name)
	taskItem, err := task.NewTask(taskName, task.TaskCreate, task.TaskScopeVm, req.TaskID, 0)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	if err := v.saveMetadata(createMetadataReq, dto.VMInfo{DiskPath: createMetadataReq.DiskPath}, taskItem.Task.ID); err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			var domain dto.VMInfo
			var err error
			if req.TemplateID != 0 {
				domain, err = v.provider.CreateFromTemplate(context.Background(), cfg)
			} else {
				domain, err = v.provider.Create(context.Background(), cfg)
			}
			if err != nil {
				return err
			}
			v.syncCreatedDefaultResourceMetadata(cfg)
			return v.saveMetadata(createMetadataReq, domain, t.Task.ID)
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) ListTemplates() ([]dto.VMTemplateInfo, error) {
	records, err := vmRepo.ListTemplates()
	if err != nil {
		return nil, err
	}
	items := make([]dto.VMTemplateInfo, 0, len(records))
	for _, record := range records {
		item := dto.VMTemplateInfo{
			ID: record.ID, Name: record.Name, SourceVM: record.SourceVM, CPU: record.CPU, Memory: record.Memory,
			DiskSize: record.DiskSize, DiskBus: defaultDiskBus(record.DiskBus), DiskPath: record.DiskPath, StoragePool: record.StoragePool,
			Network: record.Network, BootOrder: record.BootOrder, CreatedAt: record.CreatedAt, Description: record.Description,
		}
		if stat, statErr := os.Stat(record.DiskPath); statErr == nil {
			item.IsExist = true
			item.Size = stat.Size()
		}
		items = append(items, item)
	}
	return items, nil
}

func (v *VMService) CreateTemplate(req dto.VMTemplateCreate) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	if item, _ := vmRepo.GetTemplate(repo.WithByName(req.TemplateName)); item.ID != 0 {
		return dto.VMTaskResp{}, buserr.New("ErrRecordExist")
	}
	source, err := v.Get(req.Name, false)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	storagePool := req.StoragePool
	if storagePool == "" {
		storagePool = source.StoragePool
	}
	if storagePool == "" {
		storagePool = "default"
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskName := vmTemplateTaskName("VMTemplateCreateTask", req.Name, req.TemplateName)
	taskItem, err := task.NewTask(taskName, task.TaskCreate, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			tpl, err := v.provider.CreateTemplate(context.Background(), types.VMTemplateCreate{
				Name: req.TemplateName, SourceName: req.Name, StoragePool: storagePool,
			})
			if err != nil {
				return err
			}
			templateRecord := &model.VMTemplate{
				Name: req.TemplateName, SourceVM: req.Name, CPU: source.CPU, Memory: source.Memory,
				DiskSize: tpl.DiskSize, DiskBus: defaultDiskBus(source.DiskBus), DiskPath: tpl.DiskPath, StoragePool: storagePool,
				Network: source.Network, BootOrder: source.BootOrder, Description: req.Description,
			}
			if err := vmRepo.CreateTemplate(templateRecord); err != nil {
				if tpl.DiskPath != "" {
					_ = os.Remove(tpl.DiskPath)
				}
				return err
			}
			return nil
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) DeleteTemplate(req dto.VMTemplateDelete) (dto.VMTaskResp, error) {
	record, err := vmRepo.GetTemplate(repo.WithByID(req.ID))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskName := vmTaskName("VMTemplateDeleteTask", record.Name)
	taskItem, err := task.NewTask(taskName, task.TaskDelete, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		if req.DeleteDisk && record.DiskPath != "" {
			if err := os.Remove(record.DiskPath); err != nil && !os.IsNotExist(err) {
				return err
			}
		}
		return vmRepo.DeleteTemplate(repo.WithByID(record.ID))
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) Update(req dto.VMUpdate) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	if err := v.validateMemoryLimit(req.Memory, req.Name); err != nil {
		return dto.VMTaskResp{}, err
	}
	cfg := types.VMConfig{
		Name: req.Name, CPU: req.CPU, MemoryMiB: req.Memory, ISOPaths: req.ISOPaths,
		Network: req.Network, BootOrder: req.BootOrder, AutoStart: req.AutoStart,
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskName := vmTaskName("VMUpdateTask", req.Name)
	taskItem, err := task.NewTask(taskName, task.TaskUpdate, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			domain, err := v.provider.Update(context.Background(), cfg)
			if err != nil {
				return err
			}
			return v.updateMetadata(req, domain, t.Task.ID)
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) ResizeDisk(req dto.VMDiskResize) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskName := i18n.GetMsgWithMap("VMDiskResizeTask", map[string]interface{}{"name": req.Name, "disk": req.Target})
	taskItem, err := task.NewTask(taskName, task.TaskUpdate, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			domain, err := v.provider.ResizeDisk(context.Background(), req)
			if err != nil {
				return err
			}
			vars := map[string]interface{}{"disk_path": domain.DiskPath, "task_id": t.Task.ID}
			if domain.DiskSize > 0 {
				vars["disk_size"] = domain.DiskSize
			}
			return vmRepo.Update(record.ID, vars)
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) Rename(req dto.VMRename) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	if req.Name == req.NewName {
		return dto.VMTaskResp{}, buserr.New("ErrRecordExist")
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	if item, _ := vmRepo.Get(repo.WithByName(req.NewName)); item.ID != 0 {
		return dto.VMTaskResp{}, buserr.New("ErrRecordExist")
	}
	taskName := vmRenameTaskName("VMRenameTask", req.Name, req.NewName)
	taskItem, err := task.NewTask(taskName, task.TaskUpdate, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		return v.withVMRenameLock(req.Name, req.NewName, func() error {
			domain, err := v.provider.Rename(context.Background(), req.Name, req.NewName)
			if err != nil {
				return err
			}
			if err := v.renameMetadata(record, req, domain, t.Task.ID); err != nil {
				_, _ = v.provider.Rename(context.Background(), req.NewName, req.Name)
				return err
			}
			return nil
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) Delete(req dto.VMDelete) (dto.VMTaskResp, error) {
	record, err := vmRepo.Get(repo.WithByID(req.ID))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	if !v.helperExists() && !req.Force {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	taskName := vmTaskName("VMDeleteTask", record.Name)
	taskItem, err := task.NewTask(taskName, task.TaskDelete, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(taskName, func(t *task.Task) error {
		return v.withVMLock(record.Name, func() error {
			if !v.helperExists() {
				if !req.Force {
					return buserr.New("ErrVMHelperNotFound")
				}
				t.Log(i18n.GetMsgWithMap("VMForceDeleteMetadataOnly", map[string]interface{}{"name": record.Name}))
			} else {
				paths, err := protectedVMPaths(record.ID)
				if err != nil {
					return err
				}
				if record.ImagePath != "" {
					paths = append(paths, record.ImagePath)
				}
				if err := v.provider.Delete(context.Background(), record.Name, req.Force, true, paths); err != nil {
					var helperErr *protocol.Error
					if errors.As(err, &helperErr) && helperErr.Code == "MAIN_DISK_DELETE_FAILED" {
						return buserr.WithErr("ErrVMCommandFailed", err)
					}
					if !req.Force {
						return err
					}
					t.Log(i18n.GetMsgWithMap("VMForceDeleteMetadataAfterLibvirtFailed", map[string]interface{}{"name": record.Name, "err": err}))
				}
			}
			if record.ID != 0 {
				return vmRepo.Delete(repo.WithByID(record.ID))
			}
			return nil
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) ListSnapshots(req dto.VMSnapshotSearch) ([]dto.VMSnapshotInfo, error) {
	if !v.helperExists() {
		return nil, buserr.New("ErrVMHelperNotFound")
	}
	items, err := v.provider.ListSnapshots(context.Background(), req.Name)
	if err != nil {
		return nil, err
	}
	sortVMSnapshotItems(items)
	return items, nil
}

func sortVMSnapshotItems(items []dto.VMSnapshotInfo) {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].CreatedAt != items[j].CreatedAt {
			return items[i].CreatedAt > items[j].CreatedAt
		}
		return items[i].Name > items[j].Name
	})
}

func (v *VMService) CreateSnapshot(req dto.VMSnapshotReq) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem, err := task.NewTask(vmSnapshotTaskName("VMSnapshotCreateTask", req), task.TaskCreate, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(vmSnapshotTaskName("VMSnapshotCreateStep", req), func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			return v.provider.CreateSnapshot(context.Background(), req.Name, req.SnapshotName, req.Description)
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) RecoverSnapshot(req dto.VMSnapshotReq) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem, err := task.NewTask(vmSnapshotTaskName("VMSnapshotRecoverTask", req), task.TaskUpdate, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(vmSnapshotTaskName("VMSnapshotRecoverStep", req), func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			if err := v.provider.RecoverSnapshot(context.Background(), req.Name, req.SnapshotName); err != nil {
				return err
			}
			domain, err := v.provider.Get(context.Background(), req.Name, false)
			if err != nil {
				return buserr.WithErr("ErrVMSnapshotMetadataSync", err)
			}
			if domain.DiskSize <= 0 {
				return buserr.New("ErrVMSnapshotMetadataSync")
			}
			if err := vmRepo.Update(record.ID, map[string]interface{}{"disk_path": domain.DiskPath, "disk_size": domain.DiskSize}); err != nil {
				return buserr.WithErr("ErrVMSnapshotMetadataSync", err)
			}
			return nil
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func (v *VMService) DeleteSnapshot(req dto.VMSnapshotReq) (dto.VMTaskResp, error) {
	if !v.helperExists() {
		return dto.VMTaskResp{}, buserr.New("ErrVMHelperNotFound")
	}
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem, err := task.NewTask(vmSnapshotTaskName("VMSnapshotDeleteTask", req), task.TaskDelete, task.TaskScopeVm, req.TaskID, record.ID)
	if err != nil {
		return dto.VMTaskResp{}, err
	}
	taskItem.AddSubTaskWithOps(vmSnapshotTaskName("VMSnapshotDeleteStep", req), func(t *task.Task) error {
		return v.withVMLock(req.Name, func() error {
			return v.provider.DeleteSnapshot(context.Background(), req.Name, req.SnapshotName)
		})
	}, nil, 0, 30*time.Minute)
	go taskItem.Execute()
	return dto.VMTaskResp{TaskID: taskItem.Task.ID}, nil
}

func vmSnapshotTaskName(key string, req dto.VMSnapshotReq) string {
	return i18n.GetMsgWithMap(key, map[string]interface{}{
		"vm":       req.Name,
		"snapshot": req.SnapshotName,
	})
}

func vmTaskName(key, name string) string {
	return i18n.GetMsgWithMap(key, map[string]interface{}{"name": name})
}

func vmRenameTaskName(key, name, newName string) string {
	return i18n.GetMsgWithMap(key, map[string]interface{}{"name": name, "newName": newName})
}

func vmTemplateTaskName(key, name, template string) string {
	return i18n.GetMsgWithMap(key, map[string]interface{}{"name": name, "template": template})
}

func (v *VMService) Operate(req dto.VMOperate) error {
	switch req.Operate {
	case "start":
		return v.lifecycle(req.Name, v.provider.Start)
	case "shutdown":
		return v.lifecycle(req.Name, v.provider.Shutdown)
	case "reboot":
		return v.lifecycle(req.Name, v.provider.Reboot)
	case "destroy":
		return v.lifecycle(req.Name, v.provider.Destroy)
	case "suspend":
		return v.lifecycle(req.Name, v.provider.Suspend)
	case "resume":
		return v.lifecycle(req.Name, v.provider.Resume)
	default:
		return buserr.WithName("ErrNotSupportType", req.Operate)
	}
}

func (v *VMService) Console(name string) (*dto.VMGraphics, error) {
	if !v.helperExists() {
		return nil, buserr.New("ErrVMHelperNotFound")
	}
	return v.provider.Console(context.Background(), name)
}

func (v *VMService) lifecycle(name string, fn func(context.Context, string) error) error {
	if !v.helperExists() {
		return buserr.New("ErrVMHelperNotFound")
	}
	return v.withVMLock(name, func() error {
		return fn(context.Background(), name)
	})
}

func (v *VMService) helperExists() bool {
	return v.provider.HelperStatus()
}

func (v *VMService) withVMLock(name string, fn func() error) error {
	lock := v.acquireVMLock(name)
	defer v.releaseVMLock(name, lock)
	return fn()
}

func (v *VMService) acquireVMLock(name string) *vmLockEntry {
	v.locksMu.Lock()
	if v.locks == nil {
		v.locks = make(map[string]*vmLockEntry)
	}
	lock := v.locks[name]
	if lock == nil {
		lock = &vmLockEntry{}
		v.locks[name] = lock
	}
	lock.refs++
	v.locksMu.Unlock()
	lock.mu.Lock()
	return lock
}

func (v *VMService) releaseVMLock(name string, lock *vmLockEntry) {
	lock.mu.Unlock()
	v.locksMu.Lock()
	lock.refs--
	if lock.refs == 0 && v.locks[name] == lock {
		delete(v.locks, name)
	}
	v.locksMu.Unlock()
}

func (v *VMService) withVMRenameLock(name string, newName string, fn func() error) error {
	if name > newName {
		name, newName = newName, name
	}
	return v.withVMLock(name, func() error {
		return v.withVMLock(newName, fn)
	})
}

func (v *VMService) saveMetadata(req dto.VMCreate, domain dto.VMInfo, taskID string) error {
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	diskBus := defaultDiskBus(req.DiskBus)
	if domain.DiskBus != "" {
		diskBus = domain.DiskBus
	}
	if domain.UUID != "" {
		req.ISOPaths = domain.ISOPaths
	}
	data := &model.VirtualMachine{
		Name: req.Name, UUID: domain.UUID, CPU: req.CPU, Memory: req.Memory, DiskSize: req.DiskSize,
		DiskBus: diskBus, DiskPath: domain.DiskPath, StoragePool: req.StoragePool, ISOPaths: req.ISOPaths,
		ImagePath: req.ImagePath, Network: req.Network, BootOrder: req.BootOrder,
		AutoStart: req.AutoStart, TaskID: taskID, CreateTaskID: taskID, Description: req.Description,
	}
	if err == nil && record.ID != 0 {
		return vmRepo.Update(record.ID, metadataUpdateMap(*data))
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return vmRepo.Create(data)
}

func (v *VMService) buildCreateMetadataReq(req dto.VMCreate, cfg types.VMConfig) dto.VMCreate {
	out := req
	out.DiskSize = cfg.DiskSizeGiB
	out.DiskBus = cfg.DiskBus
	out.DiskPath = cfg.DiskPath
	out.StoragePool = cfg.StoragePool
	out.ISOPaths = cfg.ISOPaths
	out.BootOrder = cfg.BootOrder
	if out.StoragePool == "" {
		out.StoragePool = defaultVMResourceName
	}
	if req.TemplateID != 0 {
		out.ImagePath = ""
	}
	if out.DiskPath == "" {
		out.DiskPath = v.defaultCreateDiskPath(out.Name, out.StoragePool)
	}
	return out
}

func (v *VMService) defaultCreateDiskPath(name string, storagePool string) string {
	if storagePool == "" {
		storagePool = defaultVMResourceName
	}
	if live, ok, err := v.findVMResource("storagePool", storagePool); err == nil && ok && live.Path != "" {
		return filepath.Join(live.Path, name+".qcow2")
	}
	if storagePool == defaultVMResourceName {
		return filepath.Join(defaultVMDiskDir, name+".qcow2")
	}
	return ""
}

func resolveDiskBus(osType string, diskBus string) string {
	if diskBus != "" {
		return diskBus
	}
	if osType == osTypeWindows {
		return diskBusSATA
	}
	return diskBusVirtIO
}

func defaultDiskBus(diskBus string) string {
	if diskBus == "" {
		return diskBusVirtIO
	}
	return diskBus
}

func (v *VMService) validateMemoryLimit(memoryMiB int64, currentName string) error {
	if memoryMiB < vmMemoryMinMiB {
		return buserr.WithDetail("ErrVMInvalidInput", fmt.Sprintf("memory must be at least %d MiB", vmMemoryMinMiB), nil)
	}
	memoryInfo, err := mem.VirtualMemory()
	if err != nil || memoryInfo.Total == 0 {
		return buserr.WithDetail("ErrVMInvalidInput", "failed to load host memory limit", err)
	}
	allocatedBytes, currentBytes, err := v.loadAllocatedVMMemory(currentName)
	if err != nil {
		return err
	}
	reservedBytes := uint64(float64(memoryInfo.Total) * vmMemoryReserveRatio)
	if reservedBytes < vmMemoryReserveMin {
		reservedBytes = vmMemoryReserveMin
	}
	if reservedBytes >= memoryInfo.Total {
		return buserr.WithDetail("ErrVMInvalidInput", "host memory is insufficient", nil)
	}
	availableBytes := memoryInfo.Total - reservedBytes
	if allocatedBytes > currentBytes {
		usedByOthers := allocatedBytes - currentBytes
		if usedByOthers >= availableBytes {
			return buserr.WithDetail("ErrVMInvalidInput", "host memory is insufficient", nil)
		}
		availableBytes -= usedByOthers
	}
	requestBytes := uint64(memoryMiB) * bytePerMiB
	if requestBytes > availableBytes {
		return buserr.WithDetail("ErrVMInvalidInput", fmt.Sprintf("memory exceeds available limit: %d MiB", availableBytes/bytePerMiB), nil)
	}
	return nil
}

func (v *VMService) loadAllocatedVMMemory(currentName string) (uint64, uint64, error) {
	domains, err := v.provider.List(context.Background())
	if err != nil {
		return 0, 0, buserr.WithDetail("ErrVMInvalidInput", "failed to load virtual machine memory usage", err)
	}
	var allocated uint64
	var current uint64
	for _, domain := range domains {
		if !vmMemoryAllocatedStatus(domain.Status) {
			continue
		}
		memoryBytes := uint64(domain.Memory) * bytePerMiB
		allocated += memoryBytes
		if currentName != "" && domain.Name == currentName {
			current = memoryBytes
		}
	}
	return allocated, current, nil
}

func vmMemoryAllocatedStatus(status string) bool {
	switch strings.ToLower(status) {
	case "running", "blocked", "paused", "shutdown", "pmsuspended":
		return true
	default:
		return false
	}
}

func (v *VMService) validateDiskLimit(diskSizeGiB int64, storagePool string) error {
	if diskSizeGiB < vmDiskMinGiB {
		return buserr.WithDetail("ErrVMInvalidInput", fmt.Sprintf("disk size must be at least %d GiB", vmDiskMinGiB), nil)
	}
	if storagePool == "" {
		storagePool = defaultVMResourceName
	}
	live, ok, err := v.findVMResource("storagePool", storagePool)
	if err != nil {
		return err
	}
	if !ok || live.Available == 0 {
		return nil
	}
	if uint64(diskSizeGiB)*bytePerGiB > live.Available {
		return buserr.WithDetail("ErrVMInvalidInput", fmt.Sprintf("disk size exceeds storage pool available space: %d GiB", live.Available/bytePerGiB), nil)
	}
	return nil
}

func (v *VMService) syncCreatedDefaultResourceMetadata(cfg types.VMConfig) {
	if cfg.Network == defaultVMResourceName {
		if err := v.upsertNetworkMetadata(defaultVMResourceName); err != nil {
			global.LOG.Warnf("[xpack] sync default VM network metadata failed: %v", err)
		}
	}
	storagePool := cfg.StoragePool
	if storagePool == "" {
		storagePool = defaultVMResourceName
	}
	if storagePool == defaultVMResourceName {
		if err := v.upsertStorageMetadata(defaultVMResourceName); err != nil {
			global.LOG.Warnf("[xpack] sync default VM storage metadata failed: %v", err)
		}
	}
}

func (v *VMService) upsertNetworkMetadata(name string) error {
	live, ok, err := v.findVMResource("network", name)
	if err != nil || !ok {
		return err
	}
	vars := map[string]interface{}{
		"type":             live.SubType,
		"bridge_name":      live.BridgeName,
		"source_interface": live.SourceInterface,
		"cidr":             live.CIDR,
		"gateway":          live.Gateway,
		"dhcp_start":       live.DHCPStart,
		"dhcp_end":         live.DHCPEnd,
		"auto_start":       live.AutoStart,
		"status":           normalizeVMResourceStatus(live.Status),
	}
	record, err := vmRepo.GetNetwork(repo.WithByName(name))
	if err == nil && record.ID != 0 {
		return vmRepo.UpdateNetwork(record.ID, vars)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return vmRepo.CreateNetwork(&model.VMNetwork{
		Name: live.Name, Type: live.SubType, BridgeName: live.BridgeName, SourceInterface: live.SourceInterface,
		CIDR: live.CIDR, Gateway: live.Gateway, DHCPStart: live.DHCPStart, DHCPEnd: live.DHCPEnd,
		AutoStart: live.AutoStart, Status: normalizeVMResourceStatus(live.Status),
	})
}

func (v *VMService) upsertStorageMetadata(name string) error {
	live, ok, err := v.findVMResource("storagePool", name)
	if err != nil || !ok {
		return err
	}
	vars := map[string]interface{}{
		"type":       live.SubType,
		"path":       live.Path,
		"auto_start": live.AutoStart,
		"status":     normalizeVMResourceStatus(live.Status),
	}
	record, err := vmRepo.GetStorage(repo.WithByName(name))
	if err == nil && record.ID != 0 {
		return vmRepo.UpdateStorage(record.ID, vars)
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	return vmRepo.CreateStorage(&model.VMStorage{
		Name: live.Name, Type: live.SubType, Path: live.Path,
		AutoStart: live.AutoStart, Status: normalizeVMResourceStatus(live.Status),
	})
}

func (v *VMService) findVMResource(resourceType, name string) (dto.VMResourceInfo, bool, error) {
	return v.findVMResourceWithContext(context.Background(), resourceType, name)
}

func (v *VMService) findVMResourceWithContext(ctx context.Context, resourceType, name string) (dto.VMResourceInfo, bool, error) {
	liveItems, err := v.provider.ListResources(ctx, resourceType)
	if err != nil {
		return dto.VMResourceInfo{}, false, err
	}
	for _, live := range liveItems {
		if live.Name == name {
			return live, true, nil
		}
	}
	return dto.VMResourceInfo{}, false, nil
}

func (v *VMService) updateMetadata(req dto.VMUpdate, domain dto.VMInfo, taskID string) error {
	record, err := vmRepo.Get(repo.WithByName(req.Name))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return vmRepo.Update(record.ID, map[string]interface{}{
		"uuid": domain.UUID, "cpu": req.CPU, "memory": req.Memory, "iso_paths": domain.ISOPaths,
		"network": req.Network, "boot_order": domain.BootOrder, "auto_start": req.AutoStart,
		"description": req.Description, "task_id": taskID,
	})
}

func (v *VMService) renameMetadata(record model.VirtualMachine, req dto.VMRename, domain dto.VMInfo, taskID string) error {
	vars := metadataUpdateMap(record)
	vars["name"] = req.NewName
	vars["uuid"] = domain.UUID
	vars["cpu"] = domain.CPU
	vars["memory"] = domain.Memory
	if domain.DiskSize > 0 {
		vars["disk_size"] = domain.DiskSize
	}
	vars["disk_path"] = domain.DiskPath
	vars["iso_paths"] = domain.ISOPaths
	vars["network"] = domain.Network
	vars["boot_order"] = domain.BootOrder
	vars["auto_start"] = domain.AutoStart
	vars["task_id"] = taskID
	if err := vmRepo.Update(record.ID, vars); err != nil {
		return err
	}
	if err := vmRepo.UpdateTemplateSource(req.Name, req.NewName); err != nil {
		rollbackVars := metadataUpdateMap(record)
		rollbackVars["name"] = req.Name
		_ = vmRepo.Update(record.ID, rollbackVars)
		return err
	}
	return nil
}

func metadataUpdateMap(vm model.VirtualMachine) map[string]interface{} {
	return map[string]interface{}{
		"uuid": vm.UUID, "cpu": vm.CPU, "memory": vm.Memory, "disk_size": vm.DiskSize, "disk_bus": vm.DiskBus,
		"disk_path": vm.DiskPath, "storage_pool": vm.StoragePool, "iso_paths": vm.ISOPaths,
		"image_path": vm.ImagePath, "network": vm.Network, "boot_order": vm.BootOrder,
		"auto_start": vm.AutoStart, "task_id": vm.TaskID, "create_task_id": vm.CreateTaskID, "description": vm.Description,
	}
}

func mergeVMMetadata(item *dto.VMInfo, record model.VirtualMachine) {
	item.ID = record.ID
	item.StoragePool = record.StoragePool
	item.ImagePath = record.ImagePath
	if item.DiskPath == "" {
		item.DiskPath = record.DiskPath
	}
	if item.DiskBus == "" {
		item.DiskBus = defaultDiskBus(record.DiskBus)
	}
	item.AutoStart = record.AutoStart
	item.TaskID = record.TaskID
	item.CreateTaskID = record.CreateTaskID
	item.Description = record.Description
	if item.UUID == "" {
		item.UUID = record.UUID
	}
	if item.DiskSize == 0 {
		item.DiskSize = record.DiskSize
	}
	if item.Status == "" {
		item.Status = record.Status
	}
}

func (v *VMService) fillMetadataVMStatus(item *dto.VMInfo) {
	if item.Status != "" {
		return
	}
	taskID := item.CreateTaskID
	if taskID == "" {
		taskID = item.TaskID
	}
	if taskID == "" {
		return
	}
	taskRecord, err := taskRepo.GetFirst(taskRepo.WithByID(taskID))
	if err != nil {
		return
	}
	switch taskRecord.Status {
	case constant.StatusFailed:
		item.Status = "failed"
	case constant.StatusExecuting, constant.StatusWaiting, constant.StatusRunning:
		item.Status = "creating"
	}
}
