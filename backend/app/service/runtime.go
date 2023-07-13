package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/app/dto/request"
	"github.com/1Panel-dev/1Panel/backend/app/dto/response"
	"github.com/1Panel-dev/1Panel/backend/app/model"
	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/docker"
	"github.com/1Panel-dev/1Panel/backend/utils/files"
	"github.com/subosito/gotenv"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type RuntimeService struct {
}

type IRuntimeService interface {
	Page(req request.RuntimeSearch) (int64, []response.RuntimeRes, error)
	Create(create request.RuntimeCreate) error
	Delete(id uint) error
	Update(req request.RuntimeUpdate) error
	Get(id uint) (res *response.RuntimeRes, err error)
}

func NewRuntimeService() IRuntimeService {
	return &RuntimeService{}
}

func (r *RuntimeService) Create(create request.RuntimeCreate) (err error) {
	exist, _ := runtimeRepo.GetFirst(runtimeRepo.WithName(create.Name))
	if exist != nil {
		return buserr.New(constant.ErrNameIsExist)
	}
	if create.Resource == constant.ResourceLocal {
		runtime := &model.Runtime{
			Name:     create.Name,
			Resource: create.Resource,
			Type:     create.Type,
			Version:  create.Version,
			Status:   constant.RuntimeNormal,
		}
		return runtimeRepo.Create(context.Background(), runtime)
	}
	exist, _ = runtimeRepo.GetFirst(runtimeRepo.WithImage(create.Image))
	if exist != nil {
		return buserr.New(constant.ErrImageExist)
	}
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(create.AppDetailID))
	if err != nil {
		return err
	}
	app, err := appRepo.GetFirst(commonRepo.WithByID(appDetail.AppId))
	if err != nil {
		return err
	}
	fileOp := files.NewFileOp()
	appVersionDir := path.Join(constant.AppResourceDir, app.Resource, app.Key, appDetail.Version)
	if !fileOp.Stat(appVersionDir) || appDetail.Update {
		if err := downloadApp(app, appDetail, nil); err != nil {
			return err
		}
	}
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
	newNameDir := path.Join(runtimeDir, create.Name)
	defer func() {
		if err != nil {
			_ = fileOp.DeleteDir(newNameDir)
		}
	}()
	if oldDir != newNameDir {
		if err = fileOp.Rename(oldDir, newNameDir); err != nil {
			return
		}
		if err = fileOp.DeleteDir(tempDir); err != nil {
			return
		}
	}
	composeContent, envContent, forms, err := handleParams(create.Image, create.Type, newNameDir, create.Params)
	if err != nil {
		return
	}
	composeService, err := getComposeService(create.Name, newNameDir, composeContent, envContent, false)
	if err != nil {
		return
	}
	runtime := &model.Runtime{
		Name:          create.Name,
		DockerCompose: string(composeContent),
		Env:           string(envContent),
		AppDetailID:   create.AppDetailID,
		Type:          create.Type,
		Image:         create.Image,
		Resource:      create.Resource,
		Status:        constant.RuntimeBuildIng,
		Version:       create.Version,
		Params:        string(forms),
	}
	if err = runtimeRepo.Create(context.Background(), runtime); err != nil {
		return
	}
	go buildRuntime(runtime, composeService, "")
	return
}

