package service

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/constant"
	"github.com/1Panel-dev/1Panel/agent/global"
)

const userBrotliConf = `user  root;
worker_processes  auto;

include /usr/local/openresty/nginx/conf/modules-enabled/*.conf;

events { use epoll; }

http {
    include       mime.types;
    default_type  application/octet-stream;

    gzip on;
    gzip_comp_level 5;

    # enabled by hand, long before the panel managed it
    brotli on;
    brotli_comp_level 6;
    brotli_types text/plain text/css application/json;

    include /usr/local/openresty/nginx/conf/http.d/*.conf;
    include /usr/local/openresty/nginx/conf/conf.d/*.conf;
}
`

func TestNginxConfigDefinesDirective(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"directive present", userBrotliConf, true},
		{"absent", strings.Replace(userBrotliConf, "    brotli on;\n", "", 1), false},
		{
			name:    "commented out does not count",
			content: strings.Replace(userBrotliConf, "    brotli on;", "    # brotli on;", 1),
			want:    false,
		},
		{
			name:    "a longer directive name is not a match",
			content: "http {\n    brotli_comp_level 6;\n}\n",
			want:    false,
		},
		{
			name:    "indentation does not matter",
			content: "http {\n\t\tbrotli on;\n}\n",
			want:    true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := nginxConfigDefinesDirective(tc.content, "brotli"); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// The detection must catch any brotli* directive, in any file nginx loads it
// from, not only the primary directive in nginx.conf.
func TestNginxModuleUserDirectiveRe(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    bool
	}{
		{"primary directive", "http {\n    brotli on;\n}", true},
		{"a tuning directive alone", "server {\n    brotli_comp_level 11;\n}", true},
		{"another variant", "http {\n    brotli_types text/plain;\n}", true},
		{"server scope in a site file", "server {\n    listen 80;\n    brotli on;\n}", true},
		{"commented out does not count", "http {\n    # brotli on;\n}", false},
		{"indented comment does not count", "http {\n        # brotli_comp_level 6;\n}", false},
		{"no brotli at all", "http {\n    gzip on;\n}", false},
		{"a similarly named directive is not a match", "http {\n    gzip on;\n}", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := nginxModuleUserDirectiveRe.MatchString(tc.content); got != tc.want {
				t.Fatalf("expected %v, got %v", tc.want, got)
			}
		})
	}
}

// The panel must not emit a managed file for a module the user already
// configured: nginx rejects the same directive defined twice.
func TestUserConfiguredBrotliSuppressesManagedFile(t *testing.T) {
	if !nginxModuleUserDirectiveRe.MatchString(userBrotliConf) {
		t.Fatal("a hand-written brotli config must be detected")
	}
	clean := strings.Replace(userBrotliConf, "    brotli on;\n", "", 1)
	clean = strings.Replace(clean, "    brotli_comp_level 6;\n", "", 1)
	clean = strings.Replace(clean, "    brotli_types text/plain text/css application/json;\n", "", 1)
	if nginxModuleUserDirectiveRe.MatchString(clean) {
		t.Fatal("a config without brotli must not be treated as user-managed")
	}
}

func TestRewriteUserBrotliDirectivesInPlace(t *testing.T) {
	// Mirrors updateUserNginxBrotliParams without touching the filesystem.
	rewrite := func(content string, values map[string][]string) string {
		updated := content
		for _, name := range []string{"brotli", "brotli_comp_level", "brotli_min_length", "brotli_types"} {
			params, ok := values[name]
			if !ok || len(params) == 0 {
				continue
			}
			pattern := nginxUserDirectivePattern(name)
			if !pattern.MatchString(updated) {
				continue
			}
			updated = pattern.ReplaceAllString(updated, "${1}"+name+" "+strings.Join(params, " ")+";")
		}
		return updated
	}

	got := rewrite(userBrotliConf, map[string][]string{
		"brotli":            {"off"},
		"brotli_comp_level": {"4"},
		// brotli_min_length is absent from the user's config and must not be
		// introduced: the panel cannot know where they would want it.
		"brotli_min_length": {"2k"},
	})

	if !strings.Contains(got, "    brotli off;") {
		t.Errorf("value was not updated:\n%s", got)
	}
	if !strings.Contains(got, "    brotli_comp_level 4;") {
		t.Errorf("comp level was not updated:\n%s", got)
	}
	if strings.Contains(got, "brotli_min_length") {
		t.Error("a directive the user never wrote must not be added")
	}
	// Everything else survives, including the comment the parser would drop.
	for _, keep := range []string{
		"# enabled by hand, long before the panel managed it",
		"    gzip on;",
		"    gzip_comp_level 5;",
		"    brotli_types text/plain text/css application/json;",
		"include /usr/local/openresty/nginx/conf/conf.d/*.conf;",
		"worker_processes  auto;",
	} {
		if !strings.Contains(got, keep) {
			t.Errorf("unrelated line was altered or lost: %s", keep)
		}
	}
	if strings.Count(got, "brotli on;")+strings.Count(got, "brotli off;") != 1 {
		t.Error("the directive must remain defined exactly once")
	}
}

func TestRewriteUserBrotliPreservesIndentation(t *testing.T) {
	content := "http {\n\t\tbrotli on;\n}\n"
	pattern := nginxUserDirectivePattern("brotli")
	got := pattern.ReplaceAllString(content, "${1}brotli off;")
	if !strings.Contains(got, "\t\tbrotli off;") {
		t.Errorf("original indentation was not preserved: %q", got)
	}
}

// Detection must cover the default/ directory, which is included at http
// scope like conf.d but lives under the install directory, not the site root.
func TestNginxModuleUserConfigPathsCoversDefaultDir(t *testing.T) {
	siteDir := t.TempDir()
	installRoot := t.TempDir()
	installDir := path.Join(installRoot, "openresty", "openresty")
	defaultDir := path.Join(installDir, "conf", "default")
	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path.Join(defaultDir, "00.default.conf"), []byte("server {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	global.Dir.AppInstallDir = installRoot
	install := model.AppInstall{Name: "openresty"}
	install.App.Key = constant.AppOpenresty

	paths := nginxModuleUserConfigPathsWithSiteDir(install, siteDir)
	foundMain, foundDefault := false, false
	for _, p := range paths {
		if strings.HasSuffix(p, path.Join("conf", "nginx.conf")) {
			foundMain = true
		}
		if strings.HasSuffix(p, path.Join("conf", "default", "00.default.conf")) {
			foundDefault = true
		}
	}
	if !foundMain {
		t.Error("nginx.conf must be scanned")
	}
	if !foundDefault {
		t.Error("conf/default must be scanned: it is included at http scope")
	}
}
