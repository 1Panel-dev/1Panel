package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	BatchDelete(ids []uint) error
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

	fileSetting, err := settingRepo.Get(settingRepo.WithByKey("DaemonJsonPath"))
	if err != nil {
		return err
	}
	if len(fileSetting.Value) == 0 {
		return errors.New("error daemon.json")
	}
	if _, err := os.Stat(fileSetting.Value); err != nil && os.IsNotExist(err) {
		if err = os.MkdirAll(fileSetting.Value, os.ModePerm); err != nil {
			if err != nil {
				return err
			}
		}
	}
	if req.Protocol == "http" {
		file, err := ioutil.ReadFile(fileSetting.Value)
		if err != nil {
			return err
		}

		deamonMap := make(map[string]interface{})
		if err := json.Unmarshal(file, &deamonMap); err != nil {
			return err
		}
		if _, ok := deamonMap["insecure-registries"]; ok {
			if k, v := deamonMap["insecure-registries"].([]interface{}); v {
				deamonMap["insecure-registries"] = common.RemoveRepeatElement(append(k, req.DownloadUrl))
			}
		} else {
			deamonMap["insecure-registries"] = []string{req.DownloadUrl}
		}
		newJson, err := json.MarshalIndent(deamonMap, "", "\t")
		if err != nil {
			return err
		}
		if err := ioutil.WriteFile(fileSetting.Value, newJson, 0640); err != nil {
			return err
		}
	}

	if err := copier.Copy(&imageRepo, &req); err != nil {
		return errors.WithMessage(constant.ErrStructTransform, err.Error())
	}
	imageRepo.Status = constant.StatusSuccess
	if err := u.checkConn(req.DownloadUrl, req.Username, req.Password); err != nil {
		imageRepo.Status = constant.StatusFailed
		imageRepo.Message = err.Error()
	}
	if err := imageRepoRepo.Create(&imageRepo); err != nil {
		return err
	}

	cmd := exec.Command("systemctl", "restart", "docker")
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(stdout))
	}
	return nil
}

type DeamonJson struct {
	InsecureRegistries []string `json:"insecure-registries"`
}

func (u *ImageRepoService) BatchDelete(ids []uint) error {
	for _, id := range ids {
		if id == 1 {
			return errors.New("The default value cannot be edit !")
		}
	}
	repos, err := imageRepoRepo.List(commonRepo.WithIdsIn(ids))
	if err != nil {
		return err
	}
	fileSetting, err := settingRepo.Get(settingRepo.WithByKey("DaemonJsonPath"))
	if err != nil {
		return err
	}
	if len(fileSetting.Value) == 0 {
		return errors.New("error daemon.json")
	}

	deamonMap := make(map[string]interface{})
	file, err := ioutil.ReadFile(fileSetting.Value)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(file, &deamonMap); err != nil {
		return err
	}
	iRegistries := deamonMap["insecure-registries"]
	registries, _ := iRegistries.([]string)
	if len(registries) != 0 {
		for _, repo := range repos {
			if repo.Protocol == "http" {
				for i, regi := range registries {
					if regi == repo.DownloadUrl {
						registries = append(registries[:i], registries[i+1:]...)
					}
				}
			}
		}
	}
	if len(registries) == 0 {
		delete(deamonMap, "insecure-registries")
	}
	newJson, err := json.MarshalIndent(deamonMap, "", "\t")
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(fileSetting.Value, newJson, 0640); err != nil {
		return err
	}

	for _, repo := range repos {
		if repo.Auth {
			cmd := exec.Command("docker", "logout", fmt.Sprintf("%s://%s", repo.Protocol, repo.DownloadUrl))
			_, _ = cmd.CombinedOutput()
		}
	}
	if err := imageRepoRepo.Delete(commonRepo.WithIdsIn(ids)); err != nil {
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
