package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/buserr"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/global"
	"github.com/1Panel-dev/1Panel/backend/utils/cmd"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type ImageRepoService struct{}

type IImageRepoService interface {
	Page(search dto.SearchWithPage) (int64, interface{}, error)
	List() ([]dto.ImageRepoOption, error)
	Login(req dto.OperateByID) error
	Create(req dto.ImageRepoCreate) error
	Update(req dto.ImageRepoUpdate) error
	BatchDelete(req dto.ImageRepoDelete) error
}

func NewIImageRepoService() IImageRepoService {
	return &ImageRepoService{}
}

func (u *ImageRepoService) Page(req dto.SearchWithPage) (int64, interface{}, error) {
	total, ops, err := imageRepoRepo.Page(req.Page, req.PageSize, commonRepo.WithLikeName(req.Info), commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.ImageRepoInfo
	for _, op := range ops {
		var item dto.ImageRepoInfo
		if err := copier.Copy(&item, &op); err != nil {
			return 0, nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
		}
		dtoOps = append(dtoOps, item)
	}
	return total, dtoOps, err
}

func (u *ImageRepoService) Login(req dto.OperateByID) error {
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if repo.Auth {
		if err := u.CheckConn(repo.DownloadUrl, repo.Username, repo.Password); err != nil {
			_ = imageRepoRepo.Update(repo.ID, map[string]interface{}{"status": constant.StatusFailed, "message": err.Error()})
			return err
		}
	}
	_ = imageRepoRepo.Update(repo.ID, map[string]interface{}{"status": constant.StatusSuccess})
	return nil
}

func (u *ImageRepoService) List() ([]dto.ImageRepoOption, error) {
	ops, err := imageRepoRepo.List(commonRepo.WithOrderBy("created_at desc"))
	var dtoOps []dto.ImageRepoOption
	for _, op := range ops {
		if op.Status == constant.StatusSuccess {
			var item dto.ImageRepoOption
			if err := copier.Copy(&item, &op); err != nil {
				return nil, errors.WithMessage(constant.ErrStructTransform, err.Error())
			}
			dtoOps = append(dtoOps, item)
		}
	}
	return dtoOps, err
}

func (u *ImageRepoService) Create(req dto.ImageRepoCreate) error {
	if cmd.CheckIllegal(req.Username, req.Password, req.DownloadUrl) {
		return buserr.New(constant.ErrRepoConn)
	}
	imageRepo, _ := imageRepoRepo.Get(commonRepo.WithByName(req.Name))
	if imageRepo.ID != 0 {
		return constant.ErrRecordExist
	}
	if req.Protocol == "http" {
		_ = u.handleRegistries(req.DownloadUrl, "", "create")
		stdout, err := cmd.Exec("systemctl restart docker")
		if err != nil {
			return errors.New(string(stdout))
		}
		ticker := time.NewTicker(3 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
		if err := func() error {
			for range ticker.C {
				select {
				case <-ctx.Done():
					cancel()
					return errors.New("the docker service cannot be restarted")
				default:
					stdout, err := cmd.Exec("systemctl is-active docker")
					if string(stdout) == "active\n" && err == nil {
						global.LOG.Info("docker restart with new conf successful!")
						return nil
					}
				}
			}
			return nil
		}(); err != nil {
			return err
		}
	}

	if err := copier.Copy(&imageRepo, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	imageRepo.Status = constant.StatusSuccess
	if req.Auth {
		if err := u.CheckConn(req.DownloadUrl, req.Username, req.Password); err != nil {
			imageRepo.Status = constant.StatusFailed
			imageRepo.Message = err.Error()
		}
	}
	if err := imageRepoRepo.Create(&imageRepo); err != nil {
		return err
	}

	return nil
}

func (u *ImageRepoService) BatchDelete(req dto.ImageRepoDelete) error {
	for _, id := range req.Ids {
		if id == 1 {
			return errors.New("The default value cannot be edit !")
		}
	}
	if err := imageRepoRepo.Delete(commonRepo.WithIdsIn(req.Ids)); err != nil {
		return err
	}
	return nil
}

func (u *ImageRepoService) Update(req dto.ImageRepoUpdate) error {
	if req.ID == 1 {
		return errors.New("The default value cannot be deleted !")
	}
	if cmd.CheckIllegal(req.Username, req.Password, req.DownloadUrl) {
		return buserr.New(constant.ErrRepoConn)
	}
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if repo.DownloadUrl != req.DownloadUrl || (!repo.Auth && req.Auth) {
		_ = u.handleRegistries(req.DownloadUrl, repo.DownloadUrl, "update")
		if repo.Auth {
			_, _ = cmd.ExecWithCheck("docker", "logout", repo.DownloadUrl)
		}
		stdout, err := cmd.Exec("systemctl restart docker")
		if err != nil {
			return errors.New(string(stdout))
		}
	}

	upMap := make(map[string]interface{})
	upMap["download_url"] = req.DownloadUrl
	upMap["protocol"] = req.Protocol
	upMap["username"] = req.Username
	upMap["password"] = req.Password
	upMap["auth"] = req.Auth

	upMap["status"] = constant.StatusSuccess
	upMap["message"] = ""
	if req.Auth {
		if err := u.CheckConn(req.DownloadUrl, req.Username, req.Password); err != nil {
			upMap["status"] = constant.StatusFailed
			upMap["message"] = err.Error()
		}
	}
	return imageRepoRepo.Update(req.ID, upMap)
}

func (u *ImageRepoService) CheckConn(host, user, password string) error {
	stdout, err := cmd.ExecWithCheck("docker", "login", "-u", user, "-p", password, host)
	if err != nil {
		return fmt.Errorf("stdout: %s, stderr: %v", stdout, err)
	}
	if strings.Contains(string(stdout), "Login Succeeded") {
		return nil
	}
	return errors.New(string(stdout))
}

func (u *ImageRepoService) handleRegistries(newHost, delHost, handle string) error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			return err
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}

	daemonMap := make(map[string]interface{})
	file, err := os.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(file, &daemonMap); err != nil {
		return err
	}

	iRegistries := daemonMap["insecure-registries"]
	registries, _ := iRegistries.([]interface{})
	switch handle {
	case "create":
		registries = common.RemoveRepeatElement(append(registries, newHost))
	case "update":
		registries = common.RemoveRepeatElement(append(registries, newHost))
		for i, regi := range registries {
			if regi == delHost {
				registries = append(registries[:i], registries[i+1:]...)
			}
		}
	case "delete":
		for i, regi := range registries {
			if regi == delHost {
				registries = append(registries[:i], registries[i+1:]...)
			}
		}
	}
	if len(registries) == 0 {
		delete(daemonMap, "insecure-registries")
	} else {
		daemonMap["insecure-registries"] = registries
	}
	newJson, err := json.MarshalIndent(daemonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := os.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}
	return nil
}
