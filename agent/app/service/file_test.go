package service

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/1Panel-dev/1Panel/agent/app/dto/request"
	"github.com/1Panel-dev/1Panel/agent/buserr"
)

func TestChangeNameRejectsExistingTarget(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old.txt")
	targetPath := filepath.Join(dir, "target.txt")
	// Binary source content keeps this test independent of the file-history DB.
	if err := os.WriteFile(oldPath, []byte("old\x00"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(targetPath, []byte("target"), 0644); err != nil {
		t.Fatal(err)
	}

	err := (&FileService{}).ChangeName(request.FileRename{OldName: oldPath, NewName: targetPath})
	if err == nil {
		t.Fatal("expected an error when the target name already exists")
	}
	var bizErr buserr.BusinessError
	if !errors.As(err, &bizErr) || bizErr.Msg != "ErrFileIsExist" {
		t.Fatalf("expected ErrFileIsExist, got %v", err)
	}

	content, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != "target" {
		t.Fatalf("target file was overwritten, content = %q", content)
	}
	if _, err := os.Stat(oldPath); err != nil {
		t.Fatalf("source file should still exist after a rejected rename: %v", err)
	}
}

func TestChangeNameRenamesToNewName(t *testing.T) {
	dir := t.TempDir()
	oldPath := filepath.Join(dir, "old")
	newPath := filepath.Join(dir, "new")
	if err := os.Mkdir(oldPath, 0755); err != nil {
		t.Fatal(err)
	}

	if err := (&FileService{}).ChangeName(request.FileRename{OldName: oldPath, NewName: newPath}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(newPath); err != nil {
		t.Fatalf("renamed path does not exist: %v", err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("old path should be gone, stat err = %v", err)
	}
}
