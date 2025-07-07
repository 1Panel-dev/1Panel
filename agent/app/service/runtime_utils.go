package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	cmd2 "github.com/1Panel-dev/1Panel/agent/utils/cmd"

	"github.com/1Panel-dev/1Panel/agent/i18n"
	"github.com/1Panel-dev/1Panel/agent/utils/common"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/compose"
	"github.com/1Panel-dev/1Panel/agent/utils/docker"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"github.com/pkg/errors"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
)

func handleRuntime(create request.RuntimeCreate, runtime *model.Runtime, fileOp files.FileOp, appVersionDir string) (err error) {
	runtimeDir := path.Join(global.Dir.RuntimeDir, create.Type)
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
	runtime.Status = constant.StatusCreating
	runtime.CodeDir = create.CodeDir

	nodeDetail, err := appDetailRepo.GetFirst(repo.WithByID(runtime.AppDetailID))
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
	runtimeDir := path.Join(global.Dir.RuntimeDir, create.Type)
	if err = fileOp.CopyDirWithNewName(appVersionDir, runtimeDir, create.Name); err != nil {
		return
	}
	projectDir := path.Join(runtimeDir, create.Name)
	defer func() {
		if err != nil {
			_ = fileOp.DeleteDir(projectDir)
		}
	}()

	version, ok := create.Params["PHP_VERSION"]
	if ok {
		extensionsDir := path.Join(projectDir, "extensions", getExtensionDir(version.(string)))
		_ = fileOp.CreateDir(extensionsDir, 0755)
	}

	composeContent, envContent, forms, err := handleParams(create, projectDir)
	if err != nil {
		return
	}
	runtime.DockerCompose = string(composeContent)
	runtime.Env = string(envContent)
	runtime.Params = string(forms)
	runtime.Status = constant.StatusBuilding

	go func() {
		appDetail, err := appDetailRepo.GetFirst(repo.WithByID(runtime.AppDetailID))
		if err == nil {
			RequestDownloadCallBack(appDetail.DownloadCallBackUrl)
		}
	}()

	go buildRuntime(runtime, "", "", false)
	return
}

func startRuntime(runtime *model.Runtime) {
	if err := runComposeCmdWithLog("up", runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		runtime.Status = constant.StatusError
		runtime.Message = err.Error()
		_ = runtimeRepo.Save(runtime)
		return
	}

	if err := SyncRuntimeContainerStatus(runtime); err != nil {
		runtime.Status = constant.StatusError
		runtime.Message = err.Error()
		_ = runtimeRepo.Save(runtime)
		return
	}
}

