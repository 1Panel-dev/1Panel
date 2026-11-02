package helper

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"

	"github.com/1Panel-dev/1Panel/agent/utils/vm/protocol"
)

func EnsureEmbeddedHelper() (string, bool, error) {
	bin, ok := embeddedHelperBytes()
	if !ok {
		return "", false, nil
	}
	target, err := embeddedHelperPath()
	if err != nil {
		return "", true, err
	}
	if sameFileContent(target, bin) {
		if err := os.Chmod(target, 0755); err != nil {
			return "", true, err
		}
		return target, true, nil
	}
	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", true, err
	}
	tmp := target + ".tmp"
	if err := os.WriteFile(tmp, bin, 0755); err != nil {
		return "", true, err
	}
	if err := os.Rename(tmp, target); err != nil {
		_ = os.Remove(tmp)
		return "", true, err
	}
	if err := os.Chmod(target, 0755); err != nil {
		return "", true, err
	}
	return target, true, nil
}

func embeddedHelperPath() (string, error) {
	self, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("load executable path: %w", err)
	}
	return filepath.Join(filepath.Dir(self), protocol.HelperBinary), nil
}

func sameFileContent(path string, expected []byte) bool {
	current, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return sha256.Sum256(current) == sha256.Sum256(expected)
}
