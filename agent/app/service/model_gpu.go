package service

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/re"
	"github.com/compose-spec/compose-go/v2/interpolation"
	"github.com/compose-spec/compose-go/v2/template"
)

func LoadModelGPUDevices() dto.ModelGPUDevices {
	result := dto.ModelGPUDevices{Devices: []dto.ModelGPUDevice{}, Warnings: []string{}}
	uuidPattern := re.GetRegex(re.NvidiaGPUUUIDPattern)
	renderPattern := re.GetRegex(re.IntelGPURenderNamePattern)
	ascendPattern := re.GetRegex(re.AscendGPUDeviceNamePattern)
	for _, directory := range []string{"/proc/driver/nvidia/gpus", "/sys/class/drm", "/dev"} {
		entries, err := os.ReadDir(directory)
		if err != nil {
			if !os.IsNotExist(err) {
				result.Warnings = append(result.Warnings, err.Error())
			}
			continue
		}
		for _, entry := range entries {
			var device dto.ModelGPUDevice
			var devicePath string
			switch directory {
			case "/proc/driver/nvidia/gpus":
				informationPath := filepath.Join(directory, entry.Name(), "information")
				content, err := os.ReadFile(informationPath)
				if err != nil {
					result.Warnings = append(result.Warnings, err.Error())
					continue
				}
				fields := map[string]string{}
				for _, line := range strings.Split(string(content), "\n") {
					key, value, found := strings.Cut(line, ":")
					if found {
						fields[strings.TrimSpace(key)] = strings.TrimSpace(value)
					}
				}
				uuid := fields["GPU UUID"]
				minor, err := strconv.Atoi(fields["Device Minor"])
				if !uuidPattern.MatchString(uuid) || err != nil || minor < 0 {
					result.Warnings = append(result.Warnings, fmt.Sprintf("invalid NVIDIA GPU UUID or device minor in %s", informationPath))
					continue
				}
				name := fields["Model"]
				if name == "" {
					name = "NVIDIA"
				}
				device = dto.ModelGPUDevice{ID: "nvidia:" + uuid, Name: name + " · " + entry.Name(), Type: "nvidia", UUID: uuid}
				devicePath = filepath.Join("/dev", "nvidia"+strconv.Itoa(minor))
			case "/sys/class/drm":
				if !renderPattern.MatchString(entry.Name()) {
					continue
				}
				vendor, err := os.ReadFile(filepath.Join(directory, entry.Name(), "device/vendor"))
				if err != nil {
					result.Warnings = append(result.Warnings, err.Error())
					continue
				}
				if strings.TrimSpace(string(vendor)) != "0x8086" {
					continue
				}
				devicePath = filepath.Join("/dev/dri", entry.Name())
				device = dto.ModelGPUDevice{ID: "intel:" + devicePath, Name: "Intel · " + entry.Name(), Type: "intel"}
			case "/dev":
				if !ascendPattern.MatchString(entry.Name()) {
					continue
				}
				devicePath = filepath.Join(directory, entry.Name())
				device = dto.ModelGPUDevice{ID: "ascend:" + strings.TrimPrefix(entry.Name(), "davinci"), Name: "Ascend · " + entry.Name(), Type: "ascend"}
			}
			info, err := os.Stat(devicePath)
			if err != nil {
				result.Warnings = append(result.Warnings, err.Error())
			} else {
				device.Available = info.Mode()&os.ModeCharDevice != 0
			}
			result.Devices = append(result.Devices, device)
		}
	}
	return result
}

