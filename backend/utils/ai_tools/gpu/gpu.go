package gpu

import (
	"bytes"
	_ "embed"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/ai_tools/gpu/common"
	"github.com/1Panel-dev/1Panel/backend/utils/ai_tools/gpu/schema_v12"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
)

type NvidiaSMI struct{}

func New() (bool, NvidiaSMI) {
	return cmd.Which("nvidia-smi"), NvidiaSMI{}
}

func (n NvidiaSMI) LoadGpuInfo() (*common.GpuInfo, error) {
	std, err := cmd.ExecWithTimeOut("nvidia-smi -q -x", 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("calling nvidia-smi failed, err: %v", std)
	}
	data := []byte(std)
	schema := "v11"

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
			schema = strings.TrimSuffix(strings.TrimPrefix(s, "nvsmi_device_"), ".dtd")
		} else {
			global.LOG.Debugf("Cannot find schema version in %q", directive)
		}
		break
	}

	if schema != "v12" {
		return &common.GpuInfo{}, nil
	}
	return schema_v12.Parse(data)
}
