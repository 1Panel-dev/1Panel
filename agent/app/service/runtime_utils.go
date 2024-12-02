package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	cmd2 "github.com/1Panel-dev/1Panel/agent/utils/cmd"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func handleNodeAndJava(create request.RuntimeCreate, runtime *model.Runtime, fileOp files.FileOp, appVersionDir string) (err error) {
	runtimeDir := path.Join(constant.RuntimeDir, create.Type)
	if err = fileOp.CopyDir(appVersionDir, runtimeDir); err != nil {
		return
	}
	versionDir := path.Join(runtimeDir, filepath.Base(appVersionDir))
	projectDir := path.Join(runtimeDir, create.Name)
	defer func() {
		if err != nil {
			_ = fileOp.DeleteDir(projectDir)
		}
	}()
	if err = fileOp.Rename(versionDir, projectDir); err != nil {
		return
	}
	composeContent, envContent, _, err := handleParams(create, projectDir)
	if err != nil {
		return
	}
	runtime.DockerCompose = string(composeContent)
	runtime.Env = string(envContent)
	runtime.Status = constant.RuntimeCreating
	runtime.CodeDir = create.CodeDir

	nodeDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(runtime.AppDetailID))
	if err != nil {
		return err
	}

	go func() {
		RequestDownloadCallBack(nodeDetail.DownloadCallBackUrl)
	}()
	go startRuntime(runtime)

	return
}

func handlePHP(create request.RuntimeCreate, runtime *model.Runtime, fileOp files.FileOp, appVersionDir string) (err error) {
	runtimeDir := path.Join(constant.RuntimeDir, create.Type)
	if err = fileOp.CopyDirWithNewName(appVersionDir, runtimeDir, create.Name); err != nil {
		return
	}
	projectDir := path.Join(runtimeDir, create.Name)
	defer func() {
		if err != nil {
			_ = fileOp.DeleteDir(projectDir)
		}
	}()
	composeContent, envContent, forms, err := handleParams(create, projectDir)
	if err != nil {
		return
	}
	runtime.DockerCompose = string(composeContent)
	runtime.Env = string(envContent)
	runtime.Params = string(forms)
	runtime.Status = constant.RuntimeBuildIng

	go buildRuntime(runtime, "", "", false)
	return
}

func startRuntime(runtime *model.Runtime) {
	if err := runComposeCmdWithLog("up", runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		runtime.Status = constant.RuntimeError
		runtime.Message = err.Error()
		_ = runtimeRepo.Save(runtime)
		return
	}

	if err := SyncRuntimeContainerStatus(runtime); err != nil {
		runtime.Status = constant.RuntimeError
		runtime.Message = err.Error()
		_ = runtimeRepo.Save(runtime)
		return
	}
}

func reCreateRuntime(runtime *model.Runtime) {
	var err error
	defer func() {
		if err != nil {
			runtime.Status = constant.RuntimeError
			runtime.Message = err.Error()
			_ = runtimeRepo.Save(runtime)
		}
	}()
	if err = runComposeCmdWithLog("down", runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		return
	}
	if err = runComposeCmdWithLog("up", runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		return
	}
	if err := SyncRuntimeContainerStatus(runtime); err != nil {
		return
	}
}

func runComposeCmdWithLog(operate string, composePath string, logPath string) error {
	cmd := exec.Command("docker", "compose", "-f", composePath, operate)
	if operate == "up" {
		cmd = exec.Command("docker", "compose", "-f", composePath, operate, "-d")
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		global.LOG.Errorf("Failed to open log file: %v", err)
		return err
	}
	defer logFile.Close()
	multiWriterStdout := io.MultiWriter(os.Stdout, logFile)
	cmd.Stdout = multiWriterStdout
	var stderrBuf bytes.Buffer
	multiWriterStderr := io.MultiWriter(&stderrBuf, logFile, os.Stderr)
	cmd.Stderr = multiWriterStderr

	err = cmd.Run()
	if err != nil {
		return errors.New(buserr.New(constant.ErrRuntimeStart).Error() + ":" + err.Error())
	}
	return nil
}

