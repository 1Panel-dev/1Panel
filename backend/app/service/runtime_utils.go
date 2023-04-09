package service

import (
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/subosito/gotenv"
	"path"
	"strings"
)

func buildRuntime(runtime *model.Runtime, service *docker.ComposeService) {
	err := service.ComposeBuild()
	if err != nil {
		runtime.Status = constant.RuntimeError
		runtime.Message = buserr.New(constant.ErrImageBuildErr).Error() + ":" + err.Error()
	} else {
		runtime.Status = constant.RuntimeNormal
	}
	_ = runtimeRepo.Save(runtime)
}

func handleParams(image, runtimeType, runtimeDir string, params map[string]interface{}) (composeContent []byte, envContent []byte, forms []byte, err error) {
	fileOp := files.NewFileOp()
	composeContent, err = fileOp.GetContent(path.Join(runtimeDir, "docker-compose.yml"))
	if err != nil {
		return
	}
	env, err := gotenv.Read(path.Join(runtimeDir, ".env"))
	if err != nil {
		return
	}
	forms, err = fileOp.GetContent(path.Join(runtimeDir, "config.json"))
	if err != nil {
		return
	}
	params["IMAGE_NAME"] = image
	if runtimeType == constant.RuntimePHP {
		if extends, ok := params["PHP_EXTENSIONS"]; ok {
			if extendsArray, ok := extends.([]interface{}); ok {
				strArray := make([]string, len(extendsArray))
				for i, v := range extendsArray {
					strArray[i] = fmt.Sprintf("%v", v)
				}
				params["PHP_EXTENSIONS"] = strings.Join(strArray, ",")
			}
		}
	}
	newMap := make(map[string]string)
	handleMap(params, newMap)
	for k, v := range newMap {
		env[k] = v
	}
	envStr, err := gotenv.Marshal(env)
	if err != nil {
		return
	}
	if err = gotenv.Write(env, path.Join(runtimeDir, ".env")); err != nil {
		return
	}
	envContent = []byte(envStr)
	return
}

func getComposeService(name, runtimeDir string, composeFile, env []byte, skipNormalization bool) (*docker.ComposeService, error) {
	project, err := docker.GetComposeProject(name, runtimeDir, composeFile, env, skipNormalization)
	if err != nil {
		return nil, err
	}
	composeService, err := docker.NewComposeService()
	if err != nil {
		return nil, err
	}
	composeService.SetProject(project)
	return composeService, nil
}