func ApplyModelGPU(service map[string]interface{}, config *dto.ModelGPUConfig, params map[string]interface{}, imageType string) error {
	if config == nil {
		return nil
	}
	if err := validateModelGPU(config, imageType); err != nil {
		return err
	}
	resolved, err := resolveModelGPUService(service, params, false)
	if err != nil {
		return err
	}
	devicePaths := []string{}
	ids := []string{}
	logicalIDs := []string{}
	if config.Enabled && imageType != "nvidia" {
		for _, deviceID := range config.DeviceIDs {
			id := strings.TrimPrefix(deviceID, imageType+":")
			if !slices.Contains(ids, id) {
				ids = append(ids, id)
			}
		}
		if imageType == "ascend" {
			slices.SortFunc(ids, func(a, b string) int {
				left, _ := strconv.Atoi(a)
				right, _ := strconv.Atoi(b)
				return left - right
			})
			visibleIDs := ids
			if config.Mode == "all" || resolved["privileged"] == true || resolved["privileged"] == "true" {
				visibleIDs = loadModelAscendDeviceIDs()
				if len(visibleIDs) == 0 {
					return fmt.Errorf("no available Ascend devices")
				}
			}
			if config.Mode == "all" {
				ids = visibleIDs
			} else {
				for _, id := range ids {
					index := slices.Index(visibleIDs, id)
					if index < 0 {
						return fmt.Errorf("Ascend GPU is unavailable: %s", id)
					}
					logicalIDs = append(logicalIDs, strconv.Itoa(index))
				}
			}
			for _, id := range ids {
				devicePaths = append(devicePaths, "/dev/davinci"+id)
			}
			devicePaths = append(devicePaths, "/dev/davinci_manager", "/dev/devmm_svm", "/dev/hisi_hdc")
		} else if config.Mode == "all" {
			devicePaths = append(devicePaths, "/dev/dri")
		} else {
			devicePaths = append(devicePaths, ids...)
		}
	}
	if imageType == "nvidia" {
		delete(service, "gpus")
	}
	if imageType == "nvidia" && resolved["runtime"] == "nvidia" && !config.Enabled {
		delete(service, "runtime")
	}
	deploy, _ := service["deploy"].(map[string]interface{})
	resources, _ := deploy["resources"].(map[string]interface{})
	reservations, _ := resources["reservations"].(map[string]interface{})
	devices, _ := reservations["devices"].([]interface{})
	resolvedDevices, _ := resolved["reservations"].([]interface{})
	kept := []interface{}{}
	var gpuDevice map[string]interface{}
	for index, item := range devices {
		device, _ := resolvedDevices[index].(map[string]interface{})
		capabilities, _ := device["capabilities"].([]interface{})
		driver, _ := device["driver"].(string)
		isNvidiaGPU := driver == "nvidia" || (driver == "" && slices.Contains(capabilities, interface{}("gpu")))
		if imageType != "nvidia" || !isNvidiaGPU {
			kept = append(kept, item)
		} else if gpuDevice == nil {
			gpuDevice, _ = item.(map[string]interface{})
		}
	}
	if config.Enabled && imageType == "nvidia" {
		device := gpuDevice
		if device == nil {
			device = map[string]interface{}{}
		}
		device["driver"] = "nvidia"
		device["capabilities"] = config.Capabilities
		delete(device, "count")
		delete(device, "device_ids")
		switch config.Mode {
		case "all":
			device["count"] = "all"
		case "count":
			device["count"] = config.Count
		default:
			ids := []string{}
			for _, id := range config.DeviceIDs {
				ids = append(ids, strings.TrimPrefix(id, "nvidia:"))
			}
			device["device_ids"] = ids
		}
		kept = append(kept, device)
	}

	if len(kept) > 0 {
		if deploy == nil {
			deploy = map[string]interface{}{}
			service["deploy"] = deploy
		}
		if resources == nil {
			resources = map[string]interface{}{}
			deploy["resources"] = resources
		}
		if reservations == nil {
			reservations = map[string]interface{}{}
			resources["reservations"] = reservations
		}
		reservations["devices"] = kept
	} else {
		delete(reservations, "devices")
		if len(reservations) == 0 {
			delete(resources, "reservations")
		}
		if len(resources) == 0 {
			delete(deploy, "resources")
		}
		if len(deploy) == 0 {
			delete(service, "deploy")
		}
	}
	mappedPaths := []string{}
	for _, key := range []string{"devices", "volumes"} {
		mappings, _ := service[key].([]interface{})
		resolvedMappings, _ := resolved[key].([]interface{})
		kept := []interface{}{}
		for index, item := range mappings {
			source := modelGPUDeviceSource(resolvedMappings[index])
			managed := false
			switch imageType {
			case "nvidia":
				managed = strings.HasPrefix(source, "/dev/nvidia") || strings.HasPrefix(source, "nvidia.com/gpu=")
			case "intel":
				managed = source == "/dev/dri" || strings.HasPrefix(source, "/dev/dri/")
			case "ascend":
				managed = strings.HasPrefix(source, "/dev/davinci") || source == "/dev/devmm_svm" || source == "/dev/hisi_hdc"
			}
			if managed {
				selected := slices.Contains(devicePaths, source)
				if config.Enabled && config.Mode == "all" && imageType == "intel" {
					selected = true
				}
				if selected {
					kept = append(kept, item)
					mappedPaths = append(mappedPaths, source)
				}
				continue
			}
			kept = append(kept, item)
		}
		if len(kept) == 0 {
			delete(service, key)
		} else {
			service[key] = kept
		}
	}
	if config.Enabled && imageType != "nvidia" {
		devices, _ := service["devices"].([]interface{})
		for _, path := range devicePaths {
			if !slices.Contains(mappedPaths, path) {
				devices = append(devices, path+":"+path)
			}
		}
		if len(devices) > 0 {
			service["devices"] = devices
		}
	}
	gpuEnvironment := map[string]string{}
	if config.Enabled && imageType == "ascend" && config.Mode != "all" {
		gpuEnvironment["ASCEND_RT_VISIBLE_DEVICES"] = strings.Join(logicalIDs, ",")
	}
	var environmentKeys []string
	switch imageType {
	case "nvidia":
		environmentKeys = []string{"NVIDIA_VISIBLE_DEVICES", "NVIDIA_DRIVER_CAPABILITIES"}
	case "ascend":
		environmentKeys = []string{"ASCEND_RT_VISIBLE_DEVICES"}
		if params != nil {
			params["ASCEND_RT_VISIBLE_DEVICES"] = strings.Join(logicalIDs, ",")
		}
	default:
		return nil
	}
	switch environment := service["environment"].(type) {
	case map[string]interface{}:
		for _, key := range environmentKeys {
			delete(environment, key)
		}
		for key, value := range gpuEnvironment {
			environment[key] = value
		}
	case []interface{}:
		kept := []interface{}{}
		for _, item := range environment {
			value, ok := item.(string)
			if !ok {
				return fmt.Errorf("invalid Compose environment")
			}
			key, _, _ := strings.Cut(value, "=")
			if !slices.Contains(environmentKeys, key) {
				kept = append(kept, item)
			}
		}
		for key, value := range gpuEnvironment {
			kept = append(kept, key+"="+value)
		}
		service["environment"] = kept
	case nil:
		if len(gpuEnvironment) > 0 {
			environment := map[string]interface{}{}
			for key, value := range gpuEnvironment {
				environment[key] = value
			}
			service["environment"] = environment
		}
	default:
		return fmt.Errorf("invalid Compose environment")
	}
	return nil
}

