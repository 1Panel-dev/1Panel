package gpu

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

type provider interface {
	LoadInfo(context.Context) (*Info, error)
}

type Client struct {
	providers []provider
}

type providerResult struct {
	info *Info
	err  error
}

func New() (bool, Client) {
	client := Client{}
	if cmd.Which(nvidiaSMICommand) {
		client.providers = append(client.providers, nvidiaSMI{})
	}
	if command, ok := findAMDSMI(); ok {
		client.providers = append(client.providers, amdSMI{command: command})
	}
	return len(client.providers) > 0, client
}

func (c Client) LoadInfo() (*Info, error) {
	return c.LoadInfoContext(context.Background())
}

func (c Client) LoadInfoContext(ctx context.Context) (*Info, error) {
	results := make([]providerResult, len(c.providers))
	var wg sync.WaitGroup
	for index, item := range c.providers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[index].info, results[index].err = item.LoadInfo(ctx)
		}()
	}
	wg.Wait()

	merged := &Info{}
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
		if result.info.Type != "" {
			types = append(types, result.info.Type)
		}
		if result.info.DriverVersion != "" {
			driverVersions = append(driverVersions, formatDriverVersion(result.info.Type, result.info.DriverVersion))
		}
		if result.info.CudaVersion != "" {
			merged.CudaVersion = result.info.CudaVersion
		}
		for _, device := range result.info.Devices {
			if device.Type == "" {
				device.Type = result.info.Type
			}
			merged.Devices = append(merged.Devices, device)
		}
	}

	merged.Type = mergeTypes(types)
	merged.DriverVersion = mergeDriverVersions(driverVersions)
	if len(merged.Devices) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("calling GPU monitoring tools failed: %w", errors.Join(errs...))
	}
	return merged, nil
}

func formatDriverVersion(deviceType, version string) string {
	if deviceType == "" {
		return version
	}
	return fmt.Sprintf("%s: %s", strings.ToUpper(deviceType), version)
}

func mergeTypes(types []string) string {
	if len(types) == 1 {
		return types[0]
	}
	if len(types) > 1 {
		return "mixed"
	}
	return ""
}

func mergeDriverVersions(versions []string) string {
	if len(versions) == 1 {
		parts := strings.SplitN(versions[0], ": ", 2)
		return parts[len(parts)-1]
	}
	return strings.Join(versions, "；")
}
