package service

import (
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
)

func TestHasEnabledStaticNginxModules(t *testing.T) {
	cases := []struct {
		name    string
		modules []dto.NginxModule
		want    bool
	}{
		{
			name: "an enabled static module requires a rebuild",
			modules: []dto.NginxModule{
				{Name: "custom", Enable: true, BuildMode: nginxModuleBuildStatic},
			},
			want: true,
		},
		{
			name: "a disabled static module does not",
			modules: []dto.NginxModule{
				{Name: "custom", Enable: false, BuildMode: nginxModuleBuildStatic},
			},
			want: false,
		},
		{
			name: "dynamic modules never require a rebuild",
			modules: []dto.NginxModule{
				{Name: "ngx_brotli", Enable: true, BuildMode: nginxModuleBuildDynamic},
			},
			want: false,
		},
		{
			name:    "no modules at all",
			modules: nil,
			want:    false,
		},
		{
			name: "one enabled static module among dynamic ones is enough",
			modules: []dto.NginxModule{
				{Name: "ngx_brotli", Enable: true, BuildMode: nginxModuleBuildDynamic},
				{Name: "custom", Enable: true, BuildMode: nginxModuleBuildStatic},
			},
			want: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasEnabledStaticNginxModules(tc.modules); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// A stale RESTY_CONFIG_OPTIONS_MORE used to force the full-rebuild path even
// with no static module enabled. configureStaticNginxModules derives that value
// from the module list and runs before every build, so the rebuild it triggered
// could only ever reproduce the current image. Module state is now the only
// input; this test pins that down.
func TestStaticRebuildIgnoresLeftoverBuildOptions(t *testing.T) {
	modules := []dto.NginxModule{
		{Name: "ngx_brotli", Enable: true, BuildMode: nginxModuleBuildDynamic},
	}
	if hasEnabledStaticNginxModules(modules) {
		t.Fatal("dynamic-only modules must not select the static build path")
	}
}

// normalizeNginxModule is applied to a copy, so callers keep their entities.
func TestHasEnabledStaticNginxModulesDoesNotMutateInput(t *testing.T) {
	modules := []dto.NginxModule{
		{Name: "custom", Enable: true, BuildMode: nginxModuleBuildStatic, Packages: []string{"", "libfoo", ""}},
	}
	_ = hasEnabledStaticNginxModules(modules)
	if len(modules[0].Packages) != 3 {
		t.Fatalf("input was normalized in place: %v", modules[0].Packages)
	}
}