func ReadModelGPU(service map[string]interface{}, params map[string]interface{}, imageType string) *dto.ModelGPUConfig {
	resolved, err := resolveModelGPUService(service, params, true)
	if err != nil {
		return &dto.ModelGPUConfig{Mode: "devices", DeviceIDs: []string{}, Capabilities: []string{"gpu"}, Error: err.Error()}
	}
	if imageType == "nvidia" {
		config, err := readModelGPU(service, resolved)
		if err != nil {
			return &dto.ModelGPUConfig{Mode: "devices", DeviceIDs: []string{}, Capabilities: []string{"gpu"}, Error: err.Error()}
		}
		return config
	}
	config := &dto.ModelGPUConfig{Mode: "devices", DeviceIDs: []string{}, Capabilities: []string{"gpu"}}
	privileged := resolved["privileged"] == true || resolved["privileged"] == "true"
	all := false
	for _, key := range []string{"devices", "volumes"} {
		mappings, _ := resolved[key].([]interface{})
		for _, item := range mappings {
			source := modelGPUDeviceSource(item)
			id := ""
			if imageType == "intel" {
				switch {
				case source == "/dev/dri":
					all = true
				case re.GetRegex(re.IntelGPUDevicePathPattern).MatchString(source):
					id = imageType + ":" + source
				case strings.HasPrefix(source, "/dev/dri/"):
					config.Error = "unsupported Intel GPU mapping requires an explicit selection: " + source
					return config
				}
			}
			if imageType == "ascend" && re.GetRegex(re.AscendGPUDevicePathPattern).MatchString(source) {
				id = imageType + ":" + strings.TrimPrefix(source, "/dev/davinci")
			}
			if id != "" && !slices.Contains(config.DeviceIDs, id) {
				config.DeviceIDs = append(config.DeviceIDs, id)
			}
		}
	}
	config.Enabled = all || len(config.DeviceIDs) > 0
	config.Mode = "devices"
	if all {
		config.Mode = "all"
		config.DeviceIDs = []string{}
	}
	if imageType == "ascend" {
		const key = "ASCEND_RT_VISIBLE_DEVICES"
		visible := ""
		switch environment := service["environment"].(type) {
		case map[string]interface{}:
			if value, exists := environment[key]; exists {
				if value == nil {
					visible = "${" + key + "}"
				} else {
					visible = fmt.Sprint(value)
				}
			}
		case []interface{}:
			for _, item := range environment {
				value, _ := item.(string)
				if value == key {
					visible = "${" + key + "}"
				} else if strings.HasPrefix(value, key+"=") {
					visible = strings.TrimPrefix(value, key+"=")
				}
			}
		}
		visible, err := resolveModelGPUValue(visible, params)
		if err != nil {
			config.Error = err.Error()
			return config
		}
		visible = strings.TrimSpace(visible)
		if visible == "all" || (visible == "" && config.Enabled) {
			config.Enabled = true
			config.Mode = "all"
			config.DeviceIDs = []string{}
		} else if visible != "" {
			physicalIDs := []string{}
			for _, id := range config.DeviceIDs {
				physicalIDs = append(physicalIDs, strings.TrimPrefix(id, "ascend:"))
			}
			slices.SortFunc(physicalIDs, func(a, b string) int {
				left, _ := strconv.Atoi(a)
				right, _ := strconv.Atoi(b)
				return left - right
			})
			if privileged {
				physicalIDs = loadModelAscendDeviceIDs()
			}
			selected := []string{}
			for _, id := range strings.Split(visible, ",") {
				id = strings.TrimSpace(id)
				if !re.GetRegex(re.AscendGPUDeviceIDPattern).MatchString(id) {
					config.Error = "invalid Ascend GPU device: " + id
					return config
				}
				index, err := strconv.Atoi(id)
				if err != nil || index < 0 || index >= len(physicalIDs) {
					config.Error = "Ascend logical GPU is not available in the container: " + id
					return config
				}
				selected = append(selected, "ascend:"+physicalIDs[index])
			}
			config.Enabled = true
			config.Mode = "devices"
			config.DeviceIDs = selected
		}
	}
	return config
}

