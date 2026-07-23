package service

import (
	"encoding/json"
	"errors"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func TestNormalizeNginxModuleDoesNotInferBuildMode(t *testing.T) {
	module := dto.NginxModule{
		Name:     "legacy",
		Packages: []string{"git", "", "git", " curl "},
	}

	normalizeNginxModule(&module)

	if module.BuildMode != "" {
		t.Fatalf("build mode must be explicit, got %s", module.BuildMode)
	}
	if module.Provider != nginxModuleProviderLocal {
		t.Fatalf("expected local provider, got %s", module.Provider)
	}
	if len(module.Packages) != 2 || module.Packages[0] != "git" || module.Packages[1] != "curl" {
		t.Fatalf("unexpected normalized packages: %#v", module.Packages)
	}
}

func TestNormalizeDynamicModuleParams(t *testing.T) {
	params, err := normalizeDynamicModuleParams("--with-http_dav_module --add-module=/tmp/nginx-dav-ext-module")
	if err != nil {
		t.Fatal(err)
	}
	expected := "--with-http_dav_module --add-dynamic-module=/tmp/nginx-dav-ext-module"
	if params != expected {
		t.Fatalf("expected %q, got %q", expected, params)
	}

	if _, err = normalizeDynamicModuleParams("--add-module=/tmp/module;touch /tmp/unsafe"); err == nil {
		t.Fatal("expected shell metacharacters to be rejected")
	}
}

func TestFindCurrentAndLatestNginxModuleBuild(t *testing.T) {
	target := dto.NginxModuleTarget{Key: "target"}
	module := dto.NginxModule{
		Name:      "example",
		Params:    "--add-module=/tmp/example",
		BuildMode: nginxModuleBuildDynamic,
		Provider:  nginxModuleProviderLocal,
	}
	params, err := normalizeDynamicModuleParams(module.Params)
	if err != nil {
		t.Fatal(err)
	}
	currentHash, err := nginxModuleBuildHash(module, target, params)
	if err != nil {
		t.Fatal(err)
	}
	oldTime := time.Now().Add(-time.Hour)
	newTime := time.Now()
	module.Builds = []dto.NginxModuleBuild{
		{Hash: "old", Status: nginxModuleStatusReady, Target: target, BuiltAt: oldTime},
		{Hash: currentHash, Status: nginxModuleStatusReady, Target: target, BuiltAt: newTime},
	}

	if build := findCurrentNginxModuleBuild(module, target); build == nil || build.Hash != currentHash {
		t.Fatalf("current build was not selected: %#v", build)
	}
	if build := findLatestNginxModuleBuild(module, target); build == nil || build.Hash != currentHash {
		t.Fatalf("latest build was not selected: %#v", build)
	}

	module.Params = "--add-module=/tmp/example-v2"
	if build := findCurrentNginxModuleBuild(module, target); build != nil {
		t.Fatalf("changed module input should be stale, got %#v", build)
	}
	if build := findLatestNginxModuleBuild(module, target); build == nil || build.Hash != currentHash {
		t.Fatal("the previous ready build should remain available until replacement")
	}
}

func TestNginxModulePathNameAvoidsSanitizedNameCollisions(t *testing.T) {
	first := nginxModulePathName("example/module")
	second := nginxModulePathName("example-module")
	if first == second {
		t.Fatalf("module path names collided: %s", first)
	}
	if len(nginxModulePathName(string(make([]byte, 256)))) > 57 {
		t.Fatal("module path name should remain safe for Docker resource names")
	}
}

func TestRecordNginxModuleBuildFailureKeepsPreviousReadyBuild(t *testing.T) {
	target := dto.NginxModuleTarget{Key: "target"}
	ready := dto.NginxModuleBuild{
		Hash: "ready", Status: nginxModuleStatusReady, Target: target, BuiltAt: time.Now().Add(-time.Hour),
	}
	original := []dto.NginxModule{{
		Name: "example", BuildMode: nginxModuleBuildDynamic,
		Builds: []dto.NginxModuleBuild{ready},
	}}
	failed := dto.NginxModuleBuild{
		Hash: "candidate", Status: nginxModuleStatusFailed, Target: target, Error: "load failed", BuiltAt: time.Now(),
	}

	result := recordNginxModuleBuildFailure(original, "example", failed, &ready)

	if len(result[0].Builds) != 1 || result[0].Builds[0].Hash != "ready" {
		t.Fatalf("previous ready build was replaced: %#v", result[0].Builds)
	}
	if result[0].LastError != failed.Error {
		t.Fatalf("failure metadata was not retained: %#v", result[0])
	}
	result[0].Builds[0].Hash = "mutated"
	if original[0].Builds[0].Hash != "ready" {
		t.Fatal("module clone shares build state with the original")
	}
}

func TestHasDynamicNginxModuleBuildTask(t *testing.T) {
	dynamicEnabled := dto.NginxModule{Name: "brotli", Enable: true, BuildMode: nginxModuleBuildDynamic}
	staticEnabled := dto.NginxModule{Name: "pagespeed", Enable: true, BuildMode: nginxModuleBuildStatic}
	disabledDynamic := dto.NginxModule{Name: "waf", Enable: false, BuildMode: nginxModuleBuildDynamic}

	if hasDynamicNginxModuleBuildTask(nil, nil) {
		t.Fatal("empty module list should not require a dynamic build")
	}
	if hasDynamicNginxModuleBuildTask([]dto.NginxModule{staticEnabled}, nil) {
		t.Fatal("static-only modules should not require a dynamic build")
	}
	if hasDynamicNginxModuleBuildTask([]dto.NginxModule{dynamicEnabled}, []string{"other"}) {
		t.Fatal("enabled module outside the selection should not require a dynamic build")
	}
	if !hasDynamicNginxModuleBuildTask([]dto.NginxModule{disabledDynamic}, []string{"waf"}) {
		t.Fatal("selected module should require a dynamic build even when disabled")
	}
	if !hasDynamicNginxModuleBuildTask([]dto.NginxModule{dynamicEnabled, staticEnabled}, nil) {
		t.Fatal("enabled dynamic module should require a dynamic build")
	}

	legacy := dto.NginxModule{Name: "legacy", Enable: true}
	if hasDynamicNginxModuleBuildTask([]dto.NginxModule{legacy}, nil) {
		t.Fatal("legacy module without a build mode stays static and should not require a dynamic build")
	}
	if legacy.BuildMode != "" {
		t.Fatalf("prescan must normalize a copy, got mutated BuildMode %q", legacy.BuildMode)
	}
}

func TestNginxModuleStaticBuildErrorHintOnlyForCustomModules(t *testing.T) {
	if hint := nginxModuleStaticBuildErrorHint(dto.NginxModule{Name: "builtin"}); hint != "" {
		t.Fatalf("built-in module cannot switch build mode, got hint %q", hint)
	}
	if hint := nginxModuleStaticBuildErrorHint(dto.NginxModule{Name: "custom", Custom: true}); hint == "" {
		t.Fatal("custom module should receive the static-build alternative")
	}
}

func TestResolveNginxModuleTargetWithoutBuilder(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.27.1.2"}
	install.App.Key = constant.AppOpenresty

	_, _, err := resolveNginxModuleTarget(install)
	if !errors.Is(err, errNginxModuleBuilderMissing) {
		t.Fatalf("expected builder-missing sentinel, got %v", err)
	}
	if !strings.Contains(err.Error(), "dynamic module builder not found") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestBuildDynamicNginxModulesFailsWithoutBuilder(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty
	modules := []dto.NginxModule{{
		Name: "rtmp", Enable: true, BuildMode: nginxModuleBuildDynamic,
	}}

	_, err := buildDynamicNginxModules(install, modules, nil, false, "", "", nil)
	if !errors.Is(err, errNginxModuleBuilderMissing) {
		t.Fatalf("missing target builder must fail the dynamic build, got %v", err)
	}
}

func TestMergeOpenrestyModuleVolumes(t *testing.T) {
	newService := map[string]interface{}{}
	oldService := map[string]interface{}{
		"volumes": []interface{}{
			"./conf:/etc/nginx/conf.d:ro",
			"./modules:/usr/local/openresty/nginx/modules/1panel:ro",
			"./conf/modules-enabled:/usr/local/openresty/nginx/conf/modules-enabled:ro",
			12345,
		},
	}

	mergeOpenrestyModuleVolumes(newService, oldService)

	merged, ok := newService["volumes"].([]interface{})
	if !ok || len(merged) != 2 {
		t.Fatalf("expected two module mounts to be merged, got %#v", newService["volumes"])
	}
	for _, volume := range merged {
		if volume == "./conf:/etc/nginx/conf.d:ro" {
			t.Fatal("unrelated mount should not be merged")
		}
	}
}

func TestMergeOpenrestyModuleVolumesKeepsExisting(t *testing.T) {
	existing := "./modules:/usr/local/openresty/nginx/modules/1panel:ro"
	newService := map[string]interface{}{
		"volumes": []interface{}{existing, map[string]interface{}{"type": "bind"}},
	}
	oldService := map[string]interface{}{
		"volumes": []interface{}{
			"./modules:/usr/local/openresty/nginx/modules/1panel:ro",
			"./conf/modules-enabled:/usr/local/openresty/nginx/conf/modules-enabled:ro",
		},
	}

	mergeOpenrestyModuleVolumes(newService, oldService)

	merged, ok := newService["volumes"].([]interface{})
	if !ok || len(merged) != 3 {
		t.Fatalf("expected only the missing mount to be appended, got %#v", newService["volumes"])
	}
	if merged[0] != existing {
		t.Fatalf("existing mounts must keep their order, got %#v", merged)
	}
	if merged[2] != "./conf/modules-enabled:/usr/local/openresty/nginx/conf/modules-enabled:ro" {
		t.Fatalf("missing module mount was not appended, got %#v", merged)
	}
}

func TestResolveNginxModuleBuildMirror(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.27.1.2"}
	install.App.Key = constant.AppOpenresty
	if err := os.MkdirAll(install.GetPath(), constant.DirPerm); err != nil {
		t.Fatal(err)
	}

	if got := resolveNginxModuleBuildMirror(install, "https://mirror.example.com"); got != "https://mirror.example.com" {
		t.Fatalf("request mirror should win, got %q", got)
	}
	if got := resolveNginxModuleBuildMirror(install, ""); got != "" {
		t.Fatalf("missing env file should yield an empty mirror, got %q", got)
	}

	envContent := "CONTAINER_PACKAGE_URL=https://apt.example.com\nOTHER=value\n"
	if err := os.WriteFile(install.GetEnvPath(), []byte(envContent), constant.FilePerm); err != nil {
		t.Fatal(err)
	}
	if got := resolveNginxModuleBuildMirror(install, ""); got != "https://apt.example.com" {
		t.Fatalf("env CONTAINER_PACKAGE_URL should be the fallback, got %q", got)
	}
	if got := resolveNginxModuleBuildMirror(install, "https://mirror.example.com"); got != "https://mirror.example.com" {
		t.Fatalf("request mirror should still win over the env value, got %q", got)
	}
}

func writeNginxModuleCatalogStateFixture(t *testing.T, install model.AppInstall, catalog []dto.NginxModule, state string) {
	t.Helper()
	buildDir := path.Join(install.GetPath(), nginxModuleBuildDir)
	if err := os.MkdirAll(buildDir, constant.DirPerm); err != nil {
		t.Fatal(err)
	}
	catalogContent, err := json.Marshal(catalog)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(path.Join(buildDir, nginxModuleCatalogFile), catalogContent, constant.FilePerm); err != nil {
		t.Fatal(err)
	}
	if state != "" {
		if err = os.WriteFile(path.Join(buildDir, nginxModuleStoreFile), []byte(state), constant.FilePerm); err != nil {
			t.Fatal(err)
		}
	}
}

func TestLoadNginxModulesBuiltinStateCannotOverrideCatalogDefinition(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	catalog := []dto.NginxModule{{
		Name: "rtmp", Script: "catalog-script", Packages: []string{"unzip"}, Params: "--add-module=/tmp/rtmp",
		BuildMode: nginxModuleBuildDynamic, Provider: nginxModuleProviderLocal, LoadOrder: 20,
	}}
	state := `[{"name":"rtmp","enable":true,"script":"user-script","params":"--with-user","buildMode":"static","loadOrder":99,"lastError":"failed"}]`
	writeNginxModuleCatalogStateFixture(t, install, catalog, state)

	modules, err := loadNginxModules(install)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != 1 {
		t.Fatalf("expected one merged module, got %#v", modules)
	}
	module := modules[0]
	if !module.Enable || module.LastError != "failed" {
		t.Fatalf("builtin state was not applied: %#v", module)
	}
	if module.Script != "catalog-script" || module.Params != "--add-module=/tmp/rtmp" ||
		module.BuildMode != nginxModuleBuildDynamic || module.LoadOrder != 20 {
		t.Fatalf("builtin definition must come from catalog: %#v", module)
	}
}

func TestLoadNginxModulesWithoutStateShowsDisabledCatalogModules(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	catalog := []dto.NginxModule{
		{Name: "rtmp", Enable: true, BuildMode: nginxModuleBuildDynamic},
		{Name: "geoip2", BuildMode: nginxModuleBuildDynamic},
	}
	writeNginxModuleCatalogStateFixture(t, install, catalog, "")

	modules, err := loadNginxModules(install)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != len(catalog) {
		t.Fatalf("expected all catalog modules, got %#v", modules)
	}
	for _, module := range modules {
		if module.Enable || module.Custom {
			t.Fatalf("catalog modules must default to disabled built-ins: %#v", module)
		}
	}
}

func TestSaveNginxModulesPersistsOnlyBuiltinState(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	catalog := []dto.NginxModule{{
		Name: "rtmp", Script: "catalog-script", Params: "--add-module=/tmp/rtmp",
		BuildMode: nginxModuleBuildDynamic, Provider: nginxModuleProviderLocal, LoadOrder: 20,
	}}
	writeNginxModuleCatalogStateFixture(t, install, catalog, "")
	modules, err := loadNginxModules(install)
	if err != nil {
		t.Fatal(err)
	}
	modules[0].Enable = true
	modules[0].LastError = "failed"

	if err = saveNginxModules(install, modules); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleStoreFile))
	if err != nil {
		t.Fatal(err)
	}
	var states []map[string]any
	if err = json.Unmarshal(content, &states); err != nil {
		t.Fatal(err)
	}
	if len(states) != 1 || states[0]["name"] != "rtmp" || states[0]["enable"] != true || states[0]["lastError"] != "failed" {
		t.Fatalf("unexpected builtin state: %s", content)
	}
	for _, key := range []string{"script", "packages", "params", "buildMode", "provider", "loadOrder"} {
		if _, exists := states[0][key]; exists {
			t.Fatalf("builtin definition field %q leaked into module state: %s", key, content)
		}
	}
}

