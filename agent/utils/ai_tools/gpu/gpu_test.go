package gpu

import (
	"errors"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/common"
)

type fakeSMI struct {
	info *common.GpuInfo
	err  error
}

func (f fakeSMI) LoadGpuInfo() (*common.GpuInfo, error) {
	return f.info, f.err
}

func TestMultiSMIMergesNvidiaAndAscend(t *testing.T) {
	client := multiSMI{clients: []SMI{
		fakeSMI{info: &common.GpuInfo{
			Type:          "nvidia",
			CudaVersion:   "12.8",
			DriverVersion: "570.00",
			GPUs:          []common.GPU{{Index: 0, ProductName: "NVIDIA H100"}},
		}},
		fakeSMI{info: &common.GpuInfo{
			Type:          "ascend",
			DriverVersion: "24.1.rc3",
			GPUs:          []common.GPU{{Index: 0, ProductName: "910B2C"}},
		}},
	}}

	info, err := client.LoadGpuInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "mixed" || info.CudaVersion != "12.8" {
		t.Fatalf("unexpected merged info: %#v", info)
	}
	if info.DriverVersion != "NVIDIA: 570.00；ASCEND: 24.1.rc3" {
		t.Fatalf("unexpected driver versions: %q", info.DriverVersion)
	}
	if len(info.GPUs) != 2 || info.GPUs[0].Type != "nvidia" || info.GPUs[1].Type != "ascend" {
		t.Fatalf("unexpected merged devices: %#v", info.GPUs)
	}
}

func TestMultiSMIReturnsAvailableVendorWhenAnotherFails(t *testing.T) {
	client := multiSMI{clients: []SMI{
		fakeSMI{err: errors.New("nvidia-smi failed")},
		fakeSMI{info: &common.GpuInfo{
			Type:          "ascend",
			DriverVersion: "24.1.rc3",
			GPUs:          []common.GPU{{Type: "ascend", ProductName: "910B2C"}},
		}},
	}}

	info, err := client.LoadGpuInfo()
	if err != nil {
		t.Fatal(err)
	}
	if info.Type != "ascend" || len(info.GPUs) != 1 {
		t.Fatalf("unexpected partial result: %#v", info)
	}
}
