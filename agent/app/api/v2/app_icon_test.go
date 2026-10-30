package v2

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/appicon"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestAppIconEndpointPreservesRemoteAndHistoricalSources(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "icons.db")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.App{}, &model.AppDetail{}, &model.AppTag{}); err != nil {
		t.Fatal(err)
	}
	oldDB, oldDir, oldLog := global.DB, global.Dir, global.LOG
	global.DB = db
	global.Dir.IconCacheDir = t.TempDir()
	global.LOG = logrus.New()
	global.LOG.SetOutput(io.Discard)
	t.Cleanup(func() {
		global.DB, global.Dir, global.LOG = oldDB, oldDir, oldLog
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})

	var imageRequests atomic.Int32
	cdn := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		imageRequests.Add(1)
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(cdn.Close)
	remoteURL := cdn.URL + "/icon.png?token=a%2Bb&v=42#image"
	image := "\x89PNG\r\n\x1a\noriginal-icon"
	fileName := "app_php.png"
	if err := os.WriteFile(filepath.Join(global.Dir.IconCacheDir, fileName), []byte(image), 0644); err != nil {
		t.Fatal(err)
	}

	app := model.App{Key: "php", Resource: constant.AppResourceRemote}
	if err := db.Create(&app).Error; err != nil {
		t.Fatal(err)
	}
	router := gin.New()
	router.GET("/api/v2/apps/icon/:key", (&BaseApi{}).GetAppIcon)
	for _, tc := range []struct {
		name        string
		icon        string
		ifNoneMatch string
		status      int
		body        string
		etag        string
	}{
		{"remote URL", remoteURL, "", http.StatusFound, "", ""},
		{"remote URL ignores legacy etag", remoteURL, `"old"`, http.StatusFound, "", ""},
		{"historical file", appicon.BuildIconField(fileName, `"42"`), "", http.StatusOK, image, `"42"`},
		{"historical file cached", appicon.BuildIconField(fileName, `"42"`), `"42"`, http.StatusNotModified, "", `"42"`},
		{"historical file changed", appicon.BuildIconField(fileName, `"42"`), `"old"`, http.StatusOK, image, `"42"`},
		{"historical base64", base64.StdEncoding.EncodeToString([]byte(image)), "", http.StatusOK, image, ""},
		{"missing file", "app_missing.png?etag=old", "", http.StatusNoContent, "", ""},
		{"invalid base64", "not-an-icon", "", http.StatusNoContent, "", ""},
		{"invalid URL", "javascript:alert(1)", "", http.StatusNoContent, "", ""},
		{"empty icon", "", "", http.StatusNoContent, "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := db.Model(&app).UpdateColumn("icon", tc.icon).Error; err != nil {
				t.Fatal(err)
			}
			req := httptest.NewRequest(http.MethodGet, "/api/v2/apps/icon/php?operateNode=child", nil)
			req.Header.Set("If-None-Match", tc.ifNoneMatch)
			res := httptest.NewRecorder()
			router.ServeHTTP(res, req)
			if res.Code != tc.status {
				t.Fatalf("status = %d, want %d; body = %s", res.Code, tc.status, res.Body.String())
			}
			if res.Header().Get("ETag") != tc.etag {
				t.Fatalf("etag = %q, want %q", res.Header().Get("ETag"), tc.etag)
			}
			if tc.status == http.StatusFound {
				if res.Header().Get("Location") != remoteURL || res.Header().Get("Cache-Control") != "no-cache" {
					t.Fatalf("redirect URL or cache policy was changed: %v", res.Header())
				}
			} else if res.Body.String() != tc.body {
				t.Fatalf("body = %q, want %q", res.Body.String(), tc.body)
			}
			if imageRequests.Load() != 0 {
				t.Fatalf("icon endpoint fetched the image %d times", imageRequests.Load())
			}
		})
	}
}
