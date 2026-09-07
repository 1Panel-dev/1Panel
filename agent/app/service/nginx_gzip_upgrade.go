package service

import (
	"os"
	"path"
	"regexp"
	"sort"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/global"
)

// stockNginxGzipDirectives is the gzip block shipped by the OpenResty app
// since 1.21.4.3. The upgrade only rewrites values when the installed
// nginx.conf still carries exactly these directives and values, which proves
// the user never tuned compression. Any deviation aborts the rewrite.
var stockNginxGzipDirectives = map[string]string{
	"gzip":              "on",
	"gzip_min_length":   "1k",
	"gzip_buffers":      "4 16k",
	"gzip_http_version": "1.1",
	"gzip_comp_level":   "2",
	"gzip_types":        "text/plain application/javascript application/x-javascript text/javascript text/css application/xml",
	"gzip_vary":         "on",
	"gzip_proxied":      "expired no-cache no-store private auth",
	"gzip_disable":      `"MSIE [1-6]\."`,
}

// correctedNginxGzipDirectives replaces the stock values in place. gzip lives
// in the http block of nginx.conf and must stay there: repeating it from an
// included file would make nginx reject the configuration with a duplicate
// directive error, and the compression settings page reads and writes these
// same keys in nginx.conf.
var correctedNginxGzipDirectives = map[string]string{
	"gzip_comp_level": "5",
	"gzip_types":      strings.Join(nginxCompressibleTypes, " "),
	"gzip_proxied":    "any",
}

// obsoleteNginxGzipDirectives are dropped outright.
var obsoleteNginxGzipDirectives = map[string]struct{}{
	// A per-request User-Agent regex for browsers with no measurable share.
	"gzip_disable": {},
}

var nginxGzipDirectiveRe = regexp.MustCompile(`(?m)^[ \t]*(gzip[a-z_]*)[ \t]+([^;\n]*);[ \t]*$`)

func nginxMainConfigPath(install model.AppInstall) string {
	return path.Join(install.GetPath(), nginxModuleConfDir, "nginx.conf")
}

// upgradeStockNginxGzipConfig rewrites the factory gzip defaults in place.
//
// Upgrades deliberately preserve the user's nginx.conf, so corrected defaults
// shipped with a new OpenResty version would otherwise never reach existing
// installations.
//
// The config parser is not used: its dumper regenerates the whole file, drops
// standalone comments and reorders proxy includes, which would be destructive
// on a user's main config. Lines are edited individually so everything outside
// the gzip block stays byte-identical.
func upgradeStockNginxGzipConfig(install model.AppInstall) error {
	configPath := nginxMainConfigPath(install)
	content, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !isStockNginxGzipConfig(string(content)) {
		return nil
	}
	updated := rewriteNginxGzipDirectives(string(content))
	if updated == string(content) {
		return nil
	}
	if err = writeNginxFileAtomic(configPath, []byte(updated)); err != nil {
		return err
	}
	if err = nginxCheckAndReload(string(content), configPath, install.ContainerName); err != nil {
		return err
	}
	global.LOG.Info("updated the stock OpenResty gzip configuration to the current defaults")
	return nil
}

// isStockNginxGzipConfig reports whether every gzip directive in the config
// matches the factory defaults exactly, with none missing and none extra.
func isStockNginxGzipConfig(content string) bool {
	found := make(map[string]string)
	for _, match := range nginxGzipDirectiveRe.FindAllStringSubmatch(content, -1) {
		name := match[1]
		value := strings.Join(strings.Fields(match[2]), " ")
		if _, ok := found[name]; ok {
			// A directive repeated in the http block means the config was
			// edited by hand; leave it alone.
			return false
		}
		found[name] = value
	}
	if len(found) != len(stockNginxGzipDirectives) {
		return false
	}
	for name, expected := range stockNginxGzipDirectives {
		if found[name] != expected {
			return false
		}
	}
	return true
}

// rewriteNginxGzipDirectives updates known values in place, drops obsolete
// directives and appends directives that are missing, preserving the original
// indentation and leaving every other line untouched.
func rewriteNginxGzipDirectives(content string) string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0, len(lines))
	seen := make(map[string]struct{})
	lastGzipIndex := -1
	lastGzipIndent := "    "

	for _, line := range lines {
		match := nginxGzipDirectiveRe.FindStringSubmatch(line)
		if match == nil {
			result = append(result, line)
			continue
		}
		name := match[1]
		// The indentation belongs to the line itself; a top-level directive
		// must not inherit the indent a previous, nested directive used.
		lineIndent := line[:len(line)-len(strings.TrimLeft(line, " \t"))]
		if lineIndent == "" {
			lineIndent = "    "
		}
		lastGzipIndent = lineIndent
		if _, obsolete := obsoleteNginxGzipDirectives[name]; obsolete {
			continue
		}
		seen[name] = struct{}{}
		if replacement, ok := correctedNginxGzipDirectives[name]; ok {
			result = append(result, lineIndent+name+" "+replacement+";")
		} else {
			result = append(result, line)
		}
		lastGzipIndex = len(result) - 1
	}

	// Directives introduced by a newer default set are appended right after
	// the existing block so they stay visually grouped.
	var missing []string
	for name := range correctedNginxGzipDirectives {
		if _, ok := seen[name]; !ok {
			missing = append(missing, name)
		}
	}
	if len(missing) == 0 || lastGzipIndex < 0 {
		return strings.Join(result, "\n")
	}
	sort.Strings(missing)
	added := make([]string, 0, len(missing))
	for _, name := range missing {
		added = append(added, lastGzipIndent+name+" "+correctedNginxGzipDirectives[name]+";")
	}
	tail := append(added, result[lastGzipIndex+1:]...)
	return strings.Join(append(result[:lastGzipIndex+1], tail...), "\n")
}
