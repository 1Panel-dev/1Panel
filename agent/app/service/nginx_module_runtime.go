package service

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

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
//
// A module the user already configured by hand in nginx.conf is skipped
// entirely. Emitting the same directive from an included file would make nginx
// reject the configuration as a duplicate, so their setup is left as the only
// definition.
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
		if nginxModuleConfiguredByUser(install, module.Name) {
			continue
		}
		fileName := nginxHTTPConfigFileName(nginxModuleRuntimeOrder(module.Name), module.Name)
		current := readNginxHTTPDirectives(path.Join(nginxHTTPConfigDir(install), fileName))
		desired[fileName] = renderNginxHTTPConfig(mergeNginxRuntimeDirectives(directives, current))
	}
	return desired
}

// nginxModuleConfiguredByUser reports whether the user already manages any of
// the module's directives by hand.
//
// Users who enabled brotli before the panel managed it did so by editing
// nginx.conf or a file it includes. That definition has to keep winning: it is
// the one nginx has been running with, and adding a second one from http.d
// would break the configuration outright.
//
// Any brotli* directive counts, not just the primary one. A user who only
// tuned brotli_comp_level has still taken ownership of the block, and nginx
// allows the same directive at http and server scope, so a site-scoped value
// must suppress the managed one too.
func nginxModuleConfiguredByUser(install model.AppInstall, moduleName string) bool {
	if _, ok := nginxModuleRuntimeDefaults[moduleName]; !ok {
		return false
	}
	for _, filePath := range nginxModuleUserConfigPaths(install) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		if nginxModuleUserDirectiveRe.MatchString(string(content)) {
			return true
		}
	}
	return false
}

// nginxModuleUserDirectiveRe matches any active (non-commented) brotli*
// directive at the start of a line, wherever it was written.
var nginxModuleUserDirectiveRe = regexp.MustCompile(`(?m)^[ \t]*brotli[a-z_]*[ \t]+[^;\n]*;`)

// nginxModuleUserConfigPaths lists the files that may carry a user's brotli
// configuration: the main config and the http-scope files it includes. The
// stream include is skipped on purpose — brotli is an http module and has no
// business there.
func nginxModuleUserConfigPaths(install model.AppInstall) []string {
	websiteDir := GetWebSiteRootDir()
	return append(
		[]string{nginxMainConfigPath(install)},
		globConfFiles(path.Join(websiteDir, "conf.d"))...,
	)
}

func globConfFiles(dir string) []string {
	matches, err := filepath.Glob(path.Join(dir, "*.conf"))
	if err != nil {
		return nil
	}
	return matches
}

// nginxUserDirectivePattern matches a directive the user wrote in nginx.conf,
// capturing its indentation so a rewrite can keep the line's shape. Leading
// whitespace only, so a commented-out line never matches.
func nginxUserDirectivePattern(name string) *regexp.Regexp {
	return regexp.MustCompile(`(?m)^([ \t]*)` + regexp.QuoteMeta(name) + `[ \t]+[^;\n]*;`)
}

