package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/utils/nginx"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/components"
	"github.com/1Panel-dev/1Panel/agent/utils/nginx/parser"
)

// An HTTPS site: one server block listening on 80 and 443, 443 serving TLS.
const httpsSiteConf = `server {
    listen 80 ;
    listen 443 ssl ;
    server_name a.example.com;

    location / {
        return 200 "ok";
    }
}
`

func listenParams(server *components.Server) []string {
	var out []string
	for _, listen := range server.Listens {
		out = append(out, listen.Bind+" "+strings.Join(listen.Parameters, " "))
	}
	return out
}

// nginxAccepts runs the real nginx binary over the dumped site config. It is
// the only authority on whether the panel wrote a usable server block.
func nginxAccepts(t *testing.T, siteConfig string) (string, bool) {
	t.Helper()
	root := t.TempDir()
	site := filepath.Join(root, "site.conf")
	if err := os.WriteFile(site, []byte(siteConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	cert, key := filepath.Join(root, "cert.pem"), filepath.Join(root, "key.pem")
	if out, err := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
		"-keyout", key, "-out", cert, "-days", "1", "-subj", "/CN=a.example.com").CombinedOutput(); err != nil {
		t.Skipf("openssl unavailable, cannot validate with nginx: %s", out)
	}
	main := filepath.Join(root, "nginx.conf")
	conf := fmt.Sprintf(`pid %s;
error_log %s;
events {}
http {
    ssl_certificate %s;
    ssl_certificate_key %s;
    include %s;
}
`, filepath.Join(root, "nginx.pid"), filepath.Join(root, "error.log"), cert, key, site)
	if err := os.WriteFile(main, []byte(conf), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := exec.Command("nginx", "-t", "-c", main, "-p", root).CombinedOutput()
	return string(out), err == nil
}

// addListenAndServerName calls setListen for each new domain with
// `ssl = protocol == HTTPS && domain.SSL`. The add-domain form starts the SSL
// checkbox unchecked, so a second domain added on the site's HTTPS port passes
// ssl=false. The port must keep serving TLS: the site's other domains and its
// certificate are still bound to it.
func TestAddDomainWithoutCertificateKeepsPortTLS(t *testing.T) {
	config, err := parser.NewStringParser(httpsSiteConf).Parse()
	if err != nil {
		t.Fatal(err)
	}
	server := config.FindServers()[0]

	// Exactly the call addListenAndServerName makes for the new domain.
	setListen(server, "443", false, false, false, false)

	dumped := nginx.DumpConfig(config, nginx.IndentedStyle)
	t.Logf("listens after add: %q\n%s", listenParams(server), dumped)

	if !strings.Contains(dumped, "listen 443 ssl") {
		t.Fatalf("adding a domain without its own certificate turned the site's HTTPS port into plain HTTP:\n%s", dumped)
	}
	if strings.Contains(dumped, "listen 443 ;") {
		t.Fatalf("the unencrypted 443 listen replaced the TLS one:\n%s", dumped)
	}
	if out, ok := nginxAccepts(t, dumped); !ok {
		t.Fatalf("nginx rejected the written config: %s\n%s", out, dumped)
	}
}

// The HTTP/3 listener on the same port must survive too.
func TestAddDomainWithoutCertificateKeepsQUIC(t *testing.T) {
	const src = `server {
    listen 80 ;
    listen 443 ssl ;
    listen 443 quic ;
    server_name a.example.com;

    location / {
        return 200 "ok";
    }
}
`
	config, err := parser.NewStringParser(src).Parse()
	if err != nil {
		t.Fatal(err)
	}
	server := config.FindServers()[0]
	setListen(server, "443", false, false, false, false)

	dumped := nginx.DumpConfig(config, nginx.IndentedStyle)
	t.Logf("listens after add: %q\n%s", listenParams(server), dumped)
	if !strings.Contains(dumped, "listen 443 quic") {
		t.Fatalf("the QUIC listen was dropped:\n%s", dumped)
	}
	if !strings.Contains(dumped, "listen 443 ssl") {
		t.Fatalf("the TLS listen was dropped:\n%s", dumped)
	}
}

// A port that does not serve TLS yet still follows the new domain's checkbox.
func TestNewPortFollowsCertificateFlag(t *testing.T) {
	config, err := parser.NewStringParser(httpsSiteConf).Parse()
	if err != nil {
		t.Fatal(err)
	}
	server := config.FindServers()[0]
	setListen(server, "8080", false, false, false, false)

	dumped := nginx.DumpConfig(config, nginx.IndentedStyle)
	t.Logf("listens after add: %q\n%s", listenParams(server), dumped)
	if strings.Contains(dumped, "listen 8080 ssl") {
		t.Fatalf("a port whose domain has no certificate must not listen with TLS:\n%s", dumped)
	}
	if !strings.Contains(dumped, "listen 8080 ;") {
		t.Fatalf("missing plain listen:\n%s", dumped)
	}
}

// Disabling HTTPS still works: removeSSLListen deletes the bind first, so the
// port comes back without TLS.
func TestRemoveSSLListenStillDisablesTLS(t *testing.T) {
	config, err := parser.NewStringParser(httpsSiteConf).Parse()
	if err != nil {
		t.Fatal(err)
	}
	server := config.FindServers()[0]
	server.DeleteListen("443")
	setListen(server, "443", false, false, false, false)

	dumped := nginx.DumpConfig(config, nginx.IndentedStyle)
	t.Logf("listens after disable: %q\n%s", listenParams(server), dumped)
	if strings.Contains(dumped, "listen 443 ssl") {
		t.Fatalf("TLS was not removed when disabling HTTPS:\n%s", dumped)
	}
}
