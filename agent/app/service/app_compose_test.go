package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"gorm.io/gorm"
)

const composeTestTemplate = "services:\n  demo:\n    image: demo:latest\n    network_mode: host\n"

func composeTestDetail(t *testing.T, resource, source string) (*gorm.DB, model.App, model.AppDetail) {
	t.Helper()
	db := appCacheTestDB(t)
	app := model.App{Key: "demo", Name: "Demo", Resource: resource, Type: "website"}
	if err := db.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	detail := model.AppDetail{AppId: app.ID, Version: "1.0", LastModified: 42, Params: `{"formFields":[]}`, DownloadUrl: source}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatal(err)
	}
	return db, app, detail
}

func TestComposeOnDemandLoadsOnceAndPreservesSource(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/demo/1.0/docker-compose.yml" || r.URL.Query().Get("signature") != "a+b" || r.URL.Query().Get("v") != "42" {
			t.Errorf("unexpected Compose source: %s", r.URL)
		}
		_, _ = io.WriteString(w, composeTestTemplate)
	}))
	t.Cleanup(server.Close)
	db, app, detail := composeTestDetail(t, constant.AppResourceRemote, server.URL+"/demo/1.0/demo-1.0.tar.gz?signature=a%2Bb")
	for i := 0; i < 2; i++ {
		if err := ensureAppDetailCompose(app, &detail); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 1 || detail.DockerCompose != composeTestTemplate {
			t.Fatalf("cache miss after loading: calls=%d, detail=%+v", calls.Load(), detail)
		}
	}
	var stored model.AppDetail
	if err := db.First(&stored, detail.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.DockerCompose != composeTestTemplate || stored.LastModified != 42 || stored.DownloadUrl != detail.DownloadUrl || stored.Params != detail.Params || !stored.UpdatedAt.Equal(detail.UpdatedAt) {
		t.Fatalf("cache write altered version metadata: %+v", stored)
	}
	res, err := (AppService{}).GetAppDetailByID(detail.ID)
	if err != nil || !res.HostMode || res.DockerCompose != composeTestTemplate || calls.Load() != 1 {
		t.Fatalf("ID details did not reuse template/HostMode: %+v, %v, calls=%d", res, err, calls.Load())
	}
}

func TestComposeOnDemandIDDetailsFillMissingTemplate(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = io.WriteString(w, composeTestTemplate)
	}))
	t.Cleanup(server.Close)
	_, _, detail := composeTestDetail(t, constant.AppResourceRemote, server.URL+"/demo.tar.gz")
	res, err := (AppService{}).GetAppDetailByID(detail.ID)
	if err != nil || !res.HostMode || res.DockerCompose != composeTestTemplate || calls.Load() != 1 {
		t.Fatalf("missing template gave wrong ID details: %+v, %v, calls=%d", res, err, calls.Load())
	}
}

func TestComposeOnDemandLocalAndCustomUseTheirDirectories(t *testing.T) {
	for _, resource := range []string{constant.AppResourceLocal, constant.AppResourceCustom} {
		t.Run(resource, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { calls.Add(1) }))
			t.Cleanup(server.Close)
			_, app, detail := composeTestDetail(t, resource, server.URL+"/must-not-fetch.tar.gz")
			appCacheTestWrite(t, filepath.Join(app.GetAppResourcePath(), detail.Version, "docker-compose.yml"), composeTestTemplate)
			if err := ensureAppDetailCompose(app, &detail); err != nil {
				t.Fatal(err)
			}
			if detail.DockerCompose != composeTestTemplate || calls.Load() != 0 {
				t.Fatalf("local resource used remote template: %+v, calls=%d", detail, calls.Load())
			}
		})
	}
}

func TestComposeOnDemandRejectsDownloadAndTemplateFailures(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		status     int
	}{
		{"not found", "missing", http.StatusNotFound},
		{"server error", "failure", http.StatusServiceUnavailable},
		{"empty", "", http.StatusOK},
		{"HTML", "<html>blocked</html>", http.StatusOK},
		{"malformed YAML", "services: [", http.StatusOK},
		{"missing services", "version: '3'", http.StatusOK},
		{"empty services", "services: {}", http.StatusOK},
		{"empty service", "services:\n  demo:", http.StatusOK},
		{"string service", "services:\n  demo: blocked", http.StatusOK},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
				_, _ = io.WriteString(w, tc.body)
			}))
			t.Cleanup(server.Close)
			db, app, detail := composeTestDetail(t, constant.AppResourceRemote, server.URL+"/demo.tar.gz")
			if err := ensureAppDetailCompose(app, &detail); err == nil {
				t.Fatal("invalid template was accepted")
			}
			var stored model.AppDetail
			if err := db.First(&stored, detail.ID).Error; err != nil {
				t.Fatal(err)
			}
			if detail.DockerCompose != "" || stored.DockerCompose != "" {
				t.Fatal("failed response was cached")
			}
		})
	}
}

func TestComposeCacheRejectsConcurrentMetadataUpdates(t *testing.T) {
	for _, field := range []string{"last_modified", "params", "download_url"} {
		t.Run(field, func(t *testing.T) {
			db, _, detail := composeTestDetail(t, constant.AppResourceRemote, "https://example.com/demo.tar.gz")
			value := interface{}("new value")
			if field == "last_modified" {
				value = 43
			}
			if err := db.Model(&model.AppDetail{}).Where("id = ?", detail.ID).Updates(map[string]interface{}{field: value, "docker_compose": "new template"}).Error; err != nil {
				t.Fatal(err)
			}
			if err := cacheAppDetailCompose(&detail, []byte(composeTestTemplate)); err == nil {
				t.Fatal("stale template overwrote newer metadata")
			}
			var stored model.AppDetail
			if err := db.First(&stored, detail.ID).Error; err != nil {
				t.Fatal(err)
			}
			if stored.DockerCompose != "new template" || detail.DockerCompose != "" {
				t.Fatalf("stale cache result was retained: %+v", stored)
			}
		})
	}
}