func readModelGPU(service, resolved map[string]interface{}) (*dto.ModelGPUConfig, error) {
	config := &dto.ModelGPUConfig{Mode: "all", DeviceIDs: []string{}, Capabilities: []string{"gpu"}}
	devices, _ := resolved["reservations"].([]interface{})
	if service["gpus"] != nil {
		return nil, fmt.Errorf("custom GPU declarations require an explicit selection")
	}
	for _, item := range devices {
		device, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid GPU reservation")
		}
		driver, _ := device["driver"].(string)
		capabilities, _ := device["capabilities"].([]interface{})
		isGPU := driver == "nvidia" || (driver == "" && slices.Contains(capabilities, interface{}("gpu")))
		if !isGPU {
			continue
		}
		if config.Enabled {
			return nil, fmt.Errorf("multiple GPU declarations require an explicit selection")
		}
		config.Enabled = true
		ids, _ := device["device_ids"].([]interface{})
		if len(ids) > 0 {
			if device["count"] != nil {
				return nil, fmt.Errorf("GPU count and device_ids cannot be combined")
			}
			config.Mode = "devices"
			for _, id := range ids {
				config.DeviceIDs = append(config.DeviceIDs, "nvidia:"+fmt.Sprint(id))
			}
		} else if count := device["count"]; count != nil && fmt.Sprint(count) != "all" && fmt.Sprint(count) != "-1" {
			config.Mode = "count"
			countValue, err := strconv.Atoi(fmt.Sprint(count))
			config.Count = countValue
			if err != nil {
				return nil, fmt.Errorf("invalid GPU count: %v", count)
			}
		}
		for _, capability := range capabilities {
			value, ok := capability.(string)
			if !ok {
				return nil, fmt.Errorf("unsupported GPU capabilities: %v", capabilities)
			}
			if !slices.Contains(config.Capabilities, value) {
				config.Capabilities = append(config.Capabilities, value)
			}
		}
	}
	for _, key := range []string{"devices", "volumes"} {
		mappings, _ := resolved[key].([]interface{})
		for _, item := range mappings {
			source := modelGPUDeviceSource(item)
			if strings.HasPrefix(source, "/dev/nvidia") || strings.HasPrefix(source, "nvidia.com/gpu=") {
				return nil, fmt.Errorf("direct NVIDIA device mappings require an explicit selection")
			}
		}
	}
	if !config.Enabled && resolved["runtime"] == "nvidia" {
		return nil, fmt.Errorf("custom NVIDIA runtime configuration requires an explicit selection")
	}
	if err := validateModelGPU(config, "nvidia"); err != nil {
		return nil, err
	}
	return config, nil
}

