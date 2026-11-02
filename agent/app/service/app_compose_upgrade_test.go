package service

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const composeUpgradeCurrent = `services:
  demo:
    image: demo:1.0
    deploy:
      resources:
        limits:
          cpus: ${CPUS}
          memory: ${MEMORY_LIMIT}
`

func composeUpgradeDB(t *testing.T, resourceURL string) (*gorm.DB, model.App, model.AppInstall, []model.AppDetail) {
	t.Helper()
	db := iconSyncDB(t)
	if err := db.AutoMigrate(&model.AppIgnoreUpgrade{}); err != nil {
		t.Fatal(err)
	}
	app := model.App{Key: "demo", Name: "Demo", Type: "runtime", Resource: constant.AppResourceRemote, CrossVersionUpdate: true}
	if err := db.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	var details []model.AppDetail
	for _, version := range []string{"1.1", "1.0", "1.3", "1.2"} {
		detail := model.AppDetail{
			AppId: app.ID, Version: version, LastModified: 42, Params: `{"formFields":[]}`, Status: constant.AppNormal,
			DownloadUrl: resourceURL + "/demo/" + version + "/demo-" + version + ".tar.gz",
		}
		if err := db.Create(&detail).Error; err != nil {
			t.Fatal(err)
		}
		details = append(details, detail)
	}
	if err := db.Create(&model.AppIgnoreUpgrade{AppID: app.ID, AppDetailID: details[2].ID, Scope: "version"}).Error; err != nil {
		t.Fatal(err)
	}
	install := model.AppInstall{
		AppId: app.ID, AppDetailId: details[1].ID, Name: "demo-install", Version: "1.0", App: app,
		ServiceName: "demo", ContainerName: "demo-container", Status: constant.StatusRunning,
		Env: `{"CONTAINER_NAME":"demo-container"}`, DockerCompose: composeUpgradeCurrent,
	}
	if err := db.Omit(clause.Associations).Create(&install).Error; err != nil {
		t.Fatal(err)
	}
	return db, app, install, details
}

func TestUpgradeComposeLoadsOnlyEffectiveTarget(t *testing.T) {
	var mu sync.Mutex
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests[r.URL.Path]++
		mu.Unlock()
		if r.URL.Query().Get("v") != "42" {
			t.Errorf("missing synchronized version: %s", r.URL)
		}
		version := strings.Split(r.URL.Path, "/")[2]
		_, _ = fmt.Fprintf(w, "services:\n  source:\n    image: demo:%s\n", version)
	}))
	t.Cleanup(server.Close)
	db, _, install, details := composeUpgradeDB(t, server.URL)
	// An unselected bad cached template must not prevent listing or default preview.
	if err := db.Model(&details[0]).UpdateColumn("docker_compose", "invalid: [").Error; err != nil {
		t.Fatal(err)
	}
	service := NewIAppInstalledService()
	noCompose := false
	versions, err := service.GetUpdateVersions(request.AppUpdateVersion{AppInstallID: install.ID, LoadCompose: &noCompose})
	if err != nil || len(versions) != 2 || versions[0].Version != "1.2" || versions[1].Version != "1.1" {
		t.Fatalf("pure versions did not preserve filtering/sorting: %+v, %v", versions, err)
	}
	for _, version := range versions {
		if version.DockerCompose != "" {
			t.Fatal("pure version request returned a rendered template")
		}
	}
	if len(requests) != 0 {
		t.Fatalf("pure version request fetched templates: %v", requests)
	}
	for i := 0; i < 2; i++ {
		versions, err = service.GetUpdateVersions(request.AppUpdateVersion{AppInstallID: install.ID})
		if err != nil || len(versions) != 2 || !strings.Contains(versions[0].DockerCompose, "demo:1.2") || versions[1].DockerCompose != "" {
			t.Fatalf("default target preview = %+v, %v", versions, err)
		}
	}
	if requests["/demo/1.2/docker-compose.yml"] != 1 || len(requests) != 1 {
		t.Fatalf("default preview did not reuse only its own cache: %v", requests)
	}
	if err := db.Model(&details[0]).UpdateColumn("docker_compose", "").Error; err != nil {
		t.Fatal(err)
	}
	versions, err = service.GetUpdateVersions(request.AppUpdateVersion{AppInstallID: install.ID, UpdateVersion: "1.1"})
	if err != nil || !strings.Contains(versions[1].DockerCompose, "demo:1.1") || versions[0].DockerCompose != "" {
		t.Fatalf("selected target preview = %+v, %v", versions, err)
	}
	if requests["/demo/1.1/docker-compose.yml"] != 1 || len(requests) != 2 {
		t.Fatalf("selection prefetched other candidates: %v", requests)
	}
	if _, err := service.GetUpdateVersions(request.AppUpdateVersion{AppInstallID: install.ID, UpdateVersion: "1.3"}); err == nil {
		t.Fatal("ignored target was accepted")
	}
	if len(requests) != 2 {
		t.Fatal("invalid target triggered a template request")
	}
}

