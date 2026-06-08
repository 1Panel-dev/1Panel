package service

import (
	"os"
	"testing"
)

func TestBuiltinRewriteNameExistIgnoresCase(t *testing.T) {
	if !builtinRewriteNameExist("WordPress") {
		t.Fatal("expected WordPress to match builtin wordpress rewrite")
	}
	if !builtinRewriteNameExist("EmpireCMS") {
		t.Fatal("expected EmpireCMS to match builtin empirecms rewrite")
	}
	if builtinRewriteNameExist("not-exist-rewrite") {
		t.Fatal("expected unknown rewrite name to be absent")
	}
}

func TestCustomRewriteNameExistIgnoresCase(t *testing.T) {
	rewriteDir := t.TempDir()
	if err := os.WriteFile(rewriteDir+"/MyRewrite.conf", []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		want bool
	}{
		{name: "MyRewrite", want: true},
		{name: "myrewrite", want: true},
		{name: "MYREWRITE", want: true},
		{name: "OtherRewrite", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := customRewriteNameExist(rewriteDir, tt.name)
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Fatalf("customRewriteNameExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSafeRewriteName(t *testing.T) {
	if _, err := getSafeRewriteName("../rewrite"); err == nil {
		t.Fatal("expected path traversal name to be rejected")
	}
	if _, err := getSafeRewriteName("rewrite/name"); err == nil {
		t.Fatal("expected path separator name to be rejected")
	}
	if name, err := getSafeRewriteName("MyRewrite"); err != nil || name != "MyRewrite" {
		t.Fatalf("getSafeRewriteName() = %q, %v; want MyRewrite, nil", name, err)
	}
}