func validateModelGPU(config *dto.ModelGPUConfig, imageType string) error {
	if config.Error != "" {
		return fmt.Errorf("GPU configuration must be selected explicitly: %s", config.Error)
	}
	if !config.Enabled {
		return nil
	}
	patterns := map[string]string{
		"nvidia": re.NvidiaGPUDeviceIDPattern,
		"intel":  re.IntelGPUDevicePathPattern,
		"ascend": re.AscendGPUDeviceIDPattern,
	}
	pattern, supported := patterns[imageType]
	if !supported {
		return fmt.Errorf("unsupported GPU type: %s", imageType)
	}
	switch config.Mode {
	case "all":
		if config.Count != 0 || len(config.DeviceIDs) != 0 {
			return fmt.Errorf("all GPUs cannot be combined with count or device IDs")
		}
	case "count":
		if imageType != "nvidia" || config.Count < 1 || len(config.DeviceIDs) != 0 {
			return fmt.Errorf("GPU count requires a positive number and runtime device allocation")
		}
	case "devices", "":
		if config.Count != 0 || len(config.DeviceIDs) == 0 {
			return fmt.Errorf("select at least one GPU device")
		}
		for _, deviceID := range config.DeviceIDs {
			id := strings.TrimPrefix(deviceID, imageType+":")
			if !re.GetRegex(pattern).MatchString(id) {
				return fmt.Errorf("invalid %s GPU device: %s", imageType, deviceID)
			}
		}
	default:
		return fmt.Errorf("unsupported GPU selection mode: %s", config.Mode)
	}

	if imageType != "nvidia" {
		return nil
	}
	if !slices.Contains(config.Capabilities, "gpu") {
		return fmt.Errorf("NVIDIA capabilities must include gpu")
	}
	for _, capability := range config.Capabilities {
		if capability != "gpu" && capability != "utility" && capability != "compute" {
			return fmt.Errorf("unsupported NVIDIA capability: %s", capability)
		}
	}
	return nil
}

