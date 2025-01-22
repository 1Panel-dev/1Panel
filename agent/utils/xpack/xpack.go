//go:build !xpack

package xpack

import (
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/utils/cmd"
)

func RemoveTamper(website string) {}

func LoadGpuInfo() []interface{} {
	return nil
}

func LoadXpuInfo() []interface{} {
	return nil
}

func StartClam(startClam model.Clam, isUpdate bool) (int, error) {
	return 0, buserr.New(constant.ErrXpackNotFound)
}

func LoadNodeInfo() (bool, model.NodeInfo, error) {
	var info model.NodeInfo
	info.BaseDir = loadParams("BASE_DIR")
	info.Version = loadParams("ORIGINAL_VERSION")
	info.Scope = "master"
	return false, info, nil
}

func loadParams(param string) string {
	stdout, err := cmd.Execf("grep '^%s=' /usr/local/bin/1pctl | cut -d'=' -f2", param)
	if err != nil {
		panic(err)
	}
	info := strings.ReplaceAll(stdout, "\n", "")
	if len(info) == 0 || info == `""` {
		panic(fmt.Sprintf("error `%s` find in /usr/local/bin/1pctl", param))
	}
	return info
}

func GetImagePrefix() string {
	return ""
}

func IsUseCustomApp() bool {
	return false
}

// alert
func CreateAlert(createAlert dto.CreateOrUpdateAlert) error {
	return nil
}
func UpdateAlert(updateAlert dto.CreateOrUpdateAlert) error {
	return nil
}
func DeleteAlert(alertBase dto.AlertBase) error {
	return nil
}
func GetAlert(alertBase dto.AlertBase) uint {
	return 0
}
func PushAlert(pushAlert dto.PushAlert) error {
	return nil
}
