package service

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

func appCacheTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "cache.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.App{}, &model.AppDetail{}, &model.AppTag{}, &model.Setting{}); err != nil {
		t.Fatal(err)
	}
	oldDB, oldDir, oldLog := global.DB, global.Dir, global.LOG
	global.DB = db
	global.Dir = global.SystemDir{AppResourceDir: t.TempDir()}
	global.Dir.RemoteAppResourceDir = filepath.Join(global.Dir.AppResourceDir, constant.AppResourceRemote)
	global.Dir.LocalAppResourceDir = filepath.Join(global.Dir.AppResourceDir, constant.AppResourceLocal)
	global.Dir.CustomAppResourceDir = filepath.Join(global.Dir.AppResourceDir, constant.AppResourceCustom)
	global.LOG = logrus.New()
	global.LOG.SetOutput(io.Discard)
	t.Cleanup(func() {
		global.DB, global.Dir, global.LOG = oldDB, oldDir, oldLog
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.Create(&model.Setting{Key: "AppStoreLastModified", Value: "42"}).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func appCacheTestArchive(t *testing.T, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0644, Size: int64(len(content))}); err != nil {
		t.Fatal(err)
	}
	if _, err := io.WriteString(tw, content); err != nil {
		t.Fatal(err)
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func appCacheTestWrite(t *testing.T, name, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(name, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestAppDownloadURLPreservesSourceQuery(t *testing.T) {
	for _, rawURL := range []string{
		"https://apps.1panel.pro/php/8/php-8.tar.gz",
		"https://apps.1panel.pro/php/8/php-8.tar.gz?token=a%2Bb&region=intl&v=old#archive",
	} {
		t.Run(rawURL, func(t *testing.T) {
			got, err := appDownloadURL(rawURL, "42 /next")
			if err != nil {
				t.Fatal(err)
			}
			parsed, err := url.Parse(got)
			if err != nil {
				t.Fatal(err)
			}
			original, _ := url.Parse(rawURL)
			wantQuery := original.Query()
			wantQuery.Set("v", "42 /next")
			if !reflect.DeepEqual(parsed.Query(), wantQuery) || parsed.Path != original.Path || parsed.Fragment != original.Fragment {
				t.Fatalf("download URL lost source fields: %s", got)
			}
		})
	}
	if _, err := appDownloadURL("https://apps.1panel.pro/%zz", "42"); err == nil {
		t.Fatal("invalid source URL was accepted")
	}
}

func TestDownloadAppRefreshesAllOfficialCachedPackages(t *testing.T) {
	db := appCacheTestDB(t)
	app := model.App{Key: "mysql", Name: "MySQL", Type: "website", Resource: constant.AppResourceRemote}
	versionDir := filepath.Join(app.GetAppResourcePath(), "8")
	dataPath := filepath.Join(versionDir, "data.yml")
	stampPath := filepath.Join(versionDir, ".1panel-appstore-version")
	appCacheTestWrite(t, dataPath, "stale")
	archives := map[string][]byte{
		"42": appCacheTestArchive(t, "mysql/8/data.yml", "fresh-42"),
		"43": appCacheTestArchive(t, "mysql/8/data.yml", "fresh-43"),
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Query().Get("token") != "a+b" || r.URL.Query().Get("region") != "intl" {
			t.Errorf("source query was lost: %s", r.URL.RawQuery)
		}
		version := r.URL.Query().Get("v")
		if version != "42" && version != "43" {
			t.Errorf("unexpected app store version: %q", version)
		}
		_, _ = w.Write(archives[version])
	}))
	t.Cleanup(server.Close)
	detail := model.AppDetail{Version: "8", DownloadUrl: server.URL + "/mysql-8.tar.gz?token=a%2Bb&region=intl", Update: false}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatal(err)
	}
	for _, step := range []struct {
		version string
		calls   int32
		update  bool
	}{
		{"42", 1, false}, // An existing directory without a stamp must be refreshed.
		{"42", 1, false},
		{"43", 2, false},
		{"43", 3, true},
	} {
		if err := db.Model(&model.Setting{}).Where("key = ?", "AppStoreLastModified").Update("value", step.version).Error; err != nil {
			t.Fatal(err)
		}
		detail.Update = step.update
		if err := db.Model(&detail).UpdateColumn("update", detail.Update).Error; err != nil {
			t.Fatal(err)
		}
		if err := downloadApp(app, detail, nil, nil); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != step.calls {
			t.Fatalf("version %s: HTTP calls = %d, want %d", step.version, calls.Load(), step.calls)
		}
		for name, want := range map[string]string{dataPath: "fresh-" + step.version, stampPath: step.version} {
			data, err := os.ReadFile(name)
			if err != nil || string(data) != want {
				t.Fatalf("%s = %q, %v; want %q", name, data, err, want)
			}
		}
		var stored model.AppDetail
		if err := db.First(&stored, detail.ID).Error; err != nil {
			t.Fatal(err)
		}
		if stored.Update || stored.DownloadUrl != detail.DownloadUrl {
			t.Fatalf("refresh altered source URL or left update flag: %+v", stored)
		}
	}
}

func TestDownloadAppFailureKeepsCacheVersionAndUpdateFlag(t *testing.T) {
	for _, failure := range []string{"download", "decompress"} {
		t.Run(failure, func(t *testing.T) {
			db := appCacheTestDB(t)
			app := model.App{Key: "mysql", Resource: constant.AppResourceRemote}
			stampPath := filepath.Join(app.GetAppResourcePath(), "8", ".1panel-appstore-version")
			appCacheTestWrite(t, stampPath, "41")
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if failure == "download" {
					w.WriteHeader(http.StatusServiceUnavailable)
					return
				}
				_, _ = io.WriteString(w, "not an archive")
			}))
			t.Cleanup(server.Close)
			detail := model.AppDetail{Version: "8", DownloadUrl: server.URL + "/mysql-8.tar.gz", Update: true}
			if err := db.Create(&detail).Error; err != nil {
				t.Fatal(err)
			}
			install := model.AppInstall{}
			if err := downloadApp(app, detail, &install, nil); err == nil {
				t.Fatal("failed refresh was accepted")
			}
			if install.Status != constant.StatusDownloadErr || install.Message == "" {
				t.Fatalf("installation did not report refresh failure: %+v", install)
			}
			stamp, err := os.ReadFile(stampPath)
			if err != nil || string(stamp) != "41" {
				t.Fatalf("failed refresh updated stamp: %q, %v", stamp, err)
			}
			var stored model.AppDetail
			if err := db.First(&stored, detail.ID).Error; err != nil {
				t.Fatal(err)
			}
			if !stored.Update || stored.DownloadUrl != detail.DownloadUrl {
				t.Fatalf("failed refresh changed database state: %+v", stored)
			}
		})
	}
}

