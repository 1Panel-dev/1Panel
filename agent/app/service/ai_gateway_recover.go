package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/1Panel-dev/1Panel/agent/app/model"
	"github.com/1Panel-dev/1Panel/agent/buserr"
	"github.com/1Panel-dev/1Panel/agent/global"
	"github.com/1Panel-dev/1Panel/agent/utils/files"
)

func prepareAIGatewayAppRecover(install *model.AppInstall, recoverFile, secret string) (string, func(), error) {
	noop := func() {}
	if !isAIGatewayApp(install.App.Key) {
		return "", noop, nil
	}
	tmpDir, err := os.MkdirTemp(global.Dir.TmpDir, "ai-gateway-recover-")
	if err != nil {
		return "", noop, err
	}
	cleanup := func() { _ = os.RemoveAll(tmpDir) }
	fail := func(err error) (string, func(), error) {
		cleanup()
		return "", noop, err
	}
	if err := files.NewFileOp().TarGzExtractPro(recoverFile, tmpDir, secret); err != nil {
		return fail(err)
	}
	backupDir := filepath.Join(tmpDir, strings.ReplaceAll(filepath.Base(recoverFile), ".tar.gz", ""))
	for _, name := range []string{"app.json", "app.tar.gz"} {
		info, err := os.Lstat(filepath.Join(backupDir, name))
		if err != nil || !info.Mode().IsRegular() {
			return fail(buserr.New("AppBackupFileIncomplete"))
		}
	}
	metadata, err := os.ReadFile(filepath.Join(backupDir, "app.json"))
	if err != nil {
		return fail(err)
	}
	var backup model.AppInstall
	if err := json.Unmarshal(metadata, &backup); err != nil {
		return fail(fmt.Errorf("unmarshal app.json failed: %w", err))
	}
	if backup.App.Key != install.App.Key || backup.Name != install.Name {
		return fail(buserr.New("AppAttributesNotMatch"))
	}
	// Historical backups already include Version. Unsuffixed releases are
	// SQLite; no new metadata field is required. Missing versions are invalid,
	// not evidence that an archive contains SQLite data.
	if strings.TrimSpace(backup.Version) == "" || strings.TrimSpace(install.Version) == "" {
		return fail(buserr.New("AppBackupFileIncomplete"))
	}
	// Deliberately do not compare version order: restoring an older version of
	// the same database type is required for failed-upgrade rollbacks.
	if !isAIGatewayUpgradeVersionAllowed(install.App.Key, install.Version, backup.Version) {
		return fail(buserr.New("ErrAIGatewayRecoverDatabaseTypeMismatch"))
	}
	return backupDir, cleanup, nil
}