func resolveModelGPUService(service map[string]interface{}, params map[string]interface{}, resolveSelection bool) (map[string]interface{}, error) {
	fields := map[string]interface{}{}
	fields["runtime"] = service["runtime"]
	deploy, _ := service["deploy"].(map[string]interface{})
	resources, _ := deploy["resources"].(map[string]interface{})
	reservations, _ := resources["reservations"].(map[string]interface{})
	fields["reservations"] = reservations["devices"]
	if devices, ok := reservations["devices"].([]interface{}); ok && !resolveSelection {
		declarations := make([]interface{}, 0, len(devices))
		for _, item := range devices {
			device, _ := item.(map[string]interface{})
			declarations = append(declarations, map[string]interface{}{"driver": device["driver"], "capabilities": device["capabilities"]})
		}
		fields["reservations"] = declarations
	}
	resolved, err := interpolation.Interpolate(fields, interpolation.Options{
		Substitute: func(value string, _ template.Mapping) (string, error) { return resolveModelGPUValue(value, params) },
	})
	if err != nil {
		return nil, err
	}
	for _, key := range []string{"devices", "volumes"} {
		mappings, _ := service[key].([]interface{})
		sources := make([]interface{}, 0, len(mappings))
		for _, item := range mappings {
			source, err := template.SubstituteWithOptions(modelGPUDeviceSource(item), func(name string) (string, bool) {
				value, ok := params[name]
				if !ok || value == nil {
					return "", false
				}
				return fmt.Sprint(value), true
			}, template.WithoutLogging)
			if err != nil {
				return nil, err
			}
			if _, short := item.(string); short {
				source = modelGPUDeviceSource(source)
			}
			sources = append(sources, map[string]interface{}{"source": source})
		}
		resolved[key] = sources
	}
	resolved["privileged"] = service["privileged"]
	if value, ok := service["privileged"].(string); ok {
		if value, err := resolveModelGPUValue(value, params); err == nil {
			resolved["privileged"] = value
		}
	}
	return resolved, nil
}

func modelGPUDeviceSource(item interface{}) string {
	switch mapping := item.(type) {
	case string:
		if strings.HasPrefix(mapping, "nvidia.com/gpu=") {
			return mapping
		}
		depth := 0
		for index, ch := range mapping {
			switch ch {
			case '{':
				depth++
			case '}':
				depth--
			case ':':
				if depth == 0 {
					return mapping[:index]
				}
			}
		}
		return mapping
	case map[string]interface{}:
		source, _ := mapping["source"].(string)
		return source
	}
	return ""
}

func resolveModelGPUValue(value string, params map[string]interface{}) (string, error) {
	return template.SubstituteWithOptions(value, func(name string) (string, bool) {
		value, ok := params[name]
		if !ok || value == nil {
			return "", false
		}
		return fmt.Sprint(value), true
	}, template.WithoutLogging, template.WithReplacementFunction(func(value string, mapping template.Mapping, config *template.Config) (string, error) {
		resolved, applied, err := template.DefaultReplacementAppliedFunc(value, mapping, config)
		if err != nil {
			return "", err
		}
		if !applied {
			return "", fmt.Errorf("unresolved GPU variable: %s", value)
		}
		return resolved, nil
	}))
}

func loadModelAscendDeviceIDs() []string {
	ids := []string{}
	for _, device := range LoadModelGPUDevices().Devices {
		if device.Type == "ascend" && device.Available {
			ids = append(ids, strings.TrimPrefix(device.ID, "ascend:"))
		}
	}
	slices.SortFunc(ids, func(a, b string) int {
		left, _ := strconv.Atoi(a)
		right, _ := strconv.Atoi(b)
		return left - right
	})
	return ids
}