func TestSaveNginxModulesOmitsPristineBuiltinState(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty
	catalog := []dto.NginxModule{{Name: "rtmp", BuildMode: nginxModuleBuildDynamic}}
	writeNginxModuleCatalogStateFixture(t, install, catalog, "")

	modules, err := loadNginxModules(install)
	if err != nil {
		t.Fatal(err)
	}
	if err = saveNginxModules(install, modules); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleStoreFile))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(content)) != "[]" {
		t.Fatalf("pristine built-in state should not be persisted: %s", content)
	}
}

func TestLoadAndSaveNginxCustomModule(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	state := `[{"name":"custom","custom":true,"script":"prepare","packages":["git"],"params":"--add-module=/tmp/custom","enable":true,"buildMode":"static","provider":"local","loadOrder":100}]`
	writeNginxModuleCatalogStateFixture(t, install, nil, state)

	modules, err := loadNginxModules(install)
	if err != nil {
		t.Fatal(err)
	}
	if len(modules) != 1 || !modules[0].Custom {
		t.Fatalf("custom module source was not restored: %#v", modules)
	}
	if modules[0].Script != "prepare" || modules[0].BuildMode != nginxModuleBuildStatic || modules[0].LoadOrder != 100 {
		t.Fatalf("custom module definition was not restored: %#v", modules[0])
	}

	if err = saveNginxModules(install, modules); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(path.Join(install.GetPath(), nginxModuleBuildDir, nginxModuleStoreFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), `"custom": true`) || !strings.Contains(string(content), `"buildMode": "static"`) {
		t.Fatalf("custom module definition was not persisted: %s", content)
	}
}