func (r *RuntimeService) Page(req request.RuntimeSearch) (int64, []response.RuntimeRes, error) {
	var (
		opts []repo.DBOption
		res  []response.RuntimeRes
	)
	if req.Name != "" {
		opts = append(opts, commonRepo.WithLikeName(req.Name))
	}
	if req.Status != "" {
		opts = append(opts, runtimeRepo.WithStatus(req.Status))
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

func (r *RuntimeService) Delete(id uint) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return err
	}
	website, _ := websiteRepo.GetFirst(websiteRepo.WithRuntimeID(id))
	if website.ID > 0 {
		return buserr.New(constant.ErrDelWithWebsite)
	}
	if runtime.Resource == constant.ResourceAppstore {
		client, err := docker.NewClient()
		if err != nil {
			return err
		}
		imageID, err := client.GetImageIDByName(runtime.Image)
		if err != nil {
			return err
		}
		if imageID != "" {
			if err := client.DeleteImage(imageID); err != nil {
				global.LOG.Errorf("delete image id [%s] error %v", imageID, err)
			}
		}
		runtimeDir := path.Join(constant.RuntimeDir, runtime.Type, runtime.Name)
		if err := files.NewFileOp().DeleteDir(runtimeDir); err != nil {
			return err
		}
	}
	return runtimeRepo.DeleteBy(commonRepo.WithByID(id))
}

func (r *RuntimeService) Get(id uint) (*response.RuntimeRes, error) {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(id))
	if err != nil {
		return nil, err
	}
	res := &response.RuntimeRes{}
	res.Runtime = *runtime
	if runtime.Resource == constant.ResourceLocal {
		return res, nil
	}
	appDetail, err := appDetailRepo.GetFirst(commonRepo.WithByID(runtime.AppDetailID))
	if err != nil {
		return nil, err
	}
	res.AppID = appDetail.AppId
	var (
		appForm   dto.AppForm
		appParams []response.AppParam
	)
	if err := json.Unmarshal([]byte(runtime.Params), &appForm); err != nil {
		return nil, err
	}
	envs, err := gotenv.Unmarshal(runtime.Env)
	if err != nil {
		return nil, err
	}
	for _, form := range appForm.FormFields {
		if v, ok := envs[form.EnvKey]; ok {
			appParam := response.AppParam{
				Edit:     false,
				Key:      form.EnvKey,
				Rule:     form.Rule,
				Type:     form.Type,
				Required: form.Required,
			}
			if form.Edit {
				appParam.Edit = true
			}
			appParam.LabelZh = form.LabelZh
			appParam.LabelEn = form.LabelEn
			appParam.Multiple = form.Multiple
			appParam.Value = v
			if form.Type == "select" {
				if form.Multiple {
					if v == "" {
						appParam.Value = []string{}
					} else {
						appParam.Value = strings.Split(v, ",")
					}
				} else {
					for _, fv := range form.Values {
						if fv.Value == v {
							appParam.ShowValue = fv.Label
							break
						}
					}
				}
				appParam.Values = form.Values
			}
			appParams = append(appParams, appParam)
		}
	}
	res.AppParams = appParams
	return res, nil
}

func (r *RuntimeService) Update(req request.RuntimeUpdate) error {
	runtime, err := runtimeRepo.GetFirst(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	oldImage := runtime.Image
	if runtime.Resource == constant.ResourceLocal {
		runtime.Version = req.Version
		return runtimeRepo.Save(runtime)
	}
	exist, _ := runtimeRepo.GetFirst(runtimeRepo.WithImage(req.Name), runtimeRepo.WithNotId(req.ID))
	if exist != nil {
		return buserr.New(constant.ErrImageExist)
	}
	runtimeDir := path.Join(constant.RuntimeDir, runtime.Type, runtime.Name)
	composeContent, envContent, _, err := handleParams(req.Image, runtime.Type, runtimeDir, req.Params)
	if err != nil {
		return err
	}
	composeService, err := getComposeService(runtime.Name, runtimeDir, composeContent, envContent, false)
	if err != nil {
		return err
	}
	runtime.Image = req.Image
	runtime.Env = string(envContent)
	runtime.DockerCompose = string(composeContent)
	runtime.Status = constant.RuntimeBuildIng
	_ = runtimeRepo.Save(runtime)
	client, err := docker.NewClient()
	if err != nil {
		return err
	}
	imageID, err := client.GetImageIDByName(oldImage)
	if err != nil {
		return err
	}
	go buildRuntime(runtime, composeService, imageID)
	return nil
}
