package service

import (
	"path"

	"github.com/1Panel-dev/1Panel/agent/app/dto"
	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/constant"
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
//
// Values the user changed through the compression settings page are read back
// from the current managed file, so reconciling after an unrelated module
// change does not silently reset them to the defaults.
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
		current := readNginxHTTPDirectives(path.Join(nginxHTTPConfigDir(install), fileName))
		desired[fileName] = renderNginxHTTPConfig(mergeNginxRuntimeDirectives(directives, current))
	}
	return desired
}

// mergeNginxRuntimeDirectives keeps the declared directive set and ordering
// while preferring values already present in the managed file.
func mergeNginxRuntimeDirectives(defaults []nginxHTTPDirective, current map[string][]string) []nginxHTTPDirective {
	if len(current) == 0 {
		return defaults
	}
	merged := make([]nginxHTTPDirective, 0, len(defaults))
	for _, directive := range defaults {
		if params, ok := current[directive.Name]; ok && len(params) > 0 {
			directive.Params = params
		}
		merged = append(merged, directive)
	}
	return merged
}

// nginxBrotliModuleName is the catalog name of the brotli module.
const nginxBrotliModuleName = "ngx_brotli"

// getNginxBrotliParams reports the brotli settings currently in effect.
//
// Brotli is served from the managed http.d file instead of nginx.conf, so the
// directives can be removed together with the module. When the module is
// disabled the declared defaults are returned, which lets the settings page
// show what would be applied once it is enabled.
func getNginxBrotliParams() ([]response.NginxParam, error) {
	install, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	fileName := nginxHTTPConfigFileName(nginxModuleRuntimeOrder(nginxBrotliModuleName), nginxBrotliModuleName)
	current := readNginxHTTPDirectives(path.Join(nginxHTTPConfigDir(install), fileName))
	res := make([]response.NginxParam, 0, len(dto.BrotliKeys))
	for _, directive := range mergeNginxRuntimeDirectives(nginxModuleRuntimeDefaults[nginxBrotliModuleName], current) {
		res = append(res, response.NginxParam{Name: directive.Name, Params: directive.Params})
	}
	return res, nil
}

// updateNginxBrotliParams persists brotli settings to the managed http.d file.
//
// Writing is refused unless the module is enabled and built: the directives
// would reference a module that is not loaded and nginx would fail to start.
func updateNginxBrotliParams(params []dto.NginxParam) error {
	install, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return err
	}
	if !nginxHTTPConfigSupported(install) {
		return buserr.New("ErrBrotliUnsupported")
	}
	modules, err := loadNginxModules(install)
	if err != nil {
		return err
	}
	values := make(map[string][]string, len(params))
	for _, param := range params {
		values[param.Name] = param.Params
	}
	for i := range modules {
		if modules[i].Name != nginxBrotliModuleName {
			continue
		}
		if !modules[i].Enable {
			return buserr.New("ErrBrotliDisabled")
		}
		fileName := nginxHTTPConfigFileName(nginxModuleRuntimeOrder(nginxBrotliModuleName), nginxBrotliModuleName)
		configDir := nginxHTTPConfigDir(install)
		snapshot, snapErr := snapshotManagedNginxHTTPConfigs(configDir)
		if snapErr != nil {
			return snapErr
		}
		merged := mergeNginxRuntimeDirectives(nginxModuleRuntimeDefaults[nginxBrotliModuleName], values)
		desired := map[string][]byte{fileName: renderNginxHTTPConfig(merged)}
		for name, content := range snapshot {
			if name != fileName {
				desired[name] = content
			}
		}
		if err = applyManagedNginxHTTPConfigs(configDir, desired); err != nil {
			_ = applyManagedNginxHTTPConfigs(configDir, snapshot)
			return err
		}
		if err = opNginx(install.ContainerName, constant.NginxCheck); err != nil {
			_ = applyManagedNginxHTTPConfigs(configDir, snapshot)
			return err
		}
		if err = opNginx(install.ContainerName, constant.NginxReload); err != nil {
			_ = applyManagedNginxHTTPConfigs(configDir, snapshot)
			return err
		}
		return nil
	}
	return buserr.New("ErrBrotliDisabled")
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
