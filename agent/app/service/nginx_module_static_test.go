package service

import (
	"strings"
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

func TestNormalizeStaticModuleParams(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "dynamic add-module is reversed",
			in:   "--add-dynamic-module=/tmp/ngx_http_geoip2_module",
			want: "--add-module=/tmp/ngx_http_geoip2_module",
		},
		{
			name: "already static params are left alone",
			in:   "--add-module=/usr/local/openresty/modules/ngx_brotli",
			want: "--add-module=/usr/local/openresty/modules/ngx_brotli",
		},
		{
			name: "built-in switches drop the dynamic suffix",
			in:   "--with-http_image_filter_module=dynamic",
			want: "--with-http_image_filter_module",
		},
		{
			name: "mixed options are handled together",
			in:   "--with-http_dav_module --add-dynamic-module=/tmp/nginx-dav-ext-module",
			want: "--with-http_dav_module --add-module=/tmp/nginx-dav-ext-module",
		},
		{
			name: "surrounding whitespace is trimmed",
			in:   "  --add-dynamic-module=/tmp/x  ",
			want: "--add-module=/tmp/x",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeStaticModuleParams(tc.in); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

// The conversion must round-trip the params the dynamic path produces,
// otherwise an install without a builder would compile the wrong thing.
func TestStaticParamsRoundTripFromDynamic(t *testing.T) {
	// These are the catalog params shipped with the OpenResty app.
	for _, original := range []string{
		"--add-module=/usr/local/openresty/modules/ngx_brotli",
		"--add-module=/tmp/nginx-rtmp-module",
		"--with-http_dav_module --add-module=/tmp/nginx-dav-ext-module",
		"--add-module=/tmp/ngx_http_geoip2_module",
		"--add-module=/tmp/ngx_http_substitutions_filter_module",
	} {
		dynamic, err := normalizeDynamicModuleParams(original)
		if err != nil {
			t.Fatalf("%s: %v", original, err)
		}
		if got := normalizeStaticModuleParams(dynamic); got != original {
			t.Errorf("round trip changed the options:\n  from %q\n  via  %q\n  to   %q", original, dynamic, got)
		}
	}
}

func TestConvertNginxModulesToStatic(t *testing.T) {
	modules := []dto.NginxModule{
		{Name: "ngx_brotli", Enable: true, BuildMode: nginxModuleBuildDynamic,
			Params: "--add-dynamic-module=/usr/local/openresty/modules/ngx_brotli"},
		{Name: "rtmp", Enable: false, BuildMode: nginxModuleBuildDynamic,
			Params: "--add-dynamic-module=/tmp/nginx-rtmp-module"},
	}
	converted, err := convertNginxModulesToStatic(modules, nil)
	if err != nil {
		t.Fatal(err)
	}
	if converted[0].BuildMode != nginxModuleBuildStatic {
		t.Error("an enabled module should be retargeted to static")
	}
	if !strings.Contains(converted[0].Params, "--add-module=") ||
		strings.Contains(converted[0].Params, "--add-dynamic-module=") {
		t.Errorf("params were not converted: %q", converted[0].Params)
	}
	if converted[1].BuildMode != nginxModuleBuildDynamic {
		t.Error("a disabled module must be left alone")
	}

	// The caller's slice must survive untouched: the conversion drives a single
	// build and is never written back to module.json.
	if modules[0].BuildMode != nginxModuleBuildDynamic {
		t.Error("the input slice was mutated")
	}
	if modules[0].Params != "--add-dynamic-module=/usr/local/openresty/modules/ngx_brotli" {
		t.Errorf("input params were mutated: %q", modules[0].Params)
	}
}

func TestConvertNginxModulesToStaticHonoursSelection(t *testing.T) {
	modules := []dto.NginxModule{
		{Name: "ngx_brotli", Enable: true, BuildMode: nginxModuleBuildDynamic,
			Params: "--add-dynamic-module=/a"},
		{Name: "rtmp", Enable: true, BuildMode: nginxModuleBuildDynamic,
			Params: "--add-dynamic-module=/b"},
	}
	converted, err := convertNginxModulesToStatic(modules, []string{"rtmp"})
	if err != nil {
		t.Fatal(err)
	}
	if converted[0].BuildMode != nginxModuleBuildDynamic {
		t.Error("an unselected module must not be converted")
	}
	if converted[1].BuildMode != nginxModuleBuildStatic {
		t.Error("the selected module should be converted")
	}
}

func TestConvertNginxModulesToStaticRejectsEmptyParams(t *testing.T) {
	modules := []dto.NginxModule{
		{Name: "broken", Enable: true, BuildMode: nginxModuleBuildDynamic, Params: "   "},
	}
	if _, err := convertNginxModulesToStatic(modules, nil); err == nil {
		t.Fatal("a module with no configure options cannot be compiled in")
	}
}
