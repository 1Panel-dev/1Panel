package service

import (
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"gopkg.in/yaml.v3"
)

func getRuntimeUpdateAppDetail(runtime *model.Runtime, req request.RuntimeUpdate) (model.AppDetail, error) {
	current, err := appDetailRepo.GetFirst(repo.WithByID(runtime.AppDetailID))
	if err != nil {
		return model.AppDetail{}, err
	}
	if current.ID == 0 {
		return model.AppDetail{}, buserr.New("ErrRecordNotFound")
	}
	opts := []repo.DBOption{appDetailRepo.WithAppId(current.AppId)}
	if req.AppDetailID != 0 {
		opts = append(opts, repo.WithByID(req.AppDetailID))
	} else if req.Version != "" {
		opts = append(opts, appDetailRepo.WithVersion(req.Version))
	} else {
		return current, nil
	}
	detail, err := appDetailRepo.GetFirst(opts...)
	if err != nil {
		return model.AppDetail{}, err
	}
	if detail.ID == 0 || (req.Version != "" && detail.Version != req.Version) {
		return model.AppDetail{}, buserr.New("ErrInvalidParams")
	}
	return detail, nil
}

func updateRuntimeVersionFiles(runtime *model.Runtime, appVersionDir string) error {
	fileOp := files.NewFileOp()
	template, err := fileOp.GetContent(filepath.Join(appVersionDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	current, err := fileOp.GetContent(runtime.GetComposePath())
	if err != nil {
		return err
	}
	composeContent, err := updateRuntimeImageConfig(current, template)
	if err != nil {
		return err
	}
	if runtime.Type == constant.RuntimePHP {
		if err = fileOp.CopyDir(filepath.Join(appVersionDir, "build"), runtime.GetPath()); err != nil {
			return err
		}
		if err = fileOp.CopyFile(filepath.Join(appVersionDir, "data.yml"), runtime.GetPath()); err != nil {
			return err
		}
	} else {
		if err = fileOp.CopyFile(filepath.Join(appVersionDir, "run.sh"), runtime.GetPath()); err != nil {
			return err
		}
	}
	return fileOp.SaveFile(runtime.GetComposePath(), string(composeContent), constant.FilePerm)
}

func updateRuntimeImageConfig(current, template []byte) ([]byte, error) {
	var currentCompose, templateCompose map[string]interface{}
	if err := yaml.Unmarshal(current, &currentCompose); err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(template, &templateCompose); err != nil {
		return nil, err
	}
	currentServices, ok := currentCompose["services"].(map[string]interface{})
	if !ok || len(currentServices) != 1 {
		return nil, buserr.New("ErrFileParse")
	}
	templateServices, ok := templateCompose["services"].(map[string]interface{})
	if !ok || len(templateServices) != 1 {
		return nil, buserr.New("ErrFileParse")
	}
	// Keep container settings and mounts; only the image and build definition belong to the version.
	for _, currentService := range currentServices {
		service, ok := currentService.(map[string]interface{})
		if !ok {
			return nil, buserr.New("ErrFileParse")
		}
		for _, templateService := range templateServices {
			target, ok := templateService.(map[string]interface{})
			if !ok || (target["image"] == nil && target["build"] == nil) {
				return nil, buserr.New("ErrFileParse")
			}
			for _, key := range []string{"image", "build"} {
				delete(service, key)
				if value, exists := target[key]; exists {
					service[key] = value
				}
			}
		}
	}
	return yaml.Marshal(currentCompose)
}