func reCreateRuntime(runtime *model.Runtime) {
	var err error
	defer func() {
		if err != nil {
			runtime.Status = constant.StatusError
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

func getComposeCmd(composePath, operate string) *exec.Cmd {
	dockerCommand := global.CONF.DockerConfig.Command
	var cmd *exec.Cmd
	if dockerCommand == "docker-compose" {
		if operate == "up" {
			cmd = exec.Command("docker-compose", "-f", composePath, operate, "-d")
		} else {
			cmd = exec.Command("docker-compose", "-f", composePath, operate)
		}
	} else {
		if operate == "up" {
			cmd = exec.Command("docker", "compose", "-f", composePath, operate, "-d")
		} else {
			cmd = exec.Command("docker", "compose", "-f", composePath, operate)
		}
	}
	return cmd
}

func runComposeCmdWithLog(operate string, composePath string, logPath string) error {
	cmd := getComposeCmd(composePath, operate)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, constant.FilePerm)
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
		return errors.New(buserr.New("ErrRuntimeStart").Error() + ":" + stderrBuf.String())
	}
	return nil
}

func SyncRuntimesStatus(runtimes []model.Runtime) error {
	cli, err := docker.NewClient()
	if err != nil {
		return err
	}
	defer cli.Close()
	var containerNames []string
	runtimeContainer := make(map[string]int)
	for index, runtime := range runtimes {
		containerNames = append(containerNames, runtime.ContainerName)
		runtimeContainer["/"+runtime.ContainerName] = index
	}
	containers, err := cli.ListContainersByName(containerNames)
	if err != nil {
		return err
	}
	for _, contain := range containers {
		if index, ok := runtimeContainer[contain.Names[0]]; ok {
			if runtimes[index].Status == constant.StatusBuilding || runtimes[index].Status == constant.StatusCreating {
				delete(runtimeContainer, contain.Names[0])
				continue
			}
			switch contain.State {
			case "exited":
				runtimes[index].Status = constant.StatusError
			case "running":
				runtimes[index].Status = constant.StatusRunning
			case "paused":
				runtimes[index].Status = constant.StatusStopped
			case "restarting":
				runtimes[index].Status = constant.StatusRestarting
			}
			_ = runtimeRepo.Save(&runtimes[index])
			delete(runtimeContainer, contain.Names[0])
		}
	}
	for _, index := range runtimeContainer {
		if runtimes[index].Status != constant.StatusBuilding && runtimes[index].Status != constant.StatusCreating {
			runtimes[index].Status = constant.StatusStopped
		}
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
		runtime.Status = constant.StatusError
	case "running":
		runtime.Status = constant.StatusRunning
	case "paused":
		runtime.Status = constant.StatusStopped
	default:
		if runtime.Status != constant.StatusBuilding {
			runtime.Status = constant.StatusStopped
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

func deleteImageByID(oldImageID, imageName string, client docker.Client) {
	newImageID, err := client.GetImageIDByName(imageName)
	if err == nil && newImageID != oldImageID {
		global.LOG.Infof("delete imageID [%s] ", oldImageID)
		if err := client.DeleteImage(oldImageID); err != nil {
			global.LOG.Errorf("delete imageID [%s] error %v", oldImageID, err)
		} else {
			global.LOG.Infof("delete old image success")
		}
	}
}

func buildRuntime(runtime *model.Runtime, oldImageID string, oldEnv string, rebuild bool) {
	runtimePath := runtime.GetPath()
	composePath := runtime.GetComposePath()
	logPath := path.Join(runtimePath, "build.log")

	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, constant.FilePerm)
	if err != nil {
		global.LOG.Errorf("failed to open log file: %v", err)
		return
	}
	defer func() {
		_ = logFile.Close()
	}()

	newPHPVersion := getRuntimeEnv(runtime.Env, "PHP_VERSION")
	oldPHPVersion := getRuntimeEnv(oldEnv, "PHP_VERSION")
	if newPHPVersion != oldPHPVersion {
		_ = os.Rename(path.Join(runtimePath, "extensions", getExtensionDir(oldPHPVersion)), path.Join(runtimePath, "extensions", getExtensionDir(newPHPVersion)))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Hour)
	defer cancel()

	dockerCommand := global.CONF.DockerConfig.Command
	var cmd *exec.Cmd
	if dockerCommand == "docker-compose" {
		cmd = exec.CommandContext(ctx, "docker-compose", "-f", composePath, "build")
	} else {
		cmd = exec.CommandContext(ctx, "docker", "compose", "-f", composePath, "build")
	}
	cmd.Stdout = logFile
	var stderrBuf bytes.Buffer
	multiWriterStderr := io.MultiWriter(&stderrBuf, logFile)
	cmd.Stderr = multiWriterStderr

	err = cmd.Run()
	if err != nil {
		runtime.Status = constant.StatusError
		runtime.Message = buserr.New("ErrImageBuildErr").Error() + ":" + stderrBuf.String()
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			runtime.Message = buserr.New("ErrImageBuildErr").Error() + ":" + buserr.New("ErrCmdTimeout").Error()
		} else {
			runtime.Message = buserr.New("ErrImageBuildErr").Error() + ":" + stderrBuf.String()
		}
		_ = runtimeRepo.Save(runtime)
		return
	}
	if err = runComposeCmdWithLog(constant.RuntimeDown, runtime.GetComposePath(), runtime.GetLogPath()); err != nil {
		return
	}
	client, err := docker.NewClient()
	if err != nil {
		_, _ = logFile.WriteString(fmt.Sprintf("failed to connect to docker client: %v", err))
		return
	}
	runtime.Message = ""
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
			_, _ = logFile.WriteString(fmt.Sprintf("unInstallPHPExtension error %v", err))
		}
	}

	defer func() {
		_ = runtimeRepo.Save(runtime)
	}()

	if out, err := compose.Up(composePath); err != nil {
		runtime.Status = constant.StatusStartErr
		runtime.Message = out
		return
	}
	deleteImageID := ""
	extensions := getRuntimeEnv(runtime.Env, "PHP_EXTENSIONS")
	if extensions != "" {
		deleteImageID, _ = client.GetImageIDByName(runtime.Image)
		cmdMgr := cmd2.NewCommandMgr(cmd2.WithTimeout(60*time.Minute), cmd2.WithOutputFile(logPath))
		if err = cmdMgr.Run("docker", "exec", "-i", runtime.ContainerName, "install-ext", extensions); err != nil {
			runtime.Status = constant.StatusError
			runtime.Message = buserr.New("ErrImageBuildErr").Error() + ":" + err.Error()
			return
		}
		commitMgr := cmd2.NewCommandMgr(cmd2.WithTimeout(10*time.Minute), cmd2.WithOutputFile(logPath))
		err = commitMgr.Run("docker", "commit", runtime.ContainerName, runtime.Image)
		if err != nil {
			runtime.Status = constant.StatusError
			runtime.Message = buserr.New("ErrImageBuildErr").Error() + ":" + err.Error()
			return
		}
	}
	if oldImageID != "" {
		deleteImageByID(oldImageID, runtime.Image, client)
	}
	if deleteImageID != "" {
		deleteImageByID(deleteImageID, runtime.Image, client)
	}
	handlePHPDir(*runtime)
	if out, err := compose.DownAndUp(composePath); err != nil {
		runtime.Status = constant.StatusStartErr
		runtime.Message = out
		return
	}
	runtime.Status = constant.StatusRunning
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
		if strings.HasPrefix(k, "CONTAINER_PORT_") || strings.HasPrefix(k, "HOST_PORT_") || strings.HasPrefix(k, "HOST_IP_") || strings.Contains(k, "APP_PORT") {
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
		if siteDir.Value == "" {
			siteDir.Value = path.Join(global.Dir.BaseDir, "1panel", "www")
		}
		create.Params["PANEL_WEBSITE_DIR"] = siteDir.Value
		composeContent, err = handleEnvironments(composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeNode:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["NODE_VERSION"] = create.Version
		if create.NodeConfig.Install {
			create.Params["RUN_INSTALL"] = "1"
		} else {
			create.Params["RUN_INSTALL"] = "0"
		}
		create.Params["CONTAINER_PACKAGE_URL"] = create.Source
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeJava:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["JAVA_VERSION"] = create.Version
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeGo:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["GO_VERSION"] = create.Version
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimePython:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["PYTHON_VERSION"] = create.Version
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeDotNet:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["DOTNET_VERSION"] = create.Version
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

func handleEnvironments(composeContent []byte, create request.RuntimeCreate, projectDir string) (composeByte []byte, err error) {
	composeMap := make(map[string]interface{})
	if err = yaml.Unmarshal(composeContent, &composeMap); err != nil {
		return
	}
	services, serviceValid := composeMap["services"].(map[string]interface{})
	if !serviceValid {
		err = buserr.New("ErrFileParse")
		return
	}
	serviceName := ""
	serviceValue := make(map[string]interface{})
	for name, service := range services {
		serviceName = name
		serviceValue = service.(map[string]interface{})
		var environments []interface{}
		for _, e := range create.Environments {
			environments = append(environments, fmt.Sprintf("%s=%s", e.Key, e.Value))
		}
		delete(serviceValue, "environment")
		if len(environments) > 0 {
			serviceValue["environment"] = environments
		}
		break
	}
	services[serviceName] = serviceValue
	composeMap["services"] = services
	composeByte, err = yaml.Marshal(composeMap)
	if err != nil {
		return
	}
	fileOp := files.NewFileOp()
	_ = fileOp.SaveFile(path.Join(projectDir, "docker-compose.yml"), string(composeByte), constant.DirPerm)
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
		err = buserr.New("ErrFileParse")
		return
	}
	serviceName := ""
	serviceValue := make(map[string]interface{})
	for name, service := range services {
		serviceName = name
		serviceValue = service.(map[string]interface{})
		delete(serviceValue, "ports")
		if len(create.ExposedPorts) > 0 {
			var ports []interface{}
			for i, port := range create.ExposedPorts {
				containerPortStr := fmt.Sprintf("CONTAINER_PORT_%d", i)
				hostPortStr := fmt.Sprintf("HOST_PORT_%d", i)
				existMap[containerPortStr] = struct{}{}
				existMap[hostPortStr] = struct{}{}
				hostIPStr := fmt.Sprintf("HOST_IP_%d", i)
				ports = append(ports, fmt.Sprintf("${%s}:${%s}:${%s}", hostIPStr, hostPortStr, containerPortStr))
				create.Params[containerPortStr] = port.ContainerPort
				create.Params[hostPortStr] = port.HostPort
				create.Params[hostIPStr] = port.HostIP
			}
			if create.Type == constant.RuntimePHP {
				ports = append(ports, "127.0.0.1:${PANEL_APP_PORT_HTTP}:9000")
			}
			serviceValue["ports"] = ports
		} else {
			if create.Type == constant.RuntimePHP {
				serviceValue["ports"] = []interface{}{"127.0.0.1:${PANEL_APP_PORT_HTTP}:9000"}
			}
		}
		var environments []interface{}
		for _, e := range create.Environments {
			environments = append(environments, fmt.Sprintf("%s=%s", e.Key, e.Value))
		}
		delete(serviceValue, "environment")
		if len(environments) > 0 {
			serviceValue["environment"] = environments
		}
		var volumes []interface{}
		defaultVolumes := make(map[string]string)
		switch create.Type {
		case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimePython, constant.RuntimeDotNet:
			defaultVolumes = constant.RuntimeDefaultVolumes
		case constant.RuntimeGo:
			defaultVolumes = constant.GoDefaultVolumes
		case constant.RuntimePHP:
			defaultVolumes = constant.PHPDefaultVolumes
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
	_ = fileOp.SaveFile(path.Join(projectDir, "docker-compose.yml"), string(composeByte), constant.DirPerm)
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
		return buserr.New("ErrContainerName")
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
	if err := json.Unmarshal(nginx_conf.GetWebsiteFile("php_extensions.json"), &phpExtensions); err != nil {
		return err
	}
	phpVersion := getRuntimeEnv(runtime.Env, "PHP_VERSION")
	phpExtensionDir := path.Join(dir, "extensions", getExtensionDir(phpVersion))

	delMap := make(map[string]struct{})
	for _, ext := range phpExtensions {
		for _, del := range delExtensions {
			if ext.Name == del {
				delMap[ext.Check] = struct{}{}
				_ = fileOP.DeleteFile(path.Join(phpExtensionDir, ext.File))
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
		for key, value := range service.Environment.Variables {
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

func checkRuntimePortExist(port int, scanPort bool, runtimeID uint) error {
	errMap := make(map[string]interface{})
	errMap["port"] = port
	appInstall, _ := appInstallRepo.GetFirst(appInstallRepo.WithPort(port))
	if appInstall.ID > 0 {
		errMap["type"] = i18n.GetMsgByKey("TYPE_APP")
		errMap["name"] = appInstall.Name
		return buserr.WithMap("ErrPortExist", errMap, nil)
	}
	opts := []repo.DBOption{runtimeRepo.WithPort(port)}
	if runtimeID > 0 {
		opts = append(opts, repo.WithByNOTID(runtimeID))
	}
	runtime, _ := runtimeRepo.GetFirst(context.Background(), opts...)
	if runtime != nil {
		errMap["type"] = i18n.GetMsgByKey("TYPE_RUNTIME")
		errMap["name"] = runtime.Name
		return buserr.WithMap("ErrPortExist", errMap, nil)
	}
	domain, _ := websiteDomainRepo.GetFirst(websiteDomainRepo.WithPort(port))
	if domain.ID > 0 {
		errMap["type"] = i18n.GetMsgByKey("TYPE_DOMAIN")
		errMap["name"] = domain.Domain
		return buserr.WithMap("ErrPortExist", errMap, nil)
	}
	if scanPort && common.ScanPort(port) {
		return buserr.WithDetail("ErrPortInUsed", port, nil)
	}
	return nil
}

func getExtensionDir(version string) string {
	if strings.HasPrefix(version, "8.4") {
		return "no-debug-non-zts-20240924"
	}
	if strings.HasPrefix(version, "8.3") {
		return "no-debug-non-zts-20230831"
	}
	if strings.HasPrefix(version, "8.2") {
		return "no-debug-non-zts-20220829"
	}
	if strings.HasPrefix(version, "8.1") {
		return "no-debug-non-zts-20210902"
	}
	if strings.HasPrefix(version, "8.0") {
		return "no-debug-non-zts-20200930"
	}
	if strings.HasPrefix(version, "7.4") {
		return "no-debug-non-zts-20190902"
	}
	if strings.HasPrefix(version, "7.3") {
		return "no-debug-non-zts-20180731"
	}
	if strings.HasPrefix(version, "7.2") {
		return "no-debug-non-zts-20170718"
	}
	if strings.HasPrefix(version, "7.1") {
		return "no-debug-non-zts-20160303"
	}
	if strings.HasPrefix(version, "7.0") {
		return "no-debug-non-zts-20151012"
	}
	if strings.HasPrefix(version, "5.6") {
		return "no-debug-non-zts-20131226"
	}
	return ""
}

func RestartPHPRuntime() {
	runtimes, err := runtimeRepo.List(repo.WithByType(constant.RuntimePHP))
	if err != nil {
		return
	}
	websiteDir, _ := settingRepo.GetValueByKey("WEBSITE_DIR")
	for _, runtime := range runtimes {
		envs, err := gotenv.Unmarshal(runtime.Env)
		if err != nil {
			global.LOG.Warningf("restart php runtime failed %v", err)
			continue
		}
		envs["PANEL_WEBSITE_DIR"] = websiteDir
		if err = gotenv.Write(envs, runtime.GetEnvPath()); err != nil {
			global.LOG.Warningf("restart php runtime failed %v", err)
			continue
		}
		go func() {
			_ = restartRuntime(&runtime)
		}()
	}
}

func handleRuntimeDTO(res *response.RuntimeDTO, runtime model.Runtime) error {
	res.Params = make(map[string]interface{})
	envs, err := gotenv.Unmarshal(runtime.Env)
	if err != nil {
		return err
	}
	for k, v := range envs {
		if strings.Contains(k, "CONTAINER_PORT") || strings.Contains(k, "HOST_PORT") {
			if strings.Contains(k, "CONTAINER_PORT") {
				r := regexp.MustCompile(`_(\d+)$`)
				matches := r.FindStringSubmatch(k)
				containerPort, err := strconv.Atoi(v)
				if err != nil {
					return err
				}
				hostPort, err := strconv.Atoi(envs[fmt.Sprintf("HOST_PORT_%s", matches[1])])
				if err != nil {
					return err
				}
				hostIP := envs[fmt.Sprintf("HOST_IP_%s", matches[1])]
				if hostIP == "" {
					hostIP = "0.0.0.0"
				}
				res.ExposedPorts = append(res.ExposedPorts, request.ExposedPort{
					ContainerPort: containerPort,
					HostPort:      hostPort,
					HostIP:        hostIP,
				})
			}
		} else {
			res.Params[k] = v
		}
	}
	if v, ok := envs["CONTAINER_PACKAGE_URL"]; ok {
		res.Source = v
	}
	composeByte, err := files.NewFileOp().GetContent(runtime.GetComposePath())
	if err != nil {
		return err
	}
	res.Environments, err = getDockerComposeEnvironments(composeByte)
	if err != nil {
		return err
	}
	volumes, err := getDockerComposeVolumes(composeByte)
	if err != nil {
		return err
	}

	defaultVolumes := make(map[string]string)
	switch runtime.Type {
	case constant.RuntimeNode, constant.RuntimeJava, constant.RuntimePython, constant.RuntimeDotNet:
		defaultVolumes = constant.RuntimeDefaultVolumes
	case constant.RuntimeGo:
		defaultVolumes = constant.GoDefaultVolumes
	case constant.RuntimePHP:
		defaultVolumes = constant.PHPDefaultVolumes
	}
	for _, volume := range volumes {
		exist := false
		for key, value := range defaultVolumes {
			if key == volume.Source && value == volume.Target {
				exist = true
				break
			}
		}
		if !exist {
			res.Volumes = append(res.Volumes, volume)
		}
	}
	return nil
}

func handlePHPDir(runtime model.Runtime) {
	fileOp := files.NewFileOp()
	dirs := []string{
		path.Join(runtime.GetPath(), "conf"),
		path.Join(runtime.GetPath(), "extensions"),
		path.Join(runtime.GetPath(), "composer"),
		path.Join(runtime.GetPath(), "log"),
	}
	for _, dir := range dirs {
		_ = fileOp.ChmodR(dir, 0755, true)
		_ = fileOp.ChownR(dir, strconv.Itoa(1000), strconv.Itoa(1000), true)
	}
}

func HandleOldPHPRuntime() {
	runtimes, _ := runtimeRepo.List(repo.WithByType(constant.RuntimePHP))
	if len(runtimes) == 0 {
		return
	}
	fileOp := files.NewFileOp()
	for _, runtime := range runtimes {
		composePtah := runtime.GetComposePath()
		composeBytes, _ := fileOp.GetContent(composePtah)
		composeContent := strings.ReplaceAll(string(composeBytes), "./conf:/usr/local/etc/php", "./conf/php.ini:/usr/local/etc/php/php.ini")
		composeContent = strings.ReplaceAll(composeContent, "./conf/php-fpm.conf:/usr/local/etc/php-fpm.d/www.conf", "./conf/php-fpm.conf:/usr/local/etc/php-fpm.conf")
		composeContent = strings.ReplaceAll(composeContent, "./extensions:${EXTENSION_DIR}", "./extensions:/usr/local/lib/php/extensions")
		_ = fileOp.WriteFile(composePtah, strings.NewReader(composeContent), constant.DirPerm)
		_ = fileOp.WriteFile(runtime.GetFPMPath(), bytes.NewReader(nginx_conf.GetWebsiteFile("php-fpm.conf")), constant.DirPerm)
		supervisorConfigPath := path.Join(runtime.GetPath(), "supervisor", "supervisor.d", "php-fpm.ini")
		supervisorConfigBytes, _ := fileOp.GetContent(supervisorConfigPath)
		if !strings.Contains(string(supervisorConfigBytes), "nodaemonize") {
			newConfigContent := strings.ReplaceAll(string(supervisorConfigBytes), "command=php-fpm", "command=php-fpm --nodaemonize")
			_ = fileOp.WriteFile(supervisorConfigPath, bytes.NewReader([]byte(newConfigContent)), constant.DirPerm)
		}
		go func() {
			_ = restartRuntime(&runtime)
		}()
	}
}
