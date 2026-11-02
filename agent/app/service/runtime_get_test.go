package service

import (
	"encoding/json"
	"reflect"
	"slices"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

func TestGetPHPRuntimeRefreshesVersionOptionsWithoutChangingConfig(t *testing.T) {
	for _, resource := range []string{constant.AppResourceRemote, constant.AppResourceLocal, constant.AppResourceCustom} {
		t.Run(resource, func(t *testing.T) {
			db := appCacheTestDB(t)
			if err := db.AutoMigrate(&model.Runtime{}); err != nil {
				t.Fatal(err)
			}
			app := model.App{Key: "php", Type: constant.RuntimePHP, Resource: resource}
			if err := db.Create(&app).Error; err != nil {
				t.Fatal(err)
			}
			oldVersion := dto.AppFormValue{Label: "8.1.34", Value: "8.1.34"}
			newVersion := dto.AppFormValue{Label: "8.1.35", Value: "8.1.35"}
			savedForm := dto.AppForm{FormFields: []dto.AppFormFields{
				{EnvKey: "PHP_VERSION", Type: "select", Default: oldVersion.Value, Values: []dto.AppFormValue{oldVersion}},
				{EnvKey: "PHP_EXTENSIONS", Type: "select", Multiple: true, Values: []dto.AppFormValue{{Label: "redis", Value: "redis"}}},
				{EnvKey: "PANEL_APP_PORT_HTTP", Type: "number"},
				{EnvKey: "CONTAINER_PACKAGE_URL", Type: "select", Values: []dto.AppFormValue{{Label: "custom mirror", Value: "https://mirror.example"}}},
			}}
			encode := func(form dto.AppForm) string {
				t.Helper()
				data, err := json.Marshal(form)
				if err != nil {
					t.Fatal(err)
				}
				return string(data)
			}
			detail := model.AppDetail{AppId: app.ID, Version: "8.1", Params: encode(savedForm), Update: true}
			if err := db.Create(&detail).Error; err != nil {
				t.Fatal(err)
			}
			otherMajor := model.AppDetail{AppId: app.ID, Version: "8.2", Params: `{"formFields":[{"envKey":"PHP_VERSION","values":[{"label":"8.2.31","value":"8.2.31"}]}]}`}
			if err := db.Create(&otherMajor).Error; err != nil {
				t.Fatal(err)
			}
			runtime := model.Runtime{
				Name: "existing-php", Type: constant.RuntimePHP, Resource: constant.ResourceAppstore,
				Version: detail.Version, AppDetailID: detail.ID, Image: "custom-php:existing", Params: detail.Params,
				Env:           "PHP_VERSION=8.1.34\nPHP_EXTENSIONS=redis,custom-extension\nPANEL_APP_PORT_HTTP=9070\nCONTAINER_PACKAGE_URL=https://mirror.example\nCONTAINER_NAME=my-php\nCUSTOM_SETTING=keep-me\n",
				DockerCompose: "existing compose", Remark: "keep remark",
			}
			if err := db.Create(&runtime).Error; err != nil {
				t.Fatal(err)
			}
			service := NewRuntimeService()
			before, err := service.Get(runtime.ID)
			if err != nil {
				t.Fatal(err)
			}
			latestForm := dto.AppForm{FormFields: []dto.AppFormFields{
				{EnvKey: "PHP_VERSION", Type: "select", Default: newVersion.Value, Values: []dto.AppFormValue{newVersion, oldVersion}},
				{EnvKey: "PHP_EXTENSIONS", Type: "select", Multiple: true, Default: "opcache"},
				{EnvKey: "PANEL_APP_PORT_HTTP", Default: 9000},
			}}
			removedForm := dto.AppForm{FormFields: []dto.AppFormFields{
				{EnvKey: "PHP_VERSION", Type: "select", Default: newVersion.Value, Values: []dto.AppFormValue{newVersion}},
			}}
			for _, tc := range []struct {
				name   string
				params string
				values []dto.AppFormValue
			}{
				{"new patch on reopen", encode(latestForm), []dto.AppFormValue{newVersion, oldVersion}},
				{"selected version removed", encode(removedForm), []dto.AppFormValue{newVersion, oldVersion}},
				{"invalid metadata", "{invalid", []dto.AppFormValue{oldVersion}},
				{"missing metadata", "", []dto.AppFormValue{oldVersion}},
				{"missing version field", `{"formFields":[{"envKey":"PHP_EXTENSIONS"}]}`, []dto.AppFormValue{oldVersion}},
				{"empty version list", `{"formFields":[{"envKey":"PHP_VERSION","values":[]}]}`, []dto.AppFormValue{oldVersion}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					if err := db.Model(&detail).UpdateColumn("params", tc.params).Error; err != nil {
						t.Fatal(err)
					}
					got, err := service.Get(runtime.ID)
					if err != nil {
						t.Fatal(err)
					}
					want := *before
					want.AppParams = slices.Clone(before.AppParams)
					want.AppParams[0].Values = tc.values
					if !reflect.DeepEqual(*got, want) {
						t.Fatalf("edit response changed configuration or returned stale options:\ngot: %+v\nwant: %+v", *got, want)
					}
					var stored model.Runtime
					if err := db.First(&stored, runtime.ID).Error; err != nil {
						t.Fatal(err)
					}
					if stored.Params != runtime.Params || stored.Env != runtime.Env || stored.Image != runtime.Image || stored.DockerCompose != runtime.DockerCompose || stored.AppDetailID != runtime.AppDetailID || stored.Version != runtime.Version {
						t.Fatal("opening edit mutated the saved runtime")
					}
					var storedDetail model.AppDetail
					if err := db.First(&storedDetail, detail.ID).Error; err != nil {
						t.Fatal(err)
					}
					if !storedDetail.Update {
						t.Fatal("opening edit cleared the pending template update")
					}
				})
			}
		})
	}
}
