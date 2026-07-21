package service

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

func TestNormalizeNginxModulePreservesLegacyStaticMode(t *testing.T) {
	module := dto.NginxModule{
		Name:     "legacy",
		Packages: []string{"git", "", "git", " curl "},
	}

	normalizeNginxModule(&module)

	if module.BuildMode != nginxModuleBuildStatic {
		t.Fatalf("expected legacy module to remain static, got %s", module.BuildMode)
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
		Name: "example", BuildMode: nginxModuleBuildDynamic, DynamicSupport: nginxModuleSupportUnknown,
		Builds: []dto.NginxModuleBuild{ready},
	}}
	failed := dto.NginxModuleBuild{
		Hash: "candidate", Status: nginxModuleStatusFailed, Target: target, Error: "load failed", BuiltAt: time.Now(),
	}

	result := recordNginxModuleBuildFailure(original, "example", failed, &ready, true)

	if len(result[0].Builds) != 1 || result[0].Builds[0].Hash != "ready" {
		t.Fatalf("previous ready build was replaced: %#v", result[0].Builds)
	}
	if result[0].LastError != failed.Error || result[0].DynamicSupport != nginxModuleSupportSupported {
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
	deletedDynamic := dto.NginxModule{Name: "geoip", Enable: true, BuildMode: nginxModuleBuildDynamic, Deleted: true}
	disabledDynamic := dto.NginxModule{Name: "waf", Enable: false, BuildMode: nginxModuleBuildDynamic}

	if hasDynamicNginxModuleBuildTask(nil, nil) {
		t.Fatal("empty module list should not require a dynamic build")
	}
	if hasDynamicNginxModuleBuildTask([]dto.NginxModule{staticEnabled}, nil) {
		t.Fatal("static-only modules should not require a dynamic build")
	}
	if hasDynamicNginxModuleBuildTask([]dto.NginxModule{deletedDynamic}, nil) {
		t.Fatal("deleted modules should not require a dynamic build")
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