func TestComposeCachePropagatesWriteFailure(t *testing.T) {
	db, _, detail := composeTestDetail(t, constant.AppResourceRemote, "https://example.com/demo.tar.gz")
	if err := db.Exec("CREATE TRIGGER reject_compose BEFORE UPDATE OF docker_compose ON app_details BEGIN SELECT RAISE(FAIL, 'test write failure'); END").Error; err != nil {
		t.Fatal(err)
	}
	if err := cacheAppDetailCompose(&detail, []byte(composeTestTemplate)); err == nil || detail.DockerCompose != "" {
		t.Fatal("failed cache write was ignored")
	}
}

func TestMarkDownloadedPreservesComposeAndRejectsChangedMetadata(t *testing.T) {
	db, _, detail := composeTestDetail(t, constant.AppResourceRemote, "https://example.com/demo.tar.gz")
	if err := db.Model(&model.AppDetail{}).Where("id = ?", detail.ID).Updates(map[string]interface{}{"update": true, "docker_compose": composeTestTemplate}).Error; err != nil {
		t.Fatal(err)
	}
	stored, err := appDetailRepo.MarkDownloaded(detail)
	if err != nil || !stored {
		t.Fatalf("matching download marker failed: stored=%v err=%v", stored, err)
	}
	var fresh model.AppDetail
	if err := db.First(&fresh, detail.ID).Error; err != nil || fresh.Update || fresh.DockerCompose != composeTestTemplate {
		t.Fatalf("download marker overwrote Compose: %+v, %v", fresh, err)
	}
	for _, field := range []string{"last_modified", "params", "download_url"} {
		t.Run(field, func(t *testing.T) {
			var snapshot model.AppDetail
			if err := db.First(&snapshot, detail.ID).Error; err != nil {
				t.Fatal(err)
			}
			value := interface{}("new metadata")
			if field == "last_modified" {
				value = snapshot.LastModified + 1
			}
			if err := db.Model(&model.AppDetail{}).Where("id = ?", detail.ID).Updates(map[string]interface{}{field: value, "update": true}).Error; err != nil {
				t.Fatal(err)
			}
			stored, err := appDetailRepo.MarkDownloaded(snapshot)
			if err != nil || stored {
				t.Fatalf("stale download marked new resources ready: stored=%v err=%v", stored, err)
			}
			if err := db.First(&fresh, detail.ID).Error; err != nil || !fresh.Update || fresh.DockerCompose != composeTestTemplate {
				t.Fatalf("stale marker overwrote template or update flag: %+v, %v", fresh, err)
			}
		})
	}
}

func TestRuntimeDetailUsesValidPackageComposeWithoutAnotherGET(t *testing.T) {
	var calls atomic.Int32
	archive := appCacheTestArchive(t, "demo/1.0/docker-compose.yml", composeTestTemplate)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if !strings.HasSuffix(r.URL.Path, ".tar.gz") {
			t.Errorf("runtime requested Compose after obtaining a valid package: %s", r.URL)
		}
		_, _ = w.Write(archive)
	}))
	t.Cleanup(server.Close)
	_, app, detail := composeTestDetail(t, constant.AppResourceRemote, server.URL+"/demo.tar.gz")
	app.Type = constant.RuntimeNode
	if err := global.DB.Model(&app).UpdateColumn("type", app.Type).Error; err != nil {
		t.Fatal(err)
	}
	res, err := (AppService{}).GetAppDetail(app.ID, detail.Version, "runtime")
	if err != nil || res.DockerCompose != composeTestTemplate || !res.HostMode || calls.Load() != 1 {
		t.Fatalf("package Compose was not reused: %+v, %v, calls=%d", res, err, calls.Load())
	}
}

func TestRuntimeDetailDoesNotCacheStaleTemplateAfterPackageFailure(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if strings.HasSuffix(r.URL.Path, ".tar.gz") {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = io.WriteString(w, composeTestTemplate)
	}))
	t.Cleanup(server.Close)
	db, app, detail := composeTestDetail(t, constant.AppResourceRemote, server.URL+"/demo.tar.gz")
	app.Type = constant.RuntimeNode
	if err := db.Model(&app).UpdateColumn("type", app.Type).Error; err != nil {
		t.Fatal(err)
	}
	appCacheTestWrite(t, filepath.Join(app.GetAppResourcePath(), detail.Version, "docker-compose.yml"), "services:\n  stale:\n    image: stale:old\n")
	res, err := (AppService{}).GetAppDetail(app.ID, detail.Version, "runtime")
	if err != nil || res.DockerCompose != composeTestTemplate || calls.Load() != 2 {
		t.Fatalf("failed package download reused a stale local template: %+v, %v, calls=%d", res, err, calls.Load())
	}
	var stored model.AppDetail
	if err := db.First(&stored, detail.ID).Error; err != nil || stored.DockerCompose != composeTestTemplate {
		t.Fatalf("stale package contents were cached: %+v, %v", stored, err)
	}
}