func TestGetPHPAppDetailUsesSynchronizedOfficialParams(t *testing.T) {
	const oldYAML = "additionalProperties:\n  formFields:\n    - envKey: PHP_VERSION\n      type: select\n      default: 8.1.34\n"
	const newJSON = `{"formFields":[{"envKey":"PHP_VERSION","type":"select","default":"8.1.35"}]}`
	const compose = "services:\n  php:\n    image: php:test\n"
	for _, tc := range []struct {
		name, resource, compose string
		update                  bool
	}{
		{"remote", constant.AppResourceRemote, compose, false},
		{"local", constant.AppResourceLocal, compose, false},
		{"custom", constant.AppResourceCustom, compose, false},
		{"remote-fetches-compose", constant.AppResourceRemote, "", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := appCacheTestDB(t)
			app := model.App{Key: "php", Name: "PHP", Type: constant.RuntimePHP, Resource: tc.resource}
			if err := db.Omit(clause.Associations).Create(&app).Error; err != nil {
				t.Fatal(err)
			}
			versionDir := filepath.Join(app.GetAppResourcePath(), "8")
			appCacheTestWrite(t, filepath.Join(versionDir, "data.yml"), oldYAML)
			appCacheTestWrite(t, filepath.Join(versionDir, "docker-compose.yml"), compose)
			archive := appCacheTestArchive(t, "php/8/data.yml", oldYAML)
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if r.URL.Query().Get("v") != "42" {
					t.Errorf("missing app store version: %s", r.URL.RawQuery)
				}
				if r.URL.Path == "/docker-compose.yml" {
					_, _ = io.WriteString(w, compose)
					return
				}
				_, _ = w.Write(archive)
			}))
			t.Cleanup(server.Close)
			detail := model.AppDetail{AppId: app.ID, Version: "8", Params: newJSON, DockerCompose: tc.compose, Update: tc.update, DownloadUrl: server.URL + "/php-8.tar.gz"}
			if err := db.Create(&detail).Error; err != nil {
				t.Fatal(err)
			}
			got, err := (AppService{}).GetAppDetail(app.ID, "8", "runtime")
			if err != nil {
				t.Fatal(err)
			}
			wantVersion, wantCalls := "8.1.34", int32(0)
			if tc.resource == constant.AppResourceRemote {
				wantVersion, wantCalls = "8.1.35", 1
			}
			want := map[string]interface{}{"formFields": []interface{}{map[string]interface{}{
				"envKey": "PHP_VERSION", "type": "select", "default": wantVersion,
			}}}
			if !reflect.DeepEqual(got.Params, want) || got.Image != "php" || calls.Load() != wantCalls {
				encoded, _ := json.Marshal(got.Params)
				t.Fatalf("params=%s, image=%s, HTTP calls=%d; want version %s, calls %d", encoded, got.Image, calls.Load(), wantVersion, wantCalls)
			}
			var stored model.AppDetail
			if err := db.First(&stored, detail.ID).Error; err != nil {
				t.Fatal(err)
			}
			if stored.Update || got.Update {
				t.Fatal("fetching the compose file restored a stale update flag")
			}
		})
	}
}