func SyncRuntimeContainerStatus(runtime *model.Runtime) error {
	env, err := gotenv.Unmarshal(runtime.Env)
	if err != nil {
		return err
	}
	var containerNames []string
	if containerName, ok := env["CONTAINER_NAME"]; !ok {
		return buserr.New("ErrContainerNameNotFound")
	} else {
		containerNames = append(containerNames, containerName)
	}
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	containers, err := cli.ListContainersByName(containerNames)
	if err != nil {
		return err
	}
	if len(containers) == 0 {
		return buserr.WithNameAndErr("ErrContainerNotFound", containerNames[0], nil)
	}
	container := containers[0]

	switch container.State {
	case "exited":
		runtime.Status = constant.RuntimeError
	case "running":
		runtime.Status = constant.RuntimeRunning
	case "paused":
		runtime.Status = constant.RuntimeStopped
	default:
		if runtime.Status != constant.RuntimeBuildIng {
			runtime.Status = constant.RuntimeStopped
		}
	}

	return runtimeRepo.Save(runtime)
}

func getRuntimeEnv(envStr, key string) string {
	env, err := gotenv.Unmarshal(envStr)
	if err != nil {
		return ""
	}
	if v, ok := env[key]; ok {
		return v
	}
	return ""
}

func buildRuntime(runtime *model.Runtime, oldImageID string, oldEnv string, rebuild bool) {
	runtimePath := runtime.GetPath()
	composePath := runtime.GetComposePath()
	logPath := path.Join(runtimePath, "build.log")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		global.LOG.Errorf("failed to open log file: %v", err)
		return
	}
	defer func() {
		_ = logFile.Close()
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker-compose", "-f", composePath, "build")
	cmd.Stdout = logFile
	var stderrBuf bytes.Buffer
	multiWriterStderr := io.MultiWriter(&stderrBuf, logFile)
	cmd.Stderr = multiWriterStderr

	err = cmd.Run()
	if err != nil {
		runtime.Status = constant.RuntimeError
		runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + stderrBuf.String()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + buserr.New("ErrCmdTimeout").Error()
		} else {
			runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + stderrBuf.String()
		}
	} else {
		if err = runComposeCmdWithLog(constant.RuntimeDown, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
			return
		}
		runtime.Message = ""
		if oldImageID != "" {
			client, err := docker.NewClient()
			if err == nil {
				defer client.Close()
				newImageID, err := client.GetImageIDByName(runtime.Image)
				if err == nil && newImageID != oldImageID {
					global.LOG.Infof("delete imageID [%s] ", oldImageID)
					if err := client.DeleteImage(oldImageID); err != nil {
						global.LOG.Errorf("delete imageID [%s] error %v", oldImageID, err)
					} else {
						global.LOG.Infof("delete old image success")
					}
				}
			} else {
				global.LOG.Errorf("delete imageID [%s] error %v", oldImageID, err)
			}
		}
		if rebuild && runtime.ID > 0 {
			extensionsStr := getRuntimeEnv(runtime.Env, "PHP_EXTENSIONS")
			extensionsArray := strings.Split(extensionsStr, ",")
			oldExtensionStr := getRuntimeEnv(oldEnv, "PHP_EXTENSIONS")
			oldExtensionArray := strings.Split(oldExtensionStr, ",")
			var delExtensions []string
			for _, oldExt := range oldExtensionArray {
				exist := false
				for _, ext := range extensionsArray {
					if oldExt == ext {
						exist = true
						break
					}
				}
				if !exist {
					delExtensions = append(delExtensions, oldExt)
				}
			}

			if err = unInstallPHPExtension(runtime, delExtensions); err != nil {
				global.LOG.Errorf("unInstallPHPExtension error %v", err)
			}
		}

		if out, err := compose.Up(composePath); err != nil {
			runtime.Status = constant.RuntimeStartErr
			runtime.Message = out
		} else {
			extensions := getRuntimeEnv(runtime.Env, "PHP_EXTENSIONS")
			if extensions != "" {
				installCmd := fmt.Sprintf("docker exec -i %s %s %s", runtime.ContainerName, "install-ext", extensions)
				err = cmd2.ExecWithLogFile(installCmd, 60*time.Minute, logPath)
				if err != nil {
					runtime.Status = constant.RuntimeError
					runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + err.Error()
					_ = runtimeRepo.Save(runtime)
					return
				}
			}
			runtime.Status = constant.RuntimeRunning
		}
	}
	_ = runtimeRepo.Save(runtime)
}