func TestActivateNginxModuleCatalogReplacesCatalogAtomically(t *testing.T) {
	buildDir := t.TempDir()
	activePath := path.Join(buildDir, nginxModuleCatalogFile)
	pendingPath := path.Join(buildDir, nginxModuleCatalogPendingFile)
	sourcePath := path.Join(buildDir, "target.catalog.json")
	if err := os.WriteFile(activePath, []byte(`[{"name":"old"}]`), constant.FilePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(sourcePath, []byte(`[{"name":"new"}]`), constant.FilePerm); err != nil {
		t.Fatal(err)
	}

	if err := stageNginxModuleCatalog(sourcePath, pendingPath); err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(activePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != `[{"name":"old"}]` {
		t.Fatalf("staging target catalog changed the active catalog: %s", content)
	}
	if err := activateNginxModuleCatalog(pendingPath, activePath); err != nil {
		t.Fatal(err)
	}
	content, err = os.ReadFile(activePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != `[{"name":"new"}]` {
		t.Fatalf("active catalog was not replaced: %s", content)
	}
	if _, err = os.Stat(pendingPath); !os.IsNotExist(err) {
		t.Fatalf("pending catalog should be consumed, got %v", err)
	}
}

func TestActivateNginxModuleCatalogRestoresCatalogWhenCommitFails(t *testing.T) {
	buildDir := t.TempDir()
	activePath := path.Join(buildDir, nginxModuleCatalogFile)
	pendingPath := path.Join(buildDir, nginxModuleCatalogPendingFile)
	if err := os.WriteFile(activePath, []byte(`[{"name":"old"}]`), constant.FilePerm); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pendingPath, []byte(`[{"name":"new"}]`), constant.FilePerm); err != nil {
		t.Fatal(err)
	}
	commitErr := errors.New("save install failed")

	err := activateNginxModuleCatalogAndCommit(pendingPath, activePath, func() error {
		return commitErr
	})
	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error, got %v", err)
	}
	content, readErr := os.ReadFile(activePath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(content) != `[{"name":"old"}]` {
		t.Fatalf("failed upgrade must restore the old catalog: %s", content)
	}
}

func TestLoadNginxModulesRejectsDuplicateCatalogNames(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	catalog := []dto.NginxModule{
		{Name: "rtmp", Params: "--add-module=/tmp/one", BuildMode: nginxModuleBuildDynamic},
		{Name: "rtmp", Params: "--add-module=/tmp/two", BuildMode: nginxModuleBuildDynamic},
	}
	writeNginxModuleCatalogStateFixture(t, install, catalog, "")

	if _, err := loadNginxModules(install); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate catalog name error, got %v", err)
	}
}

func TestLoadNginxModulesRejectsDuplicateStateNames(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	catalog := []dto.NginxModule{{Name: "rtmp", Params: "--add-module=/tmp/rtmp", BuildMode: nginxModuleBuildDynamic}}
	writeNginxModuleCatalogStateFixture(t, install, catalog, `[{"name":"rtmp","enable":true},{"name":"rtmp","enable":false}]`)

	if _, err := loadNginxModules(install); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate state name error, got %v", err)
	}
}

