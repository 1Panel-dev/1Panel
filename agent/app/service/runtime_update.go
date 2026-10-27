package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/repo"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
	"gopkg.in/yaml.v3"
)

// Keep the on-disk configuration as well as the model: files may contain edits
// that have not been reflected in the database.
type runtimeUpdateSnapshot struct {
	runtime model.Runtime
	files   []runtimeUpdateFile
}

type runtimeUpdateFile struct {
	path    string
	content []byte
	mode    os.FileMode
	exists  bool
}

func snapshotRuntimeUpdate(runtime *model.Runtime) (*runtimeUpdateSnapshot, error) {
	snapshot := &runtimeUpdateSnapshot{runtime: *runtime}
	for _, name := range []string{"docker-compose.yml", ".env", "run.sh"} {
		file := runtimeUpdateFile{path: filepath.Join(runtime.GetPath(), name)}
		info, err := os.Stat(file.path)
		if errors.Is(err, os.ErrNotExist) {
			snapshot.files = append(snapshot.files, file)
			continue
		}
		if err != nil {
			return nil, err
		}
		file.content, err = os.ReadFile(file.path)
		if err != nil {
			return nil, err
		}
		file.exists, file.mode = true, info.Mode().Perm()
		snapshot.files = append(snapshot.files, file)
	}
	return snapshot, nil
}

func (s *runtimeUpdateSnapshot) restoreFiles() error {
	var result error
	for _, file := range s.files {
		var err error
		if file.exists {
			err = os.WriteFile(file.path, file.content, file.mode)
			if err == nil {
				err = os.Chmod(file.path, file.mode)
			}
		} else {
			err = os.Remove(file.path)
			if errors.Is(err, os.ErrNotExist) {
				err = nil
			}
		}
		if err != nil {
			result = errors.Join(result, fmt.Errorf("restore %s: %w", file.path, err))
		}
	}
	return result
}

// This must finish before reCreateRuntime stops the existing container. A
// failed preparation restores the previous version without touching Docker.
func prepareRuntimeUpdate(runtime *model.Runtime, snapshot *runtimeUpdateSnapshot) error {
	if err := pullRuntimeComposeImages(runtime.GetComposePath(), false, nil); err != nil {
		result := fmt.Errorf("prepare runtime images: %w", err)
		restoreErr := snapshot.restoreFiles()
		*runtime = snapshot.runtime
		if restoreErr != nil {
			result = errors.Join(result, restoreErr)
			runtime.Status = constant.StatusError
		}
		runtime.Message = result.Error()
		if saveErr := runtimeRepo.Save(runtime); saveErr != nil {
			result = errors.Join(result, fmt.Errorf("restore runtime record: %w", saveErr))
		}
		logFile, logErr := os.OpenFile(runtime.GetLogPath(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, constant.FilePerm)
		if logErr == nil {
			_, logErr = fmt.Fprintln(logFile, result.Error())
			logErr = errors.Join(logErr, logFile.Close())
		}
		return errors.Join(result, logErr)
	}
	return nil
}

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
