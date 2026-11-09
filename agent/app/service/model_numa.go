package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/compose-spec/compose-go/v2/types"
)

const modelNUMAKey = "x-1panel-numa"

type modelNUMAState struct {
	Config    dto.ModelNUMAConfig    `json:"config"`
	Original  map[string]interface{} `json:"original"`
	Generated map[string]interface{} `json:"generated"`
}

func loadModelNUMANodes(sysRoot string) ([]dto.ModelNUMANode, error) {
	nodes := []dto.ModelNUMANode{}
	content, err := os.ReadFile(filepath.Join(sysRoot, "devices/system/node/online"))
	if err != nil {
		return nodes, fmt.Errorf("NUMA topology unavailable: %w", err)
	}
	online, err := parseModelCPUList(strings.TrimSpace(string(content)))
	if err != nil {
		return nodes, err
	}
	cpuContent, err := os.ReadFile(filepath.Join(sysRoot, "devices/system/cpu/online"))
	if err != nil {
		return nodes, err
	}
	onlineCPUs, err := parseModelCPUList(strings.TrimSpace(string(cpuContent)))
	if err != nil {
		return nodes, err
	}
	for _, id := range online {
		directory := filepath.Join(sysRoot, "devices/system/node", fmt.Sprintf("node%d", id))
		cpus, err := os.ReadFile(filepath.Join(directory, "cpulist"))
		if err != nil {
			return nodes, err
		}
		meminfo, err := os.ReadFile(filepath.Join(directory, "meminfo"))
		if err != nil {
			return nodes, err
		}
		node := dto.ModelNUMANode{ID: id, CPUs: strings.TrimSpace(string(cpus))}
		nodeCPUs, err := parseModelCPUList(node.CPUs)
		if err != nil {
			return nodes, err
		}
		activeCPUs := []string{}
		for _, cpu := range nodeCPUs {
			if slices.Contains(onlineCPUs, cpu) {
				activeCPUs = append(activeCPUs, strconv.Itoa(cpu))
			}
		}
		if len(activeCPUs) != len(nodeCPUs) {
			node.CPUs = strings.Join(activeCPUs, ",")
		}
		for _, line := range strings.Split(string(meminfo), "\n") {
			_, value, found := strings.Cut(line, "MemTotal:")
			if found {
				fields := strings.Fields(value)
				if len(fields) > 0 {
					total, _ := strconv.ParseUint(fields[0], 10, 64)
					node.Memory = total > 0
				}
			}
		}
		nodes = append(nodes, node)
	}
	return nodes, nil
}

func loadModelGPUNode(sysRoot, busID string) *int {
	var domain, bus, device, function uint
	if n, err := fmt.Sscanf(strings.ToLower(busID), "%x:%x:%x.%x", &domain, &bus, &device, &function); err != nil || n != 4 || domain > 0xffff || bus > 0xff || device > 0x1f || function > 7 {
		return nil
	}
	address := fmt.Sprintf("%04x:%02x:%02x.%x", domain, bus, device, function)
	content, err := os.ReadFile(filepath.Join(sysRoot, "bus/pci/devices", address, "numa_node"))
	if err != nil {
		return nil
	}
	node, err := strconv.Atoi(strings.TrimSpace(string(content)))
	if err != nil || node < 0 {
		return nil
	}
	return &node
}

func parseModelCPUList(value string) ([]int, error) {
	values := []int{}
	if value == "" {
		return values, nil
	}
	for _, item := range strings.Split(value, ",") {
		startText, endText, ranged := strings.Cut(item, "-")
		start, err := strconv.Atoi(startText)
		if err != nil || start < 0 || start > 1048576 {
			return nil, fmt.Errorf("invalid CPU/node list: %s", value)
		}
		end := start
		if ranged {
			end, err = strconv.Atoi(endText)
		}
		if err != nil || end < start || end > 1048576 {
			return nil, fmt.Errorf("invalid CPU/node list: %s", value)
		}
		for id := start; id <= end; id++ {
			values = append(values, id)
		}
	}
	slices.Sort(values)
	return slices.Compact(values), nil
}

