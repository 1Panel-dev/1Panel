package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/app/task"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack/providers"
	"github.com/gin-gonic/gin"
	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"gorm.io/gorm"
)

type iconSyncTransport struct {
	providers.MultiNodeProvider
	transport *http.Transport
}

func (p iconSyncTransport) LoadRequestTransport() *http.Transport { return p.transport }

func iconSyncDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := appCacheTestDB(t)
	if err := db.AutoMigrate(&model.Tag{}, &model.AppInstall{}, &model.AppInstallResource{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&[]model.Setting{{Key: "SystemVersion", Value: "2.4.0"}, {Key: "AppStoreSyncStatus", Value: constant.StatusSyncSuccess}}).Error; err != nil {
		t.Fatal(err)
	}
	oldI18n := global.I18n
	global.I18n = goi18n.NewLocalizer(goi18n.NewBundle(language.English), "en")
	t.Cleanup(func() { global.I18n = oldI18n })
	return db
}

func TestOfficialIconSyncNeverDownloadsImages(t *testing.T) {
	iconSyncDB(t)
	var calls atomic.Int32
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(cdn.Close)
	taskLog := &task.Task{Logger: global.LOG}
	for _, oldIcon := range []string{"", "app_demo.png?etag=old", "b2xk", cdn.URL + "/old.png"} {
		t.Run(oldIcon, func(t *testing.T) {
			icon := cdn.URL + "/demo.png?signature=a%2Bb&v=one"
			def := dto.AppDefine{Icon: icon, AppProperty: dto.AppProperty{Key: "demo", Type: "website"}, Versions: []dto.AppConfigVersion{{Name: "1.0", AppForm: dto.AppForm{}}}}
			apps := getApps([]model.App{{Key: "demo", Icon: oldIcon}}, []dto.AppDefine{def}, "2.4.0", taskLog)
			ctx := appSyncContext{baseRemoteUrl: cdn.URL, list: &dto.AppList{LastModified: 42}}
			result := ctx.processOneApp(appWorkItem{app: apps["demo"], appDef: def})
			if result.Icon != icon || len(result.Details) != 1 || result.Details[0].Version != "1.0" || calls.Load() != 0 {
				t.Fatalf("sync changed metadata or fetched a CDN image: %+v, calls=%d", result, calls.Load())
			}
		})
	}
	for _, invalid := range []string{"", "javascript:alert(1)", "https://user:secret@example.com/a.png", "//example.com/a.png", "https://"} {
		def := dto.AppDefine{Icon: invalid, AppProperty: dto.AppProperty{Key: "demo"}}
		apps := getApps([]model.App{{Key: "demo", Icon: "b2xk"}}, []dto.AppDefine{def}, "2.4.0", taskLog)
		if apps["demo"].Icon != "b2xk" {
			t.Fatalf("invalid icon %q erased history", invalid)
		}
	}
}

