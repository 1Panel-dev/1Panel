package accelerator

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
)

type Client struct {
	providers []Provider
}

type providerResult struct {
	snapshot *ProviderSnapshot
	err      error
}

func New() (bool, Client) {
	return newClient()
}

func NewAll() (bool, Client) {
	return New()
}

func newClient() (bool, Client) {
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
	results := make([]providerResult, len(c.providers))
	var wg sync.WaitGroup
	for index, item := range c.providers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			providerSnapshot, err := item.Collect(ctx)
			if err != nil {
				err = fmt.Errorf("%s provider failed: %w", item.Kind(), err)
			}
			results[index] = providerResult{snapshot: providerSnapshot, err: err}
		}()
	}
	wg.Wait()

	snapshot := &Snapshot{DriverVersions: make(map[string]string)}
	var (
		errs       []error
		active     []*ProviderSnapshot
		xpuVersion string
	)
	for _, result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
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
		if item.DriverVersion != "" {
			snapshot.DriverVersions[item.Type] = item.DriverVersion
			if item.Type == "xpu" {
				xpuVersion = item.DriverVersion
			}
		}
	}
	snapshot.Warnings = append(snapshot.Warnings, errs...)
	if len(snapshot.Devices) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("calling accelerator monitoring tools failed: %w", errors.Join(errs...))
	}

	snapshot.Info.XPUDriverVersion = xpuVersion
	snapshot.Info.Type, snapshot.Info.DriverVersion = mergeProviderMetadata(active)
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