func TestProcessOneAppDefersComposeUntilRequested(t *testing.T) {
	db := appCacheTestDB(t)
	const compose = "services:\n  php:\n    image: php:test\n"
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		if r.URL.Path != "/php/8/docker-compose.yml" || r.URL.Query().Get("v") != "42" {
			t.Errorf("unversioned initial compose request: %s", r.URL.String())
		}
		_, _ = io.WriteString(w, compose)
	}))
	t.Cleanup(server.Close)
	app := model.App{Key: "php", Type: constant.RuntimePHP, Resource: constant.AppResourceRemote}
	if err := db.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	ctx := appSyncContext{
		baseRemoteUrl: server.URL,
		systemVersion: "2.4.0",
		list:          &dto.AppList{LastModified: 42},
	}
	result := ctx.processOneApp(appWorkItem{
		app: app,
		appDef: dto.AppDefine{
			AppProperty: dto.AppProperty{Key: "php", Type: constant.RuntimePHP},
			Versions: []dto.AppConfigVersion{{Name: "8", AppForm: dto.AppForm{
				FormFields: []dto.AppFormFields{{EnvKey: "PHP_VERSION", Default: "8.1.35", Type: "select"}},
			}}},
		},
	})
	if calls.Load() != 0 || len(result.Details) != 1 {
		t.Fatalf("initial compose refresh failed: calls=%d, result=%+v", calls.Load(), result)
	}
	detail := result.Details[0]
	if detail.DockerCompose != "" || detail.DownloadUrl != server.URL+"/php/8/php-8.tar.gz" {
		t.Fatalf("sync fetched compose or changed the canonical archive URL: %+v", detail)
	}
	detail.AppId = app.ID
	if err := db.Create(&detail).Error; err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := ensureAppDetailCompose(app, &detail); err != nil {
			t.Fatal(err)
		}
		if calls.Load() != 1 || detail.DockerCompose != compose {
			t.Fatalf("on-demand compose did not download once and cache: calls=%d, detail=%+v", calls.Load(), detail)
		}
	}
}
