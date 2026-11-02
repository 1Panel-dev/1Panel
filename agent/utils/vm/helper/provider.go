package helper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/protocol"
	"github.com/1Panel-dev/1Panel/agent/utils/vm/types"
)

const (
	defaultTimeout      = 2 * time.Minute
	longRunningTimeout  = 30 * time.Minute
	slowActionThreshold = 5 * time.Second
)

type Provider struct {
	HelperPath  string
	URI         string
	Timeout     time.Duration
	mu          sync.Mutex
	versionOnce sync.Once
	versionGate chan struct{}
	checked     bool
}

func NewProvider() (*Provider, error) {
	return &Provider{URI: protocol.DefaultURI, Timeout: defaultTimeout}, nil
}

func (p *Provider) CheckEnvironment(ctx context.Context) ([]dto.VMCheckItem, error) {
	var items []protocol.VMCheckItem
	if err := p.call(ctx, protocol.ActionCheck, nil, &items); err != nil {
		return nil, err
	}
	out := make([]dto.VMCheckItem, 0, len(items))
	for _, item := range items {
		out = append(out, dto.VMCheckItem{
			Name:     item.Name,
			Passed:   item.Passed,
			Message:  item.Message,
			IsEnable: item.IsEnable,
			IsExist:  item.IsExist,
			IsActive: item.IsActive,
		})
	}
	return out, nil
}

func (p *Provider) HealthCheck(ctx context.Context, name string) (dto.VMHealthCheckResult, error) {
	var result protocol.VMHealthCheckResult
	if err := p.callWithTimeout(
		ctx,
		protocol.ActionHealthCheck,
		protocol.VMHealthCheckRequest{Name: name},
		&result,
		longRunningTimeout,
	); err != nil {
		return dto.VMHealthCheckResult{}, err
	}
	out := dto.VMHealthCheckResult{
		Pass:  result.Pass,
		Warn:  result.Warn,
		Fail:  result.Fail,
		Items: make([]dto.VMHealthCheckItem, 0, len(result.Items)),
	}
	for _, item := range result.Items {
		out.Items = append(out.Items, dto.VMHealthCheckItem{
			Level: item.Level, Check: item.Check, Reason: item.Reason, Params: item.Params,
			Detail: item.Detail, FixCommand: item.FixCommand,
		})
	}
	return out, nil
}

func (p *Provider) HelperStatus() bool {
	if _, err := p.resolveHelperPath(); err != nil {
		return false
	}
	return true
}

func (p *Provider) List(ctx context.Context) ([]dto.VMInfo, error) {
	var items []protocol.VMInfo
	if err := p.call(ctx, protocol.ActionList, nil, &items); err != nil {
		return nil, err
	}
	out := make([]dto.VMInfo, 0, len(items))
	for _, item := range items {
		out = append(out, toDTO(item))
	}
	return out, nil
}

func (p *Provider) ListResources(ctx context.Context, resourceType string) ([]dto.VMResourceInfo, error) {
	var items []protocol.VMResourceInfo
	if err := p.call(ctx, protocol.ActionResources, protocol.VMResourceSearch{Type: resourceType}, &items); err != nil {
		return nil, err
	}
	out := make([]dto.VMResourceInfo, 0, len(items))
	for _, item := range items {
		out = append(out, dto.VMResourceInfo{
			Type: item.Type, Name: item.Name, SubType: item.SubType, BridgeName: item.BridgeName, SourceInterface: item.SourceInterface,
			CIDR: item.CIDR, Gateway: item.Gateway, DHCPStart: item.DHCPStart, DHCPEnd: item.DHCPEnd,
			Path: item.Path, Status: item.Status, AutoStart: item.AutoStart, Capacity: item.Capacity, Available: item.Available,
		})
	}
	return out, nil
}

