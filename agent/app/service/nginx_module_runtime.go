package service

import (
	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/model"
)

// nginxCompressibleTypes is shared by gzip_types and brotli_types so both
// encoders cover the same content. Already compressed formats (images other
// than SVG, woff/woff2, archives, media) are deliberately excluded:
// recompressing them costs CPU and usually grows the payload.
var nginxCompressibleTypes = []string{
	"text/plain",
	"text/css",
	"text/xml",
	"text/javascript",
	"application/json",
	"application/ld+json",
	"application/javascript",
	"application/x-javascript",
	"application/xml",
	"application/xhtml+xml",
	"application/rss+xml",
	"application/atom+xml",
	"application/wasm",
	"image/svg+xml",
	"font/ttf",
	"font/otf",
}

// nginxModuleRuntimeDefaults maps a module to the http-context directives that
// make it actually do something once loaded. Without these, enabling a module
// only emits load_module, leaving it loaded but inert.
//
// brotli_static is intentionally omitted: nginx does not verify that a .br
// file is newer than its source, so a stale artifact would be served
// indefinitely with no error.
var nginxModuleRuntimeDefaults = map[string][]nginxHTTPDirective{
	"ngx_brotli": {
		{Name: "brotli", Params: []string{"on"}},
		// Brotli level 5 reaches roughly gzip level 9 ratio at a fraction of
		// the cost. The nginx default of 6 is tuned for static assets and is
		// too expensive for dynamic responses.
		{Name: "brotli_comp_level", Params: []string{"5"}},
		{Name: "brotli_min_length", Params: []string{"1k"}},
		{Name: "brotli_types", Params: nginxCompressibleTypes},
	},
}

// nginxModuleRuntimeLoadOrder keeps managed file names stable and ordered
// independently of the module load order used for load_module.
var nginxModuleRuntimeLoadOrder = map[string]int{
	"ngx_brotli": 100,
}

func nginxModuleRuntimeOrder(name string) int {
	if order, ok := nginxModuleRuntimeLoadOrder[name]; ok {
		return order
	}
	return 900
}

// desiredNginxModuleRuntimeConfigs renders the managed http.d files for every
// enabled module that has a ready build and known runtime defaults.
func desiredNginxModuleRuntimeConfigs(install model.AppInstall, modules []dto.NginxModule, target dto.NginxModuleTarget) map[string][]byte {
	desired := make(map[string][]byte)
	for _, module := range modules {
		normalizeNginxModule(&module)
		directives, ok := nginxModuleRuntimeDefaults[module.Name]
		if !ok || !module.Enable {
			continue
		}
		if !nginxModuleRuntimeReady(module, target) {
			continue
		}
		fileName := nginxHTTPConfigFileName(nginxModuleRuntimeOrder(module.Name), module.Name)
		desired[fileName] = renderNginxHTTPConfig(directives)
	}
	return desired
}

// nginxModuleRuntimeReady reports whether the module is actually usable.
//
// Dynamic modules need a ready build for the current target, otherwise the
// .so is missing and nginx would reject the directives. Static modules are
// compiled into the binary and carry no artifacts, so an enabled static
// module is considered ready.
func nginxModuleRuntimeReady(module dto.NginxModule, target dto.NginxModuleTarget) bool {
	if module.BuildMode == nginxModuleBuildStatic {
		return true
	}
	build := findCurrentNginxModuleBuild(module, target)
	if build == nil || build.Status != nginxModuleStatusReady {
		build = findLatestNginxModuleBuild(module, target)
	}
	return build != nil && build.Status == nginxModuleStatusReady
}
