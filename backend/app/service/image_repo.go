package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/1Panel-dev/1Panel/backend/app/dto"
	"github.com/1Panel-dev/1Panel/backend/constant"
	"github.com/1Panel-dev/1Panel/backend/utils/common"
	"github.com/jinzhu/copier"
	"github.com/pkg/errors"
)

type ImageRepoService struct{}

type IImageRepoService interface {
	Page(search dto.PageInfo) (int64, interface{}, error)
	List() ([]dto.ImageRepoOption, error)
	Create(req dto.ImageRepoCreate) error
	Update(req dto.ImageRepoUpdate) error
	BatchDelete(req dto.ImageRepoDelete) error
}

func NewIImageRepoService() IImageRepoService {
	return &ImageRepoService{}
}

func (u *ImageRepoService) Page(search dto.PageInfo) (int64, interface{}, error) {
	total, ops, err := imageRepoRepo.Page(search.Page, search.PageSize, commonRepo.WithOrderBy("created_at desc"))
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
	imageRepo, _ := imageRepoRepo.Get(commonRepo.WithByName(req.Name))
	if imageRepo.ID != 0 {
		return constant.ErrRecordExist
	}
	if req.Protocol == "http" {
		_ = u.handleRegistries(req.DownloadUrl, "", "create")
	}

	if err := copier.Copy(&imageRepo, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}

	cmd := exec.Command("systemctl", "restart", "docker")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}

	imageRepo.Status = constant.StatusSuccess
	if err := u.checkConn(req.DownloadUrl, req.Username, req.Password); err != nil {
		imageRepo.Status = constant.StatusFailed
		imageRepo.Message = err.Error()
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
	if !req.DeleteInsecure {
		if err := imageRepoRepo.Delete(commonRepo.WithIdsIn(req.Ids)); err != nil {
			return err
		}
		return nil
	}
	repos, err := imageRepoRepo.List(commonRepo.WithIdsIn(req.Ids))
	if err != nil {
		return err
	}
	for _, repo := range repos {
		if repo.Protocol == "http" {
			_ = u.handleRegistries("", repo.DownloadUrl, "delete")
		}
		if repo.Auth {
			cmd := exec.Command("docker", "logout", fmt.Sprintf("%s://%s", repo.Protocol, repo.DownloadUrl))
			_, _ = cmd.CombinedOutput()
		}
	}
	if err := imageRepoRepo.Delete(commonRepo.WithIdsIn(req.Ids)); err != nil {
		return err
	}
	cmd := exec.Command("systemctl", "restart", "docker")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}

	return nil
}

func (u *ImageRepoService) Update(req dto.ImageRepoUpdate) error {
	if req.ID == 1 {
		return errors.New("The default value cannot be deleted !")
	}
	repo, err := imageRepoRepo.Get(commonRepo.WithByID(req.ID))
	if err != nil {
		return err
	}
	if repo.DownloadUrl != req.DownloadUrl {
		_ = u.handleRegistries(req.DownloadUrl, repo.DownloadUrl, "update")
		if repo.Auth {
			cmd := exec.Command("docker", "logout", repo.DownloadUrl)
			_, _ = cmd.CombinedOutput()
		}
		cmd := exec.Command("systemctl", "restart", "docker")
		stdout, err := cmd.CombinedOutput()
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
	if err := u.checkConn(req.DownloadUrl, req.Username, req.Password); err != nil {
		upMap["status"] = constant.StatusFailed
		upMap["message"] = err.Error()
	}
	return imageRepoRepo.Update(req.ID, upMap)
}

func (u *ImageRepoService) checkConn(host, user, password string) error {
	cmd := exec.Command("docker", "login", "-u", user, "-p", password, host)
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	if strings.Contains(string(stdout), "Login Succeeded") {
		return nil
	}
	return errors.New(string(stdout))
}

func (u *ImageRepoService) handleRegistries(newHost, delHost, handle string) error {
	if _, err := os.Stat(constant.DaemonJsonPath); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(path.Dir(constant.DaemonJsonPath), os.ModePerm); err != nil {
			if err != nil {
				return err
			}
		}
		_, _ = os.Create(constant.DaemonJsonPath)
	}

	deamonMap := make(map[string]interface{})
	file, err := ioutil.ReadFile(constant.DaemonJsonPath)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		return err
	}

	iRegistries := deamonMap["insecure-registries"]
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
		delete(deamonMap, "insecure-registries")
	} else {
		deamonMap["insecure-registries"] = registries
	}
	newJson, err := json.MarshalIndent(deamonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(constant.DaemonJsonPath, newJson, 0640); err != nil {
		return err
	}
	return nil
}
