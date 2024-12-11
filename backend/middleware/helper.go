package middleware

import (
	"net/http"

	"github.com/1Panel-dev/1Panel/backend/app/repo"
)

func LoadErrCode() int {
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
		return http.StatusNotFound
	case "408":
		return http.StatusRequestTimeout
	case "416":
		return http.StatusRequestedRangeNotSatisfiable
	case "500":
		return http.StatusInternalServerError
	case "444":
		return 444
	default:
		return http.StatusOK
	}
}
