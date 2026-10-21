package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/mattn/go-shellwords"
)

func AIBenchmarkSources(kind, selected string) ([]dto.AIBenchmarkSource, error) {
	if kind != "vllm" && kind != "ollama" {
		return nil, errors.New("unsupported benchmark source")
	}
	selectedID := strings.SplitN(selected, ":", 2)[0]
	type instance struct {
		id              uint
		name, container string
		port            int
	}
	instances := []instance{}
	apps, err := appRepo.GetBy(appRepo.WithKey(kind))
	if err != nil {
		return nil, err
	}
	for _, app := range apps {
		installs, err := appInstallRepo.ListBy(context.Background(), appInstallRepo.WithAppId(app.ID))
		if err != nil {
			return nil, err
		}
		for _, install := range installs {
			if selected != "" && strconv.FormatUint(uint64(install.ID), 10) != selectedID {
				continue
			}
			instances = append(instances, instance{id: install.ID, name: install.Name, container: install.ContainerName, port: install.HttpPort})
		}
	}
	results := []dto.AIBenchmarkSource{}
	if len(instances) == 0 {
		return results, nil
	}
	client, err := docker.NewDockerClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	if !strings.HasPrefix(client.DaemonHost(), "unix://") {
		return nil, errors.New("benchmark shortcuts require a Docker daemon on the main node; enter the remote service address manually")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	for _, instance := range instances {
		base := dto.AIBenchmarkSource{Kind: kind, Name: instance.name, Endpoint: "/v1/chat/completions", Models: []string{}, Status: "unknown"}
		inspected, err := client.ContainerInspect(ctx, instance.container)
		if err != nil {
			base.ID = strconv.FormatUint(uint64(instance.id), 10)
			base.Message = "container inspection unavailable"
			results = append(results, base)
			continue
		}
		if inspected.State != nil {
			base.Status = inspected.State.Status
		}
		args := []string{}
		if inspected.Config != nil {
			args = append(args, inspected.Config.Entrypoint...)
			args = append(args, inspected.Config.Cmd...)
		}
		if len(args) >= 3 && (args[0] == "sh" || args[0] == "bash" || strings.HasSuffix(args[0], "/sh") || strings.HasSuffix(args[0], "/bash")) && args[1] == "-c" {
			parser := shellwords.NewParser()
			parser.ParseEnv = false
			parser.ParseBacktick = false
			if parsed, err := parser.Parse(args[2]); err == nil {
				args = parsed
			}
		}
		flags := map[string]string{}
		for i, arg := range args {
			if key, value, ok := strings.Cut(arg, "="); ok && strings.HasPrefix(key, "--") {
				flags[key] = value
			} else if strings.HasPrefix(arg, "--") && i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
				flags[arg] = args[i+1]
			}
		}
		if name := flags["--served-model-name"]; name != "" {
			base.Models = []string{name}
		}
		tokenizer := flags["--tokenizer"]
		if tokenizer == "" {
			tokenizer = flags["--model"]
			for i, arg := range args {
				if tokenizer == "" && arg == "serve" && i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
					tokenizer = args[i+1]
				}
			}
		}
		if kind != "ollama" && tokenizer != "" {
			if !filepath.IsAbs(tokenizer) {
				base.Tokenizer = tokenizer
			} else {
				longest := 0
				for _, mount := range inspected.Mounts {
					relative, err := filepath.Rel(mount.Destination, tokenizer)
					if err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) && len(mount.Destination) > longest {
						candidate := filepath.Join(mount.Source, relative)
						if info, err := os.Stat(candidate); err == nil && info.IsDir() {
							base.Tokenizer = candidate
							longest = len(mount.Destination)
						}
					}
				}
			}
		}
		ports := map[int]string{}
		if instance.port > 0 {
			ports[instance.port] = "127.0.0.1"
		}
		if inspected.NetworkSettings != nil {
			for _, bindings := range inspected.NetworkSettings.Ports {
				for _, binding := range bindings {
					port, _ := strconv.Atoi(binding.HostPort)
					if port > 0 {
						host := binding.HostIP
						if host == "" || host == "0.0.0.0" || host == "::" {
							host = "127.0.0.1"
						}
						ports[port] = host
					}
				}
			}
		}
		if len(ports) == 0 && inspected.HostConfig != nil && inspected.HostConfig.NetworkMode == "host" {
			port, _ := strconv.Atoi(flags["--port"])
			if port > 0 {
				ports[port] = "127.0.0.1"
			}
		}
		ordered := []int{}
		for port := range ports {
			ordered = append(ordered, port)
		}
		sort.Ints(ordered)
		if len(ordered) == 0 {
			ordered = append(ordered, 0)
		}
		for _, port := range ordered {
			source := base
			source.ID = fmt.Sprintf("%d:%d", instance.id, port)
			if selected != "" && source.ID != selected {
				continue
			}
			if port > 0 {
				source.BaseURL = "http://" + net.JoinHostPort(ports[port], strconv.Itoa(port))
			}
			if selected == source.ID && source.BaseURL != "" && source.Status == "running" {
				endpoint := "/v1/models"
				if kind == "ollama" {
					endpoint = "/api/tags"
				}
				request, err := http.NewRequestWithContext(ctx, http.MethodGet, source.BaseURL+endpoint, nil)
				if err != nil {
					return nil, err
				}
				response, err := (&http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}).Do(request)
				if err != nil {
					source.Message = "model list unavailable; enter the served model name"
				} else {
					raw, readErr := io.ReadAll(io.LimitReader(response.Body, 2*1024*1024))
					response.Body.Close()
					if response.StatusCode == http.StatusOK && readErr == nil {
						var data struct {
							Data []struct {
								ID string `json:"id"`
							} `json:"data"`
							Models []struct {
								Name string `json:"name"`
							} `json:"models"`
						}
						if err := json.Unmarshal(raw, &data); err == nil {
							source.Models = []string{}
							for _, model := range data.Data {
								source.Models = append(source.Models, model.ID)
							}
							for _, model := range data.Models {
								source.Models = append(source.Models, model.Name)
							}
						}
					} else {
						source.Message = fmt.Sprintf("model list HTTP %d; provide model name and credentials if required", response.StatusCode)
					}
				}
			}
			results = append(results, source)
		}
	}
	return results, nil
}