func TestOfficialIconSameRevisionRefreshIsReadOnlyUntilSync(t *testing.T) {
	db := iconSyncDB(t)
	apps := []model.App{
		{Key: "mysql", Type: "website", Status: constant.AppNormal, Icon: "app_mysql.png", BatchInstallSupport: true},
		{Key: "demo", Name: "original", Type: "website", Status: constant.AppNormal, Icon: "b2xk", LastModified: 7},
		{Key: "future", Type: "website", Status: constant.AppNormal, Icon: "b2xk"},
		{Key: "local-demo", Type: "website", Resource: constant.AppResourceLocal, Status: constant.AppNormal, Icon: "bG9jYWw="},
		{Key: "custom-demo", Type: "website", Resource: constant.AppResourceCustom, Status: constant.AppNormal, Icon: "Y3VzdG9t"},
		{Key: "down", Type: "website", Status: constant.AppTakeDown, Icon: "b2xk"},
	}
	if err := db.Create(&apps).Error; err != nil {
		t.Fatal(err)
	}
	detail := model.AppDetail{AppId: apps[1].ID, Version: "1.0", Params: "keep", DockerCompose: "keep", Update: false}
	if err := db.Create(&detail).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&detail, detail.ID).Error; err != nil {
		t.Fatal(err)
	}
	var imageCalls, indexCalls atomic.Int32
	var failIndex atomic.Bool
	list := dto.AppList{LastModified: 42}
	for _, app := range apps {
		version := 0.0
		if app.Key == "future" {
			version = 9.9
		}
		list.Apps = append(list.Apps, dto.AppDefine{Icon: "https://cdn.example/" + app.Key + ".png?signature=a%2Bb", AppProperty: dto.AppProperty{Key: app.Key, Version: version}})
	}
	var zipped bytes.Buffer
	zw := zip.NewWriter(&zipped)
	w, err := zw.Create("1panel.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(w).Encode(list); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "1panel.json.version.txt"):
			_, _ = io.WriteString(w, "42")
		case strings.HasSuffix(r.URL.Path, "1panel.json.zip"):
			indexCalls.Add(1)
			if failIndex.Load() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			_, _ = w.Write(zipped.Bytes())
		default:
			imageCalls.Add(1)
			w.WriteHeader(http.StatusServiceUnavailable)
		}
	}))
	t.Cleanup(server.Close)
	transport := server.Client().Transport.(*http.Transport).Clone()
	transport.TLSClientConfig.InsecureSkipVerify = true
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, server.Listener.Addr().String())
	}
	oldProvider := xpack.MultiNodeProvider
	xpack.MultiNodeProvider = iconSyncTransport{MultiNodeProvider: oldProvider, transport: transport}
	t.Cleanup(func() { xpack.MultiNodeProvider = oldProvider; transport.CloseIdleConnections() })
	global.Dir.ResourceDir = t.TempDir()
	global.Dir.IconCacheDir = t.TempDir()
	appCacheTestWrite(t, filepath.Join(global.Dir.IconCacheDir, "app_mysql.png"), "keep old image")
	encodedList, err := json.Marshal(list)
	if err != nil {
		t.Fatal(err)
	}
	appCacheTestWrite(t, filepath.Join(global.Dir.ResourceDir, "1panel.json"), string(encodedList))

	res, err := (AppService{}).GetAppUpdate()
	if err != nil || res.CanUpdate || res.AppList == nil || indexCalls.Load() != 0 {
		t.Fatalf("same revision/missing files forced a full sync: %+v, %v", res, err)
	}
	var original model.App
	if err := db.First(&original, apps[1].ID).Error; err != nil || original.Icon != apps[1].Icon {
		t.Fatalf("update check modified icon: %+v, %v", original, err)
	}
	var writes atomic.Int32
	if err := db.Callback().Update().After("gorm:update").Register("test:icon-writes", func(tx *gorm.DB) {
		if values, ok := tx.Statement.Dest.(map[string]interface{}); ok && tx.Statement.Table == "apps" && values["icon"] != nil {
			writes.Add(1)
		}
	}); err != nil {
		t.Fatal(err)
	}
	var shared *appSyncContext
	for i := 0; i < 2; i++ {
		if err := (AppService{}).createSyncAppStoreTask(&shared)(&task.Task{Logger: global.LOG}); err != nil {
			t.Fatal(err)
		}
		if shared == nil || !shared.skipMetaSync || writes.Load() != 2 || imageCalls.Load() != 0 || indexCalls.Load() != 0 {
			t.Fatalf("unexpected same revision sync: writes=%d, image calls=%d, index calls=%d, context=%+v", writes.Load(), imageCalls.Load(), indexCalls.Load(), shared)
		}
	}
	var stored model.App
	if err := db.First(&stored, apps[1].ID).Error; err != nil {
		t.Fatal(err)
	}
	want := original
	want.Icon = list.Apps[1].Icon
	if !reflect.DeepEqual(stored, want) {
		t.Fatalf("light refresh changed non-icon fields: got %+v, want %+v", stored, want)
	}
	for _, original := range apps[2:] {
		stored = model.App{}
		if err := db.First(&stored, original.ID).Error; err != nil {
			t.Fatal(err)
		}
		if stored.Icon != original.Icon || stored.Status != original.Status {
			t.Fatalf("filtered app was changed: %+v", stored)
		}
	}
	var storedDetail model.AppDetail
	if err := db.First(&storedDetail, detail.ID).Error; err != nil || !reflect.DeepEqual(storedDetail, detail) {
		t.Fatalf("light refresh touched app details: %+v, %v", storedDetail, err)
	}
	failIndex.Store(true)
	appCacheTestWrite(t, filepath.Join(global.Dir.ResourceDir, "1panel.json"), "{invalid")
	if err := (AppService{}).createSyncAppStoreTask(&shared)(&task.Task{Logger: global.LOG}); err == nil || writes.Load() != 2 || indexCalls.Load() != 1 {
		t.Fatal("failed index read was accepted or modified icons")
	}
}

