package service

import (
	"strings"

	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

func validateClusterAppInstall(key, version string, cluster bool) error {
	if !cluster {
		if constant.IsClusterApp(key) {
			return buserr.New("ErrClusterAppOnly")
		}
		return nil
	}
	base, master := strings.CutSuffix(version, "-master")
	if !master {
		base, _ = strings.CutSuffix(version, "-slave")
	}
	if !constant.IsClusterApp(key) || base == "" || base == version {
		return buserr.New("ErrInvalidParams")
	}
	return nil
}
