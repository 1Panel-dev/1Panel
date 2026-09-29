package alert

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

const (
	CronJobAlertFailed  = "failed"
	CronJobAlertSuccess = "success"
	CronJobAlertBoth    = "both"
)

func cronJobAlertParams(advanced string) (map[string]json.RawMessage, error) {
	params := make(map[string]json.RawMessage)
	if strings.TrimSpace(advanced) != "" {
		if err := json.Unmarshal([]byte(advanced), &params); err != nil {
			return nil, fmt.Errorf("invalid cronjob alert advanced parameters: %w", err)
		}
	}
	if params == nil {
		params = make(map[string]json.RawMessage)
	}
	return params, nil
}

func CronJobAlertTriggerMode(advanced string) (string, error) {
	params, err := cronJobAlertParams(advanced)
	if err != nil {
		return CronJobAlertFailed, err
	}
	mode := CronJobAlertFailed
	if raw, ok := params["alertTriggerMode"]; ok {
		if err := json.Unmarshal(raw, &mode); err != nil {
			return CronJobAlertFailed, fmt.Errorf("invalid cronjob alert trigger mode: %w", err)
		}
	}
	switch mode {
	case CronJobAlertFailed, CronJobAlertSuccess, CronJobAlertBoth:
		return mode, nil
	default:
		return CronJobAlertFailed, fmt.Errorf("invalid cronjob alert trigger mode %q", mode)
	}
}

func MergeCronJobAlertParams(previous, incoming string) (string, error) {
	params, err := cronJobAlertParams(previous)
	if err != nil {
		return "", err
	}
	updates, err := cronJobAlertParams(incoming)
	if err != nil {
		return "", err
	}
	for key, value := range updates {
		params[key] = value
	}
	if _, ok := params["alertTriggerMode"]; !ok {
		params["alertTriggerMode"] = json.RawMessage(`"failed"`)
	}
	data, err := json.Marshal(params)
	if err != nil {
		return "", err
	}
	if _, err := CronJobAlertTriggerMode(string(data)); err != nil {
		return "", err
	}
	return string(data), nil
}

func MatchCronJobAlertResult(advanced, result string) bool {
	mode, err := CronJobAlertTriggerMode(advanced)
	if err != nil {
		return false
	}
	if result == "" {
		result = CronJobAlertFailed
	}
	if result != CronJobAlertFailed && result != CronJobAlertSuccess {
		return false
	}
	return mode == CronJobAlertBoth || mode == result
}

func CreateTaskAlertParams(pushAlert dto.PushAlert) []dto.Param {
	params := CreateAlertParams(GetCronJobTypeName(pushAlert.Param))
	if GetCronJobType(pushAlert.AlertType) != "cronJob" {
		return params
	}
	params = append(params, CreateCronJobResultParam(pushAlert.Result))
	return params
}

func cronJobTaskName(project string, params []dto.Param) string {
	if project != "" {
		return project
	}
	if name := getValueByIndex(params, "taskName"); name != "" {
		return name
	}
	return getValueByIndex(params, "1")
}

func CreateCronJobResultParam(result string) dto.Param {
	value := "失败"
	if result == CronJobAlertSuccess {
		value = "成功"
	}
	return dto.Param{Index: "2", Key: "result", Value: value}
}

func CronJobAlertResultFromParams(params []dto.Param) string {
	for _, param := range params {
		if param.Index == "result" {
			if param.Value == CronJobAlertSuccess {
				return CronJobAlertSuccess
			}
			return CronJobAlertFailed
		}
	}
	for _, param := range params {
		if param.Index == "2" && param.Key == "result" && param.Value == "成功" {
			return CronJobAlertSuccess
		}
	}
	return CronJobAlertFailed
}
