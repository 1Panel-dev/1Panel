package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto/response"
	"github.com/1Panel-dev/1Panel/agent/cmd/server/nginx_conf"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func TestCreateSwitchDefaultIndexFileUsesCustomDefaultPage(t *testing.T) {
	fileOp := files.NewFileOp()

	t.Run("php uses custom default html", func(t *testing.T) {
		dir := t.TempDir()
		var gotType string
		createSwitchDefaultIndexFile(fileOp, dir, true, func(resourceType string) (*response.WebsiteHtmlRes, error) {
			gotType = resourceType
			return &response.WebsiteHtmlRes{Content: "<?php echo 'custom'; ?>"}, nil
		})
		if gotType != "php" {
			t.Fatalf("expected resourceType php, got %q", gotType)
		}
		data, err := os.ReadFile(filepath.Join(dir, "index.php"))
		if err != nil {
			t.Fatalf("expected index.php to be created: %v", err)
		}
		if string(data) != "<?php echo 'custom'; ?>" {
			t.Fatalf("unexpected index.php content: %q", string(data))
		}
	})

	t.Run("static uses custom default html", func(t *testing.T) {
		dir := t.TempDir()
		var gotType string
		createSwitchDefaultIndexFile(fileOp, dir, false, func(resourceType string) (*response.WebsiteHtmlRes, error) {
			gotType = resourceType
			return &response.WebsiteHtmlRes{Content: "<h1>custom static</h1>"}, nil
		})
		if gotType != "index" {
			t.Fatalf("expected resourceType index, got %q", gotType)
		}
		data, err := os.ReadFile(filepath.Join(dir, "index.html"))
		if err != nil {
			t.Fatalf("expected index.html to be created: %v", err)
		}
		if string(data) != "<h1>custom static</h1>" {
			t.Fatalf("unexpected index.html content: %q", string(data))
		}
	})

	t.Run("does not overwrite existing file", func(t *testing.T) {
		dir := t.TempDir()
		existing := filepath.Join(dir, "index.php")
		if err := os.WriteFile(existing, []byte("<?php echo 'existing'; ?>"), 0o644); err != nil {
			t.Fatal(err)
		}
		called := false
		createSwitchDefaultIndexFile(fileOp, dir, true, func(string) (*response.WebsiteHtmlRes, error) {
			called = true
			return &response.WebsiteHtmlRes{Content: "<?php echo 'custom'; ?>"}, nil
		})
		if called {
			t.Fatal("getDefaultHtml should not be called when index file already exists")
		}
		data, err := os.ReadFile(existing)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "<?php echo 'existing'; ?>" {
			t.Fatalf("existing index.php was modified: %q", string(data))
		}
	})

	t.Run("falls back to embedded default on lookup error", func(t *testing.T) {
		dir := t.TempDir()
		createSwitchDefaultIndexFile(fileOp, dir, true, func(string) (*response.WebsiteHtmlRes, error) {
			return nil, errors.New("lookup failed")
		})
		data, err := os.ReadFile(filepath.Join(dir, "index.php"))
		if err != nil {
			t.Fatalf("expected index.php to be created: %v", err)
		}
		if string(data) != string(nginx_conf.IndexPHP) {
			t.Fatalf("expected fallback IndexPHP content, got %q", string(data))
		}
	})
}

func TestHasExistingIndexFile(t *testing.T) {
	t.Run("empty directory returns false", func(t *testing.T) {
		dir := t.TempDir()
		if hasExistingIndexFile(dir, "index.php", "index.html", "index.htm") {
			t.Fatal("expected false for empty directory")
		}
	})

	t.Run("detects index.html for php website", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<h1>hello</h1>"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !hasExistingIndexFile(dir, "index.php", "index.html", "index.htm") {
			t.Fatal("expected true when index.html exists")
		}
	})

	t.Run("detects index.htm for php and static websites", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "index.htm"), []byte("<h1>hello</h1>"), 0o644); err != nil {
			t.Fatal(err)
		}
		if !hasExistingIndexFile(dir, "index.php", "index.html", "index.htm") {
			t.Fatal("expected true when index.htm exists")
		}
		if !hasExistingIndexFile(dir, "index.html", "index.htm") {
			t.Fatal("expected true for static candidates when index.htm exists")
		}
	})

	t.Run("ignores directory with same name", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.Mkdir(filepath.Join(dir, "index.html"), 0o755); err != nil {
			t.Fatal(err)
		}
		if hasExistingIndexFile(dir, "index.php", "index.html", "index.htm") {
			t.Fatal("expected false when candidate is a directory")
		}
	})
}