func TestLoadNginxModulesRejectsOrphanBuiltinState(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	writeNginxModuleCatalogStateFixture(t, install, nil, `[{"name":"removed","enable":true}]`)

	if _, err := loadNginxModules(install); err == nil || !strings.Contains(err.Error(), "missing from the module catalog") {
		t.Fatalf("expected orphan builtin state error, got %v", err)
	}
}

func TestLoadNginxModulesRejectsCustomCatalogNameConflict(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	catalog := []dto.NginxModule{{Name: "rtmp", BuildMode: nginxModuleBuildDynamic}}
	writeNginxModuleCatalogStateFixture(t, install, catalog, `[{"name":"rtmp","custom":true,"enable":true}]`)

	if _, err := loadNginxModules(install); err == nil || !strings.Contains(err.Error(), "conflicts") {
		t.Fatalf("expected custom/catalog conflict error, got %v", err)
	}
}

func TestLoadNginxModulesRejectsInvalidBuildModes(t *testing.T) {
	oldDir := global.Dir.AppInstallDir
	global.Dir.AppInstallDir = t.TempDir()
	t.Cleanup(func() { global.Dir.AppInstallDir = oldDir })
	install := model.AppInstall{Name: "openresty", Version: "1.31.1.1"}
	install.App.Key = constant.AppOpenresty

	writeNginxModuleCatalogStateFixture(t, install, []dto.NginxModule{{
		Name: "rtmp", BuildMode: "auto",
	}}, "")
	if _, err := loadNginxModules(install); err == nil || !strings.Contains(err.Error(), "invalid build mode") {
		t.Fatalf("expected invalid catalog build mode error, got %v", err)
	}

	writeNginxModuleCatalogStateFixture(t, install, nil, `[{"name":"custom","custom":true,"buildMode":"auto"}]`)
	if _, err := loadNginxModules(install); err == nil || !strings.Contains(err.Error(), "invalid build mode") {
		t.Fatalf("expected invalid custom build mode error, got %v", err)
	}
}