func TestUpgradeComposeFetchFailurePreservesInstanceAndCache(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	t.Cleanup(server.Close)
	db, _, install, details := composeUpgradeDB(t, server.URL)
	if _, err := NewIAppInstalledService().GetUpdateVersions(request.AppUpdateVersion{AppInstallID: install.ID}); err == nil {
		t.Fatal("failed target template load was accepted")
	}
	var stored model.AppInstall
	if err := db.First(&stored, install.ID).Error; err != nil || stored.DockerCompose != install.DockerCompose || stored.Status != install.Status || stored.Version != install.Version {
		t.Fatalf("preview modified installed instance: %+v, %v", stored, err)
	}
	var detail model.AppDetail
	if err := db.First(&detail, details[3].ID).Error; err != nil || detail.DockerCompose != "" {
		t.Fatalf("failed response cached a template: %+v, %v", detail, err)
	}
}

func TestUpgradePrepareComposeFailureDoesNotCutover(t *testing.T) {
	for _, failure := range []string{"package download", "invalid package template"} {
		t.Run(failure, func(t *testing.T) {
			archive := appCacheTestArchive(t, "demo/1.2/docker-compose.yml", "services: []")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if failure == "package download" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				_, _ = w.Write(archive)
			}))
			t.Cleanup(server.Close)
			db, _, install, details := composeUpgradeDB(t, server.URL)
			upgrade := &appUpgradeContext{original: install, detail: details[3]}
			if err := upgrade.prepare(&task.Task{Logger: global.LOG}); err == nil {
				t.Fatal("invalid upgrade resources passed prepare")
			}
			if upgrade.stopAttempted || upgrade.downAttempted || upgrade.stageDir != "" {
				t.Fatal("failed template preparation reached live cutover")
			}
			var stored model.AppInstall
			if err := db.First(&stored, install.ID).Error; err != nil || stored.DockerCompose != install.DockerCompose || stored.Env != install.Env || stored.Version != install.Version || stored.Status != install.Status {
				t.Fatalf("failed prepare modified the instance: %+v, %v", stored, err)
			}
		})
	}
}

func TestExplicitConfigComposeDetectsCustomAndReusesCurrentCache(t *testing.T) {
	for _, tc := range []struct {
		name    string
		custom  bool
		failure bool
	}{
		{name: "original"},
		{name: "custom", custom: true},
		{name: "fetch failure", failure: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				if r.URL.Path != "/demo/1.0/docker-compose.yml" {
					t.Errorf("configuration loaded a different version: %s", r.URL)
				}
				if tc.failure {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				_, _ = io.WriteString(w, "services:\n  source:\n    image: demo:1.0\n")
			}))
			t.Cleanup(server.Close)
			db, _, install, _ := composeUpgradeDB(t, server.URL)
			if tc.custom {
				install.DockerCompose = strings.Replace(install.DockerCompose, "image: demo:1.0", "image: demo:1.0\n    environment:\n      USER_SETTING: preserve", 1)
				if err := db.Omit(clause.Associations).Save(&install).Error; err != nil {
					t.Fatal(err)
				}
			}
			if isEditCompose(install) || calls != 0 {
				t.Fatal("list-time custom check fetched a missing template")
			}
			result, err := NewIAppInstalledService().GetParams(install.ID)
			if tc.failure {
				if err == nil || result != nil {
					t.Fatal("failed current-template load was treated as an unedited configuration")
				}
				return
			}
			if err != nil || result.RawCompose == "" || result.IsEdit != tc.custom || result.DockerCompose != install.DockerCompose {
				t.Fatalf("explicit custom state = %+v, %v; want custom=%v", result, err, tc.custom)
			}
			if _, err := NewIAppInstalledService().GetParams(install.ID); err != nil || calls != 1 {
				t.Fatalf("configuration cache was not reused: calls=%d, %v", calls, err)
			}
			if isEditCompose(install) != tc.custom || calls != 1 {
				t.Fatal("cached list-time comparison changed custom state or made network requests")
			}
		})
	}
}

func TestVllmManagedComposeDoesNotRequireOfficialTemplate(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(server.Close)
	db, app, install, _ := composeUpgradeDB(t, server.URL)
	if err := db.Model(&app).UpdateColumn("key", vllmAppKeyForUpgrade).Error; err != nil {
		t.Fatal(err)
	}
	install.DockerCompose = strings.Replace(install.DockerCompose, "image: demo:1.0", "image: demo:1.0\n    command: [--managed-user-option]", 1)
	if err := db.Omit(clause.Associations).Save(&install).Error; err != nil {
		t.Fatal(err)
	}
	service := NewIAppInstalledService()
	config, err := service.GetParams(install.ID)
	if err != nil || config.RawCompose != install.DockerCompose || config.DockerCompose != install.DockerCompose || config.IsEdit {
		t.Fatalf("managed configuration lost the instance template: %+v, %v", config, err)
	}
	versions, err := service.GetUpdateVersions(request.AppUpdateVersion{AppInstallID: install.ID})
	if err != nil || len(versions) != 2 || versions[0].DockerCompose != install.DockerCompose {
		t.Fatalf("managed upgrade preview lost the instance template: %+v, %v", versions, err)
	}
	if calls.Load() != 0 {
		t.Fatalf("managed template attempted %d official resource requests", calls.Load())
	}
}
