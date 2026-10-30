package service

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
)

func TestComposeSyncOnlyInvalidatesChangedVersionResources(t *testing.T) {
	form := dto.AppForm{FormFields: []dto.AppFormFields{{EnvKey: "PORT", Default: 8080}}}
	params, err := json.Marshal(form)
	if err != nil {
		t.Fatal(err)
	}
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		_, _ = io.WriteString(w, "services:\n  demo:\n    image: demo:unexpected\n")
	}))
	t.Cleanup(server.Close)
	for _, appType := range []string{"runtime", "php", "node", "java", "go", "python", "dotnet", "website"} {
		t.Run(appType, func(t *testing.T) {
			for _, tc := range []struct {
				name      string
				mutate    func(*model.AppDetail, *dto.AppConfigVersion)
				keepCache bool
			}{
				{"unchanged", func(_ *model.AppDetail, _ *dto.AppConfigVersion) {}, true},
				{"package update flag does not expire compose", func(d *model.AppDetail, _ *dto.AppConfigVersion) { d.Update = true }, true},
				{"version revision changed", func(_ *model.AppDetail, v *dto.AppConfigVersion) { v.LastModified++ }, false},
				{"unknown revision", func(_ *model.AppDetail, v *dto.AppConfigVersion) { v.LastModified = 0 }, false},
				{"negative revision", func(_ *model.AppDetail, v *dto.AppConfigVersion) { v.LastModified = -1 }, false},
				{"form changed", func(d *model.AppDetail, _ *dto.AppConfigVersion) { d.Params = `{"formFields":[]}` }, false},
				{"source changed", func(d *model.AppDetail, _ *dto.AppConfigVersion) { d.DownloadUrl += ".old" }, false},
				{"new version", func(d *model.AppDetail, _ *dto.AppConfigVersion) { *d = model.AppDetail{Version: "1.0"} }, false},
			} {
				t.Run(tc.name, func(t *testing.T) {
					const cached = "services:\n  demo:\n    image: demo:cached\n"
					old := model.AppDetail{Version: "1.0", LastModified: 42, Params: string(params), DownloadUrl: server.URL + "/demo/1.0/demo-1.0.tar.gz", DockerCompose: cached}
					version := dto.AppConfigVersion{Name: "1.0", LastModified: 42, AppForm: form}
					tc.mutate(&old, &version)
					ctx := appSyncContext{baseRemoteUrl: server.URL, systemVersion: "2.4.0"}
					app := model.App{Key: "demo", Type: appType, Resource: constant.AppResourceRemote, Details: []model.AppDetail{old}}
					got := ctx.processOneApp(appWorkItem{app: app, appDef: dto.AppDefine{AppProperty: dto.AppProperty{Key: "demo"}, Versions: []dto.AppConfigVersion{version}}})
					if len(got.Details) != 1 || calls.Load() != 0 {
						t.Fatalf("sync fetched compose or lost version: %+v, calls=%d", got, calls.Load())
					}
					detail := got.Details[0]
					want := ""
					if tc.keepCache {
						want = cached
					}
					if detail.DockerCompose != want || detail.LastModified != version.LastModified || detail.Params != string(params) || !detail.Update {
						t.Fatalf("cache decision or version metadata changed: %+v", detail)
					}
					if app.Details[0].DockerCompose != old.DockerCompose {
						t.Fatal("sync mutated its input template")
					}
				})
			}
		})
	}
}