func (p *Provider) CreateResource(ctx context.Context, req types.VMResourceCreate) error {
	return p.call(ctx, protocol.ActionCreateResource, protocol.VMResourceCreate{
		Type: req.Type, Name: req.Name, SubType: req.SubType, BridgeName: req.BridgeName, SourceInterface: req.SourceInterface,
		CIDR: req.CIDR, Gateway: req.Gateway, DHCPStart: req.DHCPStart, DHCPEnd: req.DHCPEnd,
		Path: req.Path, AutoStart: req.AutoStart,
	}, nil)
}

func (p *Provider) UpdateResource(ctx context.Context, req types.VMResourceUpdate) error {
	return p.call(ctx, protocol.ActionUpdateResource, protocol.VMResourceUpdate{
		Type: req.Type, Name: req.Name, AutoStart: req.AutoStart, Status: req.Status,
	}, nil)
}

func (p *Provider) DeleteResource(ctx context.Context, resourceType string, name string) error {
	return p.call(ctx, protocol.ActionDeleteResource, protocol.VMResourceDelete{Type: resourceType, Name: name}, nil)
}

func (p *Provider) ScanOrphans(ctx context.Context, protectedPaths []string) ([]dto.VMOrphanItem, error) {
	var items []protocol.VMOrphanItem
	if err := p.callWithTimeout(ctx, protocol.ActionScanOrphans, protocol.VMOrphanScan{ProtectedPaths: protectedPaths}, &items, longRunningTimeout); err != nil {
		return nil, err
	}
	out := make([]dto.VMOrphanItem, 0, len(items))
	for _, item := range items {
		out = append(out, dto.VMOrphanItem{Type: item.Type, Path: item.Path, Size: item.Size, ModifiedAt: item.ModifiedAt})
	}
	return out, nil
}

func (p *Provider) CleanOrphans(ctx context.Context, items []dto.VMOrphanItem, protectedPaths []string) ([]dto.VMOrphanItem, error) {
	payload := protocol.VMOrphanClean{ProtectedPaths: protectedPaths, Items: make([]protocol.VMOrphanItem, 0, len(items))}
	for _, item := range items {
		payload.Items = append(payload.Items, protocol.VMOrphanItem{Type: item.Type, Path: item.Path, Size: item.Size, ModifiedAt: item.ModifiedAt})
	}
	var removed []protocol.VMOrphanItem
	if err := p.callWithTimeout(ctx, protocol.ActionCleanOrphans, payload, &removed, longRunningTimeout); err != nil {
		return nil, err
	}
	out := make([]dto.VMOrphanItem, 0, len(removed))
	for _, item := range removed {
		out = append(out, dto.VMOrphanItem{Type: item.Type, Path: item.Path, Size: item.Size, ModifiedAt: item.ModifiedAt})
	}
	return out, nil
}

func (p *Provider) Get(ctx context.Context, name string, checkDisks bool) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.call(ctx, protocol.ActionGet, protocol.VMInfoReq{Name: name, CheckDisks: checkDisks}, &item); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) Stats(ctx context.Context, name string) (dto.VMStats, error) {
	var item protocol.VMStats
	if err := p.call(ctx, protocol.ActionStats, protocol.VMName{Name: name}, &item); err != nil {
		return dto.VMStats{}, err
	}
	return dto.VMStats{
		Name: item.Name, Status: item.Status, Timestamp: item.Timestamp, CPUTime: item.CPUTime, VCPU: item.VCPU,
		Memory: dto.VMMemoryStats(item.Memory), Disk: dto.VMDiskStats(item.Disk), Network: dto.VMNetworkStats(item.Network),
	}, nil
}