func TestAppIconURLsReachListDetailAndInstalled(t *testing.T) {
	db := iconSyncDB(t)
	for _, icon := range []string{"https://cdn.example/icon.png?signature=a%2Bb", "app_legacy.png?etag=old", "bG9jYWw="} {
		app := model.App{Key: "demo-" + icon[:4], Name: "demo", Type: "website", Status: constant.AppNormal, Icon: icon}
		app.Name = app.Key
		if err := db.Create(&app).Error; err != nil {
			t.Fatal(err)
		}
		ctx, _ := gin.CreateTestContext(httptest.NewRecorder())
		ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		got, err := (AppService{}).PageApp(ctx, request.AppSearch{PageInfo: dto.PageInfo{Page: 1, PageSize: 10}, Name: app.Key})
		if err != nil || len(got.Items) != 1 {
			t.Fatalf("list failed: %+v, %v", got, err)
		}
		want := ""
		if strings.HasPrefix(icon, "https:") {
			want = icon
		}
		if got.Items[0].Icon != want {
			t.Fatalf("list icon = %q, want %q", got.Items[0].Icon, want)
		}
		detail, err := (AppService{}).GetApp(ctx, app.Key)
		if err != nil || detail.Icon != icon {
			t.Fatalf("detail icon = %+v, %v", detail, err)
		}
		installed, err := handleInstalled([]model.AppInstall{{App: app, Status: constant.StatusRunning}}, false, false, false)
		if err != nil || len(installed) != 1 || installed[0].Icon != want {
			t.Fatalf("installed icon = %+v, %v", installed, err)
		}
	}
}

func TestLocalAppLogoKeepsBase64Delivery(t *testing.T) {
	db := iconSyncDB(t)
	dir := t.TempDir()
	const image = "\x89PNG\r\n\x1a\nlocal-icon"
	appCacheTestWrite(t, filepath.Join(dir, "data.yml"), "additionalProperties:\n  key: demo\n  name: Local Demo\n  type: website\n")
	appCacheTestWrite(t, filepath.Join(dir, "README.md"), "local application")
	appCacheTestWrite(t, filepath.Join(dir, "logo.png"), image)
	app, err := handleLocalApp(dir)
	if err != nil || app.Resource != constant.AppResourceLocal || app.Icon != base64.StdEncoding.EncodeToString([]byte(image)) {
		t.Fatalf("local logo parsing changed: %+v, %v", app, err)
	}
	if err := db.Create(app).Error; err != nil {
		t.Fatal(err)
	}
	data, source, etag, err := (AppService{}).GetAppIcon(app.Key)
	if err != nil || string(data) != image || source != "" || etag != "" {
		t.Fatalf("local logo no longer uses historical delivery: %q, %q, %q, %v", data, source, etag, err)
	}
}
