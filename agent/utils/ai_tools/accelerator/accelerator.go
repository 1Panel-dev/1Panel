package accelerator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
)

type Client struct {
	providers []Provider
}

type providerResult struct {
	kind     Kind
	snapshot *ProviderSnapshot
	err      error
}

func New() (bool, Client) {
	client := Client{}
	if available, gpuClient := gpu.New(); available {
		client.providers = append(client.providers, gpuProvider{client: gpuClient})
	}
	if available, npuClient := npu.New(); available {
		client.providers = append(client.providers, npuProvider{client: npuClient})
	}
	if available, xpuClient := xpu.New(); available {
		client.providers = append(client.providers, xpuProvider{client: xpuClient})
	}
	return len(client.providers) > 0, client
}

func (c Client) LoadInfo() (*Info, error) {
	snapshot, err := c.Collect(context.Background())
	if err != nil {
		return nil, err
	}
	return &snapshot.Info, nil
}

func (c Client) Collect(ctx context.Context) (*Snapshot, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	collectCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	results := make([]providerResult, len(c.providers))
	var wg sync.WaitGroup
	for index, item := range c.providers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			providerSnapshot, err := item.Collect(collectCtx)
			if err != nil {
				err = fmt.Errorf("%s provider failed: %w", item.Kind(), err)
			}
			results[index] = providerResult{kind: item.Kind(), snapshot: providerSnapshot, err: err}
		}()
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	snapshot := &Snapshot{Info: Info{VendorWarnings: map[string][]string{}}}
	var (
		errs       []error
		active     []*ProviderSnapshot
		xpuVersion string
	)
	if err := collectCtx.Err(); err != nil {
		snapshot.Info.Warnings = append(snapshot.Info.Warnings, err.Error())
	}
	for _, result := range results {
		vendor := ""
		switch result.kind {
		case KindXPU:
			vendor = "intel"
		case KindNPU:
			vendor = "ascend"
		}
		if result.err != nil && (result.snapshot == nil || len(result.snapshot.VendorWarnings) == 0) {
			if vendor == "" {
				snapshot.Info.Warnings = append(snapshot.Info.Warnings, result.err.Error())
			} else {
				snapshot.Info.VendorWarnings[vendor] = append(snapshot.Info.VendorWarnings[vendor], result.err.Error())
			}
		}
		if result.snapshot != nil {
			for key, warnings := range result.snapshot.VendorWarnings {
				if len(warnings) > 0 {
					snapshot.Info.VendorWarnings[key] = append(snapshot.Info.VendorWarnings[key], warnings...)
				}
			}
			if vendor != "" && len(result.snapshot.Warnings) > 0 {
				snapshot.Info.VendorWarnings[vendor] = append(snapshot.Info.VendorWarnings[vendor], result.snapshot.Warnings...)
			} else {
				snapshot.Info.Warnings = append(snapshot.Info.Warnings, result.snapshot.Warnings...)
			}
		}
		if result.snapshot == nil || len(result.snapshot.Devices) == 0 {
			continue
		}
		item := result.snapshot
		active = append(active, item)
		snapshot.Devices = append(snapshot.Devices, item.Devices...)
		snapshot.Info.GPUs = append(snapshot.Info.GPUs, item.GPUs...)
		snapshot.Info.NPUs = append(snapshot.Info.NPUs, item.NPUs...)
		snapshot.Info.XPUs = append(snapshot.Info.XPUs, item.XPUs...)
		if item.CudaVersion != "" {
			snapshot.Info.CudaVersion = item.CudaVersion
		}
		if item.Type == "xpu" {
			xpuVersion = item.DriverVersion
		}
	}
	warnings := append([]string{}, snapshot.Info.Warnings...)
	for _, vendorWarnings := range snapshot.Info.VendorWarnings {
		warnings = append(warnings, vendorWarnings...)
	}
	for _, device := range snapshot.Devices {
		if device.ParentID != "" {
			continue
		}
		switch {
		case device.GPU != nil:
			warnings = append(warnings, device.GPU.Warnings...)
		case device.NPU != nil:
			warnings = append(warnings, device.NPU.Warnings...)
		case device.XPU != nil:
			warnings = append(warnings, device.XPU.Warnings...)
		}
	}
	for _, warning := range warnings {
		errs = append(errs, errors.New(warning))
	}
	snapshot.Warnings = errs
	snapshot.Info.CollectedAt = time.Now()
	snapshot.Info.XPUDriverVersion = xpuVersion
	snapshot.Info.Type, snapshot.Info.DriverVersion = mergeProviderMetadata(active)
	if len(snapshot.Devices) == 0 && len(errs) > 0 {
		return snapshot, fmt.Errorf("calling accelerator monitoring tools failed: %w", errors.Join(errs...))
	}
	return snapshot, nil
}

func mergeProviderMetadata(items []*ProviderSnapshot) (string, string) {
	if len(items) == 0 {
		return "", ""
	}
	resultType := items[0].Type
	if len(items) > 1 || resultType == "mixed" {
		resultType = "mixed"
	}
	driverVersions := make([]string, 0, len(items))
	for _, item := range items {
		if item.DriverVersion == "" {
			continue
		}
		if len(items) == 1 && item.Type != "mixed" {
			driverVersions = append(driverVersions, item.DriverVersion)
			continue
		}
		if item.Type == "mixed" {
			driverVersions = append(driverVersions, item.DriverVersion)
		} else {
			driverVersions = append(driverVersions, fmt.Sprintf("%s: %s", strings.ToUpper(item.Type), item.DriverVersion))
		}
	}
	return resultType, strings.Join(driverVersions, "；")
}
