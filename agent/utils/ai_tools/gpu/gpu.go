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

	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/common"
	"github.com/1Panel-dev/1Panel/agent/utils/ai_tools/gpu/schema"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type NvidiaSMI struct{}

func New() (bool, NvidiaSMI) {
	return cmd.Which("nvidia-smi"), NvidiaSMI{}
}

func (n NvidiaSMI) LoadGpuInfo() (*common.GpuInfo, error) {
	cmdMgr := cmd.NewCommandMgr(cmd.WithTimeout(5 * time.Second))
	itemData, err := cmdMgr.RunWithStdoutBashC("nvidia-smi -q -x")
	if err != nil {
		return nil, fmt.Errorf("calling nvidia-smi failed, err: %w", err)
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

	if version == "v12" || version == "v11" {
		return schema.Parse(data, version)
	} else {
		global.LOG.Errorf("don't support such schema version %s", version)
	}

	return &common.GpuInfo{}, nil
}
