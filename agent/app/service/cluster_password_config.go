package service

import (
	"bytes"
	"context"
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/env"
	"github.com/compose-spec/compose-go/v2/dotenv"
	"github.com/compose-spec/compose-go/v2/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

type clusterPasswordConfig struct {
	project     string
	service     string
	previousEnv []byte
	nextEnv     []byte
	environment map[string]string
	changes     map[string]string
	ports       map[clusterPasswordPortBinding]int
	recreate    bool
}

func prepareClusterPasswordConfig(ctx context.Context, cli *client.Client, db clusterDatabase, info container.InspectResponse, req ClusterPasswordMemberRequest, changes map[string]string, forceRecreate bool) (clusterPasswordConfig, error) {
	var config clusterPasswordConfig
	environmentChanged := forceRecreate
	invalid := buserr.New("ErrClusterPasswordConfiguration")
	install := db.install
	for _, filename := range []string{install.GetEnvPath(), install.GetComposePath()} {
		stat, err := os.Lstat(filename)
		if err != nil || !stat.Mode().IsRegular() {
			return config, invalid
		}
	}
	rawEnv, err := os.ReadFile(install.GetEnvPath())
	if err != nil {
		return config, invalid
	}
	nextEnv, values, err := replaceClusterPasswordEnv(rawEnv, changes)
	if err != nil {
		return config, err
	}
	rawCompose, err := os.ReadFile(install.GetComposePath())
	if err != nil {
		return config, invalid
	}
	projectName := info.Config.Labels["com.docker.compose.project"]
	serviceName := info.Config.Labels["com.docker.compose.service"]
	if projectName == "" || serviceName == "" || serviceName != install.ServiceName {
		return config, invalid
	}
	project, err := docker.GetComposeProjectWithEnv(projectName, install.GetPath(), rawCompose, values, false)
	if err != nil || project.Name != projectName {
		return config, invalid
	}
	service, ok := project.Services[serviceName]
	if !ok || service.ContainerName != install.ContainerName || service.Image == "" || service.Build != nil {
		return config, invalid
	}
	for variable, value := range clusterPasswordContainerEnvironment(req.App, changes) {
		if configured := service.Environment[variable]; configured == nil || *configured != value {
			return config, invalid
		}
		environmentChanged = environmentChanged || !containsRestoreEnv(info.Config.Env, variable+"="+value)
	}
	config.project, config.service = projectName, serviceName
	config.previousEnv, config.nextEnv, config.environment = rawEnv, nextEnv, values
	config.changes, config.recreate = changes, environmentChanged
	if config.recreate {
		if err := preserveClusterPasswordPortBindings(&config, info.HostConfig, project, install.GetPath(), rawCompose); err != nil {
			return config, err
		}
	}
	if config.recreate || !bytes.Equal(config.previousEnv, config.nextEnv) {
		probe, cancel := context.WithTimeout(ctx, 12*time.Second)
		err := compose.ValidatePasswordService(probe, install.GetComposePath(), projectName, serviceName, db.dockerHost, config.environment, config.nextEnv)
		cancel()
		if err != nil {
			return config, invalid
		}
	}
	if config.recreate {
		image, err := cli.ImageInspect(ctx, service.Image)
		if err != nil || image.ID != info.Image {
			return config, invalid
		}
		data, destination := originalRestoreMount(req.App)
		expected, err := filepath.EvalSymlinks(filepath.Join(install.GetPath(), data))
		if err != nil {
			return config, invalid
		}
		active, planned := false, false
		for _, volume := range info.Mounts {
			if volume.Destination == destination && volume.Type == mount.TypeBind && volume.RW {
				physical, err := filepath.EvalSymlinks(volume.Source)
				active = err == nil && physical == expected
			}
		}
		for _, volume := range service.Volumes {
			if volume.Target == destination && volume.Type == "bind" && !volume.ReadOnly {
				physical, err := filepath.EvalSymlinks(volume.Source)
				planned = err == nil && physical == expected
			}
		}
		if !active || !planned {
			return config, invalid
		}
	}
	return config, nil
}

type clusterPasswordPortBinding struct {
	target    uint32
	protocol  string
	published string
	hostIP    string
}

func preserveClusterPasswordPortBindings(config *clusterPasswordConfig, active *container.HostConfig, project *types.Project, workDir string, composeData []byte) error {
	mismatch := buserr.New("ErrClusterPasswordPortBindings")
	bindings, ok := clusterPasswordDockerPortBindings(active)
	if !ok {
		return mismatch
	}
	config.ports = bindings
	service := project.Services[config.service]
	planned := clusterPasswordComposePortBindings(service.Ports)
	if maps.Equal(bindings, planned) {
		return nil
	}
	var hostIP string
	var found bool
	for binding, count := range bindings {
		if count <= planned[binding] {
			continue
		}
		if found && binding.hostIP != hostIP {
			return mismatch
		}
		hostIP, found = binding.hostIP, true
	}
	if !found {
		return mismatch
	}
	changes := maps.Clone(config.changes)
	changes[constant.HostIP] = hostIP
	nextEnv, values, err := replaceClusterPasswordEnv(config.previousEnv, changes)
	if err != nil {
		return mismatch
	}
	adjusted, err := docker.GetComposeProjectWithEnv(config.project, workDir, composeData, values, false)
	if err != nil {
		return mismatch
	}
	updatedService := adjusted.Services[config.service]
	if !maps.Equal(bindings, clusterPasswordComposePortBindings(updatedService.Ports)) {
		return mismatch
	}
	updatedService.Ports = service.Ports
	adjusted.Services[config.service] = updatedService
	adjusted.Environment = project.Environment
	if !reflect.DeepEqual(adjusted, project) {
		return mismatch
	}
	config.nextEnv, config.environment, config.changes = nextEnv, values, changes
	return nil
}

func (config clusterPasswordConfig) matchesPorts(active *container.HostConfig) bool {
	bindings, ok := clusterPasswordDockerPortBindings(active)
	return ok && maps.Equal(config.ports, bindings)
}

func clusterPasswordDockerPortBindings(active *container.HostConfig) (map[clusterPasswordPortBinding]int, bool) {
	if active == nil || active.PublishAllPorts {
		return nil, false
	}
	bindings := make(map[clusterPasswordPortBinding]int)
	for port, published := range active.PortBindings {
		for _, binding := range published {
			number, err := strconv.ParseUint(binding.HostPort, 10, 16)
			if err != nil || number == 0 {
				return nil, false
			}
			bindings[clusterPasswordPortBinding{uint32(port.Int()), port.Proto(), binding.HostPort, binding.HostIP}]++
		}
	}
	return bindings, true
}

func clusterPasswordComposePortBindings(ports []types.ServicePortConfig) map[clusterPasswordPortBinding]int {
	bindings := make(map[clusterPasswordPortBinding]int, len(ports))
	for _, port := range ports {
		bindings[clusterPasswordPortBinding{port.Target, port.Protocol, port.Published, port.HostIP}]++
	}
	return bindings
}

func clusterPasswordContainerEnvironment(app string, changes map[string]string) map[string]string {
	credentials := make(map[string]string, len(changes)+1)
	for key, password := range changes {
		switch key {
		case clusterSQLPassword:
			if app == constant.AppPostgresqlCluster {
				credentials["POSTGRESQL_PASSWORD"] = password
			} else {
				credentials["MYSQL_ROOT_PASSWORD"] = password
			}
		case clusterRedisPassword:
			credentials["REDIS_PASSWORD"] = password
			credentials["REDIS_EXTRA_FLAGS"] = "--requirepass " + password
		case clusterSourcePassword:
			if app == constant.AppMysqlCluster {
				credentials["MYSQL_MASTER_ROOT_PASSWORD"] = password
			} else {
				credentials["REDIS_MASTER_PASSWORD"] = password
			}
		}
	}
	return credentials
}

func replaceClusterPasswordEnv(content []byte, changes map[string]string) ([]byte, map[string]string, error) {
	invalid := buserr.New("ErrClusterPasswordConfiguration")
	values, err := dotenv.Parse(bytes.NewReader(content))
	if err != nil {
		return nil, nil, invalid
	}
	lines := strings.Split(string(content), "\n")
	seen := make(map[string]bool, len(changes))
	for i, line := range lines {
		key, _, assignment := strings.Cut(line, "=")
		key = strings.TrimSpace(key)
		password, managed := changes[key]
		if !assignment || !managed {
			continue
		}
		if seen[key] {
			return nil, nil, invalid
		}
		seen[key] = true
		if values[key] == password {
			continue
		}
		encoded, err := env.MarshalWithOrder(map[string]string{key: password}, []string{key})
		if err != nil {
			return nil, nil, invalid
		}
		if strings.HasSuffix(line, "\r") {
			encoded += "\r"
		}
		lines[i] = encoded
	}
	if len(seen) != len(changes) {
		return nil, nil, invalid
	}
	result := []byte(strings.Join(lines, "\n"))
	nextValues, err := dotenv.Parse(bytes.NewReader(result))
	if err != nil || len(values) != len(nextValues) {
		return nil, nil, invalid
	}
	for key, old := range values {
		expected, ok := changes[key]
		if !ok {
			expected = old
		}
		if value, ok := nextValues[key]; !ok || value != expected {
			return nil, nil, invalid
		}
	}
	return result, nextValues, nil
}

func ensureRedisPasswordPersistence(ctx context.Context, db clusterDatabase, password string) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	notPersisted := buserr.New("ErrClusterPasswordRedisPersistence")
	directory, dirErr := redisClusterPasswordQuery(ctx, db, password, "CONFIG", "GET", "dir")
	settings, infoErr := redisClusterPasswordQuery(ctx, db, password, "INFO", "persistence")
	shutdown, shutdownErr := redisClusterPasswordQuery(ctx, db, password, "CONFIG", "GET", "shutdown-on-sigterm")
	fields := clusterPasswordStatusFields(settings)
	if dirErr != nil || infoErr != nil || shutdownErr != nil || directory != "dir\n/bitnami/redis/data" ||
		fields["loading"] != "0" || strings.Contains(strings.ToLower(shutdown), "nosave") {
		return notPersisted
	}
	if fields["aof_enabled"] == "1" && fields["aof_last_write_status"] == "ok" {
		return nil
	}
	save, err := redisClusterPasswordQuery(ctx, db, password, "CONFIG", "GET", "save")
	if fields["aof_enabled"] != "0" || fields["rdb_last_bgsave_status"] != "ok" || err != nil || !strings.HasPrefix(save, "save\n") {
		return notPersisted
	}
	return nil
}
