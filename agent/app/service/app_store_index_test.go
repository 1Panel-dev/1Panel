package service

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/xpack"
)

func TestAppStoreIndexUsesCacheUntilRevisionChanges(t *testing.T) {
	for _, tc := range []struct {
		name       string
		cache      string
		version    int
		failZIP    bool
		wantZIP    int32
		wantUpdate bool
		wantErr    bool
	}{
		{name: "matching cache", cache: `{"lastModified":42}`, version: 42},
		{name: "missing cache", version: 42, wantZIP: 1},
		{name: "malformed cache", cache: `{invalid`, version: 42, wantZIP: 1},
		{name: "null cache", cache: `null`, version: 42, wantZIP: 1},
		{name: "trailing corrupt data", cache: `{"lastModified":42} garbage`, version: 42, wantZIP: 1},
		{name: "outdated cache", cache: `{"lastModified":41}`, version: 42, wantZIP: 1},
		{name: "remote revision changed", cache: `{"lastModified":42}`, version: 43, wantUpdate: true},
		{name: "download failure", cache: `{invalid`, version: 42, failZIP: true, wantZIP: 1, wantErr: true},
		{name: "cached minimum panel version", cache: `{"lastModified":42,"additionalProperties":{"version":"9.9.0"}}`, version: 42, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := iconSyncDB(t)
			if err := db.Create(&model.App{Key: "mysql", Type: "website", Status: constant.AppNormal, Icon: "app_mysql.png", BatchInstallSupport: true}).Error; err != nil {
				t.Fatal(err)
			}
			global.Dir.ResourceDir = t.TempDir()
			cachePath := filepath.Join(global.Dir.ResourceDir, "1panel.json")
			if tc.cache != "" {
				appCacheTestWrite(t, cachePath, tc.cache)
			}
			list := dto.AppList{LastModified: tc.version, Apps: []dto.AppDefine{{Icon: "https://cdn.example/mysql.png", AppProperty: dto.AppProperty{Key: "mysql"}}}}
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			entry, err := zw.Create("1panel.json")
			if err != nil {
				t.Fatal(err)
			}
			if err := json.NewEncoder(entry).Encode(list); err != nil {
				t.Fatal(err)
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			var versionCalls, zipCalls atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "1panel.json.version.txt"):
					versionCalls.Add(1)
					_, _ = io.WriteString(w, strconv.Itoa(tc.version))
				case strings.HasSuffix(r.URL.Path, "1panel.json.zip"):
					zipCalls.Add(1)
					if tc.failZIP {
						w.WriteHeader(http.StatusServiceUnavailable)
						return
					}
					_, _ = w.Write(archive.Bytes())
				default:
					t.Errorf("unexpected remote resource request: %s", r.URL)
					w.WriteHeader(http.StatusNotFound)
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

			res, err := (AppService{}).GetAppUpdate()
			if (err != nil) != tc.wantErr || res == nil && !tc.wantErr {
				t.Fatalf("update check = %+v, %v; want error %v", res, err, tc.wantErr)
			}
			if zipCalls.Load() != tc.wantZIP || versionCalls.Load() != 1 {
				t.Fatalf("first check requests: version=%d, ZIP=%d; want 1/%d", versionCalls.Load(), zipCalls.Load(), tc.wantZIP)
			}
			if tc.wantErr {
				if tc.wantZIP == 0 && !strings.Contains(err.Error(), "ErrVersionTooLow") {
					t.Fatalf("cached panel-version check was lost: %v", err)
				}
				return
			}
			if res.CanUpdate != tc.wantUpdate {
				t.Fatalf("CanUpdate=%v, want %v", res.CanUpdate, tc.wantUpdate)
			}
			if tc.wantUpdate {
				// Full synchronization still fetches the new index after detecting a revision change.
				fresh, err := getAppList()
				if err != nil || fresh.LastModified != tc.version || zipCalls.Load() != 1 {
					t.Fatalf("changed revision did not fetch the new index: %+v, %v, ZIP=%d", fresh, err, zipCalls.Load())
				}
				return
			}
			if res.AppList == nil || res.AppList.LastModified != 42 {
				t.Fatalf("same-revision icon refresh lost its index: %+v", res)
			}
			if _, err := (AppService{}).GetAppUpdate(); err != nil || zipCalls.Load() != tc.wantZIP || versionCalls.Load() != 2 {
				t.Fatalf("second check redownloaded cached index: %v, version=%d, ZIP=%d", err, versionCalls.Load(), zipCalls.Load())
			}
		})
	}
}
