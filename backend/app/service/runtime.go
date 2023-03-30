package service

import (
	"context"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/subosito/gotenv"
	"path"
)

type RuntimeService struct {
}

type IRuntimeService interface {
	Page(req request.RuntimeSearch) (int64, []response.RuntimeRes, error)
	Create(create request.RuntimeCreate) error
}

func NewRuntimeService() IRuntimeService {
	return &RuntimeService{}
}

func (r *RuntimeService) Create(create request.RuntimeCreate) error {
	if create.Resource == constant.ResourceLocal {
		runtime := &model.Runtime{
			Name:     create.Name,
			Resource: create.Resource,
			Type:     create.Type,
			Status:   constant.RuntimeNormal,
		}
		return runtimeRepo.Create(context.Background(), runtime)
	}
	var err error
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(create.AppDetailID))
	if err != nil {
		return err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	buildDir := path.Join(constant.AppResourceDir, app.Key, "versions", appDetail.Version, "build")
	if !fileOp.Stat(buildDir) {
		return buserr.New(constant.ErrDirNotFound)
	}
	tempDir := path.Join(constant.RuntimeDir, app.Key)
	if err := fileOp.CopyDir(buildDir, tempDir); err != nil {
		return err
	}
	oldDir := path.Join(tempDir, "build")
	newNameDir := path.Join(tempDir, create.Name)
	defer func(defErr *error) {
		if defErr != nil {
			_ = fileOp.DeleteDir(newNameDir)
		}
	}(&err)
	if oldDir != newNameDir {
		if err := fileOp.Rename(oldDir, newNameDir); err != nil {
			return err
		}
	}
	composeFile, err := fileOp.GetContent(path.Join(newNameDir, "docker-compose.yml"))
	if err != nil {
		return err
	}
	env, err := gotenv.Read(path.Join(newNameDir, ".env"))
	if err != nil {
		return err
	}
	newMap := make(map[string]string)
	handleMap(create.Params, newMap)
	for k, v := range newMap {
		env[k] = v
	}
	envStr, err := gotenv.Marshal(env)
	if err != nil {
		return err
	}
	if err := gotenv.Write(env, path.Join(newNameDir, ".env")); err != nil {
		return err
	}
	project, err := docker.GetComposeProject(create.Name, newNameDir, composeFile, []byte(envStr))
	if err != nil {
		return err
	}
	composeService, err := docker.NewComposeService()
	if err != nil {
		return err
	}
	composeService.SetProject(project)
	if err := composeService.ComposeBuild(); err != nil {
		return err
	}
	runtime := &model.Runtime{
		Name:          create.Name,
		DockerCompose: string(composeFile),
		Env:           envStr,
		AppDetailID:   create.AppDetailID,
		Type:          create.Type,
		Image:         create.Image,
		Resource:      create.Resource,
		Status:        constant.RuntimeNormal,
	}
	return runtimeRepo.Create(context.Background(), runtime)
}

func (r *RuntimeService) Page(req request.RuntimeSearch) (int64, []response.RuntimeRes, error) {
	var (
		opts []repo.DBOption
		res  []response.RuntimeRes
	)
	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}
	total, runtimes, err := runtimeRepo.Page(req.Page, req.PageSize, opts...)
	if err != nil {
		return 0, nil, err
	}
	for _, runtime := range runtimes {
		res = append(res, response.RuntimeRes{
			Runtime: runtime,
		})
	}
	return total, res, nil
}
