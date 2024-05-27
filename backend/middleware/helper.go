package middleware

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/backend/app/repo"
	"github.com/1Panel-dev/1Panel/backend/constant"
)

func LoadErrCode(errInfo string) int {
	settingRepo := repo.NewISettingRepo()
	codeVal, err := settingRepo.Get(settingRepo.WithByKey("NoAuthSetting"))
	if err != nil {
		return 500
	}

	switch codeVal.Value {
	case "400":
		return http.StatusBadRequest
	case "401":
		return http.StatusUnauthorized
	case "403":
		return http.StatusForbidden
	case "404":
		return http.StatusFound
	case "408":
		return http.StatusRequestTimeout
	case "416":
		return http.StatusRequestedRangeNotSatisfiable
	default:
		if errInfo == "err-ip" {
			return constant.CodeErrIP
		}
		if errInfo == "err-domain" {
			return constant.CodeErrDomain
		}
		if errInfo == "err-entrance" {
			return constant.CodeErrEntrance
		}
		return http.StatusOK
	}
}
