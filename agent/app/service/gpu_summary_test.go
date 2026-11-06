package service

import (
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/accelerator"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/npu"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/xpu"
)

func gpuSummaryTestMemory(mib float64) accelerator.Metrics {
	return accelerator.Metrics{MemoryTotal: accelerator.Metric{Value: &mib, Unit: "MiB"}}
}

func TestSummarizeGPUsCountsCardsWithoutDuplicatingTilesOrChips(t *testing.T) {
	snapshot := &accelerator.Snapshot{
		Info: accelerator.Info{CudaVersion: "12.8"},
		Devices: []accelerator.Device{
			{ID: "nvidia:0", Vendor: "nvidia", DriverVersion: "570.86", GPU: &gpu.Device{}, Metrics: gpuSummaryTestMemory(24576)},
			{ID: "nvidia:1", Vendor: "nvidia", DriverVersion: "570.86", GPU: &gpu.Device{}, Metrics: gpuSummaryTestMemory(24576)},
			{ID: "intel:0", Vendor: "Intel Corporation", DriverVersion: "1.3", XPU: &xpu.Device{}, Metrics: gpuSummaryTestMemory(16384)},
			{ID: "intel:0:tile:0", ParentID: "intel:0", Vendor: "Intel Corporation", XPU: &xpu.Device{}, Metrics: gpuSummaryTestMemory(8192)},
			{ID: "intel:0:tile:1", ParentID: "intel:0", Vendor: "Intel Corporation", XPU: &xpu.Device{}, Metrics: gpuSummaryTestMemory(8192)},
			{ID: "intel:0", Vendor: "Intel Corporation", DriverVersion: "1.3", XPU: &xpu.Device{}, Metrics: gpuSummaryTestMemory(16384)},
			{ID: "ascend:0:0", Vendor: "ascend", NPUIndex: 0, ChipIndex: 0, DriverVersion: "24.1", NPU: &npu.Device{}, Metrics: gpuSummaryTestMemory(32768)},
			{ID: "ascend:0:1", Vendor: "ascend", NPUIndex: 0, ChipIndex: 1, DriverVersion: "24.1", NPU: &npu.Device{}, Metrics: gpuSummaryTestMemory(32768)},
			{ID: "ascend:1:0", Vendor: "ascend", NPUIndex: 1, ChipIndex: 0, DriverVersion: "24.1", NPU: &npu.Device{}, Metrics: gpuSummaryTestMemory(65536)},
			{ID: "amd:0", Vendor: "amd", DriverVersion: "6.4", GPU: &gpu.Device{}, Metrics: gpuSummaryTestMemory(4096)},
		},
	}
	want := map[string]struct {
		cards  int
		gib    float64
		driver string
	}{
		"nvidia": {2, 48, "570.86"},
		"intel":  {1, 16, "1.3"},
		"ascend": {2, 128, "24.1"},
		"amd":    {1, 4, "6.4"},
	}
	summaries := summarizeGPUs(snapshot)
	if len(summaries) != len(want) {
		t.Fatalf("got %d vendor summaries, want %d", len(summaries), len(want))
	}
	for _, summary := range summaries {
		expected, ok := want[summary.Vendor]
		if !ok {
			t.Fatalf("unexpected vendor %q", summary.Vendor)
		}
		if summary.CardCount == nil || *summary.CardCount != expected.cards {
			t.Errorf("%s: card count = %v, want %d physical cards", summary.Vendor, summary.CardCount, expected.cards)
		}
		if summary.MemoryTotal == nil || *summary.MemoryTotal != expected.gib*1024*1024*1024 {
			t.Errorf("%s: memory total = %v, want %g GiB", summary.Vendor, summary.MemoryTotal, expected.gib)
		}
		if summary.DriverVersion != expected.driver || summary.MemoryPartial || summary.SharedMemoryTotal != nil {
			t.Errorf("%s: incorrect summary metadata: %+v", summary.Vendor, summary)
		}
		if (summary.Vendor == "nvidia" && summary.CudaVersion != "12.8") || (summary.Vendor != "nvidia" && summary.CudaVersion != "") {
			t.Errorf("%s: CUDA version %q belongs only to NVIDIA", summary.Vendor, summary.CudaVersion)
		}
	}
}

func TestSummarizeGPUsSeparatesSharedHostMemoryAndPreservesPartialTotals(t *testing.T) {
	snapshot := &accelerator.Snapshot{Devices: []accelerator.Device{
		{ID: "nvidia:0", Vendor: "nvidia", GPU: &gpu.Device{SharedMemory: true}, Metrics: gpuSummaryTestMemory(65536)},
		{ID: "nvidia:1", Vendor: "nvidia", GPU: &gpu.Device{SharedMemory: true}, Metrics: gpuSummaryTestMemory(65536)},
		{ID: "nvidia:2", Vendor: "nvidia", GPU: &gpu.Device{}, Metrics: gpuSummaryTestMemory(24576)},
		{ID: "nvidia:3", Vendor: "nvidia", GPU: &gpu.Device{}},
	}}
	result := summarizeGPUs(snapshot)
	if len(result) != 1 {
		t.Fatalf("got %d summaries, want one NVIDIA summary", len(result))
	}
	summary := result[0]
	if summary.CardCount == nil || *summary.CardCount != 4 {
		t.Errorf("card count = %v, want 4", summary.CardCount)
	}
	if summary.MemoryTotal == nil || *summary.MemoryTotal != 24*1024*1024*1024 {
		t.Errorf("dedicated total = %v, want 24 GiB", summary.MemoryTotal)
	}
	if summary.SharedMemoryTotal == nil || *summary.SharedMemoryTotal != 64*1024*1024*1024 {
		t.Errorf("shared total = %v, want the single 64 GiB host pool", summary.SharedMemoryTotal)
	}
	if !summary.MemoryPartial {
		t.Error("missing memory data must mark the aggregate as partial")
	}
}

func TestSummarizeGPUsKeepsUnknownMemoryDistinctFromZero(t *testing.T) {
	unknown := summarizeGPUs(&accelerator.Snapshot{Devices: []accelerator.Device{{ID: "nvidia:0", Vendor: "nvidia", GPU: &gpu.Device{}}}})
	if len(unknown) != 1 || unknown[0].MemoryTotal != nil || !unknown[0].MemoryPartial {
		t.Fatalf("unknown memory must remain null and partial: %+v", unknown)
	}
	zero := summarizeGPUs(&accelerator.Snapshot{Devices: []accelerator.Device{{ID: "nvidia:0", Vendor: "nvidia", GPU: &gpu.Device{}, Metrics: gpuSummaryTestMemory(0)}}})
	if len(zero) != 1 || zero[0].MemoryTotal == nil || *zero[0].MemoryTotal != 0 || zero[0].MemoryPartial {
		t.Fatalf("a known zero value must remain distinguishable from unknown: %+v", zero)
	}
}