func handleParams(create request.RuntimeCreate, projectDir string) (composeContent []byte, envContent []byte, forms []byte, err error) {
	fileOp := files.NewFileOp()
	composeContent, err = fileOp.GetContent(path.Join(projectDir, "docker-compose.yml"))
	if err != nil {
		return
	}
	envPath := path.Join(projectDir, ".env")
	if !fileOp.Stat(envPath) {
		_ = fileOp.CreateFile(envPath)
	}
	env, err := gotenv.Read(envPath)
	if err != nil {
		return
	}
	for k := range env {
		if strings.HasPrefix(k, "CONTAINER_PORT_") || strings.HasPrefix(k, "HOST_PORT_") {
			delete(env, k)
		}
	}
	switch create.Type {
	case constant.RuntimePHP:
		create.Params["IMAGE_NAME"] = create.Image
		var fromYml []byte
		fromYml, err = fileOp.GetContent(path.Join(projectDir, "data.yml"))
		if err != nil {
			return
		}
		var data dto.PHPForm
		err = yaml.Unmarshal(fromYml, &data)
		if err != nil {
			return
		}
		formFields := data.AdditionalProperties.FormFields
		forms, err = json.MarshalIndent(map[string]interface{}{
			"formFields": formFields,
		}, "", "  ")
		if err != nil {
			return
		}
		if extends, ok := create.Params["PHP_EXTENSIONS"]; ok {
			if extendsArray, ok := extends.([]interface{}); ok {
				strArray := make([]string, len(extendsArray))
				for i, v := range extendsArray {
					strArray[i] = strings.ToLower(fmt.Sprintf("%v", v))
				}
				create.Params["PHP_EXTENSIONS"] = strings.Join(strArray, ",")
			}
		}
		create.Params["CONTAINER_PACKAGE_URL"] = create.Source
		siteDir, _ := settingRepo.Get(settingRepo.WithByKey("WEBSITE_DIR"))
		create.Params["PANEL_WEBSITE_DIR"] = siteDir.Value
	case constant.RuntimeNode:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["NODE_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
		if create.NodeConfig.Install {
			create.Params["RUN_INSTALL"] = "1"
		} else {
			create.Params["RUN_INSTALL"] = "0"
		}
		create.Params["CONTAINER_PACKAGE_URL"] = create.Source
		create.Params["NODE_APP_PORT"] = create.Params["APP_PORT"]
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeJava:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["JAVA_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
		create.Params["JAVA_APP_PORT"] = create.Params["APP_PORT"]
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeGo:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["GO_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
		create.Params["GO_APP_PORT"] = create.Params["APP_PORT"]
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimePython:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["PYTHON_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeDotNet:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["DOTNET_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	}

	newMap := make(map[string]string)
	handleMap(create.Params, newMap)
	for k, v := range newMap {
		env[k] = v
	}

	envStr, err := gotenv.Marshal(env)
	if err != nil {
		return
	}
	if err = gotenv.Write(env, envPath); err != nil {
		return
	}
	envContent = []byte(envStr)
	return
}

func handleCompose(env gotenv.Env, composeContent []byte, create request.RuntimeCreate, projectDir string) (composeByte []byte, err error) {
	existMap := make(map[string]interface{})
	composeMap := make(map[string]interface{})
	if err = yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return
	}
	services, serviceValid := composeMap["services"].(map[string]interface{})
	if !serviceValid {
		err = buserr.New(constant.ErrFileParse)
		return
	}
	serviceName := ""
	serviceValue := make(map[string]interface{})
	for name, service := range services {
		serviceName = name
		serviceValue = service.(map[string]interface{})
		_, ok := serviceValue["ports"].([]interface{})
		if ok {
			var ports []interface{}
			switch create.Type {
			case constant.RuntimeNode:
				ports = append(ports, "${HOST_IP}:${PANEL_APP_PORT_HTTP}:${NODE_APP_PORT}")
			case constant.RuntimeJava:
				ports = append(ports, "${HOST_IP}:${PANEL_APP_PORT_HTTP}:${JAVA_APP_PORT}")
			case constant.RuntimeGo:
				ports = append(ports, "${HOST_IP}:${PANEL_APP_PORT_HTTP}:${GO_APP_PORT}")
			case constant.RuntimePython, constant.RuntimeDotNet:
				ports = append(ports, "${HOST_IP}:${PANEL_APP_PORT_HTTP}:${APP_PORT}")

			}
			for i, port := range create.ExposedPorts {
				containerPortStr := fmt.Sprintf("CONTAINER_PORT_%d", i)
				hostPortStr := fmt.Sprintf("HOST_PORT_%d", i)
				existMap[containerPortStr] = struct{}{}
				existMap[hostPortStr] = struct{}{}
				ports = append(ports, fmt.Sprintf("${HOST_IP}:${%s}:${%s}", hostPortStr, containerPortStr))
				create.Params[containerPortStr] = port.ContainerPort
				create.Params[hostPortStr] = port.HostPort
			}
			serviceValue["ports"] = ports
		}
		var environments []interface{}
		for _, e := range create.Environments {
			environments = append(environments, fmt.Sprintf("%s:%s", e.Key, e.Value))
		}
		delete(serviceValue, "environment")
		if len(environments) > 0 {
			serviceValue["environment"] = environments
		}
		var volumes []interface{}
		defaultVolumes := make(map[string]string)
		switch create.Type {
		case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimePython:
			defaultVolumes = constant.RuntimeDefaultVolumes
		case constant.RuntimeGo:
			defaultVolumes = constant.GoDefaultVolumes
		}
		for k, v := range defaultVolumes {
			volumes = append(volumes, fmt.Sprintf("%s:%s", k, v))
		}
		for _, volume := range create.Volumes {
			volumes = append(volumes, fmt.Sprintf("%s:%s", volume.Source, volume.Target))
		}
		serviceValue["volumes"] = volumes
		break
	}
	for k := range env {
		if strings.Contains(k, "CONTAINER_PORT_") || strings.Contains(k, "HOST_PORT_") {
			if _, ok := existMap[k]; !ok {
				delete(env, k)
			}
		}
	}

	services[serviceName] = serviceValue
	composeMap["services"] = services
	composeByte, err = yaml.Marshal(composeMap)
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	_ = fileOp.SaveFile(path.Join(projectDir, "docker-compose.yml"), string(composeByte), 0644)
	return
}

func checkContainerName(name string) error {
	dockerCli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer dockerCli.Close()
	names, err := dockerCli.ListContainersByName([]string{name})
	if err != nil {
		return err
	}
	if len(names) > 0 {
		return buserr.New(constant.ErrContainerName)
	}
	return nil
}

func checkContainerStatus(name string) (string, error) {
	dockerCli, err := docker.NewClient()
	if err != nil {
		return "", err
	}
	defer dockerCli.Close()
	names, err := dockerCli.ListContainersByName([]string{name})
	if err != nil {
		return "", err
	}
	if len(names) > 0 {
		return names[0].State, nil
	}
	return "", nil
}

func unInstallPHPExtension(runtime *model.Runtime, delExtensions []string) error {
	dir := runtime.GetPath()
	fileOP := files.NewFileOp()
	var phpExtensions []response.SupportExtension
	if err := json.Unmarshal(nginx_conf.PHPExtensionsJson, &phpExtensions); err != nil {
		return err
	}
	delMap := make(map[string]struct{})
	for _, ext := range phpExtensions {
		for _, del := range delExtensions {
			if ext.Check == del {
				delMap[ext.Check] = struct{}{}
				_ = fileOP.DeleteFile(path.Join(dir, "extensions", ext.File))
				_ = fileOP.DeleteFile(path.Join(dir, "conf", "conf.d", "docker-php-ext-"+ext.Check+".ini"))
				_ = removePHPIniExt(path.Join(dir, "conf", "php.ini"), ext.File)
				break
			}
		}
	}
	extensions := getRuntimeEnv(runtime.Env, "PHP_EXTENSIONS")
	var (
		oldExts []string
		newExts []string
	)
	oldExts = strings.Split(extensions, ",")
	for _, ext := range oldExts {
		if _, ok := delMap[ext]; !ok {
			newExts = append(newExts, ext)
		}
	}
	newExtensions := strings.Join(newExts, ",")
	envs, err := gotenv.Unmarshal(runtime.Env)
	if err != nil {
		return err
	}
	envs["PHP_EXTENSIONS"] = newExtensions
	if err = gotenv.Write(envs, runtime.GetEnvPath()); err != nil {
		return err
	}
	envContent, err := gotenv.Marshal(envs)
	if err != nil {
		return err
	}
	runtime.Env = envContent
	return nil
}

func removePHPIniExt(filePath, extensionName string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	targetLine1 := `extension="` + extensionName + `"`
	targetLine2 := `zend_extension="` + extensionName + `"`

	tempFile, err := os.CreateTemp(path.Dir(filePath), "temp_*.txt")
	if err != nil {
		return err
	}
	defer tempFile.Close()

	scanner := bufio.NewScanner(file)
	writer := bufio.NewWriter(tempFile)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, targetLine1) && !strings.Contains(line, targetLine2) {
			_, err := writer.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return os.Rename(tempFile.Name(), filePath)
}

func restartRuntime(runtime *model.Runtime) (err error) {
	if err = runComposeCmdWithLog(constant.RuntimeDown, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		return
	}
	if err = runComposeCmdWithLog(constant.RuntimeUp, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		return
	}
	return
}

func getDockerComposeEnvironments(yml []byte) ([]request.Environment, error) {
	var (
		composeProject docker.ComposeProject
		err            error
	)
	err = yaml.Unmarshal(yml, &composeProject)
	if err != nil {
		return nil, err
	}
	var res []request.Environment
	for _, service := range composeProject.Services {
		for _, env := range service.Environment {
			envArray := strings.Split(env, ":")
			key := envArray[0]
			value := ""
			if len(envArray) > 1 {
				value = envArray[1]
			}
			res = append(res, request.Environment{
				Key:   key,
				Value: value,
			})
		}
	}
	return res, nil
}

func getDockerComposeVolumes(yml []byte) ([]request.Volume, error) {
	var (
		composeProject docker.ComposeProject
		err            error
	)
	err = yaml.Unmarshal(yml, &composeProject)
	if err != nil {
		return nil, err
	}
	var res []request.Volume
	for _, service := range composeProject.Services {
		for _, volume := range service.Volumes {
			envArray := strings.Split(volume, ":")
			source := envArray[0]
			target := ""
			if len(envArray) > 1 {
				target = envArray[1]
			}
			res = append(res, request.Volume{
				Source: source,
				Target: target,
			})
		}
	}
	return res, nil
}
