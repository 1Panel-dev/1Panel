package middleware

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/core/app/repo"
)

func LoadErrCode(errInfo string) int {
	settingRepo := repo.NewISettingRepo()
	commonRepo := repo.NewICommonRepo()
	codeVal, err := settingRepo.Get(commonRepo.WithByKey("NoAuthSetting"))
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
		return http.StatusNotFound
	case "408":
		return http.StatusRequestTimeout
	case "416":
		return http.StatusRequestedRangeNotSatisfiable
	default:
		return http.StatusOK
	}
}
