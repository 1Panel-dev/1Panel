package accelerator

import (
	"context"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
)

type Provider interface {
	Kind() Kind
	Collect(context.Context) (*ProviderSnapshot, error)
}

type gpuProvider struct {
	client gpu.Client
}

func (p gpuProvider) Kind() Kind { return KindGPU }

func (p gpuProvider) Collect(ctx context.Context) (*ProviderSnapshot, error) {
	info, err := p.client.LoadInfoContext(ctx)
	if err != nil {
		return nil, err
	}
	result := &ProviderSnapshot{
		Type:          info.Type,
		DriverVersion: info.DriverVersion,
		CudaVersion:   info.CudaVersion,
		GPUs:          info.Devices,
	}
	for index := range result.GPUs {
		result.Devices = append(result.Devices, normalizeGPU(&result.GPUs[index]))
	}
	return result, nil
}

type npuProvider struct {
	client npu.Client
}

func (p npuProvider) Kind() Kind { return KindNPU }

func (p npuProvider) Collect(ctx context.Context) (*ProviderSnapshot, error) {
	info, err := p.client.LoadInfoContext(ctx)
	if err != nil {
		return nil, err
	}
	result := &ProviderSnapshot{
		Type:          info.Type,
		DriverVersion: info.DriverVersion,
		NPUs:          info.Devices,
	}
	for index := range result.NPUs {
		result.Devices = append(result.Devices, normalizeNPU(&result.NPUs[index]))
	}
	return result, nil
}

type xpuProvider struct {
	client xpu.Client
}

func (p xpuProvider) Kind() Kind { return KindXPU }

func (p xpuProvider) Collect(ctx context.Context) (*ProviderSnapshot, error) {
	info, err := p.client.LoadInfoContext(ctx)
	if err != nil {
		return nil, err
	}
	result := &ProviderSnapshot{
		Type:          info.Type,
		DriverVersion: info.DriverVersion,
		XPUs:          info.Devices,
	}
	for index := range result.XPUs {
		result.Devices = append(result.Devices, normalizeXPU(&result.XPUs[index]))
	}
	return result, nil
}
