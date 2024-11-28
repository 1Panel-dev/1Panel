package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/pkg/errors"
	"github.com/subosito/gotenv"
	"gopkg.in/yaml.v3"
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
	buildDir := path.Join(appVersionDir, "build")
	if !fileOp.Stat(buildDir) {
		return buserr.New(constant.ErrDirNotFound)
	}
	runtimeDir := path.Join(constant.RuntimeDir, create.Type)
	tempDir := filepath.Join(runtimeDir, fmt.Sprintf("%d", time.Now().UnixNano()))
	if err = fileOp.CopyDir(buildDir, tempDir); err != nil {
		return
	}
	oldDir := path.Join(tempDir, "build")
	projectDir := path.Join(runtimeDir, create.Name)
	defer func() {
		if err != nil {
			_ = fileOp.DeleteDir(projectDir)
		}
	}()
	if oldDir != projectDir {
		if err = fileOp.Rename(oldDir, projectDir); err != nil {
			return
		}
		if err = fileOp.DeleteDir(tempDir); err != nil {
			return
		}
	}
	composeContent, envContent, forms, err := handleParams(create, projectDir)
	if err != nil {
		return
	}
	runtime.DockerCompose = string(composeContent)
	runtime.Env = string(envContent)
	runtime.Params = string(forms)
	runtime.Status = constant.RuntimeBuildIng

	go buildRuntime(runtime, "", false)
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
	cmd := exec.Command("docker-compose", "-f", composePath, operate)
	if operate == "up" {
		cmd = exec.Command("docker-compose", "-f", composePath, operate, "-d")
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
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
		return errors.New(buserr.New(constant.ErrRuntimeStart).Error() + ":" + stderrBuf.String())
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

func buildRuntime(runtime *model.Runtime, oldImageID string, rebuild bool) {
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
	multiWriterStdout := io.MultiWriter(os.Stdout, logFile)
	cmd.Stdout = multiWriterStdout
	var stderrBuf bytes.Buffer
	multiWriterStderr := io.MultiWriter(&stderrBuf, logFile, os.Stderr)
	cmd.Stderr = multiWriterStderr

	err = cmd.Run()
	if err != nil {
		runtime.Status = constant.RuntimeError
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + buserr.New("ErrCmdTimeout").Error()
		} else {
			runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + stderrBuf.String()
		}
	} else {
		runtime.Status = constant.RuntimeNormal
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
			websites, _ := websiteRepo.GetBy(websiteRepo.WithRuntimeID(runtime.ID))
			if len(websites) > 0 {
				installService := NewIAppInstalledService()
				installMap := make(map[uint]string)
				for _, website := range websites {
					if website.AppInstallID > 0 {
						installMap[website.AppInstallID] = website.PrimaryDomain
					}
				}
				for installID, domain := range installMap {
					go func(installID uint, domain string) {
						global.LOG.Infof("rebuild php runtime [%s] domain [%s]", runtime.Name, domain)
						if err := installService.Operate(request.AppInstalledOperate{
							InstallId: installID,
							Operate:   constant.Rebuild,
						}); err != nil {
							global.LOG.Errorf("rebuild php runtime [%s] domain [%s] error %v", runtime.Name, domain, err)
						}
					}(installID, domain)
				}
			}
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
	switch create.Type {
	case constant.RuntimePHP:
		create.Params["IMAGE_NAME"] = create.Image
		forms, err = fileOp.GetContent(path.Join(projectDir, "config.json"))
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

		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeJava:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["JAVA_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
		composeContent, err = handleCompose(env, composeContent, create, projectDir)
		if err != nil {
			return
		}
	case constant.RuntimeGo:
		create.Params["CODE_DIR"] = create.CodeDir
		create.Params["GO_VERSION"] = create.Version
		create.Params["PANEL_APP_PORT_HTTP"] = create.Port
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