func (p *Provider) Create(ctx context.Context, cfg types.VMConfig) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.call(ctx, protocol.ActionCreate, toProtocolConfig(cfg), &item); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) CreateFromTemplate(ctx context.Context, cfg types.VMConfig) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.callWithTimeout(ctx, protocol.ActionCreateFromTemplate, toProtocolConfig(cfg), &item, longRunningTimeout); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) CreateTemplate(ctx context.Context, req types.VMTemplateCreate) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.callWithTimeout(ctx, protocol.ActionCreateTemplate, protocol.VMTemplateCreate{
		Name: req.Name, SourceName: req.SourceName, DiskPath: req.DiskPath, StoragePool: req.StoragePool,
	}, &item, longRunningTimeout); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) Update(ctx context.Context, cfg types.VMConfig) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.call(ctx, protocol.ActionUpdate, toProtocolConfig(cfg), &item); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) ResizeDisk(ctx context.Context, req dto.VMDiskResize) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.callWithTimeout(ctx, protocol.ActionResizeDisk, protocol.VMDiskResize{Name: req.Name, Target: req.Target, SizeGiB: req.SizeGiB}, &item, longRunningTimeout); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) Rename(ctx context.Context, name string, newName string) (dto.VMInfo, error) {
	var item protocol.VMInfo
	if err := p.call(ctx, protocol.ActionRename, protocol.VMRename{Name: name, NewName: newName}, &item); err != nil {
		return dto.VMInfo{}, err
	}
	return toDTO(item), nil
}

func (p *Provider) Delete(ctx context.Context, name string, force bool, deleteDisk bool, protectedPaths []string) error {
	return p.call(ctx, protocol.ActionDelete, protocol.VMDelete{Name: name, Force: force, DeleteDisk: deleteDisk, ProtectedPaths: protectedPaths}, nil)
}

func (p *Provider) ListSnapshots(ctx context.Context, name string) ([]dto.VMSnapshotInfo, error) {
	var items []protocol.VMSnapshotInfo
	if err := p.call(ctx, protocol.ActionListSnapshots, protocol.VMName{Name: name}, &items); err != nil {
		return nil, err
	}
	out := make([]dto.VMSnapshotInfo, 0, len(items))
	for _, item := range items {
		out = append(out, dto.VMSnapshotInfo(item))
	}
	return out, nil
}

func (p *Provider) CreateSnapshot(ctx context.Context, name string, snapshotName string, description string) error {
	return p.call(ctx, protocol.ActionCreateSnapshot, protocol.VMSnapshotReq{
		Name: name, SnapshotName: snapshotName, Description: description,
	}, nil)
}

func (p *Provider) RecoverSnapshot(ctx context.Context, name string, snapshotName string) error {
	return p.call(ctx, protocol.ActionRecoverSnapshot, protocol.VMSnapshotReq{Name: name, SnapshotName: snapshotName}, nil)
}

func (p *Provider) DeleteSnapshot(ctx context.Context, name string, snapshotName string) error {
	return p.call(ctx, protocol.ActionDeleteSnapshot, protocol.VMSnapshotReq{Name: name, SnapshotName: snapshotName}, nil)
}

func (p *Provider) Start(ctx context.Context, name string) error {
	return p.lifecycle(ctx, protocol.ActionStart, name)
}

func (p *Provider) Shutdown(ctx context.Context, name string) error {
	return p.lifecycle(ctx, protocol.ActionShutdown, name)
}

func (p *Provider) Reboot(ctx context.Context, name string) error {
	return p.lifecycle(ctx, protocol.ActionReboot, name)
}

func (p *Provider) Destroy(ctx context.Context, name string) error {
	return p.lifecycle(ctx, protocol.ActionDestroy, name)
}

func (p *Provider) Suspend(ctx context.Context, name string) error {
	return p.lifecycle(ctx, protocol.ActionSuspend, name)
}

func (p *Provider) Resume(ctx context.Context, name string) error {
	return p.lifecycle(ctx, protocol.ActionResume, name)
}

func (p *Provider) Console(ctx context.Context, name string) (*dto.VMGraphics, error) {
	var item protocol.VMGraphics
	if err := p.call(ctx, protocol.ActionConsole, protocol.VMName{Name: name}, &item); err != nil {
		return nil, err
	}
	return (*dto.VMGraphics)(&item), nil
}

func (p *Provider) lifecycle(ctx context.Context, action string, name string) error {
	return p.call(ctx, action, protocol.VMName{Name: name}, nil)
}