func resolveModelNUMANodes(config *dto.ModelGPUConfig, topology dto.ModelGPUDevices, imageType, cpuset string) ([]int, error) {
	nodes := slices.Clone(config.NUMA.Nodes)
	switch config.NUMA.Mode {
	case "auto":
		if config.Mode == "count" {
			return nil, fmt.Errorf("automatic NUMA binding requires specific GPU devices or all GPUs")
		}
		nodes = nil
		selected := 0
		for _, device := range topology.Devices {
			if device.Type != imageType || (config.Mode != "all" && !slices.Contains(config.DeviceIDs, device.ID)) {
				continue
			}
			selected++
			if !device.Available || device.NUMANode == nil {
				return nil, fmt.Errorf("NUMA topology is unknown for GPU %s; select NUMA nodes manually", device.Name)
			}
			nodes = append(nodes, *device.NUMANode)
		}
		if selected == 0 || (config.Mode != "all" && selected != len(config.DeviceIDs)) {
			return nil, fmt.Errorf("selected GPUs are unavailable for automatic NUMA binding")
		}
		slices.Sort(nodes)
		nodes = slices.Compact(nodes)
		if len(nodes) != 1 {
			return nil, fmt.Errorf("selected GPUs span NUMA nodes; select nodes manually")
		}
	case "manual":
	default:
		return nil, fmt.Errorf("unsupported NUMA binding mode: %s", config.NUMA.Mode)
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("select at least one NUMA node")
	}
	allowed, err := parseModelCPUList(cpuset)
	if err != nil {
		return nil, err
	}
	for _, id := range nodes {
		index := slices.IndexFunc(topology.NUMANodes, func(node dto.ModelNUMANode) bool { return node.ID == id })
		if index < 0 {
			return nil, fmt.Errorf("NUMA node %d is unavailable", id)
		}
		node := topology.NUMANodes[index]
		cpus, err := parseModelCPUList(node.CPUs)
		if err != nil {
			return nil, err
		}
		if len(cpus) == 0 || !node.Memory {
			return nil, fmt.Errorf("NUMA node %d has no CPUs or memory", id)
		}
		if cpuset != "" && !slices.ContainsFunc(cpus, func(cpu int) bool { return slices.Contains(allowed, cpu) }) {
			return nil, fmt.Errorf("NUMA node %d conflicts with the container CPU set", id)
		}
	}
	slices.Sort(nodes)
	return slices.Compact(nodes), nil
}

func readModelNUMAState(service map[string]interface{}) (*modelNUMAState, error) {
	value, exists := service[modelNUMAKey]
	if !exists {
		return nil, nil
	}
	content, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("invalid managed NUMA configuration")
	}
	var state modelNUMAState
	if err := json.Unmarshal([]byte(content), &state); err != nil {
		return nil, fmt.Errorf("invalid managed NUMA configuration: %w", err)
	}
	if state.Original == nil || state.Generated["entrypoint"] == nil {
		return nil, fmt.Errorf("invalid managed NUMA startup configuration")
	}
	return &state, nil
}

func restoreModelNUMA(service map[string]interface{}) error {
	state, err := readModelNUMAState(service)
	if err != nil || state == nil {
		return err
	}
	for key, generated := range state.Generated {
		currentJSON, _ := json.Marshal(service[key])
		generatedJSON, _ := json.Marshal(generated)
		if string(currentJSON) != string(generatedJSON) {
			if key == "entrypoint" {
				return fmt.Errorf("managed NUMA entrypoint was edited; restore it before changing NUMA binding")
			}
			if key == "cap_add" {
				var capabilities []string
				if err := json.Unmarshal(currentJSON, &capabilities); err != nil {
					return fmt.Errorf("invalid Compose cap_add: %w", err)
				}
				capabilities = slices.DeleteFunc(capabilities, func(value string) bool { return value == "SYS_NICE" })
				if len(capabilities) == 0 {
					delete(service, key)
				} else {
					service[key] = capabilities
				}
			}
			continue
		}
		if original, exists := state.Original[key]; exists {
			service[key] = original
		} else {
			delete(service, key)
		}
	}
	delete(service, modelNUMAKey)
	return nil
}

