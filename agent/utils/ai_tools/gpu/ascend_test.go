package gpu

import "testing"

func TestParseAscendSMI910B(t *testing.T) {
	output := `+------------------------------------------------------------------------------------------------+
| npu-smi 24.1.rc3                 Version: 24.1.rc3                                             |
+---------------------------+---------------+----------------------------------------------------+
| NPU   Name                | Health        | Power(W)    Temp(C)           Hugepages-Usage(page)|
| Chip                      | Bus-Id        | AICore(%)   Memory-Usage(MB)  HBM-Usage(MB)        |
+===========================+===============+====================================================+
| 0     910B2C              | OK            | 92.1        45                0    / 0             |
| 0                         | 0000:5A:00.0  | 37          0    / 0          3383 / 65536         |
+===========================+===============+====================================================+
| 16    910 B4              | Warning       | 72.1        42                0    / 0             |
| 0                         | 0000:15:00.0  | 8           0    / 0          2900 / 32768         |
+===========================+===============+====================================================+
+---------------------------+---------------+----------------------------------------------------+
| NPU     Chip              | Process id    | Process name             | Process memory(MB)      |
+===========================+===============+====================================================+
| 16      0                 | 207848        | python                   | 116                     |
+===========================+===============+====================================================+`

	info := parseAscendSMI(output)
	if info.Type != "ascend" || info.DriverVersion != "24.1.rc3" {
		t.Fatalf("unexpected device info: %#v", info)
	}
	if len(info.GPUs) != 2 {
		t.Fatalf("expected 2 NPUs, got %d", len(info.GPUs))
	}
	first := info.GPUs[0]
	if first.Type != "ascend" || first.Index != 0 || first.ProductName != "910B2C" || first.BusID != "0000:5A:00.0" {
		t.Errorf("unexpected first NPU identity: %#v", first)
	}
	if first.GPUUtil != "37 %" || first.Temperature != "45 C" || first.PowerDraw != "92.1 W" {
		t.Errorf("unexpected first NPU metrics: %#v", first)
	}
	if first.MemUsed != "3383 MB" || first.MemTotal != "65536 MB" {
		t.Errorf("expected HBM usage, got %q / %q", first.MemUsed, first.MemTotal)
	}

	second := info.GPUs[1]
	if second.Index != 16 || second.ProductName != "910 B4" || second.PerformanceState != "Warning" {
		t.Errorf("unexpected second NPU: %#v", second)
	}
	if len(second.Processes) != 1 {
		t.Fatalf("expected one process, got %d", len(second.Processes))
	}
	process := second.Processes[0]
	if process.Pid != "207848" || process.Type != "NPU" || process.ProcessName != "python" || process.UsedMemory != "116 MB" {
		t.Errorf("unexpected process: %#v", process)
	}
}

func TestAscendMemoryUsageFallsBackToGenericMemory(t *testing.T) {
	used, total := ascendMemoryUsage("12 / 32768")
	if used != "12 MB" || total != "32768 MB" {
		t.Fatalf("unexpected memory usage: %q / %q", used, total)
	}

	used, total = ascendMemoryUsage("N/A")
	if used != "N/A" || total != "N/A" {
		t.Fatalf("unexpected unavailable memory usage: %q / %q", used, total)
	}
}
