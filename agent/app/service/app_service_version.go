package service

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/utils/common"
)

var appServiceNumericVersion = regexp.MustCompile(`^[vV]?[0-9]+(?:\.[0-9]+)*`)

func appServiceVersionAllowed(version, minimum string) bool {
	if minimum == "" {
		return true
	}
	version = appServiceNumericVersion.FindString(strings.TrimSpace(version))
	minimum = appServiceNumericVersion.FindString(strings.TrimSpace(minimum))
	return version != "" && minimum != "" && common.CompareAppVersion(version, minimum)
}

// Match installation validation to the version-filtered services offered by the API.
// Unconstrained fields keep the existing installation behavior.
func validateAppServiceVersions(formJSON string, params map[string]interface{}) error {
	if strings.TrimSpace(formJSON) == "" {
		return nil
	}
	var form struct {
		FormFields []json.RawMessage `json:"formFields"`
	}
	if err := json.Unmarshal([]byte(formJSON), &form); err != nil {
		return err
	}
	for _, rawField := range form.FormFields {
		var field struct {
			Type       string          `json:"type"`
			Required   bool            `json:"required"`
			Key        string          `json:"key"`
			EnvKey     string          `json:"envKey"`
			MinVersion string          `json:"minVersion"`
			Values     json.RawMessage `json:"values"`
			Child      struct {
				EnvKey     string `json:"envKey"`
				Required   bool   `json:"required"`
				MinVersion string `json:"minVersion"`
			} `json:"child"`
		}
		if err := json.Unmarshal(rawField, &field); err != nil {
			return err
		}
		if field.Type == "apps" && field.MinVersion == "" {
			field.MinVersion = field.Child.MinVersion
		}
		if field.MinVersion == "" || (field.Type != "service" && field.Type != "apps") {
			continue
		}
		key, envKey := field.Key, field.EnvKey
		if field.Type == "apps" {
			key, _ = params[field.EnvKey].(string)
			envKey = field.Child.EnvKey
		}
		selected, _ := params[envKey].(string)
		if selected == "" && !field.Required && !field.Child.Required {
			continue
		}
		if field.Type == "apps" {
			var values []dto.AppFormValue
			if err := json.Unmarshal(field.Values, &values); err != nil {
				return err
			}
			allowed := false
			for _, value := range values {
				if value.Value == key {
					allowed = true
					break
				}
			}
			if !allowed {
				return buserr.WithName("ErrAppParamKey", field.EnvKey)
			}
		}
		if key == "" || selected == "" {
			return buserr.WithName("ErrAppParamKey", envKey)
		}
		services, err := (&AppInstallService{}).GetServices(key, field.MinVersion)
		if err != nil {
			return err
		}
		allowed := false
		for _, service := range services {
			if service.Value == selected {
				allowed = true
				break
			}
		}
		if !allowed {
			return buserr.WithName("ErrAppParamKey", envKey)
		}
	}
	return nil
}