func TestApplyNginxModuleUpdateKeepsBuiltinDefinitionImmutable(t *testing.T) {
	modules := []dto.NginxModule{{
		Name: "rtmp", Script: "catalog-script", Params: "--add-module=/tmp/rtmp",
		BuildMode: nginxModuleBuildDynamic, Provider: nginxModuleProviderLocal, LoadOrder: 20,
	}}
	req := request.NginxModuleUpdate{
		Operate: nginxModuleOperateUpdate, Name: "rtmp", Enable: true, Script: "user-script",
		Params: "--with-user", BuildMode: nginxModuleBuildStatic, Provider: "prebuilt", LoadOrder: 99,
	}

	updated, _, err := applyNginxModuleUpdate(modules, req)
	if err != nil {
		t.Fatal(err)
	}
	if !updated[0].Enable {
		t.Fatal("builtin enable state was not updated")
	}
	if updated[0].Script != "catalog-script" || updated[0].Params != "--add-module=/tmp/rtmp" ||
		updated[0].BuildMode != nginxModuleBuildDynamic || updated[0].Provider != nginxModuleProviderLocal ||
		updated[0].LoadOrder != 20 {
		t.Fatalf("builtin definition was modified: %#v", updated[0])
	}
}

func TestApplyNginxModuleUpdateRejectsBuiltinDelete(t *testing.T) {
	modules := []dto.NginxModule{{Name: "rtmp", BuildMode: nginxModuleBuildDynamic}}

	_, _, err := applyNginxModuleUpdate(modules, request.NginxModuleUpdate{
		Operate: nginxModuleOperateDelete, Name: "rtmp",
	})
	if err == nil || !strings.Contains(err.Error(), "cannot be deleted") {
		t.Fatalf("expected builtin delete error, got %v", err)
	}
}