func applyModelNUMA(service map[string]interface{}, config *dto.ModelGPUConfig, params map[string]interface{}, imageType string) error {
	if err := restoreModelNUMA(service); err != nil {
		return err
	}
	if !config.Enabled || config.NUMA == nil || config.NUMA.Mode == "" || config.NUMA.Mode == "off" {
		return nil
	}
	var topology dto.ModelGPUDevices
	if config.NUMA.Mode == "auto" {
		topology = LoadModelGPUDevices()
	} else {
		nodes, err := loadModelNUMANodes("/sys")
		if err != nil {
			return err
		}
		topology.NUMANodes = nodes
	}
	cpuset := ""
	if value, exists := service["cpuset"]; exists {
		var err error
		cpuset, err = resolveModelGPUValue(fmt.Sprint(value), params)
		if err != nil {
			return err
		}
	}
	nodes, err := resolveModelNUMANodes(config, topology, imageType, cpuset)
	if err != nil {
		return err
	}
	var imageEntrypoint, imageCommand []string
	if service["entrypoint"] == nil {
		imageName, err := resolveModelGPUValue(fmt.Sprint(service["image"]), params)
		if err != nil {
			return err
		}
		cli, err := docker.NewDockerClient()
		if err != nil {
			return err
		}
		defer cli.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		info, err := cli.ImageInspect(ctx, imageName)
		if err != nil {
			return fmt.Errorf("cannot preserve image startup command for NUMA binding; pull image %s first: %w", imageName, err)
		}
		if info.Config != nil {
			imageEntrypoint = info.Config.Entrypoint
			imageCommand = info.Config.Cmd
		}
	}
	return wrapModelNUMA(service, *config.NUMA, nodes, params, imageEntrypoint, imageCommand)
}

func wrapModelNUMA(service map[string]interface{}, config dto.ModelNUMAConfig, nodes []int, params map[string]interface{}, imageEntrypoint, imageCommand []string) error {
	original := map[string]interface{}{}
	for _, key := range []string{"entrypoint", "command", "cap_add"} {
		if value, exists := service[key]; exists {
			original[key] = value
		}
	}
	entrypoint := imageEntrypoint
	if value := service["entrypoint"]; value != nil {
		var err error
		entrypoint, err = modelNUMACommand(value, params)
		if err != nil {
			return err
		}
	}
	ids := make([]string, len(nodes))
	for i, node := range nodes {
		ids[i] = strconv.Itoa(node)
	}
	binding := strings.Join(ids, ",")
	wrapped := []string{"numactl", "--cpunodebind=" + binding, "--membind=" + binding, "--"}
	for _, arg := range entrypoint {
		wrapped = append(wrapped, strings.ReplaceAll(arg, "$", "$$"))
	}
	generated := map[string]interface{}{"entrypoint": wrapped}
	capabilityJSON, err := json.Marshal(service["cap_add"])
	if err != nil {
		return err
	}
	var capabilities []string
	if err := json.Unmarshal(capabilityJSON, &capabilities); err != nil {
		return fmt.Errorf("invalid Compose cap_add: %w", err)
	}
	hasMemoryBindingCapability := false
	for _, capability := range capabilities {
		resolved, err := resolveModelGPUValue(capability, params)
		if err != nil {
			return err
		}
		resolved = strings.TrimPrefix(strings.ToUpper(resolved), "CAP_")
		if resolved == "SYS_NICE" || resolved == "ALL" {
			hasMemoryBindingCapability = true
		}
	}
	if !hasMemoryBindingCapability {
		generated["cap_add"] = append(capabilities, "SYS_NICE")
	}
	if service["entrypoint"] == nil && service["command"] == nil {
		args := make([]string, len(imageCommand))
		for i, arg := range imageCommand {
			args[i] = strings.ReplaceAll(arg, "$", "$$")
		}
		generated["command"] = args
	}
	effectiveCommand := service["command"]
	if value, exists := generated["command"]; exists {
		effectiveCommand = value
	}
	if len(entrypoint) == 0 {
		command, err := modelNUMACommand(effectiveCommand, params)
		if err != nil {
			return err
		}
		if len(command) == 0 {
			return fmt.Errorf("NUMA binding requires an executable startup command")
		}
	}
	config.Nodes = nodes
	state, err := json.Marshal(modelNUMAState{Config: config, Original: original, Generated: generated})
	if err != nil {
		return err
	}
	for key, value := range generated {
		service[key] = value
	}
	service[modelNUMAKey] = string(state)
	return nil
}

func modelNUMACommand(value interface{}, params map[string]interface{}) ([]string, error) {
	if value == nil {
		return nil, nil
	}
	if text, ok := value.(string); ok {
		resolved, err := resolveModelGPUValue(text, params)
		if err != nil {
			return nil, err
		}
		var command types.ShellCommand
		if err := command.DecodeMapstructure(resolved); err != nil {
			return nil, err
		}
		return command, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var args []string
	if err := json.Unmarshal(raw, &args); err != nil {
		return nil, fmt.Errorf("invalid NUMA startup command: %w", err)
	}
	for i, arg := range args {
		args[i], err = resolveModelGPUValue(arg, params)
		if err != nil {
			return nil, err
		}
	}
	return args, nil
}
