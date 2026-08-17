package gpu

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/common"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/schema"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type NvidiaSMI struct{}

type SMI interface {
	LoadGpuInfo() (*common.GpuInfo, error)
}

func New() (bool, SMI) {
	var clients []SMI
	if cmd.Which("nvidia-smi") {
		clients = append(clients, NvidiaSMI{})
	}
	if cmd.Which("npu-smi") {
		clients = append(clients, AscendSMI{})
	}
	if len(clients) == 0 {
		return false, nil
	}
	if len(clients) == 1 {
		return true, clients[0]
	}
	return true, multiSMI{clients: clients}
}

type multiSMI struct {
	clients []SMI
}

type smiResult struct {
	info *common.GpuInfo
	err  error
}

func (m multiSMI) LoadGpuInfo() (*common.GpuInfo, error) {
	results := make([]smiResult, len(m.clients))
	var wg sync.WaitGroup
	for index, client := range m.clients {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[index].info, results[index].err = client.LoadGpuInfo()
		}()
	}
	wg.Wait()

	merged := &common.GpuInfo{}
	var (
		errs           []error
		types          []string
		driverVersions []string
	)
	for _, result := range results {
		if result.err != nil {
			errs = append(errs, result.err)
			continue
		}
		if result.info == nil {
			continue
		}
		deviceType := result.info.Type
		if deviceType != "" {
			types = append(types, deviceType)
		}
		if result.info.DriverVersion != "" {
			driverVersions = append(driverVersions, fmt.Sprintf("%s: %s", strings.ToUpper(deviceType), result.info.DriverVersion))
		}
		if result.info.CudaVersion != "" {
			merged.CudaVersion = result.info.CudaVersion
		}
		for _, device := range result.info.GPUs {
			if device.Type == "" {
				device.Type = deviceType
			}
			merged.GPUs = append(merged.GPUs, device)
		}
	}

	if len(types) == 1 {
		merged.Type = types[0]
	} else if len(types) > 1 {
		merged.Type = "mixed"
	}
	if len(driverVersions) == 1 {
		parts := strings.SplitN(driverVersions[0], ": ", 2)
		merged.DriverVersion = parts[len(parts)-1]
	} else {
		merged.DriverVersion = strings.Join(driverVersions, "；")
	}
	if len(merged.GPUs) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("calling GPU monitoring tools failed: %w", errors.Join(errs...))
	}
	return merged, nil
}

func (n NvidiaSMI) LoadGpuInfo() (*common.GpuInfo, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
	itemData, err := cmdMgr.RunWithStdout("nvidia-smi", "-q", "-x")
	if err != nil {
		return nil, fmt.Errorf("calling nvidia-smi failed, %v", err)
	}
	data := []byte(itemData)
	version := "v11"

	buf := bytes.NewBuffer(data)
	decoder := xml.NewDecoder(buf)
	for {
		token, err := decoder.Token()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("reading token failed: %w", err)
		}
		d, ok := token.(xml.Directive)
		if !ok {
			continue
		}
		directive := string(d)
		if !strings.HasPrefix(directive, "DOCTYPE") {
			continue
		}
		parts := strings.Split(directive, " ")
		s := strings.Trim(parts[len(parts)-1], "\" ")
		if strings.HasPrefix(s, "nvsmi_device_") && strings.HasSuffix(s, ".dtd") {
			version = strings.TrimSuffix(strings.TrimPrefix(s, "nvsmi_device_"), ".dtd")
		} else {
			global.LOG.Debugf("Cannot find schema version in %q", directive)
		}
		break
	}

	return schema.Parse(data, version)
}