func TestApplyNginxModuleUpdateCreatesAndDeletesCustomModule(t *testing.T) {
	if _, _, err := applyNginxModuleUpdate(nil, request.NginxModuleUpdate{
		Operate: nginxModuleOperateCreate, Name: "invalid", BuildMode: "auto",
	}); err == nil || !strings.Contains(err.Error(), "invalid build mode") {
		t.Fatalf("expected invalid custom build mode error, got %v", err)
	}

	created, _, err := applyNginxModuleUpdate(nil, request.NginxModuleUpdate{
		Operate: nginxModuleOperateCreate, Name: "custom", Script: "prepare", Packages: "git,curl",
		Params: "--add-module=/tmp/custom", Enable: true, BuildMode: nginxModuleBuildStatic,
		Provider: nginxModuleProviderLocal, LoadOrder: 100,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created) != 1 || !created[0].Custom || created[0].BuildMode != nginxModuleBuildStatic {
		t.Fatalf("custom module was not created correctly: %#v", created)
	}
	updated, _, err := applyNginxModuleUpdate(created, request.NginxModuleUpdate{
		Operate: nginxModuleOperateUpdate, Name: "custom", Script: "updated",
		Params: "--add-module=/tmp/custom-v2", Enable: false, BuildMode: nginxModuleBuildDynamic,
		Provider: nginxModuleProviderLocal, LoadOrder: 75,
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated[0].Script != "updated" || updated[0].BuildMode != nginxModuleBuildDynamic || updated[0].LoadOrder != 75 {
		t.Fatalf("custom module was not updated correctly: %#v", updated[0])
	}

	remaining, deleted, err := applyNginxModuleUpdate(updated, request.NginxModuleUpdate{
		Operate: nginxModuleOperateDelete, Name: "custom",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(remaining) != 0 || deleted == nil || deleted.Name != "custom" {
		t.Fatalf("custom module was not removed: remaining=%#v deleted=%#v", remaining, deleted)
	}
}