func (p *Provider) checkVersion(ctx context.Context) error {
	var info protocol.VersionInfo
	if err := p.call(ctx, protocol.ActionVersion, nil, &info); err != nil {
		return err
	}
	if info.API != protocol.MinVersion {
		return buserr.WithErr("ErrVMCommandFailed", fmt.Errorf("unsupported kvm helper api version: %s", info.API))
	}
	return nil
}

func (p *Provider) call(ctx context.Context, action string, payload interface{}, out interface{}) error {
	return p.callWithTimeout(ctx, action, payload, out, p.Timeout)
}

func (p *Provider) callWithTimeout(ctx context.Context, action string, payload interface{}, out interface{}, timeout time.Duration) (retErr error) {
	startedAt := time.Now()
	defer func() {
		logAction(action, time.Since(startedAt), retErr)
	}()
	if ctx == nil {
		ctx = context.Background()
	}
	if timeout == 0 {
		timeout = defaultTimeout
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	if action != protocol.ActionVersion {
		if err := p.ensureVersion(runCtx); err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				return buserr.New("ErrCmdTimeout")
			}
			return err
		}
	}

	req := protocol.Request{Version: protocol.Version, Action: action, URI: p.URI, Payload: payload}
	input, err := json.Marshal(req)
	if err != nil {
		return err
	}
	helperPath, err := p.resolveHelperPath()
	if err != nil {
		return err
	}
	cmd := exec.Command(helperPath)
	cmd.Stdin = bytes.NewReader(input)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = runHelperCommand(runCtx, cmd)
	if stderr.Len() > 0 && global.LOG != nil {
		global.LOG.Warnf("kvm helper stderr: %s", stderr.String())
	}
	if errors.Is(runCtx.Err(), context.DeadlineExceeded) {
		return buserr.New("ErrCmdTimeout")
	}
	if err != nil {
		return buserr.WithErr("ErrVMCommandFailed", fmt.Errorf("helper failed: %w, stderr: %s", err, stderr.String()))
	}
	var resp protocol.Response
	if err := json.Unmarshal(stdout.Bytes(), &resp); err != nil {
		return buserr.WithErr("ErrVMCommandFailed", fmt.Errorf("invalid helper response: %w", err))
	}
	if !resp.Success {
		return mapHelperError(resp.Error)
	}
	if out == nil {
		return nil
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}

func logAction(action string, duration time.Duration, err error) {
	if global.LOG == nil {
		return
	}
	durationMs := duration.Milliseconds()
	if err != nil {
		global.LOG.Errorf("[kvm] action failed, action=%s, duration_ms=%d, err=%v", action, durationMs, err)
		return
	}
	if duration >= slowActionThreshold {
		global.LOG.Warnf("[kvm] slow action, action=%s, duration_ms=%d", action, durationMs)
	}
}

func runHelperCommand(ctx context.Context, cmd *exec.Cmd) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		if cmd.Process != nil {
			_ = cmd.Process.Signal(os.Interrupt)
		}
		timer := time.NewTimer(5 * time.Second)
		defer timer.Stop()
		select {
		case err := <-done:
			return err
		case <-timer.C:
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
			return <-done
		}
	}
}

