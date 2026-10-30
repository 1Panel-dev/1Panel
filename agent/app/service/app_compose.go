package service

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/req_helper"
	"gopkg.in/yaml.v3"
)

func ensureAppDetailCompose(app model.App, detail *model.AppDetail) error {
	if strings.TrimSpace(detail.DockerCompose) != "" && validateAppDetailCompose([]byte(detail.DockerCompose)) == nil {
		return nil
	}
	var content []byte
	var err error
	if app.IsLocalApp() || app.IsCustomApp() {
		content, err = os.ReadFile(filepath.Join(app.GetAppResourcePath(), detail.Version, "docker-compose.yml"))
	} else {
		var composeURL string
		if detail.DownloadUrl == "" {
			composeURL = fmt.Sprintf("%s/%s/1panel/%s/%s/docker-compose.yml", global.AppRepoURL(), global.CONF.Base.Mode, app.Key, detail.Version)
		} else {
			var source *url.URL
			source, err = url.Parse(detail.DownloadUrl)
			if err == nil {
				source.Path = path.Join(path.Dir(source.Path), "docker-compose.yml")
				source.RawPath = ""
				composeURL = source.String()
			}
		}
		if err == nil {
			composeURL, err = currentAppDownloadURL(composeURL)
		}
		if err == nil {
			var status int
			status, content, err = req_helper.HandleRequest(composeURL, http.MethodGet, constant.TimeOut20s)
			if status == http.StatusNotFound {
				return buserr.New("ErrAppVersionUnavailable")
			}
		}
	}
	if err != nil {
		return buserr.WithDetail("ErrGetCompose", err.Error(), err)
	}
	return cacheAppDetailCompose(detail, content)
}

func validateAppDetailCompose(content []byte) error {
	var compose struct {
		Services map[string]map[string]interface{} `yaml:"services"`
	}
	if err := yaml.Unmarshal(content, &compose); err != nil {
		return buserr.WithDetail("ErrGetCompose", err.Error(), err)
	}
	if len(compose.Services) == 0 {
		return buserr.WithDetail("ErrGetCompose", "docker-compose.yml has no services", nil)
	}
	for name, service := range compose.Services {
		if strings.TrimSpace(name) == "" || len(service) == 0 {
			return buserr.WithDetail("ErrGetCompose", "docker-compose.yml has an empty service", nil)
		}
	}
	return nil
}

func cacheAppDetailCompose(detail *model.AppDetail, content []byte) error {
	if err := validateAppDetailCompose(content); err != nil {
		return err
	}
	stored, err := appDetailRepo.CacheCompose(*detail, string(content))
	if err != nil {
		return buserr.WithDetail("ErrGetCompose", err.Error(), err)
	}
	if !stored {
		return buserr.WithDetail("ErrGetCompose", "application resources changed; retry the operation", nil)
	}
	detail.DockerCompose = string(content)
	return nil
}
