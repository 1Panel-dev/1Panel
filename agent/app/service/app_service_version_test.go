package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"gopkg.in/yaml.v3"
)

func TestAppServiceMinimumVersionComparison(t *testing.T) {
	for _, tc := range []struct {
		version, minimum string
		want             bool
	}{
		{"16.6", "17", false},
		{"17", "17", true},
		{"17.0", "17", true},
		{"18.1", "17", true},
		{"17.x", "17", true},
		{"v17.2-alpine", "17.1", true},
		{"17-alpine99", "17.1", false},
		{"17.9", "17.10", false},
		{"latest", "17", false},
		{"", "17", false},
		{"17", "latest", false},
		{"latest", "", true},
	} {
		t.Run(tc.version+"/"+tc.minimum, func(t *testing.T) {
			if got := appServiceVersionAllowed(tc.version, tc.minimum); got != tc.want {
				t.Fatalf("allowed = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestAppServiceMinimumVersionMetadataSurvivesSerialization(t *testing.T) {
	var field dto.AppFormFields
	if err := yaml.Unmarshal([]byte("type: service\nkey: postgresql\nminVersion: '17'\n"), &field); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(field)
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(encoded, &got); err != nil {
		t.Fatal(err)
	}
	if got["key"] != "postgresql" || got["minVersion"] != "17" {
		t.Fatalf("service metadata lost: %s", encoded)
	}
}

func TestAppServiceMinimumVersionFiltersAndValidatesSelection(t *testing.T) {
	db := appCacheTestDB(t)
	if err := db.AutoMigrate(&model.AppInstall{}, &model.Database{}); err != nil {
		t.Fatal(err)
	}
	local16 := model.AppInstall{Name: "pg16", Version: "16.6", Status: constant.StatusRunning}
	local17 := model.AppInstall{Name: "pg17", Version: "17.2", Status: constant.StatusRunning}
	if err := db.Create(&local16).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&local17).Error; err != nil {
		t.Fatal(err)
	}
	for _, item := range []model.Database{
		{Name: "pg16", Type: constant.AppPostgresql, AppInstallID: local16.ID, Version: "18.1"},
		{Name: "pg17", Type: constant.AppPostgresql, AppInstallID: local17.ID, Version: "16.1"},
		{Name: "remote16", Type: constant.AppPostgresql, Version: "16.x"},
		{Name: "remote17", Type: constant.AppPostgresql, Version: "17.x"},
		{Name: "unknown", Type: constant.AppPostgresql, Version: ""},
		{Name: "mysql", Type: constant.AppMysql, Version: "18"},
	} {
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, key := range []string{constant.AppPostgresql, constant.AppPostgres} {
		services, err := (&AppInstallService{}).GetServices(key, "17")
		if err != nil {
			t.Fatal(err)
		}
		var names []string
		for _, service := range services {
			names = append(names, service.Value)
		}
		if !reflect.DeepEqual(names, []string{"pg17", "remote17"}) {
			t.Fatalf("unexpected services for %s: %v", key, names)
		}
	}
	all, err := (&AppInstallService{}).GetServices(constant.AppPostgresql, "")
	if err != nil || len(all) != 5 {
		t.Fatalf("unconstrained services changed: %v, %v", all, err)
	}
	const serviceForm = `{"formFields":[{"type":"service","key":"postgresql","envKey":"PANEL_DB_HOST","required":true,"minVersion":"17"}]}`
	const appsForm = `{"formFields":[{"type":"apps","envKey":"PANEL_DB_TYPE","required":true,"minVersion":"17","values":[{"label":"PostgreSQL","value":"postgres"}],"child":{"type":"service","envKey":"PANEL_DB_HOST"}}]}`
	for _, tc := range []struct {
		name, form string
		params     map[string]interface{}
		wantError  bool
	}{
		{"local current", serviceForm, map[string]interface{}{"PANEL_DB_HOST": "pg17"}, false},
		{"local old", serviceForm, map[string]interface{}{"PANEL_DB_HOST": "pg16"}, true},
		{"remote current", serviceForm, map[string]interface{}{"PANEL_DB_HOST": "remote17"}, false},
		{"remote old", serviceForm, map[string]interface{}{"PANEL_DB_HOST": "remote16"}, true},
		{"missing version", serviceForm, map[string]interface{}{"PANEL_DB_HOST": "unknown"}, true},
		{"missing selection", serviceForm, map[string]interface{}{}, true},
		{"wrong type", serviceForm, map[string]interface{}{"PANEL_DB_HOST": "mysql"}, true},
		{"non string", serviceForm, map[string]interface{}{"PANEL_DB_HOST": 17}, true},
		{"apps current", appsForm, map[string]interface{}{"PANEL_DB_TYPE": "postgres", "PANEL_DB_HOST": "pg17"}, false},
		{"apps old", appsForm, map[string]interface{}{"PANEL_DB_TYPE": "postgres", "PANEL_DB_HOST": "pg16"}, true},
		{"apps child constraint", `{"formFields":[{"type":"apps","envKey":"TYPE","values":[{"value":"postgres"}],"child":{"envKey":"HOST","required":true,"minVersion":"17"}}]}`, map[string]interface{}{"TYPE": "postgres", "HOST": "pg16"}, true},
		{"apps parent constraint wins", `{"formFields":[{"type":"apps","envKey":"TYPE","minVersion":"17","values":[{"value":"postgres"}],"child":{"envKey":"HOST","required":true,"minVersion":"18"}}]}`, map[string]interface{}{"TYPE": "postgres", "HOST": "pg17"}, false},
		{"apps forged type", appsForm, map[string]interface{}{"PANEL_DB_TYPE": "mysql", "PANEL_DB_HOST": "mysql"}, true},
		{"unconstrained old", `{"formFields":[{"type":"service","key":"postgresql","envKey":"PANEL_DB_HOST"}]}`, map[string]interface{}{"PANEL_DB_HOST": "pg16"}, false},
		{"unconstrained numeric select", `{"formFields":[{"type":"select","envKey":"OPT","values":[{"value":1}]}]}`, map[string]interface{}{}, false},
		{"optional absent", `{"formFields":[{"type":"service","key":"postgresql","envKey":"HOST","minVersion":"17"}]}`, map[string]interface{}{}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAppServiceVersions(tc.form, tc.params)
			if (err != nil) != tc.wantError {
				t.Fatalf("error presence = %v, want %v", err != nil, tc.wantError)
			}
		})
	}
}

func TestAppServiceMinimumVersionForOrdinaryServices(t *testing.T) {
	db := appCacheTestDB(t)
	if err := db.AutoMigrate(&model.AppInstall{}); err != nil {
		t.Fatal(err)
	}
	app := model.App{Key: "demo-service", Name: "Demo Service"}
	if err := db.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	for _, item := range []model.AppInstall{
		{Name: "old", AppId: app.ID, ServiceName: "old-service", Version: "1.9", Status: constant.StatusRunning},
		{Name: "new", AppId: app.ID, ServiceName: "new-service", Version: "2.0", Status: constant.StatusRunning},
		{Name: "stopped", AppId: app.ID, ServiceName: "stopped-service", Version: "3.0", Status: constant.StatusStopped},
	} {
		if err := db.Create(&item).Error; err != nil {
			t.Fatal(err)
		}
	}
	services, err := (&AppInstallService{}).GetServices(app.Key, "2.0")
	if err != nil || len(services) != 1 || services[0].Value != "new-service" {
		t.Fatalf("unexpected ordinary service filtering: %+v, %v", services, err)
	}
	form := `{"formFields":[{"type":"service","key":"demo-service","envKey":"DEMO_HOST","required":true,"minVersion":"2.0"}]}`
	if err := validateAppServiceVersions(form, map[string]interface{}{"DEMO_HOST": "old-service"}); err == nil {
		t.Fatal("old ordinary service accepted")
	}
	if err := validateAppServiceVersions(form, map[string]interface{}{"DEMO_HOST": "new-service"}); err != nil {
		t.Fatal(err)
	}
}

func TestAppServiceMinimumVersionInstallationRejectsBeforeCreatingResources(t *testing.T) {
	db := appCacheTestDB(t)
	if err := db.AutoMigrate(&model.AppInstall{}, &model.Database{}); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/networks") {
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `[{"Name":"1panel-network"}]`)
			return
		}
		t.Errorf("unexpected Docker request: %s %s", r.Method, r.URL.Path)
		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)
	t.Setenv("DOCKER_API_VERSION", "1.47")
	if err := db.Create(&model.Setting{Key: "DockerSockPath", Value: strings.Replace(server.URL, "http://", "tcp://", 1)}).Error; err != nil {
		t.Fatal(err)
	}
	app := model.App{Key: "version-constrained", Name: "Version Constrained"}
	if err := db.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	detail := model.AppDetail{AppId: app.ID, Version: "1.0", Params: `{"formFields":[{"type":"service","key":"postgresql","envKey":"PANEL_DB_HOST","required":true,"minVersion":"17"}]}`}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.Database{Name: "remote-old", Type: constant.AppPostgresql, Version: "16.x"}).Error; err != nil {
		t.Fatal(err)
	}
	installed, err := (AppService{}).Install(request.AppInstallCreate{
		Name: "rejected-install", AppDetailId: detail.ID, Params: map[string]interface{}{"PANEL_DB_HOST": "remote-old"},
	}, false)
	validationError, ok := err.(buserr.BusinessError)
	if !ok || validationError.Msg != "ErrAppParamKey" || installed != nil {
		t.Fatalf("installation did not reject invalid service selection: installed=%+v, err=%T", installed, err)
	}
	var count int64
	if err := db.Model(&model.AppInstall{}).Count(&count).Error; err != nil || count != 0 {
		t.Fatalf("invalid installation created resources: count=%d, err=%v", count, err)
	}
}