func (p *Provider) ensureVersion(ctx context.Context) error {
	p.versionOnce.Do(func() {
		p.versionGate = make(chan struct{}, 1)
	})
	select {
	case p.versionGate <- struct{}{}:
		defer func() { <-p.versionGate }()
	case <-ctx.Done():
		return ctx.Err()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if p.checked {
		return nil
	}
	if err := p.checkVersion(ctx); err != nil {
		return err
	}
	p.checked = true
	return nil
}

func (p *Provider) resolveHelperPath() (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.HelperPath != "" {
		return p.HelperPath, nil
	}
	path, err := findHelperPath()
	if err != nil {
		return "", err
	}
	p.HelperPath = path
	return path, nil
}

func mapHelperError(item *protocol.Error) error {
	if item == nil {
		return buserr.New("ErrVMCommandFailed")
	}
	switch item.Code {
	case "MAIN_DISK_DELETE_FAILED":
		return item
	case "VM_NOT_FOUND":
		return buserr.New("ErrVMNotFound")
	case "VM_ALREADY_EXISTS":
		return buserr.New("ErrVMAlreadyExists")
	case "VM_IMAGE_ALREADY_EXISTS":
		return buserr.WithDetail("ErrVMImageAlreadyExists", item.Detail, errors.New(item.Detail))
	case "DISK_RESIZE_BLOCKED":
		if item.Detail == "referenceCheck" {
			return buserr.New("ErrVMDiskReferenceCheck")
		}
		if item.Detail == "referenced" {
			return buserr.New("ErrVMDiskReferenced")
		}
		return buserr.New("ErrVMDiskResizeBlocked")
	case "VM_RUNNING":
		return buserr.New("ErrVMRunning")
	case "LIBVIRT_CONNECT":
		return buserr.WithDetail("ErrVMLibvirtConnect", item.Detail, errors.New(item.Detail))
	case "PERMISSION_DENIED":
		return buserr.WithDetail("ErrVMPermission", item.Detail, errors.New(item.Detail))
	case "NETWORK_NOT_FOUND":
		return buserr.WithDetail("ErrVMNetworkNotFound", item.Detail, errors.New(item.Detail))
	case "INVALID_PATH":
		return buserr.WithDetail("ErrVMInvalidPath", item.Detail, errors.New(item.Detail))
	case "INVALID_INPUT":
		return buserr.WithDetail("ErrVMInvalidInput", item.Detail, errors.New(item.Detail))
	default:
		detail := item.Detail
		if detail == "" {
			detail = item.Message
		}
		return buserr.WithErr("ErrVMCommandFailed", errors.New(detail))
	}
}

func findHelperPath() (string, error) {
	if path := os.Getenv("ONEPANEL_KVM"); path != "" {
		return path, nil
	}
	self, err := os.Executable()
	if err == nil {
		candidate := filepath.Join(filepath.Dir(self), protocol.HelperBinary)
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	if path, err := exec.LookPath(protocol.HelperBinary); err == nil {
		return path, nil
	}
	return "", buserr.WithErr("ErrVMCommandFailed", fmt.Errorf("%s not found", protocol.HelperBinary))
}

func toProtocolConfig(cfg types.VMConfig) protocol.VMConfig {
	return protocol.VMConfig{
		Name: cfg.Name, CPU: cfg.CPU, MemoryMiB: cfg.MemoryMiB, DiskSizeGiB: cfg.DiskSizeGiB,
		DiskBus: cfg.DiskBus, DiskPath: cfg.DiskPath, StoragePool: cfg.StoragePool, ISOPaths: cfg.ISOPaths, ImagePath: cfg.ImagePath,
		Network: cfg.Network, BootOrder: cfg.BootOrder, AutoStart: cfg.AutoStart,
		StartAfterCreate: cfg.StartAfterCreate,
	}
}

func toDTO(item protocol.VMInfo) dto.VMInfo {
	out := dto.VMInfo{
		Name: item.Name, UUID: item.UUID, CPU: item.CPU, Memory: item.Memory, DiskSize: item.DiskSize,
		DiskBus: item.DiskBus, DiskPath: item.DiskPath, StoragePool: item.StoragePool, ISOPaths: item.ISOPaths, ImagePath: item.ImagePath,
		Network: item.Network, BootOrder: item.BootOrder, AutoStart: item.AutoStart, Status: item.Status,
	}
	for _, disk := range item.Disks {
		out.Disks = append(out.Disks, dto.VMDisk(disk))
	}
	for _, net := range item.Interfaces {
		out.Interfaces = append(out.Interfaces, dto.VMInterface(net))
	}
	if item.Graphics != nil {
		graphics := dto.VMGraphics(*item.Graphics)
		out.Graphics = &graphics
	}
	return out
}