// nginxConfigDefinesDirective reports whether a directive is set anywhere in
// the file, ignoring commented-out lines.
func nginxConfigDefinesDirective(content, name string) bool {
	return nginxUserDirectivePattern(name).MatchString(content)
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

// getNginxBrotliParams reports the brotli settings currently in effect, and
// where they come from.
//
// Brotli is normally served from the managed http.d file instead of
// nginx.conf, so the directives can be removed together with the module. When
// the module is disabled the declared defaults are returned, which lets the
// settings page show what would be applied once it is enabled.
//
// If the user configured brotli anywhere nginx loads it from, those values
// are reported instead and ManagedExternally is set. Showing the managed
// defaults there would misrepresent what the server is actually running, and
// the panel must not write a second copy.
func getNginxBrotliParams() (*response.NginxBrotliRes, error) {
	install, err := getAppInstallByKey(constant.AppOpenresty)
	if err != nil {
		return nil, err
	}
	managedExternally := nginxModuleConfiguredByUser(install, nginxBrotliModuleName)
	var current map[string][]string
	if managedExternally {
		current = readNginxUserBrotliDirectives(install)
	} else {
		fileName := nginxHTTPConfigFileName(nginxModuleRuntimeOrder(nginxBrotliModuleName), nginxBrotliModuleName)
		current = readNginxHTTPDirectives(path.Join(nginxHTTPConfigDir(install), fileName))
	}
	res := &response.NginxBrotliRes{
		ManagedExternally: managedExternally,
		// Without the include, values the panel would write would never reach
		// nginx, so they are reported as unavailable rather than shown as if
		// they were in effect.
		ManagedUnavailable: !managedExternally && !nginxHTTPIncludePresent(install),
	}
	for _, directive := range mergeNginxRuntimeDirectives(nginxModuleRuntimeDefaults[nginxBrotliModuleName], current) {
		res.Params = append(res.Params, response.NginxParam{Name: directive.Name, Params: directive.Params})
	}
	return res, nil
}

// readNginxUserBrotliDirectives collects the brotli directives the user wrote
// in any of the files nginx loads them from.
func readNginxUserBrotliDirectives(install model.AppInstall) map[string][]string {
	directives := make(map[string][]string)
	for _, filePath := range nginxModuleUserConfigPaths(install) {
		content, err := os.ReadFile(filePath)
		if err != nil {
			continue
		}
		for _, name := range dto.BrotliKeys {
			pattern := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(name) + `[ \t]+([^;\n]*);`)
			if match := pattern.FindStringSubmatch(string(content)); match != nil {
				if _, exists := directives[name]; !exists {
					directives[name] = strings.Fields(strings.TrimSpace(match[1]))
				}
			}
		}
	}
	return directives
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
		// The user configured brotli in nginx.conf before the panel managed
		// it. Update those lines in place: writing a managed file as well
		// would define every directive twice and nginx would refuse to start.
		if nginxModuleConfiguredByUser(install, nginxBrotliModuleName) {
			return updateUserNginxBrotliParams(install, values)
		}
		// A managed write needs the include. Installations missing it are
		// upgraded in place here; when nginx.conf cannot be edited safely the
		// write is refused with an actionable error instead of writing values
		// nginx would never load.
		if !nginxHTTPIncludePresent(install) {
			configPath := nginxMainConfigPath(install)
			content, readErr := os.ReadFile(configPath)
			if readErr != nil {
				return readErr
			}
			updated, insErr := insertNginxHTTPInclude(string(content))
			if insErr != nil {
				return buserr.New("ErrBrotliUnsupported")
			}
			if err = os.WriteFile(configPath, []byte(updated), constant.FilePerm); err != nil {
				return err
			}
			if err = os.MkdirAll(nginxHTTPConfigDir(install), constant.DirPerm); err != nil {
				return err
			}
			if err = nginxCheckAndReload(string(content), configPath, install.ContainerName); err != nil {
				return err
			}
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

// updateUserNginxBrotliParams rewrites the brotli directives the user wrote
// into nginx.conf, in place.
//
// Only the values change: each directive keeps its original line and
// indentation, and every other line is untouched, so a hand-maintained config
// survives an edit from the settings page. Directives the user did not write
// are not introduced, since the panel cannot know where they intended them.
func updateUserNginxBrotliParams(install model.AppInstall, values map[string][]string) error {
	configPath := nginxMainConfigPath(install)
	content, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}
	updated := string(content)
	for _, name := range dto.BrotliKeys {
		params, ok := values[name]
		if !ok || len(params) == 0 {
			continue
		}
		pattern := nginxUserDirectivePattern(name)
		if !pattern.MatchString(updated) {
			continue
		}
		replacement := "${1}" + name + " " + strings.Join(params, " ") + ";"
		updated = pattern.ReplaceAllString(updated, replacement)
	}
	if updated == string(content) {
		return nil
	}
	if err = os.WriteFile(configPath, []byte(updated), constant.FilePerm); err != nil {
		return err
	}
	return nginxCheckAndReload(string(content), configPath, install.ContainerName)
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
